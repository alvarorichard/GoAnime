package tui

import (
	"image/color"
	"math"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTheme(t *testing.T) {
	t.Parallel()

	dark := NewTheme(true)
	light := NewTheme(false)

	require.NotNil(t, dark.Primary.GetForeground())
	require.NotNil(t, light.Primary.GetForeground())
	assert.NotEqual(t, dark.Primary.GetForeground(), light.Primary.GetForeground())
	assert.True(t, dark.Header.GetBold())
	assert.Contains(t, dark.Header.Render("GOANIME"), "GOANIME")
	assert.True(t, dark.Selected.GetBold())
	assert.True(t, dark.FilterMatch.GetUnderline())
	assert.True(t, dark.PanelValue.GetBold())

	// Square corners, as on the site: a panel is a filled block, never a
	// rounded border.
	assert.False(t, dark.Panel.GetBorderTop())
	assert.Equal(t, 2, dark.Panel.GetPaddingLeft())
	assert.Equal(t, 1, dark.Panel.GetPaddingTop())

	// Every style drawn inside a panel carries the panel's fill, or the
	// ground shows through between styled runs.
	for name, style := range map[string]lipgloss.Style{
		"PanelLabel": dark.PanelLabel, "PanelText": dark.PanelText, "PanelMuted": dark.PanelMuted,
		"PanelValue": dark.PanelValue, "PanelTone": dark.PanelTone, "PanelBar": dark.PanelBar, "PanelTrack": dark.PanelTrack,
	} {
		assert.Equal(t, dark.Panel.GetBackground(), style.GetBackground(), name)
	}
	// The same for the crimson cursor row.
	for name, style := range map[string]lipgloss.Style{
		"SelectedMuted": dark.SelectedMuted, "SelectedFaint": dark.SelectedFaint,
	} {
		assert.Equal(t, dark.Selected.GetBackground(), style.GetBackground(), name)
	}
}

// The theme promises WCAG AA (4.5:1) for every pair that carries text.
func TestThemeContrast(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		isDark bool
		ground color.Color
	}{
		{"dark", true, lipgloss.Color("#0B0B0E")},
		{"light", false, lipgloss.Color("#FFFFFF")},
	} {
		theme := NewTheme(tc.isDark)
		onGround := map[string]lipgloss.Style{
			"Text": theme.Text, "Soft": theme.Soft, "Muted": theme.Muted, "Primary": theme.Primary, "Footer": theme.Footer,
		}
		for name, style := range onGround {
			assert.GreaterOrEqual(t, contrast(style.GetForeground(), tc.ground), 4.5, "%s: %s on the ground", tc.name, name)
		}
		filled := map[string]lipgloss.Style{
			"Header": theme.Header, "Selected": theme.Selected, "SelectedMuted": theme.SelectedMuted,
			"Keycap": theme.Keycap, "Tag": theme.Tag, "Notice": theme.Notice, "PanelCode": theme.PanelCode,
			"PanelLabel": theme.PanelLabel, "PanelText": theme.PanelText, "PanelMuted": theme.PanelMuted, "PanelValue": theme.PanelValue,
		}
		for name, style := range filled {
			assert.GreaterOrEqual(t, contrast(style.GetForeground(), style.GetBackground()), 4.5, "%s: %s", tc.name, name)
		}
	}
}

// contrast is the WCAG 2 contrast ratio between two colors.
func contrast(a, b color.Color) float64 {
	la, lb := luminance(a), luminance(b)
	return (math.Max(la, lb) + 0.05) / (math.Min(la, lb) + 0.05)
}

func luminance(c color.Color) float64 {
	r, g, b, _ := c.RGBA()
	channel := func(v uint32) float64 {
		s := float64(v) / 0xffff
		if s <= 0.03928 {
			return s / 12.92
		}
		return math.Pow((s+0.055)/1.055, 2.4)
	}
	return 0.2126*channel(r) + 0.7152*channel(g) + 0.0722*channel(b)
}
