package providers

import (
	"context"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeSearchQuery(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		"o-todo-poderoso":    "o todo poderoso",
		"the-boys":           "the boys",
		"jujutsu_kaisen":     "jujutsu kaisen",
		"the---boys":         "the boys",
		"  naruto  ":         "naruto",
		"naruto":             "naruto",
		"":                   "",
		"cowboy bebop":       "cowboy bebop",
		"re:zero-season-2":   "re:zero season 2",
		"fate/stay-night":    "fate/stay night",
		"mobile-suit-gundam": "mobile suit gundam",
	} {
		assert.Equalf(t, want, normalizeSearchQuery(in), "input %q", in)
	}
}

// recordingSource is a Searchable that records the query it was handed.
type recordingSource struct {
	mu   sync.Mutex
	kind source.SourceKind
	got  []string
}

func (r *recordingSource) Describe() source.Descriptor {
	return source.Descriptor{Kind: r.kind, Priority: 999}
}

func (r *recordingSource) Search(_ context.Context, query string) ([]*models.Anime, error) {
	r.mu.Lock()
	r.got = append(r.got, query)
	r.mu.Unlock()
	return []*models.Anime{{Name: "hit", Source: string(r.kind)}}, nil
}

// The rest of source.Source, unused here: this double exists to observe the
// query, not to answer anything.
func (r *recordingSource) FetchEpisodes(context.Context, *models.Anime) ([]models.Episode, error) {
	return nil, nil
}

func (r *recordingSource) FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error) {
	return "", nil
}

func (r *recordingSource) queries() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.got...)
}

// The fix has to land where EVERY source sees it.
//
// SuperFlix undid the dashes inside its own client and the other three never
// did, which is exactly the failure this guards: a per-source workaround looks
// like a fix until you count how many sources have it. Asserting on what the
// source actually receives is the only way to catch a new source that would be
// added without one.
func TestSearchAll_HandsSourcesWordsNotASlug(t *testing.T) {
	// Not parallel: registers into the process-wide source registry.
	rec := &recordingSource{kind: source.SourceKind("QueryProbe")}
	defer source.SwapRegistryForTesting(rec)()

	results, err := SearchAll(context.Background(), "o-todo-poderoso", rec.kind)
	require.NoError(t, err)
	assert.NotEmpty(t, results)

	got := rec.queries()
	require.Len(t, got, 1)
	assert.Equal(t, "o todo poderoso", got[0],
		"the source was handed the slug; a site's search engine treats the dash as a literal character")
}
