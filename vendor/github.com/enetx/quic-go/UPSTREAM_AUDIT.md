# Upstream integration, 2026-11-15

The fork integrates official [quic-go master at b638c8b5](https://github.com/quic-go/quic-go/commit/b638c8b53731c23d9a5471ba9829fc84e93ed2ec),
including all source and test changes in the 55 commits after
[v0.65.0](https://github.com/quic-go/quic-go/tree/v0.65.0)
(`aa265c3fc6489e3bfa8071b333be95e0e887a51d`). The pin in `upstream.json`
identifies this master revision; v0.65.0 is only the previous release anchor.
This replaces the earlier selective-backport plan.

Compatible package moves, compact congestion interfaces, CRYPTO ownership,
frame-sorter changes and deterministic integration harnesses are included.
Fork-owned mechanisms preserve configurable CUBIC, ACK/PTO and STREAM recovery,
fixed accounting segments, pacing, external TLS, UDP receive sizes, dynamic
QPACK and early-generation ownership. Browser values remain in Surf.

Repository-specific CI remains authoritative. Upstream cloud credentials,
container publishing, upload/scanner services and their action/cache settings
are excluded. Fuzzer toolchain changes are included. Dependency versions and
Go 1.27 requirements are retained rather than downgraded to upstream minima.

## Commit inventory

| Upstream commit | Change | Integration |
| --- | --- | --- |
| [4ba9d8a7](https://github.com/quic-go/quic-go/commit/4ba9d8a7886e6baab5d84460e86f75ff07b91ec6) | remove leftover rawConn type assertion from Transport initialization (#6135) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [8d249447](https://github.com/quic-go/quic-go/commit/8d24944738e84a1c092aacf9790f4e4796a85c75) | move UDP connection handling and packet buffers to internal/udp (#5944) | Merged. Receive-size policy and jumbo pools move with UDP ownership; root compatibility aliases remain. |
| [ef347562](https://github.com/quic-go/quic-go/commit/ef34756266885f3617e54f274d1445f4a159cbe9) | http3: add Server.ReadHeaderTimeout (#6144) | Merged and adapted. ReadHeaderTimeout also bounds blocked dynamic QPACK headers; body/trailer reads remain unrestricted after headers. |
| [72c826bb](https://github.com/quic-go/quic-go/commit/72c826bbcf275322032151540bc5a43eedc2ecf2) | http3: reject writes exceeding a Content-Length of 0 (#6142) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [06372df4](https://github.com/quic-go/quic-go/commit/06372df45c4f44c9164f22a1ebb58fae94ee24a6) | ci: run UDP tests on FreeBSD and OpenBSD (#6139) | Repository tooling only. Keep fork CI and service ownership; no runtime source is omitted. |
| [192a63a8](https://github.com/quic-go/quic-go/commit/192a63a8201b449123c6e732f6684fb14a9266b7) | set the source address of IPv4 packets sent on FreeBSD (#6145) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [22153b8f](https://github.com/quic-go/quic-go/commit/22153b8f0194b9683823795546362f862af8d3c4) | fix the source address of IPv4 packets sent on macOS (#6146) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [d7df6395](https://github.com/quic-go/quic-go/commit/d7df639575d261746c926170a0cd1ceed221446c) | ackhandler: count ECN testing packets declared lost on PTO (#6149) | Merged and adapted. ECN loss accounting preserves retained-copy PTO policies. |
| [2042fa0a](https://github.com/quic-go/quic-go/commit/2042fa0a492a2702f8e579c8477bd46137741b89) | start the servers in TestMultiplexingServerAndClientOnSameConn (#6151) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [3dcf64dc](https://github.com/quic-go/quic-go/commit/3dcf64dcee00f7b06761f65bd28e3ec791d920a7) | use synctest for the MTU tests by making SimConn implement udp.Conn (#6152) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [0d39eb7c](https://github.com/quic-go/quic-go/commit/0d39eb7c01bb8070ceb7b3d10d2699d9d541c615) | simnet: carry ECN marks (#6153) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [66e75534](https://github.com/quic-go/quic-go/commit/66e7553417ff4b2f90d11ad40508fb18ed706e50) | add integration tests for ECN validation (#6154) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [dae2e71c](https://github.com/quic-go/quic-go/commit/dae2e71cb011421a794f950d5003ebdc4ad3d7e0) | use synctest for the ACK bundling integration test (#6156) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [ed723f9e](https://github.com/quic-go/quic-go/commit/ed723f9e9b232e955b87f754edd3c352a6415ee2) | use synctest for the NAT rebinding integration tests (#6159) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [abfd7048](https://github.com/quic-go/quic-go/commit/abfd7048d28c5b0336efaa6f83f475a94bbe8159) | http3: use an atomic for the Alt-Svc header and a plain server mutex (#6163) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [1e5c48c2](https://github.com/quic-go/quic-go/commit/1e5c48c26068869524e07d81ac6d16fdd79bac41) | http3: fix Server.Close deadlock when a handler calls SetQUICHeaders (#6162) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [82b5d81b](https://github.com/quic-go/quic-go/commit/82b5d81b06a99b63c1a435b84bbf43b1eb2d67d5) | use synctest for the MITM integration tests (#6158) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [42cba333](https://github.com/quic-go/quic-go/commit/42cba33383fe63011ebf846643421e4a7512bacc) | fix flaky TestHTTP3ListenerGracefulShutdown (#6164) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [d5a8ae6b](https://github.com/quic-go/quic-go/commit/d5a8ae6bb7e9b9138cd428c3fe93f0e36dfe9a5d) | ackhandler: subtract spurious losses from PacketsLost and BytesLost (#6169) | Merged and adapted. ACK-handler statistics preserve separate wire and congestion packet identities. |
| [ab98fc2a](https://github.com/quic-go/quic-go/commit/ab98fc2ad5e67767eb411f8478836951e40a2fb3) | move the congestion window stat from the cubic sender to the ackhandler (#6171) | Merged. ACK handler publishes native congestion-window statistics, including timeout, MTU and migration changes. |
| [254a5393](https://github.com/quic-go/quic-go/commit/254a5393706bc15f9d412cbf2ef0f43ec0084c03) | fix leak of send streams reset during a blocked Write (#6157) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [1c07d8bf](https://github.com/quic-go/quic-go/commit/1c07d8bf72238a0b7a264011f5f63554e68609e4) | congestion: remove unused Cubic code, rename cubicSender to renoSender (#6175) | Merged and adapted. Native Reno is unchanged; configurable CUBIC/PRR/pacing lives in an owned sender embedding Reno state. |
| [2f766d1f](https://github.com/quic-go/quic-go/commit/2f766d1f463260c653898644ef74087787551ea2) | congestion: remove the unused OnConnectionMigration method (#6176) | Merged for native Reno. Legacy optional sender reset coverage is tracked separately from the native interface. |
| [d45648da](https://github.com/quic-go/quic-go/commit/d45648dabb2b91745db2789ebae5cf3a239fc35d) | congestion: drop unused parameters from SendAlgorithm methods (#6177) | Merged and adapted. Owned event adapters retain the rich context required by configurable policies; the native interface stays compact. |
| [d4ad6796](https://github.com/quic-go/quic-go/commit/d4ad67960d4d8c06b9a522a09bbc0ac195f2a77c) | congestion: merge SendAlgorithmWithDebugInfos into SendAlgorithm (#6178) | Merged. Native SendAlgorithm is unified and unchanged; mocks are regenerated. |
| [325ffc4c](https://github.com/quic-go/quic-go/commit/325ffc4c3d290ddcc2b8b6eb5d14de0d80081f2c) | congestion: remove dead code from the Reno sender (#6180) | Merged for native Reno. Configurable mechanisms remain outside upstream-owned sender source. |
| [2d95e479](https://github.com/quic-go/quic-go/commit/2d95e479adb9d8bc707675b5ba5c2a53b77692b9) | congestion: use the configured InitialPacketSize in the pacer (#6182) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [057500b3](https://github.com/quic-go/quic-go/commit/057500b3f1d62538f50cb867c7e461fee7729e5c) | ci: bump CodSpeedHQ/action from 5.6.0 to 5.6.3 (#6187) | Repository tooling only. Keep fork CI and service ownership; no runtime source is omitted. |
| [d8963d1b](https://github.com/quic-go/quic-go/commit/d8963d1bd8e59d0f6220a59740c1f386b7029a12) | ci: bump docker/login-action from 4.7.0 to 4.8.0 (#6188) | Repository tooling only. Keep fork CI and service ownership; no runtime source is omitted. |
| [1515bdad](https://github.com/quic-go/quic-go/commit/1515bdad6fef5030988fce297eea7a83fdeb4671) | congestion: unexport identifiers only used inside the package (#6183) | Merged for native Reno; generic CUBIC and clock helpers are fork-owned where still required. |
| [28f46208](https://github.com/quic-go/quic-go/commit/28f46208865b1692fc9cfb2494fd0b10008a5da1) | simnet: start links in NewEndpoint, remove Simnet.Start (#6165) | Merged with a compatibility adapter. NewEndpoint starts links; the existing public Start method remains a no-op. |
| [b7964ff1](https://github.com/quic-go/quic-go/commit/b7964ff1869ca0532960086a5a97eba184385c9f) | use synctest for the HTTP integration tests (#6173) | Merged and adapted. Native synctest harness retains ordered maps, raw invalid-peer requests and bounded stream-cancellation checks. |
| [7bc7b722](https://github.com/quic-go/quic-go/commit/7bc7b722fabfff706348c7726f145fa928c49728) | add benchmarks for the frame sorter (#6184) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [8d34dbcc](https://github.com/quic-go/quic-go/commit/8d34dbcccc219375fa9fbaa9b7b22caeebd7d49d) | use a two-level slice for the frame sorter gaps, raise limit to 4096 (#6185) | Merged. Two-level gap slices and the 4096-gap limit retain the fork's non-consuming Peek hook. |
| [900ea49e](https://github.com/quic-go/quic-go/commit/900ea49ebec79020bf43678ddd0cf305cb0a1104) | ci: build OSS-Fuzz with Go 1.27.4 to fix inflated coverage counts (#6196) | Merged. Fuzzer builds use Go 1.27.4; module minimum remains Go 1.27. |
| [a25e4be7](https://github.com/quic-go/quic-go/commit/a25e4be74ad0cdbd1aaf449345a3785915b292ef) | http3: use H3_NO_ERROR when the Transport closes connections (#6193) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [21e73857](https://github.com/quic-go/quic-go/commit/21e738576159a8e5891c6cc662468817480a1870) | drop the Initial and Handshake crypto streams with their keys (#6197) | Merged and adapted. CRYPTO manager solely owns streams and retires them with keys; Initial layouts and range recovery use small hooks. |
| [6128c23e](https://github.com/quic-go/quic-go/commit/6128c23eb245ae084c45ddf685255027980efb51) | qlogwriter: log the replacement version when quic-go is replaced (#6199) | Merged. Replacement versions are reported accurately. |
| [b273c05b](https://github.com/quic-go/quic-go/commit/b273c05b8ac448b5cf2d697a679044e6165e7def) | qlogwriter: replace code_version with the implementation field (#6200) | Merged with compatibility. New implementation metadata is emitted; code_version remains as an alias for existing trace consumers. |
| [9a267db0](https://github.com/quic-go/quic-go/commit/9a267db0deaaec5eec01e907dbe9bebc7155a1b1) | ackhandler: count the last ECN testing packet as ECT(0) (#6203) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [7dfdd075](https://github.com/quic-go/quic-go/commit/7dfdd075de139b1414fe67fd484b3e11398c4bf0) | ackhandler: disable ECN after path migration (#6206) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [2f22322b](https://github.com/quic-go/quic-go/commit/2f22322bb81564fab26beefa274a6d418c8abcda) | qlogwriter: test that the qlog file type can be sniffed (#6207) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [35e0af55](https://github.com/quic-go/quic-go/commit/35e0af55765b04fadf820b0ed4bd64af0510ea95) | use synctest for the server close and accept queue tests (#6208) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [6844301a](https://github.com/quic-go/quic-go/commit/6844301a563cfc3a5c3057abb14bfd523a088ae2) | ci: use a daily Go build cache for golangci-lint (#6209) | Repository tooling only. Keep fork CI and service ownership; no runtime source is omitted. |
| [6433a6b6](https://github.com/quic-go/quic-go/commit/6433a6b6d746a5da79f78a1192069939fd78f53f) | ci: disable setup-go's built-in cache (#6210) | Repository tooling only. Keep fork CI and service ownership; no runtime source is omitted. |
| [59b8e04b](https://github.com/quic-go/quic-go/commit/59b8e04b4c7e8e315bcbf4dfb46431cf866a1a7a) | use synctest for the version negotiation integration tests (#6213) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [1d3df076](https://github.com/quic-go/quic-go/commit/1d3df0769c91901a738b9d19aa6e520d77e1771a) | ci: limit push triggers to master and release branches (#6214) | Repository tooling only. Keep fork CI and service ownership; no runtime source is omitted. |
| [d79d6e4c](https://github.com/quic-go/quic-go/commit/d79d6e4c9060ad586e73e301a33e467f1575b82e) | use synctest for the early data integration test (#6215) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [6dd1dae4](https://github.com/quic-go/quic-go/commit/6dd1dae4a91259e568d5e9ae29c0ace33d640c80) | use synctest for the connection migration integration test (#6216) | Merged and adapted. Path validation remains tested separately from application STREAM routing. |
| [ae2ac386](https://github.com/quic-go/quic-go/commit/ae2ac386c87dc555f2f5f8803f3254406a828c99) | stop single-use transports when the last connection is destroyed (#6218) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [45ee4506](https://github.com/quic-go/quic-go/commit/45ee45062a17abb0fb8a94d26e2f884dfeef7065) | ackhandler: revalidate ECN after path migration (#6219) | Merged and adapted. ECN revalidation retains configurable congestion and per-space accounting. |
| [e18ca1d8](https://github.com/quic-go/quic-go/commit/e18ca1d898e9cfc17e4eddbf532a220c2d6cd346) | move the ListenAddr socket close check to the address helpers test (#6221) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [eab9cbb0](https://github.com/quic-go/quic-go/commit/eab9cbb019a08acb21770f4cdb541113d3ad2ea6) | add configuration for Anthropic's OSS Scanner (#6222) | Repository tooling only. Keep fork CI and service ownership; no runtime source is omitted. |
| [5a6c94ec](https://github.com/quic-go/quic-go/commit/5a6c94ec6cec49d927eef5881e322f7572cc0b0d) | fix stall when the send queue frees up before the run loop checks (#6223) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |
| [b638c8b5](https://github.com/quic-go/quic-go/commit/b638c8b53731c23d9a5471ba9829fc84e93ed2ec) | remove outdated Listen doc about closing established connections (#6226) | Merged, including its source/tests or documentation; fork import names and existing extension hooks are retained. |

## Integration checks

Tests retain old custom vectors alongside the current upstream suite. Targeted
regressions additionally cover default Reno with persistent congestion, blocked
dynamic QPACK header deadlines and successful connection reuse, disabled header
timeouts, dynamic trailers after the header phase, source-address control data,
receive-buffer ownership and native/configurable sender equivalence.

Public API review compares packages rather than source filenames. Windows socket
constants and the UDP maximum retain their values through aliases;
OOBCapablePacketConn retains its complete interface. Simnet.Start remains
available. ReadHeaderTimeout and simulator ECN/UDP methods are additive.
The gomock-only build-tag aliases are generation details, not runtime exports.

See [UPSTREAM.md](UPSTREAM.md) for mechanism ownership, required hooks and the
update procedure. Browser fixture tests are regression checks against pinned
captures, not a claim of new native-browser measurements or complete parity.
