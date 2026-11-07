package api

import (
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of the tests, for the Stream hand-off (source/stream_test.go,
// providers/stream_publish_test.go, providers/stream_publish_cascade_test.go
// and pomfy_stream_test.go). Same method as audio_guard_meta_test.go: each
// mutant puts back one way the hand-off can break, builds with go test
// -overlay, and the guard tests must fail against it. A mutant whose snippet is
// no longer in the code fails here too, so a refactor cannot quietly retire a
// guard.

const (
	sourcePkg                  = "./internal/api/source"
	providersPkg               = "./internal/api/providers"
	streamFile                 = "internal/api/source/stream.go"
	providersFile              = "internal/api/providers/source_providers.go"
	dispatchFile               = "internal/api/providers/dispatch.go"
	pomfyAPIFile               = "internal/api/pomfy.go"
	pobreflixFetchStreamURLSig = "func (p *pobreflixProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {\n"
	pbAPIFilePath              = "internal/api/pobreflix.go"
	topCineFetchStreamURLSig   = "func (p *topCineProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {\n"
	pomfyFetchStreamURLSig     = "func (p *pomfyProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {\n"
	startFlixFetchStreamURLSig = "func (p *startFlixProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {\n"
	fetchAndPublishP           = "func fetchAndPublish(ctx context.Context, sf source.StreamFetcher, episode *models.Episode, anime *models.Anime, quality string) (string, error) {\n"
)

var publishGuards = map[string][]string{
	sourcePkg: {
		"TestStreamPublishWritesEachField", "TestStreamPublishLeavesWhatTheStreamDoesNotSay",
		"TestStreamPublishFieldsAreIndependent", "TestStreamPublishHonoursNoSubs",
		"TestStreamPublishKeepsEveryTrackInOrder", "TestStreamPublishCopiesTheSubtitles",
	},
}

var streamPublishMutants = []mutant{
	// ---- Stream.Publish -------------------------------------------------------
	{
		name:  "the referer is never published",
		edits: []codeEdit{{streamFile, "\tif s.Referer != \"\" {\n\t\tutil.SetGlobalReferer(s.Referer)\n\t}\n", ""}},
		kills: map[string][]string{sourcePkg: {"TestStreamPublishWritesEachField", "TestStreamPublishFieldsAreIndependent"}},
	},
	{
		name:  "the origin is never published",
		edits: []codeEdit{{streamFile, "\tif s.Origin != \"\" {\n\t\tutil.SetGlobalOrigin(s.Origin)\n\t}\n", ""}},
		kills: map[string][]string{sourcePkg: {"TestStreamPublishWritesEachField", "TestStreamPublishFieldsAreIndependent"}},
	},
	{
		name:  "the audio hint is never published",
		edits: []codeEdit{{streamFile, "\tif s.AudioLang != \"\" {\n\t\tutil.SetGlobalAudioLanguage(s.AudioLang)\n\t}\n", ""}},
		kills: map[string][]string{sourcePkg: {"TestStreamPublishWritesEachField", "TestStreamPublishFieldsAreIndependent"}},
	},
	{
		name:  "an empty referer wipes the one already set",
		edits: []codeEdit{{streamFile, "\tif s.Referer != \"\" {\n", "\tif true {\n"}},
		kills: map[string][]string{sourcePkg: {"TestStreamPublishLeavesWhatTheStreamDoesNotSay", "TestStreamPublishFieldsAreIndependent"}},
	},
	{
		name:  "an empty audio hint wipes the one already set",
		edits: []codeEdit{{streamFile, "\tif s.AudioLang != \"\" {\n", "\tif true {\n"}},
		kills: map[string][]string{sourcePkg: {"TestStreamPublishLeavesWhatTheStreamDoesNotSay"}},
	},
	{
		name:  "--no-subs ignored",
		edits: []codeEdit{{streamFile, "if len(s.Subtitles) > 0 && !util.GlobalNoSubs {", "if len(s.Subtitles) > 0 {"}},
		kills: map[string][]string{sourcePkg: {"TestStreamPublishHonoursNoSubs"}},
	},
	{
		name:  "a stream without subtitles wipes the last one's",
		edits: []codeEdit{{streamFile, "if len(s.Subtitles) > 0 && !util.GlobalNoSubs {", "if !util.GlobalNoSubs {"}},
		kills: map[string][]string{sourcePkg: {"TestStreamPublishLeavesWhatTheStreamDoesNotSay", "TestStreamPublishFieldsAreIndependent"}},
	},
	{
		name:  "a subtitle's language and label swapped",
		edits: []codeEdit{{streamFile, "util.SubtitleInfo{URL: sub.URL, Language: sub.Language, Label: sub.Label}", "util.SubtitleInfo{URL: sub.URL, Language: sub.Label, Label: sub.Language}"}},
		kills: map[string][]string{sourcePkg: {"TestStreamPublishWritesEachField"}},
	},
	{
		name:  "only the first subtitle is kept",
		edits: []codeEdit{{streamFile, "for _, sub := range s.Subtitles {", "for _, sub := range s.Subtitles[:1] {"}},
		kills: map[string][]string{sourcePkg: {"TestStreamPublishKeepsEveryTrackInOrder"}},
	},

	// ---- fetchAndPublish ------------------------------------------------------
	{
		name:  "the resolved stream is never published",
		edits: []codeEdit{{providersFile, "\tstream.Publish()\n\treturn stream.URL, nil\n}", "\treturn stream.URL, nil\n}"}},
		kills: map[string][]string{providersPkg: {"TestFetchAndPublishResolvesThenPublishes", "TestPomfyFetchStreamURLPublishesLikeItAlwaysDid", "TestMigratedProvidersFetchStreamURLPublishes", "TestFetchStreamURLPrefersFetchStreamWhenTheSourceHasIt", "TestPomfyStreamPublishCascade", "TestStartFlixStreamPublishCascade"}},
		// the cascade must name the hop that carries the player's state
		stopsAt: "6 the player gets this stream's referer, audio and subtitles",
	},
	{
		name:  "a failed resolve still publishes what it returned",
		edits: []codeEdit{{providersFile, "\tstream, err := sf.FetchStream(ctx, episode, anime, quality)\n\tif err != nil {\n\t\treturn \"\", err\n\t}\n\tstream.Publish()", "\tstream, err := sf.FetchStream(ctx, episode, anime, quality)\n\tstream.Publish()\n\tif err != nil {\n\t\treturn \"\", err\n\t}"}},
		kills: map[string][]string{providersPkg: {"TestFetchAndPublishPublishesNothingOnFailure", "TestFetchStreamURLPropagatesAFailedResolve"}},
	},
	{
		name:  "a cancelled context still resolves",
		edits: []codeEdit{{providersFile, fetchAndPublishP + "\tif err := ctx.Err(); err != nil {\n\t\treturn \"\", err\n\t}\n", fetchAndPublishP}},
		kills: map[string][]string{providersPkg: {"TestFetchAndPublishDoesNothingOnACancelledContext"}},
	},
	{
		name:  "the last episode's subtitles survive into the next",
		edits: []codeEdit{{providersFile, "\tutil.ClearGlobalSubtitles()\n\tif anime.Source != \"\" {\n\t\tutil.SetGlobalAnimeSource(anime.Source)\n\t}\n\tstream, err := sf.FetchStream", "\tif anime.Source != \"\" {\n\t\tutil.SetGlobalAnimeSource(anime.Source)\n\t}\n\tstream, err := sf.FetchStream"}},
		kills: map[string][]string{providersPkg: {"TestFetchAndPublishClearsThePreviousStreamsSubtitlesBeforeResolving"}},
	},
	{
		name:  "the current source is never tagged",
		edits: []codeEdit{{providersFile, "\tif anime.Source != \"\" {\n\t\tutil.SetGlobalAnimeSource(anime.Source)\n\t}\n\tstream, err := sf.FetchStream", "\tstream, err := sf.FetchStream"}},
		kills: map[string][]string{providersPkg: {"TestFetchAndPublishTagsTheCurrentSourceBeforeResolving", "TestPomfyStreamPublishCascade", "TestStartFlixStreamPublishCascade"}},
	},
	{
		name:  "an anime with no source wipes the tagged one",
		edits: []codeEdit{{providersFile, "\tif anime.Source != \"\" {\n\t\tutil.SetGlobalAnimeSource(anime.Source)\n\t}\n\tstream, err := sf.FetchStream", "\tutil.SetGlobalAnimeSource(anime.Source)\n\tstream, err := sf.FetchStream"}},
		kills: map[string][]string{providersPkg: {"TestFetchAndPublishLeavesTheSourceAloneWhenTheAnimeHasNone"}},
	},
	{
		name:  "the URL is dropped on success",
		edits: []codeEdit{{providersFile, "\tstream.Publish()\n\treturn stream.URL, nil\n}", "\tstream.Publish()\n\treturn \"\", nil\n}"}},
		kills: map[string][]string{providersPkg: {"TestFetchAndPublishResolvesThenPublishes", "TestFetchStreamURLPrefersFetchStreamWhenTheSourceHasIt"}},
	},

	// ---- the dispatcher -------------------------------------------------------
	{
		name:  "the dispatcher ignores FetchStream",
		edits: []codeEdit{{dispatchFile, "\tif sf, ok := src.(source.StreamFetcher); ok {\n\t\treturn fetchAndPublish(ctx, sf, episode, anime, quality)\n\t}\n", ""}},
		kills: map[string][]string{providersPkg: {"TestFetchStreamURLPrefersFetchStreamWhenTheSourceHasIt"}},
	},
	{
		name:  "every source is sent down the FetchStream path",
		edits: []codeEdit{{dispatchFile, "\tif sf, ok := src.(source.StreamFetcher); ok {\n", "\tif sf, ok := src.(source.StreamFetcher); ok || true {\n"}},
		// a source that is not a StreamFetcher makes the assertion yield a nil
		// interface; calling it must not go unnoticed
		kills: map[string][]string{providersPkg: {"TestFetchStreamURLStillUsesTheLegacyPathForOtherSources"}},
	},
	{
		name:  "the last episode's referer and origin are not cleared",
		edits: []codeEdit{{dispatchFile, "\tutil.ClearGlobalReferer()\n\tutil.ClearGlobalOrigin()\n\tutil.Debug(\"Fetching stream via registry\"", "\tutil.Debug(\"Fetching stream via registry\""}},
		kills: map[string][]string{providersPkg: {"TestFetchStreamURLClearsTheLastEpisodesRefererAndOriginForLegacySources"}},
	},
	{
		name:  "the last episode's origin is not cleared",
		edits: []codeEdit{{dispatchFile, "\tutil.ClearGlobalReferer()\n\tutil.ClearGlobalOrigin()\n\tutil.Debug(\"Fetching stream via registry\"", "\tutil.ClearGlobalReferer()\n\tutil.Debug(\"Fetching stream via registry\""}},
		kills: map[string][]string{providersPkg: {"TestFetchStreamURLClearsTheLastEpisodesRefererAndOriginForLegacySources"}},
	},

	// ---- the Pomfy provider ---------------------------------------------------
	{
		name:  "the Pomfy resolver is reached with a cancelled context",
		edits: []codeEdit{{providersFile, "(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (source.Stream, error) {\n\tif err := ctx.Err(); err != nil {\n\t\treturn source.Stream{}, err\n\t}\n\treturn pomfyStreamFn", "(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (source.Stream, error) {\n\treturn pomfyStreamFn"}},
		kills: map[string][]string{providersPkg: {"TestPomfyFetchStreamPropagatesErrorsAndCancellation", "TestMigratedProvidersFetchStreamPropagatesErrorsAndCancellation"}},
	},
	{
		name:  "Pomfy's FetchStreamURL stops publishing",
		edits: []codeEdit{{providersFile, pomfyFetchStreamURLSig + "\treturn fetchAndPublish(ctx, p, episode, anime, quality)\n}", pomfyFetchStreamURLSig + "\ts, err := p.FetchStream(ctx, episode, anime, quality)\n\treturn s.URL, err\n}"}},
		kills: map[string][]string{providersPkg: {"TestPomfyFetchStreamURLPublishesLikeItAlwaysDid", "TestMigratedProvidersFetchStreamURLPublishes"}},
	},
	{
		name:  "Pomfy's FetchStream publishes behind the dispatcher's back",
		edits: []codeEdit{{providersFile, "\treturn pomfyStreamFn(anime, episode, quality)\n}", "\ts, err := pomfyStreamFn(anime, episode, quality)\n\ts.Publish()\n\treturn s, err\n}"}},
		kills: map[string][]string{providersPkg: {"TestPomfyFetchStreamWritesNoPlaybackState", "TestMigratedProvidersFetchStreamWritesNoPlaybackState"}},
	},

	// ---- api: ResolvePomfyStream / GetPomfyStreamURL --------------------------
	{
		name:  "GetPomfyStreamURL stops publishing",
		edits: []codeEdit{{pomfyAPIFile, "\tstream.Publish()\n\treturn stream.URL, nil\n}", "\treturn stream.URL, nil\n}"}},
		kills: map[string][]string{apiPkg: {"TestGetPomfyStreamURLLeavesTheMetadataContract", "TestResolvePomfyStreamThenPublishEqualsGetPomfyStreamURL"}},
	},
	{
		name:  "ResolvePomfyStream publishes as it resolves",
		edits: []codeEdit{{pomfyAPIFile, "\tutil.Debug(\"Pomfy stream URL obtained\", \"host\", stream.Host, \"url\", stream.URL[:min(len(stream.URL), 80)])\n\treturn out, nil", "\tutil.Debug(\"Pomfy stream URL obtained\", \"host\", stream.Host, \"url\", stream.URL[:min(len(stream.URL), 80)])\n\tout.Publish()\n\treturn out, nil"}},
		kills: map[string][]string{apiPkg: {"TestResolvePomfyStreamReturnsTheContextAndPublishesNothing"}},
	},
	{
		name:  "the referer is lost on the way into the Stream",
		edits: []codeEdit{{pomfyAPIFile, "\t\tReferer: stream.Referer,\n", ""}},
		kills: map[string][]string{apiPkg: {"TestResolvePomfyStreamReturnsTheContextAndPublishesNothing", "TestGetPomfyStreamURLLeavesTheMetadataContract"}},
	},
	{
		name:  "the audio hint is lost on the way into the Stream",
		edits: []codeEdit{{pomfyAPIFile, "\t\tAudioLang: pomfyStreamLang(stream.URL),\n", ""}},
		kills: map[string][]string{apiPkg: {"TestResolvePomfyStreamReturnsTheContextAndPublishesNothing", "TestGetPomfyStreamURLLeavesTheMetadataContract"}},
	},
	{
		name:  "a subtitle's language and label swapped on the way in",
		edits: []codeEdit{{pomfyAPIFile, "source.Subtitle{URL: s.URL, Language: s.Language, Label: s.Label}", "source.Subtitle{URL: s.URL, Language: s.Label, Label: s.Language}"}},
		kills: map[string][]string{apiPkg: {"TestResolvePomfyStreamReturnsTheContextAndPublishesNothing", "TestGetPomfyStreamURLLeavesTheMetadataContract"}},
	},
	{
		name:  "the subtitles are dropped on the way in",
		edits: []codeEdit{{pomfyAPIFile, "\tfor _, s := range stream.Subtitles {\n\t\tout.Subtitles = append(out.Subtitles, source.Subtitle{URL: s.URL, Language: s.Language, Label: s.Label})\n\t}\n", ""}},
		kills: map[string][]string{apiPkg: {"TestResolvePomfyStreamReturnsTheContextAndPublishesNothing", "TestGetPomfyStreamURLLeavesTheMetadataContract"}},
	},
	{
		name:  "a missing episode is let through",
		edits: []codeEdit{{pomfyAPIFile, "\tif media == nil || episode == nil {\n\t\treturn source.Stream{}, errors.New(\"no Pomfy title or episode to play\")", "\tif media == nil && episode == nil {\n\t\treturn source.Stream{}, errors.New(\"no Pomfy title or episode to play\")"}},
		kills: map[string][]string{apiPkg: {"TestResolvePomfyStreamRejectsAMissingTitleOrEpisode"}},
	},

	// ---- the StartFlix provider -------------------------------------------------
	{
		name:  "the StartFlix resolver is reached with a cancelled context",
		edits: []codeEdit{{providersFile, "func (p *startFlixProvider) FetchStream(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (source.Stream, error) {\n\tif err := ctx.Err(); err != nil {\n\t\treturn source.Stream{}, err\n\t}\n", "func (p *startFlixProvider) FetchStream(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (source.Stream, error) {\n"}},
		kills: map[string][]string{providersPkg: {"TestMigratedProvidersFetchStreamPropagatesErrorsAndCancellation"}},
	},
	{
		name:  "StartFlix's FetchStreamURL stops publishing",
		edits: []codeEdit{{providersFile, startFlixFetchStreamURLSig + "\treturn fetchAndPublish(ctx, p, episode, anime, quality)\n}", startFlixFetchStreamURLSig + "\ts, err := p.FetchStream(ctx, episode, anime, quality)\n\treturn s.URL, err\n}"}},
		kills: map[string][]string{providersPkg: {"TestMigratedProvidersFetchStreamURLPublishes", "TestStartFlixProvider_FetchStreamURL"}},
	},
	{
		name:  "StartFlix's FetchStream publishes behind the dispatcher's back",
		edits: []codeEdit{{providersFile, "\treturn startFlixStreamFn(anime, episode, quality)\n}", "\ts, err := startFlixStreamFn(anime, episode, quality)\n\ts.Publish()\n\treturn s, err\n}"}},
		kills: map[string][]string{providersPkg: {"TestMigratedProvidersFetchStreamWritesNoPlaybackState"}},
	},

	// ---- api: ResolveStartFlixStream / GetStartFlixStreamURL --------------------
	{
		name:  "GetStartFlixStreamURL stops publishing",
		edits: []codeEdit{{sfxAPIFile, "\tstream.Publish()\n\treturn stream.URL, nil\n}", "\treturn stream.URL, nil\n}"}},
		kills: map[string][]string{apiPkg: {"TestGetStartFlixStreamURL_SetsPlaybackGlobals", "TestResolveStartFlixStreamThenPublishEqualsGetStartFlixStreamURL"}},
	},
	{
		name:  "ResolveStartFlixStream publishes as it resolves",
		edits: []codeEdit{{sfxAPIFile, "\tutil.Debug(\"StartFlix stream URL obtained\", \"host\", stream.Host, \"url\", stream.URL[:min(len(stream.URL), 80)])\n\treturn out, nil", "\tutil.Debug(\"StartFlix stream URL obtained\", \"host\", stream.Host, \"url\", stream.URL[:min(len(stream.URL), 80)])\n\tout.Publish()\n\treturn out, nil"}},
		kills: map[string][]string{apiPkg: {"TestResolveStartFlixStreamReturnsTheContextAndPublishesNothing"}},
	},
	{
		name:  "the StartFlix referer is lost on the way into the Stream",
		edits: []codeEdit{{sfxAPIFile, "out := source.Stream{URL: stream.URL, Referer: stream.Referer}", "out := source.Stream{URL: stream.URL}"}},
		kills: map[string][]string{apiPkg: {"TestResolveStartFlixStreamReturnsTheContextAndPublishesNothing", "TestGetStartFlixStreamURL_SetsPlaybackGlobals"}},
	},
	{
		name:  "a pinned --audio is overridden by the stream's audio",
		edits: []codeEdit{{sfxAPIFile, "\tif startFlixPinnedAudioLang() == \"\" {\n\t\tout.AudioLang = audioLangForStartFlix(have)\n\t}\n", "\tout.AudioLang = audioLangForStartFlix(have)\n"}},
		kills: map[string][]string{apiPkg: {"TestResolveStartFlixStreamKeepsTheAudioTheUserPinned"}},
	},
	{
		name:  "mpv is never told the StartFlix audio",
		edits: []codeEdit{{sfxAPIFile, "\tif startFlixPinnedAudioLang() == \"\" {\n\t\tout.AudioLang = audioLangForStartFlix(have)\n\t}\n", ""}},
		kills: map[string][]string{apiPkg: {"TestResolveStartFlixStreamAlwaysTellsMpvTheAudioWhenNothingIsPinned", "TestResolveStartFlixStreamReturnsTheContextAndPublishesNothing"}},
	},
	{
		name:  "the StartFlix subtitles are dropped on the way in",
		edits: []codeEdit{{sfxAPIFile, "\tfor _, s := range stream.Subtitles {\n\t\tout.Subtitles = append(out.Subtitles, source.Subtitle{URL: s.URL, Language: s.Language, Label: s.Label})\n\t}\n", ""}},
		kills: map[string][]string{apiPkg: {"TestResolveStartFlixStreamReturnsTheContextAndPublishesNothing", "TestGetStartFlixStreamURL_SetsPlaybackGlobals"}},
	},
	{
		name:  "a StartFlix subtitle's language and label swapped",
		edits: []codeEdit{{sfxAPIFile, "source.Subtitle{URL: s.URL, Language: s.Language, Label: s.Label}", "source.Subtitle{URL: s.URL, Language: s.Label, Label: s.Language}"}},
		kills: map[string][]string{apiPkg: {"TestResolveStartFlixStreamReturnsTheContextAndPublishesNothing"}},
	},
	{
		name:  "a missing StartFlix episode is let through",
		edits: []codeEdit{{sfxAPIFile, "if media == nil || episode == nil {\n\t\treturn source.Stream{}, fmt.Errorf(\"no %s title or episode to play\"", "if media == nil && episode == nil {\n\t\treturn source.Stream{}, fmt.Errorf(\"no %s title or episode to play\""}},
		kills: map[string][]string{apiPkg: {"TestResolveStartFlixStreamRejectsAMissingTitleOrEpisode"}},
	},

	// ---- the TopCine provider ---------------------------------------------------
	{
		name:  "the TopCine resolver is reached with a cancelled context",
		edits: []codeEdit{{providersFile, "func (p *topCineProvider) FetchStream(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (source.Stream, error) {\n\tif err := ctx.Err(); err != nil {\n\t\treturn source.Stream{}, err\n\t}\n", "func (p *topCineProvider) FetchStream(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (source.Stream, error) {\n"}},
		kills: map[string][]string{providersPkg: {"TestMigratedProvidersFetchStreamPropagatesErrorsAndCancellation"}},
	},
	{
		name:  "TopCine's FetchStreamURL stops publishing",
		edits: []codeEdit{{providersFile, topCineFetchStreamURLSig + "\treturn fetchAndPublish(ctx, p, episode, anime, quality)\n}", topCineFetchStreamURLSig + "\ts, err := p.FetchStream(ctx, episode, anime, quality)\n\treturn s.URL, err\n}"}},
		kills: map[string][]string{providersPkg: {"TestMigratedProvidersFetchStreamURLPublishes", "TestTopCineProvider_FetchStreamURL", "TestTopCineStreamPublishCascade"}},
	},
	{
		name:  "TopCine's FetchStream publishes behind the dispatcher's back",
		edits: []codeEdit{{providersFile, "\treturn topCineStreamFn(anime, episode, quality)\n}", "\ts, err := topCineStreamFn(anime, episode, quality)\n\ts.Publish()\n\treturn s, err\n}"}},
		kills: map[string][]string{providersPkg: {"TestMigratedProvidersFetchStreamWritesNoPlaybackState"}},
	},
	{
		name:  "TopCine resolves nothing",
		edits: []codeEdit{{"internal/api/topcine.go", "\treturn ResolveStartFlixStream(media, episode, quality)\n}", "\treturn source.Stream{}, nil\n}"}},
		kills: map[string][]string{apiPkg: {"TestResolveTopCineStreamIsTheStartFlixStreamWithoutSideEffects", "TestResolveTopCineStreamRejectsAMissingTitleOrEpisode"}},
	},
	{
		name:  "GetTopCineStreamURL stops publishing",
		edits: []codeEdit{{"internal/api/topcine.go", "\treturn GetStartFlixStreamURL(media, episode, quality)\n}", "\ts, err := ResolveStartFlixStream(media, episode, quality)\n\treturn s.URL, err\n}"}},
		kills: map[string][]string{apiPkg: {"TestGetTopCineStreamURLStillPublishes"}},
	},

	// ---- fetchAndPublish clears the slate on any entrance ------------------------
	{
		name:  "the last stream's referer survives a direct provider call",
		edits: []codeEdit{{providersFile, "\tutil.ClearGlobalReferer()\n\tutil.ClearGlobalOrigin()\n\tutil.ClearGlobalSubtitles()\n\tif anime.Source != \"\" {", "\tutil.ClearGlobalOrigin()\n\tutil.ClearGlobalSubtitles()\n\tif anime.Source != \"\" {"}},
		kills: map[string][]string{providersPkg: {"TestFetchAndPublishClearsThePreviousStreamsRefererAndOriginOnAnyEntrance"}},
	},
	{
		name:  "the last stream's origin survives a direct provider call",
		edits: []codeEdit{{providersFile, "\tutil.ClearGlobalReferer()\n\tutil.ClearGlobalOrigin()\n\tutil.ClearGlobalSubtitles()\n\tif anime.Source != \"\" {", "\tutil.ClearGlobalReferer()\n\tutil.ClearGlobalSubtitles()\n\tif anime.Source != \"\" {"}},
		kills: map[string][]string{providersPkg: {"TestFetchAndPublishClearsThePreviousStreamsRefererAndOriginOnAnyEntrance"}},
	},

	// ---- the Pobreflix provider ----------------------------------------------------
	{
		name:  "the Pobreflix resolver is reached with a cancelled context",
		edits: []codeEdit{{providersFile, "func (p *pobreflixProvider) FetchStream(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (source.Stream, error) {\n\tif err := ctx.Err(); err != nil {\n\t\treturn source.Stream{}, err\n\t}\n", "func (p *pobreflixProvider) FetchStream(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (source.Stream, error) {\n"}},
		kills: map[string][]string{providersPkg: {"TestMigratedProvidersFetchStreamPropagatesErrorsAndCancellation"}},
	},
	{
		name:  "Pobreflix's FetchStreamURL stops publishing",
		edits: []codeEdit{{providersFile, pobreflixFetchStreamURLSig + "\treturn fetchAndPublish(ctx, p, episode, anime, quality)\n}", pobreflixFetchStreamURLSig + "\ts, err := p.FetchStream(ctx, episode, anime, quality)\n\treturn s.URL, err\n}"}},
		kills: map[string][]string{providersPkg: {"TestMigratedProvidersFetchStreamURLPublishes", "TestPobreflixStreamPublishCascade"}},
	},
	{
		name:  "Pobreflix's FetchStream publishes behind the dispatcher's back",
		edits: []codeEdit{{providersFile, "\treturn pobreflixStreamFn(anime, episode, quality)\n}", "\ts, err := pobreflixStreamFn(anime, episode, quality)\n\ts.Publish()\n\treturn s, err\n}"}},
		kills: map[string][]string{providersPkg: {"TestMigratedProvidersFetchStreamWritesNoPlaybackState"}},
	},

	// ---- api: ResolvePobreflixStream / GetPobreflixStreamURL ---------------------
	{
		name:  "GetPobreflixStreamURL stops publishing",
		edits: []codeEdit{{pbAPIFilePath, "\tutil.ClearGlobalReferer()\n\tutil.ClearGlobalSubtitles()\n\tstream.Publish()\n\treturn stream.URL, nil", "\tutil.ClearGlobalReferer()\n\tutil.ClearGlobalSubtitles()\n\treturn stream.URL, nil"}},
		kills: map[string][]string{apiPkg: {"TestGetPobreflixStreamURLDirectMovieClearsTheLastStreamsState", "TestPobreflixGatedDubPlaysThroughCfsolve", "TestResolvePobreflixStreamThenPublishEqualsGetPobreflixStreamURL"}},
	},
	{
		name:  "GetPobreflixStreamURL leaves the last stream's referer",
		edits: []codeEdit{{pbAPIFilePath, "\tutil.ClearGlobalReferer()\n\tutil.ClearGlobalSubtitles()\n\tstream.Publish()", "\tutil.ClearGlobalSubtitles()\n\tstream.Publish()"}},
		kills: map[string][]string{apiPkg: {"TestGetPobreflixStreamURLDirectMovieClearsTheLastStreamsState"}},
	},
	{
		name:  "GetPobreflixStreamURL leaves the last stream's subtitles",
		edits: []codeEdit{{pbAPIFilePath, "\tutil.ClearGlobalReferer()\n\tutil.ClearGlobalSubtitles()\n\tstream.Publish()", "\tutil.ClearGlobalReferer()\n\tstream.Publish()"}},
		kills: map[string][]string{apiPkg: {"TestGetPobreflixStreamURLDirectMovieClearsTheLastStreamsState"}},
	},
	{
		name:  "ResolvePobreflixStream publishes as it resolves",
		edits: []codeEdit{{pbAPIFilePath, "\t\tif err == nil && stream.URL != \"\" {\n\t\t\treturn stream, nil", "\t\tif err == nil && stream.URL != \"\" {\n\t\t\tstream.Publish()\n\t\t\treturn stream, nil"}},
		kills: map[string][]string{apiPkg: {"TestResolvePobreflixStreamReturnsTheContextAndPublishesNothing"}},
	},
	{
		name:  "the panel fallback is dropped when the direct path fails",
		edits: []codeEdit{{pbAPIFilePath, "\treturn ResolveStartFlixStream(media, episode, quality)\n}", "\treturn source.Stream{}, nil\n}"}},
		kills: map[string][]string{apiPkg: {"TestPobreflixDeadPanelPlaysThroughVIDSRC"}},
	},
	{
		name:  "a Pobreflix file is not told to mpv as the dub",
		edits: []codeEdit{{pbAPIFilePath, "\t\ts.AudioLang = audioLangForStartFlix(startflixAudio(audio))\n", "\t\ts.AudioLang = audioLangForStartFlix(startflix.AudioSubtitled)\n"}},
		kills: map[string][]string{apiPkg: {"TestPobreflixStreamBuildsTheContext", "TestResolvePobreflixStreamReturnsTheContextAndPublishesNothing"}},
	},
	{
		name:  "a pinned --audio is overridden by the Pobreflix file",
		edits: []codeEdit{{pbAPIFilePath, "\tif startFlixPinnedAudioLang() == \"\" {\n\t\ts.AudioLang = audioLangForStartFlix(startflixAudio(audio))\n\t}\n", "\ts.AudioLang = audioLangForStartFlix(startflixAudio(audio))\n"}},
		kills: map[string][]string{apiPkg: {"TestPobreflixStreamBuildsTheContext", "TestResolvePobreflixStreamKeepsTheAudioTheUserPinned", "TestGetPobreflixStreamURLDirectMovieKeepsTheAudioTheUserPinned"}},
	},
	{
		name:  "mpv is never told the Pobreflix audio",
		edits: []codeEdit{{pbAPIFilePath, "\tif startFlixPinnedAudioLang() == \"\" {\n\t\ts.AudioLang = audioLangForStartFlix(startflixAudio(audio))\n\t}\n", ""}},
		kills: map[string][]string{apiPkg: {"TestPobreflixStreamBuildsTheContext", "TestGetPobreflixStreamURLDirectMovieClearsTheLastStreamsState"}},
	},
	{
		name:  "the Pobreflix gate's referer is lost",
		edits: []codeEdit{{pbAPIFilePath, "return pobreflixStream(stream.URL, have, stream.Referer), nil", "return pobreflixStream(stream.URL, have, \"\"), nil"}},
		kills: map[string][]string{apiPkg: {"TestPobreflixGatedDubPlaysThroughCfsolve"}},
	},
	{
		name:  "the Pobreflix referer is dropped from the Stream",
		edits: []codeEdit{{pbAPIFilePath, "s := source.Stream{URL: url, Referer: referer}", "s := source.Stream{URL: url}"}},
		kills: map[string][]string{apiPkg: {"TestPobreflixStreamBuildsTheContext", "TestPobreflixGatedDubPlaysThroughCfsolve"}},
	},
	{
		name:  "a missing Pobreflix episode is let through",
		edits: []codeEdit{{pbAPIFilePath, "if media == nil || episode == nil {\n\t\treturn source.Stream{}, fmt.Errorf(\"no Pobreflix title or episode to play\")", "if media == nil && episode == nil {\n\t\treturn source.Stream{}, fmt.Errorf(\"no Pobreflix title or episode to play\")"}},
		kills: map[string][]string{apiPkg: {"TestResolvePobreflixStreamRejectsAMissingTitleOrEpisode"}},
	},
}

func TestStreamPublishGuardKillsMutants(t *testing.T) {
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go toolchain on PATH to build the mutants with")
	}
	root := moduleRoot(t)

	// Control: the guard tests run and pass on the real code. Without it a
	// test that always fails — or one renamed away, which -run would select
	// nothing of — would "kill" every mutant.
	guards := map[string][]string{}
	for pkg, tests := range publishGuards {
		guards[pkg] = append(guards[pkg], tests...)
	}
	for _, m := range streamPublishMutants {
		for pkg, tests := range m.kills {
			guards[pkg] = append(guards[pkg], tests...)
		}
	}
	for pkg, tests := range guards {
		out, err := runGuardTests(t, goBin, root, "", pkg, tests)
		require.NoError(t, err, "the guard tests must pass on the real code:\n%s", out)
		for _, name := range tests {
			require.Contains(t, out, "--- PASS: "+name+" (", "guard %s did not run and pass in %s", name, pkg)
		}
	}

	for _, m := range streamPublishMutants {
		t.Run(m.name, func(t *testing.T) {
			t.Parallel()
			overlay := writeMutantOverlay(t, root, m.edits)
			for pkg, tests := range m.kills {
				out, err := runGuardTests(t, goBin, root, overlay, pkg, tests)
				require.NotContains(t, out, "[build failed]", "the mutant must build, so the tests — not the compiler — catch it:\n%s", out)
				require.Error(t, err, "mutant survived %s: the guard tests passed against it\n%s", pkg, out)
				killed := false
				for _, name := range tests {
					killed = killed || strings.Contains(out, "--- FAIL: "+name+" (")
				}
				assert.True(t, killed, "%s failed, but not in a guard test %v:\n%s", pkg, tests, out)
				if m.stopsAt != "" {
					assert.Contains(t, out, `cascade stopped: stage "`+m.stopsAt+`" failed`)
					broken, err := strconv.Atoi(strings.Fields(m.stopsAt)[0])
					require.NoError(t, err, "a cascade stage name starts with its number")
					for _, match := range cascadeStageRe.FindAllStringSubmatch(out, -1) {
						ran, _ := strconv.Atoi(match[1])
						assert.LessOrEqual(t, ran, broken, "stage %d ran after the broken stage %d", ran, broken)
					}
				}
			}
		})
	}
}
