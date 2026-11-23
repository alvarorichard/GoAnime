# Upstream maintenance

The supported baseline is Go 1.27+. Exact versions and commits live in
[upstream.json](upstream.json). The current sources are Go 1.27.4, x/net v0.62.0
(public HTTP/2 adapters). HTTP/3 is maintained separately in
`github.com/enetx/quic-go/http3`.

## Required maintenance policy

Ease of future upstream updates is a design requirement. Preserve upstream
package/file responsibilities, accept compatible structural refactors and keep
fork mechanisms in small owned files with minimal hooks in upstream sources.
Do not maintain obsolete parallel upstream implementations where a small adapter
preserves the required behavior. Keep an explicit inventory of each local
mechanism, hook and reason. Remove superseded adapters only after protocol,
concurrency, public-API and Surf fingerprint/header-order regressions pass.
Browser values and policies belong in Surf; native engine defaults stay intact.

## Required coverage of fork changes

All code changes owned by this fork must have **100% test coverage**, including
existing extensions, new files and changed hooks in upstream sources. Measure
executable additions/modifications against the exact pinned upstream revision;
an overall upstream percentage is not an acceptance criterion. Cover changed
statements and each reachable branch, error, boundary, reset/reuse and
cancellation outcome with meaningful assertions. Go statement coverage alone
must not be presented as proof of branch or protocol-state coverage.

Execute platform-specific behavior on the appropriate OS and combine coverage;
cross-compilation alone is not execution. Types/constants and public API changes
without instrumented statements still need appropriate API/compile/wire checks.
Retain independent browser-source and wire evidence, with actual histories and
normalization limits recorded. Do not exclude owned code, weaken assertions or
change goldens without evidence to obtain the number. Remove obsolete code only
after proving required behavior is preserved. Complete missing coverage before
publishing a new code release; report unmet targets and measurement limits
explicitly rather than claiming 100% prematurely.

## Audited fork behavior

Before migration, all source files and repository histories were compared with
Go 1.26.0 and x/net v0.50.0. The original repositories were http `ef38f4c`, http2
`7d3b7bd`, and http3 `ba6a502`. The old HTTP/2 and HTTP/3 repositories are archived
under the workspace's archive directory as source snapshots.

The retained behavior is covered by protocol and Surf regression tests:

- `Header-Order:`, `PHeader-Order:` and `Header-Case:` are metadata, bypass normal field validation
  where needed and never appear on the wire or in header-size accounting.
- One stateless sorter in `internal/headerorder` handles HTTP/1 and HTTP/2.
  Its `headerorder` public adapter is also used by the separate QUIC module.
  Listed headers precede unlisted ones; unlisted headers sort lexicographically.
  Multiple values keep their order. There are no shared mutable exclusion maps,
  retained pool order or global header locks.
- HTTP/1 preserves canonical wire names and ordered Host, User-Agent,
  computed request Content-Length and chunked Transfer-Encoding.
  Response header ordering remains opt-in through `EnableHeaderOrder`.
- Cookies preserve embedded double quotes. Current cookie-count limits remain
  available through the small standalone GODEBUG reader.
- HTTP/2 preserves the exact SETTINGS list/order, initial stream ID, connection
  WINDOW_UPDATE increment, PRIORITY prelude and HEADERS priority parameters.
  Actual receive windows/decoder limits agree with the advertised settings.
- HTTP/2 first-request lifecycle checks account for the configured stream ID
  and PRIORITY prelude. Error GOAWAY frames do not retry that first request;
  graceful GOAWAY keeps upstream retry behavior. Single-use connections accept
  their first request even when its stream ID is greater than 1.


The old exported `H2transport` assignment is replaced by native registration and
connection management. Public HTTP/2 frame/error types alias the native engine,
so framing and error identity are shared instead of copied.

## Maintained HTTP/2 adapters

The fork maintains `http2.Transport`, `Server`, `ClientConn` and the `h2c` handler as
supported adapters over the native engine. Upstream x/net deprecated these in
favor of the standard library; the fork retains their configuration API for Surf,
including fingerprint controls and HTTP/1.1 Upgrade to h2c. Those upstream
replacement notices are removed from the supported adapter APIs. Warnings remain
for options that do not work (`Server.MaxHandlers` and `RoundTripOpt.OnlyCachedConn`
without a custom pool). Standard-library API deprecations remain unchanged.

