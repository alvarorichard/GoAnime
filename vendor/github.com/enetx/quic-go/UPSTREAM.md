# QUIC and HTTP/3 maintenance

Source: official [quic-go master](https://github.com/quic-go/quic-go/commit/b638c8b53731c23d9a5471ba9829fc84e93ed2ec),
commit `b638c8b53731c23d9a5471ba9829fc84e93ed2ec`. Go 1.27+.
The preceding release anchor is v0.65.0; the pinned source includes the next 55
commits with the fork adaptations described below.
Pins are recorded in [upstream.json](upstream.json).

The current owned-change measurement and the individually reviewed architectural
exceptions are recorded in [COVERAGE_AUDIT.md](COVERAGE_AUDIT.md). Raw coverage,
branch evidence and browser comparisons remain separate checks.

The module is `github.com/enetx/quic-go`. Imports are mechanically renamed so Surf
can depend on a published fork without requiring its users to add replacements.
The HTTP/3 package retains Surf's ordered SETTINGS and request headers, local
header limits and stream-priority adapters. HTTP types come from `enetx/http`.
The stateless header sorter is shared through `enetx/http/headerorder`; there is
no second implementation here.

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

The owner permits a narrowly documented exception for a noncritical path that
cannot be exercised through the supported architecture. Review its callers and
invariants first, and prefer a useful, small refactor when that makes behavior
independently testable without complicating future upstream integration. Keep
the uncovered statements in the raw report and record each accepted exception
and its consequences. Difficult tests, platform execution and critical protocol,
security, cancellation, ownership and data-integrity behavior are not exceptions.
Do not report raw 100% when exceptions remain.

Request replay after rejected early data has a separate private preparation
function in `http3/client_early_restart.go`, covering method normalization,
unopened body retention, `GetBody` failures and close ownership independently
of connection-generation scheduling. No public API or upstream file changes are
needed for this separation.

QPACK send streams share the private `rawConn.watchQPACKSendStream` mechanism.
It closes the connection after peer cancellation of either reserved or lazy
critical streams, including while a feedback writer is idle. See
[RFC 9204 section 4.2](https://www.rfc-editor.org/rfc/rfc9204.html#section-4.2).
Rejected early generations and connection shutdown keep their existing lifecycle.

## Full upstream integration

The [2026-11-15 review](UPSTREAM_AUDIT.md) inventories all 55 commits through
upstream master `b638c8b53731c23d9a5471ba9829fc84e93ed2ec`. Compatible source,
package moves, refactors and test harnesses are integrated. Repository-owned CI
replaces upstream service/publishing workflows; module/toolchain requirements
retain the fork's supported versions. `upstream.json` pins the integrated master,
not the previous release or an obsolete selective-backport list.

UDP implementation and pool ownership now follow `internal/udp`. Native
`reno_sender.go` and `interface.go` follow upstream without custom bodies.
`configured_sender.go` embeds native Reno state and owns configurable CUBIC,
PRR and pacing. `sender_events.go` supplies full event context only to that
controller, allowing the native compact interface to stay unchanged. The CRYPTO
manager solely owns Initial/Handshake streams and drops them with their keys;
owned layout/recovery hooks operate through this owner. The two-level frame
sorter and deterministic synctest harness are adopted.

Server ReadHeaderTimeout covers blocked dynamic QPACK decoding as well as raw
header reads. Successful headers clear the deadline before body and trailer
processing. Qlog emits the new implementation field and retains code_version
for older trace consumers. UDP constants/interfaces and Simnet.Start retain
source compatibility through small adapters.

## Ownership

The fork contains protocol controls, no browser names, presets or browser-valued
defaults. All profiles, template construction, GREASE policy and extension values
belong in Surf. QUIC does not depend on uTLS. `tlsquic` is a small interface
using standard-library TLS event types; Surf owns the TLS implementation adapter.

`Config.ClientHandshake == nil` preserves upstream client behavior. Servers keep
the upstream TLS path. Controls are per configuration/connection, never globals.
Known parameter values stay coupled to the engine: stream credit configures the
actual receive controllers; UDP limits provision real receive buffers; ACK delay
sets the receive scheduler; CID and datagram limits configure their validators.
Raw additional parameters may only describe extensions outside engine ownership.
Advertising an extension does not implement it.

## Concurrency corrections

The 0-RTT outgoing-stream fix follows the proposed upstream
[PR #6011](https://github.com/quic-go/quic-go/pull/6011), reviewed at
`ec20e6e567f417325633c6aedc4810b216fc4230`. It was **not merged** at import.
Keep this correction separate from the pinned release baseline and remove it
when an official update includes the same behavior.

Outgoing maps retain applied send windows and RESET_STREAM_AT capability under
their existing locks. Creating a stream no longer reads the parameter pointer
that the handshake can replace, or uses larger newly received limits before
handshake completion. Existing MAX_STREAM_DATA increases remain effective.
The corresponding upstream regression tests are included.

A local stress regression also reproduced a pointer race in SendDatagram;
ConnectionState accesses the same pointer. Publication and these reads use the existing
connection-state mutex; nil parameters reject datagrams normally. No browser
values, presets or new configuration options are involved.

## Small hooks

Initial byte-range recovery is isolated in `crypto_range_recovery.go`.
`ClientHandshake` optionally selects range tracking, a recovery layout, Initial
PTO allowance/preparation, number skipping and reactions to peer probes. Defaults
retain native behavior; browser values belong in Surf. The packet packer invokes
the recovery policy with live budgets and absolute offsets, including after HRR.
PTO preparation can retain original ACK callbacks without declaring those packets
lost. Keep the interval-oracle, late-ACK, HRR and raw-wire regressions when merging
upstream; remove local mechanisms only when upstream supplies equivalent hooks.

| Extension | Own implementation | Existing source hooks |
| --- | --- | --- |
| Client controls, per-attempt factories, receive windows, datagram/CID/ACK limits, Initial ID/number | `client_handshake.go` | `interface.go`, `config.go`, `connection.go`, `transport.go` |
| Explicit local initial RTT and initial PTO factor, measured-RTT metadata | client controls; `initial_rtt_test.go`, `internal/utils/initial_estimate_test.go` | one client initialization hook; optional fields in `RTTStats`; atomic measurement flag and `ConnectionStats.RTTMeasured` |
| UDP receive size (including the omitted protocol default) | `internal/udp/receive_size*.go`, `packet_receive_size.go` | `internal/udp/conn*.go`, `transport.go`, one receive-only pool branch in `internal/udp/buffer_pool.go` |
| Optional UDP read batch capacity | `internal/udp/receive_batch*.go` | `Transport.ReceiveBatchSize`, validated Wrap option, dynamic OOB buffers and refill capacity |
| Configurable congestion and pacing | `internal/congestion/configured_sender.go`, `sender_config.go`, `sender_events.go`, `recovery.go`, `pacing_*.go`, `cubic*.go`, `clock.go` | compact event adapters and persistent-congestion dispatch in ACK handler; native Reno stays unchanged |
| ACK scheduling and active CID limits | `internal/ackhandler/receive_config.go`, client controls | one tracker field/alarm hook and the manager limit hook |
| Optional client receive ACK policy | `ack_policy.go`, `internal/ackhandler/ack_policy.go` | one Application Data tracker branch and client setup; native and server defaults retained |
| Dynamic QPACK receive tables and bounded feedback | `internal/qpack/`, `http3/qpack.go`, `read_stream_wait.go` | HTTP/3 stream/connection decoder adapters |
| Optional QPACK stream order, type timing and encoder capacity announcements | `http3/qpack_client.go` | one transport option, client setup, peer SETTINGS and request header/trailer hooks |
| Ordered/additional client parameters and external GREASE | `transportparams/` | one field and return adapter in `internal/wire/transport_parameters.go` |
| External TLS client | `tlsquic/` | client-only factory and error propagation in `internal/handshake/crypto_setup.go` |
| Initial SETTINGS before request writes / optional first-flight staging | `http3/client_settings.go`, `http3/client_initial_settings.go`, `early_data_ready.go` | readiness channels, nonblocking control/QPACK-type preparation and per-dial application/early-ready callbacks |
| Buffered peer SETTINGS before normal request encoding | `http3/client_buffered_settings.go` | client handshake notification after the current receive batch; nonblocking receive peek/accept; one initial HTTP/3 input handoff |
| Authenticated initial peer SETTINGS and selected-connection request headers | `http3/client_peer_settings.go`, `http3/request_preparation.go` | optional Transport fields, common client-constructor forwarding, normal/raw/control initialization, shared setting validation and field-limit lookup; see `http3/UPSTREAM.md` |
| HTTP/3 ordering and priority | `http3/` custom files and existing small hooks | see `http3/UPSTREAM.md` |
| First HTTP/3 body read and cancellation of blocked uploads | `http3/request_body_reads.go` | optional wire callback, validated header collection and body-copy hooks |
| Encoded packet-number length, temporary coalescing-budget snapshots and recovery-driven ACK-only probes | `packet_policy.go`, `internal/ackhandler/client_packet_policy.go` | send-handler length/MTU/Retry hooks, coalesced/PTO packing and short-header reset; policy stays outside the engine |
| Optional early-data send boundaries, small protected tails and creation-order stream service | `early_data_ready.go`, `stream_creation_scheduling.go` | restored-parameter readiness, one send-loop hook, per-packet STREAM threshold and normal Initial coalescing guard; one creation hook in the stream flow-controller factory |
| Optional Initial datagram padding, separated handshake ACK flight and NEW_CONNECTION_ID order | `initial_datagram_padding.go`, `connection_id_frame_order.go` | client packer options, Initial padding accounting and a control-queue ordering branch |
| Optional client Initial protection, address-based packet size, ACK-only PING and receive-credit batching | `initial_packet_protection.go`, `client_handshake.go` | packet serialization/handshake hooks, per-client flow-controller divisor; no browser defaults |
| Optional ordered Initial CRYPTO ranges with live packet budgets | `initial_crypto_layout.go`, `packet_packer_initial.go` | client configuration, manager-owned Initial stream, existing protection queues and one pre-serialization preparation step; native recovery retained |

A separate upstream bug fix in `crypto_stream.go` sorts absent SNI cuts after
real ECH cuts. ECH without SNI (IP origins) otherwise leaves HasData false and
stalls the handshake. A byte-reassembly regression covers this case.

`InitialRTT` seeds local RTT/pacing/recovery until a real ACK measurement;
`InitialPTOMultiplier` optionally controls PTO before that first measurement.
Zero fields retain native behavior, including token restoration. Real samples
replace estimates, path migration forgets the old estimate, and clones retain
the explicit policy. No transport parameter or browser value is added here.
`ConnectionStats.RTTMeasured` distinguishes an observed RTT from an initial/token
estimate, including measured paths whose RTT variance has reached zero. Its
underlying flag is atomic so the public statistics API remains safe to read.

Test-only adaptations: HTTP integration tests/examples use fork HTTP types and
ordered maps. `hkdf_test.go` uses fixed Expand-Label vectors instead of private
crypto/tls linknames rejected by Go 1.27. The config-field audit includes the new
optional field. Keep upstream tests with their source packages when private
access is required; public extension tests live under `tests/`.
Malformed QPACK peers wait for a request before injecting an error, avoiding a
legitimate pre-request transport retry. Stream cancellation completion waits
use the integration suite's time scale and retain all cancellation assertions.

## Audit and update

Early staging uses `OnEarlyDataReady` only when configured. By default its
bounded, nonblocking writes can share the Initial flight. `FlushEarlyData`
instead gives currently sendable Initial CRYPTO and setup data separate send
opportunities before publishing the connection. Pacing, congestion and queue
limits are never bypassed. Nil callbacks preserve native readiness. HTTP/3
chains callbacks from cloned configurations and shares one connection owner.

`CoalesceSmallEarlyPackets` relaxes the 128-byte heuristics only for 0-RTT;
actual long-header, AEAD, sample and frame sizes still bound writes.
`InitialCryptoCoalescingBarrier` also prevents later spaces sharing an Initial
while more Initial CRYPTO is queued. Both default off.
`StreamSchedulingByCreation` replaces local urgency/incremental service with
open-stream creation order, while preserving retransmission precedence and
skipping credit-blocked streams. Its linked state contains IDs only and is
removed on completion or 0-RTT rejection. Other connections keep native queues.
Within that scheduler, `PrioritizeFinishedStreams` optionally services queued
bytes and FIN of write-closed streams in the retransmission class. Lost ranges
and finished writes retain creation order together; ordinary new data follows.
The stream's existing mutex protects its write-side state. Reset/canceled
streams are not promoted. The option has no effect on urgency/incremental
scheduling and does not alter HTTP priority fields, flow-control credit, packet
sizes or send timing. It defaults off and is copied with the handshake options.
Regressions cover both defaults, finished data spanning packets, empty FIN,
creation order across loss and new finished data, blocked credit and rejection
cleanup. All policy selections remain in Surf's public builder/presets.
These are generic opt-in mechanisms; all browser selections remain in Surf.

`python3 scripts/upstream.py audit` compares with the pinned Git source after
normalizing module/HTTP imports and gofmt. `audit --patch` shows the substantive
patch. New extension files are listed separately. This requires `gofmt` on PATH.
The script is a maintenance tool, not a runtime dependency.

Before updating, commit reviewed local changes. Fetch the new upstream tag and
create a separate Git worktree from the fork's HEAD. Merge the pinned tag or commit there with
`--no-commit --no-ff`; resolve the small hooks using the inventory above. Normalize
new upstream package imports to `github.com/enetx/quic-go`, and HTTP/3/integration
HTTP imports to `github.com/enetx/http`. Keep source links pointing upstream.
Update the pin only after reviewing and testing the merged candidate. Never merge
upstream changes directly over an uncommitted Surf working tree.

Validation:

```sh
go mod tidy
go test -race -count=1 ./...
go vet ./...
python3 scripts/upstream.py audit
```

Also run Surf's full race/coverage command and independent Initial-packet wire
tests against the candidate. Check that nil controls retain upstream behavior,
custom values reach the peer, session metadata survives resumption, invalid TLS
templates fail explicitly, and HTTP/1/2 fixtures remain unchanged.

## Module boundaries

Surf imports HTTP/3 from `github.com/enetx/quic-go/http3` and uses released
HTTP and QUIC modules. Local replacements are optional development overrides.
HTTP/3 source and tests live only in this module. The duplicate implementation
was removed from the HTTP module after migration and regression comparison.

Published versions depend on the released HTTP module with its exported sorter.
No local replacement is needed by module consumers. Release HTTP changes first
when QUIC needs a new shared HTTP API, then release QUIC and update Surf.

The fork owns `.github/workflows/go.yml`: Go 1.27 checks formatting, module
consistency, the complete race-enabled suite and vet on Linux, macOS and Windows.
Upstream Docker publishing, cloud fuzzing and third-party upload workflows are
not installed in the fork; they require upstream-specific credentials/services.

## Verified behavior and remaining limits

The receive-side dynamic QPACK implementation follows RFC 9204. It reuses the
unchanged official qpack module for the static table/outbound encoder and x/net's
Huffman codec. Capacity, blocked streams, decoded header size and feedback queues
are bounded. Cancellation, deadlines, connection shutdown and critical-stream
errors retain native QUIC semantics. Zero capacity keeps the upstream static path.
RFC Appendix B vectors, fragmented instructions, table eviction/wrap, blocked
streams, errors, cancellation and real HTTP/3 client/server exchanges are tested.
Outbound dynamic encoding now lives in `internal/qpack/encoder.go`. It reuses
official static-table entries and x/net Huffman coding, without another dependency
fork. Generic options select indexing, strings, Base, instruction names, draining,
and bounded outstanding sections; browser choices remain in Surf.

`http3.QPACKConfig` optionally reserves encoder/decoder streams in application
order, sends their types eagerly or with their first instruction, and queues an
encoder capacity announcement after SETTINGS or with the next HEADERS block.
The value never exceeds the peer's limit or the configured local budget; absent
or zero peer capacity keeps requests static. Zero announcements are optional.
One writer per critical stream
preserves instruction order and respects flow control without blocking request
headers or the control parser. Feedback queues remain bounded. Nil configuration
retains native behavior, and server behavior is unchanged. Browser values and
mode selection belong in Surf; the fork contains no browser detection or sleeps.
The protocol matrix, reserved IDs, lazy type bytes, capacity bounds, snapshots,
limited stream credit, critical-stream STOP_SENDING and dynamic response feedback
are tested in `tests/http3/qpack_client_test.go`.

Surf owns independent Chromium 158 / Firefox 160 QUIC ClientHello fixtures.
Its tests decrypt Initial packets independently and compare TLS, ordered SETTINGS,
headers and supported transport parameters. Manual warmed requests to the external
HTTP/3 endpoint pass, including the Firefox response that previously exposed the
dynamic-QPACK gap. The external endpoint is not a CI dependency.

This does not establish complete browser identity. Upstream quic-go lacks
compatible in-flight version negotiation. This fork now adds opt-in client
v1/v2 switching with authenticated validation; see Client version negotiation
below. Native server negotiation policy remains unchanged.
The native ACK_FREQUENCY parser has different last-field semantics from the
legacy format. The fork exposes an isolated opt-in legacy receiver and local
min_ack_delay advertisement; see below. It does not generate adaptive requests
for the peer's ACK rate.
ACK ratio/frequency, packet scheduling/coalescing, congestion/loss behavior and
resource-specific priority policy still differ from browser engines. Browser
defaults and workaround decisions belong in Surf.

The outgoing encoder commits a section only after its complete instruction batch
is accepted by the bounded writer. Backpressure falls back to static encoding,
without references to unsent insertions. Eviction protects unacknowledged entries,
outstanding references and references acquired in the current section. RIC modulo
uses the peer's SETTINGS maximum independently from the chosen capacity. Feedback
is FIFO per stream, with separate section ACK, cancellation and increment handling.
Request trailers share the connection encoder and deterministic header ordering.
Without encoder-stream credit, sections remain static until the critical stream
can open; queued capacity instructions cannot create references to unsendable data.

State management was cross-checked against the independent Go implementation in
[`mach`](https://github.com/Lemon4ksan/mach/tree/17ea25018f3b1d8a2810f4f409974d225a6eb2bb/proto/codec/qpack)
and RFC 9204. It is a reference, not a new module dependency. The implementation
remains isolated here; static/Huffman tables and browser policies are not copied.

`tests/qpack/encoder_test.go` and `encoder_fuzz_test.go` cover native static bytes,
independent wire vectors, Huffman/N bits, both Base forms, wrapping, eviction,
duplication, tracking/blocking budgets, backpressure rollback, partial/invalid
feedback, concurrency and options. `tests/http3/qpack_encoder_test.go` exercises
real HTTP/3 connections, concurrent bodies/trailers and connection-level errors.
Surf's separate browser audit compares raw QPACK blocks and encoder instructions.

The connection-migration regression checks STREAM packet tuples after a switch;
late validation responses may still use a nondefault tuple. Explicit probe,
packet-count, address and connection-ID unlinkability assertions are retained.

Client receive ACK policy is now configurable: initial/decimated packet thresholds,
transition by packet-number span, RTT-scaled delay and floor, fixed earliest deadline,
gap-window/immediate reordering and ECN response. Browser policy remains in Surf.
The native tracker branch and Initial/Handshake spaces are preserved. Policy is
cloned per configuration and connection. These local controls do not negotiate ACK_FREQUENCY; the separate opt-in legacy
receiver is described below. They do not change congestion control, packet pacing
or version switching.
Deterministic `internal/ackhandler/ack_policy_test.go` cases cover boundary numbers,
non-eliciting packets, timer expiry, real RTT samples, gaps, ECN and snapshots.

Optional client packet controls remain independent of receive ACK scheduling.
`InitialPacketProtection` fragments/reorders actual CRYPTO and distributes the
existing padding budget. The multi-packet layout preserves stream offsets and
native fragment retransmission handlers. Configuration is snapshotted per
connection; browser values and policies remain in Surf. Property tests cover
partial ClientHello writes, packet budgets, large messages, recovery and later
TLS messages.

`ClientHandshake.InitialCryptoLayout` optionally chooses ordered ranges for new
data from the first complete ClientHello. Its callback receives the message's
length and SNI/ECH positions, a private ascending `Remaining` range snapshot,
the current plaintext `PacketSizeLimit` (including the QUIC header, excluding
AEAD expansion), and `FrameBudget` after the header and any ACK. Range offsets
exclude CRYPTO frame headers. Missing SNI/ECH positions are `-1` with zero length.

All returned groups must cover exactly `Remaining`, with positive lengths and
no duplicate, omitted, out-of-bounds or previously consumed bytes. Invalid output
fails before serialization. The engine copies ranges and selects only the first
group for this packet; the other groups establish coverage, not a fixed plan for
later packets. Actual popped CRYPTO bytes, including budget-forced splits, are
removed from the next snapshot. Each later packet carrying new ClientHello data
gets fresh header/ACK budgets. Callbacks run in the send loop and must not block
or reenter the connection.

Nil output on the first call retains native preparation, including the existing
scrambling/protection configuration. After a custom group, nil output selects
the remaining ranges sequentially. Explicit custom groups retain their order
through the ordinary writer instead of its native frame shuffle. Separately
enabled `InitialPacketProtection` can still split and shuffle those frames as
requested. Application policy chooses the layout; the fork contains no browser
hostname-slicing rule, balancing constant or preset.

This hook builds on the pinned upstream
[`crypto_stream.go`](https://github.com/quic-go/quic-go/blob/v0.65.0/crypto_stream.go)
and [`packet_packer.go`](https://github.com/quic-go/quic-go/blob/v0.65.0/packet_packer.go)
paths; it is a fork extension, not functionality supplied by that release.
It waits for a complete first ClientHello. Partial-message emission, later TLS
messages (including a second ClientHello), and retransmitted CRYPTO retain native
handling. Recovery preserves bytes and offsets, but does not rerun the layout
callback over previously consumed/lost ranges. In particular, loss of an entire
flight is not evidence that its retransmission layout matches an application
that slices the restored byte ranges again.

`initial_crypto_layout_test.go` and `initial_crypto_layout_dynamic_test.go` cover
native nil fallback, SNI/ECH and fragmented TLS writes, exact coverage validation,
caller ownership, actual serialized frame order, changing live budgets, forced
splits, remaining-range updates, subsequent TLS data and client/server isolation.

`ReceiveWindowUpdateDivisor` applies to real stream/connection controllers.
`AckOnlyPingThreshold` can retain, replace or disable periodic PING insertion.
`AckOnlyMaxDataThreshold` queues a repeat of current connection credit to solicit
an ACK, without increasing credit. `OmitInitialAckOnHandshake` omits an ACK-only
Initial only when Handshake data/keys are ready. Initial DCID length and
address-based UDP-size factories run per dial attempt; errors/invalid results
fail before packet emission. Default client/server behavior remains native.

Optional `PacketNumberLengthPolicy` receives a recovery snapshot and selects encoded
bytes. Invalid lengths fall back to native; too-short lengths are increased to
retain a safe decoding range. Server policies are never installed. The handler
tracks the current datagram size for congestion-window estimates.
`PacketNumberLengthInfo.CoalescedMaxPacketsInFlight` additionally estimates that
window using the remaining unpadded long-header datagram budget from the previous
packet in the same number space. It is zero without a usable temporary budget;
the existing `MaxPacketsInFlight` remains based on the path's real datagram size.
Normal coalesced packets and long-header PTO probes update this private snapshot.
Remainders too small for another protected header are ignored. A number-space
change, short-header construction, Retry or an MTU update discards the temporary
limit; it cannot reappear when returning to an earlier space. This does not
change the path MTU, congestion state, padding, packet numbers or RFC 9000 safety
checks. The application decides whether the extra estimate affects its selected
width. No policy leaves native selection unchanged, and older policies can
ignore the new field. `coalesced_packet_policy_test.go` decodes actual packet
widths and covers unusable remainders, 1-RTT isolation and PTO updates; the
ackhandler tests cover snapshot values, number-space/Retry/MTU reset and servers.

`AckProbePolicy` receives tracked application-data packet count (including
ACK-only sends), last eliciting send age and unbacked-off PTO. It only adds a
PING to an ACK-bearing packet when congestion and packet space permit it; native
PTO recovery remains independent. Lost-packet retention still follows quic-go.

`DisableConnectionIDRotation` suppresses unsolicited changes on the current path,
including the first post-handshake change. Peer retirement and path migration
remain mandatory and have explicit regressions. `OnApplicationReady` is a
nonblocking staging callback after verified TLS and applied peer parameters,
before releasing `HandshakeComplete`. It must never call blocking stream writes
or `CloseWithError`. HTTP/3 uses `TryWriteAll` and hands blocked writes to a goroutine.

`InitialPaddingMode` optionally moves the datagram size padding outside coalesced
long-header packets. Header-protection sample padding remains encrypted. A final
short-header packet instead receives encrypted padding because its ciphertext
extends to the UDP boundary; appending plaintext would invalidate AEAD. Datagram
padding is charged to the Initial sent-packet size, while qlog packet lengths
remain their actual encoded lengths. Initial probes and close packets use the
same policy. Servers and zero-valued client controls keep native behavior.

`HandshakeAckBeforeData` optionally packs queued Initial/Handshake ACKs before
consuming local Handshake CRYPTO. The ACK-only prefix never includes application
data. Native disposal of Initial keys on the first Handshake send then allows the
following Finished flight to remain unpadded. There is no injected timer, TLS
verification bypass or asynchronous TLS-engine modification. Certificate
verification still completes synchronously; this reproduces flight composition,
not NSS authentication latency.

`SortConnectionIDFrames` controls the NEW_CONNECTION_ID subsequence independently
of other frame types. Queue admission preserves ascending order across packet
budgets, and serialization restores it after native control-frame shuffling.
Every frame moves with its loss/ACK handler. Default queues and servers retain
native ordering. Unit tests cover packet budgets and 100 shuffle seeds; live
transport tests cover each padding/ACK combination, rejected certificates,
0-RTT acceptance/rejection and loss of each of the first three outgoing datagrams.

HTTP/3 also accepts a per-connection `WireConfig.SettingsGrease` factory. Reserved
identifier validation and ordered snapshotting are isolated in
`http3/settings_grease.go`; browser distributions remain outside the engine.

## Optional packet-number spaces

`ClientHandshake.SeparatePacketNumberSpaces` keeps wire generators and histories
for Initial, Handshake and Application Data independently. Only Initial receives
the configured starting number; the other spaces start at zero.
`DisablePacketNumberSkipping` suppresses random and PTO-induced Application Data
gaps; the default preserves native skipping and skipped-ACK rejection. 0-RTT and 1-RTT
share the Application Data space. The default remains the native shared counter.

The isolated `internal/ackhandler/packet_number_spaces.go` selects histories.
`packet.CongestionNumber` is a separate monotonic transmission sequence, used
only for congestion/recovery callbacks; qlog, ACKs, ECN and retransmission
bookkeeping retain wire numbers. Lost-packet entries preserve congestion IDs
for later spurious-loss reports. Counts across histories still enforce native
packet-storage limits. Retry and migration do not reset wire generators.
Tests cover overlapping numbers, ACK isolation, unsent/skipped ACK rejection,
0-RTT rejection, Retry, key disposal, loss/undo and migration. Surf's Chromium
profile deliberately keeps shared allocation; its Firefox profile enables the
independent spaces after verification of the respective browser sources.

## Legacy ACK_FREQUENCY receive mode

`ClientHandshake.LegacyAckFrequencyMinDelay` enables receive support and emits
local parameter 0xff02de1a. Zero retains native parsing and advertisement; positive
whole-microsecond values must fit the configured max_ack_delay. Profiles and the
chosen minimum remain in Surf. `Config.LegacyAckFrequencyMinDelay` also enables
the same receiver on either endpoint, including opt-in test servers. The existing
client-only value overrides it when nonzero. Sender policy is independent.

`internal/wire/legacy_ack_frequency_frame.go` isolates the three-varint plus
single-byte ignore_order format. It is deliberately separate from the existing
newer draft parser. Legacy mode accepts it only in 1-RTT and does not enable
IMMEDIATE_ACK. Tolerance zero and flags other than 0/1 are frame errors; requested
delay below the local advertised minimum is a protocol violation. Oversized delay
conversion and alarm deadlines saturate without integer wrap.

The tracker uses the requested tolerance, delay and ordering flag after the first
valid update. Equal sequence numbers replace state, older sequences are ignored,
and an already pending ACK deadline is retained. Initial/Handshake trackers and
native behavior when unset are preserved. Source references:
[Firefox legacy frame parser](https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/third_party/rust/neqo-transport/src/frame.rs),
[receiver validation](https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/third_party/rust/neqo-transport/src/connection/mod.rs), and
[sequence/timer rules](https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/third_party/rust/neqo-transport/src/tracking.rs).

Transport-parameter layout preserves this known engine field even though its
number is congruent to 27 modulo 31. Public custom-parameter injection remains
forbidden for engine-owned IDs. Private parser/tracker tests, independent Neqo
bytes, qlog serialization and real client/server frame injection cover the mode,
including disabled and invalid-delay connection errors. Surf additionally decrypts
Initial bytes to verify the Firefox advertisement, ordering, overrides and opt-out.

## Outgoing legacy ACK_FREQUENCY

`ClientHandshake.PeerAckFrequencyPolicy` enables a per-connection controller only
after the actual handshake and an opted-in decode of peer parameter 0xff02de1a.
Native clients still ignore this parameter. Capability is not saved for 0-RTT.
The callback supplies policy from the application's live congestion window,
datagram size and smoothed RTT; browser constants and hysteresis remain in Surf.
Nil retains native behavior. Callbacks must be concurrency safe across connections
and must not block or reenter a connection.

`peer_ack_frequency.go` isolates sequencing, packet budgets and frame delivery.
Only one request is outstanding; loss allows a new sequence with the latest
target instead of retransmitting stale settings. Late feedback cannot clear a
newer outstanding request or restore an older acknowledged state. Recovery uses
the maximum of current, target and sent-but-unconfirmed delay, and rearms PTO on
changes. The unconfirmed high-water bound is deliberately conservative compared
with Neqo's current/target bound; it clears when the latest request is acknowledged.
Targets must have a positive QUIC integer tolerance and a whole-microsecond delay
at least the peer minimum and at most 16383ms, the engine's max_ack_delay range.
Invalid callback results close the local connection.

`ClientHandshake.InitialCongestionWindowPackets` exposes only the initial window
(zero retains native defaults; explicit range 2..10000). Migration restores that
configured window against the new accounting segment size. The optional
congestion selectors described below extend this constructor; the native default
remains Reno with the original HyStart and native pacing. Tests cover independent wire bytes, real negotiated
and unnegotiated connections, invalid policies, delayed feedback, sequence
exhaustion, packet budgets, PTO rearming, migration and non-resumption of capability.

## HTTP/3 connection preparation

`http3/transport_preconnect.go` adds `Transport.Preconnect(ctx, hostPort)` using
the transport's existing dial coalescing, client factory, SETTINGS setup and pool.
It waits for a verified handshake, including when the underlying dial uses early
data, without sending a request. Failed or canceled preparation removes/closes an
attempt only when no other pool user holds it. A shared dial survives cancellation
of one waiting caller. Successful preparation leaves the connection in the normal
pool; idle cleanup and Close retain ownership. `PreconnectEarly` exposes the
same pool entry when early keys are ready, returning its transport-owned Conn;
it neither sends an HTTP request nor treats early readiness as authentication.
Cold connections still wait for the handshake. Tests cover real HTTP/3 reuse,
shared cancellation and a gated resumable handshake.

With `CoalesceInitialSettings`, the small dial hook in `http3/transport.go`
keeps setup per candidate connection. A custom address dialer can invoke several
application-ready callbacks and return a different winner. Setup must follow
that winner, rather than the first callback; temporary candidate references are
released when dial completes. A two-candidate regression reproduces the failure
with the old single sync.Once and passes with candidate-specific setup.

Protocol competition, timers, browser policy and body ownership remain in Surf.

`SessionVersion` and `ClientSessionMetadata` read the existing client-side
session-state metadata block. Its local format revision is 7: negotiated version,
measured RTT (zero without a sample), Initial key-context CID and remembered
transport parameters. Revisions 5 and 6 remain readable for ordinary
resumption/0-RTT; version-bearing revision 6 has no RTT or CID. Returned byte
slices are snapshots. The number and order of application Extra entries are
preserved. This metadata never changes the opaque server ticket or TLS wire
format. Generic metadata, malformed/legacy states and authenticated compatible
switches are tested. Choosing an initial version from a live ticket remains
Surf policy, including origin isolation, allowed versions, cache eviction and
expiry; the QUIC engine does not add an origin-history database.

`ClientHandshake.OnEarlyDataRejected` observes the client's existing TLS
0-RTT rejection event before its encryption level is discarded. The callback
runs on the connection loop and must return promptly without reentering it.
Surf uses it to update its remaining ticket queue; no browser cache policy or
origin database enters this fork. Standard TLS and custom TLS-engine rejection
paths are exercised by Surf's actual HTTP/3 handshake tests.

## Client congestion control

`ClientHandshake.CongestionControl` selects native Reno, explicit Reno,
RFC 9438 CUBIC with whole-segment ACK credit, or fixed-point byte-mode CUBIC.
`SlowStart` selects original HyStart or classic byte growth (not HyStart++).
`CongestionControlSegmentSize` can fix the accounting segment independently of
UDP MTU; zero follows MTU. Byte-mode `CongestionControlConnections` configures
connection emulation. Browser choices and all preset values remain in Surf.

The RFC controller is isolated in `internal/congestion/cubic_rfc9438.go`.
ACK-frame batching and send-time window utilization preserve byte accounting
when delayed ACKs cover several packets. Loss reduces by 0.7; ECN uses 0.85;
spurious-loss undo restores the curve, ACK credit and saved window. Byte-mode
CUBIC supports an explicit segment size and matches QUICHE's unsigned intermediates
and conversion after the floating-point addition. The upstream Reno-friendly test
helper is adjusted to that corrected rounding; independent C++ vectors validate it.
Migration recreates the configured controller; timeout resets its transient state.
Client options are ignored by server constructors. Defaults and server behavior
retain their native defaults. Optional scheduling and recovery controls are
described below; these mechanisms do not establish full browser recovery parity. The source-derived suites execute 1120 original Neqo and
540 original QUICHE controller vectors; their provenance and runners are in
`internal/congestion/testdata`. They are not browser packet captures.

## Client version negotiation

`ClientHandshake.CompatibleVersionNegotiation` enables RFC 9368 / RFC 9369
v1/v2 switching and authenticated version negotiation checks. Native clients
and server policy retain their defaults. `Config.Versions` supplies permitted
versions and the initial version; a client layout may provide parameter 0x11
with per-connection preference / GREASE. The engine validates and promotes that
parameter into the typed codec, retaining its position and exact version values.
No browser preferences are embedded in the fork.

`compatible_version.go` owns selection guards, alternate Initial verification,
server version-information checks and version-scoped token keys. It uses the
existing bounded undecryptable-packet queue for reordered Handshake packets.
Initial keys and packet key labels change only after a candidate Initial passes
AEAD verification. The first server CRYPTO locks the version. Initial PN space,
TLS transcript, congestion state and Retry's current key CID are retained.
`ConnectionState.Version` reports the negotiated version. An authenticated
mismatch, unavailable choice or induced downgrade fails with transport error 0x11.
Malformed version information fails with TRANSPORT_PARAMETER_ERROR; the TLS
error wrapper now preserves existing transport errors instead of masking them
as INTERNAL_ERROR. The RFC 9368
section 8 exception for old v1-only servers is covered.

The handshake's isolated helper changes packet key labels and scopes client
session-cache keys to the negotiated version. Packet packing retains original
version headers and keys for 0-RTT after a switch. NEW_TOKEN entries are likewise
isolated by version. A ticket from v2 is not offered by an original v1 connection;
choosing the next connection's original version from browser history is separate
application policy. No TLS library fork is required.

Private-package tests cover both directions, Retry, an original-version ACK,
reordered Handshake, a lost flight, corrupted candidate packets, classic VN and
VN followed by compatible switching, old servers, malformed / contradictory
information, downgrade rejection, PN decoding, 0-RTT packet types, cache isolation
and real TLS application keys. The UDP packet peer drives actual client Dial and
verifies client Finished. Its TLS/packet primitives are native and it is not an
independent implementation of selection. Additional manual HTTP/3 interoperability
checks use unmodified official aioquic 1.3.0, with private test certificates trusted
by the Go clients, covering both directions, Retry and accepted original-version
0-RTT across a switch. Surf preserves the sanitized reference results separately.

## Optional send scheduling and recovery

`ClientHandshake.MaxCongestionWindowPackets` retains the native ceiling at zero,
sets an explicit segment ceiling when positive, and removes that fixed ceiling
at -1. Accounting segment and path MTU remain independent. Integer operations
and packet-history limits still bound the engine; no memory limits are disabled.

`PacingConfig` selects a leaky bucket with signed debt or scheduled pacing with
initial/idle and lumpy bursts. Applications supply gains, burst sizes, bandwidth
threshold and granularity. `ProportionalRateReduction` enables RFC 6937;
`RecoveryBurstPackets` supplies a bounded recovery allowance.
`PersistentCongestionThreshold` requires a contiguous loss interval bounded by
ack-eliciting packets after the first RTT sample; a PTO expiration alone never
collapses the window. Application-limited notifications from empty packet packs
stop the scheduled pacer from accumulating credit during application pauses.

The code is isolated in `internal/congestion/pacing_policy.go`, `recovery.go`,
`internal/ackhandler/persistent_congestion.go` and `client_send_config.go`. The
connection packing path has two optional notification hooks.
`ClientSendConfig` keeps PN/loss bookkeeping separate from congestion arithmetic.
No browser presets, uTLS dependency, DNS resolver or browser sleeps enter this fork.
Browser policy and connection discovery remain in Surf.

In addition to CUBIC reference fixtures, tests execute 120 pacing rows generated
from unmodified Firefox 160 Neqo `pace.rs`, and 40 PRR states (200 allowance
comparisons) from unmodified pinned QUICHE source/header. Go tests need neither
Rust nor C++; reproducible developer runners and hashes are in
`internal/congestion/testdata/README.md`. Surf additionally verifies complete
bidirectional transfers under deterministic packet loss with both profiles.
These are source and transport checks, not proof of identical browser NIC timing.


## Optional Handshake and Application PTO policies

`HandshakeProbePackets` and `ApplicationProbePackets` independently select a
bounded 0..2 PTO allowance; zero preserves the native allowance.
`HandshakeProbeSkipPacketNumber` controls the Handshake gap. Optional
`ApplicationProbeSkipPacketNumber` separates the Application PTO gap from random
number skipping. Nil preserves `DisablePacketNumberSkipping`'s original authority;
its pointer value is privately cloned. Skipped-number
ACK validation remains enabled. `HandshakeProbeOnlyWithData` suppresses only an
empty follow-up probe after the first Handshake PTO packet.

`HandshakeCryptoRecovery` reuses `crypto_range_recovery.go` for byte ranges and
late original PTO acknowledgments. It releases the transcript when Handshake
keys retire. `HandshakeCryptoResendOnPeerProbe` enables PING/non-progressing
CRYPTO reactions; duplicate Initial CRYPTO can also reschedule Handshake bytes.
The upstream default STREAM retransmission accounting is unchanged.

`PeerPacketRecoveryPolicy` receives a routed packet's missing/retired-key state
and buffer counts. Authentication failures never invoke it. Actions can schedule
unacknowledged Initial bytes or a Handshake PING if its sealer is available;
neither action accepts ciphertext or weakens TLS verification. The application
supplies conditions and limits. `ImmediateACKOnPeerPing` optionally queues an
immediate ACK for an authenticated Application Data PING, independently of the
normal or negotiated ACK policy. All controls default to native behavior.

Keep the actual ACK-handler, missing-key/authentication-failure and encrypted
UDP loss tests when updating upstream. Browser policies and independent native
capture evidence live in Surf. This change does not implement complete STREAM
copy tracking, simultaneous-space Neqo probe batching or early-data peer hints.

## Grouped application stream writes

`Conn.TryWriteStreamBatch` stages a bounded list of existing local-stream writes
and optional FINs in one connection-loop turn. It never waits for stream credit
or a competing writer, returns the accepted prefix, and preserves normal packet
size and congestion limits. The generic implementation lives in
`stream_write_batch.go`; the connection loop has one channel and two dispatch
hooks. Default request writes do not use the operation.

Optional `WireConfig.CoalesceRequestHeaders` groups initial priority and HEADERS,
including FIN only for automatically managed bodyless requests. Manual request
streams stay open. The single control writer retains ordering and uses the
asynchronous path under backpressure. Thirty-request packet tests, manual DATA,
flow-control blockage, partial batches and close/error cases are covered.

`QPACKConfig.EncoderCapacityRequiresSetting` distinguishes an omitted peer table
setting from an explicit zero while preserving the independent
EncoderZeroCapacity switch. Browser choices remain exclusively in Surf.

## 32-bit build coverage

Keep QUIC's 62-bit bounds explicitly typed as `uint64` at variadic formatting
and assertion boundaries. An untyped constant passed through `any` defaults to
`int`, which overflows on 32-bit targets. The datagram-limit diagnostic retains
its exact value and validation behavior with this conversion.

CI builds every package and compiles all tests for Linux 386, ARMv6 and ARMv7
with CGO disabled. These cross-compilation checks skip execution; native race
tests continue to run separately on Linux, macOS and Windows.

## Rejected early stream generations

`ClientHandshake.ResetStreamsOnEarlyDataRejection` enables fresh stream maps at
verified application readiness after TLS rejection. The optional epoch sender
drains existing callbacks and suppresses callbacks from retired streams before
IDs are reused. A locally finished early send stream still receives its shutdown
error, preventing a late RESET_STREAM from targeting a new stream with that ID.
Nil/default clients preserve explicit `NextConnection` handling and incur no
epoch-wrapper callbacks. Browser selection belongs to Surf, not the engine.

## Application STREAM byte-range recovery

`ClientHandshake.StreamDataRecovery` selects an optional SendStream range tracker.
Each sent frame has immutable range/FIN metadata; any original/copy ACK retires
covered queued copies and logical live-frame counts. Probes preserve original
ACK handlers without declaring congestion loss. Actual losses retain only STREAM
handler metadata separately from packet flight accounting. `StreamLossRetention`
receives elapsed loss time, current PTO and the matching packet-number domain;
browser constants remain in Surf profiles. Nil policy retains useful metadata
up to the existing generic packet-history bound. Default clients/servers retain
the upstream frame path. No new TLS provider or configuration interface is added.

Range recovery lives in `stream_range_recovery.go`; optional packet metadata and
probe scheduling live in `internal/ackhandler/stream_recovery.go`. The STREAM pool
copy helper explicitly resets FIN. Regressions cover ranges, FIN, split/partial
copies, RESET_STREAM_AT, shutdown/early rejection, late loss ACKs, number domains,
Retry, bounded metadata and actual verified-TLS UDP loss/PTO scenarios. Surf
retains independent native and engine capture reports under `stream_recovery`.

### Shared PTO datagrams

`ClientHandshake.CoalescePTOProbes` enables one datagram allowance and cumulative
retransmission byte budget across eligible encryption spaces. Crypto timer
baselines, key-gated Handshake priming and per-space force masks are isolated in
`internal/ackhandler/pto_batch.go` and `packet_packer_pto.go`. The independent
`InitialACKResetsPTO` option controls pre-validation backoff reset. Both default
to false; browser policy stays in Surf. Tests cover v1/v2 packet composition,
ACK-only traffic, Retry, key retirement and early rejection.

### Peer-triggered early data recovery

`PeerPacketRecovery.ResendEarlyData` and
`ClientHandshake.EarlyDataResendOnPeerCrypto` expose early recovery tokens without
TLS rejection, congestion loss or PTO reset. The concrete optional helper in
`internal/ackhandler/early_data_recovery.go` removes early packet accounting and
reschedules owned tokens. A new Application ACK, including retained actual-loss
metadata, ends the mechanism. Authentication failures and retired keys do not
trigger it. Firefox's selection policy remains in Surf; default clients and
servers do not enable it. PING only retains the existing CRYPTO/ACK behavior.

## Receive-loop ACK policy and packet-space snapshots

`AckPolicy.FlushOnPacket` gives an already queued application ACK a send
opportunity before the next received UDP datagram, including a full receive
queue. It preserves socket batch reads; native batching remains the default.
No browser values live in the engine. The hook in `Conn.handlePackets` only
queries pending ACK state, without consuming ranges or resetting timers.
Socket backpressure can still postpone transmission.

Client packet-number and ACK-probe policies inspect `history(space)` so explicit
separate number spaces include the correct 0-RTT/1-RTT history. Tests cover full
receive queues, wakeups, ACK reset, native batching, removed slots and path probes
in both numbering modes. Linux captures of actual browser transfers are retained
as sanitized observations in Surf's `linux_load` fixtures.

## Optional peer transport extensions

`Config.PeerTransportParameterIDs` selects otherwise unknown extension bytes for
the current handshake. `ValidatePeerTransportParameters` receives owned copies
before publication; errors fail the handshake with TRANSPORT_PARAMETER_ERROR.
`Conn.PeerTransportParameter` exposes another private copy only after verified
handshake completion. These bytes never come from session-ticket restoration.
Native duplicate/known-parameter validation remains authoritative. Nil capture
and validator retain native handling. Browser tag formats and defaults live in
Surf, outside this generic extension point.

`Config.ServerTransportParameters` optionally applies the existing validated
layout encoder to local server parameters, including extension fixtures. Clone
and populate operations own the layout and capture-ID slices. The local layout
and captured fields are never serialized as ticket state.

The changes are isolated in `peer_transport_parameters.go`, configuration
copying, the transport-parameter parser and handshake event handoff. Regressions
cover current-handshake ownership, clone immutability, rejected state and tickets.

## HTTP/3 preparation and ready pool snapshots

Connection-start notification, confirmed cached attempts and drained-owner
lifecycle are isolated in `http3/connection_start.go`, `connection_reuse.go`
and `connection_retirement.go`. Transport hooks only publish one shared start
event, reject a pending ready-only probe, and retain active retired owners.
See [HTTP/3 maintenance](http3/UPSTREAM.md#connection-readiness-and-retirement).
Browser delay and request-admission policies remain entirely in Surf.

## Small early packets and rejected critical streams

A FIN-only send frame validates its complete header against the available byte
budget before consuming FIN state. This keeps small coalesced tails within the
path limit and leaves an unsent FIN queued. `send_stream_packet_budget_test.go`
checks IDs/offset varint boundaries and exactly-once FIN; the packet packer
regression reproduces the original 1253-versus-1252 failure deterministically.
No browser frame layout or padding value is added to the engine.

HTTP/3 control and QPACK workers distinguish `Err0RTTRejected` from critical
stream failure, including before the accept loop publishes generation retirement.
They retain the original rejection for normal application retry/restart handling;
genuine stream failures still close the connection. The gated trusted-TLS
regression confirms resumed TLS, rejected early data and a live QUIC connection.
Both regressions fail against the preceding published source tree.

The small-tail framer also sizes the omitted STREAM length from the actual
stream-ID/offset header and data capacity, under the stream lock. Using the
varint width of the entire remaining packet budget overestimates the saving at
64-byte boundaries. `stream_packet_budget.go` isolates the mechanism and retains
the existing frame-getter interface. `framer_packet_budget_test.go` checks both
schedulers, new/retransmitted data, nonzero offsets, tiny/varint-boundary budgets
and delivery without lost or repeated bytes. The preceding code emits 65 bytes
for a 64-byte budget in the deterministic regression.

### Explicit HTTP/3 pool retirement

`http3/transport_retirement.go` adds generic `Transport.RetireConnections`.
It removes current exact/coalesced owners, cancels pending handshakes, preserves
active streams, then closes each drained connection. TLS caches and future
transport use remain intact. `rawConn.trackStream` serializes retirement against
new request tracking; the successful streams-empty hook handles final closure.
`earlyRestartClient` carries retirement through rejected-generation replacement.
No browser profile, network observer, timing constant or additional interface is
part of this mechanism. Existing Close retains ownership of every retired client.

`http3/transport_retirement_test.go` covers active bodies, idle/pending owners,
cross-origin reuse, genuine critical errors and both replacement/retirement
orderings after an early-generation rejection.

## Optional UDP read batch capacity

`Transport.ReceiveBatchSize` selects one read's maximum message count (1–255);
zero preserves native defaults. `udp.Wrap` validates packet size and batch count
before touching the socket, even for an already wrapped connection. Existing
`udp.Conn` implementations retain their own reader. Basic sockets and Darwin's
single-message path remain single-read. OOB platforms allocate the selected
message/buffer arrays once before the first read, preserve each packet buffer's
ownership across partial refills, and retain ancillary ECN/address parsing.
This setting does not alter advertised transport parameters, send grouping,
congestion control, ACK policy or pacing. Browser values live only in Surf.
