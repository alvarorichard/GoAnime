package source

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// Descriptor is a source's self-declared identity (Model B).
// It carries the same matching data as SourceDefinition plus an explicit
// Priority, so resolution order is deterministic data — never init() order.
type Descriptor struct {
	Kind        SourceKind
	Priority    int                // Lower value = matched first among non-explicit criteria
	Explicit    []string           // Values that may appear in anime.Source
	Tags        []string           // Lowercase tags in anime.Name, e.g. "[animefire]"
	URLMatchers []string           // Lowercase substrings to match in anime.URL
	MediaTypes  []models.MediaType // MediaType values that map to this source
	ShortID     bool               // If true, accepts AllAnime-style short alphanumeric IDs
}

// definition converts a Descriptor to a SourceDefinition so both resolution
// paths share the exact same matching logic during the migration.
func (d Descriptor) definition() SourceDefinition {
	return SourceDefinition{
		Kind:        d.Kind,
		Explicit:    d.Explicit,
		Tags:        d.Tags,
		URLMatchers: d.URLMatchers,
		MediaTypes:  d.MediaTypes,
		ShortID:     d.ShortID,
	}
}

// Source is a self-describing media source: identity (Describe) and behavior
// (Fetch*) live in one unit, registered from the source's own init().
type Source interface {
	Describe() Descriptor
	FetchEpisodes(ctx context.Context, anime *models.Anime) ([]models.Episode, error)
	FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error)
}

var (
	registryMu sync.RWMutex
	registry   = make(map[SourceKind]Source)
)

// Register adds a source to the registry. Called from init() in each source's
// package. Registering the same Kind again replaces the previous entry, so
// re-registration is safe in tests.
func Register(s Source) {
	if s == nil {
		panic("source: Register called with nil Source")
	}
	d := s.Describe()
	if d.Kind == "" {
		panic("source: Register called with empty Descriptor.Kind")
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[d.Kind] = s
}

// Registered returns the Source registered for kind, if any.
func Registered(kind SourceKind) (Source, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	s, ok := registry[kind]
	return s, ok
}

// registeredByPriority returns a snapshot of the registry ordered by
// (Priority, Kind). The Kind tie-break keeps resolution deterministic even if
// two sources declare the same Priority.
func registeredByPriority() []Source {
	registryMu.RLock()
	srcs := make([]Source, 0, len(registry))
	for _, s := range registry {
		srcs = append(srcs, s)
	}
	registryMu.RUnlock()

	sort.Slice(srcs, func(i, j int) bool {
		di, dj := srcs[i].Describe(), srcs[j].Describe()
		if di.Priority != dj.Priority {
			return di.Priority < dj.Priority
		}
		return di.Kind < dj.Kind
	})
	return srcs
}

// ResolveSource determines the source for an anime by scanning the registered
// sources' Describe() data. Same precedence as Resolve:
//
//  1. Explicit anime.Source field (checked across ALL sources first)
//  2. anime.MediaType
//  3. Tags in anime.Name
//  4. URL pattern / short ID
//
// If nothing matches, returns (nil, Kind=Unknown) with a warning log — the
// caller decides whether to fall back (BestEffortKind + Registered).
func ResolveSource(anime *models.Anime) (Source, ResolvedSource) {
	if anime == nil {
		return nil, ResolvedSource{Kind: Unknown, Reason: "nil anime"}
	}

	srcs := registeredByPriority()

	// Priority 1: Explicit Source field (highest priority, check all sources first)
	if anime.Source != "" {
		for _, s := range srcs {
			d := s.Describe()
			for _, e := range d.Explicit {
				if anime.Source == e {
					return s, ResolvedSource{Kind: d.Kind, Reason: "explicit Source=" + e}
				}
			}
		}
	}

	// Priority 2+: MediaType, tags, URL, shortID (lowest Priority wins)
	for _, s := range srcs {
		d := s.Describe()
		def := d.definition()
		if reason, ok := def.matchNonExplicit(anime); ok {
			return s, ResolvedSource{Kind: d.Kind, Reason: reason}
		}
	}

	// PT-BR tag without specific source → default AnimeFire
	if anime.Name != "" {
		lower := strings.ToLower(anime.Name)
		if strings.Contains(lower, "[pt-br]") || strings.Contains(lower, "[portugu") {
			if s, ok := Registered(AnimeFire); ok {
				return s, ResolvedSource{Kind: AnimeFire, Reason: "PT-BR language tag (default AnimeFire)"}
			}
		}
	}

	util.Warn("source resolution fell through to Unknown", "anime", anime.Name, "url", anime.URL)
	return nil, ResolvedSource{Kind: Unknown, Reason: "no match, best-effort AllAnime"}
}

// ResolveSourceURL resolves a source from a raw URL string only, scanning the
// registered sources. Used when no models.Anime context is available.
func ResolveSourceURL(rawURL string) (Source, ResolvedSource) {
	if rawURL == "" {
		return nil, ResolvedSource{Kind: Unknown, Reason: "empty URL"}
	}

	for _, s := range registeredByPriority() {
		d := s.Describe()
		def := d.definition()
		if reason, ok := def.matchURL(rawURL); ok {
			return s, ResolvedSource{Kind: d.Kind, Reason: reason}
		}
	}

	return nil, ResolvedSource{Kind: Unknown, Reason: "URL not matched"}
}

// SwapRegistryForTesting replaces the registry with the given sources and
// returns a restore func. Only for tests.
func SwapRegistryForTesting(srcs ...Source) (restore func()) {
	registryMu.Lock()
	old := registry
	registry = make(map[SourceKind]Source, len(srcs))
	for _, s := range srcs {
		registry[s.Describe().Kind] = s
	}
	registryMu.Unlock()

	return func() {
		registryMu.Lock()
		registry = old
		registryMu.Unlock()
	}
}
