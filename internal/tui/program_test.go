package tui

import (
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// minimalModel is the simplest possible Bubble Tea model for construction tests.
type minimalModel struct{}

func (m minimalModel) Init() tea.Cmd                        { return nil }
func (m minimalModel) Update(tea.Msg) (tea.Model, tea.Cmd) { return m, nil }
func (m minimalModel) View() tea.View                      { return tea.NewView("") }

func TestBubbleTeaProgramOptions_ReturnsNonEmptySlice(t *testing.T) {
	t.Parallel()
	opts := BubbleTeaProgramOptions()
	assert.NotEmpty(t, opts, "BubbleTeaProgramOptions must return at least one option")
}

func TestBubbleTeaProgramOptions_AppendsExtra(t *testing.T) {
	t.Parallel()
	base := BubbleTeaProgramOptions()
	withExtra := BubbleTeaProgramOptions(tea.WithoutSignals())
	// Extra option causes the slice to grow
	assert.Greater(t, len(withExtra), len(base))
}

func TestNewProgram_ReturnsNonNil(t *testing.T) {
	t.Parallel()
	prog := NewProgram(minimalModel{})
	require.NotNil(t, prog, "NewProgram must return a non-nil *tea.Program")
}

func TestNewProgram_WithExtraOptions(t *testing.T) {
	t.Parallel()
	prog := NewProgram(minimalModel{}, tea.WithoutSignals())
	require.NotNil(t, prog)
}

// TestFind_Pin keeps the symbol referenced. Find requires a live TTY
// (fuzzyfinder/tcell); exercising it headlessly is not supported.
func TestFind_Pin(t *testing.T) {
	t.Parallel()
	_ = Find[string] // symbol pin — TTY required to drive fuzzyfinder
}

func TestDrainTerminalResponses_NoPanic(t *testing.T) {
	t.Parallel()
	// On Windows this is a no-op; on Unix it drains terminal escape responses.
	DrainTerminalResponses(0)
}