Wire tests under `tests/h2c` cover ordinary HTTP/1, prior knowledge, Upgrade with
a request body, and an additional request over the upgraded connection.

## Keep modifications isolated

| Extension | Files |
| --- | --- |
| Shared sorting and metadata | `internal/headerorder` |
| HTTP/1 ordering glue | `header_order.go`; small hooks in `header.go`, `request.go`, `transfer.go`, `server.go` |
| Computed request Content-Length | `header_order.go`, `request.go`, `transfer.go`, `internal/httpcommon/header_order.go`; narrow codec and HTTP/2 transport hooks |
| HTTP/2 wire configuration | `http2_fingerprint.go`, `internal/http2/fingerprint.go`, `http2/fingerprint.go` |
| HTTP/2 initial peer settings | `internal/http2/peer_settings.go`; connection initialization and shared wire-SETTINGS hooks in `transport.go`; existing fingerprint/public transport adapters |
| HTTP/2 peer-error policy | `internal/http2/connection_error_policy.go`; two read-loop hooks and public fingerprint/transport adapters |
| HTTP/2 request header policy | `internal/http2/header_policy.go`, `hpack_literal.go`, `indexed_fields.go`; narrow hooks in `transport.go` |
| HTTP/2 receive-credit batching and initial per-stream credit increase | `internal/http2/receive_window.go`; hooks in `flow.go`, `fingerprint.go`, `transport.go`; public fingerprint aliases |
| Shared HTTP/2 header codec hooks | `internal/httpcommon/header_order.go` |
| Public HTTP/2 frame aliases | `http2/frame_aliases.go`, generated by `http2/gen_aliases.go` |
| Standalone network tracing | `httptrace/trace.go` bridges net hooks to the public standard-library httptrace API |
| Runtime-independent settings | `internal/godebug` |

Do not reintroduce the old generated `h2_bundle.go`, a second HTTP/2 engine,
version-switch implementations for Go <1.27, private Go runtime hooks, deprecated
scheduler no-ops, or unrelated profiling/toolchain/proxy packages. Keep the Go
source files needed by Surf and the HTTP/2 dependency closure. Server,
`httptest` and `cookiejar` are required by the active tests and examples.

## HTTP/2 initial peer settings

`HTTP2Fingerprint.InitialPeerSettings` (also available on `http2.Transport`)
receives the selected connection once, after HPACK and outbound flow control
exist and before the preface, read loop or requests. Callers supply already
authenticated peer instructions; the HTTP engine contains no TLS decoder,
browser values or browser policy. The callback must return promptly, support
concurrent connections and avoid connection I/O or transport reentry. Its result
is copied and each setting uses the same validation/update path as wire SETTINGS.
An error closes the selected connection before any HTTP/2 output. Nil keeps the
native initialization path.

Initial input neither emits an ACK nor marks wire SETTINGS received. The first
wire frame remains required, including the existing extended-CONNECT wait.
Explicit initial MAX_CONCURRENT_STREAMS survives omission from that frame;
subsequent wire values replace it and wake queued requests, including a zero
limit raised to a positive value. Header-table updates retain the existing
optional repeated-update policy. An explicit peer MAX_HEADER_LIST_SIZE of zero
rejects new requests before HPACK/body output, while empty trailers retain their
normal END_STREAM behavior. The generic shared header codec's zero-value API is
unchanged. Unknown valid settings remain ignored as on the native wire path.

The mechanism is an additive fork extension against the unchanged exact Go and
x/net revisions in `upstream.json`, not an upstream source merge or an ALPS/TLS
implementation. Keep the private state/wire tests and independent public-adapter
HPACK/DATA/error tests when updating the native settings switch. Root fingerprint
and public frame types already alias this engine, so no bundled engine or new
alias generator is needed.

## HTTP/2 receive-credit policy

