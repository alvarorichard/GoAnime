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
	sourcePkg        = "./internal/api/source"
	providersPkg     = "./internal/api/providers"
	streamFile       = "internal/api/source/stream.go"
	providersFile    = "internal/api/providers/source_providers.go"
	dispatchFile     = "internal/api/providers/dispatch.go"
	pomfyAPIFile     = "internal/api/pomfy.go"
	fetchAndPublishP = "func fetchAndPublish(ctx context.Context, sf source.StreamFetcher, episode *models.Episode, anime *models.Anime, quality string) (string, error) {\n"
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
		kills: map[string][]string{providersPkg: {"TestFetchAndPublishResolvesThenPublishes", "TestPomfyFetchStreamURLPublishesLikeItAlwaysDid", "TestFetchStreamURLPrefersFetchStreamWhenTheSourceHasIt", "TestPomfyStreamPublishCascade"}},
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
		kills: map[string][]string{providersPkg: {"TestFetchAndPublishTagsTheCurrentSourceBeforeResolving", "TestPomfyStreamPublishCascade"}},
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
		kills: map[string][]string{providersPkg: {"TestFetchStreamURLClearsTheLastEpisodesRefererAndOrigin", "TestPomfyStreamPublishCascade"}},
		// the cascade must name the hop that checks the cleanup
		stopsAt: "4 the last episode's playback state is gone before the resolve",
	},
	{
		name:  "the last episode's origin is not cleared",
		edits: []codeEdit{{dispatchFile, "\tutil.ClearGlobalReferer()\n\tutil.ClearGlobalOrigin()\n\tutil.Debug(\"Fetching stream via registry\"", "\tutil.ClearGlobalReferer()\n\tutil.Debug(\"Fetching stream via registry\""}},
		kills: map[string][]string{providersPkg: {"TestFetchStreamURLClearsTheLastEpisodesRefererAndOrigin"}},
	},

	// ---- the Pomfy provider ---------------------------------------------------
	{
		name:  "the Pomfy resolver is reached with a cancelled context",
		edits: []codeEdit{{providersFile, "(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (source.Stream, error) {\n\tif err := ctx.Err(); err != nil {\n\t\treturn source.Stream{}, err\n\t}\n\treturn pomfyStreamFn", "(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (source.Stream, error) {\n\treturn pomfyStreamFn"}},
		kills: map[string][]string{providersPkg: {"TestPomfyFetchStreamPropagatesErrorsAndCancellation"}},
	},
	{
		name:  "Pomfy's FetchStreamURL stops publishing",
		edits: []codeEdit{{providersFile, "\treturn fetchAndPublish(ctx, p, episode, anime, quality)\n}", "\ts, err := p.FetchStream(ctx, episode, anime, quality)\n\treturn s.URL, err\n}"}},
		kills: map[string][]string{providersPkg: {"TestPomfyFetchStreamURLPublishesLikeItAlwaysDid"}},
	},
	{
		name:  "Pomfy's FetchStream publishes behind the dispatcher's back",
		edits: []codeEdit{{providersFile, "\treturn pomfyStreamFn(anime, episode, quality)\n}", "\ts, err := pomfyStreamFn(anime, episode, quality)\n\ts.Publish()\n\treturn s, err\n}"}},
		kills: map[string][]string{providersPkg: {"TestPomfyFetchStreamWritesNoPlaybackState"}},
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
