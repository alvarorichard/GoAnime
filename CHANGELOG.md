# GoAnime Release Notes - Version 1.7

Release date: 2026-04-08

## Highlights

- **AllAnime Cloudflare Bypass**: All AllAnime API requests now use POST with JSON body, bypassing Cloudflare challenge pages that were blocking GET requests.
- **9Anime Integration**: Full support for 9Anime as a new anime streaming source, including search, streaming, downloading, and subtitle embedding.
- **PT-BR Sources (Goyabu + AnimeFire)**: New Goyabu scraper and improved AnimeFire integration for Brazilian Portuguese anime content with `-source ptbr` flag.
- **Blogger Proxy with Chrome TLS Impersonation**: Refactored Blogger video extraction to use a local Go HTTP proxy with Chrome TLS fingerprinting, solving 403 errors from Google CDN.
- **Anime4K Upscaling**: New real-time video upscaling using Anime4K shaders with automatic shader installation, plus experimental GAN UUL upscaling mode for enhanced visual quality.
- **Security Hardening**: SSRF protection for all HTTP clients, response body size limits, data race fixes, and Go 1.26.2 with patched `crypto/x509` and `crypto/tls` vulnerabilities.

## Features

- Add 9Anime as a new streaming source with search, episode listing, server selection, subtitle selection, and downloading support.
- Add Goyabu scraper as a new PT-BR anime source with search, episode parsing, and stream URL extraction.
- Add `-source ptbr` flag with `[PT-BR]` tags and automatic legendado/dublado detection.
- Implement AnimeFire `GetAnimeEpisodes` and `GetEpisodeStreamURL` for full AnimeFire playback and download support.
- AllAnime API requests switched from GET to POST with JSON body to bypass Cloudflare challenges; referer updated to `allmanga.to`.
- Implement Blogger CDN direct downloading with multi-threading and improved error handling.
- Implement Go HTTP proxy for Blogger video extraction with Chrome TLS impersonation (fixes 403 from Google CDN).
- Add `extractRomajiFromURL` for PT-BR AniList title resolution, enriching AniList lookup with romaji from URL slugs.
- Add media path handling for decryption API in FlixHQ and SFlix clients.
- Add fallback API support for FlixHQ and SFlix clients.
- Enhance subtitle management with global source tracking, user prompts, and 9Anime subtitle embedding.
- Add support for overriding media title in MPV with cleaned anime names.
- Add Wayland GPU context support in mpv playback.
- Add `runWithSpinner` function for improved loading feedback during network calls.
- Add `govulncheck` for vulnerability scanning in CI workflow.
- Enhance `SanitizeForFilename` to handle 9Anime-specific metadata and multilanguage tags.
- Enhance `CleanTitle` to handle 9Anime titles, `[Multilanguage]` prefixes, and episode info patterns.
- Add Anime4K real-time upscaling with automatic shader installation and MPV integration.
- Add experimental GAN UUL upscaling mode for enhanced visual quality.
- Add download location printing for episodes and movies.
- Enhance season handling propagation in player and Discord Rich Presence.

## Improvements

- Optimize HTTP client usage with connection pre-warming and shared singleton clients for better performance.
- Optimize mpv path lookup and enhance async operations for improved playback performance.
- Optimize timeouts and improve parallel search handling in anime fetching with concurrent scraper execution.
- Refactor playback handling and simplify user input menus; remove `isMovieOrTV` parameter from `showPlayerMenu`.
- Refactor scrapers for improved performance and maintainability; reduce ScraperManager instances.
- Refactor media state management to prevent data races and improve concurrency safety.
- Streamline debug logging by replacing conditional checks with unified debug functions.
- Improve log file naming with unique session identification and file output support.
- Enhance yt-dlp integration with unsafe extension handling and direct HTTP fallback for downloads.
- Improve error handling and logging for episode downloads.
- Enhance URL handling with accent normalization.
- Update Go version to 1.26.2 (fixes CVE in `crypto/x509` and `crypto/tls`).
- Update dependencies: `go-ytdlp` v1.3.2, `go-crypto` v1.4.0, `golang.org/x/net` v0.51.0, `charm.land/bubbles`, and others.
- CI: enhance Windows dependency installation with dynamic mpv download, yt-dlp integration, GCC and Inno Setup checks with retries.
- CI: add GitHub token authentication for MPV release downloads.
- CI: update golangci-lint action version, increase timeout, and add configuration file.

## Bug Fixes

- Fix AllAnime API blocked by Cloudflare by switching from GET to POST requests.
- Fix SSRF vulnerability: implement SSRF protection for all HTTP clients and refactor transport settings.
- Fix excessive memory usage by limiting response body size on all HTTP responses.
- Fix data races in media state management with proper concurrency controls.
- Fix Blogger video extraction preferring 720p (itag=22) over 360p (itag=18).
- Fix playback: replace `Fatal` calls with proper error returns in playback package (3 bugs fixed).
- Fix AnimeFire video URL resolution before streaming and download.
- Fix Goyabu search reliability, episode parsing, and test coverage.
- Fix duplicate quality menu and clean PT-BR title formatting.
- Fix `nil` check for `streamInfo` in `GetFlixHQStreamURL` to prevent panics.
- Fix result channel drain before early exit in concurrent search to prevent goroutine leaks.
- Fix FlixHQ tests to use httptest mocks instead of real API calls.
- Fix transient error handling and skip tests for unavailable external services.
- Fix Codacy issues in `goyabu.go` (unexported constants, unused variables).
- Fix Go 1.26.2 standard library vulnerabilities (GO-2026-4947, GO-2026-4946, GO-2026-4870, GO-2026-4866).

## Developer Notes

- Replace FlixHQ real API calls in tests with `httptest` mocks for reliability.
- Add tests for Blogger video extraction, playback error returns, Goyabu scraper, and `CleanTitle` variations.
- Add doc comments to exported functions in `appflow` package.
- Enhance contribution guidelines with mandatory security and testing requirements.
- CI workflows updated to Go 1.26.2 across `ci.yml`, `coverage.yml`, and `release.yml`.