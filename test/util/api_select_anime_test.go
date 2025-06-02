package test_util_test

import (
	"errors"
	"sort"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// FuzzyFinder is an interface to abstract the fuzzyfinder interaction
type FuzzyFinder interface {
	Find(slice interface{}, itemFunc func(i int) string) (int, error)
}

// MockFuzzyFinder is a mock implementation of the FuzzyFinder interface
type MockFuzzyFinder struct {
	mock.Mock
}

func (m *MockFuzzyFinder) Find(slice interface{}, itemFunc func(i int) string) (int, error) {
	args := m.Called(slice, itemFunc)
	return args.Int(0), args.Error(1)
}

// sortAnimes is a helper function to sort animes
func sortAnimes(animeList []models.Anime) []models.Anime {
	sort.Slice(animeList, func(i, j int) bool {
		return animeList[i].Name < animeList[j].Name
	})
	return animeList
}

// selectAnimeWithGoFuzzyFinder is a modified version of the original function for testing
func selectAnimeWithGoFuzzyFinder(finder FuzzyFinder, animes []models.Anime) (string, error) {
	if len(animes) == 0 {
		return "", errors.New("no anime provided")
	}

	animeNames := make([]string, len(animes))
	sortedAnimes := sortAnimes(animes)
	for i, anime := range sortedAnimes {
		animeNames[i] = anime.Name
	}

	idx, err := finder.Find(
		animeNames,
		func(i int) string {
			return animeNames[i]
		},
	)
	if err != nil {
		return "", errors.New("failed to select anime with go-fuzzyfinder: " + err.Error())
	}

	if idx < 0 || idx >= len(animes) {
		return "", errors.New("invalid index returned by fuzzyfinder")
	}

	return sortedAnimes[idx].Name, nil
}

func TestSelectAnimeWithGoFuzzyFinder(t *testing.T) {
	tests := []struct {
		name        string
		animes      []models.Anime
		findResult  int
		findError   error
		expected    string
		expectedErr bool
	}{
		{
			name: "Single anime selection",
			animes: []models.Anime{
				{Name: "Anime One"},
				{Name: "Anime Three"},
				{Name: "Anime Two"},
			},
			findResult:  1,
			findError:   nil,
			expected:    "Anime Three",
			expectedErr: false,
		},
		{
			name:        "No anime provided",
			animes:      []models.Anime{},
			findResult:  0,
			findError:   nil,
			expected:    "",
			expectedErr: true,
		},
		{
			name: "Find function error",
			animes: []models.Anime{
				{Name: "Anime One"},
				{Name: "Anime Three"},
				{Name: "Anime Two"},
			},
			findResult:  0,
			findError:   errors.New("fuzzyfinder error"),
			expected:    "",
			expectedErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFinder := new(MockFuzzyFinder)

			if len(tt.animes) > 0 {
				mockFinder.On("Find", mock.Anything, mock.Anything).Return(tt.findResult, tt.findError)
			}

			result, err := selectAnimeWithGoFuzzyFinder(mockFinder, tt.animes)

			if tt.expectedErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.expected, result)
			mockFinder.AssertExpectations(t)
		})
	}
}
