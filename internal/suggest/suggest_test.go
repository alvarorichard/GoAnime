package suggest

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeProvider answers from a script and counts its calls.
type fakeProvider struct {
	name  string
	kinds []Kind
	items []Suggestion
	err   error
	block bool // wait for ctx instead of answering
	calls atomic.Int32
	mu    sync.Mutex
	last  struct {
		query string
		limit int
	}
}

func (f *fakeProvider) Name() string  { return f.name }
func (f *fakeProvider) Kinds() []Kind { return f.kinds }
func (f *fakeProvider) Suggest(ctx context.Context, query string, limit int) ([]Suggestion, error) {
	f.calls.Add(1)
	f.mu.Lock()
	f.last.query, f.last.limit = query, limit
	f.mu.Unlock()
	if f.block {
		<-ctx.Done()
		return nil, ctx.Err()
	}
	return f.items, f.err
}

func anime(title string, year int) Suggestion {
	return Suggestion{Title: title, Year: year, Kind: KindAnime}
}
func series(title string) Suggestion { return Suggestion{Title: title, Kind: KindSeries} }
func movie(title string) Suggestion  { return Suggestion{Title: title, Kind: KindMovie} }

func titles(items []Suggestion) []string {
	out := make([]string, 0, len(items))
	for _, s := range items {
		out = append(out, s.Title)
	}
	return out
}

func TestDetail(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in   Suggestion
		want string
	}{
		{Suggestion{Kind: KindAnime, Year: 2002}, "Anime · 2002"},
		{Suggestion{Kind: KindMovie, Year: 1999}, "Movie · 1999"},
		{Suggestion{Kind: KindSeries}, "Series"},
		{Suggestion{Year: 2020}, "2020"},
		{Suggestion{}, ""},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, c.in.Detail())
	}
}

func TestNormalize(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "one piece", Normalize("  one \t piece\n"))
	assert.Empty(t, Normalize(" a "), "one rune is below MinQuery")
	assert.Empty(t, Normalize("   "))
	assert.Equal(t, "鬼滅", Normalize("鬼滅"), "min is counted in runes, not bytes")
	long := Normalize(string(make([]rune, 0)) + repeat("ab ", 60))
	assert.LessOrEqual(t, len([]rune(long)), MaxQuery)
	assert.Equal(t, long, trimSpace(long), "the cut never leaves trailing space")
}

func repeat(s string, n int) string {
	out := ""
	for range n {
		out += s
	}
	return out
}

func trimSpace(s string) string {
	for s != "" && s[len(s)-1] == ' ' {
		s = s[:len(s)-1]
	}
	return s
}

func TestKindsForScope(t *testing.T) {
	t.Parallel()
	assert.Nil(t, KindsForScope(""))
	assert.Nil(t, KindsForScope("whatever"))
	assert.Equal(t, []Kind{KindAnime}, KindsForScope("ptbr"))
	assert.Equal(t, []Kind{KindAnime}, KindsForScope(" HiAnime "))
	assert.Equal(t, []Kind{KindMovie, KindSeries}, KindsForScope("movietv"))
}

func TestFold(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "re zero", fold("Re:Zero"))
	assert.Equal(t, "re zero", fold("  re - zero!! "))
	assert.Equal(t, "鬼滅の刃", fold("鬼滅の刃"))
	assert.Equal(t, "", fold("::"))
	assert.Equal(t, "a1 b2", fold("A1/B2"))
}

func TestSuggestShortQueryAndNoProviders(t *testing.T) {
	t.Parallel()
	p := &fakeProvider{name: "a", kinds: []Kind{KindAnime}}
	e := New(Options{Providers: []Provider{p}})

	got, err := e.Suggest(t.Context(), " x ", nil)
	require.NoError(t, err)
	assert.Nil(t, got)
	assert.Zero(t, p.calls.Load(), "a too-short query never reaches a provider")

	got, err = e.Suggest(t.Context(), "naruto", []Kind{KindMovie})
	require.NoError(t, err)
	assert.Nil(t, got, "no provider serves movies")
	assert.Zero(t, p.calls.Load())
}

