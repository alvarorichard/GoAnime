# GoAnime Library - Usage Examples

This package provides a simple and clean API for searching and scraping anime content from multiple sources.

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
    "github.com/alvarorichard/Goanime/pkg/goanime/types"
)

func main() {
    // Create a new client
    client := goanime.NewClient()

    // Search for anime across all sources
    results, err := client.SearchAnime("One Piece", nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d results\n", len(results))
    for i, anime := range results {
        fmt.Printf("%d. %s [%s]\n", i+1, anime.Name, anime.Source)
    }
}
```

## Examples

### 1. Search Anime in a Specific Source

```go
package main

import (
    "fmt"
    "log"

    "github.com/alvarorichard/Goanime/pkg/goanime"
    "github.com/alvarorichard/Goanime/pkg/goanime/types"
)

func main() {
    client := goanime.NewClient()

    // Search only in AllAnime
    source := types.SourceAllAnime
    results, err := client.SearchAnime("Naruto", &source)
    if err != nil {
        log.Fatal(err)
    }

    for _, anime := range results {
        fmt.Printf("Name: %s\n", anime.Name)
        fmt.Printf("URL: %s\n", anime.URL)
        fmt.Printf("Source: %s\n", anime.Source)
        fmt.Println("---")
    }
}
```

### 2. Get Episodes for an Anime

```go
package main

import (
    "fmt"
    "log"

    "github.com/alvarorichard/Goanime/pkg/goanime"
    "github.com/alvarorichard/Goanime/pkg/goanime/types"
)

func main() {
    client := goanime.NewClient()

    // First, search for anime
    results, err := client.SearchAnime("Demon Slayer", nil)
    if err != nil {
        log.Fatal(err)
    }

    if len(results) == 0 {
        log.Fatal("No anime found")
    }

    // Get episodes for the first result
    anime := results[0]
    
    // Parse the source to use the correct scraper
    source, err := types.ParseSource(anime.Source)
    if err != nil {
        log.Fatal(err)
    }

    episodes, err := client.GetAnimeEpisodes(anime.URL, source)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d episodes\n", len(episodes))
    for _, ep := range episodes {
        fmt.Printf("Episode %s: %s\n", ep.Number, ep.URL)
        if ep.Title != nil {
            fmt.Printf("  Title: %s\n", ep.Title.English)
        }
    }
}
```

### 3. Get Stream URL for an Episode

```go
package main

import (
    "fmt"
    "log"

    "github.com/alvarorichard/Goanime/pkg/goanime"
    "github.com/alvarorichard/Goanime/pkg/goanime/types"
)

func main() {
    client := goanime.NewClient()

    // Search for anime
    results, err := client.SearchAnime("Attack on Titan", nil)
    if err != nil {
        log.Fatal(err)
    }

    if len(results) == 0 {
        log.Fatal("No anime found")
    }

    anime := results[0]
    source, _ := types.ParseSource(anime.Source)

    // Get episodes
    episodes, err := client.GetAnimeEpisodes(anime.URL, source)
    if err != nil {
        log.Fatal(err)
    }

    if len(episodes) == 0 {
        log.Fatal("No episodes found")
    }

    // Get stream URL for first episode
    streamURL, headers, err := client.GetStreamURL(episodes[0].URL, source)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Stream URL: %s\n", streamURL)
    fmt.Println("Headers:")
    for key, value := range headers {
        fmt.Printf("  %s: %s\n", key, value)
    }
}
```

### 4. List Available Sources

```go
package main

import (
    "fmt"

    "github.com/alvarorichard/Goanime/pkg/goanime"
)

func main() {
    client := goanime.NewClient()

    sources := client.GetAvailableSources()
    fmt.Println("Available sources:")
    for _, source := range sources {
        fmt.Printf("- %s\n", source.String())
    }
}
```

### 5. Complete Example: Search, List Episodes, and Get Stream

```go
package main

import (
    "fmt"
    "log"

    "github.com/alvarorichard/Goanime/pkg/goanime"
    "github.com/alvarorichard/Goanime/pkg/goanime/types"
)

func main() {
    client := goanime.NewClient()

    // 1. Search for anime
    fmt.Println("Searching for anime...")
    results, err := client.SearchAnime("Jujutsu Kaisen", nil)
    if err != nil {
        log.Fatal(err)
    }

    if len(results) == 0 {
        log.Fatal("No anime found")
    }

    // Display search results
    fmt.Printf("\nFound %d results:\n", len(results))
    for i, anime := range results {
        fmt.Printf("%d. %s [%s]\n", i+1, anime.Name, anime.Source)
    }

    // Select first result
    selectedAnime := results[0]
    fmt.Printf("\nSelected: %s\n", selectedAnime.Name)

    // 2. Get episodes
    source, err := types.ParseSource(selectedAnime.Source)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("\nFetching episodes...")
    episodes, err := client.GetAnimeEpisodes(selectedAnime.URL, source)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d episodes\n", len(episodes))

    // Display first 5 episodes
    displayCount := 5
    if len(episodes) < displayCount {
        displayCount = len(episodes)
    }

    fmt.Printf("\nFirst %d episodes:\n", displayCount)
    for i := 0; i < displayCount; i++ {
        ep := episodes[i]
        title := "N/A"
        if ep.Title != nil && ep.Title.English != "" {
            title = ep.Title.English
        } else if ep.Title != nil && ep.Title.Romaji != "" {
            title = ep.Title.Romaji
        }
        fmt.Printf("  Episode %s: %s\n", ep.Number, title)
    }

    // 3. Get stream URL for first episode
    if len(episodes) > 0 {
        fmt.Println("\nGetting stream URL for episode 1...")
        streamURL, headers, err := client.GetStreamURL(episodes[0].URL, source)
        if err != nil {
            log.Printf("Error getting stream URL: %v\n", err)
        } else {
            fmt.Printf("Stream URL: %s\n", streamURL)
            if len(headers) > 0 {
                fmt.Println("Required headers:")
                for key, value := range headers {
                    fmt.Printf("  %s: %s\n", key, value)
                }
            }
        }
    }

    fmt.Println("\nDone!")
}
```

## API Reference

### Client

#### `NewClient() *Client`
Creates a new GoAnime client with all available scrapers initialized.

#### `SearchAnime(query string, source *types.Source) ([]*types.Anime, error)`
Searches for anime by name. If `source` is `nil`, searches across all sources.

#### `GetAnimeEpisodes(animeURL string, source types.Source) ([]*types.Episode, error)`
Retrieves all episodes for a specific anime using its URL and source.

#### `GetStreamURL(episodeURL string, source types.Source, options ...interface{}) (string, map[string]string, error)`
Gets the streaming URL and required headers for a specific episode.

#### `GetAvailableSources() []types.Source`
Returns a list of all available scraper sources.

### Types

#### `types.Source`
- `types.SourceAllAnime` - AllAnime source
- `types.SourceAnimeFire` - AnimeFire source

#### `types.Anime`
Represents an anime with properties like Name, URL, ImageURL, Episodes, AnilistID, etc.

#### `types.Episode`
Represents an episode with properties like Number, URL, Title, Duration, etc.

## Error Handling

All methods return errors that should be properly handled:

```go
results, err := client.SearchAnime("One Piece", nil)
if err != nil {
    // Handle error appropriately
    log.Printf("Search failed: %v", err)
    return
}
```

## Notes

- The library automatically handles rate limiting and retries for API calls
- Stream URLs may expire after some time
- Some sources may require specific headers for streaming
- Not all anime metadata (like AniList ID) may be available for all sources
