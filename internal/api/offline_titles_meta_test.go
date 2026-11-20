package api

import "testing"

// Tests of the tests, for hiding AnimeFire titles whose video was removed:
// playback records a title (rememberRemovedTitle), the offline cache keeps it,
// and the next AnimeFire search leaves it out (HideOfflineTitles). Same method
// as stream_publish_meta_test.go; the runner is shared.

const (
	otCachePkg     = "./internal/scraper/offlinecache"
	otCacheFile    = "internal/scraper/offlinecache/cache.go"
	otProvidersGo  = "internal/api/providers/source_providers.go"
	otFallbackFile = "internal/playback/fallback.go"
)

var offlineTitlesMutants = []mutant{
	// ---- the store -----------------------------------------------------------------
	{
		name:  "a recorded title never expires",
		edits: []codeEdit{{otCacheFile, "ok && now < exp {\n\t\t\tout[u] = true", "ok || now < exp || true {\n\t\t\tout[u] = true"}},
		kills: map[string][]string{otCachePkg: {"TestExpiry"}},
	},
	{
		name:  "the source is matched case-sensitively",
		edits: []codeEdit{{otCacheFile, "strings.ToLower(source) + \"|\" + titleURL", "source + \"|\" + titleURL"}},
		kills: map[string][]string{otCachePkg: {"TestMarkThenSet"}},
	},
	{
		name:  "a verdict on one source applies to every source",
		edits: []codeEdit{{otCacheFile, "return strings.ToLower(source) + \"|\" + titleURL", "return titleURL"}},
		kills: map[string][]string{otCachePkg: {"TestMarkThenSet"}},
	},
	{
		name:  "a save drops what another process recorded",
		edits: []codeEdit{{otCacheFile, "\tc.mergeDiskLocked()\n\tc.entries[key(source, titleURL)] =", "\tc.entries[key(source, titleURL)] ="}},
		kills: map[string][]string{otCachePkg: {"TestTwoInstancesDoNotEraseEachOther"}},
	},
	{
		name:  "the switch that turns hiding off does nothing",
		edits: []codeEdit{{otCacheFile, "\t\treturn false\n\t}\n\treturn true", "\t\treturn true\n\t}\n\treturn true"}},
		kills: map[string][]string{otCachePkg: {"TestHidingEnabled"}},
	},

	// ---- the search filter ---------------------------------------------------------
	{
		name:  "the search keeps the titles that were recorded and drops the rest",
		edits: []codeEdit{{otProvidersGo, "\t\tif !recorded[r.URL] {", "\t\tif recorded[r.URL] {"}},
		kills: map[string][]string{providersPkg: {"TestHideOfflineTitles"}},
	},
	{
		name:  "a search whose every title is recorded comes back empty",
		edits: []codeEdit{{otProvidersGo, "\tif len(kept) == 0 {\n\t\treturn results\n\t}\n", ""}},
		kills: map[string][]string{providersPkg: {"TestHideOfflineTitles_AllRecordedKeepsList"}},
	},
	{
		name:  "the switch that turns hiding off is ignored by the search",
		edits: []codeEdit{{otProvidersGo, "if !offlinecache.HidingEnabled() || len(results) == 0 {", "if len(results) == 0 {"}},
		kills: map[string][]string{providersPkg: {"TestHideOfflineTitles_EnvSwitchesOff"}},
	},

	// ---- playback records, search hides: the two ends together ---------------------
	{
		name:  "playback records a title under another source than the search reads",
		edits: []codeEdit{{otFallbackFile, "kind := string(apisource.AnimeFire)", "kind := \"Elsewhere\""}},
		kills: map[string][]string{playbackPkg: {"TestRememberedRemovedTitleIsHiddenFromTheNextSearch"}},
	},
	{
		name:  "playback records another address than the one the search lists",
		edits: []codeEdit{{otFallbackFile, "MarkOffline(kind, anime.URL)", "MarkOffline(kind, anime.URL+\"/\")"}},
		kills: map[string][]string{playbackPkg: {"TestRememberedRemovedTitleIsHiddenFromTheNextSearch"}},
	},
	{
		name:  "a failure that is not a removal is recorded",
		edits: []codeEdit{{otFallbackFile, "!errors.Is(videoErr, player.ErrSourceVideoRemoved)", "errors.Is(videoErr, player.ErrSourceVideoRemoved) && false"}},
		kills: map[string][]string{playbackPkg: {"TestRememberRemovedTitle", "TestRememberedRemovedTitleIsHiddenFromTheNextSearch"}},
	},
	{
		name:  "a title from another source is recorded as AnimeFire's",
		edits: []codeEdit{{otFallbackFile, "if !strings.HasPrefix(strings.ToLower(anime.Source), strings.ToLower(kind)) {", "if false {"}},
		kills: map[string][]string{playbackPkg: {"TestRememberRemovedTitle"}},
	},
}

var offlineTitlesGuards = map[string][]string{
	otCachePkg:   {"TestMarkThenSet"},
	providersPkg: {"TestHideOfflineTitles"},
	playbackPkg:  {"TestRememberedRemovedTitleIsHiddenFromTheNextSearch"},
}

func TestOfflineTitlesGuardKillsMutants(t *testing.T) {
	skipMutantsWhenShort(t)
	killMutants(t, offlineTitlesGuards, offlineTitlesMutants)
}
