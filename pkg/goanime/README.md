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
available for the anime sources and for StartFlix (`Source.SupportsEpisodes()`).
StartFlix is a movie and series catalogue: its series go through `GetSeasons`
and `GetSeasonEpisodes` (see below).

## Client

### `func NewClient() *Client`

Creates a client backed by the same source registry as the command-line
tool. It performs no network requests.

### `func (c *Client) SearchAnime(query string, source *types.Source) ([]*types.Anime, error)`

Searches for `query`. With a `nil` source it searches every enabled source in
parallel; with a source such as `&types.SourceAnimeFire` it searches that one only.

### `func (c *Client) GetAnimeEpisodes(animeURL string, source types.Source) ([]*types.Episode, error)`

Returns the episodes of an AnimeFire title. `animeURL` is the `URL` of a
search result.

### `func (c *Client) GetEpisodeStreamURL(anime *types.Anime, episode *types.Episode, options *StreamOptions) (string, map[string]string, error)`

Returns the stream URL of an episode and a metadata map whose `source` key
names the source. `anime.Source` must be the name of one of the sources below; anything
else returns an `unknown source` error, and a source that is search-only
returns `ErrEpisodesUnsupported`. `options` may be `nil`.

### `func (c *Client) GetStreamURL(episodeURL string, source types.Source, options ...any) (string, map[string]string, error)`

Deprecated. Returns the same values as `GetEpisodeStreamURL`, taking the
episode URL directly.

### `func (c *Client) GetSeasons(anime *types.Anime) ([]*types.Season, error)`

For a catalogue source (`Source.IsCatalog()`: StartFlix). Lists a series'
seasons, each with the audios it is offered in (`Dublado`, `Legendado`). A
movie has no seasons: the list is empty, and `GetAnimeEpisodes` returns the
movie as its one episode. `anime` is a `SearchAnime` result of that source;
an anime source returns `ErrNotCatalog`.

### `func (c *Client) GetSeasonEpisodes(anime *types.Anime, season int, audio string) ([]*types.Episode, error)`

Lists one season in one audio. An empty `audio` takes the dub when the season
has one. A season the series lacks is `ErrSeasonNotFound` and an audio the
season lacks is `ErrAudioUnavailable`; both messages say what exists.
`GetAnimeEpisodes` on a series returns `ErrSeasonRequired`.

### `func (c *Client) GetAvailableSources() []types.Source`

Returns the sources that support episode lists and stream URLs: AnimeFire,
Goyabu, HiAnime, StartFlix, AnimesDigital, AnimeDrive and AnimesROLL. `types.Sources()`
lists every source that can be searched. A source that is switched off
(Goyabu is, by default; see `GOANIME_ENABLED_SOURCES`) is still listed but
returns no search results.

### `type StreamOptions`

```go
type StreamOptions struct {
    Quality string // "best", "worst", "1080p", "720p", "480p", "360p"
    Mode    string // "sub" or "dub"
}
```

`DefaultStreamOptions()` returns `{Quality: "best", Mode: "sub"}`. `Audio`
(`"Dublado"` or `"Legendado"`) is applied for catalogue sources: an audio
that does not play is `ErrAudioUnavailable`, never the other audio in its
place, and `metadata["audio"]` says what played. An empty `Audio` plays the
episode in the audio it was listed in (`Episode.Audio`). When the panel's servers
fail, or only play the other audio, a Legendado request is retried on the
independent vidsrc fallback by the title's ids; a Dublado request never is,
because the fallback serves the original audio only. That needs the episode as
`GetSeasonEpisodes` returned it (`Season` and `Number` are what it looks up). `Quality` and `Mode` are
accepted but not applied yet: the stream comes back in the source's default
quality.

## Types

### `types.Source`

| Value                         | `String()`        |
| ----------------------------- | ----------------- |
| `types.SourceAnimeFire`       | `"Animefire.io"`  |
| `types.SourceGoyabu`          | `"Goyabu"`        |
| `types.SourceHiAnime`         | `"HiAnime"`       |
| `types.SourceStartFlix`       | `"StartFlix"`     |
| `types.SourceTopCine`         | `"TopCine"`       |
| `types.SourceAnimesDigital`   | `"AnimesDigital"` |
| `types.SourceAnimeDrive`      | `"AnimeDrive"`    |
| `types.SourcePobreflix`       | `"Pobreflix"`     |
| `types.SourceAnroll`          | `"AnimesROLL"`    |
| `types.SourcePomfy`           | `"Pomfy"`         |

`types.Sources()` returns all of them. All can be searched. StartFlix, TopCine,
Pobreflix and Pomfy are movie and series catalogues; only StartFlix lists
episodes through this package so far (`Source.IsCatalog()`). For the other
three, `GetAnimeEpisodes`, `GetStreamURL` and `GetEpisodeStreamURL` return
`goanime.ErrEpisodesUnsupported`; `Source.SupportsEpisodes()` tells which.
Some catalogues index PT-BR titles only, so search `A Origem`, not `Inception`.

`types.ParseSource(s string) (Source, error)`
accepts each `String()` spelling, the lowercase name of each source (and
`fire` for AnimeFire), and returns an error for anything else. New values are
only appended, so a stored `Source` keeps its meaning.

### `types.Season`

```go
type Season struct {
    Number int
    Audios []string // "Dublado" and/or "Legendado", as the site offers them
}
```

Audios are written `Dublado` and `Legendado` (the library also accepts `dub`
and `sub`, in any case). Check `Season.Audios` before asking: many titles have
only the dub, and an audio the season lacks is `ErrAudioUnavailable`.

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

- A Legendado stream can fail on some titles even when it is listed: the
  servers that carry it come and go. The error says what played instead and
  how to ask for it; try again later or ask for `Dublado`.
- Calls take no `context`. A failing stream call can take about 5 seconds, and
  the worst case is bounded at 100 seconds.
- Streams from the anime sources and the Dublado streams of StartFlix are URLs
  on a local proxy (`http://127.0.0.1:PORT/...`) that lives while your program
  runs: play them before it exits. Legendado streams are HLS (`.m3u8`) on a
  host that changes per call, and need `metadata["referer"]` as the Referer.
- Titles are listed under their Portuguese names, so `breaking bad` finds
  "Breaking Bad: A Química do Mal".

- Stream URLs expire. Request a new one shortly before playing.
- AnimeFire's video host expects a `Referer: https://animefire.io` header.
  Pass it on to your player, for example
  `mpv --referrer=https://animefire.io <url>`.
- The sources are third-party websites. When one changes, calls that reach it
  fail until GoAnime is updated.
