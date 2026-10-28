//go:build windows

package updater

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// runWindowsHelper executes the post-exit helper synchronously. PID 999999 is
// not expected to exist, so the wait loop finishes immediately instead of
// waiting for the test process to exit. pathPrefix, when set, is prepended to
// PATH inside a wrapper batch: Wine ignores the Unix PATH variable, so setting
// it from Go would not reach cmd there.
func runWindowsHelper(t *testing.T, plan windowsUpdatePlan, scriptPath, pathPrefix string) ([]byte, error) {
	t.Helper()

	if pathPrefix != "" {
		wrapper := filepath.Join(filepath.Dir(scriptPath), "run-helper.bat")
		content := "@echo off\r\n" +
			"set \"PATH=" + pathPrefix + ";%PATH%\"\r\n" +
			"call \"" + scriptPath + "\"\r\n" +
			"exit /b %ERRORLEVEL%\r\n"
		require.NoError(t, os.WriteFile(wrapper, []byte(content), 0o600))
		scriptPath = wrapper
	}

	cmd := exec.Command("cmd", "/C", scriptPath)
	cmd.Env = append(os.Environ(),
		"GOANIME_UPDATE_TARGET="+plan.currentExe,
		"GOANIME_UPDATE_STAGED="+plan.stagedExe,
		"GOANIME_UPDATE_BACKUP="+plan.backupFile,
		"GOANIME_UPDATE_MARKER="+plan.markerFile,
		"GOANIME_UPDATE_SHA256="+plan.digest,
		"GOANIME_UPDATE_PID=999999",
	)
	return cmd.CombinedOutput()
}

func newHelperFixture(t *testing.T) (windowsUpdatePlan, string) {
	t.Helper()

	dir := t.TempDir()
	target := filepath.Join(dir, "goanime.exe")
	downloaded := filepath.Join(dir, "downloaded.exe")
	require.NoError(t, os.WriteFile(target, []byte("OLD"), 0o755))
	require.NoError(t, os.WriteFile(downloaded, []byte("NEW"), 0o755))

	plan, err := prepareWindowsUpdate(target, downloaded)
	require.NoError(t, err)

	scriptPath := filepath.Join(dir, "helper.bat")
	require.NoError(t, os.WriteFile(scriptPath, []byte(windowsUpdateScript()), 0o600))
	return plan, scriptPath
}

// requireRealCertutil skips the test where certutil cannot hash a file, which
// is the case under Wine. GitHub's windows-latest runners have the real tool.
func requireRealCertutil(t *testing.T) {
	t.Helper()

	out, err := exec.Command("certutil", "-hashfile", os.Args[0], "SHA256").Output()
	if err != nil {
		t.Skipf("certutil is unavailable: %v", err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimSpace(strings.ReplaceAll(line, " ", ""))
		if len(line) == 64 {
			if _, err := parseUpdateMarker([]byte(line)); err == nil {
				return
			}
		}
	}
	t.Skip("certutil produced no SHA-256 digest (Wine)")
}

func TestWindowsUpdateHelperInstallsAndMarks(t *testing.T) {
	requireRealCertutil(t)

	plan, scriptPath := newHelperFixture(t)

	out, err := runWindowsHelper(t, plan, scriptPath, "")
	require.NoError(t, err, "helper output:\n%s", out)

	installed, err := os.ReadFile(plan.currentExe)
	require.NoError(t, err)
	assert.Equal(t, "NEW", string(installed), "the verified update must be installed")

	backup, err := os.ReadFile(plan.backupFile)
	require.NoError(t, err)
	assert.Equal(t, "OLD", string(backup), "the previous binary stays available for rollback")

	marker, err := os.ReadFile(plan.markerFile)
	require.NoError(t, err)
	assert.Equal(t, plan.digest, strings.TrimSpace(string(marker)),
		"the marker records the digest of the installed binary")

	assert.NoFileExists(t, plan.stagedExe, "the staged file is removed after a verified copy")
	assert.Eventually(t, func() bool {
		_, statErr := os.Stat(scriptPath)
		return os.IsNotExist(statErr)
	}, 5*time.Second, 100*time.Millisecond, "the helper removes itself")
}

func TestWindowsUpdateHelperRestoresOnHashMismatch(t *testing.T) {
	plan, scriptPath := newHelperFixture(t)

	// A fake certutil reports the expected hash for the staged file and a
	// wrong hash for the installed file, forcing the post-copy verification to
	// fail. A .bat earlier in PATH wins over certutil.exe in System32.
	fakeDir := filepath.Join(t.TempDir(), "fakebin")
	require.NoError(t, os.Mkdir(fakeDir, 0o755))
	fakeCertutil := "@echo off\r\n" +
		"echo SHA256 hash of file %2:\r\n" +
		"if /I \"%~nx2\"==\"goanime.exe.new\" (\r\n" +
		"  echo %GOANIME_UPDATE_SHA256%\r\n" +
		") else (\r\n" +
		"  echo 0000000000000000000000000000000000000000000000000000000000000000\r\n" +
		")\r\n" +
		"echo CertUtil: -hashfile command completed successfully.\r\n"
	require.NoError(t, os.WriteFile(filepath.Join(fakeDir, "certutil.bat"), []byte(fakeCertutil), 0o600))

	out, err := runWindowsHelper(t, plan, scriptPath, fakeDir)
	require.Error(t, err, "the helper must fail when the installed hash does not match; output:\n%s", out)

	installed, err := os.ReadFile(plan.currentExe)
	require.NoError(t, err)
	assert.Equal(t, "OLD", string(installed), "a failed verification restores the previous binary")

	backup, err := os.ReadFile(plan.backupFile)
	require.NoError(t, err)
	assert.Equal(t, "OLD", string(backup), "the backup is kept for the user")

	assert.NoFileExists(t, plan.markerFile, "no marker is written without a verified copy")
	assert.NoFileExists(t, plan.stagedExe, "the staged file is removed after the failure")
	assert.Eventually(t, func() bool {
		_, statErr := os.Stat(scriptPath)
		return os.IsNotExist(statErr)
	}, 5*time.Second, 100*time.Millisecond, "the helper removes itself")
}
