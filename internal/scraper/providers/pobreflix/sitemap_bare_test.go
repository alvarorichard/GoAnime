package pobreflix

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 2026-10-26: the sitemaps began listing title pages without ".html"
// (https://pobreflixtv.gay/filme/<slug>), while the catalog API and every page
// link still use the ".html" form. The corpus only knew the ".html" form, so
// it indexed nothing, and a search the API answered with no hits ended as
// "sitemap shards produced no catalog entries" — an error, where an empty
// result was the truth. The URLs below are copied from the live sitemaps.

func TestEntryFromURL_BareSitemapURLs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		raw      string
		wantKind MediaKind
		wantURL  string
		wantBase string
	}{
		{"https://pobreflixtv.gay/filme/todo-poderoso", KindMovie, "https://pobreflixtv.gay/filme/todo-poderoso.html", "todo poderoso"},
		{"https://pobreflixtv.gay/serie/1-contra-todos", KindSeries, "https://pobreflixtv.gay/serie/1-contra-todos.html", "1 contra todos"},
		{"https://pobreflixtv.gay/filme/todo-poderoso.html", KindMovie, "https://pobreflixtv.gay/filme/todo-poderoso.html", "todo poderoso"},
		{"https://pobreflixtv.gay/serie/011ze-nova-geracao-2026?x=1#y", KindSeries, "https://pobreflixtv.gay/serie/011ze-nova-geracao-2026.html", "011ze nova geracao 2026"},
	}
	for _, tt := range tests {
		t.Run(tt.raw, func(t *testing.T) {
			t.Parallel()
			e, ok := entryFromURL(tt.raw)
			require.True(t, ok, "a title page the sitemap lists must be indexed")
			assert.Equal(t, tt.wantKind, e.Kind)
			assert.Equal(t, tt.wantURL, e.URL, "the app opens the .html form")
			assert.Equal(t, tt.wantBase, e.Tokens)
		})
	}
}

func TestEntryFromURL_BareEpisodesAndNoiseStayOut(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{
		"https://pobreflixtv.gay/assistir/stranger-things-t1-ep1",
		"https://pobreflixtv.gay/assistir/stranger-things-t1-ep1.html",
		"https://pobreflixtv.gay/sitemap-movies.xml",
		"https://pobreflixtv.gay/filme/",
		"",
	} {
		_, ok := entryFromURL(raw)
		assert.False(t, ok, "%q is not a title page", raw)
	}
}

func TestRank_FindsBareSitemapTitles(t *testing.T) {
	t.Parallel()
	cc := &corpus{}
	for _, raw := range []string{
		"https://pobreflixtv.gay/filme/todo-poderoso",
		"https://pobreflixtv.gay/filme/a-volta-do-todo-poderoso",
		"https://pobreflixtv.gay/serie/ben-hur",
	} {
		e, ok := entryFromURL(raw)
		require.True(t, ok)
		cc.entries = append(cc.entries, e)
	}

	out := cc.rank("todo poderoso")

	require.NotEmpty(t, out)
	assert.Equal(t, "Todo Poderoso", out[0].Title)
	assert.Equal(t, "https://pobreflixtv.gay/filme/todo-poderoso.html", out[0].URL)
}
