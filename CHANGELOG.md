# GoAnime Release Notes - Version 1.8.6

Release date: 2026-07-24

## Highlights

- **Source Registry Architecture (Model C)**: Sources are now self-describing. Each provider implements `source.Source` with a `Describe()` descriptor, and video/episode resolution dispatches through a registry instead of hardcoded branching. New optional capability interfaces — `Seasoned` (season-aware sources) and `BrowserGated` (sources needing a browser warm-up) — let the dispatcher adapt per source without special-casing call sites.
- **Scraper Package Restructure**: `internal/scraper` split into `internal/scraper/netx` (transport, SSRF guard, diagnostics, circuit breaker) and `internal/scraper/providers/{allanime,animefire,goyabu,superflix}`. The monolithic `unified.go`/`media_manager.go` pair was replaced by `manager.go` with lazy, `sync.Once`-guarded adapter loading. Public API preserved.
- **SuperFlix Revival with Headless Turnstile Auto-Solver**: SuperFlix is back and fully automatic. A Playwright-driven solver clears Cloudflare managed challenges with zero human interaction, backed by a stream cache, a dedicated transport with `Retry-After` handling, TVmaze episode mapping, and next-episode prefetching.
- **AllAnime Dynamic Key Derivation**: The AllAnime client no longer relies on a static key. It derives the AES key per epoch from the referer page (`epoch` + `partB`) masked against the entry bundle, and caches it — used for both the `aaReq` token and the `tobeparsed` blob.
- **Concurrent Multi-Source Search**: Search now fans out across all enabled sources in parallel with error tolerance — a dead source degrades the result set instead of failing the query.
- **Manual Source Kill-Switch**: `GOANIME_DISABLED_SOURCES` / `GOANIME_ENABLED_SOURCES` turn sources off or on without a rebuild, complementing the automatic per-source circuit breaker.
- **New TUI Layer**: Native anime-results screen, item picker, shell and theme, plus Windows ANSI/VT processing and color-profile detection.
- **Dependency & Toolchain Refresh**: Go 1.26.5, Playwright migrated to `mxschmitt/playwright-go`, and project-wide dependency bumps (see *Dependencies*).
- **CI Hardening**: CodeQL scanning, Dependabot, and a daily live source-health workflow added; the standalone coverage workflow was folded into `ci.yml` with a hard **66.0%** coverage gate.

## Features

### Architecture & Sources

- Introduce `internal/api/source` with the `Source` interface, `Descriptor` struct, and resolution logic that prioritizes explicit source fields, media types, tags, and URL patterns.
- Migrate AllAnime, AnimeFire, Goyabu and SuperFlix to the self-describing model; registration supports both legacy and new paths during the phased migration, guarded by anti-drift tests.
- Add `Seasoned` and `BrowserGated` capability interfaces; browser-gated sources are warmed up (`WarmUp`) before a video URL is fetched.
- Add `internal/api/providers/dispatch` with `FetchEpisodes` and `FetchStreamURL`, routing episode and stream retrieval through the registry.
- Add result tagging (`internal/api/providers/tagging.go`): language tagging and source identification on search results.
- Implement concurrent search across multiple sources with per-source error tolerance.
- Add manual source kill-switch (`internal/util/source_toggle.go`), parsed in the leaf `util` package so both the dispatch and search layers honor it without an import cycle.
- Add strict-source handling (`GOANIME_STRICT_SOURCE`) and URL-only resolution for video URL retrieval.
- Add `internal/api/source/enablement.go` and `capabilities.go` with dedicated tests.

### SuperFlix

- Headless Cloudflare Turnstile auto-solver driven by Playwright (`browser.go`, `cf.go`), with configurable behavior via `GOANIME_SF_HEADLESS`, `GOANIME_SF_BUNDLED`, `GOANIME_SF_CHROME_CHANNEL` and `GOANIME_SF_MASK`.
- Stream cache with CDN liveness probing (`streamcache.go`), dedicated HTTP transport honoring `Retry-After` (`transport.go`), and TVmaze episode mapping (`tvmaze.go`).
- Next-episode prefetching, disableable via `GOANIME_SF_NO_PREFETCH`.
- Interactive episode-flow selection and improved media handling.
- Retry logic for transient embed stream-sniff failures.
- Dedicated content-signal errors so the UI can explain *why* nothing played: `ErrSuperFlixNoServers`, `ErrSuperFlixRestricted`, `ErrSuperFlixNoEpisodeList`.
- Migrate host to the current live domain `superflixapi.pro` (previous aliases `.rest`, `.online`, `.best`, `.fit`, `.cyou`, `.lifestyle` 301-redirect, which downgraded POSTs to GET and broke `/player/bootstrap`).

### AllAnime

- Per-epoch key derivation: fetch `epoch` and `partB` from the referer page, derive the AES key with the entry-bundle mask, and cache it. No network call when a valid cached key exists.
- Split into `client.go`, `crypto.go`, `keys.go`, `search.go`, `episodes.go`, `stream.go` with regression tests for the CTR path and key derivation.

### Player & Playback

