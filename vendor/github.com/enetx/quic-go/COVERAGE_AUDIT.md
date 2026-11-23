# Coverage of maintained changes

The HINTS-02 candidate audit measures executable additions and modified hooks against the
normalized upstream revision in `upstream.json`, using Go 1.27.4. It counts whole
coverage blocks intersecting changed Go tokens, conservatively including adjacent
statements. Repository-wide upstream coverage is a different measure.

The raw result is **6009 / 6022 statements, 99.78%**. The 13 remaining statements
are retained in the raw profile and reviewed below under the owner's exception
for noncritical paths unavailable through the supported architecture. This is
**not raw 100% coverage**. No production file, assertion or currently uncovered
counter was excluded to obtain it.

Coverage combines complete macOS/Linux runs, Surf integration and applicable
Windows execution. Reused counters must describe unchanged production files with
identical coordinates. After a refactor, complete current-file counters replace
obsolete coordinates. CI saves raw profiles, Go version, Git revision and module
versions separately for each OS.

The v1.0.36 receive-batch additions and modified hooks have **26/26 measured
statements** against v1.0.35. Full macOS/Linux race tests cover default and custom
capacities, full/partial refills, ECN and retained-buffer ownership, both bounds,
single-read sockets and actual queued UDP. Linux socket-buffer forcing tests
require CAP_NET_ADMIN; the full UDP package passes with that capability. A
separate strace observes native recvmmsg capacity/return 45. Source-identical
baseline blocks are retained only when every line of the block maps unchanged;
modified blocks require fresh counters. The 13 reviewed exceptions are unchanged.

## Authenticated peer settings and request preparation

The reviewed starting revision is `140bc59be3d6f8bd2bd058ef1c6f2624aca1de4e`
(v1.0.36). The optional authenticated peer-settings and selected-connection
header-preparation additions, shared validation correction and modified hooks
have **257/257 measured statements** against that revision. These counters are
fresh measurements of the current production source. The complete owned-fork
inventory remains visible against upstream
`b638c8b53731c23d9a5471ba9829fc84e93ed2ec`; its same 13 exceptions are unchanged.
No new exception is requested.

The retained `.maintenance/hints02/README.md` and reports record concurrency,
request/generation cancellation, retired generation publication, actual accepted
and rejected 0-RTT, warm initialization, reserved SETTINGS 2 through 5, shared
QPACK ownership, independently decoded request headers, field limits and
exactly-once unsent body ownership. Local startup/settings/header errors preserve
the original cause and are terminal, including when a secondary connection-close
error races with them. The Surf SETTINGS-generator regression passes ten race
iterations. Native unopened GOAWAY/network failures retain existing retry/body
ownership. No browser capture or broader parity claim is introduced.

Go 1.27.4 macOS full race tests use ordinary per-package coverage. Linux full race
tests run without module-wide instrumentation, plus current HTTP/3 package
coverage. Module-wide `-coverpkg=./...` instrumentation hit unchanged pure-QUIC
integration fixtures' two-second deadlines; a subsequent ENOSPC-interrupted
coverage write is also excluded. Neither failed attempt supplies acceptance or
coverage counters. No timeout or assertion was weakened. The final audit merges
successful current measurements with the retained v1.0.36 profile only for
source-identical blocks. Published HTTP v1.0.49 is used without local replaces.
Windows HTTP/3 cross-compilation checks this API change; unchanged platform
execution evidence remains separately identified, without calling a cross-build
execution.

## Reviewed defensive statements

Each row records the source invariant and the risk if a future update invalidates
it. These exceptions do not extend automatically to new code or difficult tests.

