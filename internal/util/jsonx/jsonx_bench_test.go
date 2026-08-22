package jsonx

import (
	"bytes"
	jsonv1 "encoding/json"
	"fmt"
	"io"
	"strings"
	"testing"
)

type benchEpisode struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Season  int     `json:"season"`
	Number  int     `json:"number"`
	Airdate string  `json:"airdate"`
	Runtime int     `json:"runtime"`
	Rating  float64 `json:"rating"`
	Summary string  `json:"summary"`
	URL     string  `json:"url"`
}

// benchPayload mimics the shape and size of a real episode-list response.
func benchPayload(n int) []byte {
	var b strings.Builder
	b.WriteByte('[')
	for i := range n {
		if i > 0 {
			b.WriteByte(',')
		}
		fmt.Fprintf(&b, `{"id":%d,"name":"Episode %d","season":1,"number":%d,"airdate":"2024-01-%02d",`+
			`"runtime":24,"rating":8.5,"summary":"<p>A summary long enough to look like the real thing.</p>",`+
			`"url":"https://example.com/episodes/%d"}`, i, i, i, (i%28)+1, i)
	}
	b.WriteByte(']')
	return []byte(b.String())
}

var benchData = benchPayload(400)

// BenchmarkJSONXVsV1 is the measurement behind the migration: the buffered path
// against encoding/json, and the streaming path against the
// io.ReadAll(io.LimitReader(...)) + json.Unmarshal pattern it replaces.
func BenchmarkJSONXVsV1(b *testing.B) {
	b.Run("v1/Unmarshal", func(b *testing.B) {
		b.SetBytes(int64(len(benchData)))
		for b.Loop() {
			var out []benchEpisode
			if err := jsonv1.Unmarshal(benchData, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("jsonx/Unmarshal", func(b *testing.B) {
		b.SetBytes(int64(len(benchData)))
		for b.Loop() {
			var out []benchEpisode
			if err := Unmarshal(benchData, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("v1/ReadAll+Unmarshal", func(b *testing.B) {
		b.SetBytes(int64(len(benchData)))
		for b.Loop() {
			buf, err := io.ReadAll(io.LimitReader(bytes.NewReader(benchData), 5<<20))
			if err != nil {
				b.Fatal(err)
			}
			var out []benchEpisode
			if err := jsonv1.Unmarshal(buf, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
	b.Run("jsonx/Decode", func(b *testing.B) {
		b.SetBytes(int64(len(benchData)))
		for b.Loop() {
			var out []benchEpisode
			if err := Decode(bytes.NewReader(benchData), 5<<20, &out); err != nil {
				b.Fatal(err)
			}
		}
	})
}