An installed `ReceiveWindowUpdatePolicy` owns the decision to return accumulated
credit for its connection or stream. Returning false defers WINDOW_UPDATE even
when the peer's available window is zero. `Available` describes remaining peer
credit, not unread response-buffer bytes: the read loop can accept the full
advertised window before the application has drained that data. `Pending` is
consumed credit that has not yet been returned to the peer. Deferred callbacks
do not reset the elapsed-time baseline; only an emitted update does.

The callback must eventually release enough pending credit for progress. An
always-false policy deliberately withholds credit and can stall the peer after
the read buffer is drained; emergency release, if wanted, belongs in the policy.
The engine does not override false with an implicit zero-window exception.
Nil policy keeps the upstream batching thresholds and immediate exhausted-credit
recovery unchanged. Maximum-window and over-credit checks remain independent
of the policy. No browser threshold or timing value is selected here.

`internal/http2/receive_window_test.go` covers strict threshold boundaries with
an exhausted window and buffered data, repeated refresh cycles, explicit policy
authority, unchanged deferred timestamps, optional emergency decisions, native
nil-policy recovery and overflow protection. Surf owns browser-source comparisons
and actual WINDOW_UPDATE wire tests.

## Request Content-Length controls

An explicit `content-length` position in `HeaderOrderKey` places the computed
request length in that position when framing requires the field. Supplied header
values never replace the actual body length. Field names and order slots match
case-insensitively; an ordered field is emitted once. A slot alone does not add
CL:0 to an ordinary empty GET or HEAD.

A nil or empty Content-Length value slice suppresses automatic CL:0 only when
the sanitized outgoing body length is exactly zero. Positive and unknown lengths
ignore this opt-out and retain their normal framing. A slice containing one empty
string is not the opt-out. For a known zero-length upload, a non-nil
`Request.GetBody` plus an explicit length slot can select CL:0 on methods that
otherwise omit it; header preparation does not call GetBody. The nil/empty
opt-out still takes precedence. With neither an explicit slot nor a nil/empty
slice, upstream method and framing defaults remain unchanged.

The request-only HTTP/1 path in [request.go](request.go) passes a negative length
sentinel when [transfer.go](transfer.go) decides not to send the field.
[header_order.go](header_order.go) uses private copies for computed framing,
including ordered Host and chunked encoding. The zero-length opt-out does not
change response encoding or the explicit length passed to `Header.Write`.
HTTP/2 carries GetBody presence through
[internal/http2/transport.go](internal/http2/transport.go) into the shared
[request codec](internal/httpcommon/httpcommon.go); its
[ordering helper](internal/httpcommon/header_order.go) applies the same rules to
both ordered synthesis and the implicit-field fallback. Keep these small hooks
when comparing the request framing code with the sources pinned in
[upstream.json](upstream.json). Browser values and profile policy stay in Surf.

## Update procedure

The Go 1.27.4 update changes only `server.go` in the retained source tree.
It imports the upstream fix for concurrent `MaxBytesReader` limit handling and
response-header writes (CVE-2026-97054). The body reader records an atomic flag;
the response writer owns header changes and connection-reuse decisions.
`tests/http1/max_bytes_reader_test.go` adapts the upstream public-API regression
and checks that an over-limit HTTP/1 connection is not reused. HTTP/2 remains
on its existing source baseline; Go 1.27.4 changes no native HTTP/2 files.

1. Run `python3 scripts/upstream.py audit` to review the local delta against
   the exact pinned source. Import grouping and gofmt-only changes are ignored.
2. Run `python3 scripts/upstream.py prepare --go VERSION --net VERSION`.
   The helper normalizes package paths, does three-way merges against the pinned
   baseline, discovers new files in retained packages and reports conflicts.
   It writes a separate candidate; the working tree remains unchanged. Missing
   upstream files are reported for manual review, never silently deleted.
3. In the candidate, resolve conflicts and inspect `summary.json` and `changes.diff`.
   Review any newly imported internal packages or API changes. The helper updates
   `upstream.json` and the primary module versions in the candidate.
4. Run `go mod tidy`, `go generate ./http2`, `go test -race ./...` and `go vet ./...`.
   Also run `python3 -m unittest discover -s tests/upstream`.
   Regenerate aliases twice and confirm the second pass has no changes.
