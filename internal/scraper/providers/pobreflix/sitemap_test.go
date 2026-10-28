package pobreflix

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The corpus's unit shapes, isolated from the network: this is where the
// slug grammar, the scoring ties and the fold bookkeeping get pinned down
// without a server.

// TestRank_TieBreaks checks the corpus-side sort keys: score first, the dub
// tie-break second, then the newer record id.
func TestRank_TieBreaks(t *testing.T) {
	t.Parallel()
	cc := &corpus{}
	add := func(url string) {
		if e, ok := entryFromURL(url); ok {
			cc.entries = append(cc.entries, e)
		}
	}
	add("https://x.test/filmes/online/todo-poderoso-legendado-7002/")
	add("https://x.test/filmes/online/todo-poderoso-dublado-7001/")
	add("https://x.test/series/online/todo-mundo-dublado-90/")
	add("https://x.test/filmes/online/todo-dublado-5/")

	out := cc.rank("todo")
	require.Len(t, out, 3, "the two same-title variants dedupe")
	assert.Equal(t, "Todo", out[0].Title, "exact match wins")
	assert.Equal(t, "Todo Poderoso", out[1].Title, "contains, dub over leg")
	assert.Equal(t, "Todo Mundo", out[2].Title)
}

// TestRank_SilentOnGlueWords: a query of only stopwords/years/lone letters
// must return nothing — never a page of containment noise.
func TestRank_SilentOnGlueWords(t *testing.T) {
	t.Parallel()
	cc := &corpus{}
	for _, u := range []string{
		"https://x.test/filmes/online/todo-poderoso-dublado-7001/",
		"https://x.test/series/online/ben-hur-a-serie-dublado-83633/",
	} {
		e, ok := entryFromURL(u)
		require.True(t, ok)
		cc.entries = append(cc.entries, e)
	}
	for _, q := range []string{"a", "de", "2003", "o", "que"} {
		assert.Empty(t, cc.rank(q), "query %q has no significant word", q)
	}
}

// TestFold_Bookkeeping: fetched shards are marked done only when their
// bodies merge, a full rebuild replaces entries instead of appending
// duplicates, and the fresh date settles only on a complete round.
func TestFold_Bookkeeping(t *testing.T) {
	t.Parallel()
	cc := &corpus{}

	var partial syncMap
	partial.store("shard_a", []string{"https://x.test/filmes/online/um-filme-dublado-1/"})
	cc.fold(&partial, true, true, 2)
	assert.Len(t, cc.entries, 1)
	assert.True(t, cc.shardsDone["shard_a"])
	assert.False(t, cc.fresh(), "1 of 2 shards: stale until the rest lands")

	// A retry that adds the missing shard completes the corpus.
	var second syncMap
	second.store("shard_b", []string{"https://x.test/series/online/uma-serie-dublado-2/"})
	cc.fold(&second, false, true, 2)
	assert.Len(t, cc.entries, 2, "a partial round appends")
	assert.True(t, cc.fresh(), "2 of 2 shards: the date moves")

	// A full rebuild is a replacement, never an append — refetching the
	// same two shards must not duplicate the entries.
	var rebuilt syncMap
	rebuilt.store("shard_a", []string{"https://x.test/filmes/online/um-filme-dublado-1/"})
	rebuilt.store("shard_b", []string{"https://x.test/series/online/uma-serie-dublado-2/"})
	cc.fold(&rebuilt, true, true, 2)
	assert.Len(t, cc.entries, 2, "the rescan replaces")
}

// TestSortShards_NumericOrder: shard 10 sorts after shard 2, not lexically
// before it.
func TestSortShards_NumericOrder(t *testing.T) {
	t.Parallel()
	in := []shard{
		{name: "sitemap_content_videobox_Video_10", url: "https://x.test/?f=10"},
		{name: "sitemap_content_videobox_Video_2", url: "https://x.test/?f=2"},
		{name: "sitemap_content_videobox_Video_1", url: "https://x.test/?f=1"},
	}
	sortShards(in)
	assert.Equal(t, "sitemap_content_videobox_Video_1", in[0].name)
	assert.Equal(t, "sitemap_content_videobox_Video_2", in[1].name)
	assert.Equal(t, "sitemap_content_videobox_Video_10", in[2].name)
}

// TestParseShardLocs: the locator regex reads the <loc> entries of the real
// fixture shard shape.
func TestParseShardLocs(t *testing.T) {
	t.Parallel()
	body := []byte(`<?xml version="1.0"?><urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` +
		`<url><loc>https://www.pobreflixtv.sale/filmes/online/verity-dublado-84506/</loc></url>` +
		`<url><loc>https://www.pobreflixtv.sale/series/online/ben-hur-a-serie-dublado-83633/</loc></url></urlset>`)
	locs := parseShardLocs(body)
	require.Len(t, locs, 2)
	assert.Equal(t, "https://www.pobreflixtv.sale/filmes/online/verity-dublado-84506/", locs[0])
}