func TestSuggestMergesRanksAndDedupes(t *testing.T) {
	t.Parallel()
	kitsu := &fakeProvider{name: "k", kinds: []Kind{KindAnime}, items: []Suggestion{
		anime("Boruto: Naruto Next Generations", 2017), anime("Naruto", 2002), anime("Naruto: Shippuuden", 2007),
	}}
	imdb := &fakeProvider{name: "i", kinds: []Kind{KindMovie, KindSeries}, items: []Suggestion{
		series("Naruto"), series("Like Naruto"), movie("Naru Barker"), series("  "),
	}}
	e := New(Options{Providers: []Provider{kitsu, imdb}})

	got, err := e.Suggest(t.Context(), "naruto", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"Naruto", "Naruto: Shippuuden", "Boruto: Naruto Next Generations", "Like Naruto", "Naru Barker"}, titles(got))
	assert.Equal(t, KindAnime, got[0].Kind, "the anime copy wins the duplicate: kitsu answers first")
}

func TestSuggestRankTiers(t *testing.T) {
	t.Parallel()
	p := &fakeProvider{name: "k", kinds: []Kind{KindAnime}, items: []Suggestion{
		anime("Unrelated", 0), anime("Xnaruto", 0), anime("Big Naruto Show", 0), anime("Naruto Show", 0), anime("naruto", 0),
	}}
	e := New(Options{Providers: []Provider{p}})
	got, err := e.Suggest(t.Context(), "Naruto", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"naruto", "Naruto Show", "Big Naruto Show", "Xnaruto", "Unrelated"}, titles(got),
		"exact, prefix, word prefix, substring, other")
}

func TestSuggestShortestPrefixFirst(t *testing.T) {
	t.Parallel()
	p := &fakeProvider{name: "k", kinds: []Kind{KindAnime}, items: []Suggestion{
		anime("Naruto: Shippuuden", 2007), anime("Naruto the Movie", 2004), anime("Naruto", 2002),
	}}
	got, err := New(Options{Providers: []Provider{p}}).Suggest(t.Context(), "narut", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"Naruto", "Naruto the Movie", "Naruto: Shippuuden"}, titles(got))
}

func TestSuggestLimitAndKindFilter(t *testing.T) {
	t.Parallel()
	p := &fakeProvider{name: "m", kinds: []Kind{KindMovie, KindSeries}, items: []Suggestion{
		movie("Dune"), series("Dune: Prophecy"), movie("Dune Two"), movie("Dune Three"),
	}}
	e := New(Options{Providers: []Provider{p}, Limit: 2})

	got, err := e.Suggest(t.Context(), "dune", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"Dune", "Dune Two"}, titles(got), "shortest prefix first, then cut to the limit")
	assert.Equal(t, 2, p.last.limit, "the engine passes its limit upstream")

	got, err = e.Suggest(t.Context(), "dune two", []Kind{KindSeries})
	require.NoError(t, err)
	assert.Equal(t, []string{"Dune: Prophecy"}, titles(got), "movies are filtered out of a series-only scope")
}

func TestSuggestPartialAndTotalFailure(t *testing.T) {
	t.Parallel()
	boom := errors.New("boom")
	bad := &fakeProvider{name: "bad", kinds: []Kind{KindAnime}, err: boom}
	good := &fakeProvider{name: "good", kinds: []Kind{KindMovie}, items: []Suggestion{movie("Akira")}}

	got, err := New(Options{Providers: []Provider{bad, good}}).Suggest(t.Context(), "akira", nil)
	require.NoError(t, err, "one provider down never breaks the box")
	assert.Equal(t, []string{"Akira"}, titles(got))

	_, err = New(Options{Providers: []Provider{bad}}).Suggest(t.Context(), "akira", nil)
	require.ErrorIs(t, err, ErrAllFailed)
	require.ErrorIs(t, err, boom)
	assert.Contains(t, err.Error(), "bad:")
}

func TestSuggestCachesSuccessOnly(t *testing.T) {
	t.Parallel()
	good := &fakeProvider{name: "g", kinds: []Kind{KindAnime}, items: []Suggestion{anime("Akira", 1988)}}
	bad := &fakeProvider{name: "b", kinds: []Kind{KindMovie}, err: errors.New("down")}
	e := New(Options{Providers: []Provider{good, bad}})

	for range 3 {
		_, err := e.Suggest(t.Context(), "Akira", nil)
		require.NoError(t, err)
	}
	// "AKIRA" and "akira " fold to the same cache key.
	_, err := e.Suggest(t.Context(), "AKIRA ", nil)
	require.NoError(t, err)
	assert.EqualValues(t, 1, good.calls.Load(), "the success is cached across spelling")
	assert.EqualValues(t, 4, bad.calls.Load(), "a failure is never cached")
}