5. Point Surf's local replace at the candidate and run its complete test suite
   with the race detector. Preserve the browser wire golden files unless the
   browser profiles themselves intentionally change.
6. Transfer reviewed changes, including updated pins and sums. Keep the local
   Surf replacement until the combined http module is released; then update the
   required version and remove the replacement for publication.

`--go-source` and `--net-source` reuse already downloaded pinned
sources. A supplied Go checkout is checked against its version and pinned commit.
Prepare was verified against identical versions: zero changes, zero conflicts,
with every retained custom source file unchanged.

## Regression suites

Request compression continues to use official `golang.org/x/net/http2/hpack`.
The header policy adapter selects literal representations, Huffman mode and
name references. Its optional bounded membership table records only entries the
official encoder actually inserts, following its capacity updates and eviction.
This permits an insertion-size limit without losing reusable exact entries.
`HPACKNameReferenceFirst` selects the first static name without a per-connection
shadow table. `HPACKIncrementalIndexing` retains exact references and emits an
incremental literal even for an oversized entry. The adapter then empties the
official encoder through its public capacity API, discarding only the private
reset's synthetic updates. Real pending peer updates and the encoder's configured
capacity limit are retained. No codec internals or additional fork are required.
Do not copy an HPACK codec or browser-specific policy into this module.

`tests/http2/header_policy_test.go` exercises multiple requests on a real
connection, sensitive and non-indexed fields, per-request priorities, open-stream
dependencies, trailers and invalid callbacks. Private codec tests cover multiple
table-size updates, eviction, static and dynamic name references and Huffman modes.

`tests/http1/fork_headers_test.go` exercises HTTP/1 wire order, generated length/encoding,
response opt-in, input immutability, independent orders and embedded cookie quotes.
[tests/http1/request_content_length_test.go](tests/http1/request_content_length_test.go)
and [internal/http2/request_content_length_test.go](internal/http2/request_content_length_test.go)
cover nil/empty controls, replayable empty uploads, computed positive lengths,
unknown streams, stale header values, case aliases, absent/unlisted/ordered slots
and unchanged inputs. HTTP/1 tests parse the resulting request bytes and separately
check unchanged response and `Header.Write` behavior; HTTP/2 tests decode HPACK.
`tests/http2/fingerprint_test.go` checks actual preface/settings/priority/window/header
bytes through both the public adapter and native `HTTP2Config`, decoder limits,
invalid configuration and unknown frame gaps. Shared sorting, size accounting
and trace metadata have direct tests in `tests/headerorder` and
`tests/headercodec`. Current upstream native HTTP/2 tests are included, together
with portable HTTP, `cookiejar`, `httptest` and dump tests. HTTP/3 and qlog tests
live in the separate QUIC module.
Surf additionally validates browser wire profiles, h2c, proxies, TLS, sessions,
HTTP/3 and its other client features.

## Test layout

Public-API suites live under `tests/`. Private codec/state-machine tests stay in
`internal/http2` and their source packages because Go requires same-package
access and `export_test.go` hooks. No test hooks are compiled into production to
force a directory move.

In the root package, `upstream_cookie_test.go` checks private
sanitizers and error identities; `upstream_mapping_test.go`, `upstream_pattern_test.go`,
`upstream_routing_index_test.go` and `upstream_routing_tree_test.go` inspect private
routing structures. They remain in the root, while request parser tests and their
reflection helper live in `tests/http1/` because they use only public API.

`upstream.json` records test relocations so future merges follow the original
upstream files even when their tests moved.

## HTTP/3 ownership

HTTP/3 source, tests, qlog support and QUIC dependencies were migrated to
`github.com/enetx/quic-go`; the duplicate `http3/`, `tests/http3/` and
`tests/http3_qlog/` directories have been removed. The update helper only handles
Go and x/net and cannot reintroduce the old package through a QUIC update.

The Go 1.27 HTTP source still contains protocol-selection and registration hooks
for external HTTP/3 implementations. These are upstream HTTP interfaces, not a
second QUIC engine, and introduce no QUIC dependency. Keeping them avoids an
unnecessary patch against the standard-library baseline.

