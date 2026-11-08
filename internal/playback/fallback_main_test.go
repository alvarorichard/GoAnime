//go:build !live

package playback

import (
	"context"
	"errors"
	"os"
	"testing"

	apisource "github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
)

// TestMain keeps every test off the network: the fallback finds nothing unless
// a test says what the other sources answer. The live tests (-tags live) use
// the real sources instead.
func TestMain(m *testing.M) {
	searchOnSource = func(context.Context, apisource.SourceKind, string) ([]*models.Anime, error) {
		return nil, errors.New("no network in tests")
	}
	os.Exit(m.Run())
}
