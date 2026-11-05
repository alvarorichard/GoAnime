package playback

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// TestHandleDownloadAndPlay_CallSites_PassDistinctIDFields is a structural
// regression guard for the original bug: anime.MalID was being routed
// through HandleDownloadAndPlay's anilistID slot.
//
// The fix added a separate animeAnilistID parameter (positional index 6,
// counting from 0). Without this test the player-package routing tests
// would pass even if a future refactor wrote
//
//	player.HandleDownloadAndPlay(..., anime.MalID, anime.MalID, ...)
//
// at the call site — the bug would compile, would not break any unit test
// downstream, and would silently restore the original failure mode.
//
// This test parses the two call sites in the playback package and asserts:
//  1. positional index 5 (the 6th argument) reads field MalID, and
//  2. positional index 6 (the 7th argument) reads field AnilistID.
//
// If either invariant breaks, the test fails with a precise message.
func TestHandleDownloadAndPlay_CallSites_PassDistinctIDFields(t *testing.T) {
	// Files that contain a call to player.HandleDownloadAndPlay.
	// If a new call site is added, register it here so the contract
	// is enforced everywhere.
	callSites := []string{
		"common.go",
		"movie.go",
	}

	for _, path := range callSites {
		t.Run(path, func(t *testing.T) {
			fset := token.NewFileSet()
			f, err := parser.ParseFile(fset, path, nil, parser.SkipObjectResolution)
			if err != nil {
				t.Fatalf("parse %s: %v", path, err)
			}

			var found bool
			ast.Inspect(f, func(n ast.Node) bool {
				call, ok := n.(*ast.CallExpr)
				if !ok {
					return true
				}
				sel, ok := call.Fun.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				recv, ok := sel.X.(*ast.Ident)
				if !ok {
					return true
				}
				// The call is player.HandleDownloadAndPlay itself, or the
				// flowEnv field that stands in for it (same arguments).
				direct := recv.Name == "player" && sel.Sel.Name == "HandleDownloadAndPlay"
				viaEnv := recv.Name == "env" && sel.Sel.Name == "downloadAndPlay"
				if !direct && !viaEnv {
					return true
				}
				found = true

				// The function's contract: 11 positional args. Bail explicitly
				// if the signature changed so the failure message is useful.
				if len(call.Args) < 11 {
					t.Fatalf("HandleDownloadAndPlay called with %d args at %s, want 11",
						len(call.Args), fset.Position(call.Pos()))
				}

				const (
					malIdx     = 5 // 6th positional arg → animeMalID
					anilistIdx = 6 // 7th positional arg → animeAnilistID
				)

				malField := selectorTail(call.Args[malIdx])
				anilistField := selectorTail(call.Args[anilistIdx])

				if malField != "MalID" {
					t.Errorf("at %s: arg #%d (animeMalID slot) reads field %q, want %q",
						fset.Position(call.Args[malIdx].Pos()), malIdx+1, malField, "MalID")
				}
				if anilistField != "AnilistID" {
					t.Errorf("at %s: arg #%d (animeAnilistID slot) reads field %q, want %q",
						fset.Position(call.Args[anilistIdx].Pos()), anilistIdx+1, anilistField, "AnilistID")
				}

				if malField == anilistField && malField != "" {
					t.Errorf("at %s: BUG REGRESSION — same field %q passed to both "+
						"animeMalID and animeAnilistID. The original bug (MAL ID treated "+
						"as AniList ID) has returned.",
						fset.Position(call.Pos()), malField)
				}
				return true
			})

			if !found {
				t.Fatalf("no player.HandleDownloadAndPlay (or env.downloadAndPlay) call found in %s — "+
					"either the call site was moved (update this test) or removed", path)
			}
		})
	}
}

// selectorTail returns the trailing field name of a selector expression
// (e.g. anime.MalID -> "MalID"), or "" if the expression is not a selector.
func selectorTail(e ast.Expr) string {
	if s, ok := e.(*ast.SelectorExpr); ok {
		return s.Sel.Name
	}
	return ""
}