| Location | Statements | Invariant, evidence and retained risk |
| --- | ---: | --- |
| `internal/qpack/encoder.go:257` | 1 | Table size equals the sum of entry sizes and never exceeds capacity. Evicting the complete table leaves zero, so another eviction cannot be required. Codec tests cover eviction, pinned entries, rollback and limits. The guard protects future accounting changes. |
| `internal/qpack/encoder.go:397` | 1 | The pinned official codec contains every static index 0 through 98 requested here. `TestEncoderStaticCompatibility` independently exercises all 99. A future incompatible codec must fail this guard and those tests. |
| `crypto_stream.go:196` | 1 | Cuts describe disjoint SNI/ECH locations or one unused slot; with neither extension present, sorting is skipped. Equal starts are unavailable from the current parser. Equality remains necessary for the comparator contract and future sort/layout changes. |
| `stream_range_recovery.go:157` | 1 | Fragment preparation subtracts queued ranges and admits FIN once, leaving distinct queued starts. The byte-model tests exercise overlapping loss/probe ranges. Equality remains safe if a future scheduler or sort supplies an equal entry. |
| `stream_range_recovery.go:255` | 1 | Useful lost bytes/FIN create a queue, preventing completion. ACK and cancellation retire already satisfied transmissions and perform completion themselves. Reliable-reset and late-loss regressions require exactly one notification. This guard protects future native completion rules. |
| `http3/client_initial_settings.go:75`; `http3/wire_config.go:207` | 2 | These writers own concrete native `*quic.SendStream` values. A short Write returns an error, never a short count with nil error. Cancellation, deadline, rejection and partial-write regressions exercise the native outcomes. The guards retain the Writer contract if native behavior changes. |
| `http3/client_session_settings.go:325` | 1 | Binding precedes publication/peer input. A fresh validated encoder receives its first bounded remembered limits; at most one capacity instruction enters a queue of at least 32 slots. Limit changes, overload and invalid cache metadata have separate tests. Future initialization reordering must revalidate this guard. |
| `http3/qpack_encode.go:24`; `http3/request_writer.go:364` | 2 | The official static encoder writes to its constructor-owned `*bytes.Buffer`, whose writes cannot fail. Its valid representations also satisfy the never-index adapter. Dynamic encoding collects fields and reports errors after enumeration. Malformed representations/dynamic failures have separate tests. Replacing the concrete sink/codec requires new error-path tests. |
| `http3/transport.go:689` | 1 | Retired owners contain concrete native connections. `Conn.CloseWithError` waits for closure and always returns nil; the nil-connection owner also returns nil. Close, pending dial, active body and retirement regressions cover ownership. A future fallible Close requires exercising propagation. |
| `initial_packet_protection.go:185` | 1 | The concrete `wire.AckFrame.Append` has no non-nil error return. ACK vectors, Initial shuffling/padding and frame-encoding failures remain tested. A future fallible ACK encoder requires testing this guard. |
| `packet_packer.go:442` | 1 | Initial-only application layout callbacks never run for Handshake. Handshake recovery uses native ranges derived from its retained transcript and has no application layout callback. Its validation cannot fail under those invariants. Initial layout failures and Handshake recovery are tested; future Handshake layout controls require an error regression. |

No reachable authentication, version switching, cancellation, replay ownership,
accepted-count, flow-control or byte-integrity scenario is exempted. These have
explicit regressions beyond statement coverage, including actual
encrypted UDP loss and TLS early-data rejection. Eager QPACK backpressure uses a
real three-byte connection window exhausted by the control prelude; it is covered.

## Non-instrumented compatibility surface

`ack_policy.go`, `congestion_control.go`, `stream_recovery.go` and
`internal/ackhandler/interfaces.go` contain aliases/constants or interfaces.
`tlsquic/conn.go` contains interface/function types and a native compile assertion.
`udp_windows.go` retains two aliases checked by a native Windows test.
`http3/control_frame.go` contains public data declarations and package error
initializers, without an instrumented body. These files hide no missing runtime
body. API, platform builds and wire regressions check their contracts; a
cross-compile alone is not reported as execution.

Browser comparisons remain separately scoped in Surf's `BROWSER_AUDIT.md`.
Statement coverage does not establish complete browser, timing or packet parity.
