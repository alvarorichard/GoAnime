package source

import (
	"context"
	"sort"
	"strings"
	"sync"

	"github.com/alvarorichard/Goanime/internal/models"
)

// Descriptor is a source's self-declared identity (Model B).
// It carries the source's matching data plus an explicit
// Priority, so resolution order is deterministic data — never init() order.
type Descriptor struct {
	Kind        SourceKind
	Priority    int                // Lower value = matched first among non-explicit criteria
	Explicit    []string           // Values that may appear in anime.Source
	Tags        []string           // Lowercase tags in anime.Name, e.g. "[animefire]"
	URLMatchers []string           // Lowercase substrings to match in anime.URL
	MediaTypes  []models.MediaType // MediaType values that map to this source
	ShortID     bool               // If true, accepts AllAnime-style short alphanumeric IDs

	// DefaultDisabled marks a source that is OFF unless the user opts in via
	// GOANIME_ENABLED_SOURCES (ARCHITECTURE.md §7 S1). Use it for experimental
	// or fragile sources that shouldn't ship live. Independent of the always-
	// available GOANIME_DISABLED_SOURCES kill-switch.
	DefaultDisabled bool
}

// matchNonExplicit checks all match criteria except the explicit Source field.
// Resolve handles explicit matching in a separate first pass so the Source
// field always wins regardless of Priority ordering.
func (d Descriptor) matchNonExplicit(anime *models.Anime) (string, bool) {
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

// matchURL checks whether a raw URL matches this descriptor's URL patterns.
func (d Descriptor) matchURL(url string) (string, bool) {
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
	snapshot := make([]Source, 0, len(registry))
	for _, s := range registry {
		snapshot = append(snapshot, s)
	}
	registryMu.RUnlock()

	// Drop config-disabled sources so they never resolve (S1 kill-switch).
	srcs := filterEnabled(snapshot)

	sort.Slice(srcs, func(i, j int) bool {
		di, dj := srcs[i].Describe(), srcs[j].Describe()
		if di.Priority != dj.Priority {
			return di.Priority < dj.Priority
		}
		return di.Kind < dj.Kind
	})
	return srcs
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
