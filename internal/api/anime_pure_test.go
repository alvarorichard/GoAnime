package api

import (
	"errors"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetStringValue(t *testing.T) {
	t.Parallel()
	data := map[string]any{
		"str":   "hello",
		"empty": "",
		"num":   42.0,
	}
	tests := []struct {
		name  string
		field string
		want  string
	}{
		{"present string", "str", "hello"},
		{"empty string", "empty", ""},
		{"missing key", "missing", ""},
		{"non-string", "num", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, getStringValue(data, tt.field))
		})
	}
}

func TestGetIntValue(t *testing.T) {
	t.Parallel()
	data := map[string]any{
		"float":  42.7,
		"int":    int(7),
		"int64":  int64(99),
		"string": "no",
		"nil":    nil,
	}
	tests := []struct {
		name  string
		field string
		want  int
	}{
		{"float", "float", 42},
		{"int", "int", 7},
		{"int64", "int64", 99},
		{"string", "string", 0},
		{"missing", "missing", 0},
		{"nil", "nil", 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, getIntValue(data, tt.field))
		})
	}
}

func TestGetBoolValue(t *testing.T) {
	t.Parallel()
	data := map[string]any{
		"true":  true,
		"false": false,
		"str":   "yes",
	}
	tests := []struct {
		name  string
		field string
		want  bool
	}{
		{"true", "true", true},
		{"false", "false", false},
		{"non-bool", "str", false},
		{"missing", "missing", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, getBoolValue(data, tt.field))
		})
	}
}

func TestResolveURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		base string
		ref  string
		want string
	}{
		{"relative path", "https://x.com/a/b", "/c", "https://x.com/c"},
		{"absolute ref", "https://x.com", "https://other.com/y", "https://other.com/y"},
		{"empty ref", "https://x.com/a", "", "https://x.com/a"},
		{"animefire", models.AnimeFireURL, "/animes/naruto", "https://animefire.io/animes/naruto"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, resolveURL(tt.base, tt.ref))
		})
	}
}

func TestNormalizeAccents(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"no accents", "Naruto", "Naruto"},
		{"vowels lower", "ãéíóú", "aeiou"},
		{"vowels upper", "ÁÉÍÓÚ", "AEIOU"},
		{"cedilla", "Clássico", "Classico"},
		{"tilde n", "señor", "senor"},
		{"mixed", "Pokémon: A Lenda do Trovão", "Pokemon: A Lenda do Trovao"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, normalizeAccents(tt.in))
		})
	}
}

func TestParseAnimes(t *testing.T) {
	t.Parallel()

	t.Run("parses anchor list", func(t *testing.T) {
		t.Parallel()
		html := `<html><body>
<div class="row ml-1 mr-1">
  <a href="/animes/naruto">Naruto</a>
  <a href="/animes/bleach">Bleach</a>
</div>
</body></html>`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)
		got := ParseAnimes(doc)
		require.Len(t, got, 2)
		assert.Equal(t, "Naruto", got[0].Name)
		assert.Contains(t, got[0].URL, "/animes/naruto")
	})

	t.Run("empty doc returns empty", func(t *testing.T) {
		t.Parallel()
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(`<html></html>`))
		require.NoError(t, err)
		got := ParseAnimes(doc)
		assert.Empty(t, got)
	})

	t.Run("anchor without href skipped", func(t *testing.T) {
		t.Parallel()
		html := `<div class="row ml-1 mr-1"><a>NoHref</a></div>`
		doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
		require.NoError(t, err)
		got := ParseAnimes(doc)
		assert.Empty(t, got)
	})
}

type errCloser struct{ err error }

func (e errCloser) Close() error { return e.err }

func TestSafeClose(t *testing.T) {
	t.Parallel()

	t.Run("no error", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() { safeClose(errCloser{nil}, "ok") })
	})

	t.Run("error logged", func(t *testing.T) {
		t.Parallel()
		assert.NotPanics(t, func() { safeClose(errCloser{errors.New("boom")}, "bad") })
	})
}