`headerorder` exports the existing internal sorter for the separate QUIC module.
Keep its adapter small and preserve the shared wire-order tests when updating.

## Connection preparation

`transport_preconnect.go` adds `Transport.Preconnect(*Request)` as an isolated
adapter around the native connection acquisition and idle-pool return paths.
It performs no application request or body access. HTTP/2 initialization and proxy
handshakes follow the normal transport. `http2/transport_preconnect.go` supplies
the existing wrapper's dial context before delegating, preserving custom TLS,
SETTINGS and native pooling without modifying upstream transport files.

Preparation does not reserve the connection; subsequent RoundTrip still applies
native pool limits and retry rules. Keep-alives are required, and custom HTTP/2
ConnPool implementations are unsupported. Cancellation retains native shared-dial
semantics: detached TCP dials may finish and become idle. CloseIdleConnections
also cancels abandoned native dials. Public tests cover body isolation, HTTP/1
and HTTP/2 reuse, validation and cancellation.

`transport_connection_group.go` adds `Transport.HasConnectionGroup(*Request)`
as a read-only snapshot of native TCP pool ownership. It includes occupied H1
connections, pooled native connections and pending dials; it does not promise
an idle/reusable socket or reserve one. It uses the existing tracked read-loop
owners and native dial queue, with no additional lifecycle hooks or counters.
Route selection uses the normal connect-method key, including proxy and
HTTP/1-only ownership. No application body or connection trace is touched.
Reusable HTTP/2 sessions retain their separate `HasHTTP2Connection` observer.
Public tests cover cold/pending/occupied/idle/failed/closed states, route
isolation, concurrent observation and non-mutating body/trace ownership.
The new observer has 100% measured statement coverage. The owned-change audit
now covers 1100/1100 statements, combining the unchanged v1.0.47 source's
previous complete profile with the current full race profile; pre-existing Go
source files have identical contents and coverage locations. Full macOS/Linux
race suites and vet pass. The observed group still does not reserve a socket
or certify browser scheduling/packet-timing parity.


## Explicit HTTP/1 field spelling

`HeaderCaseKey` lists exact field spellings selected by applications; other
HTTP/1 fields retain canonical spelling. Invalid spellings are ignored and
inputs are not mutated. Browser choices remain in Surf. The shared
`headerorder.IsMetadataKey` classifier excludes all private controls from
sorting and validation in HTTP/1, HTTP/2 and external HTTP/3 adapters. Changes
stay in `header_order.go`, small header/transport hooks and the shared sorter.

## Protocol selection and request headers

`Transport.PrepareHeaders` is an optional callback on a private request, called
for the actual HTTP major version before transmission. Its helper is isolated
in `transport_header_policy.go`; one HTTP/1 hook and the HTTP/2 request adapter
cover fresh and cached connections. Header validation, body closure and stream
reservation cleanup remain native. Context metadata and the prepared response
request are retained for redirects; a nil callback keeps the original path.
The callback can abort preparation by assigning a canceled derived context to
its private request. The helper returns that context's exact cancellation cause
before header validation, stream allocation or body reads. Canceling the private
child does not cancel the caller or retire the selected reusable connection.
Keep the public wire, clone, streaming-body and invalid-header regressions when
merging upstream.

A linked `http2.Transport` with HTTP/1 explicitly enabled preserves a negotiated
HTTP/1 socket in the linked transport's normal pool. Its dial followers stop
coalescing with a non-multiplexed connection rather than waiting for that
request's response. HTTP/2-only configurations retain their existing path.
These hooks contain no browser settings or TLS implementation dependency.

## HTTP/2 peer-error shutdown policy

`HTTP2Fingerprint.ConnectionErrorPolicy` (also available on `http2.Transport`)
receives the underlying connection, accepted frame type, peer error code and
reset stream ID. Returning true immediately retires that connection from the
pool and reports the peer error to active requests, without sending GOAWAY in
reply. RST_STREAM for an inactive stream is ignored before policy evaluation.
Callbacks run on the reader without connection locks and must not block or
operate on the connection. Nil keeps the native handling, including GOAWAY
LastStreamID semantics and stream-local resets. No browser code, cache capacity,
origin key or retry decision lives in the engine.

