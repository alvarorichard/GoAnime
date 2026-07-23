package upscaler

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetShaderDir(t *testing.T) {
	t.Parallel()
	got := GetShaderDir()
	assert.NotEmpty(t, got)
}

func TestShadersInstalled_DoesNotPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() { _ = ShadersInstalled() })
}

func TestGANShadersInstalled_DoesNotPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, func() { _ = GANShadersInstalled() })
}

func TestGetAllShaderModes(t *testing.T) {
	t.Parallel()
	modes := GetAllShaderModes()
	require.NotEmpty(t, modes)
}

func TestGetAdvancedShaderModes(t *testing.T) {
	t.Parallel()
	modes := GetAdvancedShaderModes()
	require.NotEmpty(t, modes)
}

func TestGetShaderModeName(t *testing.T) {
	t.Parallel()
	for _, m := range GetAllShaderModes() {
		name := GetShaderModeName(m)
		assert.NotEmpty(t, name, m)
	}
}

func TestGetShaderModeDescription(t *testing.T) {
	t.Parallel()
	for _, m := range GetAllShaderModes() {
		desc := GetShaderModeDescription(m)
		assert.NotEmpty(t, desc, m)
	}
}

func TestGetMPVShaderArgs_AllModes(t *testing.T) {
	t.Parallel()
	for _, m := range GetAllShaderModes() {
		args := GetMPVShaderArgs(m)
		_ = args // args may be empty if shaders not installed; just exercise call
	}
}

func TestSetAndCycleShaderMode(t *testing.T) {
	all := GetAllShaderModes()
	if len(all) < 2 {
		t.Skip("need >=2 modes")
	}
	SetShaderMode(all[0])
	next := CycleShaderMode()
	assert.NotEqual(t, all[0], next)
}

func TestShaderModeConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	modeCount := len(GetAllShaderModes())
	for i := range 8 {
		wg.Go(func() {
			for j := range 1000 {
				SetShaderMode(ShaderMode((i + j) % modeCount))
				_ = GetShaderMode()
				_ = CycleShaderMode()
			}
		})
	}
	wg.Wait()
}
