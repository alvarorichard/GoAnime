package models

import "testing"

func TestRelevanceScore_TodoPoderoso(t *testing.T) {
	cases := []struct {
		title string
		year  string
		want  int // 0 = unrelated, >0 = related; exact value checked loosely
	}{
		{"[Movie] [PT-BR] Todo Poderoso", "2003", 80},
		{"[Movie] [PT-BR] A Volta do Todo Poderoso", "2007", 80},
		{"[Movie] [PT-BR] A Casa Mórbida", "2024", 0},
		{"[Movie] [PT-BR] Eu Voltarei", "2023", 0},
		{"[Movie] [PT-BR] Conclave", "2024", 0},
		{"[Movie] [PT-BR] Sonic 3: O Filme", "2024", 0},
	}
	for _, c := range cases {
		got := RelevanceScore("todo poderoso", c.title, c.year)
		if c.want == 0 && got != 0 {
			t.Errorf("RelevanceScore(todo poderoso, %q) = %d, want 0", c.title, got)
		}
		if c.want > 0 && got <= 0 {
			t.Errorf("RelevanceScore(todo poderoso, %q) = %d, want >0", c.title, got)
		}
	}
	if got := RelevanceScore("todo poderoso", "[Movie] [PT-BR] Todo Poderoso", "2003"); got < 80 {
		t.Errorf("exact should be >=80, got %d", got)
	}
	// Year bonus: query with year prefers the matching cut.
	a := RelevanceScore("todo poderoso 2003", "[Movie] [PT-BR] Todo Poderoso", "2003")
	b := RelevanceScore("todo poderoso 2003", "[Movie] [PT-BR] A Volta do Todo Poderoso", "2007")
	if a <= b {
		t.Errorf("year match should outrank mismatch: 2003=%d 2007=%d", a, b)
	}
}
