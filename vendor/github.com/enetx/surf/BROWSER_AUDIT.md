# Desktop browser audit v1

## Fixed target

Protocol references are Chromium 158.0.8127.0/158.0.8127.18 and Firefox 160.0
(build 20261104181552), with release sources pinned in the linked reports.
Surf advertises the selected Windows Chrome/Firefox identity. Native Linux/macOS
captures establish engine behavior; they do not establish a Windows TCP/IP stack.
The Chromium-to-Chrome branding overlay is explicit and separately documented.

This audit has exactly the 24 scenarios below. A result applies only to the
listed scenario and reference version. Native random keys, GREASE values,
connection IDs, ephemeral addresses and multipart boundaries are normalized by
their documented rules; deterministic fields and ordered header values are not
discarded. Wall-clock scheduling and TCP segment boundaries are not literal
byte-equality criteria.

`match` requires the applicable source rule, current-version native evidence and
passing Surf regressions. `different` requires a retained counterexample.
`unverified` means the evidence is incomplete; it does not mean a confirmed
implementation bug. Historical 156/159 or 157 captures retain their actual
versions and cannot independently establish a 158/160 match. Passing tests alone
never promotes `unverified` to `match`.

The detailed evidence and limitations remain in [BROWSER_PARITY.md](BROWSER_PARITY.md).
Mobile/WebKit, browser DOM/resource scheduling, disk persistence, complete browser
certificate-store policy and arbitrary OS/NIC timing are outside this desktop
contract. They are separate work rather than additional acceptance conditions.

## Implementation plan

| Step | Deliverable | Acceptance |
| --- | --- | --- |
| 1 — verified | HTTP/1 header/body packing and body read sizes | Chromium's 1400-byte merge boundary and 16 KiB body reads; Firefox's 32 KiB buffered transaction stream; [56 current native full request/record matches](tests/testdata/wire/desktop_audit/README.md), streaming/Expect/Continue/body ownership and ordered headers preserved |
| 2 — verified | Firefox ordinary HTTP/2 GOAWAY | Correct idle and peer-close behavior; no duplicate GOAWAY; active responses remain readable; Chromium ordinary close remains quiet; exact code/last-stream/debug comparison on 24 current native connections, including invalid DATA stream 0; error-path and active-stream regressions |
| 3 — verified | Desktop HTTP/2 upload write boundaries | Chromium 16375-byte and Firefox buffered-source policies; 64 current native fixed-body DATA/END_STREAM and normalized TLS body-record matches; deterministic source-boundary, short-read and flow-credit regressions; arbitrary producer/kernel scheduling remains outside literal record equality |
| 4 — verified | Desktop TCP keepalive policy | Chrome/Firefox defaults and Firefox HTTP/1 lifecycle/HTTP2 disabling verified against pinned source, native socket state and long-idle capture; explicit application overrides retained |
| 5 — verified | Final audit of the fixed matrix | Source/evidence/test references and per-browser verdict for all 24 rows; full macOS/Linux race tests and vet with published HTTP v1.0.43 and the Go workspace disabled; cross-platform builds; [executed validation](tests/testdata/wire/desktop_audit/validation.json) retains remaining evidence gaps without treating them as newly required features |

Browser constants belong in Surf profiles. Existing HTTP/TLS/QUIC forks expose
generic mechanisms. Public configuration uses concrete fluent builders; no new
configuration interface is required. Each implementation step is checked before
starting the next. Publication follows completion and verification of all four
implementation steps and the final audit.

## Scenario matrix

Final per-browser verdicts and executed verification are recorded below. An
unverified row remains visibly unverified at completion.

