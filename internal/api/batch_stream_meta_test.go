package api

import "testing"

// Tests of the tests, for the first step of the player reading a Stream:
// providers.FetchStream (a resolve with no side effects) and the batch
// download that takes its referer, origin and subtitles from it instead of from
// the player's globals. Same method as stream_publish_meta_test.go; the runner
// is shared.

const (
	playerPkg    = "./internal/player"
	downloadFile = "internal/player/download.go"
	fetchStreamP = "\tif sf, ok := src.(source.StreamFetcher); ok {\n\t\treturn sf.FetchStream(ctx, episode, anime, quality)\n\t}\n"
	resolveBatch = "func resolveBatchEpisode(episode models.Episode, anime *models.Anime) (batchEpisode, error) {\n"
)

var batchStreamGuards = map[string][]string{
	playerPkg: {"TestResolveBatchEpisodeKeepsTheStreamsContext", "TestBatchEpisodesKeepTheirOwnRefererAndSubtitles"},
}

var batchStreamMutants = []mutant{
	// ---- providers.FetchStream: a resolve with no side effects -----------------
	{
		name:  "FetchStream clears the playing episode's referer",
		edits: []codeEdit{{dispatchFile, fetchStreamP, "\tutil.ClearGlobalReferer()\n" + fetchStreamP}},
		kills: map[string][]string{providersPkg: {"TestFetchStreamTouchesNoPlaybackState"}},
	},
	{
		name:  "FetchStream clears the playing episode's subtitles",
		edits: []codeEdit{{dispatchFile, fetchStreamP, "\tutil.ClearGlobalSubtitles()\n" + fetchStreamP}},
		kills: map[string][]string{providersPkg: {"TestFetchStreamTouchesNoPlaybackState"}},
	},
	{
		name:  "FetchStream tags the source behind the caller's back",
		edits: []codeEdit{{dispatchFile, fetchStreamP, "\tutil.SetGlobalAnimeSource(anime.Source)\n" + fetchStreamP}},
		kills: map[string][]string{providersPkg: {"TestFetchStreamTouchesNoPlaybackState"}},
	},
	{
		name:  "FetchStream publishes what it resolves",
		edits: []codeEdit{{dispatchFile, fetchStreamP, "\tif sf, ok := src.(source.StreamFetcher); ok {\n\t\ts, err := sf.FetchStream(ctx, episode, anime, quality)\n\t\ts.Publish()\n\t\treturn s, err\n\t}\n"}},
		kills: map[string][]string{providersPkg: {"TestFetchStreamTouchesNoPlaybackState", "TestFetchStreamURLEqualsFetchStreamThenPublish"}},
	},
	{
		name:  "FetchStream ignores a source's own Stream",
		edits: []codeEdit{{dispatchFile, fetchStreamP, ""}},
		kills: map[string][]string{providersPkg: {"TestFetchStreamPassesTheArgumentsToTheSource", "TestFetchStreamTouchesNoPlaybackState"}},
	},
	{
		name:  "FetchStream drops the URL of a legacy source",
		edits: []codeEdit{{dispatchFile, "return source.Stream{URL: url}, err", "_ = url\n\treturn source.Stream{}, err"}},
		kills: map[string][]string{providersPkg: {"TestFetchStreamForALegacySourceGivesJustTheURL"}},
	},
	{
		name:  "FetchStream resolves a nil anime",
		edits: []codeEdit{{dispatchFile, "\tif anime == nil {\n\t\treturn nil, source.ResolvedSource{}", "\tif anime == nil && false {\n\t\treturn nil, source.ResolvedSource{}"}},
		kills: map[string][]string{providersPkg: {"TestFetchStreamRejectsANilAnime"}},
	},
	{
		name:  "the strict-source refusal is skipped",
		edits: []codeEdit{{dispatchFile, "\t\tif util.StrictSourceResolution() {\n\t\t\treturn nil, resolved, fmt.Errorf(", "\t\tif util.StrictSourceResolution() && false {\n\t\t\treturn nil, resolved, fmt.Errorf("}},
		kills: map[string][]string{providersPkg: {"TestFetchStreamRefusesAnUnrecognizedSourceWhenStrict"}},
	},
	{
		name:  "a browser-gated source is not warmed up",
		edits: []codeEdit{{dispatchFile, "\tif err := source.WarmUp(ctx, src); err != nil {\n\t\treturn nil, resolved, err\n\t}\n", ""}},
		kills: map[string][]string{providersPkg: {"TestFetchStreamWarmsUpABrowserGatedSourceFirst", "TestFetchStreamReportsAWarmUpThatCannotRun"}},
	},
	{
		name:  "a warm-up that cannot run is ignored",
		edits: []codeEdit{{dispatchFile, "\tif err := source.WarmUp(ctx, src); err != nil {\n\t\treturn nil, resolved, err\n\t}\n", "\t_ = source.WarmUp(ctx, src)\n"}},
		kills: map[string][]string{providersPkg: {"TestFetchStreamReportsAWarmUpThatCannotRun"}},
	},

	// ---- the batch takes its context from the Stream ----------------------------
	{
		name:  "the batch reads the global referer, not the stream's",
		edits: []codeEdit{{downloadFile, "be.referer, be.origin = stream.Referer, stream.Origin", "be.referer, be.origin = util.GetGlobalReferer(), stream.Origin"}},
		kills: map[string][]string{playerPkg: {"TestResolveBatchEpisodeKeepsTheStreamsContext", "TestResolveBatchEpisodeNeverReadsThePlayersGlobals"}},
	},
	{
		name:  "the batch reads the global origin, not the stream's",
		edits: []codeEdit{{downloadFile, "be.referer, be.origin = stream.Referer, stream.Origin", "be.referer, be.origin = stream.Referer, util.GetGlobalOrigin()"}},
		kills: map[string][]string{playerPkg: {"TestResolveBatchEpisodeKeepsTheStreamsContext", "TestResolveBatchEpisodeNeverReadsThePlayersGlobals"}},
	},
	{
		name:  "the batch loses each episode's origin",
		edits: []codeEdit{{downloadFile, "be.referer, be.origin = stream.Referer, stream.Origin", "be.referer = stream.Referer"}},
		kills: map[string][]string{playerPkg: {"TestResolveBatchEpisodeKeepsTheStreamsContext", "TestBatchEpisodesKeepTheirOwnOriginInParallel"}},
	},
	{
		name:  "the batch reads the global subtitles, not the stream's",
		edits: []codeEdit{{downloadFile, "be.subtitles = streamSubtitles(stream)", "be.subtitles = util.GetGlobalSubtitles()"}},
		kills: map[string][]string{playerPkg: {"TestResolveBatchEpisodeKeepsTheStreamsContext", "TestBatchEpisodesKeepTheirOwnRefererAndSubtitles"}},
	},
	{
		name:  "9Anime's chosen tracks are replaced by the stream's",
		edits: []codeEdit{{downloadFile, "\tif !nineAnime {\n\t\tbe.subtitles = streamSubtitles(stream)", "\tif true {\n\t\tbe.subtitles = streamSubtitles(stream)"}},
		kills: map[string][]string{playerPkg: {"TestResolveBatchEpisodeKeepsTheUsersLanguageChoiceFor9Anime"}},
	},
	{
		name:  "9Anime's chosen tracks are cleared",
		edits: []codeEdit{{downloadFile, "\tnineAnime := util.Is9AnimeSource()\n\tif !nineAnime {\n\t\t// A batch leaves", "\tnineAnime := util.Is9AnimeSource()\n\tif true {\n\t\t// A batch leaves"}},
		kills: map[string][]string{playerPkg: {"TestResolveBatchEpisodeKeepsTheUsersLanguageChoiceFor9Anime"}},
	},
	{
		name:  "an earlier episode's online tracks are left for a downloaded file",
		edits: []codeEdit{{downloadFile, "\t\tutil.ClearGlobalSubtitles()\n\t}\n\t// The tracking of what is watched afterwards", "\t}\n\t// The tracking of what is watched afterwards"}},
		kills: map[string][]string{playerPkg: {"TestResolveBatchEpisodeClearsTheTracksAnEarlierEpisodeLeft"}},
	},
	{
		name:  "the batch stops tagging the source",
		edits: []codeEdit{{downloadFile, "\tif anime != nil && anime.Source != \"\" {\n\t\tutil.SetGlobalAnimeSource(anime.Source)\n\t}\n", ""}},
		kills: map[string][]string{playerPkg: {"TestResolveBatchEpisodeTagsTheSource"}},
	},
	{
		name:  "an anime with no source clears the tag",
		edits: []codeEdit{{downloadFile, "\tif anime != nil && anime.Source != \"\" {\n\t\tutil.SetGlobalAnimeSource(anime.Source)", "\tif anime != nil {\n\t\tutil.SetGlobalAnimeSource(anime.Source)"}},
		kills: map[string][]string{playerPkg: {"TestResolveBatchEpisodeTagsTheSource"}},
	},
	{
		name:  "a missing anime crashes the batch",
		edits: []codeEdit{{downloadFile, "\tif anime != nil && anime.Source != \"\" {\n\t\tutil.SetGlobalAnimeSource(anime.Source)", "\tif anime.Source != \"\" {\n\t\tutil.SetGlobalAnimeSource(anime.Source)"}},
		kills: map[string][]string{playerPkg: {"TestResolveBatchEpisodeTagsTheSource"}},
	},
	{
		name:  "resolving a batch episode overwrites the playing referer",
		edits: []codeEdit{{downloadFile, resolveBatch + "\tnineAnime := util.Is9AnimeSource()\n", resolveBatch + "\tnineAnime := util.Is9AnimeSource()\n\tutil.ClearGlobalReferer()\n"}},
		kills: map[string][]string{playerPkg: {"TestResolveBatchEpisodeLeavesThePlayingRefererOriginAndAudioAlone"}},
	},
	{
		name:  "an empty stream URL is taken as resolved",
		edits: []codeEdit{{downloadFile, "if err != nil || stream.URL == \"\" {", "if err != nil {"}},
		kills: map[string][]string{playerPkg: {"TestResolveBatchEpisodeFailures"}},
	},
	{
		name:  "the downloaded URL is not the stream's",
		edits: []codeEdit{{downloadFile, "be.sourceURL = stream.URL\n\tif be.url, err = resolveDownloadURL(stream.URL)", "be.sourceURL = stream.URL\n\tif be.url, err = resolveDownloadURL(be.sourceURL + \"?x\")"}},
		kills: map[string][]string{playerPkg: {"TestResolveBatchEpisodeKeepsTheStreamsContext"}},
	},
	{
		name:  "a subtitle's language and label are swapped",
		edits: []codeEdit{{downloadFile, "util.SubtitleInfo{URL: sub.URL, Language: sub.Language, Label: sub.Label}", "util.SubtitleInfo{URL: sub.URL, Language: sub.Label, Label: sub.Language}"}},
		kills: map[string][]string{playerPkg: {"TestResolveBatchEpisodeKeepsTheStreamsContext"}},
	},
	{
		name:  "only the first subtitle track survives",
		edits: []codeEdit{{downloadFile, "for _, sub := range s.Subtitles {\n\t\tsubs = append(subs, util.SubtitleInfo{", "for _, sub := range s.Subtitles[:1] {\n\t\tsubs = append(subs, util.SubtitleInfo{"}},
		kills: map[string][]string{playerPkg: {"TestStreamSubtitlesKeepsEveryTrack"}},
	},

	// ---- the per-download request context ----------------------------------------
	{
		name:  "a pinned origin is ignored",
		edits: []codeEdit{{downloadFile, "if c := v.(downloadContext); c.hasOrigin {", "if c := v.(downloadContext); c.hasOrigin && false {"}},
		kills: map[string][]string{playerPkg: {"TestDownloadContextPinsAreIndependentOfTheGlobals", "TestNativeHLSSendsThePinnedOrigin"}},
	},
	{
		name:  "an empty pinned origin falls back to the global one",
		edits: []codeEdit{{downloadFile, "if c := v.(downloadContext); c.hasOrigin {\n\t\t\treturn c.origin\n\t\t}", "if c := v.(downloadContext); c.hasOrigin && c.origin != \"\" {\n\t\t\treturn c.origin\n\t\t}"}},
		kills: map[string][]string{playerPkg: {"TestDownloadContextPinsAreIndependentOfTheGlobals", "TestBatchEpisodesKeepTheirOwnOriginInParallel", "TestNativeHLSSendsThePinnedOrigin"}},
	},
	{
		name:  "rememberDownloadContext forgets to say the origin is pinned",
		edits: []codeEdit{{downloadFile, "downloadContext{referer: referer, origin: origin, hasOrigin: true}", "downloadContext{referer: referer, origin: origin}"}},
		kills: map[string][]string{playerPkg: {"TestDownloadContextPinsAreIndependentOfTheGlobals", "TestNativeHLSSendsThePinnedOrigin"}},
	},
	{
		name:  "a referer-only pin claims the origin too",
		edits: []codeEdit{{downloadFile, "downloadContexts.Store(url, downloadContext{referer: referer})", "downloadContexts.Store(url, downloadContext{referer: referer, hasOrigin: true})"}},
		kills: map[string][]string{playerPkg: {"TestDownloadContextPinsAreIndependentOfTheGlobals", "TestNativeHLSSendsThePinnedOrigin"}},
	},
	{
		name:  "forgetting a pin leaves it behind",
		edits: []codeEdit{{downloadFile, "downloadContexts.Store(url, downloadContext{referer: referer, origin: origin, hasOrigin: true})\n\treturn func() { downloadContexts.Delete(url) }", "downloadContexts.Store(url, downloadContext{referer: referer, origin: origin, hasOrigin: true})\n\treturn func() {}"}},
		kills: map[string][]string{playerPkg: {"TestDownloadContextPinsAreIndependentOfTheGlobals"}},
	},
	{
		name:  "native HLS reads the global origin",
		edits: []codeEdit{{downloadFile, "origin := downloadOrigin(streamURL)", "origin := util.GetGlobalOrigin()"}},
		kills: map[string][]string{playerPkg: {"TestNativeHLSSendsThePinnedOrigin"}},
	},
	{
		name:  "a batch download pins only the referer",
		edits: []codeEdit{{downloadFile, "defer rememberDownloadContext(be.url, be.referer, be.origin)()", "defer rememberDownloadReferer(be.url, be.referer)()"}},
		kills: map[string][]string{playerPkg: {"TestBatchEpisodesKeepTheirOwnOriginInParallel"}},
	},

	// ---- getBestQualityStream ------------------------------------------------------
	{
		name:  "the resolver's error is not named",
		edits: []codeEdit{{downloadFile, "return source.Stream{}, fmt.Errorf(\"failed to get stream URL from %s for episode %s: %w\", src, epNumber, err)", "return source.Stream{}, err"}},
		kills: map[string][]string{playerPkg: {"TestGetBestQualityStreamNamesTheSourceAndEpisodeOnFailure"}},
	},
	{
		name:  "an empty stream URL passes the resolver",
		edits: []codeEdit{{downloadFile, "\t\tif stream.URL == \"\" {\n\t\t\treturn source.Stream{}, fmt.Errorf(\"empty stream URL from %s for episode %s\", src, epNumber)\n\t\t}\n", ""}},
		kills: map[string][]string{playerPkg: {"TestGetBestQualityStreamNamesTheSourceAndEpisodeOnFailure"}},
	},
	{
		name:  "an episode with only a Num is not numbered",
		edits: []codeEdit{{downloadFile, "if epNumber == \"\" && episode.Num > 0 {", "if false && epNumber == \"\" && episode.Num > 0 {"}},
		kills: map[string][]string{playerPkg: {"TestGetBestQualityStreamNumbersAnEpisodeThatHasOnlyNum"}},
	},
	{
		name:  "the episode's season never reaches the source",
		edits: []codeEdit{{downloadFile, "\t\t\tDataID:   episode.DataID,\n\t\t\tSeasonID: episode.SeasonID,\n\t\t}\n\n\t\tutil.Debugf(\"getBestQualityStream", "\t\t\tDataID:   episode.DataID,\n\t\t}\n\n\t\tutil.Debugf(\"getBestQualityStream"}},
		kills: map[string][]string{playerPkg: {"TestGetBestQualityStreamReturnsTheSourcesStreamWithoutPublishingIt"}},
	},
	{
		name:  "the configured quality never reaches the source",
		edits: []codeEdit{{downloadFile, "providers.FetchStream(context.Background(), ep, anime, util.GlobalQuality)", "providers.FetchStream(context.Background(), ep, anime, \"\")"}},
		kills: map[string][]string{playerPkg: {"TestGetBestQualityStreamReturnsTheSourcesStreamWithoutPublishingIt"}},
	},
	{
		name:  "getBestQualityStream publishes the stream",
		edits: []codeEdit{{downloadFile, "\t\treturn stream, nil\n\t}\n\n\t// Legacy HTTP page URL path.", "\t\tstream.Publish()\n\t\treturn stream, nil\n\t}\n\n\t// Legacy HTTP page URL path."}},
		kills: map[string][]string{playerPkg: {"TestGetBestQualityStreamReturnsTheSourcesStreamWithoutPublishingIt"}},
	},
}

func TestBatchStreamGuardKillsMutants(t *testing.T) {
	killMutants(t, batchStreamGuards, batchStreamMutants)
}
