package suggest

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
)

// Kind is the sort of title a suggestion names.
type Kind string

// The kinds the autocomplete tells apart.
const (
	KindAnime  Kind = "anime"
	KindMovie  Kind = "movie"
	KindSeries Kind = "series"
)

// Defaults for a zero Options value.
const (
	// MinQuery is the shortest query the engine answers.
	MinQuery = 2
	// MaxQuery caps the query length (in runes) sent upstream.
	MaxQuery        = 64
	defaultLimit    = 8
	defaultTimeout  = 2500 * time.Millisecond
	defaultCacheLen = 256
	defaultTTL      = 10 * time.Minute
)

// ErrAllFailed means every provider asked for a query failed.
var ErrAllFailed = errors.New("suggest: every provider failed")

// Suggestion is one autocomplete row.
type Suggestion struct {
	Title  string
	Year   int // 0 when unknown
	Kind   Kind
	Source string // provider name, for diagnostics
}

// Detail is the row's secondary text: "Anime · 2002", "Series".
func (s Suggestion) Detail() string {
	label := ""
	switch s.Kind {
	case KindAnime:
		label = "Anime"
	case KindMovie:
		label = "Movie"
	case KindSeries:
		label = "Series"
	}
	switch {
	case label != "" && s.Year > 0:
		return fmt.Sprintf("%s · %d", label, s.Year)
	case s.Year > 0:
		return fmt.Sprintf("%d", s.Year)
	default:
		return label
	}
}

// Provider is one upstream suggestion source.
type Provider interface {
	// Name identifies the provider in the cache and in diagnostics.
	Name() string
	// Kinds lists the kinds of title the provider can return.
	Kinds() []Kind
	// Suggest returns up to limit suggestions for an already normalized query.
	Suggest(ctx context.Context, query string, limit int) ([]Suggestion, error)
}

// Options configure an Engine; zero fields take the defaults.
type Options struct {
	Providers []Provider
	// Limit caps the suggestions returned (default 8).
	Limit int
	// Timeout caps each provider call (default 2.5s).
	Timeout time.Duration
	// CacheSize is how many provider answers stay cached (default 256).
	CacheSize int
	// TTL is how long a cached answer lives (default 10 minutes).
	TTL time.Duration
	// Now is the clock; nil means time.Now.
	Now func() time.Time
}

// Engine fans a query out to its providers, then merges, ranks and caches.
// It is safe for concurrent use.
type Engine struct {
	providers []Provider
	limit     int
	timeout   time.Duration
	cache     *cache
}

// New builds an Engine from opts.
func New(opts Options) *Engine {
	e := &Engine{
		providers: opts.Providers,
		limit:     opts.Limit,
		timeout:   opts.Timeout,
	}
	if e.limit <= 0 {
		e.limit = defaultLimit
	}
	if e.timeout <= 0 {
		e.timeout = defaultTimeout
	}
	size, ttl, now := opts.CacheSize, opts.TTL, opts.Now
	if size <= 0 {
		size = defaultCacheLen
	}
	if ttl <= 0 {
		ttl = defaultTTL
	}
	if now == nil {
		now = time.Now
	}
	e.cache = newCache(size, ttl, now)
	return e
}

// Normalize trims a query, collapses its whitespace and caps its length. It
// returns "" for anything shorter than MinQuery runes.
func Normalize(query string) string {
	q := strings.Join(strings.FieldsFunc(query, unicode.IsSpace), " ")
	if r := []rune(q); len(r) > MaxQuery {
		q = strings.TrimSpace(string(r[:MaxQuery]))
	}
	if len([]rune(q)) < MinQuery {
		return ""
	}
	return q
}

// KindsForScope maps the search home's scope selector to the kinds worth
// suggesting there: "" is everything, the anime scopes want anime only, the
// movie scope wants movies and series.
func KindsForScope(scope string) []Kind {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "ptbr", "hianime":
		return []Kind{KindAnime}
	case "movietv":
		return []Kind{KindMovie, KindSeries}
	default:
		return nil
	}
}