| ID | Scenario and comparison boundary | Primary retained evidence |
| --- | --- | --- |
| D01 | TCP TLS 1.2/1.3 handshake, ALPN, normalized ClientHello and same-kernel IPv4/IPv6 SYN | [TLS/TCP](tests/testdata/wire/tcp_tls/README.md), [Chrome 158](tests/testdata/wire/chrome_158/README.md), [ALPN](tests/testdata/wire/tcp_alpn/README.md) |
| D02 | QUIC TLS ClientHello, advertised H3 SETTINGS/transport parameters and supported versions | [Chrome 158](tests/testdata/wire/chrome_158/README.md), [Firefox 160](tests/testdata/wire/firefox_160_endpoint.json) |
| D03 | Accepted/rejected ECH, authenticated retry and resumed inner ClientHello over TCP/QUIC | [ECH resumption](tests/testdata/wire/ech/resumption/README.md) |
| D04 | TLS 1.3 ticket queue consumption, expiry, client ownership and cache pressure | [Session cache](tests/testdata/wire/session_cache/README.md), [Firefox storage](tests/testdata/wire/session_cache/storage/README.md) |
| D05 | HTTP/1 navigation/fetch/forms, seven ordinary methods, redirects, empty bodies and trustworthy/insecure origins | [Native header corpus](tests/testdata/wire/header_presets/README.md) |
| D06 | HTTP/2 navigation and GET/POST fetch field order, SETTINGS, window and priority | [Chrome 158](tests/testdata/wire/chrome_158/README.md), [fresh desktop audit](tests/testdata/wire/release_235/README.md) |
| D07 | HTTP/3 navigation/fetch field order, SETTINGS, control streams and initial priority | [Chrome 158](tests/testdata/wire/chrome_158/README.md), [Firefox 160](tests/testdata/wire/firefox_160_http3.json) |
| D08 | Client Hints opt-in, Critical-CH, redirect/scenario retention and user-owned ordered fields | [Header corpus](tests/testdata/wire/header_presets/README.md), [Client Hints](tests/testdata/wire/client_hints/README.md) |
| D09 | Repeated HPACK fields, credentials, table shrink/grow/zero/oversized entries | [Sequence corpus, actual capture versions](tests/testdata/wire/browser_sequences/README.md) |
| D10 | QPACK static/dynamic fields, Huffman choice, capacity 0/128/4096/65536 and stream lifetime | [QPACK corpus](tests/testdata/wire/qpack/README.md), [cold first requests](tests/testdata/wire/cold_http3/README.md) |
| D11 | HTTP/1 small/large/streaming uploads: header/body packing and TLS record boundaries | [TLS/TCP captures](tests/testdata/wire/tcp_tls/README.md) |
| D12 | HTTP/2 small/large/streaming uploads: DATA/write boundaries, TLS records and receive credit | [TLS/TCP captures](tests/testdata/wire/tcp_tls/README.md), [flow audit](tests/testdata/wire/release_235/README.md) |
| D13 | HTTP/3 fixed/streaming uploads: body read policy, framing and flow-control behavior | [Linux load](tests/testdata/wire/linux_load/README.md), [H3 behavior](tests/testdata/wire/http3_behavior/README.md) |
| D14 | QUIC first flight: IDs, packet-number widths, CRYPTO ranges, GREASE and compatible version negotiation | [Initial recovery](tests/testdata/wire/initial_recovery/README.md), [desktop audit](tests/testdata/wire/release_235/README.md) |
| D15 | QUIC ACK/PTO and recovery: Initial/Handshake/Application loss, late original ACKs and STREAM copies | [PTO recovery](tests/testdata/wire/pto_recovery/README.md), [STREAM recovery](tests/testdata/wire/stream_recovery/README.md), [Linux load](tests/testdata/wire/linux_load/README.md) |
| D16 | HTTP/3 0-RTT accepted/rejected requests, changed peer SETTINGS and safe replay | [Early rejection](tests/testdata/wire/early_rejection/README.md), [ticket SETTINGS](tests/testdata/wire/http3_session_settings/README.md) |
| D17 | Cold/warm discovery and TCP/QUIC racing; fallback before/after informational/final response headers | [Connection history](tests/testdata/wire/connection_history/README.md), [preparation](tests/testdata/wire/connection_preparation/README.md) |
| D18 | Client-local pooling, H3 coalescing and HTTP_1_1_REQUIRED handling | [Coalescing](tests/testdata/wire/connection_coalescing/README.md), [ALPN and HTTP/1 requirement](tests/testdata/wire/tcp_alpn/README.md) |
| D19 | CONNECT/SOCKS5 routes, UDP availability, TCP fallback and address-family ownership | [Verified proxy and IPv6 regressions](tests/http3_proxy_ipv6_test.go), [proxy pools](tests/chrome_transport_audit_test.go) |
| D20 | TLS close_notify policy, idle/peer closes, repeated/concurrent close and FIN/RST observations | [TLS/TCP](tests/testdata/wire/tcp_tls/README.md) |
| D21 | Ordinary HTTP/2 idle/peer/error close and GOAWAY ordering | [TLS/TCP](tests/testdata/wire/tcp_tls/README.md), [network retirement](tests/testdata/wire/network_change/README.md) |
| D22 | Address/link/route notifications, pool retirement and retained ticket/RTT state | [Network change](tests/testdata/wire/network_change/README.md) |
| D23 | TCP keepalive idle/interval/count, Firefox short/long-lived transition and HTTP/2 disabling | [Pinned socket/connection sources](tests/testdata/wire/tcp_tls/README.md) |
| D24 | Complete large transfers under baseline, kernel delay/loss and jitter/reordering; payload hashes and capture integrity | [Linux load and its unresolved cold-path observation](tests/testdata/wire/linux_load/README.md) |


