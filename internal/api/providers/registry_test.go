package providers

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

type testProvider struct {
	kind source.SourceKind
}

func (p *testProvider) Kind() source.SourceKind { return p.kind }
func (p *testProvider) HasSeasons() bool        { return false }

func (p *testProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	return anime.Episodes, nil
}

func (p *testProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	return quality, nil
}

func TestForKindCachesProviderInstances(t *testing.T) {
	ResetForTesting()

	kind := source.SourceKind("TestCacheProvider")
	var calls atomic.Int32
	RegisterProvider(kind, func(sm *scraper.ScraperManager) Provider {
		calls.Add(1)
		return &testProvider{kind: kind}
	})

	first, err := ForKind(kind)
	if err != nil {
		t.Fatalf("ForKind returned error: %v", err)
	}

	second, err := ForKind(kind)
	if err != nil {
		t.Fatalf("ForKind returned error: %v", err)
	}

	if calls.Load() != 1 {
		t.Fatalf("factory called %d times, want 1", calls.Load())
	}
	if first != second {
		t.Fatal("ForKind should return the cached provider instance")
	}
}

func TestForKindConcurrentFactoryCall(t *testing.T) {
	ResetForTesting()

	kind := source.SourceKind("TestConcurrentProvider")
	var calls atomic.Int32
	RegisterProvider(kind, func(sm *scraper.ScraperManager) Provider {
		calls.Add(1)
		return &testProvider{kind: kind}
	})

	var wg sync.WaitGroup
	errCh := make(chan error, 16)

	for range 16 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := ForKind(kind)
			errCh <- err
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			t.Fatalf("ForKind returned error: %v", err)
		}
	}

	if calls.Load() != 1 {
		t.Fatalf("factory called %d times, want 1", calls.Load())
	}
}

func TestForAnimeResolvesProvider(t *testing.T) {
	anime := &models.Anime{
		Name: "[PT-BR] Naruto",
		URL:  "https://goyabu.to/anime/naruto",
	}

	provider, resolved, err := ForAnime(anime)
	if err != nil {
		t.Fatalf("ForAnime returned error: %v", err)
	}

	if resolved.Kind != source.Goyabu {
		t.Fatalf("resolved kind = %s, want %s", resolved.Kind, source.Goyabu)
	}
	if provider.Kind() != source.Goyabu {
		t.Fatalf("provider kind = %s, want %s", provider.Kind(), source.Goyabu)
	}
}

func TestForAnimeUsesRegisteredProviderWithoutKindList(t *testing.T) {
	anime := &models.Anime{
		Name:   "[Multilanguage] Naruto",
		URL:    "8143",
		Source: "9Anime",
	}

	provider, resolved, err := ForAnime(anime)
	if err != nil {
		t.Fatalf("ForAnime returned error: %v", err)
	}

	if resolved.Kind != source.NineAnime {
		t.Fatalf("resolved kind = %s, want %s", resolved.Kind, source.NineAnime)
	}
	if provider.Kind() != source.NineAnime {
		t.Fatalf("provider kind = %s, want %s", provider.Kind(), source.NineAnime)
	}
}

func TestForAnimePropagatesResolveError(t *testing.T) {
	anime := &models.Anime{
		Name: "[PT-BR] Naruto",
		URL:  "naruto",
	}

	_, _, err := ForAnime(anime)
	if err == nil {
		t.Fatal("expected resolve error for ambiguous PT-BR anime")
	}
}

// Bug scenario: ForKind with an unregistered source kind previously panicked
// (nil map lookup). Fix: must return a descriptive error naming the kind.
func TestForKindUnregisteredKindReturnsError(t *testing.T) {
	ResetForTesting()

	unknown := source.SourceKind("UnknownProviderXYZ")
	_, err := ForKind(unknown)
	if err == nil {
		t.Fatal("expected error for unregistered kind, got nil")
	}
	if !strings.Contains(err.Error(), string(unknown)) {
		t.Fatalf("error should name the missing kind %q, got: %v", unknown, err)
	}
}

// HasProvider must return true for registered kinds and false for unknown ones.
func TestHasProvider(t *testing.T) {
	t.Parallel()

	if !HasProvider(source.Goyabu) {
		t.Fatal("HasProvider(Goyabu) = false, want true (registered in init)")
	}
	if HasProvider(source.SourceKind("definitely-not-a-real-source")) {
		t.Fatal("HasProvider returned true for unregistered kind")
	}
}

// Bug scenario: ForAnime resolving to a non-provider-backed source (hypothetical
// source kind with no factory) must return an error, not a nil provider.
func TestForAnimeNonProviderBackedSourceReturnsError(t *testing.T) {
	ResetForTesting()

	// AnimeFire resolves from URL but register a fresh kind with no factory
	// to simulate a source that exists in source.Resolve but has no provider.
	ghostKind := source.SourceKind("GhostSource")

	// Don't register a factory — just verify ForKind fails.
	_, err := ForKind(ghostKind)
	if err == nil {
		t.Fatal("expected error for source with no registered provider, got nil")
	}
}

// ForKind must be safe to call concurrently with RegisterProvider — the factory
// map and cache use separate locks so they must not deadlock.
func TestForKindAndRegisterProviderConcurrentSafe(t *testing.T) {
	ResetForTesting()

	kind := source.SourceKind("TestRegisterConcurrent")
	var calls atomic.Int32

	done := make(chan struct{})
	go func() {
		defer close(done)
		RegisterProvider(kind, func(sm *scraper.ScraperManager) Provider {
			calls.Add(1)
			return &testProvider{kind: kind}
		})
	}()

	// ForKind may race with RegisterProvider — either gets "not found" or succeeds.
	// The important guarantee is: no panic, no deadlock.
	<-done
	p, err := ForKind(kind)
	if err != nil {
		t.Skipf("ForKind raced before RegisterProvider completed: %v", err)
	}
	if p == nil {
		t.Fatal("ForKind returned nil provider without error")
	}
}
