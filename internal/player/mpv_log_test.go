package player

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestJoinMPVDiagnostics(t *testing.T) {
	assert.Equal(t, "", joinMPVDiagnostics("", ""))
	assert.Equal(t, "log", joinMPVDiagnostics("", "log"))
	assert.Equal(t, "err\nlog", joinMPVDiagnostics("err\n", "log"))
}

func TestTailLines(t *testing.T) {
	assert.Equal(t, "", tailLines("", 3))
	assert.Equal(t, "a\nb", tailLines("a\nb\n", 3))
	assert.Equal(t, "c\nd", tailLines("a\nb\nc\nd", 2))
}

func TestMPVExitCode(t *testing.T) {
	assert.Equal(t, 0, mpvExitCode(nil))
	assert.Equal(t, -1, mpvExitCode(errors.New("not an exit error")))
	if runtime.GOOS == "windows" {
		t.Skip("needs sh")
	}
	err := exec.Command("sh", "-c", "exit 7").Run()
	assert.Equal(t, 7, mpvExitCode(err))
}

// reapMPV must hand back the exit code and mpv's own log, and delete the file.
func TestReapMPVReadsLogAndRemovesIt(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("needs sh")
	}
	logPath := filepath.Join(t.TempDir(), "mpv.log")
	require.NoError(t, os.WriteFile(logPath, []byte("[vo] Error: no usable video output\n"), 0o600))
	var stderr bytes.Buffer
	cmd := exec.Command("sh", "-c", "exit 1")
	require.NoError(t, cmd.Start())

	ex := <-reapMPV(cmd, logPath, &stderr, time.Now())
	assert.Equal(t, 1, ex.code)
	assert.Contains(t, ex.log, "no usable video output")
	_, err := os.Stat(logPath)
	assert.True(t, os.IsNotExist(err), "the log file must be removed once mpv has exited")
}

func TestMPVProblemLinesPrefersErrorsOverVerboseNoise(t *testing.T) {
	log := "[ 0.0][v][cplayer] Setting option\n[ 0.1][d][global] config path\n[ 0.2][e][vo] no usable video output\n[ 0.3][v][x] more\n[ 0.4][w][ao] slow\n"
	got := mpvProblemLines(log, 15)
	assert.Equal(t, "[ 0.2][e][vo] no usable video output\n[ 0.4][w][ao] slow", got)

	plain := "one\ntwo\nthree"
	assert.Equal(t, "two\nthree", mpvProblemLines(plain, 2), "no levelled lines: fall back to the tail")
}
