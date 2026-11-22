package api

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// A --source that names no source must be refused: searching every source in
// its place told the user nothing (a typo like "animfire" searched them all).

func TestEverySourceSelectorIsSearched(t *testing.T) {
	for _, name := range SourceSelectors {
		if registryKindsForSource(name) == nil {
			t.Errorf("--source %s is listed but searches nothing of its own: it would be refused", name)
		}
		if err := CheckSourceSelector(name); err != nil {
			t.Errorf("--source %s refused: %v", name, err)
		}
		if err := CheckSourceSelector(" " + strings.ToUpper(name) + " "); err != nil {
			t.Errorf("--source %s is refused in capitals or with spaces: %v", name, err)
		}
	}
}

func TestEmptySourceMeansEverySource(t *testing.T) {
	if err := CheckSourceSelector(""); err != nil {
		t.Fatalf("no --source refused: %v", err)
	}
}

func TestAnUnknownSourceIsRefusedAndTheValidOnesAreNamed(t *testing.T) {
	for _, bad := range []string{"animfire", "naoexiste", "netflix", "all"} {
		err := CheckSourceSelector(bad)
		if err == nil {
			t.Errorf("--source %s accepted: it would search every source without a word", bad)
			continue
		}
		msg := err.Error()
		if !strings.Contains(msg, bad) {
			t.Errorf("the refusal does not name what was typed (%q): %s", bad, msg)
		}
		for _, name := range SourceSelectors {
			if !strings.Contains(msg, name) {
				t.Errorf("the refusal does not list %s among the valid sources: %s", name, msg)
			}
		}
	}
}

// What the help and the flag's own description list is what a user types:
// each must be accepted, and every selector must be listed (Pomfy was missing
// from both).
func TestTheHelpListsEverySourceSelectorAndNoOther(t *testing.T) {
	read := func(rel string) string {
		b, err := os.ReadFile(filepath.Join("..", "..", rel))
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	help := regexp.MustCompile(`"--source", "Search a single source: ([^"]*)"`).FindStringSubmatch(read("internal/util/help.go"))
	flag := regexp.MustCompile(`fs\.String\("source", "", "specify source \(([^)]*)\)"\)`).FindStringSubmatch(read("internal/util/util.go"))
	if help == nil || flag == nil {
		t.Fatal("the --source help or flag description moved; update this test with it")
	}
	words := regexp.MustCompile(`[a-z][a-z-]*`)
	for where, text := range map[string]string{"help": help[1], "flag description": flag[1]} {
		var listed []string
		for _, w := range words.FindAllString(strings.ToLower(text), -1) {
			switch w {
			case "or", "use", "for", "movie", "tv", "sources", "portuguese", "default", "search", "all", "enabled":
				continue
			}
			listed = append(listed, w)
			if err := CheckSourceSelector(w); err != nil {
				t.Errorf("the %s lists %q, which is refused: %v", where, w, err)
			}
		}
		for _, name := range SourceSelectors {
			if !slices.Contains(listed, name) {
				t.Errorf("the %s does not list --source %s", where, name)
			}
		}
	}
}
