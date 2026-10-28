package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormTheme(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)
	styles := FormTheme().Theme(true)

	assert.Equal(t, theme.Primary.GetForeground(), styles.Focused.Title.GetForeground())
	assert.Equal(t, theme.Shadow.GetForeground(), styles.Focused.Base.GetBorderLeftForeground(), "the focused field's bar is crimson")
	assert.Equal(t, theme.Selected.GetBackground(), styles.Focused.FocusedButton.GetBackground())
	assert.Equal(t, theme.Keycap.GetBackground(), styles.Focused.BlurredButton.GetBackground())
	assert.Equal(t, styles.Focused.Title, styles.Group.Title)

	light := FormTheme().Theme(false)
	assert.Equal(t, styles.Focused.Title.GetForeground(), light.Focused.Title.GetForeground(),
		"always dark, like the list screens: huh guesses light when the query is suppressed")
}

func TestSpinnerTheme(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)
	styles := SpinnerTheme().Theme(true)
	assert.Equal(t, theme.Primary.GetForeground(), styles.Spinner.GetForeground())
	assert.Equal(t, theme.Text.GetForeground(), styles.Title.GetForeground())
	assert.Equal(t, styles.Spinner.GetForeground(), SpinnerTheme().Theme(false).Spinner.GetForeground())
}
