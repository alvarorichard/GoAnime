package superflix

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

// These tests mutate process-wide environment via t.Setenv, which is
// incompatible with t.Parallel — so they intentionally run serially.

func TestLoadSuperflixConfig(t *testing.T) {
	tests := []struct {
		name     string
		env      map[string]string
		expected superflixConfig
	}{
		{
			name:     "all unset",
			env:      map[string]string{},
			expected: superflixConfig{},
		},
		{
			name: "all set",
			env: map[string]string{
				"GOANIME_SF_HEADLESS":       "1",
				"GOANIME_SF_BUNDLED":        "1",
				"GOANIME_SF_CHROME_CHANNEL": "msedge",
				"GOANIME_SF_MASK":           "1",
			},
			expected: superflixConfig{
				Headless:     true,
				ForceBundled: true,
				Channel:      "msedge",
				Mask:         true,
			},
		},
		{
			name: "channel only",
			env:  map[string]string{"GOANIME_SF_CHROME_CHANNEL": "chrome-beta"},
			expected: superflixConfig{
				Channel: "chrome-beta",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear all four knobs first, then apply this case's overrides, so a
			// value leaking from the real environment can't skew the result.
			for _, k := range []string{
				"GOANIME_SF_HEADLESS", "GOANIME_SF_BUNDLED",
				"GOANIME_SF_CHROME_CHANNEL", "GOANIME_SF_MASK",
			} {
				t.Setenv(k, "")
			}
			for k, v := range tt.env {
				t.Setenv(k, v)
			}
			assert.Equal(t, tt.expected, loadSuperflixConfig())
		})
	}
}

func TestSuperflixConfig_ResolveChannel(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		cfg  superflixConfig
		want string
	}{
		{"default auto-detects chrome", superflixConfig{}, "chrome"},
		{"explicit channel wins", superflixConfig{Channel: "msedge"}, "msedge"},
		{"explicit channel wins over bundled", superflixConfig{Channel: "msedge", ForceBundled: true}, "msedge"},
		{"forced bundled yields empty", superflixConfig{ForceBundled: true}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.cfg.resolveChannel())
		})
	}
}

func TestDisplayAvailable(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		assert.True(t, displayAvailable(), "windows/darwin always report a display")
		return
	}
	tests := []struct {
		name    string
		display string
		wayland string
		want    bool
	}{
		{"no display vars", "", "", false},
		{"x11 only", ":0", "", true},
		{"wayland only", "", "wayland-0", true},
		{"both", ":0", "wayland-0", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DISPLAY", tt.display)
			t.Setenv("WAYLAND_DISPLAY", tt.wayland)
			assert.Equal(t, tt.want, displayAvailable())
		})
	}
}

func TestHeadlessEnvironment(t *testing.T) {
	if runtime.GOOS == "windows" || runtime.GOOS == "darwin" {
		t.Skip("display is always present on windows/darwin")
	}
	tests := []struct {
		name     string
		display  string
		wayland  string
		headless string
		want     bool
	}{
		{"no display, not opted headless -> warn", "", "", "", true},
		{"has display -> no warn", ":0", "", "", false},
		{"no display but opted headless -> no warn", "", "", "1", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("DISPLAY", tt.display)
			t.Setenv("WAYLAND_DISPLAY", tt.wayland)
			t.Setenv("GOANIME_SF_HEADLESS", tt.headless)
			assert.Equal(t, tt.want, HeadlessEnvironment())
		})
	}
}
