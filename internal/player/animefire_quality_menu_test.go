// Package player — regression tests for the AnimeFire quality picker.
//
// 2026-04-28: the AnimeFire quality picker (extractActualVideoURL inline)
// had several visible UX bugs surfaced by a screenshot from the user:
//
//  1. items were displayed in whatever order AnimeFire returned them, so
//     users could see 360p above 1080p depending on API order;
//  2. labels like "FHD" / "HD" / empty rendered raw — confusing or empty
//     menu rows;
//  3. same-quality mirrors produced visually identical entries — the user
//     had no way to tell them apart from inside the picker;
//  4. Esc surfaced as a generic "failed to select quality" error instead
//     of routing back to the menu.
//
// The fix moved the sort + label-rendering logic into
// buildAnimeFireQualityItems and made the call site treat fuzzyfinder.ErrAbort
// as ErrBackRequested. These tests pin the new contract for #1–#3.
package player

import (
	"slices"
	"strings"
	"testing"
)

// Bug 1 regression: descending sort independent of upstream order.
func TestBuildAnimeFireQualityItems_SortsDescendingByParsedResolution(t *testing.T) {
	t.Parallel()

	in := []VideoData{
		{Label: "360p", Src: "https://cdn/360.mp4"},
		{Label: "1080p", Src: "https://cdn/1080.mp4"},
		{Label: "720p", Src: "https://cdn/720.mp4"},
	}
	sorted, labels := buildAnimeFireQualityItems(in)

	wantOrder := []string{"1080p", "720p", "360p"}
	for i, w := range wantOrder {
		if labels[i] != w {
			t.Fatalf("labels[%d] = %q, want %q (full=%v)", i, labels[i], w, labels)
		}
		if sorted[i].Label != w {
			t.Fatalf("sorted[%d].Label = %q, want %q", i, sorted[i].Label, w)
		}
	}
}

// AnimeFire occasionally serves non-numeric labels ("FHD", "HD", "SD").
// Parsed resolution is unknown, so they must render with a usable
// fallback — never an empty string and never the raw Src URL.
func TestBuildAnimeFireQualityItems_RendersFriendlyLabelForNonNumeric(t *testing.T) {
	t.Parallel()

	in := []VideoData{
		{Label: "FHD", Src: "https://cdn/fhd.mp4"},
		{Label: "", Src: "https://cdn/blank.mp4"},
		{Label: "HD", Src: "https://cdn/hd.mp4"},
	}
	_, labels := buildAnimeFireQualityItems(in)

	for i, l := range labels {
		if l == "" {
			t.Fatalf("labels[%d] is empty — picker would render a blank row", i)
		}
		if strings.HasPrefix(l, "http") {
			t.Fatalf("labels[%d] = %q — raw URL must never be shown", i, l)
		}
	}

	// Empty label specifically must collapse to "Auto", not the URL.
	foundAuto := slices.Contains(labels, "Auto")
	if !foundAuto {
		t.Fatalf("expected one entry to render as 'Auto' for the empty label; got %v", labels)
	}
}

// Bug 3 regression: same numeric quality across mirrors must produce
// distinct labels — otherwise the index returned by the fuzzyfinder picks
// the wrong source under the previous string-match logic.
func TestBuildAnimeFireQualityItems_DisambiguatesSameQualityMirrors(t *testing.T) {
	t.Parallel()

	in := []VideoData{
		{Label: "720p", Src: "https://mirror-a/720.mp4"},
		{Label: "720p", Src: "https://mirror-b/720.mp4"},
		{Label: "720p", Src: "https://mirror-c/720.mp4"},
	}
	sorted, labels := buildAnimeFireQualityItems(in)

	uniq := map[string]struct{}{}
	for _, l := range labels {
		uniq[l] = struct{}{}
	}
	if len(uniq) != 3 {
		t.Fatalf("labels must be unique to keep selection unambiguous, got %v", labels)
	}
	if labels[0] != "720p" {
		t.Fatalf("first occurrence must keep bare label, got %q", labels[0])
	}
	for i := 1; i < 3; i++ {
		if !strings.Contains(labels[i], "mirror") {
			t.Fatalf("labels[%d] must include 'mirror' suffix, got %q", i, labels[i])
		}
	}

	// Stable: order of equal-resolution items mirrors input order so
	// the index-based selection in the call site picks the exact source
	// the user highlighted.
	wantSrc := []string{
		"https://mirror-a/720.mp4",
		"https://mirror-b/720.mp4",
		"https://mirror-c/720.mp4",
	}
	for i, w := range wantSrc {
		if sorted[i].Src != w {
			t.Fatalf("sorted[%d].Src = %q, want %q (stable sort broken)", i, sorted[i].Src, w)
		}
	}
}

// Locks the alignment invariant the call site depends on: labels[i]
// describes sorted[i], and indexing the user's chosen idx into sorted
// returns the exact VideoData they picked.
func TestBuildAnimeFireQualityItems_LengthsAndAlignment(t *testing.T) {
	t.Parallel()

	in := []VideoData{
		{Label: "1080p", Src: "https://cdn/1080.mp4"},
		{Label: "720p", Src: "https://m1/720.mp4"},
		{Label: "720p", Src: "https://m2/720.mp4"},
		{Label: "FHD", Src: "https://cdn/fhd.mp4"},
		{Label: "", Src: "https://cdn/blank.mp4"},
	}
	sorted, labels := buildAnimeFireQualityItems(in)

	if len(sorted) != len(labels) {
		t.Fatalf("len(sorted)=%d must equal len(labels)=%d", len(sorted), len(labels))
	}
	if len(sorted) != len(in) {
		t.Fatalf("len(sorted)=%d must preserve input cardinality (%d)", len(sorted), len(in))
	}
	for i := range sorted {
		for j := i + 1; j < len(labels); j++ {
			if labels[i] == labels[j] {
				t.Fatalf("labels[%d] == labels[%d] == %q breaks index alignment", i, j, labels[i])
			}
		}
	}
}

func TestBuildAnimeFireQualityItems_DoesNotMutateInput(t *testing.T) {
	t.Parallel()

	in := []VideoData{
		{Label: "360p", Src: "https://cdn/360.mp4"},
		{Label: "1080p", Src: "https://cdn/1080.mp4"},
		{Label: "720p", Src: "https://cdn/720.mp4"},
	}
	snapshot := append([]VideoData(nil), in...)

	_, _ = buildAnimeFireQualityItems(in)

	for i := range snapshot {
		if in[i] != snapshot[i] {
			t.Fatalf("input mutated at %d: got %+v want %+v", i, in[i], snapshot[i])
		}
	}
}

func TestBuildAnimeFireQualityItems_EmptyAndSingle(t *testing.T) {
	t.Parallel()

	sorted, labels := buildAnimeFireQualityItems(nil)
	if len(sorted) != 0 || len(labels) != 0 {
		t.Fatalf("nil input must yield empty slices; got sorted=%v labels=%v", sorted, labels)
	}

	in := []VideoData{{Label: "1080p", Src: "https://cdn/1080.mp4"}}
	sorted, labels = buildAnimeFireQualityItems(in)
	if len(sorted) != 1 || sorted[0].Src != "https://cdn/1080.mp4" {
		t.Fatalf("single source must be preserved; got %+v", sorted)
	}
	if len(labels) != 1 || labels[0] != "1080p" {
		t.Fatalf("single source must render bare label; got %v", labels)
	}
}
