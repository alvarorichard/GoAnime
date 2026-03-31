package providers

import (
	"strings"
	"sync"

	"github.com/alvarorichard/Goanime/internal/models"
)

var (
	providerCache   = make(map[string]EpisodeProvider)
	providerCacheMu sync.RWMutex
)

type providerFactory func() EpisodeProvider

var sourceRegistry = map[string]providerFactory{
	"goyabu": func() EpisodeProvider { return NewGoyabuProvider() },
}

// ForSource resolves the appropriate EpisodeProvider for the given anime.
// It uses a priority chain: Source field > Name tags > URL patterns > fallback.
// Returns nil if the source hasn't been migrated yet.
func ForSource(anime *models.Anime) EpisodeProvider {
	name := ResolveSourceName(anime)
	if name == "" {
		return nil
	}
	return getOrCreateProvider(name)
}

// ForSourceName returns the EpisodeProvider for a given normalized source name string.
// Returns nil if the source hasn't been migrated yet.
func ForSourceName(source string) EpisodeProvider {
	name := normalizeSource(source)
	if name == "" {
		return nil
	}
	return getOrCreateProvider(name)
}

// ResolveSourceName determines the normalized source key for an anime using the priority chain.
func ResolveSourceName(anime *models.Anime) string {
	if anime == nil {
		return ""
	}

	if n := normalizeSource(anime.Source); n != "" {
		return n
	}

	if n := detectFromTags(anime.Name); n != "" {
		return n
	}

	if n := DetectFromTagsAndURL(anime.Name, anime.URL); n != "" {
		return n
	}

	if n := detectFromURL(anime.URL); n != "" {
		return n
	}

	return ""
}

func getOrCreateProvider(name string) EpisodeProvider {
	providerCacheMu.RLock()
	if p, ok := providerCache[name]; ok {
		providerCacheMu.RUnlock()
		return p
	}
	providerCacheMu.RUnlock()

	providerCacheMu.Lock()
	defer providerCacheMu.Unlock()

	if p, ok := providerCache[name]; ok {
		return p
	}

	factory, exists := sourceRegistry[name]
	if !exists {
		return nil // Not migrated yet
	}
	p := factory()
	providerCache[name] = p
	return p
}

// --- Convenience helpers (replace scattered is*Source functions) ---

// IsGoyabu checks if the anime source resolves to goyabu.
func IsGoyabu(anime *models.Anime) bool {
	return ResolveSourceName(anime) == "goyabu"
}

// --- Source detection internals ---

var sourceAliases = map[string]string{
	"goyabu": "goyabu",
}

func normalizeSource(source string) string {
	if source == "" {
		return ""
	}
	key := strings.ToLower(strings.TrimSpace(source))
	if n, ok := sourceAliases[key]; ok {
		return n
	}
	return ""
}

func detectFromTags(name string) string {
	return ""
}

func detectFromURL(urlStr string) string {
	if urlStr == "" {
		return ""
	}
	lower := strings.ToLower(urlStr)

	if strings.Contains(lower, "goyabu") {
		return "goyabu"
	}

	return ""
}

// DetectFromTagsAndURL combines tag and URL detection for PT-BR disambiguation.
// PT-BR tags don't uniquely identify a source, so URL patterns are needed.
func DetectFromTagsAndURL(name, urlStr string) string {
	lower := strings.ToLower(name)

	if strings.Contains(lower, "[pt-br]") || strings.Contains(lower, "[português]") || strings.Contains(lower, "[portuguese]") {
		lowerURL := strings.ToLower(urlStr)
		if strings.Contains(lowerURL, "goyabu") {
			return "goyabu"
		}
	}
	return ""
}
