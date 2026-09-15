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