`tests/http2/connection_error_policy_test.go` reads actual TCP frames through
both the public adapter and native transport. Its concurrent-stream matrix
checks unchanged nil-policy behavior, promoted resets/GOAWAY, error identity
and absence of an outgoing GOAWAY. Surf owns the Chrome-specific policy and
its TLS/proxy/header/body regressions.

## Existing HTTP/2 pool inspection

`Transport.HasHTTP2Connection(*Request)` and the public `http2.Transport`
adapter inspect the actual native pool for the same selected proxy route.
They send nothing, reserve no stream and never dial. Busy reusable sessions
remain visible, while GOAWAY, expired idle sessions, closed connections,
HTTP/1-only policy and unavailable proxy routes report false. Custom protocol
implementations and custom HTTP/2 pools are not introspected.

The implementation is isolated in `transport_pool*.go`,
`http2/transport_pool.go` and `internal/http2/connection_pool.go`; it requires no
upstream core hook. Preserve `nethttpomithttp2` build coverage. Real trusted TLS
regressions in `tests/transport_pool_test.go` cover both adapters, occupied
streams, wire GOAWAY, route isolation, cancellation and concurrent probes.

### Pool maintenance and network activity

`transport_connections.go` exposes concrete pooled HTTP/2 maintenance handles;
`transport_pool_http2.go` and `http2/transport_pool.go` share the native pool.
No browser policy is embedded. Retirement options optionally probe and/or send
NO_ERROR GOAWAY after drain; default retirement remains quiet. Protocol errors
retain their own close path. Socket I/O activity is isolated in
`internal/http2/connection_activity.go`; the real TLS net.Conn is unchanged.
`internal/http2/connection_pool.go` contains the maintenance implementation.
The only drain hooks in native transport are the two successful-idle close paths.

HTTP/1 liveness and retirement live in `transport_liveness.go`, with small
registration/read/write/retirement hooks in `transport.go`.
`TrackConnectionActivity`, copied by Clone, keeps streaming uploads on the counted
write path when enabled. Its default preserves the native ReaderFrom bypass.
The public liveness API requires enabling this field before first use. No global
transport defaults select network events or browser timeouts.

`tests/transport_connections_test.go`, `tests/transport_liveness_test.go` and
`tests/transport_retirement_wire_test.go` cover active response ownership, new
pool admission, upload/read progress, stale maintenance handles, shared fixed
probe payloads, and a wire barrier proving GOAWAY follows the active response.
`nethttpomithttp2` production builds and the applicable HTTP/1 liveness tests
remain supported; tests that explicitly request HTTP/2 require that engine.

## HTTP/1 write policy

`Transport.HTTP1WritePolicy` selects header flushing and bounded body reads from
actual serialized header bytes, content length and body classification. The
implementation lives in `transport_write_policy.go`, with small request,
transfer-writer and transport-context hooks. It never reads a body to classify
it, suppresses optional copy fast paths only when a policy requests bounded
reads, and cannot suppress mandatory streaming/Expect header flushes. Clone
retains the callback. A different transport shadows stale policy context rather
than inheriting another transport's settings. Nil retains the native fast path.
`tests/transport_write_policy_test.go` covers wire boundaries, ownership,
Expect accept/reject, short reads and cross-transport context reuse.

## HTTP/2 closing and upload policy

`internal/http2/connection_close_policy.go` owns the optional terminal GOAWAY
policy. Existing explicit/protocol-error frames are not duplicated; protocol
errors retain their required code. The callback may supply bounded debug data.
Shutdown remains bounded if the writer blocks. A nil callback preserves native
behavior. Keep the once/blocked-write/protocol/retirement tests when updating the
few close hooks in `internal/http2/transport.go`.

