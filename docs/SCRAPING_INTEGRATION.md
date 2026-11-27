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
a `referer`, an `origin` or a `subtitles` list, `providers.streamFromMetadata`
puts them in the resolved stream, and `Stream.Publish` hands them to mpv and
the downloader.

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
  as `metadata["referer"]`, and `providers.streamFromMetadata` is what puts
  it where mpv and the downloader read it.
- `GOANIME_HIANIME_AUDIO=dub` switches the preferred audio track; subtitled is
  the default.

`--source anidb` and `anidb.app` URLs still route here, so an existing alias or
a title restored from history keeps working.

Regression coverage:

```bash
go test ./internal/scraper/providers/hianime/ -count=1
```

### Miruro

miruro.cx is the English source upstream ani-cli moved to on 2026-11-25
(pystardust/ani-cli#1993). HiAnime still worked then and stays beside it.

The site is a SvelteKit app. Every page has a data endpoint, the page path
followed by `/__data.json`, that answers with the page's load data in
SvelteKit's devalue format: each value stored once in a flat array and
referred to by index. Three requests, with no key and no token:

| Stage | Request | Answers with |
| --- | --- | --- |
| search | `GET /search/__data.json?q=<q>` | `items`, AniList metadata including the AniList and MAL ids |
| episodes | `GET /watch/<id>/<slug>/__data.json?ep=1` | the whole episode list |
| stream | `GET /watch/<id>/<slug>/__data.json?ep=<n>` | that episode's sources |

Notes that matter when this breaks:

- The watch page **streams**. Its first line holds the page, and the episode
  list and the sources are `["Promise", i]` values settled by later
  `{"type":"chunk","id":N,...}` lines. The decoder in `devalue.go` resolves
  them by id, not by line order.
- Any slug is accepted in a watch URL: the app answers with a redirect to the
  canonical one, which the client follows (paths on the site only). Episode
  URLs are built with the canonical slug, so playing one is a single request.
- Sources come per track: `ssub` (subtitles as separate files), `sub` (burned
  in) and `dub`, each listing providers in the site's `providerOrder`, each
  provider several servers. A server with only an embed page is skipped.
- The CDNs **check headers**, and not all the same ones. Measured 2026-11-26:
  the `sun` family answers 403 without both Referer and Origin, `bee` and
  `hop` without the Referer, `knob` checks nothing. Each server names its own
  in `headers`; the scraper returns them as `metadata["referer"]` and
  `metadata["origin"]`.
- Mirrors fail on their own (the `sun` family refused every request for a
  while on 2026-11-26 while `bee` played). The client fetches each candidate's
  playlist with its headers and plays the first that answers, trying at most
  eight.
- `sun`'s subtitle files are ASS served as `.jpeg` with `"format":"vtt"`; mpv
  reads them by content. Only English and Portuguese files are passed on,
  because a provider can list twenty languages and its `default` flag once
  named Arabic.
- `GOANIME_MIRURO_AUDIO=dub` switches the preferred track; subtitled is the
  default.

Regression coverage:

```bash
go test ./internal/scraper/providers/miruro/ -count=1
GOANIME_LIVE=1 go test -run TestLiveMiruroChain -v ./internal/scraper/providers/miruro/
```
