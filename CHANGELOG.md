# GoAnime Release Notes - Version 1.8.1

Release date: 2026-04-20

## Highlights

- **New PT-BR Source (SuperFlix)**: Added a robust new source for movies, anime, series, and doramas, entirely in Brazilian Portuguese (PT-BR).
- **Jellyfin & Plex Compatibility**: Reorganized the media storage structure for movies, series, anime, and doramas, making them out-of-the-box compatible with Jellyfin and Plex.
- **Automatic Season Inference**: GoAnime now automatically detects the season number from the anime title (e.g., "Mushoku Tensei Season 2"), eliminating manual configuration and ensuring episodes are organized under the correct season.
- **More Robust Downloads with Smart Fallback**: The download system has been completely revamped with intelligent source selection, automatic fallback to alternative URLs (especially for AnimeFire), and improved progress tracking for batch downloads.
- **AllAnime Encrypted API Support**: Implemented AES-256-CTR decryption for AllAnime's new encrypted "tobeparsed" API responses, ensuring continued access to the AllAnime source after their API migration.
- **More Stable Terminal Interface**: Critical fixes to terminal interaction prevent display issues during asynchronous operations such as downloads and updates.

## Features

- Add new PT-BR source (SuperFlix) covering movies, anime, series, and doramas.
- Implement automated media organization optimized for Jellyfin and Plex playback with compatible folder naming and external IDs.
- Implement download-all mode and interactive selection menu for anime and TV show downloads.
- Implement automatic season number inference from anime titles using patterns like "Season N" and "Nth Season", with support for English and Romaji titles via AniList.
- Add intelligent AnimeFire download source resolution with preferred quality selection (best, worst, or specific resolution) and automatic candidate ordering.
- Implement fallback mechanism for AnimeFire downloads: when a URL returns 404, the system automatically fetches an alternative source from the video API.
- Add hierarchical progress model (`childProgress`) for batch downloads, enabling accurate tracking of total and received bytes across multiple simultaneous episodes.
- Introduce console output suppression mechanism (`SuppressConsole`) during asynchronous operations to prevent interference with progress bars.
- Implement episode-specific tracking keys to prevent data overlap in the database.
- Implement AES-256-CTR decryption for AllAnime's new encrypted "tobeparsed" source URL blobs, with JSON parsing and regex fallback extraction.
- Integrate images and Discord RPC for the SuperFlix source.

## Bug Fixes

- Fix case-sensitive comparison of source names (AllAnime, AnimeFire, AnimeDrive, Goyabu) in stream selection, preventing failures when the `Source` field had different capitalization.
- Fix episode resolution for anime with selected season greater than 1, ensuring local episodes are correctly mapped within the season instead of using absolute numbering.
- Fix terminal capability queries (CSI response suppression) that caused corrupted output during spinners and progress bars.
- Fix `USERPROFILE` environment variable in download tests for Windows compatibility.
- Ensure all sources appear in scraper search results.
- Fix Windows path compatibility using `filepath.ToSlash` in media naming functions.
- Fix AllAnime source URL decoding by replacing the incomplete hex substitution table with the full cipher from ani-cli's `provider_init`, covering all uppercase/lowercase letters, digits, and special characters.
- Fix AllAnime `extractSourceURLs` to fall through to regex extraction when JSON parsing returns zero URLs, preventing silent failures.
- Fix HTTP2 transport race conditions by replacing the shared fast client with dedicated instances.

## Improvements

- Refactor `RunClean` in the TUI package to encapsulate interactive components (spinners, prompts) and prevent interference with standard output.
- Refactor the AllAnime scraper for improved code readability and maintainability.
- Significantly expand test coverage: +1,900 lines of AllAnime scraper tests, new download regression tests, season propagation tests, and progress aggregation tests.
- Optimize download performance with larger buffers and improved concurrency settings.
- Enhance media type handling and sorting for PT-BR entries, including type disambiguation for FlixHQ/SFlix results.
- Update dependencies: bubbletea v2.0.6, lipgloss v2.0.3, and other packages in `go.mod`.
