package tui

import (
	"testing"

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
	assert.True(t, dark.Panel.GetBorderLeft())
	assert.True(t, dark.Panel.GetBorderRight())
	assert.True(t, dark.Panel.GetBorderTop())
	assert.True(t, dark.Panel.GetBorderBottom())
	assert.Equal(t, 2, dark.Panel.GetPaddingLeft())
	assert.Equal(t, 1, dark.Panel.GetPaddingTop())
	assert.True(t, dark.SelectedTitle.GetBorderLeft())
	assert.False(t, dark.SelectedTitle.GetBorderRight())
	assert.True(t, dark.FilterMatch.GetUnderline())
	assert.True(t, dark.Value.GetBold())
	assert.Contains(t, dark.Header.Render("GoAnime"), "GoAnime")
}
