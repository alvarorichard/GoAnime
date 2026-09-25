# GoAnime Enhanced Web Scraping Integration

This integration adds powerful web scraping capabilities to GoAnime, inspired by the popular `ani-cli` script. It supports multiple anime streaming sources with automatic fallback and enhanced download features.

##  New Features

### Multi-Source Support
- **hianime.at**: Subbed and dubbed HLS streams with multiple resolutions
- **Animefire.io**: Brazilian anime streaming site with Portuguese content
- **Automatic Fallback**: If one source fails, automatically tries others

### Enhanced CLI Options
```bash
# New command-line flags
--source <source>     # Specify source (hianime, animefire, superflix)
                      # goyabu is off by default: GOANIME_ENABLED_SOURCES=goyabu
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
goanime -d --source hianime "one piece" 1

# Download with quality preference
goanime -d --quality 720p "attack on titan" 5

# Download range with specific source and quality
goanime -d -r --source animefire --quality best "demon slayer" 1-12
```

### Advanced Usage
```bash
# Use HiAnime for subbed/dubbed content
goanime -d --source hianime --quality 1080p "jujutsu kaisen" 10

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
│   ├── hianime/        # hianime.at scraper
│   ├── animefire.go    # Animefire.io scraper
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
1. **Frontend API Integration** (HiAnime)
2. **HTML Parsing** (AnimeFire)
3. **Video Link Extraction**
4. **Quality Selection Logic**
5. **Error Handling with Fallbacks**
6. **Metadata Extraction**

### HiAnime Provider Notes

This slot has held three hosts. AllAnime went first, when mkissa.to removed the
material its per-request key was derived from. anidb.app took over and then went
to a site-wide 503 "Under Maintenance" — homepage, `/browse` and its API alike —
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
  lives in a third-party repository — the same dependency shape that made
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
go test ./internal/scraper -run TestProcessSourceURLsConcurrentFallsBackToFast4SpeedDirectSource -count=1
```



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
export GOANIME_DEFAULT_SOURCE=HiAnime

# Set download directory
export GOANIME_DOWNLOAD_DIR=/path/to/downloads
```

### Source Priority
When no source is specified, the system tries sources in this order:
1. HiAnime (generally higher quality)
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
goanime --debug -d --source HiAnime "your anime" 1
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

