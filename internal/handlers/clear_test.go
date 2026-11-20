package handlers

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/tracking"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --clear-continue, --clear-recent and --clear-history. The history is a fake
// that counts what was cleared; the terminal is a reader and a buffer.

func clearRun(t *testing.T, req *util.ClearRequest, stdin string, interactive bool) (out string, err error) {
	t.Helper()
	var buf bytes.Buffer
	err = runClear(req, strings.NewReader(stdin), &buf, interactive)
	return buf.String(), err
}

func TestRunClearWithNothingToClearIsAnError(t *testing.T) {
	h := &historyFake{}
	useHistoryFake(t, h)

	for name, req := range map[string]*util.ClearRequest{
		"no request":      nil,
		"an empty one":    {},
		"only --yes":      {Yes: true},
		"neither list on": {Yes: true, Continue: false, Recent: false},
	} {
		_, err := clearRun(t, req, "", true)
		require.Error(t, err, name)
		assert.Contains(t, err.Error(), "nothing to clear", name)
	}
	assert.Zero(t, h.clearedCon+h.clearedRec)
}

func TestRunClearYesClearsTheListsAskedForAndOnlyThose(t *testing.T) {
	for name, tc := range map[string]struct {
		req             util.ClearRequest
		con, rec        int
		wantOut, notOut string
	}{
		"continue": {util.ClearRequest{Continue: true, Yes: true}, 1, 0, "Cleared 7 episodes from CONTINUE.\n", "RECENT"},
		"recent":   {util.ClearRequest{Recent: true, Yes: true}, 0, 1, "Cleared 4 searches from RECENT.\n", "CONTINUE"},
		"both":     {util.ClearRequest{Continue: true, Recent: true, Yes: true}, 1, 1, "Cleared 7 episodes from CONTINUE.\nCleared 4 searches from RECENT.\n", ""},
	} {
		h := &historyFake{}
		useHistoryFake(t, h)

		out, err := clearRun(t, &tc.req, "", false)

		require.NoError(t, err, name)
		assert.Equal(t, tc.con, h.clearedCon, name)
		assert.Equal(t, tc.rec, h.clearedRec, name)
		assert.Equal(t, tc.wantOut, out, name)
		if tc.notOut != "" {
			assert.NotContains(t, out, tc.notOut, name)
		}
		assert.NotContains(t, out, "[y/N]", "--yes asks nothing: "+name)
	}
}

func TestRunClearSaysSingularAndAlreadyEmpty(t *testing.T) {
	h := &historyFake{}
	useHistoryFake(t, h)
	clearContinueFn = func() (int, error) { return 1, nil }
	clearRecentFn = func() (int, error) { return 0, nil }

	out, err := clearRun(t, &util.ClearRequest{Continue: true, Recent: true, Yes: true}, "", false)

	require.NoError(t, err)
	assert.Equal(t, "Cleared 1 episode from CONTINUE.\nRECENT was already empty.\n", out)

	clearRecentFn = func() (int, error) { return 1, nil }
	out, err = clearRun(t, &util.ClearRequest{Recent: true, Yes: true}, "", false)
	require.NoError(t, err)
	assert.Equal(t, "Cleared 1 search from RECENT.\n", out)
}

func TestRunClearWithoutATerminalAndWithoutYesClearsNothing(t *testing.T) {
	h := &historyFake{}
	useHistoryFake(t, h)

	out, err := clearRun(t, &util.ClearRequest{Continue: true, Recent: true}, "y\n", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "not clearing the CONTINUE list and the RECENT searches")
	assert.Contains(t, err.Error(), "--yes", "and says how to do it anyway")
	assert.Zero(t, h.clearedCon+h.clearedRec, "a piped 'y' is not the person answering")
	assert.Empty(t, out)
}

func TestRunClearAsksAndClearsOnYes(t *testing.T) {
	for _, answer := range []string{"y\n", "Y\n", "yes\n", "YES\n", " yes \n", "s\n", "sim\n", "y"} {
		h := &historyFake{}
		useHistoryFake(t, h)

		out, err := clearRun(t, &util.ClearRequest{Continue: true}, answer, true)

		require.NoError(t, err, answer)
		assert.Equal(t, 1, h.clearedCon, "%q clears", answer)
		assert.Contains(t, out, "Clear the CONTINUE list? This cannot be undone. [y/N] ", answer)
		assert.Contains(t, out, "Cleared 7 episodes from CONTINUE.", answer)
	}
}

func TestRunClearKeepsEverythingOnAnythingButYes(t *testing.T) {
	for _, answer := range []string{"n\n", "no\n", "\n", "", "maybe\n", "yy\n", "ye\n", "nope\n"} {
		h := &historyFake{}
		useHistoryFake(t, h)

		out, err := clearRun(t, &util.ClearRequest{Continue: true, Recent: true}, answer, true)

		require.NoError(t, err, "declining is not a failure: %q", answer)
		assert.Zero(t, h.clearedCon+h.clearedRec, "%q clears nothing", answer)
		assert.Contains(t, out, "Clear the CONTINUE list and the RECENT searches? This cannot be undone. [y/N] ")
		assert.Contains(t, out, "Nothing was cleared.\n", answer)
		assert.NotContains(t, out, "Cleared ", answer)
	}
}

