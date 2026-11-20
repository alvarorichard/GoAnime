package main

import (
	"os"
	"strings"
	"testing"
)

// "go run cmd/goanime/main.go" compiles that one file and nothing else in the
// folder. A second .go file here would leave main.go with undefined names and
// break that way of running the program (it did, once). Put new code in
// internal/ and call it from main.go.
func TestMainIsASingleFileSoGoRunMainDotGoWorks(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		name := e.Name()
		if strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go") && name != "main.go" {
			t.Errorf("%s: cmd/goanime must hold only main.go (go run cmd/goanime/main.go builds that file alone); move it into internal/", name)
		}
	}
}
