package test_util

import (
	"github.com/pkg/errors"
	"regexp"
	"strconv"
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
		{"No Episode Number", 0, true}, // Deve retornar erro porque não encontra número
		{"Special 45 Episode", 45, false},
		{"", 0, true}, // Deve retornar erro porque não encontra número
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

// Função a ser testada (copiada do código original para garantir que o teste funciona de forma independente)
func parseEpisodenumber(episodeNum string) (int, error) {
	numRe := regexp.MustCompile(`\d+`)
	numStr := numRe.FindString(episodeNum)
	if numStr == "" {
		return 0, errors.New("no episode number found")
	}
	return strconv.Atoi(numStr)
}
