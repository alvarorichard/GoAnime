package api

import "testing"

// Tests of the tests, for the AnimesDigital guards of issue #229: the player
// must be sent only what it can play, and sent with the Referer the video host
// checks. Each mutant breaks one of those, builds, and has to be failed by a
// guard test it names. Same method and runner as stream_publish_meta_test.go.

const (
	animesDigitalPkg    = "./internal/scraper/providers/animesdigital"
	animesDigitalClient = "internal/scraper/providers/animesdigital/client.go"
	animesDigitalAPI    = "internal/api/animesdigital.go"
	scraperManagerFile  = "internal/scraper/manager.go"
	scraperPkg          = "./internal/scraper"
)

var animesDigitalGuards = map[string][]string{
	animesDigitalPkg: {
		"TestStreamFallsBackWhenOnlyThePiecesAreRefused", "TestEpisodeWithRefusedPiecesAndNoPlayerTwoIsOffline",
		"TestPlaylistAnswersFindsThePieceTheWayThePlayerDoes", "TestPlaylistAnswersSaysWhichPieceFailed",
		"TestPlaylistAnswersAsksTheWayThePlayerWill", "TestStreamRefererIsThePlayersForPlayerOneOnly",
		"TestStreamIsPlayerOneWhenItsPlaylistAnswers", "TestStreamFallsBackToPlayerTwoWhenPlayerOneDoesNotAnswer",
		// Issue #237: the players the site changed to on 2026-11-18.
		"TestStreamReadsPlayerOnesPageWhenItIsACDNFrame", "TestStreamFallsToPlayerTwoWhenPlayerOnesPageIsRefused",
		"TestStreamFallsToPlayerTwoWhenPlayerOnesPageNamesNoPlaylist", "TestEpisodeWhosePlayerOnesPageIsRefusedAndHasNoPlayerTwoIsOffline",
		"TestPlayerTwoReadsTheOlderAddress", "TestPlayerTwoRefusesAnOlderAddressThatIsWrong", "TestTheLegacyPlayerOneAddressStillWorks",
	},
	providersPkg: {
		"TestAnimesDigitalStreamCarriesThePlayersReferer",
		"TestAdapterProvidersFetchStreamReturnsTheContextAndWritesNone",
	},
	scraperPkg: {"TestAnimesDigitalAdapterTellsThePlayerTheReferer"},
	apiPkg:     {"TestGetAnimesDigitalEpisodesGoesBackWhenTheAudioPickerIsLeft"},
}

