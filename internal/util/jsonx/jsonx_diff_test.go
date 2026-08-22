package jsonx

import (
	jsonv1 "encoding/json"
	"reflect"
	"strings"
	"testing"
)

// The structs below mirror the shapes GoAnime actually decodes: AniList's
// deeply nested GraphQL envelope, TVmaze's flat episode list, and the
// provider result objects that mix numbers, nulls and free-form maps.

type aniListEnvelope struct {
	Data struct {
		Media struct {
			ID    int `json:"id"`
			Title struct {
				Romaji  string `json:"romaji"`
				English string `json:"english"`
				Native  string `json:"native"`
			} `json:"title"`
			Episodes     int      `json:"episodes"`
			Genres       []string `json:"genres"`
			AverageScore int      `json:"averageScore"`
			Description  string   `json:"description"`
			CoverImage   struct {
				Large string `json:"large"`
			} `json:"coverImage"`
			Relations struct {
				Edges []struct {
					RelationType string `json:"relationType"`
					Node         struct {
						ID    int `json:"id"`
						Title struct {
							Romaji string `json:"romaji"`
						} `json:"title"`
					} `json:"node"`
				} `json:"edges"`
			} `json:"relations"`
		} `json:"Media"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
		Status  int    `json:"status"`
	} `json:"errors"`
}

type tvmazeEpisode struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Season  int    `json:"season"`
	Number  *int   `json:"number"`
	Airdate string `json:"airdate"`
	Runtime *int   `json:"runtime"`
	Rating  struct {
		Average *float64 `json:"average"`
	} `json:"rating"`
	Image map[string]string `json:"image"`
}

type providerResult struct {
	Status  bool           `json:"status"`
	Message string         `json:"message"`
	Data    map[string]any `json:"data"`
	List    []any          `json:"list"`
	Count   int64          `json:"count"`
	Ratio   float64        `json:"ratio"`
}

// diffCorpus holds payloads exercising the v1/v2 behaviour differences that
// matter for provider responses: casing, nulls, duplicates, empty containers,
// numeric shapes and non-ASCII text.
var diffCorpus = []string{
	`{}`,
	`null`,
	`{"data":{"Media":{"id":1,"title":{"romaji":"Naruto","english":null,"native":"ナルト"},"episodes":220,"genres":["Action","Adventure"],"averageScore":79,"description":"<p>Ninja</p>","coverImage":{"large":"https://x/y.jpg"},"relations":{"edges":[{"relationType":"SEQUEL","node":{"id":2,"title":{"romaji":"Shippuuden"}}}]}}}}`,
	`{"DATA":{"media":{"ID":1,"TITLE":{"ROMAJI":"Naruto"}}}}`,
	`{"data":{"Media":{"id":1,"genres":null,"relations":{"edges":null}}}}`,
	`{"data":{"Media":{"id":1,"genres":[]}},"errors":[]}`,
	`{"errors":[{"message":"Not Found","status":404}]}`,
	`{"data":{"Media":{"id":1,"id":2}}}`,
	`{"data":{"Media":{"episodes":0,"averageScore":-1}}}`,
	`{"unknownTopLevel":123,"data":{"Media":{"id":5}}}`,
}

var episodeCorpus = []string{
	`[]`,
	`null`,
	`[{"id":1,"name":"Ep 1","season":1,"number":1,"airdate":"2024-01-01","runtime":24,"rating":{"average":8.1},"image":{"medium":"m.jpg","original":"o.jpg"}}]`,
	`[{"id":1,"name":"Ep 1","number":null,"runtime":null,"rating":{"average":null},"image":null}]`,
	`[{"ID":1,"NAME":"Ep 1","Season":2,"NUMBER":3}]`,
	`[{"id":1},{"id":2},{"id":3}]`,
	`[{"id":1,"name":"蟲師 \u00e9pisode"}]`,
	`[{"id":1,"extra":"ignored","rating":{"average":0}}]`,
}

var resultCorpus = []string{
	`{"status":true,"message":"ok","data":{"a":1,"b":"x","c":null,"d":[1,2],"e":{"f":true}},"list":[1,"two",null,{"k":"v"}],"count":9007199254740993,"ratio":0.1}`,
	`{"status":false,"data":{},"list":[]}`,
	`{"status":true,"data":null,"list":null}`,
	`{"count":-5,"ratio":-0.0}`,
	`{"ratio":1e10,"count":0}`,
	`{"data":{"dup":1,"dup":2}}`,
	`{"message":"caf\u00e9 \u30a2\u30cb\u30e1"}`,
}

// runDiff decodes every payload with both implementations and requires the
// results to be indistinguishable: same success/failure, same decoded value.
func runDiff[T any](t *testing.T, corpus []string) {
	t.Helper()
	for _, payload := range corpus {
		var viaV1, viaJSONX T
		errV1 := jsonv1.Unmarshal([]byte(payload), &viaV1)
		errX := Unmarshal([]byte(payload), &viaJSONX)

		switch {
		case errV1 == nil && errX != nil:
			t.Errorf("payload %s: v1 accepted but jsonx rejected: %v", truncate(payload), errX)
			continue
		case errV1 != nil && errX == nil:
			t.Errorf("payload %s: v1 rejected (%v) but jsonx accepted", truncate(payload), errV1)
			continue
		case errV1 != nil:
			continue // both rejected: error text is allowed to differ
		}
		if !reflect.DeepEqual(viaV1, viaJSONX) {
			t.Errorf("payload %s:\n  v1    = %+v\n  jsonx = %+v", truncate(payload), viaV1, viaJSONX)
		}
	}
}

// runDiffStreaming does the same through the streaming Decode path, which must
// agree with the buffered one byte for byte.
func runDiffStreaming[T any](t *testing.T, corpus []string) {
	t.Helper()
	for _, payload := range corpus {
		var viaV1, viaJSONX T
		errV1 := jsonv1.Unmarshal([]byte(payload), &viaV1)
		errX := Decode(strings.NewReader(payload), 1<<20, &viaJSONX)
		if (errV1 == nil) != (errX == nil) {
			t.Errorf("payload %s: v1 err=%v, Decode err=%v", truncate(payload), errV1, errX)
			continue
		}
		if errV1 != nil {
			continue
		}
		if !reflect.DeepEqual(viaV1, viaJSONX) {
			t.Errorf("payload %s (streaming):\n  v1    = %+v\n  jsonx = %+v", truncate(payload), viaV1, viaJSONX)
		}
	}
}

func TestDiffAniListEnvelope(t *testing.T)  { runDiff[aniListEnvelope](t, diffCorpus) }
func TestDiffTVMazeEpisodes(t *testing.T)   { runDiff[[]tvmazeEpisode](t, episodeCorpus) }
func TestDiffProviderResult(t *testing.T)   { runDiff[providerResult](t, resultCorpus) }
func TestDiffGenericMap(t *testing.T)       { runDiff[map[string]any](t, resultCorpus) }
func TestDiffAniListStreaming(t *testing.T) { runDiffStreaming[aniListEnvelope](t, diffCorpus) }
func TestDiffEpisodesStreaming(t *testing.T) {
	runDiffStreaming[[]tvmazeEpisode](t, episodeCorpus)
}
func TestDiffGenericMapStreaming(t *testing.T) { runDiffStreaming[map[string]any](t, resultCorpus) }

// TestDecodeAgreesWithUnmarshal guards the invariant the migration relies on:
// call sites that keep the bytes (Unmarshal) and call sites that stream
// (Decode) must produce the same value for the same payload.
func TestDecodeAgreesWithUnmarshal(t *testing.T) {
	for _, payload := range append(append(append([]string{}, diffCorpus...), episodeCorpus...), resultCorpus...) {
		var buffered, streamed map[string]any
		errB := Unmarshal([]byte(payload), &buffered)
		errS := Decode(strings.NewReader(payload), 1<<20, &streamed)
		if (errB == nil) != (errS == nil) {
			t.Errorf("payload %s: Unmarshal err=%v, Decode err=%v", truncate(payload), errB, errS)
			continue
		}
		if errB == nil && !reflect.DeepEqual(buffered, streamed) {
			t.Errorf("payload %s: buffered %+v != streamed %+v", truncate(payload), buffered, streamed)
		}
	}
}

func truncate(s string) string {
	if len(s) <= 90 {
		return s
	}
	return s[:90] + "..."
}

// FuzzUnmarshalMatchesV1 explores beyond the hand-written corpus: any input v1
// accepts must also be accepted by jsonx, with an identical decoded value.
func FuzzUnmarshalMatchesV1(f *testing.F) {
	for _, s := range diffCorpus {
		f.Add(s)
	}
	for _, s := range resultCorpus {
		f.Add(s)
	}
	for _, s := range episodeCorpus {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, payload string) {
		var viaV1, viaJSONX any
		errV1 := jsonv1.Unmarshal([]byte(payload), &viaV1)
		errX := Unmarshal([]byte(payload), &viaJSONX)
		if errV1 != nil {
			return // v1 rejects it: jsonx is free to reject it too
		}
		if errX != nil {
			t.Fatalf("v1 accepted %q but jsonx rejected it: %v", payload, errX)
		}
		if !reflect.DeepEqual(viaV1, viaJSONX) {
			t.Fatalf("payload %q: v1 = %#v, jsonx = %#v", payload, viaV1, viaJSONX)
		}
	})
}

// FuzzUnmarshalStructMatchesV1 is the typed counterpart of
// FuzzUnmarshalMatchesV1: decoding into a concrete struct exercises the field
// matching rules (case folding, delimiters, unknown members) that decoding into
// `any` never touches.
func FuzzUnmarshalStructMatchesV1(f *testing.F) {
	for _, s := range episodeCorpus {
		f.Add(s)
	}
	for _, s := range diffCorpus {
		f.Add(s)
	}
	f.Fuzz(func(t *testing.T, payload string) {
		var viaV1, viaJSONX []tvmazeEpisode
		if err := jsonv1.Unmarshal([]byte(payload), &viaV1); err != nil {
			return
		}
		if err := Unmarshal([]byte(payload), &viaJSONX); err != nil {
			t.Fatalf("v1 accepted %q but jsonx rejected it: %v", payload, err)
		}
		if !reflect.DeepEqual(viaV1, viaJSONX) {
			t.Fatalf("payload %q: v1 = %#v, jsonx = %#v", payload, viaV1, viaJSONX)
		}
	})
}
