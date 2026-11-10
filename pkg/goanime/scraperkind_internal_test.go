package goanime

import (
	"testing"

	apisource "github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var scraperKindCases = []struct {
	st   scraper.ScraperType
	want apisource.SourceKind
	ok   bool
}{
	{scraper.AnimefireType, apisource.AnimeFire, true},
	{scraper.GoyabuType, apisource.Goyabu, true},
	{scraper.StartFlixType, apisource.StartFlix, true},
	{scraper.TopCineType, apisource.TopCine, true},
	{scraper.HiAnimeType, apisource.HiAnime, true},
	{scraper.AnimesDigitalType, apisource.AnimesDigital, true},
	{scraper.AnimeDriveType, apisource.AnimeDrive, true},
	{scraper.PobreflixType, apisource.Pobreflix, true},
	{scraper.AnrollType, apisource.Anroll, true},
	{scraper.PomfyType, apisource.Pomfy, true},
	{scraper.ScraperType(9999), "", false},
	{scraper.ScraperType(-1), "", false},
}

func scraperKindFailures(impl func(scraper.ScraperType) (apisource.SourceKind, bool)) []string {
	var f []string
	for _, tc := range scraperKindCases {
		got, ok := impl(tc.st)
		if got != tc.want || ok != tc.ok {
			f = append(f, "type "+string(rune('0'+int(tc.st)%10))+": mismatch")
		}
	}
	return f
}

func TestScraperKind(t *testing.T) {
	assert.Empty(t, scraperKindFailures(scraperKind))
}

func TestScraperKindMeta(t *testing.T) {
	require.Empty(t, scraperKindFailures(scraperKind))
	mutants := map[string]func(scraper.ScraperType) (apisource.SourceKind, bool){
		"always animefire": func(scraper.ScraperType) (apisource.SourceKind, bool) { return apisource.AnimeFire, true },
		"unknown accepted": func(st scraper.ScraperType) (apisource.SourceKind, bool) {
			k, _ := scraperKind(st)
			return k, true
		},
		"known rejected": func(st scraper.ScraperType) (apisource.SourceKind, bool) {
			k, _ := scraperKind(st)
			return k, false
		},
		"goyabu and startflix swapped": func(st scraper.ScraperType) (apisource.SourceKind, bool) {
			switch st {
			case scraper.GoyabuType:
				return apisource.StartFlix, true
			case scraper.StartFlixType:
				return apisource.Goyabu, true
			}
			return scraperKind(st)
		},
		"topcine missing": func(st scraper.ScraperType) (apisource.SourceKind, bool) {
			if st == scraper.TopCineType {
				return "", false
			}
			return scraperKind(st)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, scraperKindFailures(m), "mutant %q survived", name)
	}
}

func TestRegistryManagerRejectsUnknownScraperType(t *testing.T) {
	bad := scraper.ScraperType(9999)
	got, err := registryManager{}.SearchAnime("x", &bad)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown scraper type")
	assert.Nil(t, got)
}

func TestRegistryManagerGetScraperBuildsAnAdapter(t *testing.T) {
	scr, err := registryManager{}.GetScraper(scraper.AnimefireType)
	require.NoError(t, err)
	require.NotNil(t, scr)
	assert.Equal(t, scraper.AnimefireType, scr.GetType())

	_, err = registryManager{}.GetScraper(scraper.ScraperType(9999))
	assert.Error(t, err, "an unknown type has no adapter")
}

func TestNewClientUsesTheRegistry(t *testing.T) {
	c := NewClient()
	require.NotNil(t, c)
	assert.IsType(t, registryManager{}, c.manager)
}