`internal/http2/request_body_policy.go` describes optional upload read buffers,
source-cycle boundaries, virtual header prefixes and DATA caps. `MaxReadSize`
caps every new application read independently, while `ReadBufferSize` models a
persistent source-buffer boundary. A short read consumes only the latter budget;
it never reduces a later `MaxReadSize` allowance. Combined options enforce both
limits. The engine
collects effective ordered fields during its existing header encoding, retains
flow-control/cancellation/body ownership, and never waits to fill a short read.
Configured reads are capped even when a pool returns a larger scratch buffer.
Source boundaries advance by transmitted bytes, and an optional cycle reset
follows an actual receive-credit wait. Nil/zero options retain native framing.
`Fingerprint.WriteBufferSize` selects the connection writer buffer separately.
Tests force source boundaries, peer credit exhaustion/resume, invalid policies
and unchanged native framing. No browser constants belong in these files.

`Fingerprint.HealthCheckPayload` optionally supplies opaque data for automatic
idle health PINGs. Clone copies the value. Explicit Ping and peer PING echoes
retain their original behavior; timeout scheduling remains the existing HTTP/2
configuration. `internal/http2/health_check_payload_test.go` verifies repeated
fixed probes, immutable snapshots and peer echoing.

The public `http2.Transport` adapter forwards these options to the same engine;
it introduces no second HTTP/2 implementation or generated browser bundle.

### Network trace composition

The standard-net adapter in `httptrace.WithClientTrace` bridges only newly
supplied network hooks before composing the fork's protocol trace. Re-bridging
an already composed trace would call application DNS/connect hooks twice each
time a transport added a GotConn-only layer. Capturing each original DNS function
also prevents a later composition from changing an earlier bridge. Public tests
in `tests/httptrace/composition_test.go` verify callback order, multiple protocol
layers and coexistence with an independently supplied standard-library trace.

### Removing adapter leftovers

Unused private compatibility members from the preceding adapter implementation
have been removed: the old HTTP/3 protocol setter bridge, server initialization
Once, unused connection-error aliases, RoundTripOpt.allowHTTP and a duplicate
no-cached-connection classifier. Public fields, methods and error markers stay
available. The public HTTP/2 adapter passes Staticcheck U1000. Unused declarations
in unchanged upstream engine/test files retain upstream shape for straightforward
future source synchronization; they are not Surf-specific compatibility layers.


### Blocked HTTP/1 upload cancellation

The error path in `persistConn.mapRoundTripError` closes the request body before
waiting for `writeLoopDone`, after the writer has taken that body-bearing
request. A small per-request start signal preserves an unstarted unread body
for safe connection retries without unnecessary `GetBody` calls; bodyless
requests allocate no extra signal. Closing only the socket cannot unblock an
application reader waiting for its next portion. The existing atomic `readTrackingBody.Close`
keeps this close idempotent. `tests/http1/streaming_cancel_test.go` covers caller
cancellation, client timeout and abrupt peer close, with known/unknown lengths
and exactly-once closure. `tests/http1/stale_body_test.go` separately verifies
an unread body can survive an idle-peer close without reopening. Preserve the
start signal, writer hook and error-path select together when syncing Go.


### Concurrent HTTP/2 handoffs with keep-alives disabled

An HTTP/2 pool admission reserves its selected session before retiring other idle
connections when `DisableKeepAlives` is set. Concurrent TLS handoffs can otherwise
leave a never-used single-use socket behind with no idle timer. Reservation and
active-stream checks in the existing `closeIfIdle` preserve running requests.
Keep the admission cleanup in both reservation paths of
`internal/http2/client_conn_pool.go`. The deterministic pool test retains active
and reserved sessions and ordinary keep-alive behavior; the upstream concurrent
no-keepalive test also runs repeatedly. This addresses the retained macOS CI
leak without relaxing or skipping the test.


### Retirement PING registration

`internal/http2/connection_ping.go` separates synchronous payload registration
from the existing asynchronous writer and acknowledgement wait. `Ping` retains
its blocking API; `Retire` registers or joins an outstanding payload before
returning, then waits asynchronously. This prevents an ACK from overtaking a
joining retirement goroutine and creating a duplicate probe. A later call after
ACK can still request a new PING; this is not a once-only retirement policy.
Preserve the small `ping` wrapper and `Retire` registration hook during updates.
A synchronized test checks immediate registration, shared ACK, active-response
safety and a later independent probe; the public wire retirement test runs
repeatedly.
