package api

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests of the tests, for forgetting history: x and X on the SEARCH home
// screen's CONTINUE and RECENT lists, and --clear-continue, --clear-recent and
// --clear-history on the command line. Same method as stream_publish_meta_test.go
// (the runner is shared) plus two guards of their own:
//
//   - the files that hold nothing but this feature have no function that is
//     not listed here, so a function added to them cannot slip in unlisted;
//   - every listed function is fully covered when the tests run, so a listed
//     function without a test fails.
//
// A mutant breaks one decision of the feature; the guard tests it names must
// fail on it.

const (
	fhTrackingPkg = "./internal/tracking"
	fhHandlersPkg = "./internal/handlers"
	fhUtilPkg     = "./internal/util"

	fhClearFile         = "internal/tracking/clear.go"
	fhHomeFile          = "internal/tui/search_home.go"
	fhHandlersHomeFile  = "internal/handlers/searchhome.go"
	fhHandlersClearFile = "internal/handlers/clear.go"
	fhClearRequestFile  = "internal/util/clear_request.go"
	fhUtilFile          = "internal/util/util.go"
	fhHelpFile          = "internal/util/help.go"
)

var forgetHistoryMutants = []mutant{
	{
		name:  "a query is stored with its words squeezed but looked up without",
		edits: []codeEdit{{fhClearFile, "strings.Join(strings.Fields(query), \" \")", "strings.TrimSpace(query)"}},
		kills: map[string][]string{fhTrackingPkg: {"TestNormalizeQuery", "TestDeleteRecentSearchFindsItBothWaysItIsSpelled"}},
	},
	{
		name:  "a failed delete counts the rows it touched",
		edits: []codeEdit{{fhClearFile, "\tif err != nil {\n\t\treturn 0, err\n\t}\n\tn, err := res.RowsAffected()", "\tn, err := res.RowsAffected()"}},
		kills: map[string][]string{fhTrackingPkg: {"TestAffected"}},
	},
	{
		name:  "forgetting a search that is not there says it was",
		edits: []codeEdit{{fhClearFile, "return n > 0, err", "return n >= 0, err"}},
		kills: map[string][]string{fhTrackingPkg: {"TestDeleteRecentSearchThatIsNotThere"}},
	},
	{
		name:  "a search is looked up as typed, not as stored",
		edits: []codeEdit{{fhClearFile, "`DELETE FROM recent_searches WHERE query = ?`, normalizeQuery(query))", "`DELETE FROM recent_searches WHERE query = ?`, query)"}},
		kills: map[string][]string{fhTrackingPkg: {"TestDeleteRecentSearchFindsItBothWaysItIsSpelled"}},
	},
	{
		name:  "forgetting one search forgets them all",
		edits: []codeEdit{{fhClearFile, "t.db.Exec(`DELETE FROM recent_searches WHERE query = ?`, normalizeQuery(query))", "t.db.Exec(`DELETE FROM recent_searches`)"}},
		kills: map[string][]string{fhTrackingPkg: {"TestDeleteRecentSearch"}},
	},
	{
		name:  "clearing RECENT clears CONTINUE too",
		edits: []codeEdit{{fhClearFile, "return affected(t.db.Exec(`DELETE FROM recent_searches`))", "return affected(t.db.Exec(`DELETE FROM recent_searches; DELETE FROM media_progress`))"}},
		kills: map[string][]string{fhTrackingPkg: {"TestClearRecentSearches"}},
	},
	{
		name:  "clearing CONTINUE clears RECENT too",
		edits: []codeEdit{{fhClearFile, "return affected(t.db.Exec(`DELETE FROM media_progress`))", "return affected(t.db.Exec(`DELETE FROM media_progress; DELETE FROM recent_searches`))"}},
		kills: map[string][]string{fhTrackingPkg: {"TestClearProgress"}},
	},
	{
		name:  "clearing CONTINUE clears nothing",
		edits: []codeEdit{{fhClearFile, "return affected(t.db.Exec(`DELETE FROM media_progress`))", "return affected(t.db.Exec(`DELETE FROM media_progress WHERE 0`))"}},
		kills: map[string][]string{fhTrackingPkg: {"TestClearProgress"}},
	},
	{
		name:  "clearing RECENT clears nothing",
		edits: []codeEdit{{fhClearFile, "return affected(t.db.Exec(`DELETE FROM recent_searches`))", "return affected(t.db.Exec(`DELETE FROM recent_searches WHERE 0`))"}},
		kills: map[string][]string{fhTrackingPkg: {"TestClearRecentSearches"}},
	},
	{
		name:  "a tracker that is not there is cleared anyway",
		edits: []codeEdit{{fhClearFile, "\tif t == nil || t.db == nil {\n\t\treturn 0, ErrTrackerNotInited\n\t}\n\treturn affected(t.db.Exec(`DELETE FROM media_progress`))", "\treturn affected(t.db.Exec(`DELETE FROM media_progress`))"}},
		kills: map[string][]string{fhTrackingPkg: {"TestClearingOnATrackerThatIsNotThere"}},
	},
	{
		name:  "forgetting a search without history succeeds",
		edits: []codeEdit{{fhClearFile, "\tif tracker == nil {\n\t\treturn ErrHistoryUnavailable\n\t}\n\t_, err := tracker.DeleteRecentSearch(query)", "\tif tracker == nil {\n\t\treturn nil\n\t}\n\t_, err := tracker.DeleteRecentSearch(query)"}},
		kills: map[string][]string{fhTrackingPkg: {"TestClearingWithoutHistoryIsReportedNotSwallowed"}},
	},
	{
		name:  "clearing CONTINUE clears RECENT",
		edits: []codeEdit{{fhClearFile, "return tracker.ClearProgress()", "return tracker.ClearRecentSearches()"}},
		kills: map[string][]string{fhTrackingPkg: {"TestForgetSearchClearRecentAndClearContinueOnTheSharedDatabase"}},
	},
	{
		name:  "clearing RECENT clears CONTINUE",
		edits: []codeEdit{{fhClearFile, "return tracker.ClearRecentSearches()\n}\n\n// ClearContinue", "return tracker.ClearProgress()\n}\n\n// ClearContinue"}},
		kills: map[string][]string{fhTrackingPkg: {"TestForgetSearchClearRecentAndClearContinueOnTheSharedDatabase"}},
	},
	{
		name:  "x does nothing on a RECENT row",
		edits: []codeEdit{{fhHomeFile, "m.forgetRecent(row.index)", "_ = row.index"}},
		kills: map[string][]string{tuiPkg: {"TestXRemovesTheHighlightedRecentSearch"}},
	},
	{
		name:  "X asks nothing",
		edits: []codeEdit{{fhHomeFile, "m.askToClear(row.kind)", "_ = row.kind"}},
		kills: map[string][]string{tuiPkg: {"TestXOnAContinueRowAsksBeforeClearingAnything", "TestXOnARecentRowAsksAboutRecent"}},
	},
	{
		name:  "a RECENT row that fails to go is dropped anyway",
		edits: []codeEdit{{fhHomeFile, "m.failed(\"remove the search\", err)\n\t\treturn\n", "m.failed(\"remove the search\", err)\n"}},
		kills: map[string][]string{tuiPkg: {"TestARecentSearchThatFailsToGoStaysAndTheScreenSaysWhy"}},
	},
	{
		name:  "a CONTINUE row that fails to go is dropped anyway",
		edits: []codeEdit{{fhHomeFile, "m.failed(\"remove the episode\", err)\n\t\treturn\n", "m.failed(\"remove the episode\", err)\n"}},
		kills: map[string][]string{tuiPkg: {"TestAnEpisodeThatFailsToGoStaysAndTheScreenSaysWhy"}},
	},
	{
		name:  "x removes the first RECENT row, not the highlighted one",
		edits: []codeEdit{{fhHomeFile, "m.recentAll = append(m.recentAll[:index], m.recentAll[index+1:]...)", "m.recentAll = m.recentAll[1:]"}},
		kills: map[string][]string{tuiPkg: {"TestXRemovesTheSecondRecentSearchWithoutTouchingTheFirst"}},
	},
	{
		name:  "x forgets a blank search",
		edits: []codeEdit{{fhHomeFile, "m.deleteRecent(m.recentAll[index].Query)", "m.deleteRecent(\"\")"}},
		kills: map[string][]string{tuiPkg: {"TestXRemovesTheHighlightedRecentSearch"}},
	},
	{
		name:  "x without a remover still calls it",
		edits: []codeEdit{{fhHomeFile, "\tif m.deleteRecent == nil {\n\t\treturn\n\t}\n", ""}},
		kills: map[string][]string{tuiPkg: {"TestXOnARecentSearchThatCannotBeRemovedDoesNothing"}},
	},
	{
		name:  "the question about CONTINUE is about RECENT",
		edits: []codeEdit{{fhHomeFile, "m.clearing = clearContinueRows", "m.clearing = clearRecentRows"}},
		kills: map[string][]string{tuiPkg: {"TestXOnAContinueRowAsksBeforeClearingAnything"}},
	},
	{
		name:  "the question about RECENT is about CONTINUE",
		edits: []codeEdit{{fhHomeFile, "m.clearing = clearRecentRows", "m.clearing = clearContinueRows"}},
		kills: map[string][]string{tuiPkg: {"TestXOnARecentRowAsksAboutRecent"}},
	},
	{
		name:  "an empty CONTINUE list is asked about",
		edits: []codeEdit{{fhHomeFile, "m.clearContinue != nil && len(m.continueAll) > 0:", "m.clearContinue != nil:"}},
		kills: map[string][]string{tuiPkg: {"TestXAsksNothingWhenTheListCannotBeClearedOrIsEmpty"}},
	},
	{
		name:  "a CONTINUE list that cannot be cleared is asked about",
		edits: []codeEdit{{fhHomeFile, "kind == homeRowContinue && m.clearContinue != nil && len(m.continueAll) > 0:", "kind == homeRowContinue && len(m.continueAll) > 0:"}},
		kills: map[string][]string{tuiPkg: {"TestXAsksNothingWhenTheListCannotBeClearedOrIsEmpty"}},
	},
	{
		name:  "a RECENT list that cannot be cleared is asked about",
		edits: []codeEdit{{fhHomeFile, "kind == homeRowRecent && m.clearRecent != nil && len(m.recentAll) > 0:", "kind == homeRowRecent && len(m.recentAll) > 0:"}},
		kills: map[string][]string{tuiPkg: {"TestXAsksNothingWhenTheListCannotBeClearedOrIsEmpty"}},
	},
	{
		name:  "y does not clear",
		edits: []codeEdit{{fhHomeFile, "\tcase \"y\", \"Y\":\n\t\tm.clearConfirmed()", "\tcase \"y\", \"Y\":\n\t\tm.clearing = clearNothing"}},
		kills: map[string][]string{tuiPkg: {"TestYClearsTheContinueListAndOnlyThat", "TestYClearsTheRecentListAndOnlyThat"}},
	},
	{
		name:  "esc does not close the question",
		edits: []codeEdit{{fhHomeFile, "\tcase \"n\", \"N\", \"esc\":\n\t\tm.clearing = clearNothing", "\tcase \"n\", \"N\":\n\t\tm.clearing = clearNothing"}},
		kills: map[string][]string{tuiPkg: {"TestNAndEscKeepTheList"}},
	},
	{
		name:  "n does not close the question",
		edits: []codeEdit{{fhHomeFile, "\tcase \"n\", \"N\", \"esc\":\n\t\tm.clearing = clearNothing", "\tcase \"N\", \"esc\":\n\t\tm.clearing = clearNothing"}},
		kills: map[string][]string{tuiPkg: {"TestNAndEscKeepTheList"}},
	},
	{
		name:  "ctrl+c is ignored by the question",
		edits: []codeEdit{{fhHomeFile, "\tcase \"ctrl+c\":\n\t\tm.err = ErrSearchHomeCancelled\n\t\treturn m, tea.Quit\n\t}\n\treturn m, nil\n}\n\n// clearConfirmed", "\t}\n\treturn m, nil\n}\n\n// clearConfirmed"}},
		kills: map[string][]string{tuiPkg: {"TestCtrlCQuitsFromTheQuestion"}},
	},
	{
		name:  "other keys get through the question",
		edits: []codeEdit{{fhHomeFile, "\t\t\treturn m.updateClearQuestion(msg)", "\t\t\t_ = msg"}},
		kills: map[string][]string{tuiPkg: {"TestOtherKeysDoNotAnswerTheQuestionNorLeakThrough"}},
	},
	{
		name:  "CONTINUE stays when its clear fails: not kept",
		edits: []codeEdit{{fhHomeFile, "m.failed(\"clear CONTINUE\", err)\n\t\t\treturn\n", "m.failed(\"clear CONTINUE\", err)\n"}},
		kills: map[string][]string{tuiPkg: {"TestAListThatFailsToClearStaysAndTheScreenSaysWhy"}},
	},
	{
		name:  "RECENT is emptied although its clear failed",
		edits: []codeEdit{{fhHomeFile, "m.failed(\"clear RECENT\", err)\n\t\t\treturn\n", "m.failed(\"clear RECENT\", err)\n"}},
		kills: map[string][]string{tuiPkg: {"TestAListThatFailsToClearStaysAndTheScreenSaysWhy"}},
	},
	{
		name:  "clearing CONTINUE empties RECENT",
		edits: []codeEdit{{fhHomeFile, "\t\tm.continueAll = nil\n", "\t\tm.recentAll = nil\n"}},
		kills: map[string][]string{tuiPkg: {"TestYClearsTheContinueListAndOnlyThat"}},
	},
	{
		name:  "clearing RECENT empties CONTINUE",
		edits: []codeEdit{{fhHomeFile, "\t\tm.recentAll = nil\n", "\t\tm.continueAll = nil\n"}},
		kills: map[string][]string{tuiPkg: {"TestYClearsTheRecentListAndOnlyThat"}},
	},
	{
		name:  "clearing CONTINUE calls the RECENT clearer",
		edits: []codeEdit{{fhHomeFile, "if err := m.clearContinue(); err != nil {", "if err := m.clearRecent(); err != nil {"}},
		kills: map[string][]string{tuiPkg: {"TestYClearsTheContinueListAndOnlyThat"}},
	},
	{
		name:  "clearing RECENT calls the CONTINUE clearer",
		edits: []codeEdit{{fhHomeFile, "if err := m.clearRecent(); err != nil {", "if err := m.clearContinue(); err != nil {"}},
		kills: map[string][]string{tuiPkg: {"TestYClearsTheRecentListAndOnlyThat"}},
	},
	{
		name:  "the box does not get the focus back",
		edits: []codeEdit{{fhHomeFile, "\tif len(m.rows) == 0 {\n\t\tm.focusInput = true\n\t\tm.input.Focus()\n\t}\n}\n\n// failed", "\tif len(m.rows) == 0 {\n\t\tm.focusInput = true\n\t}\n}\n\n// failed"}},
		kills: map[string][]string{tuiPkg: {"TestClearingTheOnlyListHandsTheBoxItsFocusBack", "TestXOnTheLastRecentSearchHandsTheBoxItsFocusBack"}},
	},
	{
		name:  "a failure is reported without saying what failed",
		edits: []codeEdit{{fhHomeFile, "\"Could not %s: %v\"", "\"%s: %v\""}},
		kills: map[string][]string{tuiPkg: {"TestAnEpisodeThatFailsToGoStaysAndTheScreenSaysWhy"}},
	},
	{
		name:  "the question counts every list as plural",
		edits: []codeEdit{{fhHomeFile, "countOf(len(m.continueAll), \"episode\", \"episodes\")", "countOf(len(m.continueAll), \"episodes\", \"episodes\")"}},
		kills: map[string][]string{tuiPkg: {"TestTheQuestionSaysOneInTheSingular"}},
	},
	{
		name:  "one search is \"1 searchs\"",
		edits: []codeEdit{{fhHomeFile, "countOf(len(m.recentAll), \"search\", \"searches\")", "countOf(len(m.recentAll), \"search\", \"searchs\")"}},
		kills: map[string][]string{tuiPkg: {"TestXOnARecentRowAsksAboutRecent"}},
	},
	{
		name:  "one is plural",
		edits: []codeEdit{{fhHomeFile, "func countOf(n int, one, many string) string {\n\tif n == 1 {", "func countOf(n int, one, many string) string {\n\tif n == 100 {"}},
		kills: map[string][]string{tuiPkg: {"TestCountOf", "TestTheQuestionSaysOneInTheSingular"}},
	},
	{
		name:  "the question has no footer",
		edits: []codeEdit{{fhHomeFile, "{Key: \"y\", Desc: \"clear all\", Essential: true},\n", ""}},
		kills: map[string][]string{tuiPkg: {"TestTheQuestionHasItsOwnFooter"}},
	},
	{
		name:  "the footer never offers x and X",
		edits: []codeEdit{{fhHomeFile, "\tif !m.focusInput {\n\t\thints = append(hints[:3], append([]Hint{{Key: \"x\", Desc: \"remove\", Essential: true}", "\tif false {\n\t\thints = append(hints[:3], append([]Hint{{Key: \"x\", Desc: \"remove\", Essential: true}"}},
		kills: map[string][]string{tuiPkg: {"TestTheFooterOffersRemoveAndClearOnlyOverARow"}},
	},
	{
		name:  "a failure stays on screen after the next key",
		edits: []codeEdit{{fhHomeFile, "\t\tm.notice = \"\"\n", ""}},
		kills: map[string][]string{tuiPkg: {"TestTheFailureNoticeGoesAwayWithTheNextKey"}},
	},
	{
		name:  "x and X act while typing in the box",
		edits: []codeEdit{{fhHomeFile, "\t\t\tif !m.focusInput {\n\t\t\t\tif cmd := m.rowAction(msg.String()); cmd != nil {", "\t\t\tif true {\n\t\t\t\tif cmd := m.rowAction(msg.String()); cmd != nil {"}},
		kills: map[string][]string{tuiPkg: {"TestXInTheBoxTypesAnX"}},
	},
	{
		name:  "the help does not mention clearing",
		edits: []codeEdit{{fhHomeFile, "\t\t{Key: \"X\", Desc: \"clear the list\"},\n", ""}},
		kills: map[string][]string{tuiPkg: {"TestTheHelpScreenListsClearTheList"}},
	},
	{
		name:  "the screen is not given the history actions",
		edits: []codeEdit{{fhHomeFile, "deleteRecent:   opts.DeleteRecent,", "deleteRecent:   nil,"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeWiresHistoryOptionsIntoTheModel"}},
	},
	{
		name:  "the screen is not given the CONTINUE clearer",
		edits: []codeEdit{{fhHomeFile, "clearContinue:  opts.ClearContinue,", "clearContinue:  nil,"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeWiresHistoryOptionsIntoTheModel"}},
	},
	{
		name:  "the screen is not given the RECENT clearer",
		edits: []codeEdit{{fhHomeFile, "clearRecent:    opts.ClearRecent,", "clearRecent:    nil,"}},
		kills: map[string][]string{tuiPkg: {"TestSearchHomeWiresHistoryOptionsIntoTheModel"}},
	},
	{
		name:  "the question is not drawn",
		edits: []codeEdit{{fhHomeFile, "\tif question := m.questionView(width); question != \"\" {\n\t\tsections = append(sections, question)\n\t}\n", ""}},
		kills: map[string][]string{tuiPkg: {"TestXOnAContinueRowAsksBeforeClearingAnything"}},
	},
	{
		name:  "removing a show removes only its latest episode",
		edits: []codeEdit{{fhHandlersHomeFile, "if continueGroupKey(r) == group {", "if r.AllanimeID == key {"}},
		kills: map[string][]string{fhHandlersPkg: {"TestKeysOfShow", "TestRemovingAShowFromContinueRemovesEveryEpisodeOfIt"}},
	},
	{
		name:  "a failure removing an episode is swallowed",
		edits: []codeEdit{{fhHandlersHomeFile, "\t\t\treturn err\n\t\t}\n\t}\n\treturn nil\n}\n\n// forgetRecentSearch", "\t\t\tcontinue\n\t\t}\n\t}\n\treturn nil\n}\n\n// forgetRecentSearch"}},
		kills: map[string][]string{fhHandlersPkg: {"TestDeleteContinueEntryStopsAtTheFirstFailure"}},
	},
	{
		name:  "a history that cannot be read is deleted from anyway",
		edits: []codeEdit{{fhHandlersHomeFile, "\trows, err := continueRows()\n\tif err != nil {\n\t\treturn err\n\t}\n", "\trows, _ := continueRows()\n"}},
		kills: map[string][]string{fhHandlersPkg: {"TestDeleteContinueEntryReportsAHistoryItCannotRead"}},
	},
	{
		name:  "forgetting a search does nothing",
		edits: []codeEdit{{fhHandlersHomeFile, "return forgetSearchFn(query)", "return nil"}},
		kills: map[string][]string{fhHandlersPkg: {"TestForgetRecentSearchAndTheClearsPassThroughWhatHistoryAnswers", "TestTheHomeScreenIsGivenTheHistoryActions"}},
	},
	{
		name:  "clearing CONTINUE clears RECENT (handler)",
		edits: []codeEdit{{fhHandlersHomeFile, "_, err := clearContinueFn()", "_, err := clearRecentFn()"}},
		kills: map[string][]string{fhHandlersPkg: {"TestClearContinueAndRecentAreSeparateCalls", "TestTheHomeScreenIsGivenTheHistoryActions"}},
	},
	{
		name:  "clearing RECENT clears CONTINUE (handler)",
		edits: []codeEdit{{fhHandlersHomeFile, "_, err := clearRecentFn()", "_, err := clearContinueFn()"}},
		kills: map[string][]string{fhHandlersPkg: {"TestClearContinueAndRecentAreSeparateCalls", "TestTheHomeScreenIsGivenTheHistoryActions"}},
	},
	{
		name:  "the screen is not given a RECENT remover",
		edits: []codeEdit{{fhHandlersHomeFile, "\t\tDeleteRecent:  forgetRecentSearch,\n", ""}},
		kills: map[string][]string{fhHandlersPkg: {"TestTheHomeScreenIsGivenTheHistoryActions"}},
	},
	{
		name:  "the screen is given the RECENT clearer for CONTINUE",
		edits: []codeEdit{{fhHandlersHomeFile, "ClearContinue: clearContinueHistory,", "ClearContinue: clearRecentHistory,"}},
		kills: map[string][]string{fhHandlersPkg: {"TestTheHomeScreenIsGivenTheHistoryActions"}},
	},
	{
		name:  "the screen is given the CONTINUE clearer for RECENT",
		edits: []codeEdit{{fhHandlersHomeFile, "ClearRecent:   clearRecentHistory,", "ClearRecent:   clearContinueHistory,"}},
		kills: map[string][]string{fhHandlersPkg: {"TestTheHomeScreenIsGivenTheHistoryActions"}},
	},
	{
		name:  "a missing history is deleted from anyway",
		edits: []codeEdit{{fhHandlersHomeFile, "\tif tracker == nil {\n\t\treturn fmt.Errorf(\"watch history unavailable\")\n\t}\n\treturn tracker.DeleteAnime(0, key)", "\treturn tracker.DeleteAnime(0, key)"}},
		kills: map[string][]string{fhHandlersPkg: {"TestDeleteTrackedRowWithoutHistoryReportsIt"}},
	},
	{
		name:  "the command clears without asking and without a terminal",
		edits: []codeEdit{{fhHandlersClearFile, "\t\tif !interactive {", "\t\tif false {"}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearWithoutATerminalAndWithoutYesClearsNothing"}},
	},
	{
		name:  "--yes still asks",
		edits: []codeEdit{{fhHandlersClearFile, "\tif !req.Yes {", "\tif true {"}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearYesClearsTheListsAskedForAndOnlyThose"}},
	},
	{
		name:  "anything is a yes",
		edits: []codeEdit{{fhHandlersClearFile, "\t\treturn true\n\t}\n\treturn false\n}\n\n// appendClear", "\t\treturn true\n\t}\n\treturn true\n}\n\n// appendClear"}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearKeepsEverythingOnAnythingButYes", "TestAnsweredYes"}},
	},
	{
		name:  "n is a yes",
		edits: []codeEdit{{fhHandlersClearFile, "case \"y\", \"yes\", \"s\", \"sim\":", "case \"y\", \"yes\", \"s\", \"sim\", \"n\":"}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearKeepsEverythingOnAnythingButYes", "TestAnsweredYes"}},
	},
	{
		name:  "yes is not a yes",
		edits: []codeEdit{{fhHandlersClearFile, "case \"y\", \"yes\", \"s\", \"sim\":", "case \"y\", \"s\", \"sim\":"}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearAsksAndClearsOnYes", "TestAnsweredYes"}},
	},
	{
		name:  "RECENT is never cleared",
		edits: []codeEdit{{fhHandlersClearFile, "\tif req.Recent {", "\tif false {"}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearYesClearsTheListsAskedForAndOnlyThose"}},
	},
	{
		name:  "CONTINUE is never cleared",
		edits: []codeEdit{{fhHandlersClearFile, "\tif req.Continue {", "\tif false {"}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearYesClearsTheListsAskedForAndOnlyThose"}},
	},
	{
		name:  "CONTINUE is cleared with the RECENT clearer",
		edits: []codeEdit{{fhHandlersClearFile, "\"episodes\", clearContinueFn)", "\"episodes\", clearRecentFn)"}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearYesClearsTheListsAskedForAndOnlyThose"}},
	},
	{
		name:  "a failed clear is not reported",
		edits: []codeEdit{{fhHandlersClearFile, "return append(failures, fmt.Errorf(\"clear %s: %w\", list, err))", "return failures"}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearReportsEveryFailureAndStillTriesTheOtherList"}},
	},
	{
		name:  "failures are dropped",
		edits: []codeEdit{{fhHandlersClearFile, "return errors.Join(failures...)", "return nil"}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearReportsEveryFailureAndStillTriesTheOtherList"}},
	},
	{
		name:  "a failed clear loses its cause",
		edits: []codeEdit{{fhHandlersClearFile, "fmt.Errorf(\"clear %s: %w\", list, err)", "fmt.Errorf(\"clear %s: %v\", list, err)"}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearKeepsTheCauseOfAFailureInspectable"}},
	},
	{
		name:  "both lists are described as CONTINUE",
		edits: []codeEdit{{fhHandlersClearFile, "case req.Continue && req.Recent:", "case req.Continue:"}},
		kills: map[string][]string{fhHandlersPkg: {"TestDescribeClear"}},
	},
	{
		name:  "an empty request is not an error",
		edits: []codeEdit{{fhHandlersClearFile, "if req == nil || (!req.Continue && !req.Recent) {", "if req == nil {"}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearWithNothingToClearIsAnError"}},
	},
	{
		name:  "an empty list is reported as cleared",
		edits: []codeEdit{{fhHandlersClearFile, "\tif n == 0 {", "\tif n < 0 {"}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearSaysSingularAndAlreadyEmpty"}},
	},
	{
		name:  "one is plural (command)",
		edits: []codeEdit{{fhHandlersClearFile, "if n == 1 {\n\t\tnoun = one", "if n == 100 {\n\t\tnoun = one"}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearSaysSingularAndAlreadyEmpty"}},
	},
	{
		name:  "the handler always thinks there is a terminal",
		edits: []codeEdit{{fhHandlersClearFile, "os.Stdin, os.Stdout, stdinIsTerminal())", "os.Stdin, os.Stdout, true)"}},
		kills: map[string][]string{fhHandlersPkg: {"TestHandleClearRequestAsksOnlyWhereStdinIsATerminal"}},
	},
	{
		name:  "the handler ignores the stored request",
		edits: []codeEdit{{fhHandlersClearFile, "runClear(util.GlobalClearRequest,", "runClear(nil,"}},
		kills: map[string][]string{fhHandlersPkg: {"TestHandleClearRequestUsesTheStoredRequest"}},
	},
	{
		name:  "declining prints nothing",
		edits: []codeEdit{{fhHandlersClearFile, "_, _ = fmt.Fprintln(out, \"Nothing was cleared.\")", ""}},
		kills: map[string][]string{fhHandlersPkg: {"TestRunClearKeepsEverythingOnAnythingButYes"}},
	},
	{
		name:  "--clear-history clears no CONTINUE",
		edits: []codeEdit{{fhClearRequestFile, "Continue: continueList || history,", "Continue: continueList,"}},
		kills: map[string][]string{fhUtilPkg: {"TestClearRequestFrom", "TestFlagParserClearFlags"}},
	},
	{
		name:  "--clear-history clears no RECENT",
		edits: []codeEdit{{fhClearRequestFile, "Recent:   recent || history,", "Recent:   recent,"}},
		kills: map[string][]string{fhUtilPkg: {"TestClearRequestFrom", "TestFlagParserClearFlags"}},
	},
	{
		name:  "--yes is ignored",
		edits: []codeEdit{{fhClearRequestFile, "Yes:      yes,", "Yes:      false,"}},
		kills: map[string][]string{fhUtilPkg: {"TestClearRequestFrom", "TestFlagParserClearFlags"}},
	},
	{
		name:  "--clear-history alone asks for nothing",
		edits: []codeEdit{{fhClearRequestFile, "if !continueList && !recent && !history {", "if !continueList && !recent {"}},
		kills: map[string][]string{fhUtilPkg: {"TestClearRequestFrom", "TestFlagParserClearFlags"}},
	},
	{
		name:  "the request is not stored",
		edits: []codeEdit{{fhUtilFile, "\t\tGlobalClearRequest = clearRequest\n", ""}},
		kills: map[string][]string{fhUtilPkg: {"TestFlagParserClearFlags"}},
	},
	{
		name:  "-y is not --yes",
		edits: []codeEdit{{fhUtilFile, "*yesFlag || *yesShortFlag", "*yesFlag || (*yesShortFlag && false)"}},
		kills: map[string][]string{fhUtilPkg: {"TestFlagParserClearFlags"}},
	},
	{
		name:  "the clear flags do not stop the program",
		edits: []codeEdit{{fhUtilFile, "\t\treturn \"\", ErrClearRequested\n", "\t\treturn \"\", nil\n"}},
		kills: map[string][]string{fhUtilPkg: {"TestFlagParserClearFlags"}},
	},
	{
		name:  "the clear commands probe the terminal",
		edits: []codeEdit{{fhUtilFile, "\tif clearRequest == nil {\n\t\tprobeTerminal()\n\t}", "\tif true {\n\t\tprobeTerminal()\n\t}"}},
		kills: map[string][]string{fhUtilPkg: {"TestTheClearCommandsDoNotProbeTheTerminal"}},
	},
	{
		name:  "nothing probes the terminal",
		edits: []codeEdit{{fhUtilFile, "\tif clearRequest == nil {\n\t\tprobeTerminal()\n\t}", "\tif false {\n\t\tprobeTerminal()\n\t}"}},
		kills: map[string][]string{fhUtilPkg: {"TestEverythingElseStillProbesTheTerminalOnce"}},
	},
	{
		name:  "the help forgets --clear-history",
		edits: []codeEdit{{fhHelpFile, "\"--clear-history\"", "\"--clear-all\""}},
		kills: map[string][]string{fhUtilPkg: {"TestTheHelpListsTheClearFlags"}},
	},
}

func TestForgetHistoryGuardKillsMutants(t *testing.T) {
	skipMutantsWhenShort(t)
	killMutants(t, nil, forgetHistoryMutants)
}

// forgetHistorySurface is every function the feature added or rewrote. A file
// marked whole holds only this feature: it must list every function in it.
var forgetHistorySurface = []struct {
	pkg   string // module-relative package directory
	file  string
	whole bool
	funcs []string
}{
	{"internal/tracking", "clear.go", true, []string{
		"normalizeQuery", "affected", "DeleteRecentSearch", "ClearRecentSearches", "ClearProgress",
		"ForgetSearch", "ClearRecent", "ClearContinue",
	}},
	{"internal/tui", "search_home.go", false, []string{
		"rowAction", "forgetContinue", "forgetRecent", "askToClear", "updateClearQuestion",
		"clearConfirmed", "rowsChanged", "failed", "questionView", "countOf", "hints",
	}},
	{"internal/handlers", "searchhome.go", false, []string{
		"deleteTrackedRow", "keysOfShow", "deleteContinueEntry", "forgetRecentSearch",
		"clearContinueHistory", "clearRecentHistory",
	}},
	{"internal/handlers", "clear.go", true, []string{
		"HandleClearRequest", "runClear", "describeClear", "answeredYes", "appendClear",
	}},
	{"internal/util", "clear_request.go", true, []string{"clearRequestFrom"}},
}

func declaredFuncs(t *testing.T, path string) []string {
	t.Helper()
	file, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.SkipObjectResolution)
	require.NoError(t, err)
	var names []string
	for _, decl := range file.Decls {
		if fn, ok := decl.(*ast.FuncDecl); ok {
			names = append(names, fn.Name.Name)
		}
	}
	sort.Strings(names)
	return names
}

func TestForgetHistoryHasNoFunctionThatIsNotListed(t *testing.T) {
	root := moduleRoot(t)
	for _, entry := range forgetHistorySurface {
		declared := declaredFuncs(t, filepath.Join(root, filepath.FromSlash(entry.pkg), entry.file))

		if entry.whole {
			want := append([]string(nil), entry.funcs...)
			sort.Strings(want)
			assert.Equal(t, want, declared, "%s/%s holds only this feature: a function added to it must be listed in forgetHistorySurface, where TestForgetHistoryFunctionsAreFullyCovered will demand its tests", entry.pkg, entry.file)
		}
		for _, name := range entry.funcs {
			assert.Contains(t, declared, name, "%s/%s no longer declares %s: the list is stale", entry.pkg, entry.file, name)
		}
	}
}

// notFullyCovered reads `go tool cover -func` output and returns the functions
// of file that are below 100%.
func notFullyCovered(report, file string, funcs []string) []string {
	want := map[string]bool{}
	for _, f := range funcs {
		want[f] = true
	}
	var short []string
	for _, line := range strings.Split(report, "\n") {
		fields := strings.Fields(line)
		if len(fields) != 3 || !strings.Contains(fields[0], "/"+file+":") {
			continue
		}
		if want[fields[1]] && fields[2] != "100.0%" {
			short = append(short, fields[1]+" "+fields[2])
		}
		delete(want, fields[1])
	}
	for f := range want {
		short = append(short, f+" (not in the report)")
	}
	sort.Strings(short)
	return short
}

func TestForgetHistoryFunctionsAreFullyCovered(t *testing.T) {
	skipMutantsWhenShort(t)
	if os.Getenv("CGO_ENABLED") == "0" || runtime.GOOS == "windows" {
		t.Skip("the history is SQLite: its tests need CGO, and Windows keeps it elsewhere")
	}
	goBin, err := exec.LookPath("go")
	if err != nil {
		t.Skip("no go toolchain on PATH")
	}
	root := moduleRoot(t)
	byPkg := map[string][]struct {
		file  string
		funcs []string
	}{}
	for _, entry := range forgetHistorySurface {
		byPkg[entry.pkg] = append(byPkg[entry.pkg], struct {
			file  string
			funcs []string
		}{entry.file, entry.funcs})
	}
	pkgs := make([]string, 0, len(byPkg))
	for pkg := range byPkg {
		pkgs = append(pkgs, pkg)
	}
	sort.Strings(pkgs)

	for _, pkg := range pkgs {
		profile := filepath.Join(t.TempDir(), "cover.out")
		run := exec.Command(goBin, "test", "-count=1", "-coverprofile="+profile, "./"+pkg)
		run.Dir = root
		out, err := run.CombinedOutput()
		require.NoError(t, err, "the package tests must pass to be measured:\n%s", out)

		report := exec.Command(goBin, "tool", "cover", "-func="+profile)
		report.Dir = root
		table, err := report.CombinedOutput()
		require.NoError(t, err, string(table))

		for _, entry := range byPkg[pkg] {
			assert.Empty(t, notFullyCovered(string(table), entry.file, entry.funcs), "%s/%s: functions of this feature the tests do not fully cover", pkg, entry.file)
		}
	}
}

func TestNotFullyCoveredReadsTheCoverageReport(t *testing.T) {
	report := "x/internal/tracking/clear.go:15:\tnormalizeQuery\t100.0%\n" +
		"x/internal/tracking/clear.go:20:\taffected\t66.7%\n" +
		"x/internal/tracking/recent.go:13:\tDBPath\t10.0%\n" +
		"total:\t(statements)\t80.1%\n"

	assert.Equal(t, []string{"affected 66.7%"}, notFullyCovered(report, "clear.go", []string{"normalizeQuery", "affected"}))
	assert.Equal(t, []string{"missing (not in the report)"}, notFullyCovered(report, "clear.go", []string{"missing"}), "a function the report does not list is not covered")
	assert.Empty(t, notFullyCovered(report, "clear.go", []string{"normalizeQuery"}))
	assert.Empty(t, notFullyCovered(report, "recent.go", nil), "functions not asked for are not judged")
}
