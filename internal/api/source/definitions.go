// Package source defines source kinds and their associated metadata definitions.
package source

import "github.com/alvarorichard/Goanime/internal/models"

// SourceDefinition is a declarative description of a media source.
// Adding a new source to GoAnime costs one entry in the sourceDefs slice
// and one Provider implementation — no Resolve() changes needed.
//
//revive:disable-next-line:exported
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
