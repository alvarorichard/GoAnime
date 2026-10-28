package providers

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// An empty origin must not clear one already set, the same way an empty
// referer does not.
func TestApplyPlaybackMetadataKeepsOriginWhenAbsent(t *testing.T) {
	t.Cleanup(func() {
		util.ClearGlobalReferer()
		util.ClearGlobalOrigin()
	})
	util.SetGlobalOrigin("https://kept.example")
	applyPlaybackMetadata(map[string]string{"referer": "https://ref.example/"})
	if got := util.GetGlobalOrigin(); got != "https://kept.example" {
		t.Fatalf("origin = %q, want it kept", got)
	}
	applyPlaybackMetadata(map[string]string{"origin": "https://new.example"})
	if got := util.GetGlobalOrigin(); got != "https://new.example" {
		t.Fatalf("origin = %q, want the new one", got)
	}
}
