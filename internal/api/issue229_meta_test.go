package api

import "testing"

// Tests of the tests, for the guard on issue #229: the reason an episode did
// not load has to be on the episode list the user is sent back to, because
// text printed to the terminal first is hidden behind that list. Same method
// as stream_publish_meta_test.go; the runner is shared.

const (
	tuiPkg       = "./internal/tui"
	playbackPkg  = "./internal/playback"
	animefirePkg = "./internal/scraper/providers/animefire"
	pickerFile   = "internal/tui/picker.go"
	playbackFile = "internal/playback/common.go"
	animefireAPI = "internal/scraper/providers/animefire/api.go"
	animefireCli = "internal/scraper/providers/animefire/client.go"
	mpvLaunch    = "internal/player/mpv_launch.go"
	playerGo     = "internal/player/player.go"

	pickerPrepend   = "\tif !m.showHelp {\n\t\tif notice := renderNotice(&m.theme, m.notice, width); notice != \"\" {\n\t\t\tbody = notice + \"\\n\" + body\n\t\t}\n\t}\n"
	pickerLayoutH   = "\theight = max(height-m.noticeHeight(width), 1)\n\tm.areaHeight = height\n\tm.panelSide"
	pendingFn       = "func withPendingNotice(opts PickOptions) PickOptions {\n\tif pending := takePickerNotice(); opts.Notice == \"\" {\n\t\topts.Notice = pending\n\t}\n\treturn opts\n}"
	reportRemovedFn = "\tutil.Errorf(\"%s\", problem)\n\tutil.Infof(\"%s\", hint)\n\tnoticeOnNextPicker(problem + \" \" + hint)\n"
	reportLoadFn    = "\tutil.Warnf(\"Failed to extract video URL: %v\", err)\n\tnoticeOnNextPicker(episodeLoadFailedNotice(episodeNumberStr, err))\n"
	warnTail        = "\t// The episode list opens next and covers what was just printed.\n\tif c.notice != nil {\n\t\tc.notice(titleOfflineNotice(len(episodes)))\n\t}\n"
)

