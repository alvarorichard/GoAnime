package test_util

import (
	"testing"
)

func TestParseEpisodeNumber(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		err      bool
	}{
		{"Episode 1", 1, false},
		{"Episode 2", 2, false},
		{"Episode 10", 10, false},
		{"No Episode Number", 0, true}, // Should return error because it doesn't find number
		{"Special 45 Episode", 45, false},
		{"", 0, true}, // Should return error because it doesn't find number
		{"123", 123, false},
		{"Episode-15", 15, false},
		{"15th Episode", 15, false},
	}

	for _, test := range tests {
		result, err := parseEpisodeNumber(test.input)
		if (err != nil) != test.err {
			t.Errorf("parseEpisodeNumber(%q) returned error %v, expected error: %v", test.input, err, test.err)
		}
		if result != test.expected {
			t.Errorf("parseEpisodeNumber(%q) = %d, expected %d", test.input, result, test.expected)
		}
	}
}
