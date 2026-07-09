package util

import "testing"

func TestSourceDisabled(t *testing.T) {
	// Uses t.Setenv — not parallel.
	tests := []struct {
		name  string
		env   string
		probe string
		want  bool
	}{
		{"unset", "", "AllAnime", false},
		{"exact match", "AllAnime", "AllAnime", true},
		{"case-insensitive", "allanime", "AllAnime", true},
		{"dot-forgiving list side", "Animefire.io", "AnimeFire", true},
		{"dot-forgiving probe side", "animefire", "Animefire.io", true},
		{"one of several", "Goyabu, SuperFlix ,AllAnime", "SuperFlix", true},
		{"whitespace tolerant", "  AllAnime  ", "AllAnime", true},
		{"not listed", "Goyabu", "AllAnime", false},
		{"empty probe", "AllAnime", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv(disabledSourcesEnv, tt.env)
			if got := SourceDisabled(tt.probe); got != tt.want {
				t.Errorf("SourceDisabled(%q) with env %q = %v, want %v", tt.probe, tt.env, got, tt.want)
			}
		})
	}
}

func TestSourceForceEnabled(t *testing.T) {
	// Uses t.Setenv — not parallel.
	t.Run("unset is false", func(t *testing.T) {
		t.Setenv(enabledSourcesEnv, "")
		if SourceForceEnabled("Experimental") {
			t.Error("no env should mean not force-enabled")
		}
	})
	t.Run("listed is true", func(t *testing.T) {
		t.Setenv(enabledSourcesEnv, "Experimental,Beta")
		if !SourceForceEnabled("beta") {
			t.Error("listed source should be force-enabled (case-insensitive)")
		}
	})
	t.Run("disabled and enabled envs are independent", func(t *testing.T) {
		t.Setenv(disabledSourcesEnv, "AllAnime")
		t.Setenv(enabledSourcesEnv, "Beta")
		if SourceForceEnabled("AllAnime") {
			t.Error("SourceForceEnabled must read ENABLED env, not DISABLED")
		}
		if !SourceDisabled("AllAnime") {
			t.Error("SourceDisabled must read DISABLED env")
		}
	})
}
