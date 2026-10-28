# pkg/goanime

Package `goanime` is the public Go API for GoAnime's search and scraping code.
For a walkthrough, see [docs/LIBRARY_USAGE.md](../../docs/LIBRARY_USAGE.md).
Complete programs are in [examples](examples).

```bash
go get github.com/alvarorichard/Goanime/v2
```

```go
import (
    "github.com/alvarorichard/Goanime/v2/pkg/goanime"
    "github.com/alvarorichard/Goanime/v2/pkg/goanime/types"
)
```

Search covers every enabled source. Episode lists and stream URLs are
available for AnimeFire results only.

## Client

### `func NewClient() *Client`

Creates a client backed by the same source registry as the command-line
tool. It performs no network requests.

### `func (c *Client) SearchAnime(query string, source *types.Source) ([]*types.Anime, error)`

Searches for `query`. With a `nil` source it searches every enabled source in
parallel; with `&types.SourceAnimeFire` it searches AnimeFire only.

### `func (c *Client) GetAnimeEpisodes(animeURL string, source types.Source) ([]*types.Episode, error)`

Returns the episodes of an AnimeFire title. `animeURL` is the `URL` of a
search result.

### `func (c *Client) GetEpisodeStreamURL(anime *types.Anime, episode *types.Episode, options *StreamOptions) (string, map[string]string, error)`

Returns the stream URL of an episode and a metadata map whose `source` key
names the source. `anime` must be an AnimeFire result; any other source
returns an `unknown source` error. `options` may be `nil`.

### `func (c *Client) GetStreamURL(episodeURL string, source types.Source, options ...any) (string, map[string]string, error)`

Deprecated. Returns the same values as `GetEpisodeStreamURL`, taking the
episode URL directly.

### `func (c *Client) GetAvailableSources() []types.Source`

Returns the sources that support episode lists and stream URLs. Today that is
`types.SourceAnimeFire` only.

### `type StreamOptions`

```go
type StreamOptions struct {
    Quality string // "best", "worst", "1080p", "720p", "480p", "360p"
    Mode    string // "sub" or "dub"
}
```

`DefaultStreamOptions()` returns `{Quality: "best", Mode: "sub"}`. The options
are accepted but not applied yet: the stream comes back in the source's
default quality.

## Types

### `types.Source`

| Value                   | `String()`       |
| ----------------------- | ---------------- |
| `types.SourceAnimeFire` | `"Animefire.io"` |

`types.ParseSource(s string) (Source, error)` accepts `AnimeFire`,
`animefire`, `fire`, `Animefire.io` and `animefire.io`, and returns an error
for anything else.

### `types.Anime`

```go
type Anime struct {
    Name      string          // title, with a language tag such as [English] or [PT-BR]
    URL       string          // title page on the source
    ImageURL  string          // cover image
    Episodes  []*Episode
    AnilistID int             // 0 when unknown
    MalID     int             // 0 when unknown
    Source    string          // name of the source the result came from
    Details   *AniListDetails // always set; fields are zero when AniList has no data
}
```

### `types.Episode`

```go
type Episode struct {
    Number    string        // as the source writes it, e.g. "1" or "1.5"
    Num       int           // Number as an integer
    URL       string        // episode page, used to get the stream URL
    Title     *TitleDetails // nil when the source has no titles
    Aired     string
    Duration  int           // seconds
    IsFiller  bool
    IsRecap   bool
    Synopsis  string
    SkipTimes *SkipTimes    // opening and ending timestamps, nil when unknown
}
```

`TitleDetails`, `SkipTimes`, `SkipTime`, `AniListDetails`, `Title` and
`CoverImages` are plain structs; see [types/anime.go](types/anime.go).

## Notes

- Stream URLs expire. Request a new one shortly before playing.
- AnimeFire's video host expects a `Referer: https://animefire.io` header.
  Pass it on to your player, for example
  `mpv --referrer=https://animefire.io <url>`.
- The sources are third-party websites. When one changes, calls that reach it
  fail until GoAnime is updated.
