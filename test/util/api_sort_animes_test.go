package util_test

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/stretchr/testify/assert"
)

// sortAnimes is a helper function to sort animes

func TestSortAnimes(t *testing.T) {
	tests := []struct {
		name     string
		input    []api.Anime
		expected []api.Anime
	}{
		{
			name: "Basic sorting",
			input: []api.Anime{
				{Name: "Naruto"},
				{Name: "Bleach"},
				{Name: "One Piece"},
				{Name: "Dragon Ball"},
			},
			expected: []api.Anime{
				{Name: "Bleach"},
				{Name: "Dragon Ball"},
				{Name: "Naruto"},
				{Name: "One Piece"},
			},
		},
		{
			name: "Already sorted",
			input: []api.Anime{
				{Name: "Bleach"},
				{Name: "Dragon Ball"},
				{Name: "Naruto"},
				{Name: "One Piece"},
			},
			expected: []api.Anime{
				{Name: "Bleach"},
				{Name: "Dragon Ball"},
				{Name: "Naruto"},
				{Name: "One Piece"},
			},
		},
		{
			name: "Reverse order",
			input: []api.Anime{
				{Name: "One Piece"},
				{Name: "Naruto"},
				{Name: "Dragon Ball"},
				{Name: "Bleach"},
			},
			expected: []api.Anime{
				{Name: "Bleach"},
				{Name: "Dragon Ball"},
				{Name: "Naruto"},
				{Name: "One Piece"},
			},
		},
		{
			name: "Single element",
			input: []api.Anime{
				{Name: "Naruto"},
			},
			expected: []api.Anime{
				{Name: "Naruto"},
			},
		},
		{
			name:     "Empty list",
			input:    []api.Anime{},
			expected: []api.Anime{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sortAnimes(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
