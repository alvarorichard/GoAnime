# Scraping notes

How GoAnime talks to the sites behind its sources, and what is known about
each of them. For usage, see the [README](../README.md). To add or remove a
source, follow [ADDING_A_SOURCE.md](ADDING_A_SOURCE.md).

## How a search runs

Each source registers itself in the source registry (`internal/api/source`).
A search fans out to every active source in parallel
(`providers.SearchAll` in `internal/api/providers/dispatch.go`), each with its
own timeout, and merges the results. A source whose circuit breaker is open,
after three consecutive origin or blocking failures, is skipped for ten
minutes. `--source` narrows the fan-out to one source, or to the Portuguese
sources with `ptbr`.

Every source's client lives in `internal/scraper/providers/<name>` and is
reached through this interface (`internal/scraper/manager.go`):

```go
type UnifiedScraper interface {
    SearchAnime(query string, options ...any) ([]*models.Anime, error)
    GetAnimeEpisodes(animeURL string) ([]models.Episode, error)
    GetStreamURL(episodeURL string, options ...any) (string, map[string]string, error)
    GetType() ScraperType
}
```

`GetStreamURL` returns the stream URL plus metadata. When the metadata carries
a `referer` or a `subtitles` list, `providers.applyPlaybackMetadata` hands them
to mpv and the downloader.

Failures are classified as described in
[source-diagnostics-test-plan.md](source-diagnostics-test-plan.md).

## Source notes

### HiAnime

This slot has held three hosts. AllAnime went first, when mkissa.to removed the
material its per-request key was derived from. anidb.app took over and then went
to a site-wide 503 "Under Maintenance" (homepage, `/browse` and its API alike),
where it still was on 2026-09-22. hianime.at is the current one.

The chain is four plain requests, with no challenge and no key derivation:

| Stage | Request | Answers with |
| --- | --- | --- |
| search | `GET /search?keyword=<q>` | HTML, one `.flw-item` per result |
| episodes | `GET /api/theme/episode/list/<animeID>` | `{"status","totalItems","html"}` |
| servers | `GET /api/theme/episode/servers?episodeId=<id>` | `{"status","html"}` |
| stream | `GET <embed URL>` | the player config |

Notes that matter when this breaks:

- The two API endpoints return **HTML inside a JSON envelope**, which is why the
  client both decodes and parses in one call. A wrong path answers 404 with a
  Laravel `{"message":"The route … could not be found."}`, not an HTML page.
- A server's `data-hash` is **plain base64** of its embed URL, despite the name.
- Only the **ZokoAnime** server is read. Its page carries the whole player
  config in `window.__P`: base64, XOR'd with the build tag `otaku-embed-v1`.
  The MegaPlay servers instead return an AES-encrypted `enc` field whose key
  lives in a third-party repository, the same dependency shape that made
  AllAnime unmaintainable.
- The stream CDN **checks the Referer**, but not on every URL. Sampled
  2026-09-22 across three titles with freshly resolved links, the master
  playlist answered 200 with or without one, while the variant playlist below it
  answered 403 without one on two of the three. A player that omits it therefore
  appears to work until it hits a title where it matters. The scraper returns it
  as `metadata["referer"]`, and `providers.applyPlaybackMetadata` is what puts
  it where mpv and the downloader read it.
- `GOANIME_HIANIME_AUDIO=dub` switches the preferred audio track; subtitled is
  the default.

`--source anidb` and `anidb.app` URLs still route here, so an existing alias or
a title restored from history keeps working.

Regression coverage:

```bash
go test ./internal/scraper/providers/hianime/ -count=1
```