var animesDigitalMutants = []mutant{
	{
		name:  "player 1's /cdn/ page is not read",
		edits: []codeEdit{{animesDigitalClient, "} else if isCDNPlayer(u) {", "} else if false && isCDNPlayer(u) {"}},
		kills: map[string][]string{animesDigitalPkg: {"TestStreamReadsPlayerOnesPageWhenItIsACDNFrame"}},
	},
	{
		name:  "player 1's page is asked for with the site's address, not the episode's",
		edits: []codeEdit{{animesDigitalClient, "req.Header.Set(\"Referer\", episodeURL)", "req.Header.Set(\"Referer\", PlayerReferer)"}},
		kills: map[string][]string{animesDigitalPkg: {"TestStreamReadsPlayerOnesPageWhenItIsACDNFrame"}},
	},
	{
		name:  "the page's escaped slashes are left as they are",
		edits: []codeEdit{{animesDigitalClient, "strings.TrimSpace(strings.ReplaceAll(string(m[1]), `\\/`, \"/\"))", "strings.TrimSpace(string(m[1]))"}},
		kills: map[string][]string{animesDigitalPkg: {"TestStreamReadsPlayerOnesPageWhenItIsACDNFrame"}},
	},
	{
		name:  "any file the page names is taken for the playlist",
		edits: []codeEdit{{animesDigitalClient, "strings.HasPrefix(file, \"https://\") && isPlaylistPath(file)", "file != \"\""}},
		kills: map[string][]string{animesDigitalPkg: {"TestStreamFallsToPlayerTwoWhenPlayerOnesPageNamesNoPlaylist"}},
	},
	{
		name:  "player 2's ROT13-scrambled post is not unscrambled",
		edits: []codeEdit{{animesDigitalClient, "[]string{segment, rot13(segment)}", "[]string{segment}"}},
		kills: map[string][]string{animesDigitalPkg: {"TestPlayerTwoReadsTheOlderAddress"}},
	},
	{
		name:  "player 2's older four-part address is refused",
		edits: []codeEdit{{animesDigitalClient, "(len(parts) != 3 && len(parts) != 4)", "len(parts) != 3"}},
		kills: map[string][]string{animesDigitalPkg: {"TestPlayerTwoReadsTheOlderAddress"}},
	},
	{
		name:  "player 2's video is picked by the constant segment, not the index",
		edits: []codeEdit{{animesDigitalClient, "strconv.Atoi(parts[len(parts)-2])", "strconv.Atoi(parts[1])"}},
		kills: map[string][]string{animesDigitalPkg: {"TestPlayerTwoReadsTheOlderAddress"}},
	},
	{
		name:  "the playlist is trusted without asking for its first piece",
		edits: []codeEdit{{animesDigitalClient, "\treturn c.pieceAnswers(ctx, piece)\n", "\t_ = piece\n\treturn nil\n"}},
		kills: map[string][]string{animesDigitalPkg: {"TestStreamFallsBackWhenOnlyThePiecesAreRefused", "TestPlaylistAnswersSaysWhichPieceFailed"}},
	},
	{
		name:  "a piece that answers 403 counts as answering",
		edits: []codeEdit{{animesDigitalClient, "resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent", "resp.StatusCode >= 500"}},
		kills: map[string][]string{animesDigitalPkg: {"TestStreamFallsBackWhenOnlyThePiecesAreRefused"}},
	},
	{
		name:  "the probe asks with the site's Referer, not the player's",
		edits: []codeEdit{{animesDigitalClient, "req.Header.Set(\"Referer\", PlayerReferer)\n\treq.Header.Set(\"Accept\", accept)", "req.Header.Set(\"Referer\", c.baseURL+\"/\")\n\treq.Header.Set(\"Accept\", accept)"}},
		kills: map[string][]string{animesDigitalPkg: {"TestPlaylistAnswersAsksTheWayThePlayerWill", "TestStreamIsPlayerOneWhenItsPlaylistAnswers"}},
	},
	{
		name:  "a playlist that names no video is taken for one",
		edits: []codeEdit{{animesDigitalClient, "return \"\", errors.New(\"the playlist lists no video\")", "return target, nil"}},
		kills: map[string][]string{animesDigitalPkg: {"TestPlaylistAnswersFindsThePieceTheWayThePlayerDoes"}},
	},
	{
		name:  "a master playlist's variants are not followed",
		edits: []codeEdit{{animesDigitalClient, "if depth == 0 && isPlaylistPath(next) {", "if false && isPlaylistPath(next) {"}},
		kills: map[string][]string{animesDigitalPkg: {"TestPlaylistAnswersFindsThePieceTheWayThePlayerDoes"}},
	},
	{
		name:  "the player's Referer is another page",
		edits: []codeEdit{{animesDigitalClient, "const PlayerReferer = \"https://api.anivideo.net/\"", "const PlayerReferer = \"https://streameeeeee.site/\""}},
		kills: map[string][]string{animesDigitalPkg: {"TestStreamRefererIsThePlayersForPlayerOneOnly"}},
	},
	{
		name:  "player 1 is given no Referer",
		edits: []codeEdit{{animesDigitalClient, "\tif stream == \"\" || strings.HasPrefix(stream, bloggerVideoPrefix) {\n\t\treturn \"\"\n\t}\n\treturn PlayerReferer", "\treturn \"\""}},
		kills: map[string][]string{animesDigitalPkg: {"TestStreamRefererIsThePlayersForPlayerOneOnly"}, scraperPkg: {"TestAnimesDigitalAdapterTellsThePlayerTheReferer"}},
	},
	{
		name:  "player 2's Blogger video is given player 1's Referer",
		edits: []codeEdit{{animesDigitalClient, "\tif stream == \"\" || strings.HasPrefix(stream, bloggerVideoPrefix) {\n\t\treturn \"\"\n\t}\n\treturn PlayerReferer", "\tif stream == \"\" {\n\t\treturn \"\"\n\t}\n\treturn PlayerReferer"}},
		kills: map[string][]string{animesDigitalPkg: {"TestStreamRefererIsThePlayersForPlayerOneOnly"}},
	},
	{
		name:  "the adapter drops the Referer",
		edits: []codeEdit{{scraperManagerFile, "\t\tmetadata[\"referer\"] = referer\n", "\t\t_ = referer\n"}},
		kills: map[string][]string{scraperPkg: {"TestAnimesDigitalAdapterTellsThePlayerTheReferer"}, providersPkg: {"TestAnimesDigitalStreamCarriesThePlayersReferer"}},
	},
	{
		name:  "the provider drops the adapter's Referer",
		edits: []codeEdit{{providersFile, "AnimesDigital\")\n\t}\n\treturn streamFromMetadata(streamURL, metadata), nil", "AnimesDigital\")\n\t}\n\treturn source.Stream{URL: streamURL, Referer: metadata[\"\"]}, nil"}},
		kills: map[string][]string{providersPkg: {"TestAnimesDigitalStreamCarriesThePlayersReferer", "TestAdapterProvidersFetchStreamReturnsTheContextAndWritesNone"}},
	},
	{
		name:  "ESC on the audio question takes subtitles",
		edits: []codeEdit{{animesDigitalAPI, "errors.Is(err, tui.ErrPickBack) || errors.Is(err, tui.ErrPickCancelled)", "errors.Is(err, tui.ErrPickBack) && errors.Is(err, tui.ErrPickCancelled)"}},
		kills: map[string][]string{apiPkg: {"TestGetAnimesDigitalEpisodesGoesBackWhenTheAudioPickerIsLeft"}},
	},
}

func TestAnimesDigitalGuardKillsMutants(t *testing.T) {
	skipMutantsWhenShort(t)
	killMutants(t, animesDigitalGuards, animesDigitalMutants)
}