## Final evidence verdicts

These verdicts cover each complete row, rather than promoting a partial capture
or an older browser version to a current-version match. `unverified` is an
explicit evidence gap; the four implementation steps above do not promise to
remove all such gaps. In particular, a ready fixed body is not a native timed
streaming-producer reference. The new engine policies preserve streaming and
flow-control semantics, with deterministic regressions, independently of this
stricter native-evidence verdict.

| ID | Chromium 158 | Firefox 160 | Boundary / remaining evidence |
| --- | --- | --- | --- |
| D01 | match | match | Current native TCP/TLS matrix, same Linux kernel; stable hello/ALPN/SYN fields only. [Certificate transcript integrity](tests/testdata/wire/tcp_tls/README.md#tls-13-certificate-authentication) has separate signed-message regressions. Windows kernel behavior is outside this comparison. |
| D02 | match | match | Retained current native hello, advertised SETTINGS/transport parameters and supported-version values; randomized fields normalized. |
| D03 | unverified | unverified | Accepted resumed ECH has current independent references, including the supplemental Chromium 158 audit. Not every rejection/authenticated retry branch has a current native transcript. |
| D04 | unverified | different | Ticket consumption has current evidence, but not every Chrome expiry/pressure branch has a fresh native oracle. NSS and Go token representations produce different Firefox byte charges/eviction boundaries; this remains an explicit implementation limit. |
| D05 | match | match | Current seven-method/form/redirect/trustworthy/insecure header corpus; explicit identity branding/OS overlay retained. |
| D06 | match | match | Current native ordinary navigation/fetch field order, settings, window and priority; arbitrary document scheduling excluded. |
| D07 | match | match | Current native ordinary H3 navigation/fetch representation; dynamic compression sequences belong to D10. |
| D08 | match | match | Current Chromium hint/restart/redirect corpus and Firefox's unsupported-hints defaults; explicit application field ownership remains Surf's API contract. |
| D09 | match | match | Current 158/160 eleven-request HPACK sequences, repeated/shrink/grow/zero/oversized SETTINGS, sensitive fields and independent raw/semantic decode match. Firefox initial SETTINGS arrival retains the documented first-two-request race. [Evidence](tests/testdata/wire/desktop_sequences/README.md#hpack). |
| D10 | match | match | Prepared codec sequences retain their scoped 96-section/eight-stream match. New cold captures match all sixteen first sections plus 96 delayed-SETTINGS sections/eight encoder streams without preparation; pylsqpack independently decodes all 384 captured sections. Already buffered SETTINGS now precede normal header encoding. Fast later feedback races and whole discovery/packet scheduling remain excluded. [Prepared](tests/testdata/wire/desktop_sequences/README.md#qpack), [cold](tests/testdata/wire/cold_http3/README.md). |
| D11 | unverified | unverified | **56/56 current native ready-body request/record vectors match.** Native timed streaming-source record equality is not established; streaming header flush/body ownership/Expect safety is covered by regressions. |
| D12 | unverified | unverified | **64/64 current native ready-body vectors match; three paced Chromium DATA/END_STREAM/TLS vectors now also match.** Firefox Stable does not support the equivalent DOM streaming API. Arbitrary producer/credit schedules and differing DOM/Go cancellation EOF remain outside exact equality. [Evidence](tests/testdata/wire/desktop_sequences/README.md#paced-streaming-post). |
| D13 | unverified | unverified | Current fixed/streaming integrity has capture coverage, including three new paced Chromium H3 DATA/FIN plans and cancellation codes. Complete framing under arbitrary producer/credit schedules, and DOM versus Go cancellation EOF, is not equal or fully established. |
| D14 | unverified | unverified | Current first-flight/loss/HRR ranges have independent native references; the complete authenticated v1/v2 switch matrix is source/interop tested rather than fully native captured. |
| D15 | unverified | unverified | Current Initial/Handshake/Application and STREAM anchors exist. Every late-original ACK/copy-retirement branch is not independently native captured. |
| D16 | unverified | unverified | Accepted early-data references and rejection/changed-SETTINGS engine tests retain their distinct scopes. A complete controlled native rejection/replay matrix is missing. |
| D17 | unverified | unverified | Current preparation/support-history observations exist; all response-progress and impaired-path race branches lack equivalent controlled native captures. |
| D18 | unverified | unverified | Current coalescing/pool observations and source audits exist; the complete combined origin/privacy/error-code matrix is not independently native captured. |
| D19 | unverified | unverified | Trusted CONNECT/SOCKS/IPv6 Surf regressions pass. Complete equivalent current native proxy-route captures remain absent. |
| D20 | match | match | Current verified-TLS close-alert/peer/idle references; repeated and concurrent application closes retain the same single socket shutdown in regression tests. |
| D21 | match | match | New 24-connection comparison covers local/peer/protocol-error closes. Exact invalid DATA-stream-0 diagnostics match; every possible decoder diagnostic string is not claimed. |
| D22 | match | match | Retained current Linux address/link observations and policy tests. Windows/macOS host notification regressions do not assert native Windows/macOS browser packet parity. |
| D23 | match | match | Direct active-idle socket policy, Firefox transition and H2 payload/threshold match the new 16-connection evidence. Kernel rounding, browser event-loop slack and Chromium activity-triggered preface probes retain the explicit limits in the report. |
| D24 | match | unverified | Retained successful bulk baseline/loss/jitter cases validate payload/protocol/capture integrity. [New peer-side evidence](tests/testdata/wire/cold_http3_path/README.md) shows complete ClientHello reception and prepared replies for all three original Firefox failures; exact reply loss/buffering remains unestablished. Fresh paired runs pass, without relabeling older failures or claiming complete impaired-path equality. |

No new required feature is inferred from these evidence gaps. Mobile, browser
DOM/resource scheduling, complete certificate-store policy and persistence stay
outside this fixed desktop audit.

## Upstream integration revalidation, 2026-11-15

The QUIC update integrates official master `b638c8b5`, retaining the pinned
browser profiles and all existing fixture assertions. Complete macOS/Linux
race suites pass for the engine and Surf; that initial check measured 89.0%
production coverage.
The upstream inventory and owned mechanisms are in the fork's `UPSTREAM_AUDIT.md`
and `UPSTREAM.md`.

A fresh Docker installation of **unbranded Chromium 158.0.8127.0**, official
snapshot 1727526, supplements the existing 158.0.8127.18 references. Firefox's
reference remains 160.0. Local H2/H3 sequences cover capacities 0, 128, 4096 and
65536, and Chromium upload/priority cases. All 348 H3 sections are independently
decoded with pylsqpack; all 138 native/exact-input request field lists match in
order and value. Both browsers' 11 HPACK field encodings match; Firefox's leading
size-update timing can differ. Raw dynamic QPACK blocks are not all identical:
SETTINGS/feedback timing and observed native connection splits remain recorded,
including comparison with the previously published engine. These local captures
use explicit protocol/trust diagnostics and synthetic packet envelopes; they
are not ordinary discovery or OS packet-scheduling evidence.

Separately, an ordinary visible Chromium 158 discovers the public HTTP/3 endpoint
through Alt-Svc with full certificate verification. Compared with Surf's Windows
Chrome profile, normalized TLS JA3n, CID length, SETTINGS, GREASE settings, ordered
header names and transport-parameter values match. UA/platform/brand values
retain the documented Chromium-Linux to Chrome-Windows identity overlay;
transport-parameter order is randomized. The real interface capture contains
818 records with zero kernel drops. This does not establish loss/racing,
resumption, all extension bytes or socket timing parity. Exact hashes, versions,
case results and limitations accompany the release validation asset.

| Revalidation boundary | Result | Remaining evidence limit |
| --- | --- | --- |
| H3 decoding and ordered fields | 348 sections independently decoded; 138 cross-client ordered field comparisons pass across H2/H3 | Preset identity and DOM-derived request context remain explicit inputs. |
| Dynamic QPACK bytes | Byte equality varies with first SETTINGS, feedback and connection history; old/new engine comparisons retain these differences | Align the actual peer-instruction and connection histories before treating a byte difference as an encoder regression. The strict analyzer's rejected histories are retained. |
| Public Chromium 158 H3 fingerprint | TLS JA3n, advertised values and header order match | Windows Chrome branding and randomized parameter order are recorded separately; no complete packet-timing claim. |

## Supplemental current-runtime diagnostics

The latest supplemental local diagnostic matrix uses the current fork runtime
with Chromium 158.0.8127.0 and Firefox 160.0, H2 and H3 peer capacities 0, 128,
4096 and 65536. All 330 requests completed; all 110 native/exact-input ordered
field-list comparisons match at the receiving peer. Firefox's Windows User-Agent
override and Chromium's unbranded Linux identity remain explicit diagnostic
inputs.

Seven H3 cases independently decode all 231 sections using pylsqpack. The
original Firefox capacity-4096 capture retains a Wireshark decryption failure
for one Surf connection and therefore does **not** pass independent decoding.
A separate fresh repetition of that case decodes all 33 sections and matches
all 11 native/exact raw blocks. The accepted cases thus provide 264 independently
decoded sections without relabeling the failed capture as successful. Its
decryption discrepancy remains a capture/decoder investigation, not a confirmed
production defect. Strict sequence analysis also retains rejected native
connection histories; raw dynamic QPACK equality is not universal, including
Firefox capacity 65536. No reference assertion or analyzer was weakened.

After the final runtime changes, the complete matrix was repeated again: all
330 requests completed, all 110 native/exact-input ordered field lists matched,
and **all 264 H3 sections across all eight H3 cases independently decoded** with
pylsqpack 0.3.23. This successful run does not relabel earlier failed captures or
establish universal raw dynamic QPACK equality. The protocol/trust, synthetic
envelope and identity-overlay limits above still apply.

## Required coverage of Surf changes

Every new or modified Surf production-code change requires **100% coverage**.
Record the reviewed starting Git revision, measure its added/modified executable
code and test reachable branches, errors, boundaries, reset/reuse and cancellation
outcomes. Existing browser/fingerprint/header-order regressions remain required.
The overall repository percentage does not establish completion of this gate.
Platform-specific behavior requires execution on the relevant OS; Go statement
coverage alone does not prove all branch outcomes. Do not exclude owned code,
weaken assertions or change references without evidence to reach the number.
Complete known missing coverage before publishing a new code release. The HTTP,
QUIC/HTTP3 and uTLS forks follow the same rule against their pinned upstreams.

The owner's narrow architectural exception applies only to a noncritical path
that cannot be reached through the supported architecture, after reviewing its
callers, invariants and failure consequences. Record its exact location, reason
and residual risk while retaining its uncovered counters in the raw report.
Difficult, nondeterministic and platform-specific cases, and critical protocol,
security, cancellation, ownership or data-integrity behavior, remain required.
A report with documented exceptions must not be called raw 100% coverage.

QUIC v1.0.35 retains a raw **5751/5764 (99.77%)** owned-change result and records
each of its 13 noncritical architectural exceptions in the fork's
[`COVERAGE_AUDIT.md`](https://github.com/enetx/quic-go/blob/v1.0.35/COVERAGE_AUDIT.md).
HTTP v1.0.47 and uTLS v1.0.9 have complete measured owned-change statement
coverage. Surf production Go code is unchanged in this dependency/test update;
its existing fingerprint, ordered-header, ECH, early-data and browser-policy
regressions remain part of the full validation.
The current complete macOS Surf run measures 89.2% across production packages; this
overall percentage is distinct from coverage of added or modified code.
