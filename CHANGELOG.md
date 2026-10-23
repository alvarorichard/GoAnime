# GoAnime Release Notes - Unreleased

## Windows Terminal by default

- **A standalone launch reopens in Windows Terminal**: a double-click, Start menu entry or desktop shortcut that lands in a classic conhost now hands the same executable, arguments and working directory to `wt.exe -w -1 new-tab`. The TUI gets color, UTF-8 and font fallback instead of degrading to ASCII. The decision is narrow on purpose: only a classic console window (not a ConPTY host such as Windows Terminal, VS Code or WezTerm) with GoAnime as its sole attached process. A shell keeps its console, because the user already picked a terminal; pipes, CI and scripts stay headless.
- **A missing Windows Terminal installs itself once**: `winget install --id Microsoft.WindowsTerminal` runs per user and streams its progress, then GoAnime waits for the `wt.exe` app execution alias and reopens. An elevated launch skips winget (a setup started as administrator could land the Store package in another account) but still unpacks the private copy, which needs no privileges. A disabled alias is detected instead of reinstalling on every launch, and a missing `winget` falls back to the official unpackaged x64 Windows Terminal build **embedded in `goanime.exe`** (12.1 MB, SHA-256 pinned and test-enforced), unpacked into `%LOCALAPPDATA%\GoAnime\WindowsTerminal`. That is what makes the handoff work in Windows Sandbox without winget and without network access; other architectures, or a failed extraction, fall back to downloading the official archive. Without any usable copy, GoAnime stays in conhost.
- **The installer launches GoAnime as the original user**: the postinstall entry uses `runasoriginaluser` instead of `shellexec`, so the app does not inherit setup's elevated token and its private terminal copy lands in the profile of the user who ran the installer.
- **Arguments survive the handoff**: `--` keeps option-like GoAnime arguments (`--debug`, `--update`, `-h`) out of wt's parser, and literal semicolons are escaped because wt reads them as command separators even after `--`.
- **`GOANIME_TERMINAL=conhost` opts out**, documented in both READMEs. Receipt: `wt.exe -w -1 new-tab -d <dir> -- goanime.exe` put the process under `WindowsTerminal.exe` on 1.25.2733.0, and the decision table is pinned by tests.

## A real Windows icon

