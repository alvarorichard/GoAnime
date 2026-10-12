# Using GoAnime as a library

The `pkg/goanime` package exposes GoAnime's search and scraping code to other
Go programs. It uses the same source registry as the command-line tool, so it
searches the same sources and honours the same environment variables, such as
`GOANIME_ENABLED_SOURCES` and `GOANIME_DISABLED_SOURCES`.

This guide walks through a search, an episode list and a stream URL. The
methods and types are listed in [pkg/goanime/README.md](../pkg/goanime/README.md),
and [pkg/goanime/examples](../pkg/goanime/examples) has complete programs.

## Current limits

- Search covers every enabled source.
- Episode lists and stream URLs are available for AnimeFire results only.
  `types.Source` has a single value, `types.SourceAnimeFire`, and passing a
  result from another source to `GetEpisodeStreamURL` returns an
  `unknown source` error.
- `StreamOptions` is accepted but not applied yet: the stream comes back in
  the source's default quality.

## Installation

The module requires Go 1.27 or later.

```bash
go get github.com/alvarorichard/Goanime
```

## Searching

```go
package main

import (
    "fmt"
    "log"

    "github.com/alvarorichard/Goanime/pkg/goanime"
)

func main() {
    client := goanime.NewClient()

    results, err := client.SearchAnime("One Piece", nil)
    if err != nil {
        log.Fatal(err)
    }

    for _, anime := range results {
        fmt.Printf("%s [%s]\n", anime.Name, anime.Source)
    }
}
```

Passing `nil` searches every enabled source. Result names carry a language tag
such as `[English]` or `[PT-BR]`, and `Source` holds the name of the source the
result came from.

To search AnimeFire only, pass its source:

```go
import "github.com/alvarorichard/Goanime/pkg/goanime/types"

source := types.SourceAnimeFire
results, err := client.SearchAnime("One Piece", &source)
```

## Listing episodes

```go
anime := results[0]

episodes, err := client.GetAnimeEpisodes(anime.URL, types.SourceAnimeFire)
if err != nil {
    log.Fatal(err)
}

for _, ep := range episodes {
    fmt.Println("Episode", ep.Number)
}
```

`Title`, `SkipTimes` and the other metadata fields are filled in only when the
source provides them, so check pointer fields for `nil` before using them.

## Getting a stream URL

```go
streamURL, metadata, err := client.GetEpisodeStreamURL(anime, episodes[0], nil)
if err != nil {
    log.Fatal(err)
}

fmt.Println(streamURL)
fmt.Println(metadata["source"]) // "animefire"
```

Stream URLs expire, so request a new one shortly before playing. AnimeFire's
video host refuses a request without a `User-Agent`; mpv sends one, so the URL
plays as is:

```bash
mpv "<stream URL>"
```

An episode whose video the host no longer has returns an error ending in
`the source has no file for this episode` instead of a URL that cannot play.

## Running the examples

```bash
go run ./pkg/goanime/examples/search
go run ./pkg/goanime/examples/source_specific
go run ./pkg/goanime/examples/episodes
go run ./pkg/goanime/examples/stream
```

The examples contact the real sites, so they fail when a source is down.

## Tests

```bash
go test -short ./pkg/goanime/...
```

## Contributing to the library

Keep the API small and backward compatible: other projects depend on it.
Adding sources to `types.Source` changes the public API, so discuss it in an
issue first. New behavior needs tests, and this guide and
[pkg/goanime/README.md](../pkg/goanime/README.md) need to be updated with it.
