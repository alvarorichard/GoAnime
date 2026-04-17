package providers

import (
	"fmt"
	"sync"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

// ProviderFactory creates a Provider given a ScraperManager.
// Registered via RegisterProvider during init().
type ProviderFactory func(sm *scraper.ScraperManager) Provider

var (
	factories   = make(map[source.SourceKind]ProviderFactory)
	factoriesMu sync.RWMutex

	cache   = make(map[source.SourceKind]Provider)
	cacheMu sync.RWMutex
)

// RegisterProvider registers a provider factory for a given SourceKind.
// Typically called from init() in each provider's file.
// This makes adding a new source self-contained: the provider file registers itself.
func RegisterProvider(kind source.SourceKind, factory ProviderFactory) {
	factoriesMu.Lock()
	defer factoriesMu.Unlock()
	factories[kind] = factory
}

// ForKind returns the Provider for a given SourceKind with lazy init and cache.
// Thread-safe via double-checked locking.
func ForKind(kind source.SourceKind) (Provider, error) {
	// Fast path: cached
	cacheMu.RLock()
	if p, ok := cache[kind]; ok {
		cacheMu.RUnlock()
		return p, nil
	}
	cacheMu.RUnlock()

	// Slow path: create
	cacheMu.Lock()
	defer cacheMu.Unlock()

	// Double-check after acquiring write lock
	if p, ok := cache[kind]; ok {
		return p, nil
	}

	factoriesMu.RLock()
	factory, ok := factories[kind]
	factoriesMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("no provider registered for source %s", kind)
	}

	sm := scraper.NewScraperManager()
	p := factory(sm)
	cache[kind] = p
	return p, nil
}

// ForAnime resolves the source and returns the appropriate provider.
// Convenience function combining source.Resolve() + ForKind().
func ForAnime(anime *models.Anime) (Provider, source.ResolvedSource, error) {
	resolved := source.Resolve(anime)
	kind := resolved.BestEffortKind()
	p, err := ForKind(kind)
	if err != nil {
		return nil, resolved, err
	}
	return p, resolved, nil
}

// HasProvider returns true if a provider is registered for the given SourceKind.
func HasProvider(kind source.SourceKind) bool {
	factoriesMu.RLock()
	defer factoriesMu.RUnlock()
	_, ok := factories[kind]
	return ok
}

// ResetForTesting clears the provider cache. Only for tests.
func ResetForTesting() {
	cacheMu.Lock()
	defer cacheMu.Unlock()
	cache = make(map[source.SourceKind]Provider)
}
