# GoAnime Public API

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](../../LICENSE)

A clean and simple Go library for searching and scraping anime content from multiple sources.

## Features

-  **Multi-source Search**: Search across multiple anime sources (AllAnime, AnimeFire, etc.)
-  **Episode Management**: Get detailed episode lists with metadata
-  **Stream URLs**: Extract streaming URLs and required headers
-  **Rich Metadata**: Access AniList IDs, MAL IDs, genres, scores, and more
-  **Simple API**: Clean and intuitive interface
-  **Type-safe**: Full Go type safety with documented structs
-  **Fast**: Concurrent searches across sources

## Installation

```bash
go get github.com/alvarorichard/Goanime
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"

    "github.com/alvarorichard/Goanime/pkg/goanime"
)

func main() {
    // Create a new client
    client := goanime.NewClient()

    // Search for anime
    results, err := client.SearchAnime("One Piece", nil)
    if err != nil {
        log.Fatal(err)
    }

    // Display results
    for _, anime := range results {
        fmt.Printf("%s [%s]\n", anime.Name, anime.Source)
    }
}
```

## Usage Examples

### 1. Search Anime

```go
client := goanime.NewClient()

// Search all sources
results, _ := client.SearchAnime("Naruto", nil)

// Search specific source
source := types.SourceAllAnime
results, _ := client.SearchAnime("Naruto", &source)
```

### 2. Get Episodes

```go
client := goanime.NewClient()

// Search and get anime
results, _ := client.SearchAnime("Attack on Titan", nil)
anime := results[0]

// Parse source
source, _ := types.ParseSource(anime.Source)

// Get episodes
episodes, _ := client.GetAnimeEpisodes(anime.URL, source)

for _, ep := range episodes {
    fmt.Printf("Episode %s\n", ep.Number)
}
```

### 3. Get Stream URL

```go
client := goanime.NewClient()

// ... get anime and episodes as above ...

// Get stream URL for an episode
streamURL, headers, err := client.GetStreamURL(episode.URL, source)
if err != nil {
    log.Fatal(err)
}

fmt.Printf("Stream: %s\n", streamURL)
fmt.Printf("Headers: %v\n", headers)
```

### 4. Available Sources

```go
client := goanime.NewClient()

sources := client.GetAvailableSources()
for _, source := range sources {
    fmt.Println(source.String())
}
// Output:
// AllAnime
// AnimeFire
```

## API Reference

### Types

#### `Client`
Main client for interacting with anime sources.

**Methods:**
- `NewClient() *Client` - Create a new client
- `SearchAnime(query string, source *types.Source) ([]*types.Anime, error)` - Search for anime
- `GetAnimeEpisodes(animeURL string, source types.Source) ([]*types.Episode, error)` - Get episodes
- `GetStreamURL(episodeURL string, source types.Source, options ...interface{}) (string, map[string]string, error)` - Get stream URL
- `GetAvailableSources() []types.Source` - List available sources

#### `types.Anime`

```go
type Anime struct {
    Name      string              // Anime title
    URL       string              // Source-specific URL
    ImageURL  string              // Cover image URL
    Episodes  []*Episode          // List of episodes
    AnilistID int                 // AniList ID
    MalID     int                 // MyAnimeList ID
    Source    string              // Source name
    Details   *AniListDetails     // Extended metadata
}
```

#### `types.Episode`

```go
type Episode struct {
    Number    string           // Episode number (e.g., "1", "1.5")
    Num       int              // Episode number as int
    URL       string           // Episode URL
    Title     *TitleDetails    // Episode title
    Aired     string           // Air date
    Duration  int              // Duration in seconds
    IsFiller  bool             // Is filler episode
    IsRecap   bool             // Is recap episode
    Synopsis  string           // Episode description
    SkipTimes *SkipTimes       // Intro/outro timestamps
}
```

#### `types.Source`

Available sources:
- `types.SourceAllAnime` - AllAnime source
- `types.SourceAnimeFire` - AnimeFire source

**Methods:**
- `String() string` - Get source name
- `ToScraperType() scraper.ScraperType` - Convert to internal type
- `ParseSource(s string) (Source, error)` - Parse string to Source

## Examples

Complete working examples are available in the `examples/` directory:

- [`search/`](goanime/examples/search/main.go) - Basic anime search
- [`episodes/`](goanime/examples/episodes/main.go) - Get episode list
- [`stream/`](goanime/examples/stream/main.go) - Get streaming URL
- [`source_specific/`](goanime/examples/source_specific/main.go) - Search specific sources

To run an example:

```bash
go run ./pkg/goanime/examples/search/main.go
```

## Advanced Usage

### Error Handling

```go
results, err := client.SearchAnime("One Piece", nil)
if err != nil {
    // Handle specific errors
    if strings.Contains(err.Error(), "no anime found") {
        fmt.Println("No results found")
    } else {
        log.Printf("Search error: %v", err)
    }
    return
}
```

### Working with Metadata

```go
for _, anime := range results {
    if anime.Details != nil {
        fmt.Printf("Score: %d/100\n", anime.Details.AverageScore)
        fmt.Printf("Genres: %v\n", anime.Details.Genres)
        fmt.Printf("Status: %s\n", anime.Details.Status)
        fmt.Printf("Episodes: %d\n", anime.Details.Episodes)
    }
}
```

### Episode Filtering

```go
episodes, _ := client.GetAnimeEpisodes(anime.URL, source)

// Filter out filler episodes
mainEpisodes := make([]*types.Episode, 0)
for _, ep := range episodes {
    if !ep.IsFiller {
        mainEpisodes = append(mainEpisodes, ep)
    }
}
```

## Integration Examples

### Use with MPV Player

```go
streamURL, headers, _ := client.GetStreamURL(episode.URL, source)

// Build MPV command with headers
args := []string{streamURL}
for key, value := range headers {
    args = append(args, fmt.Sprintf("--http-header-fields=%s: %s", key, value))
}

cmd := exec.Command("mpv", args...)
cmd.Run()
```

### Use with Custom HTTP Client

```go
streamURL, headers, _ := client.GetStreamURL(episode.URL, source)

req, _ := http.NewRequest("GET", streamURL, nil)
for key, value := range headers {
    req.Header.Set(key, value)
}

resp, _ := http.DefaultClient.Do(req)
defer resp.Body.Close()

// Process video stream...
```

## Documentation

For detailed documentation, see:
- [Complete API Documentation](goanime/README.md)
- [Code Examples](goanime/examples/)
- [GoDoc](https://pkg.go.dev/github.com/alvarorichard/Goanime/pkg/goanime)

## Contributing

This is a public API extracted from the internal implementation. When contributing:

1. Keep the API simple and clean
2. Maintain backward compatibility
3. Add tests for new features
4. Update documentation

## Notes

- Stream URLs may expire after some time
- Some sources may require specific headers for streaming
- Not all metadata is available for all sources
- The library handles rate limiting automatically

## License

MIT License - see [LICENSE](../../LICENSE) for details

## Related Projects

- [GoAnime CLI](../../) - Full-featured CLI application using this library
- Main project documentation in the [root README](../../README.md)
