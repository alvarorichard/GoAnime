# GoAnime PT-BR Enhancement Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add robust PT-BR anime sources (AnimeFire complete + Goyabu new) with `-source ptbr` flag and legendado/dublado detection.

**Architecture:** Complete the existing AnimeFire scraper stubs, add Goyabu as a new source, register both as PT-BR sources searchable via `-source ptbr` flag.

**Tech Stack:** Go 1.25, goquery, charmbracelet TUI, existing UnifiedScraper pattern

---

### Task 1: Complete AnimeFire GetAnimeEpisodes in Scraper

**Files:**
- Modify: `internal/scraper/animefire.go:209-211` (replace stub)

**What to do:**
The `GetAnimeEpisodes` method is currently a stub returning an error. Implement it by:
1. Making an HTTP GET request to `animeURL` using the existing `decorateRequest` pattern
2. Parsing the HTML with goquery
3. Extracting episodes using selector `a.lEp.epT.divNumEp.smallbox.px-2.mx-1.text-left.d-flex`
4. Parsing episode numbers with regex `(?i)epis[oó]dio\s+(\d+)`
5. Sorting episodes by number

Reference: `internal/api/episodes.go` has the exact same logic but uses `SafeGet` instead of the client's HTTP setup.

**Implementation:**
```go
func (c *AnimefireClient) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	var lastErr error
	attempts := c.maxRetries + 1

	for attempt := 0; attempt < attempts; attempt++ {
		req, err := http.NewRequest("GET", animeURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		c.decorateRequest(req)

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("failed to fetch anime page: %w", err)
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		if resp.StatusCode != http.StatusOK {
			lastErr = c.handleStatusError(resp)
			_ = resp.Body.Close()
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		doc, err := goquery.NewDocumentFromReader(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("failed to parse HTML: %w", err)
		}

		if c.isChallengePage(doc) {
			lastErr = errors.New("animefire returned a challenge page (try VPN or wait)")
			if c.shouldRetry(attempt) {
				c.sleep()
				continue
			}
			return nil, lastErr
		}

		episodes := c.parseEpisodes(doc)
		sort.Slice(episodes, func(i, j int) bool {
			return episodes[i].Num < episodes[j].Num
		})

		return episodes, nil
	}

	if lastErr != nil {
		return nil, lastErr
	}
	return nil, errors.New("failed to retrieve episodes from AnimeFire")
}

func (c *AnimefireClient) parseEpisodes(doc *goquery.Document) []models.Episode {
	var episodes []models.Episode
	re := regexp.MustCompile(`(?i)epis[oó]dio\s+(\d+)`)

	doc.Find("a.lEp.epT.divNumEp.smallbox.px-2.mx-1.text-left.d-flex").Each(func(i int, s *goquery.Selection) {
		episodeNum := strings.TrimSpace(s.Text())
		episodeURL, _ := s.Attr("href")

		num := i + 1
		matches := re.FindStringSubmatch(episodeNum)
		if len(matches) >= 2 {
			if parsed, err := strconv.Atoi(matches[1]); err == nil {
				num = parsed
			}
		}

		episodes = append(episodes, models.Episode{
			Number: episodeNum,
			Num:    num,
			URL:    c.resolveURL(c.baseURL, episodeURL),
		})
	})

	return episodes
}
```

Add imports: `"regexp"`, `"sort"`, `"strconv"` to animefire.go.

---

### Task 2: Complete AnimeFire GetEpisodeStreamURL in Scraper

**Files:**
- Modify: `internal/scraper/animefire.go:215-218` (replace stub)

**What to do:**
AnimeFire episode pages contain video data in a `<video>` element or via a JSON/API endpoint. The stream URL extraction:
1. Fetch the episode page HTML
2. Look for `video[data-video-src]` attribute or `<source>` tag with `src`
3. Also check for an API endpoint pattern like `/api/v1/episode/{id}` in embedded scripts
4. Return the direct video URL