- Improved Windows support: better IPC socket handling and video-output configuration.
- Playback argument assembly for HLS streams, including HLS demuxer args selected by video source type.
- Subtitle handling: one `--sub-file=` per track, avoiding URL-separator corruption.
- Episode navigation validated against the real episode list; `ListAllEpisodes` returns real episode numbers instead of fabricated ones.
- Progress tracking and download-method selection per episode.

### TUI

- New anime-results screen with filtering, navigation and selection (`anime_results.go`).
- New generic item picker (`picker.go`), shell (`shell.go`) and theme (`theme.go`); footer instructions adapt to terminal width.
- Windows ANSI/VT processing (`vt_windows.go` / `vt_other.go`) and color-profile handling (`color_profile.go`).
- Terminal-state restore helper (`restore.go`).

### Downloads & SDK

- Workflow enrichment for `HandleDownloadRequest` with path-handling tests.
- `pkg/goanime` client updated to the registry-based fetch paths; integration tests run in the source-health workflow.

## Bug Fixes

- **Data race in status rendering**: lock the mutex while rendering helper status output.
- **Race conditions across concurrent components**: add mutex protection to `RichPresenceUpdater` state transitions; isolate episode-data mutation in `PlayEpisode`; make `ResponseCache` stop handing out mutable buffers; add synchronized accessors for global playback state and shader-mode management.
- **Terminal state corruption**: `HandlePlaybackMode` now fetches anime details and episodes sequentially instead of concurrently.
- Fix player menu and resume-dialog error handling so a failed action no longer triggers an unintended auto-advance.
- Prefer unsigned HLS playlists over secured links during video URL extraction; evict stale cache entries via CDN liveness probe.
- Fix AniList 403s by sending a non-browser `User-Agent` on AniList requests.
- Simplify season year-range logic in `season_select` (clarity fix for wrong-season inference).
- Handle empty SuperFlix episode lists explicitly instead of reporting a bare "no seasons found".
- Ensure the SuperFlix browser solver window closes on every resolve path, and release the browser earlier to avoid lingering contexts.
- Normalize line endings in shader source files for consistent processing.
- Ensure pre-warm goroutines complete before test cleanup, preventing teardown races.
- Fix the GoAnime Windows Installer link in `README` (#187, thanks @Jgmro).

## Improvements

### Security

- SSRF protection moved to `netx` with `SafeDialFunc` and `SafeScraperTransport` applied to scraper HTTP clients.
- Shader URL validation and secure fetching before any shader is downloaded.
- `crypto/rand` replaces `math/rand` for jitter (G404).
- Browser profile-dir segment sanitized to block path traversal via `GOANIME_SF_CHROME_CHANNEL` (G703), with a regression test.
- Cookie `SameSite` set when converting Playwright cookies (G124).
- CodeQL workflow added for static analysis on the repository.

### Diagnostics & Error Handling

- `netx.SourceDiagnostic` gives structured, unwrappable per-source failure diagnostics; all `ErrSourceUnavailable` references now route through `netx`.
- Origin probe added for better failure attribution, plus tighter search timeout handling.
- Source health checks call adapters directly; the old `source_circuit` implementation was removed and replaced by the `netx` circuit breaker.
- Kitsu API integration uses a configurable base URL, making it testable without live network.
- Connection pre-warming for known hosts to cut first-request latency.

### Code Quality

- Named return values on several method signatures for readability (`GetStreamURL`, `GetEpisodeStreamURL`, `GetUpscaledDimensions`, and test helpers).
- `http.NoBody` replaces `nil` request bodies throughout the scraper and client code.
- File-permission constants migrated to Go's `0o` octal syntax.
- Pre-compiled regex patterns for title cleaning and season-number inference, with benchmarks (`perf_bench_test.go`) across `api`, `naming`, `hls`, `superflix` and `util`.
- Playwright deprecated `QuerySelector` calls migrated to the Locator API.
- `golangci-lint` action updated to v9 with adjusted timeouts.

### Testing

- **Coverage: 66.8%** with `go test -short -race -count=1 -covermode=atomic ./...`, enforced in CI by a **66.0%** gate to block regressions. Project target remains **≥ 70.0%**.
- `MockScraper` added for network-free scraper testing; `wireEpisodesSeam` / `wireStreamSeam` let tests inject episode and stream data without importing the providers package.
- Live and real-browser tests gated behind `skipInCI` (`CI` / `GITHUB_ACTIONS`) and `-short`, so CI never launches Chrome or hits the network.
- Fuzz corpus added for `FuzzFitBlockRespectsBounds` (TUI block fitting).
- New daily `source-health.yml` workflow runs live source diagnostics, the Goyabu Blogger playback diagnostic, and the public SDK integration tests.
- Test files renamed to intent-revealing names (`ci_skip_test.go` → `skip_in_ci_test.go`, `source_diagnostic_extras_test.go` → `source_diagnostic_errors_test.go`).

### Documentation & Repo Hygiene

- Add `docs/ARCHITECTURE.md` and `docs/architecture/ARCH_STAGES.md` documenting the Model B → Model C source-registry migration.
- Move test planning docs into `docs/testing/` (`TEST_PLAN.md`, `TEST_STRATEGY.md`, `TEST_STAGES.md`, `TEST_PLAN_FUNCTIONS.md`).
- Add Dependabot for Go modules and GitHub Actions (weekly, minor/patch grouped).
- Untrack the committed `.DS_Store`; fold `coverage.yml` into `ci.yml`.

## Dependencies

**Toolchain:** Go `1.26.3` → **`1.26.5`** (also bumped in the CI and source-health workflows).

**Module graph:** 116 modules total — 22 direct, the rest indirect.

**Added (direct):**

| Module | Version | Why |
|---|---|---|
| `github.com/mxschmitt/playwright-go` | `v0.6100.0` | Browser automation for the SuperFlix Turnstile solver; replaces `playwright-community/playwright-go`, which is no longer the maintained fork |
| `github.com/charmbracelet/x/ansi` | `v0.11.7` | Promoted from indirect — used directly by the new TUI VT/color-profile code |
| `golang.org/x/net` | `v0.57.0` | Promoted from indirect — used directly by scraper HTML parsing |

**Added (indirect):** `deckarep/golang-set/v2 v2.9.0`, `go-jose/go-jose/v3 v3.0.5`, `go-stack/stack v1.8.1`, `go.mongodb.org/mongo-driver v1.17.9` — all pulled in transitively by `playwright-go`; `sahilm/fuzzy v0.1.3` via `charm.land/bubbles/v2/list`.

> Note: `playwright-go` pulls `mongo-driver` only for its BSON serialization used by `golang-set`. It adds ~4 modules to the graph but no additional runtime services.

**Upgraded:**

| Module | From | To |
|---|---|---|
| `charm.land/bubbles/v2` | `v2.1.0` | `v2.1.1` |
| `charm.land/bubbletea/v2` | `v2.0.6` | `v2.0.8` |
| `charm.land/lipgloss/v2` | `v2.0.3` | `v2.0.5` |
| `github.com/enetx/g` | `v1.0.224` | `v1.0.225` |
| `github.com/enetx/surf` | `v1.0.200` | `v1.0.201` |
| `github.com/enetx/http3` | `v1.0.7` | `v1.0.8` |
| `github.com/quic-go/quic-go` | `v0.59.1` | `v0.60.0` |
| `github.com/klauspost/compress` | `v1.18.6` | `v1.19.0` |
| `github.com/cloudflare/circl` | `v1.6.3` | `v1.6.4` |
| `github.com/mattn/go-sqlite3` | `v1.14.44` | `v1.14.48` |
| `github.com/gdamore/tcell/v2` | `v2.13.9` | `v2.13.10` |
| `github.com/andybalholm/brotli` | `v1.2.1` | `v1.2.2` |
| `github.com/andybalholm/cascadia` | `v1.3.3` | `v1.3.4` |
| `github.com/mattn/go-runewidth` | `v0.0.23` | `v0.0.24` |
| `github.com/refraction-networking/utls` | `v1.8.3-0.20260301…` | `v1.8.3-0.20260623…` |
| `github.com/charmbracelet/ultraviolet` | `20260511…` | `20260703…` |
| `golang.org/x/crypto` | `v0.51.0` | `v0.54.0` |
| `golang.org/x/net` | `v0.54.0` | `v0.57.0` |
| `golang.org/x/sys` | `v0.44.0` | `v0.47.0` |
| `golang.org/x/term` | `v0.43.0` | `v0.45.0` |
| `golang.org/x/text` | `v0.37.0` | `v0.40.0` |
| `golang.org/x/sync` | `v0.20.0` | `v0.22.0` |
| `golang.org/x/exp` | `20260508…` | `20260611…` |

**Watch list:**

- `refraction-networking/utls` and `charmbracelet/ultraviolet` are pinned to pseudo-versions (untagged commits) — they need manual review on each bump since Dependabot cannot reason about their semver.
- `mxschmitt/playwright-go` requires a matching browser download at runtime; its version is coupled to the Playwright driver, so it should not be bumped blindly.
- `quic-go` and `utls` sit on the network hot path (`enetx/surf`); regressions there surface as scraping failures rather than build breaks.

---

## Environment Variables

New knobs introduced in this release:

| Variable | Effect |
|---|---|
| `GOANIME_DISABLED_SOURCES` | Comma-separated source names to turn off (case- and dot-insensitive) |
| `GOANIME_ENABLED_SOURCES` | Opt into a source whose descriptor marks it `DefaultDisabled` |
| `GOANIME_STRICT_SOURCE` | Fail instead of falling back when the requested source cannot serve the request |
| `GOANIME_SF_HEADLESS` | Run the SuperFlix Turnstile solver headless |
| `GOANIME_SF_BUNDLED` | Use the bundled browser instead of a system install |
| `GOANIME_SF_CHROME_CHANNEL` | Select the Chrome channel for the solver (sanitized against path traversal) |
| `GOANIME_SF_MASK` | Mask applied by the SuperFlix solver |
| `GOANIME_SF_NO_PREFETCH` | Disable next-episode prefetching |

---
