package util

import (
	"fmt"
	"testing"
)

func TestSanitizeForFilename(t *testing.T) {
	tests := []struct {
		in, want string
	}{
		{"Black Clover (Dublado) 7.27 A14", "Black Clover (Dublado)"},
		{"Naruto Shippuuden 8.50 A12", "Naruto Shippuuden"},
		{"One Piece 9.12 L", "One Piece"},
		{"Demon Slayer", "Demon Slayer"},
		{"Attack on Titan (Legendado) 9.00 AL", "Attack on Titan (Legendado)"},
		{"Jujutsu Kaisen 2nd Season 8.60 A14", "Jujutsu Kaisen 2nd Season"},
		{"Solo Leveling 8.21 A14", "Solo Leveling"},
		{"My Hero Academia 7.50 L", "My Hero Academia"},
		{"Bleach: Thousand-Year Blood War", "Bleach Thousand-Year Blood War"},
		{"[Movies/TV] Dexter", "Dexter"},
		{"[English] Naruto", "Naruto"},
		{"[Portuguese] One Piece", "One Piece"},
	}
	for _, tc := range tests {
		t.Run(tc.in, func(t *testing.T) {
			got := SanitizeForFilename(tc.in)
			if got != tc.want {
				t.Errorf("SanitizeForFilename(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestFormatPlexEpisodePath(t *testing.T) {
	path := FormatPlexEpisodePath("/media/anime", "Black Clover (Dublado) 7.27 A14", 1, 3)
	want := "/media/anime/Black Clover (Dublado)/Season 01/Black Clover (Dublado) - s01e03.mp4"
	if path != want {
		t.Errorf("FormatPlexEpisodePath = %q, want %q", path, want)
	}
}

func TestPlexEpisodeFilename(t *testing.T) {
	fn := PlexEpisodeFilename("Naruto Shippuuden 8.50 A12", 2, 15)
	want := "Naruto Shippuuden - s02e15.mp4"
	if fn != want {
		t.Errorf("PlexEpisodeFilename = %q, want %q", fn, want)
	}
}

func ExampleSanitizeForFilename() {
	fmt.Println(SanitizeForFilename("Black Clover (Dublado) 7.27 A14"))
	// Output: Black Clover (Dublado)
}
