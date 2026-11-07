package providers

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/stretchr/testify/assert"
)

func TestStreamFromMetadata(t *testing.T) {
	cases := map[string]struct {
		metadata map[string]string
		want     source.Stream
	}{
		"nothing":                  {nil, source.Stream{URL: "u"}},
		"empty map":                {map[string]string{}, source.Stream{URL: "u"}},
		"referer only":             {map[string]string{"referer": "https://r/"}, source.Stream{URL: "u", Referer: "https://r/"}},
		"origin only":              {map[string]string{"origin": "https://o"}, source.Stream{URL: "u", Origin: "https://o"}},
		"values are trimmed":       {map[string]string{"referer": "  https://r/\n", "origin": "\thttps://o  "}, source.Stream{URL: "u", Referer: "https://r/", Origin: "https://o"}},
		"blank values are absent":  {map[string]string{"referer": "   ", "origin": "\n"}, source.Stream{URL: "u"}},
		"unknown keys are ignored": {map[string]string{"x-other": "v"}, source.Stream{URL: "u"}},
		"subtitles": {
			map[string]string{"subtitles": `[{"url":"a","language":"pt","label":"P"},{"url":"b","language":"en","label":"E"}]`},
			source.Stream{URL: "u", Subtitles: []source.Subtitle{{URL: "a", Language: "pt", Label: "P"}, {URL: "b", Language: "en", Label: "E"}}},
		},
		"subtitles with surrounding space": {
			map[string]string{"subtitles": "  [{\"url\":\"a\"}]\n"},
			source.Stream{URL: "u", Subtitles: []source.Subtitle{{URL: "a"}}},
		},
		"an empty list": {map[string]string{"subtitles": `[]`}, source.Stream{URL: "u", Subtitles: []source.Subtitle{}}},
		"a blank list":  {map[string]string{"subtitles": "  "}, source.Stream{URL: "u"}},
		"malformed subtitles are dropped, the rest kept": {
			map[string]string{"referer": "https://r/", "subtitles": `{not json`},
			source.Stream{URL: "u", Referer: "https://r/"},
		},
		"a subtitle list of the wrong shape is dropped": {
			map[string]string{"referer": "https://r/", "subtitles": `{"url":"a"}`},
			source.Stream{URL: "u", Referer: "https://r/"},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			got := streamFromMetadata("u", tc.metadata)
			if len(tc.want.Subtitles) == 0 && len(got.Subtitles) == 0 {
				tc.want.Subtitles, got.Subtitles = nil, nil // an empty list and no list are the same to the player
			}
			assert.Equal(t, tc.want, got)
		})
	}
}
