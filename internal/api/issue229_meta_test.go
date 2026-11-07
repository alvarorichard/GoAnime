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
}

func TestIssue229GuardKillsMutants(t *testing.T) {
	killMutants(t, issue229Guards, issue229Mutants)
}

var issue229Guards = map[string][]string{
	tuiPkg:      {"TestPickerShowsItsNoticeAboveTheList", "TestNoticeOnNextPickerIsTakenOnce"},
	playbackPkg: {"TestSourceVideoRemovedMessage", "TestPlayEpisodeExplainsARemovedVideoOnTheListItReturnsTo"},
}
