# GoAnime Release Notes - Version 2.0.0

Release date: 2026-10-28

## Highlights

- **157 commits since 1.8.7** (2026-09-14), by 6 contributors: 613 files changed, +75,280 / −31,255 lines. 331 of those files are tests: the unit, cascade and mutation tests added around each source are what let this many providers move at once.
- **StartFlix replaces SuperFlix, and GoAnime ships no browser anymore.** SuperFlix and its Playwright/Cloudflare-bypass stack are gone (`playwright-go` left `go.mod`). Movies and TV now come from StartFlix, TopCine and Pobreflix over plain HTTP, with an id-keyed vidsrc/2Embed fallback for titles whose panel players are dead.
- **Four new sources**: TopCine, Pobreflix, AnimeDrive and AnimesDigital, and **HiAnime replaces the dead anidb.app**. AnimeFire is read through its JSON API. Goyabu is opt-in while Cloudflare gates it.
- **Updates are verified**: `--update` checks every download against the release's SHA-256 manifest, and the manifest itself is **signed with minisign** (the public key is embedded in the binary), so a compromised release cannot swap a binary and its checksum together. A build with no key refuses every update.
- **Windows is a first-class platform**: standalone launches reopen in Windows Terminal (installed or unpacked from a copy embedded in the exe when missing), the installer is per-user and needs no administrator, the exe has a real icon, and Japanese glosses render without boxes.
- **Pobreflix survives its own redesign**: on 2026-10-27 every page started pointing at a Turnstile-gated player. GoAnime no longer hands that challenge page to mpv; it takes the fallbacks and says why the Portuguese dub cannot play.
- **New terminal identity**: the TUI, downloads and upscales draw the goanime.online look; the details panel fills from what each source reports.
- **The module is now `github.com/alvarorichard/Goanime/v2`** (see *Breaking Changes*).

## Breaking Changes

- **The Go module path is now `github.com/alvarorichard/Goanime/v2`.** Go requires the `/v2` suffix for a v2 module, so a `v2.0.0` tag without it would break `go install` and leave `@latest` on 1.8.7. Library users must update imports from `github.com/alvarorichard/Goanime/pkg/goanime` to `github.com/alvarorichard/Goanime/v2/pkg/goanime`, and source installs become `go install github.com/alvarorichard/Goanime/v2/cmd/goanime@latest`. The v1 module stays available at the old path for existing consumers.
- **SuperFlix was removed.** `--source superflix`, `--sf-window`, `--sf-offscreen` and the `GOANIME_SF_*` variables no longer exist. Use `--source startflix`, `topcine` or `pobreflix`, or `--source movietv` for all movie/TV sources.
- **anidb.app was replaced by HiAnime.** `--source anidb` is still accepted and means HiAnime; `GOANIME_ANIDB_LANG` is replaced by `GOANIME_HIANIME_AUDIO`.
- **Goyabu is off by default.** Its Cloudflare gate cannot be passed without a browser; `GOANIME_ENABLED_SOURCES=goyabu` turns it on where Cloudflare does not challenge.
- **`--update` refuses a release it cannot verify**: a release without `checksums-sha256.txt` and its `.minisig` signature is not installed.

## Features

### Sources

