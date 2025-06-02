package test_util

import (
	"reflect"
	"testing"
)

// Define the Episode struct
type EPisode struct {
	Number string
	Num    int
	URL    string
}

// Test function for sortEpisodesByNum
func TestSortEpisodesByNum(t *testing.T) {
	tests := []struct {
		input    []Episode
		expected []Episode
	}{
		{
			input: []Episode{
				{Number: "Episode 3", Num: 3, URL: "/episode3"},
				{Number: "Episode 1", Num: 1, URL: "/episode1"},
				{Number: "Episode 2", Num: 2, URL: "/episode2"},
			},
			expected: []Episode{
				{Number: "Episode 1", Num: 1, URL: "/episode1"},
				{Number: "Episode 2", Num: 2, URL: "/episode2"},
				{Number: "Episode 3", Num: 3, URL: "/episode3"},
			},
		},
		{
			input: []Episode{
				{Number: "Episode 5", Num: 5, URL: "/episode5"},
				{Number: "Episode 4", Num: 4, URL: "/episode4"},
				{Number: "Episode 6", Num: 6, URL: "/episode6"},
			},
			expected: []Episode{
				{Number: "Episode 4", Num: 4, URL: "/episode4"},
				{Number: "Episode 5", Num: 5, URL: "/episode5"},
				{Number: "Episode 6", Num: 6, URL: "/episode6"},
			},
		},
		{
			input: []Episode{
				{Number: "Episode 2", Num: 2, URL: "/episode2"},
				{Number: "Episode 1", Num: 1, URL: "/episode1"},
			},
			expected: []Episode{
				{Number: "Episode 1", Num: 1, URL: "/episode1"},
				{Number: "Episode 2", Num: 2, URL: "/episode2"},
			},
		},
	}

	for _, test := range tests {
		sortEpisodesByNum(test.input)
		if !reflect.DeepEqual(test.input, test.expected) {
			t.Errorf("sortEpisodesByNum(%v) = %v, expected %v", test.input, test.input, test.expected)
		}
	}
}
