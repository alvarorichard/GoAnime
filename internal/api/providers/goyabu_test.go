package providers_test

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/providers"
)

func TestGoyabuProvider_Name(t *testing.T) {
	p := providers.NewGoyabuProvider()
	if p.Name() != "Goyabu" {
		t.Errorf("Name() = %q, want %q", p.Name(), "Goyabu")
	}
	if p.HasSeasons() {
		t.Errorf("HasSeasons() = %v, want false", p.HasSeasons())
	}
}
