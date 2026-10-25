package api

import (
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/startflix"
	"github.com/stretchr/testify/assert"
)

// Movies offered only on unplayable hosts (old titles whose panel lists
// only long-dead file hosts) are stale like dead panels: hidden from later
// searches in the same session once the yt-dlp fallback has missed.
func TestIsPanelDeadErr_UnsupportedOnly(t *testing.T) {
	t.Parallel()
	assert.True(t, isPanelDeadErr(&startflix.NoStreamError{Unsupported: []string{"actelecup.com"}}))
	assert.True(t, isPanelDeadErr(&friendlyError{
		cause: &startflix.NoStreamError{Unsupported: []string{"actelecup.com"}},
		msg:   "wrapped",
	}), "friendlyError layers must not hide the cause")
	assert.False(t, isPanelDeadErr(&startflix.NoStreamError{
		Unsupported: []string{"x.test"},
		Failures:    []error{errors.New("boom")},
	}), "a host that failed (vs unknown) may heal on retry")
	assert.True(t, isPanelDeadErr(startflix.ErrNotOnPanel))
	assert.False(t, isPanelDeadErr(errors.New("plain")))
}
