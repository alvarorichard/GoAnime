package source

import (
	"context"
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// plainSource implements only the base Source interface — no optional
// capabilities. It is the "simple anime source" the Model C design keeps free
// of season/browser methods.
type plainSource struct{ kind SourceKind }

func (p plainSource) Describe() Descriptor { return Descriptor{Kind: p.kind} }
func (p plainSource) FetchEpisodes(context.Context, *models.Anime) ([]models.Episode, error) {
	return nil, nil
}
func (p plainSource) FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error) {
	return "", nil
}

// seasonedSource additionally implements Seasoned.
type seasonedSource struct {
	plainSource
	hasSeasons bool
}

func (s seasonedSource) HasSeasons() bool { return s.hasSeasons }

// gatedSource additionally implements BrowserGated.
type gatedSource struct {
	plainSource
	warmUpErr  error
	warmUpCtx  context.Context
	warmUpRuns int
}

func (g *gatedSource) WarmUp(ctx context.Context) error {
	g.warmUpRuns++
	g.warmUpCtx = ctx
	return g.warmUpErr
}

func TestIsSeasoned(t *testing.T) {
	t.Parallel()
	assert.False(t, IsSeasoned(plainSource{kind: "plain"}),
		"a source without the Seasoned capability is not seasoned")
	assert.True(t, IsSeasoned(seasonedSource{plainSource{"tv"}, true}),
		"a Seasoned source reporting true is seasoned")
	assert.False(t, IsSeasoned(seasonedSource{plainSource{"tv"}, false}),
		"a Seasoned source may still report false")
}

func TestIsBrowserGated(t *testing.T) {
	t.Parallel()
	assert.False(t, IsBrowserGated(plainSource{kind: "plain"}),
		"a pure-HTTP source is not browser-gated")
	assert.True(t, IsBrowserGated(&gatedSource{plainSource: plainSource{"gated"}}),
		"a source implementing BrowserGated is browser-gated")
}

func TestWarmUp(t *testing.T) {
	t.Parallel()

	t.Run("non-gated source is an explicit no-op", func(t *testing.T) {
		t.Parallel()
		err := WarmUp(context.Background(), plainSource{kind: "plain"})
		assert.NoError(t, err, "a non-gated source must warm up cleanly (no-op)")
	})

	t.Run("gated source that can run returns nil", func(t *testing.T) {
		t.Parallel()
		g := &gatedSource{plainSource: plainSource{"gated"}}
		err := WarmUp(context.Background(), g)
		require.NoError(t, err)
		assert.Equal(t, 1, g.warmUpRuns, "WarmUp must be invoked exactly once")
	})

	t.Run("gated source that cannot run surfaces the error", func(t *testing.T) {
		t.Parallel()
		sentinel := errors.New("no display")
		g := &gatedSource{plainSource: plainSource{"gated"}, warmUpErr: sentinel}
		err := WarmUp(context.Background(), g)
		require.ErrorIs(t, err, sentinel, "a doomed browser environment must be surfaced, not swallowed")
	})

	t.Run("context is threaded to the capability", func(t *testing.T) {
		t.Parallel()
		type ctxKey struct{}
		ctx := context.WithValue(context.Background(), ctxKey{}, "v")
		g := &gatedSource{plainSource: plainSource{"gated"}}
		require.NoError(t, WarmUp(ctx, g))
		assert.Equal(t, "v", g.warmUpCtx.Value(ctxKey{}), "the caller's ctx must reach WarmUp")
	})
}