func TestRunClearTheQuestionNamesTheListAsked(t *testing.T) {
	h := &historyFake{}
	useHistoryFake(t, h)

	out, err := clearRun(t, &util.ClearRequest{Recent: true}, "n\n", true)

	require.NoError(t, err)
	assert.Contains(t, out, "Clear the RECENT searches? ")
	assert.NotContains(t, out, "CONTINUE")
}

func TestRunClearReportsEveryFailureAndStillTriesTheOtherList(t *testing.T) {
	h := &historyFake{clearErr: errors.New("database is locked")}
	useHistoryFake(t, h)

	out, err := clearRun(t, &util.ClearRequest{Continue: true, Recent: true, Yes: true}, "", false)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "clear CONTINUE: database is locked")
	assert.Contains(t, err.Error(), "clear RECENT: database is locked")
	assert.Equal(t, 1, h.clearedCon)
	assert.Equal(t, 1, h.clearedRec, "one list failing does not leave the other uncleared")
	assert.NotContains(t, out, "Cleared ")
}

func TestRunClearKeepsTheCauseOfAFailureInspectable(t *testing.T) {
	h := &historyFake{clearErr: tracking.ErrHistoryUnavailable}
	useHistoryFake(t, h)

	_, err := clearRun(t, &util.ClearRequest{Recent: true, Yes: true}, "", false)

	assert.ErrorIs(t, err, tracking.ErrHistoryUnavailable)
}

func TestRunClearOneListFailingStillReportsTheOtherCleared(t *testing.T) {
	h := &historyFake{}
	useHistoryFake(t, h)
	clearRecentFn = func() (int, error) { return 0, errors.New("disk full") }

	out, err := clearRun(t, &util.ClearRequest{Continue: true, Recent: true, Yes: true}, "", false)

	require.ErrorContains(t, err, "clear RECENT: disk full")
	assert.NotContains(t, err.Error(), "CONTINUE")
	assert.Equal(t, "Cleared 7 episodes from CONTINUE.\n", out)
}

func TestDescribeClear(t *testing.T) {
	assert.Equal(t, "the CONTINUE list", describeClear(&util.ClearRequest{Continue: true}))
	assert.Equal(t, "the RECENT searches", describeClear(&util.ClearRequest{Recent: true}))
	assert.Equal(t, "the CONTINUE list and the RECENT searches", describeClear(&util.ClearRequest{Continue: true, Recent: true}))
}

func TestAnsweredYes(t *testing.T) {
	for in, want := range map[string]bool{
		"y": true, "y\n": true, "Y\r\n": true, "yes": true, "Sim\n": true, "s\n": true, "  y  \n": true,
		"": false, "\n": false, "n\n": false, "no": false, "yeah\n": false, "y y\n": false,
	} {
		assert.Equal(t, want, answeredYes(strings.NewReader(in)), "%q", in)
	}
	assert.False(t, answeredYes(errReader{}), "an input that fails answers no")
}

type errReader struct{}

func (errReader) Read([]byte) (int, error) { return 0, errors.New("stdin closed") }

// ---- HandleClearRequest is runClear on the real terminal -----------------------

func captureStdout(t *testing.T) func() string {
	t.Helper()
	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stdout
	os.Stdout = w
	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.String()
	}()
	t.Cleanup(func() { os.Stdout = orig })
	return func() string {
		os.Stdout = orig
		require.NoError(t, w.Close())
		out := <-done
		_ = r.Close()
		return out
	}
}

func TestHandleClearRequestUsesTheStoredRequest(t *testing.T) {
	h := &historyFake{}
	useHistoryFake(t, h)
	prev := util.GlobalClearRequest
	util.GlobalClearRequest = &util.ClearRequest{Recent: true, Yes: true}
	t.Cleanup(func() { util.GlobalClearRequest = prev })
	read := captureStdout(t)

	err := HandleClearRequest()
	out := read()

	require.NoError(t, err)
	assert.Equal(t, 1, h.clearedRec)
	assert.Zero(t, h.clearedCon)
	assert.Equal(t, "Cleared 4 searches from RECENT.\n", out)
}

func TestHandleClearRequestWithoutARequestIsAnError(t *testing.T) {
	prev := util.GlobalClearRequest
	util.GlobalClearRequest = nil
	t.Cleanup(func() { util.GlobalClearRequest = prev })

	require.ErrorContains(t, HandleClearRequest(), "nothing to clear")
}

func TestHandleClearRequestAsksOnlyWhereStdinIsATerminal(t *testing.T) {
	h := &historyFake{}
	useHistoryFake(t, h)
	prev, prevTerm := util.GlobalClearRequest, stdinIsTerminal
	util.GlobalClearRequest = &util.ClearRequest{Continue: true}
	stdinIsTerminal = func() bool { return false }
	t.Cleanup(func() { util.GlobalClearRequest, stdinIsTerminal = prev, prevTerm })

	err := HandleClearRequest()

	require.ErrorContains(t, err, "--yes")
	assert.Zero(t, h.clearedCon)
}