var issue229Mutants = []mutant{
	// ---- the picker draws the notice ---------------------------------------------
	{
		name:  "the picker never draws its notice",
		edits: []codeEdit{{pickerFile, pickerPrepend, ""}},
		kills: map[string][]string{tuiPkg: {"TestPickerShowsItsNoticeAboveTheList", "TestPickerNoticeSurvivesNavigationAndFiltering", "TestPickerNoticeWithAPanel", "TestPickerWithoutANoticeDrawsNothingExtra"}},
	},
	{
		name:  "the notice is drawn below the list",
		edits: []codeEdit{{pickerFile, "\t\t\tbody = notice + \"\\n\" + body\n", "\t\t\tbody = body + \"\\n\" + notice\n"}},
		kills: map[string][]string{tuiPkg: {"TestPickerShowsItsNoticeAboveTheList"}},
	},
	{
		name:  "the notice is drawn over the help overlay",
		edits: []codeEdit{{pickerFile, pickerPrepend, "\tif true {\n\t\tif notice := renderNotice(&m.theme, m.notice, width); notice != \"\" {\n\t\t\tbody = notice + \"\\n\" + body\n\t\t}\n\t}\n"}},
		kills: map[string][]string{tuiPkg: {"TestPickerHelpOverlayTakesTheWholeScreen"}},
	},
	{
		name:  "the list keeps its full height under a notice (layout)",
		edits: []codeEdit{{pickerFile, pickerLayoutH, "\tm.areaHeight = height\n\tm.panelSide"}},
		kills: map[string][]string{tuiPkg: {"TestPickerNoticeTakesExactlyItsRowsFromTheList", "TestPickerNoticeIsCutRatherThanFlooding", "TestPickerNoticeWithAPanel"}},
	},
	{
		name:  "the notice is never given to the picker",
		edits: []codeEdit{{pickerFile, "notice:        singleLine(opts.Notice),\n", ""}},
		kills: map[string][]string{tuiPkg: {"TestPickerShowsItsNoticeAboveTheList", "TestPickerNoticeIsOneLineWhateverItHolds"}},
	},
	{
		name:  "the notice is not flattened to one line",
		edits: []codeEdit{{pickerFile, "notice:        singleLine(opts.Notice),", "notice:        opts.Notice,"}},
		kills: map[string][]string{tuiPkg: {"TestPickerNoticeIsOneLineWhateverItHolds"}},
	},
	{
		name:  "the view draws in the full height, not the list's",
		edits: []codeEdit{{pickerFile, "\t\theight = m.areaHeight\n", "\t\theight = m.areaHeight + m.noticeHeight(width)\n"}},
		kills: map[string][]string{tuiPkg: {"TestPickerNoticeNeverCutsThePanel"}},
	},
	{
		name:  "a notice takes no room",
		edits: []codeEdit{{pickerFile, "func (m *pickerModel) noticeHeight(width int) int {\n\tif m.notice == \"\" {", "func (m *pickerModel) noticeHeight(width int) int {\n\tif true {"}},
		kills: map[string][]string{tuiPkg: {"TestPickerNoticeTakesExactlyItsRowsFromTheList", "TestPickerNoticeWithAPanel"}},
	},

	// ---- the pending notice --------------------------------------------------------
	{
		name:  "a notice is shown on every picker, not once",
		edits: []codeEdit{{pickerFile, "\ttext := pickerNotice.text\n\tpickerNotice.text = \"\"\n", "\ttext := pickerNotice.text\n"}},
		kills: map[string][]string{tuiPkg: {"TestNoticeOnNextPickerIsTakenOnce", "TestWithPendingNotice"}},
	},
	{
		name:  "the first notice wins over the latest",
		edits: []codeEdit{{pickerFile, "\tpickerNotice.text = singleLine(text)\n", "\tif pickerNotice.text == \"\" {\n\t\tpickerNotice.text = singleLine(text)\n\t}\n"}},
		kills: map[string][]string{tuiPkg: {"TestNoticeOnNextPickerKeepsOnlyTheLatest"}},
	},
	{
		name:  "the pending notice is not flattened",
		edits: []codeEdit{{pickerFile, "\tpickerNotice.text = singleLine(text)\n\tpickerNotice.Unlock()", "\tpickerNotice.text = text\n\tpickerNotice.Unlock()"}},
		kills: map[string][]string{tuiPkg: {"TestNoticeOnNextPickerCollapsesNewlines"}},
	},
	{
		name:  "the pending notice overrides a picker's own",
		edits: []codeEdit{{pickerFile, pendingFn, "func withPendingNotice(opts PickOptions) PickOptions {\n\tif pending := takePickerNotice(); true {\n\t\topts.Notice = pending\n\t}\n\treturn opts\n}"}},
		kills: map[string][]string{tuiPkg: {"TestWithPendingNotice"}},
	},
	{
		name:  "a picker with its own notice leaves the pending one for a later screen",
		edits: []codeEdit{{pickerFile, pendingFn, "func withPendingNotice(opts PickOptions) PickOptions {\n\tif opts.Notice == \"\" {\n\t\topts.Notice = takePickerNotice()\n\t}\n\treturn opts\n}"}},
		kills: map[string][]string{tuiPkg: {"TestWithPendingNotice"}},
	},
	{
		name:  "a picker with no notice of its own does not take the pending one",
		edits: []codeEdit{{pickerFile, pendingFn, "func withPendingNotice(opts PickOptions) PickOptions {\n\tif pending := takePickerNotice(); opts.Notice == \"\" && false {\n\t\topts.Notice = pending\n\t}\n\treturn opts\n}"}},
		kills: map[string][]string{tuiPkg: {"TestWithPendingNotice"}},
	},
	{
		name:  "PickActioned never takes the pending notice",
		edits: []codeEdit{{pickerFile, "\topts = withPendingNotice(opts)\n", ""}},
		kills: map[string][]string{tuiPkg: {"TestPickActionedTakesThePendingNoticeEvenWhenItCannotOpen", "TestPickActionedTakesThePendingNoticeEvenWithNoItems"}},
	},

	// ---- what the user is told -----------------------------------------------------
	{
		name:  "a removed video is not put on the list",
		edits: []codeEdit{{playbackFile, reportRemovedFn, "\tutil.Errorf(\"%s\", problem)\n\tutil.Infof(\"%s\", hint)\n"}},
		kills: map[string][]string{playbackPkg: {"TestReportSourceVideoRemovedPutsTheReasonOnTheList", "TestReportEpisodeVideoErrorPicksTheNoticeByCause", "TestPlayEpisodeExplainsARemovedVideoOnTheListItReturnsTo"}},
	},
	{
		name:  "the list notice leaves out how to go on",
		edits: []codeEdit{{playbackFile, "noticeOnNextPicker(problem + \" \" + hint)", "noticeOnNextPicker(problem)"}},
		kills: map[string][]string{playbackPkg: {"TestReportSourceVideoRemovedPutsTheReasonOnTheList"}},
	},
	{
		name:  "the list notice leaves out what happened",
		edits: []codeEdit{{playbackFile, "noticeOnNextPicker(problem + \" \" + hint)", "noticeOnNextPicker(hint)"}},
		kills: map[string][]string{playbackPkg: {"TestReportSourceVideoRemovedPutsTheReasonOnTheList"}},
	},
	{
		name:  "a load failure is not put on the list",
		edits: []codeEdit{{playbackFile, reportLoadFn, "\tutil.Warnf(\"Failed to extract video URL: %v\", err)\n"}},
		kills: map[string][]string{playbackPkg: {"TestReportEpisodeVideoErrorPicksTheNoticeByCause", "TestReportEpisodeLoadFailedNeverPanicsOnAMissingEpisodeLabel", "TestPlayEpisodeExplainsAnyOtherLoadFailureToo"}},
	},
	{
		name:  "a removed video is reported as any other failure",
		edits: []codeEdit{{playbackFile, "\tcase errors.Is(videoErr, player.ErrSourceVideoRemoved):", "\tcase false:"}},
		kills: map[string][]string{playbackPkg: {"TestReportEpisodeVideoErrorPicksTheNoticeByCause", "TestPlayEpisodeExplainsARemovedVideoOnTheListItReturnsTo"}},
	},
	{
		name:  "the user's own back is explained to them",
		edits: []codeEdit{{playbackFile, "\tcase errors.Is(videoErr, player.ErrBackToEpisodeSelection):\n\t\t// User-initiated: nothing to report.\n", ""}},
		kills: map[string][]string{playbackPkg: {"TestReportEpisodeVideoErrorPicksTheNoticeByCause", "TestPlayEpisodeSaysNothingWhenTheUserWentBack"}},
	},
	{
		name:  "PlayEpisode stops reporting why the episode did not load",
		edits: []codeEdit{{playbackFile, "\t\treportEpisodeVideoError(anime, episodeNumberStr, videoErr)\n", ""}},
		kills: map[string][]string{playbackPkg: {"TestPlayEpisodeExplainsARemovedVideoOnTheListItReturnsTo", "TestPlayEpisodeExplainsAnyOtherLoadFailureToo"}},
	},
	{
		name:  "a failed load no longer sends the user back to the list",
		edits: []codeEdit{{playbackFile, "\t\treportEpisodeVideoError(anime, episodeNumberStr, videoErr)\n\t\treturn player.ErrBackToEpisodeSelection\n", "\t\treportEpisodeVideoError(anime, episodeNumberStr, videoErr)\n\t\treturn videoErr\n"}},
		kills: map[string][]string{playbackPkg: {"TestPlayEpisodeExplainsARemovedVideoOnTheListItReturnsTo"}},
	},
	{
		name:  "the notice does not name the episode",
		edits: []codeEdit{{playbackFile, "return fmt.Sprintf(\"%s did not load: %s Pick another episode, or press ESC to go back.\", label, strings.TrimRight(reason, \". \")+\".\")", "return fmt.Sprintf(\"Did not load: %s Pick another episode, or press ESC to go back.\", strings.TrimRight(reason, \". \")+\".\")"}},
		kills: map[string][]string{playbackPkg: {"TestEpisodeLoadFailedNotice", "TestReportEpisodeVideoErrorPicksTheNoticeByCause"}},
	},
	{
		name:  "the notice doubles the cause's full stop",
		edits: []codeEdit{{playbackFile, "strings.TrimRight(reason, \". \")+\".\")", "reason+\".\")"}},
		kills: map[string][]string{playbackPkg: {"TestEpisodeLoadFailedNotice"}},
	},
	{
		name:  "an empty cause makes an empty notice",
		edits: []codeEdit{{playbackFile, "\tif reason == \"\" {\n\t\treason = \"no reason was given\"\n\t}\n", ""}},
		kills: map[string][]string{playbackPkg: {"TestEpisodeLoadFailedNotice"}},
	},
	{
		name:  "no error still makes a notice",
		edits: []codeEdit{{playbackFile, "func episodeLoadFailedNotice(episodeNumberStr string, err error) string {\n\tif err == nil {\n\t\treturn \"\"\n\t}\n", "func episodeLoadFailedNotice(episodeNumberStr string, err error) string {\n"}},
		kills: map[string][]string{playbackPkg: {"TestEpisodeLoadFailedNotice"}},
	},
	{
		name:  "the other sources are no longer named",
		edits: []codeEdit{{playbackFile, "hint = fmt.Sprintf(\"Press ESC to leave the episode list and pick the same title from another source: %s.\", strings.Join(others, \", \"))", "hint = \"Press ESC to leave the episode list and search again.\""}},
		kills: map[string][]string{playbackPkg: {"TestSourceVideoRemovedMessage", "TestReportSourceVideoRemovedPutsTheReasonOnTheList"}},
	},
	{
		name:  "a missing anime crashes the report",
		edits: []codeEdit{{playbackFile, "\tcurrent := \"\"\n\tif anime != nil {\n\t\tcurrent = anime.Source\n\t}\n\tsrc := current\n\tif src == \"\" {\n\t\tsrc = \"this source\"\n\t}\n\tlabel := strings.TrimSpace(episodeNumberStr)\n\tif label == \"\" {\n\t\tlabel = \"This episode\"\n\t}\n\tproblem = ", "\tcurrent := anime.Source\n\tsrc := current\n\tif src == \"\" {\n\t\tsrc = \"this source\"\n\t}\n\tlabel := strings.TrimSpace(episodeNumberStr)\n\tif label == \"\" {\n\t\tlabel = \"This episode\"\n\t}\n\tproblem = "}},
		kills: map[string][]string{playbackPkg: {"TestSourceVideoRemovedMessageFallsBackToWhatItKnows", "TestReportSourceVideoRemoved_HandlesMissingContext"}},
	},

	// ---- AnimeFire's own warning ---------------------------------------------------
	{
		name:  "AnimeFire's offline warning never reaches the list",
		edits: []codeEdit{{animefireAPI, warnTail, ""}},
		kills: map[string][]string{animefirePkg: {"TestWarnIfTitleLooksOffline_TellsTheListWhenBothProbesAreOffline", "TestGetAnimeEpisodes_TellsTheListWhenTheWholeTitleIsOffline"}},
	},
	{
		name:  "AnimeFire's offline warning crashes without a notifier",
		edits: []codeEdit{{animefireAPI, "\tif c.notice != nil {\n\t\tc.notice(titleOfflineNotice(len(episodes)))\n\t}\n", "\tc.notice(titleOfflineNotice(len(episodes)))\n"}},
		kills: map[string][]string{animefirePkg: {"TestWarnIfTitleLooksOffline_WorksWithoutANotifier"}},
	},
	{
		name:  "the offline warning forgets how many episodes are listed",
		edits: []codeEdit{{animefireAPI, "c.notice(titleOfflineNotice(len(episodes)))", "c.notice(titleOfflineNotice(len(episodes) - 1))"}},
		kills: map[string][]string{animefirePkg: {"TestWarnIfTitleLooksOffline_TellsTheListWhenBothProbesAreOffline", "TestGetAnimeEpisodes_TellsTheListWhenTheWholeTitleIsOffline"}},
	},
	{
		name:  "the offline notice does not say how to leave",
		edits: []codeEdit{{animefireAPI, "Press ESC and pick the same title from StartFlix or HiAnime.", "Try another source."}},
		kills: map[string][]string{animefirePkg: {"TestTitleOfflineNoticeCountsTheEpisodes"}},
	},
	{
		name:  "only the first episode is probed",
		edits: []codeEdit{{animefireAPI, "[]int{0, len(episodes) / 2}", "[]int{0}"}},
		kills: map[string][]string{animefirePkg: {"TestWarnIfTitleLooksOffline_StaysQuietWhenSomethingPlays", "TestWarnIfTitleLooksOffline_ProbesTheFirstAndTheMiddleEpisode"}},
	},
	{
		name:  "a title with a playable episode is condemned",
		edits: []codeEdit{{animefireAPI, "if _, err := c.streamAPI(episodes[idx].URL); !errors.Is(err, netx.ErrMediaOffline) {", "if _, err := c.streamAPI(episodes[idx].URL); errors.Is(err, netx.ErrMediaOffline) {"}},
		kills: map[string][]string{animefirePkg: {"TestWarnIfTitleLooksOffline_TellsTheListWhenBothProbesAreOffline", "TestWarnIfTitleLooksOffline_StaysQuietWhenSomethingPlays"}},
	},
	{
		name:  "a failing API condemns a title",
		edits: []codeEdit{{animefireAPI, "!errors.Is(err, netx.ErrMediaOffline) {\n\t\t\treturn // something is playable; nothing to warn about", "err == nil {\n\t\t\treturn // something is playable; nothing to warn about"}},
		kills: map[string][]string{animefirePkg: {"TestWarnIfTitleLooksOffline_IgnoresAnErrorThatIsNotOffline"}},
	},
	{
		name:  "a one-episode title is probed",
		edits: []codeEdit{{animefireAPI, "\tif len(episodes) < 2 {\n\t\treturn\n\t}\n\tfor _, idx := range", "\tif len(episodes) < 1 {\n\t\treturn\n\t}\n\tfor _, idx := range"}},
		kills: map[string][]string{animefirePkg: {"TestWarnIfTitleLooksOffline_NeedsAtLeastTwoEpisodes"}},
	},
	{
		name:  "listing the episodes no longer checks the title",
		edits: []codeEdit{{animefireAPI, "\tc.warnIfTitleLooksOffline(episodes)\n", ""}},
		kills: map[string][]string{animefirePkg: {"TestGetAnimeEpisodes_TellsTheListWhenTheWholeTitleIsOffline"}},
	},
	{
		name:  "the AnimeFire client has no real notifier",
		edits: []codeEdit{{animefireCli, "\t\tnotice:     util.NoticeOnNextPicker,\n", ""}},
		kills: map[string][]string{animefirePkg: {"TestClientsTellTheListThroughTheRealNotifierByDefault"}},
	},

	// ---- starting mpv: the Store alias that came first on PATH ----------------------
	{
		name:  "no option counts as tuning",
		edits: []codeEdit{{mpvLaunch, "for _, prefix := range tuningMPVArgs {", "for _, prefix := range []string{} {"}},
		kills: map[string][]string{playerPkg: {"TestEssentialMPVArgsKeepsWhatPlaybackNeeds", "TestStartVideoAcrossRetriesWithoutTheTuningOptionsAnMPVRejects"}},
	},
	{
		name:  "every option counts as tuning",
		edits: []codeEdit{{mpvLaunch, "\t\ttuning := false\n", "\t\ttuning := true\n"}},
		kills: map[string][]string{playerPkg: {"TestEssentialMPVArgsKeepsWhatPlaybackNeeds", "TestStartVideoAcrossRetriesWithoutTheTuningOptionsAnMPVRejects"}},
	},
	{
		name:  "the tuning is kept and the rest dropped",
		edits: []codeEdit{{mpvLaunch, "\t\tif tuning {\n\t\t\tdropped = append(dropped, a)\n\t\t} else {\n\t\t\tkept = append(kept, a)\n\t\t}", "\t\tif tuning {\n\t\t\tkept = append(kept, a)\n\t\t} else {\n\t\t\tdropped = append(dropped, a)\n\t\t}"}},
		kills: map[string][]string{playerPkg: {"TestEssentialMPVArgsKeepsWhatPlaybackNeeds", "TestEssentialMPVArgsEdges"}},
	},
	{
		name:  "an mpv that cannot be started ends the search",
		edits: []codeEdit{{mpvLaunch, "return errors.As(err, &early) || errors.As(err, &start)", "_ = &start\n\treturn errors.As(err, &early)"}},
		kills: map[string][]string{playerPkg: {"TestFailedToLaunch", "TestStartVideoAcrossFallsThroughWhenAnMPVCannotBeStartedAtAll"}},
	},
	{
		name:  "an mpv that went silent sends the search on",
		edits: []codeEdit{{mpvLaunch, "return errors.As(err, &early) || errors.As(err, &start)", "_, _ = &early, &start\n\treturn err != nil"}},
		kills: map[string][]string{playerPkg: {"TestFailedToLaunch", "TestStartVideoAcrossDoesNotWaitOnASecondMPVWhenTheFirstCameUpAndWentSilent"}},
	},
	{
		name:  "an mpv that exits is the end of the search",
		edits: []codeEdit{{mpvLaunch, "\t\tif !failedToLaunch(err) {\n\t\t\treturn \"\", err\n\t\t}\n\t\tfailed = append(failed, mpvAttempt{path: path, err: err})", "\t\tif true {\n\t\t\treturn \"\", err\n\t\t}\n\t\tfailed = append(failed, mpvAttempt{path: path, err: err})"}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossFallsThroughToTheNextMPVWhenThePathOneExits", "TestStartVideoAcrossFallsThroughWhenAnMPVCannotBeStartedAtAll"}},
	},
	{
		name:  "the search goes on after an mpv that went silent",
		edits: []codeEdit{{mpvLaunch, "\t\tif !failedToLaunch(err) {\n\t\t\treturn \"\", err\n\t\t}\n", ""}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossDoesNotWaitOnASecondMPVWhenTheFirstCameUpAndWentSilent"}},
	},
	{
		name:  "the mpv that worked is not remembered",
		edits: []codeEdit{{mpvLaunch, "\t\tif err == nil {\n\t\t\trememberWorkingMPV(path)\n\t\t\tif len(failed) > 0 {", "\t\tif err == nil {\n\t\t\tif len(failed) > 0 {"}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossRemembersTheMPVThatWorkedSoTheStubIsNotTriedAgain"}},
	},
	{
		name:  "the remembered mpv is not put first",
		edits: []codeEdit{{mpvLaunch, "\tfor _, p := range paths {\n\t\tif p == working {\n\t\t\tout = append(out, p)\n\t\t}\n\t}\n", ""}},
		kills: map[string][]string{playerPkg: {"TestOrderedMPVPutsTheOneThatWorkedFirst"}},
	},
	{
		name:  "the other mpv are dropped from the order",
		edits: []codeEdit{{mpvLaunch, "\tfor _, p := range paths {\n\t\tif p != working {\n\t\t\tout = append(out, p)\n\t\t}\n\t}\n\treturn out", "\treturn out"}},
		kills: map[string][]string{playerPkg: {"TestOrderedMPVPutsTheOneThatWorkedFirst", "TestStartVideoAcrossFallsThroughToTheNextMPVWhenThePathOneExits"}},
	},
	{
		name:  "an mpv known to need the minimal options is given the full ones",
		edits: []codeEdit{{mpvLaunch, "\t\tif mpvNeedsMinimal(path) {\n\t\t\tuseArgs, _ = essentialMPVArgs(args)\n\t\t}\n", ""}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossGoesStraightToTheMinimalOptionsForAnMPVThatNeedsThem"}},
	},
	{
		name:  "the minimal options are not remembered",
		edits: []codeEdit{{mpvLaunch, "\t\t\trememberMinimalMPV(path)\n\t\t\trememberWorkingMPV(path)", "\t\t\trememberWorkingMPV(path)"}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossGoesStraightToTheMinimalOptionsForAnMPVThatNeedsThem"}},
	},
	{
		name:  "an mpv that played without its tuning is not remembered as working",
		edits: []codeEdit{{mpvLaunch, "\t\t\trememberMinimalMPV(path)\n\t\t\trememberWorkingMPV(path)", "\t\t\trememberMinimalMPV(path)"}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossRemembersWhichMPVNeededTheMinimalOptions"}},
	},
	{
		name:  "the tuning is stripped after any exit, not only an init failure",
		edits: []codeEdit{{mpvLaunch, "if !errors.As(failed[i].err, &early) || early.code != 1 || mpvNeedsMinimal(failed[i].path) {", "if !errors.As(failed[i].err, &early) || mpvNeedsMinimal(failed[i].path) {"}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossDoesNotStripOptionsAfterAnExitThatIsNotAnInitFailure"}},
	},
	{
		name:  "the tuning is stripped only after exit code 2",
		edits: []codeEdit{{mpvLaunch, "early.code != 1 ||", "early.code != 2 ||"}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossRetriesWithoutTheTuningOptionsAnMPVRejects"}},
	},
	{
		name:  "a launch with nothing to leave out is repeated",
		edits: []codeEdit{{mpvLaunch, "\tif len(dropped) == 0 {\n\t\treturn \"\", false\n\t}\n\tfor i := range failed {", "\tfor i := range failed {"}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossHasNothingToStripWhenThereIsNoTuning"}},
	},
	{
		name:  "a minimal launch is retried with the minimal options again",
		edits: []codeEdit{{mpvLaunch, "|| mpvNeedsMinimal(failed[i].path) {", "{"}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossDoesNotRetryAMinimalLaunchWithTheMinimalOptions"}},
	},
	{
		name:  "the stripped retry is not made at all",
		edits: []codeEdit{{mpvLaunch, "\tif socket, ok := startMinimal(failed, safeLink, args); ok {\n\t\treturn socket, nil\n\t}\n", ""}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossRetriesWithoutTheTuningOptionsAnMPVRejects", "TestStartVideoAcrossGoesStraightToTheMinimalOptionsForAnMPVThatNeedsThem"}},
	},
	{
		name:  "the retry does not say it did not help",
		edits: []codeEdit{{mpvLaunch, "\t\tfailed[i].err = fmt.Errorf(\"%w\\n\\nTrying again without GoAnime's video options (%s) did not help: %v\",\n\t\t\tfailed[i].err, strings.Join(dropped, \" \"), firstLine(err.Error()))\n", ""}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossReportsWhatEachMPVSaid"}},
	},
	{
		name:  "the stripped retry is made before the other mpv are tried",
		edits: []codeEdit{{mpvLaunch, "\t\tfailed = append(failed, mpvAttempt{path: path, err: err})\n\t}\n", "\t\tfailed = append(failed, mpvAttempt{path: path, err: err})\n\t\tif socket, ok := startMinimal(failed, safeLink, args); ok {\n\t\t\treturn socket, nil\n\t\t}\n\t}\n"}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossTriesEveryMPVBeforeStrippingOptions"}},
	},
	{
		name:  "the other mpv are not listed in the error",
		edits: []codeEdit{{mpvLaunch, "\t\tfmt.Fprintf(&extra, \"\\nAlso tried %s: %s\", a.path, firstLine(a.err.Error()))\n", "\t\t_ = a\n"}},
		kills: map[string][]string{playerPkg: {"TestDescribeMPVFailures", "TestStartVideoAcrossReportsWhatEachMPVSaid"}},
	},
	{
		name:  "the Store alias is not recognised",
		edits: []codeEdit{{mpvLaunch, "\"/microsoft/windowsapps/\")", "\"/microsoft/windowsapps-x/\")"}},
		kills: map[string][]string{playerPkg: {"TestIsStoreAlias", "TestDescribeMPVFailures"}},
	},
	{
		name:  "every mpv is taken for a Store alias",
		edits: []codeEdit{{mpvLaunch, "func isStoreAlias(path string) bool {\n", "func isStoreAlias(path string) bool {\n\treturn true\n"}},
		kills: map[string][]string{playerPkg: {"TestIsStoreAlias", "TestDescribeMPVFailures"}},
	},
	{
		name:  "the Store alias is not told what to do",
		edits: []codeEdit{{mpvLaunch, "\tif isStoreAlias(first.path) {", "\tif false && isStoreAlias(first.path) {"}},
		kills: map[string][]string{playerPkg: {"TestDescribeMPVFailures"}},
	},
	{
		name:  "the cached mpv is not kept first",
		edits: []codeEdit{{mpvLaunch, "\tout := make([]string, 0, len(paths)+1)\n\tout = append(out, first)\n", "\tout := make([]string, 0, len(paths)+1)\n"}},
		kills: map[string][]string{playerPkg: {"TestLeadWith"}},
	},
	{
		name:  "the cached mpv is listed twice",
		edits: []codeEdit{{mpvLaunch, "\t\tif p != first {\n\t\t\tout = append(out, p)\n\t\t}", "\t\tout = append(out, p)"}},
		kills: map[string][]string{playerPkg: {"TestLeadWith"}},
	},
	{
		name:  "a launch that fails to start is not typed",
		edits: []codeEdit{{playerGo, "return \"\", &mpvStartError{err: fmt.Errorf(\"failed to start mpv: %w (stderr: %s)\", err, stderr.String())}", "return \"\", fmt.Errorf(\"failed to start mpv: %w (stderr: %s)\", err, stderr.String())"}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossFallsThroughWhenAnMPVCannotBeStartedAtAll"}},
	},
	{
		name:  "an early exit forgets its code",
		edits: []codeEdit{{playerGo, "\t\t\t\tcode: ex.code,\n", "\t\t\t\tcode: 0,\n"}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossRetriesWithoutTheTuningOptionsAnMPVRejects", "TestStartVideoAcrossReportsWhatEachMPVSaid"}},
	},
	{
		name:  "an early exit is not typed",
		edits: []codeEdit{{playerGo, "\t\t\treturn \"\", &mpvEarlyExitError{\n\t\t\t\tcode: ex.code,\n\t\t\t\terr:  formatMPVEarlyExitError(ex.err, joinMPVDiagnostics(stderr.String(), mpvProblemLines(ex.log, 15)), mpvPath),\n\t\t\t}", "\t\t\treturn \"\", formatMPVEarlyExitError(ex.err, joinMPVDiagnostics(stderr.String(), mpvProblemLines(ex.log, 15)), mpvPath)"}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossFallsThroughToTheNextMPVWhenThePathOneExits"}},
	},
	{
		name:  "the wait for mpv's socket ignores the timeout setting",
		edits: []codeEdit{{playerGo, "maxWaitTime := mpvStartTimeout\n", "maxWaitTime := 12 * time.Second\n"}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossDoesNotWaitOnASecondMPVWhenTheFirstCameUpAndWentSilent"}},
	},
	{
		name:  "mpv's stderr is read while it is still being written",
		race:  true,
		edits: []codeEdit{{playerGo, "\tvar stderr lockedBuffer\n", "\tvar stderr bytes.Buffer\n"}},
		kills: map[string][]string{playerPkg: {"TestStartVideoAcrossReadsStderrSafelyWhileMPVIsStillWritingIt"}},
	},
}

func TestIssue229GuardKillsMutants(t *testing.T) {
	killMutants(t, issue229Guards, issue229Mutants)
}

var issue229Guards = map[string][]string{
	tuiPkg:      {"TestPickerShowsItsNoticeAboveTheList", "TestNoticeOnNextPickerIsTakenOnce"},
	playbackPkg: {"TestSourceVideoRemovedMessage", "TestPlayEpisodeExplainsARemovedVideoOnTheListItReturnsTo"},
	playerPkg:   {"TestEssentialMPVArgsKeepsWhatPlaybackNeeds", "TestStartVideoAcrossStartsTheFirstMPVThatComesUp"},
}
