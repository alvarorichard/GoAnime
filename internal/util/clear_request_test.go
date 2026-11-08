package util

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --clear-continue, --clear-recent and --clear-history: which list each one
// empties, how the question is skipped, and that the help says so.

func TestClearRequestFrom(t *testing.T) {
	for name, tc := range map[string]struct {
		cont, recent, history, yes bool
		want                       *ClearRequest
	}{
		"nothing asked":            {want: nil},
		"--yes alone asks nothing": {yes: true, want: nil},
		"continue only":            {cont: true, want: &ClearRequest{Continue: true}},
		"recent only":              {recent: true, want: &ClearRequest{Recent: true}},
		"history is both":          {history: true, want: &ClearRequest{Continue: true, Recent: true}},
		"both lists by name":       {cont: true, recent: true, want: &ClearRequest{Continue: true, Recent: true}},
		"history and continue":     {history: true, cont: true, want: &ClearRequest{Continue: true, Recent: true}},
		"yes skips the question":   {recent: true, yes: true, want: &ClearRequest{Recent: true, Yes: true}},
		"history with yes":         {history: true, yes: true, want: &ClearRequest{Continue: true, Recent: true, Yes: true}},
	} {
		assert.Equal(t, tc.want, clearRequestFrom(tc.cont, tc.recent, tc.history, tc.yes), name)
	}
}

func parseClearArgs(t *testing.T, args ...string) (string, error) {
	t.Helper()
	snapshotOsArgs(t)
	snapshotGlobalRequest(t)
	GlobalClearRequest = nil
	os.Args = append([]string{"goanime"}, args...)
	return FlagParser()
}

func TestFlagParserClearFlags(t *testing.T) {
	for name, tc := range map[string]struct {
		args []string
		want ClearRequest
	}{
		"--clear-continue":           {[]string{"--clear-continue"}, ClearRequest{Continue: true}},
		"--clear-recent":             {[]string{"--clear-recent"}, ClearRequest{Recent: true}},
		"--clear-history":            {[]string{"--clear-history"}, ClearRequest{Continue: true, Recent: true}},
		"both flags":                 {[]string{"--clear-continue", "--clear-recent"}, ClearRequest{Continue: true, Recent: true}},
		"--yes":                      {[]string{"--clear-recent", "--yes"}, ClearRequest{Recent: true, Yes: true}},
		"-y":                         {[]string{"-y", "--clear-continue"}, ClearRequest{Continue: true, Yes: true}},
		"a title does not cancel it": {[]string{"--clear-recent", "naruto"}, ClearRequest{Recent: true}},
	} {
		name, tc := name, tc
		t.Run(name, func(t *testing.T) {
			title, err := parseClearArgs(t, tc.args...)

			require.ErrorIs(t, err, ErrClearRequested)
			assert.Empty(t, title, "it clears and exits: there is nothing to search")
			require.NotNil(t, GlobalClearRequest)
			assert.Equal(t, tc.want, *GlobalClearRequest)
		})
	}
}

func TestFlagParserYesAloneClearsNothing(t *testing.T) {
	for _, flag := range []string{"--yes", "-y"} {
		title, err := parseClearArgs(t, flag)

		require.NoError(t, err, flag)
		assert.Empty(t, title, flag)
		assert.Nil(t, GlobalClearRequest, flag)
	}
}

func TestFlagParserWithoutClearFlagsStoresNoRequest(t *testing.T) {
	_, err := parseClearArgs(t, "naruto")

	require.NoError(t, err)
	assert.Nil(t, GlobalClearRequest)
}

func TestErrClearRequestedIsItsOwnSignal(t *testing.T) {
	for _, other := range []error{ErrUpdateRequested, ErrDownloadRequested, ErrUpscaleRequested, ErrMovieDownloadRequested, ErrHelpRequested} {
		assert.NotErrorIs(t, ErrClearRequested, other)
	}
}

func TestTheHelpListsTheClearFlags(t *testing.T) {
	r, w, err := os.Pipe()
	require.NoError(t, err)
	orig := os.Stdout
	os.Stdout = w
	done := make(chan []byte, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		done <- buf.Bytes()
	}()

	ShowBeautifulHelp()

	os.Stdout = orig
	require.NoError(t, w.Close())
	help := ansi.Strip(string(<-done))
	_ = r.Close()

	for _, want := range []string{"--clear-continue", "--clear-recent", "--clear-history", "--yes / -y"} {
		assert.Contains(t, help, want)
	}
}

// ---- the clear commands do not probe the terminal ------------------------------

// countProbes replaces the terminal probes with a counter for one test.
func countProbes(t *testing.T) *int {
	t.Helper()
	var n int
	prev := probeTerminal
	probeTerminal = func() { n++ }
	t.Cleanup(func() { probeTerminal = prev })
	return &n
}

func TestTheClearCommandsDoNotProbeTheTerminal(t *testing.T) {
	for _, args := range [][]string{
		{"--clear-continue"}, {"--clear-recent"}, {"--clear-history"}, {"--clear-history", "--yes"},
	} {
		probes := countProbes(t)

		_, err := parseClearArgs(t, args...)

		require.ErrorIs(t, err, ErrClearRequested, "%v", args)
		assert.Zero(t, *probes, "%v: a probe nobody answers leaves a reader on stdin that swallows the y/n", args)
	}
}

func TestEverythingElseStillProbesTheTerminalOnce(t *testing.T) {
	for _, args := range [][]string{{"naruto"}, {}, {"--yes"}, {"--plain", "naruto"}} {
		probes := countProbes(t)

		_, err := parseClearArgs(t, args...)

		require.NoError(t, err, "%v", args)
		assert.Equal(t, 1, *probes, "%v", args)
	}
}
