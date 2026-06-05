# GoAnime Release Notes - Version 1.8.6

Release date: 2026-06-03

## Highlights

- **SuperFlix Source Removed**: SuperFlix integration deleted entirely. The upstream site placed every player page behind an interactive Cloudflare Turnstile captcha that cannot be solved by a TLS-impersonating HTTP client or by headless-browser automation (go-rod and Playwright, both Chromium and Firefox, were all detected and blocked). With stream extraction no longer possible, the scraper, source kind, provider registration, season-map enrichment path, and all associated tests were purged. Remaining sources (AllAnime, AnimeFire, Goyabu) are unaffected.

## Removed

- Delete `internal/scraper/superflix.go` (`SuperFlixClient` and the full player→tokens→bootstrap→source→video pipeline) and its dedicated test suites.
- Remove the `SuperFlix` source kind (`internal/api/source`), the `SuperFlixType` scraper type, the `SuperFlixAdapter`, and the `superFlixProvider` registration.
- Remove `GetSuperFlixEpisodes` / `GetSuperFlixStreamURL` and the SuperFlix dispatch branches from the enhanced API; drop the SuperFlix entry from the source breakdown.
- Remove the SuperFlix-based TMDB season-map fallback (`buildSeasonMapFromSuperFlix`) from metadata enrichment.
- Drop SuperFlix from the `--source` CLI options, PT-BR multi-source search, help text, and source/name-tag regexes.

---

# GoAnime Release Notes - Version 1.8.5

Release date: 2026-05-20

## Highlights

- **Go 1.26.3 Toolchain**: Project upgraded to Go 1.26.3. Updated all dependencies to latest compatible versions, including `enetx/surf` and `quic-go`.
- **SuperFlix Host Migration**: SuperFlix references updated from `superflixapi.online` to `superflixapi.best`. The previous `.online` host began returning 5xx errors; `.best` is now the canonical endpoint.
- **SFlix Source Removed**: SFlix integration deleted entirely (previously disabled). Dead provider code and associated tests purged.
- **Media Manager & Scraper Restructure**: Internal media manager and scraper layout refactored for cleaner provider separation. Public API preserved (no breaking changes).
- **Major Test Coverage Push**: Phases 1–14 implemented; coverage raised to **52.8%**. Phases 15–17 still in progress, targeting the remaining 165 functions at 0% across `api`, `util`, `playback`, `handlers`, `discord`, `upscaler`, `updater`, `scraper`, `providers`, `downloader`, and `SDK`.
- **Windows CI Stability**: Interactive fuzzy-finder tests now skip in CI environments without TTY. Resolves persistent 10-minute deadlocks on `windows-latest` runners caused by `tcell` `winTty.getConsoleInput` syscall blocking.

## Features

- Upgrade Go toolchain to 1.26.3; bump module dependencies project-wide.
- Update `enetx/surf` and `quic-go` to latest releases.
- Add comprehensive unit tests for scraper provider lookup, episode resolution, and upscaler pipeline (`internal/scraper`, `internal/upscaler`).
- Add player functionality tests: streaming flow, progress aggregation across episodes, error-path coverage (`internal/player`).
- Add unit tests for movie/provider entry points and media-type routing (`internal/scraper/providers`).
- Add Discord manager and handler tests: presence updates, rich-presence formatting, lifecycle (`internal/discord`).
- Implement Phases 11–14 of the test plan: every targeted exported and unexported function in those phases now has a dedicated test.
- Phases 15–17 in progress: tests being added for `api` + `util`, then `playback` + `handlers` + `discord` + `upscaler` + `updater`, then `scraper` + `providers` + `downloader` + `SDK`.

## Bug Fixes

- **Critical**: Update SuperFlix host references from `superflixapi.online` to `superflixapi.best`. The `.online` host began failing; `.best` is the new canonical endpoint. Applied across scraper code, tests, and fixtures.
- Fix Windows CI deadlock in `TestHandleUpscaleFromMenu_DoesNotPanic`, `TestAskForDownload_ReturnsValidMarker`, `TestAskForPlayOffline_DoesNotPanic`, and sibling interactive tests: each guard with `if os.Getenv("CI") != "" { t.Skip("Skipping interactive fuzzy-finder test in CI (no TTY available)") }`. Root cause: `tcell` `winTty.getConsoleInput` syscall blocks indefinitely without a console, hitting the 10-minute package timeout.
- Fix `TestSanitizeMediaTarget/plain_path_cleaned` on Windows: wrap expected path in `filepath.FromSlash(...)` so OS-native separators match `filepath.Clean` output.
- Fix global progress tracker cleanup order in CI to prevent intermittent races between teardown and lingering goroutines.

## Improvements

- Remove SFlix scraper, provider registration, and all associated tests. Dead code purged.
- Refactor media manager and scraper directory layout: cleaner separation of provider concerns; internal types reorganized without changing the public surface.
- Tighten interactive-test contract: tests that invoke `tui.Find` now check `os.Getenv("CI")` and skip when no TTY is attached, instead of relying on host-OS error behavior.
- Update `go.mod`/`go.sum` for upgraded toolchain and dependency bumps (`enetx/surf`, `quic-go`, transitive updates).
- Phases 1–14 implemented; Phases 15–17 underway, targeting the remaining 165 functions at 0%.

---