- **The installed program wears the dragon**: a multi-size `docs/assets/logo.ico` (16, 20, 24, 32, 48, 64, 128, 256 — LANCZOS from the transparent profile-dragon art at 2048) is linked into `goanime.exe` through a committed `cmd/goanime/rsrc_windows_amd64.syso`. Explorer, the taskbar and shortcuts now show the crimson sail instead of the blank generic square; the silhouette was chosen because it reads on dark and light desktops alike, down to 16 px, in the console's near-black + crimson identity.
- **The embedding cannot break the other platforms**: the `.syso` carries the `_windows_amd64` suffix, so `go build` links it only for windows/amd64; linux/darwin/freebsd builds of `cmd/goanime` and per-GOOS `go vet` run green with the file in-tree (machine-checked). Regenerating uses only Pillow: `go run github.com/akavel/rsrc@v0.10.2 -ico docs/assets/logo.ico -o cmd/goanime/rsrc_windows_amd64.syso`.
- **The Inno Setup installer wears it too**: `build/install.iss` compiles with `SetupIconFile=..\docs\assets\logo.ico` (was commented out — installers used to ship with Inno's generic icon), and the desktop-shortcut task is checked by default so the dragon lands at desktop install time; shortcuts and the uninstall entry already inherit the exe's embedded icon. Receipt: compiled locally with Inno 7.1.0, `ExtractAssociatedIcon` pulled the dragon out of the resulting `GoAnime-Installer-1.8.7.exe`.

## Console typography

- **CJK renders without boxes everywhere**: Japanese glosses (音声, 続き, 最近…) now go through a font-coverage probe. On a Windows console whose fonts lack Japanese — Windows Sandbox ships none — GoAnime installs an embedded `GoAnime Glyphs` subset of Noto Sans Mono CJK JP into the user's font set and re-checks at the next launch; nothing degrades into tofu boxes.
- **Decorative glyphs degrade to ASCII**: the `❯` prompt, `▶` cursor, `✓`/`◐`/`●` marks, `⁖` and the braille halftone/spinner are now drawn only when the console font verifiably covers them (probed through the console's own font and, under Windows Terminal, any installed face); otherwise their ASCII pair (`> x o -`, `= -`) takes over. Strings with no verifiable font draw nothing but ASCII.
- **UTF-8 consoles**: GoAnime sets the output code page to 65001 in classic Windows consoles (restoring it on exit), so UTF-8 status lines decode as written regardless of the host's legacy code page.
- **Every emoji left the source code**: update prompts, warning messages (`⚠️` prefixes are now plain wording), audio picker labels, perf-report headers and ratings (`★ 8.7` → `8.7/10`) are emoji-free. Scraper-title sanitization still strips the web's emoji source tags — the regexes spell them by code point (`\x{1F525}`) so no emoji glyph appears in the tree.
- **A rework-proofing scan** (`TestGlyphFontCoversEveryDecorativeGlyph`) fails the build when any future string literal carries a console glyph outside the safe classes (ASCII, latin-1, cp1252 extras, WGL4 shapes, basic arrows) without registering the rune and its ASCII fallback in `flavorRunes`.

## The font actually installs now

- **The install had two bugs and never landed a glyph on Windows**: `AddFontResourceW` was sought in `kernel32.dll` (it lives in `gdi32.dll`), so the first real install panicked instead of installing; and the embedded subset carried **CFF2 outlines**, which the classic font loader refuses — the file was written, the load returned 0, every gloss stayed a box. The font is rebuilt as TrueType (`fonts/build.py` against `NotoSansMonoCJKjp-VF.ttf`): glyf outlines load everywhere.
- **The install verifies itself on the machine**: after loading, gdi must map every Japanese probe glyph (ゴ ー 検) or the note owns up ("did not answer this session") instead of promising glyphs that will not draw. A verified install says so.
- **The install runs unless a Japanese face is already accounted for** — missing *or* unknown probe answers trigger it in both ConPTY hosts and classic consoles; pending/present never reinstall, so there is exactly one install across sessions.
- **The note is not debug gossip**: the install result, or the reason the glosses stay hidden, prints via the info logger in any mode (a copy still lands in the file log under `--debug`).
- **The guarantee runs at startup, not by handler courtesy**: `cmd/goanime/main.go` calls `tui.EnsureJapaneseDetected()` right after the flags parse — help and version stay silent, plain mode probes nothing — and a meta guard fails when that call disappears. On this machine the external GDI receipt answered `DRAWN DRAWN DRAWN` for the installed face while Consolas and a nonexistent family answer `ABSENT ABSENT ABSENT` — the coverage probe discriminates.
- **Machine-verified tests**: the Windows install tests run for real in a scratch per-user folder (`LOCALAPPDATA` pointed at a temp dir, registry writes off) — both files land, the freshly loaded face draws ゴ ー 検, a second install changes nothing, the HKCU font registry is never touched by a test, and `TestEveryFontProcLivesInItsDll` pins every font proc to the DLL it is declared in.

---

# GoAnime Release Notes - Version 1.8.7

Release date: 2026-09-14

## Highlights

- **SuperFlix survives domain rotations without a new release**: SuperFlix moved its domain five times since 1.8.6 (`.pro` → `.sbs` → `.beer` → `.baby` → `.monster`), and every move used to break playback until a new version shipped. GoAnime now finds the live host on its own, through independent layers that cover each other's failures: a manual pin, the last host that worked (remembered on disk), every alias ever observed (walked concurrently), and — only when all of those fail — a one-line pointer in this repository that maintainers can update to repair installed copies without a release. Fixes #199.
- **SuperFlix plays again, end to end**: the player CDN, signing endpoint, server picker and embed markup all changed. Streams, series, movies and Blogger-hosted titles resolve and play in mpv again.
- **Movies no longer hang before mpv opens**: remote subtitle tracks forced through the HLS demuxer kept mpv stalled for over two minutes. The same tracks are embedded in the stream, so nothing is lost.
- **SuperFlix downloads work again**: current playlist URLs were being sent to the plain MP4 downloader, which saved the playlist text as the episode file.
- **Much faster repeat plays**: replaying an episode now skips the browser entirely. Measured during development: a cached replay went from ~6.8 s to under 200 ms, and a cold resolve from ~19 s to ~8 s.
- **The bypass browser stays out of your way**: it starts minimized and only surfaces when the Cloudflare check actually needs you.
- **New AniDB source**, and **AllAnime has been removed** (see *Breaking Changes*).
- **Metadata survives an AniList outage**: when AniList disables its API, GoAnime says so plainly and falls back to MyAnimeList.

## Breaking Changes

- **AllAnime was removed as a source.** `--source allanime` now fails as an unknown source, and the `--allanime-smart` flag no longer exists. Use `--source anidb`, `animefire`, `goyabu` or `superflix`, or omit `--source` to search all of them.

## Features

### SuperFlix

- **Layered host discovery** resolves the live domain at runtime, in order:
  1. `GOANIME_SF_HOST` — manual pin, no network.
  2. The last host that verified, remembered in the user cache directory.
  3. Every SuperFlix alias ever observed, walked concurrently; the first live one wins, so a dead alias costs its own slot instead of the whole budget.
  4. The repository pointer [`superflix-host.txt`](https://github.com/alvarorichard/GoAnime/blob/main/superflix-host.txt), read only when every alias above has failed — a normal launch never contacts GitHub.
  5. The compiled default, now `superflixapi.monster`.
- **Smart offscreen solver** (on by default): the Cloudflare bypass browser starts minimized and is brought on screen only when the challenge needs a human — when Cloudflare reports a load failure or the check stops clearing — then closes when the solve ends. `--sf-window` restores the always-visible window.
- **Challenge retry**: when the Turnstile widget fails to load ("Não foi possível carregar a verificação"), the solver presses the page's own retry control instead of waiting out its timeout.
- **Browser-free re-signing** through the player's current `/layer/` endpoint, which replaced the retired `getVideo` endpoint. The legacy endpoint remains as a fallback.
- **Signed stream URLs are cached** with the Referer and User-Agent they are bound to, and verified with a single probe before reuse.
- **Blogger-hosted titles** are recognized and handed to the Blogger playback path instead of being treated as a FirePlayer host.

### Sources

- **AniDB source** (`anidb.app`): search, episode listing and stream resolution, included in the daily source-health checks. `GOANIME_ANIDB_LANG` selects subtitled (`jpn`/`sub`, default) or dubbed (`eng`/`dub`) releases.

### Metadata

- **AniList outage handling**: when AniList answers with its "API temporarily disabled" notice, the lookup names the outage instead of reporting a bare `403 Forbidden`, stops re-asking for the rest of the session, and falls back to MyAnimeList (Jikan) for titles, cover art, MAL id, genres and synopsis. Ordinary failures such as rate limits are still treated as transient.
- Jikan requests retry transient failures (5xx, 429) with a backoff above Jikan's published rate limit.
- Season lookups use the same live SuperFlix host as the player, so they follow a domain rotation too.

## Bug Fixes

### SuperFlix playback

- Fix playback breaking on every domain rotation: Go's HTTP client downgraded the player's `POST` requests to `GET` across the redirect, so `/player/bootstrap` returned HTML (#199).
- Fix every stream being rejected with `403` by the player CDN. It now requires the player's `/video/<hash>` page as Referer, the exact User-Agent that obtained the signed URL, the browser's `Accept-Language`, and the `Sec-CH-UA-Mobile`/`Sec-CH-UA-Platform` client hints. The liveness probe was sending incomplete headers and discarding working streams as "dead hosts".
- Fix the automatic server pick never clicking, after the server chooser switched to `.player_select_item` cards.
- Recover the player host and content hash from the captured stream request now that the player no longer calls `getVideo`.
- Fix a spurious "the chosen server failed" warning on healthy servers, caused by the retired `getVideo` endpoint answering `403`.
- Fix Blogger-hosted titles failing with a `404` from a FirePlayer request sent to `blogger.com`. The bogus cache entries this left behind (`host=blogger.com`, `hash=video.g`) are now discarded on read.
- Fix a solve that could hang for minutes: a Playwright protocol round-trip was made on the event dispatch goroutine.
- Remove a fixed 8-second wait for an endpoint that no longer exists, which had been added to every play.
- Fix Cloudflare challenge detection misclassifying `404` responses as challenges.
- Poster images served as direct TMDB URLs are now upgraded to `w500`, like proxied ones.

### Player

- Fix mpv never opening on some SuperFlix movies ("O Fim da Rua"): `--demuxer-lavf-format=hls` applies to every file mpv opens, so each external subtitle failed to open, and SuperFlix passed 27 remote ones. External subtitle files are no longer passed when the HLS format is forced; the same tracks are embedded in the stream and remain selectable.
- Send each CDN header as its own `--http-header-fields-append`, so the comma inside `Accept-Language` is no longer split into malformed fields, and replace mpv's default `libmpv` User-Agent with the one the stream was signed for.

### Downloads

- Fix SuperFlix downloads saving the playlist text as the video. Playlist detection still required a `/cdn/hls/` path segment the player no longer uses; it now keys on a trailing `master.txt`, matching playback. The HLS check is also consistent across all three download paths (addresses #193).
- Fix completed downloads being reported as failed with `read |0: file already closed`: the ffmpeg progress pipe was closed by `Wait` while still being read.
- Fix HTTP/2 fallback issues and improve ffmpeg integration for SuperFlix downloads, including installing a bundled ffmpeg when none is on `PATH`.
- Headers required by the SuperFlix CDN are now passed to ffmpeg and ffprobe as well.

### Other

- Handle Blogger RPC error responses and status codes, so a dead video token fails fast instead of being retried.
- Unit tests no longer make live network calls: SuperFlix host discovery and the AniList wrapper test previously reached real servers and leaked state between tests.

## Improvements

### Security

- Host discovery only follows redirects inside the `superflixapi.*` domain family, and only accepts a host whose page proves it is SuperFlix (its page title, or a genuine Cloudflare challenge). A retired domain re-registered as a parking page cannot be adopted.
- Hosts read from the repository pointer or from the remembered-host file go through the same family check and page verification. The remembered-host file is written with `0600` permissions.

### Diagnostics & Error Handling

- `ErrAniListAPIDisabled` identifies an upstream AniList outage; when the MyAnimeList fallback also fails, the error names both.
- Host-discovery failures report every alias that was tried and why each one failed.

### Testing & CI

- New CI gates: a goroutine leak gate using Go 1.27's `goroutineleak` profile, and a differential fuzz test guarding the JSON decoder migration (`internal/util/jsonx`). The 66.0% coverage gate remains.
- Offline tests drive every host-discovery layer with the real SuperFlix domain names, including dead, parked and hanging aliases.
- Real-browser tests cover the challenge retry button and the minimize/reveal behavior.
- Regression tests pin the CDN header contract, download routing for every observed playlist URL shape, playback and download agreeing on playlist detection, and subtitle handling under a forced HLS demuxer.
- `go-critic` was removed from CI; `golangci-lint` updated to v2.13.1.

### Repo Hygiene

- Add [`superflix-host.txt`](https://github.com/alvarorichard/GoAnime/blob/main/superflix-host.txt), the maintainer-updatable pointer used by host discovery. A test fails if it disagrees with the compiled default.
- Drop stale `go.sum` entries left by earlier dependency bumps.

## Dependencies

**Toolchain:** Go `1.26.5` → **`1.27.1`** (also bumped in the CI, release and source-health workflows).

**Direct dependencies updated:**

| Module | From | To |
|---|---|---|
| `charm.land/bubbles/v2` | `v2.1.1` | `v2.2.1` |
| `charm.land/bubbletea/v2` | `v2.0.8` | `v2.0.9` |
| `charm.land/lipgloss/v2` | `v2.0.5` | `v2.0.6` |
| `charm.land/log/v2` | `v2.0.0` | `v2.0.1` |
| `github.com/PuerkitoBio/goquery` | `v1.12.0` | `v1.13.0` |
| `github.com/enetx/g` | `v1.0.225` | `v1.1.1` |
| `github.com/enetx/surf` | `v1.0.201` | `v1.0.206` |
| `github.com/lrstanley/go-ytdlp` | `v1.3.5` | `v1.5.2` |
| `github.com/mattn/go-sqlite3` | `v1.14.48` | `v1.14.52` |
| `github.com/mxschmitt/playwright-go` | `v0.6100.0` | `v0.6201.1` |
| `github.com/stretchr/testify` | `v1.11.1` | `v1.12.1` |
| `golang.org/x/net` | `v0.57.0` | `v0.59.0` |
| `golang.org/x/sys` | `v0.47.0` | `v0.48.0` |
| `golang.org/x/term` | `v0.45.0` | `v0.46.0` |

**Notable indirect updates:** `quic-go/quic-go` `v0.60.0` → `v0.62.0`, `klauspost/compress` `v1.19.0` → `v1.20.0`, `andybalholm/brotli` `v1.2.2` → `v1.2.4`, `golang.org/x/crypto` `v0.54.0` → `v0.57.0`, `go.mongodb.org/mongo-driver` `v1.17.9` → `v1.17.10`; `gopkg.in/yaml.v3` replaced by `go.yaml.in/yaml/v3`.

**GitHub Actions:** `actions/checkout` v7, `actions/setup-go` v7, `actions/upload-artifact` v7, `actions/download-artifact` v8, `actions/dependency-review-action` v5, `github/codeql-action` v4, `softprops/action-gh-release` v3, `KSXGitHub/github-actions-deploy-aur` v4.2.0.

**Watch list:**

- This is the first release built with the updated `upload-artifact`/`download-artifact` major versions; check that the release job collects every platform binary.
- `mxschmitt/playwright-go` is coupled to a matching browser download at runtime, so it should not be bumped blindly.
- `charmbracelet/ultraviolet` is pinned to a pseudo-version (untagged commit) and needs manual review on each bump.

---

## Environment Variables & Flags

New in this release:

| Variable / Flag | Effect |
|---|---|
| `GOANIME_SF_HOST` | Pin the SuperFlix host (bare host or full origin), skipping discovery |
| `GOANIME_SF_OFFSCREEN` | Bypass browser starts minimized — **on by default**; set `0`, `false`, `no` or `off` to always show it |
| `GOANIME_ANIDB_LANG` | AniDB release language: `jpn`/`sub` (default) or `eng`/`dub` |
| `--sf-window` | Always show the bypass browser window |
| `--sf-offscreen` | Keep the bypass browser minimized (now the default; kept for existing commands) |

Removed: `--allanime-smart`.

---