// Suggest returns ranked suggestions for query, limited to kinds (nil means
// every kind). A query that is too short, or kinds no provider serves, gives
// no suggestions and no error. When some providers fail the others still
// answer; ErrAllFailed (wrapping each cause) is returned only when none did.
func (e *Engine) Suggest(ctx context.Context, query string, kinds []Kind) ([]Suggestion, error) {
	q := Normalize(query)
	if q == "" {
		return nil, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	chosen := e.providersFor(kinds)
	if len(chosen) == 0 {
		return nil, nil
	}

	type answer struct {
		items []Suggestion
		err   error
	}
	answers := make([]answer, len(chosen))
	var wg sync.WaitGroup
	for i, p := range chosen {
		wg.Add(1)
		go func() {
			defer wg.Done()
			answers[i].items, answers[i].err = e.ask(ctx, p, q)
		}()
	}
	wg.Wait()

	if err := ctx.Err(); err != nil {
		return nil, err
	}
	var errs []error
	var lists [][]Suggestion
	for i, a := range answers {
		if a.err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", chosen[i].Name(), a.err))
			continue
		}
		lists = append(lists, a.items)
	}
	if len(lists) == 0 {
		return nil, fmt.Errorf("%w: %w", ErrAllFailed, errors.Join(errs...))
	}
	return rank(q, merge(lists, kinds), e.limit), nil
}

// providersFor keeps the providers serving at least one wanted kind.
func (e *Engine) providersFor(kinds []Kind) []Provider {
	if len(kinds) == 0 {
		return e.providers
	}
	var out []Provider
	for _, p := range e.providers {
		if servesAny(p.Kinds(), kinds) {
			out = append(out, p)
		}
	}
	return out
}

func servesAny(have, want []Kind) bool {
	for _, h := range have {
		for _, w := range want {
			if h == w {
				return true
			}
		}
	}
	return false
}

// ask returns a provider's answer from the cache, or asks upstream under the
// per-provider timeout and caches a successful answer.
func (e *Engine) ask(ctx context.Context, p Provider, q string) ([]Suggestion, error) {
	key := p.Name() + "\x00" + fold(q)
	if items, ok := e.cache.get(key); ok {
		return items, nil
	}
	callCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()
	items, err := p.Suggest(callCtx, q, e.limit)
	if err != nil {
		return nil, err
	}
	e.cache.put(key, items)
	return items, nil
}

// merge concatenates the provider lists in provider order, drops kinds not
// asked for, rows without a title, and later rows repeating an earlier title.
func merge(lists [][]Suggestion, kinds []Kind) []Suggestion {
	seen := make(map[string]bool)
	var out []Suggestion
	for _, list := range lists {
		for _, s := range list {
			s.Title = strings.Join(strings.Fields(s.Title), " ")
			key := fold(s.Title)
			if key == "" || seen[key] {
				continue
			}
			if len(kinds) > 0 && !servesAny([]Kind{s.Kind}, kinds) {
				continue
			}
			seen[key] = true
			out = append(out, s)
		}
	}
	return out
}

// rank orders suggestions by how well the title matches the typed text
// (exact, prefix, word prefix, substring, other), shortest first among
// prefixes and in provider order inside the other tiers, then cuts the list to limit.
func rank(q string, items []Suggestion, limit int) []Suggestion {
	needle := fold(q)
	tier := func(s Suggestion) int {
		title := fold(s.Title)
		switch {
		case title == needle:
			return 0
		case strings.HasPrefix(title, needle):
			return 1
		case strings.Contains(" "+title, " "+needle):
			return 2
		case strings.Contains(title, needle):
			return 3
		default:
			return 4
		}
	}
	sort.SliceStable(items, func(i, j int) bool {
		ti, tj := tier(items[i]), tier(items[j])
		if ti != tj {
			return ti < tj
		}
		// Among completions of what was typed, the shortest reads first:
		// "Naruto" before "Naruto: Shippuuden" for "narut".
		if ti == 1 {
			return len(fold(items[i].Title)) < len(fold(items[j].Title))
		}
		return false
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items
}

// fold lowercases s and keeps letters and digits, turning every other run
// into one space, so "Re:Zero" and "re zero" compare equal.
func fold(s string) string {
	var sb strings.Builder
	space := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			if space && sb.Len() > 0 {
				sb.WriteByte(' ')
			}
			space = false
			sb.WriteRune(unicode.ToLower(r))
			continue
		}
		space = true
	}
	return sb.String()
}

var (
	defaultOnce   sync.Once
	defaultEngine *Engine
)