func TestSuggestCacheExpires(t *testing.T) {
	t.Parallel()
	now := time.Unix(1_000, 0)
	var mu sync.Mutex
	clock := func() time.Time { mu.Lock(); defer mu.Unlock(); return now }
	p := &fakeProvider{name: "g", kinds: []Kind{KindAnime}, items: []Suggestion{anime("Akira", 0)}}
	e := New(Options{Providers: []Provider{p}, TTL: time.Minute, Now: clock})

	_, _ = e.Suggest(t.Context(), "akira", nil)
	_, _ = e.Suggest(t.Context(), "akira", nil)
	assert.EqualValues(t, 1, p.calls.Load())
	mu.Lock()
	now = now.Add(2 * time.Minute)
	mu.Unlock()
	_, _ = e.Suggest(t.Context(), "akira", nil)
	assert.EqualValues(t, 2, p.calls.Load(), "an expired entry is asked again")
}

func TestSuggestTimeoutAndCancel(t *testing.T) {
	t.Parallel()
	slow := &fakeProvider{name: "slow", kinds: []Kind{KindAnime}, block: true}
	fast := &fakeProvider{name: "fast", kinds: []Kind{KindMovie}, items: []Suggestion{movie("Akira")}}

	e := New(Options{Providers: []Provider{slow, fast}, Timeout: 20 * time.Millisecond})
	start := time.Now()
	got, err := e.Suggest(t.Context(), "akira", nil)
	require.NoError(t, err)
	assert.Equal(t, []string{"Akira"}, titles(got), "the slow provider times out, the fast one answers")
	assert.Less(t, time.Since(start), 2*time.Second)

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err = e.Suggest(ctx, "akira", nil)
	require.ErrorIs(t, err, context.Canceled, "a dead context short-circuits")

	ctx, cancel = context.WithCancel(t.Context())
	e = New(Options{Providers: []Provider{slow}, Timeout: time.Minute})
	go func() { time.Sleep(20 * time.Millisecond); cancel() }()
	_, err = e.Suggest(ctx, "akira", nil)
	require.ErrorIs(t, err, context.Canceled)
	require.NotErrorIs(t, err, ErrAllFailed, "cancelling mid-flight returns the cancellation, not ErrAllFailed")
}

func TestNewDefaults(t *testing.T) {
	t.Parallel()
	e := New(Options{})
	assert.Equal(t, defaultLimit, e.limit)
	assert.Equal(t, defaultTimeout, e.timeout)
	assert.Equal(t, defaultCacheLen, e.cache.max)
	assert.Equal(t, defaultTTL, e.cache.ttl)
	got, err := e.Suggest(t.Context(), "anything", nil)
	require.NoError(t, err)
	assert.Nil(t, got, "no providers, no suggestions")
}

func TestSuggestConcurrent(t *testing.T) {
	t.Parallel()
	p := &fakeProvider{name: "g", kinds: []Kind{KindAnime}, items: []Suggestion{anime("Akira", 0), anime("Akira Two", 0)}}
	e := New(Options{Providers: []Provider{p}, CacheSize: 2})
	var wg sync.WaitGroup
	for i := range 32 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			q := []string{"akira", "akira t", "akira tw"}[i%3]
			got, err := e.Suggest(t.Context(), q, nil)
			assert.NoError(t, err)
			assert.NotEmpty(t, got)
		}()
	}
	wg.Wait()
}

func TestCacheLRUAndCopies(t *testing.T) {
	t.Parallel()
	now := time.Unix(0, 0)
	c := newCache(2, time.Hour, func() time.Time { return now })
	c.put("a", []Suggestion{{Title: "A"}})
	c.put("b", []Suggestion{{Title: "B"}})
	_, ok := c.get("a") // a is now the most recent
	require.True(t, ok)
	c.put("c", []Suggestion{{Title: "C"}})
	assert.Equal(t, 2, c.len())
	_, ok = c.get("b")
	assert.False(t, ok, "b was the least recently used")
	_, ok = c.get("a")
	assert.True(t, ok)

	c.put("a", []Suggestion{{Title: "A2"}})
	got, _ := c.get("a")
	assert.Equal(t, "A2", got[0].Title, "put overwrites in place")
	assert.Equal(t, 2, c.len())

	got[0].Title = "mutated"
	again, _ := c.get("a")
	assert.Equal(t, "A2", again[0].Title, "callers get copies")

	now = now.Add(2 * time.Hour)
	_, ok = c.get("a")
	assert.False(t, ok)
	assert.Equal(t, 1, c.len(), "the expired entry is dropped on read")
}
