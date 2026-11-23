# HTTP/3 upstream maintenance

HTTP/3 is part of the `github.com/enetx/quic-go` module and is imported as
`github.com/enetx/quic-go/http3`. The supported baseline is Go 1.27+.

- Source: [quic-go v0.65.0](https://github.com/quic-go/quic-go/tree/v0.65.0/http3).
- Upstream commit: `aa265c3fc6489e3bfa8071b333be95e0e887a51d`.
- Previous standalone fork: `ba6a50293c3c477f83648f3faa9e58adaa807206` (`v1.0.9`).

The former standalone source snapshot is retained in `archive/http3/` at the
workspace root. Use the shared
[update procedure](../UPSTREAM.md) and [source pins](../upstream.json) for future
updates. HTTP/3 has no separate `go.mod`.

## Audited behavior

Before updating, all production files, local tests and the standalone fork
history were compared against quic-go v0.60.0. The fork had selected later
connection and frame fixes. Its custom changes were in `client.go`, `conn.go`,
`frames.go`, `request_writer.go`, `server.go`, `transport.go`, and the former
`httpcommon` sorter.

Preserve these behaviors when updating:

- Use the fork's HTTP, tracing, test-server and HTTP/3 qlog packages.
- `Header-Order:` controls regular request header order. Listed fields precede
  unlisted fields; unlisted fields sort lexicographically. Multiple values retain
  their order. The stateless sorter exported by `enetx/http/headerorder` is shared
  with HTTP/1 and HTTP/2.
- `PHeader-Order:` controls pseudo-header order, including `:protocol` for Extended
  CONNECT. The default order is `:authority`, `:method`, `:path`, `:scheme`.
  CONNECT omits `:path` and `:scheme` where the protocol requires it.
- Both order keys are metadata and never reach the wire. Other invalid headers
  remain errors.
- `AdditionalSettings`, `Settings.Other` and `settingsFrame.Other` use
  `g.MapOrd[uint64, uint64]`. Unknown settings preserve insertion order and retain
  duplicate validation. Only qlog snapshots convert them to ordinary maps.
- Clients omit automatic `SETTINGS_MAX_FIELD_SECTION_SIZE` while enforcing
  `MaxResponseHeaderBytes` locally. Explicit profile settings remain supported.
- `settingsFrame.Append` emits `MaxFieldSectionSize` only when positive.
  Parsing an explicit peer-supplied zero remains valid. The client initializes
  this field to `-1`, so qlog agrees with its omission from the wire.
- The 0-RTT compatibility check treats zero as omitted, matching the encoder.
- Invalid-header-value errors retain the detailed diagnostics from the fork.

## Client HTTP/3 resumption state

Configured QPACK clients dispatch a complete already buffered peer SETTINGS
prelude before normal request headers. `client_buffered_settings.go` uses
`ReceiveStream.TryPeek` / `Conn.TryAcceptUniStream`; absent, partial or large
preludes continue on the ordinary asynchronous reader. A per-dial input handoff
is shared by early-restart generations. Genuine 0-RTT does not wait for it.
Client handshake notification follows the current QUIC receive batch, including
coalesced 1-RTT packets, without awaiting additional network input. Preserve this
ordering when updating `connection.go`, stream maps and the HTTP/3 dial loop.

This fixes a captured first-request static/dynamic QPACK race. The engine carries
no browser capacity or delay policy. Real-wire regressions create full/partial
buffered SETTINGS before HTTP/3, prove missing bytes never hold a request, and
check late activation plus the restart wrapper. Native first-section vectors
cover both desktop profiles at four capacities; Surf retains the full cold
capture manifest. Stream peeks preserve offsets, EOF, cancellation, deadlines
and concurrent reader ownership.

`client_session_settings.go` binds the received server SETTINGS to the exact
TLS ticket through `tls.SessionState.Extra`. The QUIC engine exposes a single
per-connection TLS factory decorator; it contains no HTTP/3 or browser policy.
The HTTP/3 transport composes that decorator with native or custom TLS factories,
including Surf's uTLS adapter. No TLS-engine patch is needed for this feature.

Tickets received before SETTINGS wait in a bounded per-connection queue. Pending
insertions retain their opaque cache keys and order; a racing connection cannot
supply their state. Missing or invalid metadata disables early data while keeping
ordinary TLS resumption. Application-owned ticket objects remain unchanged.

On 0-RTT, restore peer QPACK limits and the maximum uncompressed field-section
size. QPACK tables themselves are always fresh. Compatible actual SETTINGS retain
new table entries and do not repeat the capacity instruction. Incompatible
accepted early data closes the connection; rejected early data uses fresh limits
on Surf's post-handshake retry. Unknown settings remain metadata, not implemented
extension capabilities. Request headers and trailers enforce the peer's known
field-section limit before changing the encoder or emitting their HEADERS frame.

References: [HTTP/3](https://www.rfc-editor.org/rfc/rfc9114#section-7.2.4.2),
[QPACK](https://www.rfc-editor.org/rfc/rfc9204#section-3.2.3),
[datagrams](https://www.rfc-editor.org/rfc/rfc9297#section-2.1.1).
Keep the cache, TLS-event and raw-wire regressions when updating upstream;
`settings_0rtt.go` separately retains the server's encrypted-ticket checks.

The former pooled sorter has been removed. The shared stateless sorter preserves
its observable order without cached mutable state or a global lock.

## Authenticated initial peer settings and request preparation

`Transport.InitialPeerSettings` supplies raw identifier/value pairs for the exact
selected QUIC connection, once per authenticated HTTP/3 protocol generation.
`client_peer_settings.go` shares SETTINGS validation, accepted-0RTT compatibility,
QPACK limits and field-section limits with the control stream. Explicit values
overlay existing limits; omitted IDs retain the prelude/remembered values.
Explicit boolean zero remains distinct from omission. Callback results are copied
into generation-owned state. A canceled waiter can leave initialization owned by
another request; an early-retired generation cannot publish a late callback.

The initial input does not close `ReceivedSettings`, acknowledge SETTINGS, flush
pending TLS tickets or commit ticket metadata. The definitive wire SETTINGS still
does that existing work; ticket metadata records the exact wire frame rather
than its effective overlay. One QPACK owner retains compatible table entries and
does not repeat an unchanged capacity instruction. The shared validator rejects
reserved HTTP/2 SETTINGS identifiers 2 through 5 for both initial input and the
actual control stream. This corrects a pinned upstream parser omission and closes
the wire connection with H3_SETTINGS_ERROR, as required by RFC 9114 section
7.2.4.1 and the pinned QUICHE OnSetting cases 1302–1315. Unrelated unknown
identifiers remain accepted.

`Transport.PrepareHeaders` runs on a private request/header clone after normal
TLS, initial-input and initial-peer-settings gates, before extended-CONNECT
admission, request stream allocation, QPACK encoding or body reads. Only Header
and a Context derived from the request context may change. Normalized early
GET/HEAD attempts run preparation without waiting for fresh peer settings; they
do not gain a fresh-ALPS wait or replay. A later ordinary request initializes
authenticated input. Low-level `OpenRequestStream` shares settings initialization
while `SendRequestHeader` retains explicit caller header ownership.

`IsEarlyRequest` reads the frozen preparation-path decision from this private
request context. It reports that authenticated initialization was skipped, not
that packets used 0-RTT. Derived contexts and `Response.Request` retain the marker;
ordinary authenticated preparation explicitly clears an inherited early marker.
With either hook configured, magic methods with an already-complete handshake
initialize normally and report false. Nil hooks retain the existing native early
fast path and buffered-input bypass. A genuinely pending early request stays true
if TLS completes during preparation. This lets a live connection policy remain available for warm
requests after caller cache clearing. Rejection replay normalizes the method and
gets a fresh false marker after its replacement generation initializes. Unmarked
caller requests and nil return false.

`request_preparation.go` shares ordinary header validation with the Transport and
rejects callback-added invalid fields before stream allocation. Local errors use
`RequestPreparationError` and cannot trigger protocol fallback, connection retry
or early-rejection replay. Optional-hook paths share an idempotent body-close
owner across direct/pooled/cached clients and rejection handling; genuine
unopened failures (including GOAWAY and early rejection) retain the body for the
existing retry. Initial cache misses keep caller body ownership. Prepared
responses keep the derived context in `Response.Request`. Nil callbacks preserve existing wire
behavior. Browser values, ALPS framing/origin policy and hint selection remain
outside the engine.

The hooks are minimal additions to `Transport`, `ClientConn`, the control-stream
SETTINGS path, `frames.go`, `client_session_settings.go`, the existing common
constructor and early-restart terminal-error check. Generic implementation and
tests stay in the two owned files. The upstream pin remains the full integration
recorded in `../upstream.json`; these are fork extensions, not upstream backports.

Primary protocol/source references:
[RFC 9114 SETTINGS](https://www.rfc-editor.org/rfc/rfc9114#section-7.2.4),
[RFC 9204 peer limits](https://www.rfc-editor.org/rfc/rfc9204#section-3.2.3),
[QUICHE initial versus wire SETTINGS](https://github.com/google/quiche/blob/132bfd90ea370faeacee616c70d10210bb12ba66/quiche/quic/core/http/quic_spdy_session.cc#L1090).

## Computed request Content-Length

[request_writer.go](request_writer.go) places a computed Content-Length in an
explicit `Header-Order:` slot instead of appending it after ordinary fields.
It uses the sanitized body length, removes supplied length aliases from a private
ordered snapshot and emits the field once. Supplied numeric values never override
framing. Without an explicit slot, implicit fields retain upstream insertion order.

A nil or empty Content-Length value slice suppresses automatic CL:0 only for an
actual zero-length body. Matching is case-insensitive. Positive and unknown
lengths ignore this opt-out and retain normal framing; an empty string value is
not an empty slice. A known zero-length upload with non-nil `Request.GetBody`
can select CL:0 through an explicit length slot, including on methods that normally
omit it. The nil/empty opt-out still takes precedence. Header preparation neither
reads the body nor calls GetBody. Requests without these explicit controls retain
upstream method and framing defaults.

The change is limited to request-header preparation and private length helpers;
it needs no additional public API from the HTTP module. It applies to both static
and connection QPACK encoding, preserves header validation and leaves caller data
unchanged. Compare it with the [pinned upstream request writer](https://github.com/quic-go/quic-go/blob/v0.65.0/http3/request_writer.go)
when updating. Browser constants and profile policy remain in Surf.

## Surf API usage

Optional `WireConfig` supplies reserved control/request frame factories and
independent initial `PRIORITY_UPDATE` emission from explicit Priority headers.
Control frames use one bounded writer; a blocked control stream cannot stall
request header preparation indefinitely. Local factory errors return immediately
without waiting for a peer response. Closed critical streams close the connection.
Qlog records the added frames without changing their bytes.

The shutdown integration tests wait for handler admission before initiating
graceful shutdown. Deadline tests assert minimum lifetime and bound completion
with a client timeout, instead of treating TLS startup and scheduler latency as
part of a 25ms shutdown timing tolerance. Preserve this synchronization when
merging upstream tests.

`QPACKConfig.SplitCookies` and `NeverIndex` control request field representations.
Nil configuration retains the official static encoder. Optional client encoding
uses the isolated `internal/qpack/encoder.go` with the official static table and
Huffman codec. No new QPACK or TLS dependency fork is required. All browser values,
random distributions and policy selection remain in Surf. The client adapter owns
ordered bounded instruction queues, negotiates peer capacity/blocking limits and
passes decoder feedback to the connection encoder. Headers and trailers share it.

`tests/http3/wire_config_test.go` covers independent GREASE/priority settings,
real stream IDs across requests, bounded queues, startup errors and critical-stream
closure. Private encoding tests cover all static N-bit forms, cookie order, trace
parity and request factory failures. Preserve these alongside QPACK startup tests.

Surf uses `Transport` with `TLSClientConfig`, `QUICConfig`, `AdditionalSettings`,
`MaxResponseHeaderBytes`, optional `QPACKConfig` and `Dial`, plus `RoundTrip`, `CloseIdleConnections` and
`Close`. Its tests also use `Server`, `Serve` and `Close`.

Surf does not use capsule or `HTTPStreamer` APIs. The old `ParseCapsule` API was
upstream behavior rather than a fork extension and has no compatibility wrapper.
The fork follows the current `NewCapsuleParser` / `CapsuleReader` API,
`HTTPStreamer.HTTPStream` signature and h3-events-14 qlog schema.

## Regression coverage

- `request_writer_order_test.go` and `fork_wire_test.go`: default and custom
  regular/pseudo-header order, CONNECT, Extended CONNECT, QPACK bytes with and
  without qlog, repeated requests, multiple values and omitted metadata.
- [request_writer_content_length_test.go](request_writer_content_length_test.go)
  and [request_writer_empty_body_test.go](request_writer_empty_body_test.go):
  independently decoded QPACK bytes for ordered computed lengths, nil/empty
  zero-length suppression, explicit empty uploads, positive/unknown lengths,
  stale values and case aliases. Both encoder paths, implicit-field order,
  request immutability, body isolation and unchanged validation are covered.
- The HTTP module's `tests/headerorder/`: shared ordering, partial orders,
  lexicographic fallback, excluded metadata and independent calls.
- `frames_settings_test.go` and `fork_wire_test.go`: ordered settings on real
  client/server streams, explicit profiles, omitted/zero/positive header limits,
  peer-supplied zero, qlog snapshots, local limits and 0-RTT compatibility.
- Upstream HTTP/3 and qlog suites, with ordered-setting fixtures and detailed
  error expectations adapted to the fork.

- `../tests/qpack/` and `../tests/http3/qpack_test.go`: RFC 9204 dynamic
  receive tables, all field representations, blocking, eviction, feedback,
  cancellation/deadlines, bounded decoding and real client/server requests.
  Outgoing dynamic encoding adds independent wire vectors, both Base forms,
  bounded references, ACK/cancel/increment state, draining, rollback and fuzz cases.
- `../tests/http3/qpack_encoder_test.go`: real outgoing dynamic requests,
  parallel bodies/trailers, zero-capacity fallback and invalid decoder feedback.
- `../tests/http3/qpack_client_test.go`: configurable stream reservation order,
  eager/lazy type bytes, immediate/deferred capacity announcements, zero/omitted
  peer limits, backpressure, immutable snapshots and critical stream errors.

Run validation from the parent QUIC module, then from the sibling `surf` module
against the published version, or using a temporary Go workspace during development:

```sh
go test -race ./...
go vet ./...
```

## Request body and live priority controls

`WireConfig.RequestBodyBufferSize` and `RequestBodyBufferSizePolicy` select
request body reads; `MaxDataFrameSize` caps DATA payloads without controlling QUIC
packetization. Zero preserves native sizing/framing. Size validation runs before
HEADERS/QPACK state changes. `Stream.Write` retains unfinished-frame handling;
`TryWriteAll` queues all capped frames atomically or leaves its state unchanged.
Browser sizing policy stays in Surf.

`RequestBodyFirstReadSizePolicy` receives a private copy of the validated wire
fields and the effective buffer size before connection QPACK state changes.
It can limit the first nonempty read independently of subsequent reads. Empty
reads retain the limit; short reads are passed through without coalescing or
waiting for more producer data. Invalid results fail request preparation.
The upload reader is closed on stream cancellation even when response HEADERS
have already arrived, so a blocking streaming source can release its goroutine.

`WithPriorityUpdateHandler` publishes a concurrency-safe per-attempt updater after
HEADERS. `RequestStream.UpdatePriority` exposes it for low-level users. Updates
share the bounded connection control writer, suppress repeated values only after
successful queue admission, support a default-priority reset, and stop on response
completion, cancellation or connection closure. `CompactPriorityUpdates` controls
serialization only. Local `SetPriority` scheduling remains independent. Nil wire
configuration now installs the bounded writer to allow explicit live updates;
no extra control frames are emitted without a caller request.

`tests/http3/request_body_test.go` covers body reads, framing, unknown lengths and
policy failures. Private stream tests cover every partial header/payload deadline
and atomic retry. Priority tests cover real control frames, deduplication, reset,
cancellation and deterministic queue backpressure. Surf additionally verifies
browser-profile sizing, TLS 0-RTT with standard TLS, GOAWAY and changed SETTINGS
across fresh QPACK connections. No browser values or TLS fork are added here.

## Initial control-stream staging

`WireConfig.CoalesceInitialSettings` optionally creates the client from a per-dial
`OnApplicationReady` callback before the QUIC handshake channel is released. With
restored early keys it also prepares that same owner from `OnEarlyDataReady`,
before Initial packetization. `FlushEarlySettings` selects the alternative early
flush boundary after sendable Initial data. The
caller configuration is cloned and its callback is chained. `sync.Once` shares
exactly one client between this callback and `Dial` / `DialEarly` completion.
The constructor reserves its control/QPACK streams once and tries a nonblocking
SETTINGS+reserved-frame write. Insufficient flow-control credit hands the same
stream and bytes to an asynchronous writer; cancellation releases that writer.
Eager QPACK type bytes are staged in their configured order before the readiness
channel wakes either instruction writer. A staged prefix is not sent twice;
missing stream credit retains the existing asynchronous owner.
Failures are published through the existing readiness channel. Nil/default wire
configuration retains asynchronous startup. Custom dialers and early-data
acceptance/rejection are covered alongside native startup tests.

HTTP/3 settings can share the client's first post-handshake flight; the option
cannot guarantee NIC batching, a fixed datagram size, identical handshake ACK
scheduling or shared congestion state with another implementation.

## Ready QPACK instructions with request headers

`WireConfig.CoalesceRequestHeaders` also stages already queued encoder
instructions with HEADERS and optional FIN. `qpack_batch.go` borrows an idle
encoder writer before header encoding and submits the ready queue snapshot in
the existing `Conn.TryWriteStreamBatch` operation. The snapshot is bounded by
the encoder queue; concurrent producers cannot extend the drain indefinitely.
Stream type and encoder instruction order remain unchanged.

The asynchronous writer claims ownership before dequeuing instructions. It
first tries `SendStream.TryWriteAll`, including after a previously staged type
or capacity prefix. A blocking `Write` would otherwise wait for that reserved
prefix to be packetized, keeping the writer busy through the request's send
opportunity. Busy writers, missing stream credit and flow-control backpressure
retain their asynchronous path; requests never wait for the encoder writer.

After a partially accepted batch, only unaccepted encoder bytes remain pending.
Accepted prefixes are never replayed. Encoder ownership is released before a
request's blocking fallback write, including cancellation and validation errors.
Priority updates retain the same accepted-prefix accounting. Nil/default wire
configuration retains the upstream request path; QPACK itself remains opt-in.
When a control/encoder element blocks before the request is attempted, the
request is retried independently with its own credit and FIN. Critical-stream
backpressure therefore cannot force an otherwise sendable bodyless request
through a separate blocking write/close sequence.

`qpack_batch_test.go` covers real packet grouping, FIN, busy writers, validation,
encoder/request/control backpressure, batch-size limits, partial acceptance,
concurrent writers and cancellation. `../tests/http3/qpack_encoder_test.go`
compares dynamic decoding and feedback with batching enabled and disabled,
including concurrent POST bodies and trailers. Browser values and policy remain
in Surf. This operation does not prescribe a stream order or packet count.

## Per-connection reserved SETTINGS

`WireConfig.SettingsGrease` generates one reserved SETTINGS identifier/value when
initial control bytes are prepared. It runs once per client connection, including
staged startup, rather than once per request or transport. The factory must be
concurrency safe, nonblocking and must not reenter the transport. Its identifier
must follow `0x1f * N + 0x21`; both fields must fit a QUIC varint. Invalid values,
collisions with configured settings and factory errors fail preparation before
writing the control prefix. Valid output is appended to a private ordered snapshot;
configured SETTINGS and QPACK state are unchanged. Nil retains native behavior.
The engine contains no browser random distribution or profile defaults.

Private validation and real HTTP/3 tests cover bounds, duplicates, errors, input
immutability, request reuse and fresh connections. Surf supplies Chromium's two
independent uint32 draws through its fluent `GreaseFactory` profile control.

## Response header observation

`WireConfig.OnResponseHeaders` observes parsed informational and final headers
before tracing or application delivery. Returning an error cancels the request.
The nil default preserves upstream behavior. Browser retry/backup policy belongs
in Surf, not this generic hook. The patch is limited to `wire_config.go` and the
response loop in `client.go`; `response_headers_test.go` exercises real QUIC
streams, ordering and cancellation. The adjacent trace call now propagates
`Got1xxResponse` errors as required by the existing httptrace contract.

## Same-connection early-data restart

`Transport.EarlyDataRestart` wraps pooled clients with successive HTTP/3
protocol generations of one QUIC connection. Synchronous rejection retires the
old control/QPACK workers; verified application readiness installs the fresh
state exactly once. Pending requests restart only after rejection, honoring
cancellation and body replayability. Unconsumed bodies can be reused; consumed
bodies need `GetBody`. Explicit low-level client connections keep their existing
application-controlled behavior. The generic default remains disabled.

The narrow engine hooks and HTTP/3 generation cancellation are kept separate
from Firefox profile policy. Regressions cover same-connection wire operation,
unknown tickets, HRR, changed SETTINGS, parallel methods, cancellation, TLS
verification failure, explicit opt-out and consumed-body replay limits.

## Optional connection reuse and control extensions

`Transport.ConnectionReusePolicy` can authorize a live, confirmed native client
for another authority. Certificate, route and privacy policy belong to the
caller. Candidates are pinned outside pool locks, checked again after callbacks,
and keep one pool owner; no alias entries duplicate closure or QPACK state.
All candidates receive a no-network pass before any network-enabled pass.
`ConnectionPoolPartition` separates exact and cross-authority lookups, always
retaining the authority in the actual key. Nil callbacks preserve native pooling.
Preconnect shares the policy; abandoning a borrowed preconnect cannot retire its
owner. Idle closure preserves active response streams. Request-local failures
cannot evict a healthy shared owner.

`RoundTripCached` performs an initial cache-only probe. Its exact ErrNoCachedConn
leaves the body caller-owned; after an attempt starts, ordinary retry ownership
applies. Existing RoundTripOpt behavior is unchanged. A valid response header,
including 1xx, commits the attempt; callback errors wrapping REQUEST_REJECTED or
a subsequent reset cannot authorize an internal replay.

`WireConfig.OnControlFrame` observes otherwise unknown server control extensions
after SETTINGS through a bounded reader. Unread bytes are drained; truncation is
an error. The optional handler enforces SETTINGS-first framing. Nil keeps native
unknown-frame handling. This engine contains no ORIGIN, NOIP or browser policy.
Surf supplies those through its fluent public builder.

The narrow implementation lives in `connection_reuse.go`, `control_frame.go`
and `response_progress.go`, with small transport/parser hooks. Real QUIC tests
cover ownership, partitions, cancellation, closure, GOAWAY, active bodies,
preconnect, callback errors, informational responses and truncated extensions.

## Connection readiness and retirement

`HasCachedConnection` inspects an exact partition's confirmed, live connection
without DNS, dialing or reuse callbacks. `RoundTripCachedReady` additionally
permits the normal authorized cross-authority lookup, but never joins a pending
dial or an early-only connection. An exact cache miss leaves the request body
caller-owned. Existing `RoundTripCached` keeps its pending-dial behavior.

`PreconnectWithOptions` adds a per-waiter connection-start observer and selects
early or authenticated readiness. Start notification belongs to the actual
shared pool dial, so later waiters observe a start already reported. The default
dialer reports after DNS; custom dialers call `NotifyConnectionStarted` after
DNS/socket preparation and before QUIC dialing. Callbacks run outside pool
locks. Failure, cancellation and shared ownership follow existing preconnect
rules; no browser timer or policy value is embedded here.

A normal lookup retires a drained or closed exact owner before preparing its
replacement. Drained live owners remain transport-owned while their response
streams are active. Idle cleanup preserves those streams; full Close closes
retired owners too. Ended owners are removed by context completion callbacks.
Temporary early-data protocol restarts are not GOAWAY. Keep the real-connection
readiness, partition, shared-start, cancellation, active-body and closure tests
when updating upstream.
