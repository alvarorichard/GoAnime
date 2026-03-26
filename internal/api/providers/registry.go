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
	"allanime":   func() EpisodeProvider { return NewAllAnimeProvider() },
	"animefire":  func() EpisodeProvider { return NewAnimeFireProvider() },
	"animedrive": func() EpisodeProvider { return NewAnimeDriveProvider() },
	"flixhq":     func() EpisodeProvider { return NewFlixHQProvider() },
	"9anime":     func() EpisodeProvider { return NewNineAnimeProvider() },
	"goyabu":     func() EpisodeProvider { return NewGoyabuProvider() },
}

// ForSource resolves the appropriate EpisodeProvider for the given anime.
// It uses a priority chain: Source field > Name tags > URL patterns > fallback (AllAnime).
// Never returns nil.
func ForSource(anime *models.Anime) EpisodeProvider {
	name := ResolveSourceName(anime)
	return getOrCreateProvider(name)
}

// ForSourceName returns the EpisodeProvider for a given normalized source name string.
// Useful when the caller already knows the source key (e.g. from CLI --source flag).
func ForSourceName(source string) EpisodeProvider {
	name := normalizeSource(source)
	if name == "" {
		name = "allanime"
	}
	return getOrCreateProvider(name)
}

// ResolveSourceName determines the normalized source key for an anime using the priority chain.
func ResolveSourceName(anime *models.Anime) string {
	if anime == nil {
		return "allanime"
	}

	if n := normalizeSource(anime.Source); n != "" {
		return n
	}

	if anime.MediaType == models.MediaTypeMovie || anime.MediaType == models.MediaTypeTV {
		return "flixhq"
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

	return "allanime"
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
		factory = sourceRegistry["allanime"]
	}
	p := factory()
	providerCache[name] = p
	return p
}

// --- Convenience helpers (replace scattered is*Source functions) ---

func IsAllAnime(anime *models.Anime) bool {
	return ResolveSourceName(anime) == "allanime"
}

func IsAnimeFire(anime *models.Anime) bool {
	return ResolveSourceName(anime) == "animefire"
}

func IsAnimeDrive(anime *models.Anime) bool {
	return ResolveSourceName(anime) == "animedrive"
}

func IsFlixHQ(anime *models.Anime) bool {
	return ResolveSourceName(anime) == "flixhq"
}

func Is9Anime(anime *models.Anime) bool {
	return ResolveSourceName(anime) == "9anime"
}

func IsGoyabu(anime *models.Anime) bool {
	return ResolveSourceName(anime) == "goyabu"
}

// --- Source detection internals ---

var sourceAliases = map[string]string{
	"allanime":     "allanime",
	"animefire":    "animefire",
	"animefire.io": "animefire",
	"animedrive":   "animedrive",
	"flixhq":       "flixhq",
	"sflix":        "flixhq",
	"movie":        "flixhq",
	"tv":           "flixhq",
	"9anime":       "9anime",
	"nineanime":    "9anime",
	"goyabu":       "goyabu",
	"ptbr":         "animefire",
	"pt-br":        "animefire",
}

func normalizeSource(source string) string {
	if source == "" {
		return ""
	}
	key := strings.ToLower(strings.TrimSpace(source))
	if n, ok := sourceAliases[key]; ok {
		return n
	}
	if strings.Contains(key, "animefire") {
		return "animefire"
	}
	if strings.Contains(key, "allanime") {
		return "allanime"
	}
	return ""
}

func detectFromTags(name string) string {
	lower := strings.ToLower(name)

	if strings.Contains(lower, "[english]") {
		return "allanime"
	}
	if strings.Contains(lower, "[multilanguage]") {
		return "9anime"
	}
	if strings.Contains(lower, "[movie]") || strings.Contains(lower, "[tv]") || strings.Contains(lower, "[movies/tv]") {
		return "flixhq"
	}
	if strings.Contains(lower, "[pt-br]") || strings.Contains(lower, "[português]") || strings.Contains(lower, "[portuguese]") {
		return ""
	}
	return ""
}

func detectFromURL(urlStr string) string {
	if urlStr == "" {
		return ""
	}
	lower := strings.ToLower(urlStr)

	if strings.Contains(lower, "animefire") {
		return "animefire"
	}
	if strings.Contains(lower, "animesdrive") {
		return "animedrive"
	}
	if strings.Contains(lower, "goyabu") {
		return "goyabu"
	}
	if strings.Contains(lower, "flixhq") {
		return "flixhq"
	}
	if strings.Contains(lower, "allanime") {
		return "allanime"
	}

	if isLikelyAllAnimeID(urlStr) {
		return "allanime"
	}

	return ""
}

// DetectFromTagsAndURL combines tag and URL detection for PT-BR disambiguation.
// PT-BR tags don't uniquely identify a source, so URL patterns are needed.
func DetectFromTagsAndURL(name, urlStr string) string {
	lower := strings.ToLower(name)

	if strings.Contains(lower, "[pt-br]") || strings.Contains(lower, "[português]") || strings.Contains(lower, "[portuguese]") {
		lowerURL := strings.ToLower(urlStr)
		if strings.Contains(lowerURL, "animesdrive") {
			return "animedrive"
		}
		if strings.Contains(lowerURL, "goyabu") {
			return "goyabu"
		}
		return "animefire"
	}
	return ""
}

func isLikelyAllAnimeID(s string) bool {
	if strings.Contains(s, "http") {
		return false
	}
	if len(s) < 6 || len(s) >= 30 {
		return false
	}
	hasLetter := false
	for _, c := range s {
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
			hasLetter = true
			break
		}
	}
	return hasLetter
}
