# GoAnime Enhanced Web Scraping Integration

This integration adds powerful web scraping capabilities to GoAnime, inspired by the popular `ani-cli` script. It supports multiple anime streaming sources with automatic fallback and enhanced download features.

##  New Features

### Multi-Source Support
- **AllAnime.day**: High-quality streams with multiple resolution options
- **AnimeFire.plus**: Brazilian anime streaming site with Portuguese content
- **Automatic Fallback**: If one source fails, automatically tries others

### Enhanced CLI Options
```bash
# New command-line flags
--source <source>     # Specify anime source (allanime, animefire)
--quality <quality>   # Specify video quality (best, worst, 720p, 1080p, etc.)
```

### Quality Selection
- **best**: Automatically selects the highest available quality
- **worst**: Selects the lowest quality (for limited bandwidth)
- **720p, 1080p, 480p**: Specific resolution selection
- **hls**: HLS/m3u8 streams for better compatibility

##  Usage Examples

### Basic Usage
```bash
# Search all sources
goanime "naruto"

# Download with specific source
goanime -d --source allanime "one piece" 1

# Download with quality preference
goanime -d --quality 720p "attack on titan" 5

# Download range with specific source and quality
goanime -d -r --source animefire --quality best "demon slayer" 1-12
```

### Advanced Usage
```bash
# Use AllAnime for high-quality content
goanime -d --source allanime --quality 1080p "jujutsu kaisen" 10

# Use AnimeFire for Portuguese content
goanime -d --source animefire "naruto" 25

# Let the system choose the best source automatically
goanime -d --quality best "bleach" 100
```

##  Technical Implementation

### Architecture Overview
```
internal/
├── scraper/
│   ├── allanime.go     # AllAnime.day scraper
│   ├── animefire.go    # AnimeFire.plus scraper
│   └── unified.go      # Unified scraper interface
├── api/
│   └── enhanced.go     # Enhanced API with multi-source support
└── download/
    └── workflow.go     # Updated download workflow
```

### Scraper Interface
```go
type UnifiedScraper interface {
    SearchAnime(query string, options ...interface{}) ([]*models.Anime, error)
    GetAnimeEpisodes(animeURL string) ([]models.Episode, error)
    GetStreamURL(episodeURL string, options ...interface{}) (string, map[string]string, error)
    GetType() ScraperType
}
```

### Features Implemented
1. **GraphQL API Integration** (AllAnime)
2. **HTML Parsing** (AnimeFire)
3. **Video Link Extraction**
4. **Quality Selection Logic**
5. **Error Handling with Fallbacks**
6. **Metadata Extraction**



### Command Equivalents
```bash
# ani-cli examples -> GoAnime equivalents

# ani-cli -d "anime name" episode
goanime -d "anime name" episode

# ani-cli -d -r "anime name" 1-5
goanime -d -r "anime name" 1-5

# ani-cli -q 720p "anime name"
goanime -d --quality 720p "anime name" 1
```

##  Configuration

### Environment Variables
```bash
# Set default quality
export GOANIME_DEFAULT_QUALITY=720p

# Set default source
export GOANIME_DEFAULT_SOURCE=allanime

# Set download directory
export GOANIME_DOWNLOAD_DIR=/path/to/downloads
```

### Source Priority
When no source is specified, the system tries sources in this order:
1. AllAnime (generally higher quality)
2. AnimeFire (fallback option)

## 🐛 Troubleshooting

### Common Issues

**No results found**
```bash
# Try different sources
goanime -d --source animefire "your anime" 1
```

**Stream URL not found**
```bash
# Try different quality
goanime -d --quality worst "your anime" 1
```

**Download fails**
```bash
# Enable debug mode
goanime --debug -d "your anime" 1
```

### Debug Mode
Enable verbose logging to troubleshoot issues:
```bash
goanime --debug -d --source allanime "your anime" 1
```

##  Future Enhancements

### Planned Features
- [ ] Additional streaming sources
- [ ] Subtitle download and embedding
- [ ] Playlist generation (m3u8)
- [ ] Resume interrupted downloads
- [ ] Parallel episode downloads
- [ ] Quality auto-detection based on bandwidth
- [ ] Source quality comparison
- [ ] Custom user-agent rotation
- [ ] Proxy support for geo-restricted content

### Integration Ideas
- [ ] MyAnimeList integration for metadata
- [ ] AniList sync for watch progress
- [ ] Discord Rich Presence with source info
- [ ] Web UI for remote control
- [ ] Mobile app companion

##  Contributing

To add a new anime source:

1. Create a new scraper in `internal/scraper/newsource.go`
2. Implement the `UnifiedScraper` interface
3. Add the scraper to `ScraperManager`
4. Update the CLI flags and help text
5. Add tests and documentation

Example scraper template:
```go
type NewSourceClient struct {
    client *http.Client
    baseURL string
}

func (c *NewSourceClient) SearchAnime(query string, options ...interface{}) ([]*models.Anime, error) {
    // Implementation here
}
```