**Implementation:**
```go
func (c *AnimefireClient) GetEpisodeStreamURL(episodeURL string) (string, error) {
	req, err := http.NewRequest("GET", episodeURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	c.decorateRequest(req)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch episode page: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", c.handleStatusError(resp)
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse episode HTML: %w", err)
	}

	// Strategy 1: Look for video data-video-src attribute
	if src, exists := doc.Find("video[data-video-src]").Attr("data-video-src"); exists && src != "" {
		return src, nil
	}

	// Strategy 2: Look for source tag inside video element
	if src, exists := doc.Find("video source").Attr("src"); exists && src != "" {
		return src, nil
	}

	// Strategy 3: Look for iframe with video embed
	if src, exists := doc.Find("iframe").Attr("src"); exists && src != "" {
		return src, nil
	}

	// Strategy 4: Look for data-src or data-url attributes on video containers
	var streamURL string
	doc.Find("[data-video-src], [data-src], [data-url]").Each(func(i int, s *goquery.Selection) {
		if streamURL != "" {
			return
		}
		for _, attr := range []string{"data-video-src", "data-src", "data-url"} {
			if val, exists := s.Attr(attr); exists && val != "" {
				streamURL = val
				return
			}
		}
	})

	if streamURL != "" {
		return streamURL, nil
	}

	return "", fmt.Errorf("could not find stream URL in episode page")
}
```

---

### Task 3: Register AnimeFire as Complete PT-BR Source

**Files:**
- Modify: `internal/api/enhanced.go` — update source routing for "animefire" in episodes and stream
- Modify: `internal/scraper/unified.go` — update language tag from `[Portuguese]` to `[PT-BR]`

**What to do:**
1. In `unified.go:getLanguageTag`, change AnimeFire/AnimeDrive from `[Portuguese]` to `[PT-BR]`
2. In `enhanced.go:GetAnimeEpisodesEnhanced`, update the AnimeFire branch to use the scraper directly instead of the API layer
3. Update tag references in `enhanced.go` from `[Portuguese]` to `[PT-BR]`

---

### Task 4: Add `-source ptbr` Flag

**Files:**
- Modify: `internal/util/util.go` — add "ptbr"/"pt-br" to source flag help
- Modify: `internal/api/enhanced.go:SearchAnimeEnhanced` — add ptbr source routing
- Modify: `internal/scraper/unified.go` — add `SearchAnimePTBR` method

**What to do:**
1. Add `"ptbr"` and `"pt-br"` as valid source values in `FlagParser`
2. In `SearchAnimeEnhanced`, when source is "ptbr" or "pt-br", search only PT-BR scrapers (AnimeFire + Goyabu when available)
3. Add a `SearchAnimePTBR` method to `ScraperManager` that runs only PT-BR scrapers concurrently

---

### Task 5: Add Legendado/Dublado Tags

**Files:**
- Modify: `internal/scraper/animefire.go` — detect dub/sub from anime name/URL
- Modify: `internal/scraper/unified.go:tagResults` — append dub/sub info

**What to do:**
AnimeFire URLs and titles contain indicators like "dublado" or "legendado". Detect these and add `(Legendado)` or `(Dublado)` after the `[PT-BR]` tag.

---

### Task 6: Create Goyabu Scraper

**Files:**
- Create: `internal/scraper/goyabu.go`
- Modify: `internal/scraper/unified.go` — add GoyabuType and adapter

**What to do:**
Create a new scraper for goyabu.cc following the AnimeFire pattern:
1. `GoyabuClient` struct with HTTP client, baseURL, retry logic
2. `SearchAnime` — search via goyabu.cc search
3. `GetAnimeEpisodes` — parse episode list
4. `GetEpisodeStreamURL` — extract video URL
5. Register as `GoyabuType` in unified.go with `[PT-BR]` tag

---

### Task 7: Manual Testing and Adjustments

Test the complete flow:
1. `go build ./cmd/goanime/`
2. Test AnimeFire search, episodes, streaming
3. Test `-source ptbr` flag
4. Test Goyabu if site is accessible
5. Fix any issues found

---

### Task 8: Prepare and Submit PR

1. Clean up code, ensure tests pass
2. Create PR to upstream alvarorichard/GoAnime
3. Write descriptive PR with feature summary
