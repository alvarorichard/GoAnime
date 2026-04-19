package source

import (
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
)

// SourceDefinition is a declarative description of a media source.
// Adding a new source to GoAnime costs one entry in the sourceDefs slice
// and one Provider implementation — no Resolve() changes needed.
type SourceDefinition struct {
	Kind             SourceKind
	Explicit         []string           // Exact values that may appear in anime.Source
	ExplicitContains []string           // Lowercase substrings accepted in anime.Source
	MediaTypes       []models.MediaType // MediaType values that map to this source
	Tags             []string           // Lowercase tags in anime.Name, e.g. "[animefire]"
	URLMatchers      []string           // Lowercase substrings to match in anime.URL
	ShortID          bool               // If true, accepts AllAnime-style short alphanumeric IDs
	PTBR             bool               // If true, generic [PT-BR] tags may resolve to this source via URL
}

// sourceDefs is the ordered list of source definitions.
// Resolve iterates this slice and returns the first match.
// CONTRACT: first match wins, more specific entries must come first.
var sourceDefs = []SourceDefinition{
	{
		Kind:             SuperFlix,
		Explicit:         []string{"SuperFlix"},
		ExplicitContains: []string{"superflix"},
		Tags:             []string{"[superflix]"},
		URLMatchers:      []string{"superflix"},
		PTBR:             true,
	},
	{
		Kind:             SFlix,
		Explicit:         []string{"SFlix", "sflix"},
		ExplicitContains: []string{"sflix"},
		Tags:             []string{"[sflix]"},
		URLMatchers:      []string{"sflix"},
	},
	{
		Kind:             AnimeDrive,
		Explicit:         []string{"AnimeDrive"},
		ExplicitContains: []string{"animedrive"},
		Tags:             []string{"[animedrive]"},
		URLMatchers:      []string{"animesdrive"},
		PTBR:             true,
	},
	{
		Kind:             AnimeFire,
		Explicit:         []string{"Animefire.io", "AnimeFire"},
		ExplicitContains: []string{"animefire"},
		Tags:             []string{"[animefire]"},
		URLMatchers:      []string{"animefire"},
		PTBR:             true,
	},
	{
		Kind:             Goyabu,
		Explicit:         []string{"Goyabu"},
		ExplicitContains: []string{"goyabu"},
		Tags:             []string{"[goyabu]"},
		URLMatchers:      []string{"goyabu"},
		PTBR:             true,
	},
	{
		Kind:             FlixHQ,
		Explicit:         []string{"FlixHQ", "movie", "tv"},
		ExplicitContains: []string{"flixhq"},
		Tags:             []string{"[flixhq]", "[movie]", "[tv]"},
		URLMatchers:      []string{"flixhq"},
		MediaTypes:       []models.MediaType{models.MediaTypeMovie, models.MediaTypeTV},
	},
	{
		Kind:             NineAnime,
		Explicit:         []string{"9Anime", "NineAnime", "nineanime"},
		ExplicitContains: []string{"nineanime"},
		Tags:             []string{"[9anime]", "[multilanguage]"},
		URLMatchers:      []string{"9anime"},
	},
	{
		Kind:             AllAnime,
		Explicit:         []string{"AllAnime"},
		ExplicitContains: []string{"allanime"},
		Tags:             []string{"[english]"},
		URLMatchers:      []string{"allanime"},
		ShortID:          true,
	},
}

func (d SourceDefinition) matchesExplicit(value string) bool {
	lowerValue := strings.ToLower(strings.TrimSpace(value))
	for _, explicit := range d.Explicit {
		if lowerValue == strings.ToLower(explicit) {
			return true
		}
	}
	for _, pattern := range d.ExplicitContains {
		if strings.Contains(lowerValue, strings.ToLower(pattern)) {
			return true
		}
	}
	return false
}

func (d SourceDefinition) matchesMediaType(mediaType models.MediaType) bool {
	for _, candidate := range d.MediaTypes {
		if mediaType == candidate {
			return true
		}
	}
	return false
}

func (d SourceDefinition) matchesTag(name string) bool {
	lowerName := strings.ToLower(name)
	for _, tag := range d.Tags {
		if strings.Contains(lowerName, tag) {
			return true
		}
	}
	return false
}

// matchURL checks whether a raw URL matches this definition's URL patterns.
func (d SourceDefinition) matchURL(url string) (string, bool) {
	if url == "" {
		return "", false
	}
	lower := strings.ToLower(url)
	for _, pat := range d.URLMatchers {
		if strings.Contains(lower, pat) {
			return "URL contains " + pat, true
		}
	}
	if d.ShortID && IsAllAnimeShortID(url) {
		return "short ID", true
	}
	return "", false
}