- **StartFlix** (`startflix.biz`) is the movie/TV source, replacing SuperFlix. Its panel plays Byse/Filemoon hosts over plain HTTP; downloads are named after the official English title, found keylessly; the best quality on offer is always picked; dead movie panels are tracked and filtered from later searches.
- **Abyss playback through a local stream proxy**, with chunked fetching, a shared part cache with per-reader read-ahead, stale-fetch cancellation, and recovery of truncated file tails (a movie no longer stalls at 99%).
- **TopCine** (`topcine3.site`) is a movie/TV source next to StartFlix, played through the panel by TMDB/IMDb id.
- **Pobreflix** (`pobreflixtv.gay`): search asks the site's `/api` catalog first (the real name and slug) and falls back to the sitemap corpus. Current `/filme/`, `/serie/` and `/assistir/` pages are read from their embedded JSON, and playback picks the embedded public MP4 server by audio, quality and resolution when the page offers one. `GOANIME_POBREFLIX_URL` repoints the catalog without a release.
- **AnimeDrive** (#212, by @Reennaan): search, episodes, block-based servers, quality order and an availability probe, enabled by default. `GOANIME_ANIMEDRIVE_URL` points it at a reachable mirror.
- **AnimesDigital** (`animesdigital.org`, #211, by @william-navarro) as a PT-BR anime source. `GOANIME_ANIMESDIGITAL_URL` overrides the host.
- **HiAnime** (`hianime.at`) replaces anidb.app, which went to a site-wide 503.
- **AnimeFire** is read through its JSON API, fMP4 downloads work, episodes are numbered across seasons, and the CDN playlist is checked before a stream is handed over, so a title the CDN no longer serves is reported as offline instead of failing inside mpv.
- **Fallbacks for dead panels**: an id-keyed vidsrc.sh source and a 2Embed (`vidsrc.buzz`) chain save movies whose panel hosts no longer answer (Todo Poderoso 2003 plays this way). The user is asked before the original audio replaces a Portuguese dub, with the reason.
- **Goyabu** clears Cloudflare with a shared browser where it can, and is parked off by default until the gate can be passed without one.
- **Source health** runs every source daily, including Pobreflix and a canary that reports when its player stops being gated.

### Search

- **Results are ranked by relevance** instead of provider order, and panel hosts that need it fall back to yt-dlp.
- **Provider timeouts are race-safe** (#215, by @Dhekki and #219, by @Vittuu): a parent deadline finalizes pending providers without tripping the circuit breaker, cancellation errors survive, and AnimeDrive honors the caller's deadline.
- **A failed search says what failed**, not "no anime found"; a query with no hit is an empty result, not an error.
- **The user returns to the results** when a title lists no episodes.

### Updates and distribution

- **Checksum verification**: `--update` checks every download against `checksums-sha256.txt` before anything is installed, and before a Windows zip is extracted.
- **Signed manifest**: the manifest carries a minisign signature made with an offline key; the updater verifies it against the key embedded in the binary before reading a single checksum, and fails closed. Release assets are signed by the `release` job after a reviewer approves it.
- **Recoverable Windows updates and a per-user installer**: `%LOCALAPPDATA%\Programs\GoAnime`, no administrator, `PATH` only for the user; the backup is kept until the new binary has actually started.
- **Release and AUR publishing run in a protected `release` environment**, and the Linux runners are pinned to `ubuntu-24.04`.

### Playback and downloads

- **mpv prefers `gpu-next`** with per-OS fallbacks, and Anime4K upscaling gets the same video-output fallbacks.
- **Every source's subtitles reach mpv.**
- **HLS**: references resolve per RFC 3986 and AES-128 segments are decrypted; the direct HTTP downloader never saves a playlist or a web page as a video; proxied files stay alive through a whole batch.
- **Movie playback no longer waits on Jikan**, and StartFlix titles stay away from AniList on play and download.
- **`GetAnimeSeason`** retrieves the current anime season.

### Terminal UI

- **The goanime.online identity** in the terminal: menus, downloads and upscales, and a details panel filled from what the sources report.
- **Japanese glosses render on Linux** with a font-coverage probe and advice when no CJK font is installed.

### Windows

#### Windows Terminal by default

- **A standalone launch reopens in Windows Terminal**: a double-click, Start menu entry or desktop shortcut that lands in a classic conhost now hands the same executable, arguments and working directory to `wt.exe -w -1 new-tab`. The TUI gets color, UTF-8 and font fallback instead of degrading to ASCII. The decision is narrow on purpose: only a classic console window (not a ConPTY host such as Windows Terminal, VS Code or WezTerm) with GoAnime as its sole attached process. A shell keeps its console, because the user already picked a terminal; pipes, CI and scripts stay headless.
- **A missing Windows Terminal installs itself once**: `winget install --id Microsoft.WindowsTerminal` runs per user and streams its progress, then GoAnime waits for the `wt.exe` app execution alias and reopens. An elevated launch skips winget (a setup started as administrator could land the Store package in another account) but still unpacks the private copy, which needs no privileges. A disabled alias is detected instead of reinstalling on every launch, and a missing `winget` falls back to the official unpackaged x64 Windows Terminal build **embedded in `goanime.exe`** (12.1 MB, SHA-256 pinned and test-enforced), unpacked into `%LOCALAPPDATA%\GoAnime\WindowsTerminal`. That is what makes the handoff work in Windows Sandbox without winget and without network access; other architectures, or a failed extraction, fall back to downloading the official archive. Without any usable copy, GoAnime stays in conhost.
- **The installer launches GoAnime as the original user**: the postinstall entry uses `runasoriginaluser` instead of `shellexec`, so the app does not inherit setup's elevated token and its private terminal copy lands in the profile of the user who ran the installer.
- **Arguments survive the handoff**: `--` keeps option-like GoAnime arguments (`--debug`, `--update`, `-h`) out of wt's parser, and literal semicolons are escaped because wt reads them as command separators even after `--`.
- **`GOANIME_TERMINAL=conhost` opts out**, documented in both READMEs. Receipt: `wt.exe -w -1 new-tab -d <dir> -- goanime.exe` put the process under `WindowsTerminal.exe` on 1.25.2733.0, and the decision table is pinned by tests.

#### A real Windows icon

- **The installed program wears the dragon**: a multi-size `docs/assets/logo.ico` (16, 20, 24, 32, 48, 64, 128, 256 — LANCZOS from the transparent profile-dragon art at 2048) is linked into `goanime.exe` through a committed `cmd/goanime/rsrc_windows_amd64.syso`. Explorer, the taskbar and shortcuts now show the crimson sail instead of the blank generic square; the silhouette was chosen because it reads on dark and light desktops alike, down to 16 px, in the console's near-black + crimson identity.
- **The embedding cannot break the other platforms**: the `.syso` carries the `_windows_amd64` suffix, so `go build` links it only for windows/amd64; linux/darwin/freebsd builds of `cmd/goanime` and per-GOOS `go vet` run green with the file in-tree (machine-checked). Regenerating uses only Pillow: `go run github.com/akavel/rsrc@v0.10.2 -ico docs/assets/logo.ico -o cmd/goanime/rsrc_windows_amd64.syso`.
- **The Inno Setup installer wears it too**: `build/install.iss` compiles with `SetupIconFile=..\docs\assets\logo.ico` (was commented out — installers used to ship with Inno's generic icon), and the desktop-shortcut task is checked by default so the dragon lands at desktop install time; shortcuts and the uninstall entry already inherit the exe's embedded icon. Receipt: compiled locally with Inno 7.1.0, `ExtractAssociatedIcon` pulled the dragon out of the resulting `GoAnime-Installer-1.8.7.exe`.

#### Console typography

- **CJK renders without boxes everywhere**: Japanese glosses (音声, 続き, 最近…) now go through a font-coverage probe. On a Windows console whose fonts lack Japanese — Windows Sandbox ships none — GoAnime installs an embedded `GoAnime Glyphs` subset of Noto Sans Mono CJK JP into the user's font set and re-checks at the next launch; nothing degrades into tofu boxes.
- **Decorative glyphs degrade to ASCII**: the `❯` prompt, `▶` cursor, `✓`/`◐`/`●` marks, `⁖` and the braille halftone/spinner are now drawn only when the console font verifiably covers them (probed through the console's own font and, under Windows Terminal, any installed face); otherwise their ASCII pair (`> x o -`, `= -`) takes over. Strings with no verifiable font draw nothing but ASCII.
- **UTF-8 consoles**: GoAnime sets the output code page to 65001 in classic Windows consoles (restoring it on exit), so UTF-8 status lines decode as written regardless of the host's legacy code page.
- **Every emoji left the source code**: update prompts, warning messages (`⚠️` prefixes are now plain wording), audio picker labels, perf-report headers and ratings (`★ 8.7` → `8.7/10`) are emoji-free. Scraper-title sanitization still strips the web's emoji source tags — the regexes spell them by code point (`\x{1F525}`) so no emoji glyph appears in the tree.
- **A rework-proofing scan** (`TestGlyphFontCoversEveryDecorativeGlyph`) fails the build when any future string literal carries a console glyph outside the safe classes (ASCII, latin-1, cp1252 extras, WGL4 shapes, basic arrows) without registering the rune and its ASCII fallback in `flavorRunes`.

#### The font actually installs

- **The install had two bugs and never landed a glyph on Windows**: `AddFontResourceW` was sought in `kernel32.dll` (it lives in `gdi32.dll`), so the first real install panicked instead of installing; and the embedded subset carried **CFF2 outlines**, which the classic font loader refuses — the file was written, the load returned 0, every gloss stayed a box. The font is rebuilt as TrueType (`fonts/build.py` against `NotoSansMonoCJKjp-VF.ttf`): glyf outlines load everywhere.
- **The install verifies itself on the machine**: after loading, gdi must map every Japanese probe glyph (ゴ ー 検) or the note owns up ("did not answer this session") instead of promising glyphs that will not draw. A verified install says so.
- **The install runs unless a Japanese face is already accounted for** — missing *or* unknown probe answers trigger it in both ConPTY hosts and classic consoles; pending/present never reinstall, so there is exactly one install across sessions.
- **The note is not debug gossip**: the install result, or the reason the glosses stay hidden, prints via the info logger in any mode (a copy still lands in the file log under `--debug`).
- **The guarantee runs at startup, not by handler courtesy**: `cmd/goanime/main.go` calls `tui.EnsureJapaneseDetected()` right after the flags parse — help and version stay silent, plain mode probes nothing — and a meta guard fails when that call disappears. On this machine the external GDI receipt answered `DRAWN DRAWN DRAWN` for the installed face while Consolas and a nonexistent family answer `ABSENT ABSENT ABSENT` — the coverage probe discriminates.
- **Machine-verified tests**: the Windows install tests run for real in a scratch per-user folder (`LOCALAPPDATA` pointed at a temp dir, registry writes off) — both files land, the freshly loaded face draws ゴ ー 検, a second install changes nothing, the HKCU font registry is never touched by a test, and `TestEveryFontProcLivesInItsDll` pins every font proc to the DLL it is declared in.

## Bug Fixes

### Sources and search

- Fix a malformed Byse record crashing the whole process (StartFlix).
- Fix Pobreflix handing a Turnstile-gated player page to mpv as if it were a file, and its episode pages failing to decode after the episode id became a string (2026-10-27).
- Fix Pobreflix search failing with "sitemap shards produced no catalog entries" when the sitemaps began listing pages without `.html`.
- Fix the StartFlix panel reading a single player from a script, and AnimeDrive search ignoring context cancellation.
- Fix an SSRF gap: the dial IP is checked before connecting, and link-local addresses are blocked.
- Fix a search query being slugged into a miss, and a miss being reported as a failure.

### Terminal and platform

- Fix the OSC 11 theme probe eating early keystrokes on Windows, and add fail-fast TTY checks so headless runs never block.
- Fix the home directory resolving differently across the player: `HOME` is used everywhere.
- Fix the `-o` help text pointing outside the home folder, and make the help list every flag with what it does.
- Fix Windows font probing: typed Win32 bindings, and a partial glyph answer fails closed.

## Improvements

### Security

- Go `1.27.1` → **`1.27.3`** (a security upgrade), gosec pinned to a commit that reads Go 1.27 export data, and the CI gates (gosec, govulncheck, CodeQL, golangci-lint `v2.14.0`) cleaned of existing findings.
- The SSRF guard checks the resolved IP, not just the host name.
- The release workflow signs and verifies the manifest before publishing, and refuses to publish when the signing key does not match the one embedded in the binary.

### Documentation and tooling

- README rewritten (and a demo recorded from public-domain films), plus a contributing guide, pull request template, bug-report and feature-request forms, and corrected build, update, scraping, library and diagnostics guides.
- `docs/RELEASE_READINESS.md` records the release audit and the zero-cost distribution plan.
- Removed unused CircleCI, Codacy and VS Code configs.

### Tests

- Cascade tests (one stage per hop, stopping at the first broken one) and mutation tests (`go test -overlay`, so no file is touched) guard the audio, Pobreflix and sitemap fixes: each puts back one way the bug can return and requires the guard tests to fail against it.
- Live checks follow the catalogs as they are (`GOANIME_LIVE=1`, never in CI).

## Dependencies

**Toolchain:** Go `1.27.1` → **`1.27.3`**.

**Direct dependencies:**

| Module | From | To |
|---|---|---|
| `charm.land/bubbletea/v2` | `v2.0.9` | `v2.1.0` |
| `github.com/Microsoft/go-winio` | `v0.6.2` | `v0.6.4` |
| `github.com/charmbracelet/x/ansi` | `v0.11.8` | `v0.11.9` |
| `github.com/lrstanley/go-ytdlp` | `v1.5.2` | `v1.5.6` |
| `golang.org/x/net` | `v0.59.0` | `v0.62.0` |
| `golang.org/x/sys` | `v0.48.0` | `v0.49.0` |
| `golang.org/x/term` | `v0.46.0` | `v0.47.0` |
| `github.com/tetratelabs/wazero` | — | `v1.12.0` |
| `github.com/zzl/go-win32api` | — | `v1.2.0` |
| `github.com/mxschmitt/playwright-go` | `v0.6201.1` | removed |

**Notable indirect updates:** `quic-go/quic-go` `v0.62.0` → `v0.64.1`, `klauspost/compress` `v1.20.0` → `v1.20.3`, `andybalholm/brotli` `v1.2.4` → `v1.2.7`, `ProtonMail/go-crypto` `v1.4.1` → `v1.5.2`, `golang.org/x/crypto` `v0.57.0` → `v0.58.0` (now direct, for the manifest signature).

**Watch list:**

- `charmbracelet/ultraviolet` is pinned to a pseudo-version (untagged commit) and needs manual review on each bump.
- The first release built on `ubuntu-24.04` with the signing job: check that the release carries `checksums-sha256.txt.minisig`.
- Windows builds moved to the MSYS2 `UCRT64` toolchain.

---

## Known Limitations

- **The Portuguese dub of Todo Poderoso (2003), and of most Pobreflix and TopCine titles, does not play.** Their only players sit behind a Cloudflare Turnstile that GoAnime does not solve (it ships no browser). GoAnime asks before playing the original audio in its place and names the player that stopped the dub.
- **AnimeFire lists many titles whose CDN has no file**; they are reported as offline.
- **Goyabu** needs a network Cloudflare does not challenge.
- **Windows and macOS binaries are not code-signed.** Verify the download against the minisign-signed checksums; see the README.

## Environment Variables & Flags

New in this release:

| Variable / Flag | Effect |
|---|---|
| `GOANIME_ENABLED_SOURCES` | Turn on opt-in sources (`goyabu`) |
| `GOANIME_DISABLED_SOURCES` | Kill-switch for sources |
| `GOANIME_POBREFLIX_URL` | Repoint the Pobreflix catalog without a release |
| `GOANIME_ANIMEDRIVE_URL` | Point AnimeDrive at a reachable mirror |
| `GOANIME_ANIMESDIGITAL_URL` | Override the AnimesDigital host |
| `GOANIME_TOPCINE_URL`, `GOANIME_STARTFLIX_URL`, `GOANIME_STARTFLIX_PANEL_URL` | Override the TopCine and StartFlix hosts |
| `GOANIME_VIDSRC_URL`, `GOANIME_2EMBED_URL` | Override the vidsrc and 2Embed fallbacks |
| `GOANIME_ANIMEFIRE_API` | Override the AnimeFire API origin |
| `GOANIME_HIANIME_AUDIO` | HiAnime release audio (replaces `GOANIME_ANIDB_LANG`) |
| `GOANIME_TERMINAL=conhost` | Stay in the classic Windows console instead of reopening in Windows Terminal |
| `GOANIME_JAPANESE=1\|0` | Force Japanese glosses on or off |
| `GOANIME_PLAIN=1` | Plain-text TUI (same as `--plain`) |
| `--plain` | Plain-text TUI |
| `--source movietv`, `--source ptbr` | Search every movie/TV source, or every Portuguese source |

Removed: `--sf-window`, `--sf-offscreen`, `GOANIME_SF_HOST`, `GOANIME_SF_OFFSCREEN`, `GOANIME_ANIDB_LANG`.

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
