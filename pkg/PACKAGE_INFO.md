# GoAnime Library Package

This directory contains the public API for using GoAnime as a library in other Go projects.

## Structure

```
pkg/
├── README.md                          # Main documentation
└── goanime/                          # Main package
    ├── client.go                     # Client implementation
    ├── client_test.go                # Unit tests
    ├── doc.go                        # Package documentation
    ├── README.md                     # Detailed usage guide
    ├── types/                        # Public type definitions
    │   ├── anime.go                  # Anime and Episode types
    │   └── source.go                 # Source enum and helpers
    └── examples/                     # Usage examples
        ├── search/                   # Basic search example
        ├── episodes/                 # Episode listing example
        ├── stream/                   # Stream URL example
        └── source_specific/          # Source-specific search
```

## Quick Links

- [Main Documentation](README.md)
- [API Reference](goanime/README.md)
- [Usage Examples](goanime/examples/)

## Installation

```bash
go get github.com/alvarorichard/Goanime
```

## Quick Example

```go
package main

import (
    "fmt"
    "github.com/alvarorichard/Goanime/pkg/goanime"
)

func main() {
    client := goanime.NewClient()
    results, _ := client.SearchAnime("One Piece", nil)
    
    for _, anime := range results {
        fmt.Printf("%s [%s]\n", anime.Name, anime.Source)
    }
}
```

## Features

✅ Multi-source anime search (AllAnime, AnimeFire)  
✅ Episode listing with metadata  
✅ Stream URL extraction  
✅ Rich anime metadata (AniList, MAL)  
✅ Type-safe API  
✅ Comprehensive examples  
✅ Full test coverage  

## Testing

Run tests:
```bash
go test ./pkg/goanime/...
```

Run tests without integration tests:
```bash
go test -short ./pkg/goanime/...
```

## Building Examples

```bash
# Build search example
go build -o search ./pkg/goanime/examples/search/

# Build episodes example
go build -o episodes ./pkg/goanime/examples/episodes/

# Build stream example
go build -o stream ./pkg/goanime/examples/stream/

# Build source-specific example
go build -o source ./pkg/goanime/examples/source_specific/
```

## Documentation

For complete documentation, see:
- [pkg/README.md](README.md) - Overview and quick start
- [pkg/goanime/README.md](goanime/README.md) - Detailed API documentation
- Examples in [pkg/goanime/examples/](goanime/examples/)

## License

MIT License - See [LICENSE](../LICENSE)
