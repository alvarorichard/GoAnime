package source

import (
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
)

// SourceDefinition is a declarative description of a media source.
// Adding a new source to GoAnime costs one entry in the sourceDefs slice
// and one Provider implementation — no Resolve() changes needed.
type SourceDefinition struct {
	Kind        SourceKind
	Explicit    []string           // Values that may appear in anime.Source
	Tags        []string           // Lowercase tags in anime.Name, e.g. "[animefire]"
	URLMatchers []string           // Lowercase substrings to match in anime.URL
	MediaTypes  []models.MediaType // MediaType values that map to this source
	ShortID     bool               // If true, accepts AllAnime-style short alphanumeric IDs
}

// sourceDefs is the ordered list of source definitions.
// Resolve() iterates this slice and returns the first match.
// CONTRACT: more specific entries come first (e.g. AnimeDrive before AnimeFire
// because "animesdrive" contains "anime").
var sourceDefs = []SourceDefinition{
	{
		Kind:        AnimeDrive,
		Explicit:    []string{"AnimeDrive"},
		Tags:        []string{"[animedrive]"},
		URLMatchers: []string{"animesdrive"},
	},
	{
		Kind:        AnimeFire,
		Explicit:    []string{"Animefire.io", "AnimeFire"},
		Tags:        []string{"[animefire]"},
		URLMatchers: []string{"animefire"},
	},
	{
		Kind:        Goyabu,
		Explicit:    []string{"Goyabu"},
		Tags:        []string{"[goyabu]"},
		URLMatchers: []string{"goyabu"},
	},
	{
		Kind:        SuperFlix,
		Explicit:    []string{"SuperFlix"},
		Tags:        []string{"[superflix]"},
		URLMatchers: []string{"superflix"},
	},
	{
		Kind:        FlixHQ,
		Explicit:    []string{"FlixHQ"},
		Tags:        []string{"[flixhq]", "[movie]", "[tv]"},
		URLMatchers: []string{"flixhq"},
		MediaTypes:  []models.MediaType{models.MediaTypeMovie, models.MediaTypeTV},
	},
	{
		Kind:        SFlix,
		Explicit:    []string{"SFlix"},
		Tags:        []string{"[sflix]"},
		URLMatchers: []string{"sflix"},
	},
	{
		Kind:        NineAnime,
		Explicit:    []string{"9Anime"},
		Tags:        []string{"[9anime]", "[multilanguage]"},
		URLMatchers: []string{"9anime"},
	},
	{
		Kind:        AllAnime,
		Explicit:    []string{"AllAnime"},
		Tags:        []string{"[english]"},
		URLMatchers: []string{"allanime"},
		ShortID:     true,
	},
}

// matchNonExplicit checks all match criteria except explicit Source field.
// Resolve() handles explicit matching in a separate first pass to ensure
// Source field always wins regardless of sourceDefs ordering.
func (d *SourceDefinition) matchNonExplicit(anime *models.Anime) (string, bool) {
	// Priority 2: MediaType
	if anime.MediaType != "" {
		for _, mt := range d.MediaTypes {
			if anime.MediaType == mt {
				return "MediaType=" + string(mt), true
			}
		}
	}

	// Priority 3: Name tags
	if anime.Name != "" {
		lower := strings.ToLower(anime.Name)
		for _, tag := range d.Tags {
			if strings.Contains(lower, tag) {
				return "tag " + tag, true
			}
		}
	}

	// Priority 4: URL patterns
	if anime.URL != "" {
		lowerURL := strings.ToLower(anime.URL)
		for _, pat := range d.URLMatchers {
			if strings.Contains(lowerURL, pat) {
				return "URL contains " + pat, true
			}
		}
	}

	// Priority 5: Short ID (AllAnime-style)
	if d.ShortID && IsAllAnimeShortID(anime.URL) {
		return "short ID", true
	}

	return "", false
}

// matchURL checks whether a raw URL matches this definition's URL patterns.
func (d *SourceDefinition) matchURL(url string) (string, bool) {
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
