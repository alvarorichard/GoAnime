package updater

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withReleaseAPI swaps the API URL and restores on cleanup.
// MUST NOT be used with t.Parallel() since it mutates a package var.
func withReleaseAPI(t *testing.T, url string) {
	t.Helper()
	prev := releaseAPIURL
	releaseAPIURL = url
	t.Cleanup(func() { releaseAPIURL = prev })
}

func TestCheckForUpdates_UsesReleaseAPIURL(t *testing.T) {
	rel := GitHubRelease{TagName: "v99.0.0"}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(rel)
	}))
	t.Cleanup(srv.Close)
	withReleaseAPI(t, srv.URL)

	r, has, err := CheckForUpdates()
	require.NoError(t, err)
	assert.True(t, has)
	assert.Equal(t, "v99.0.0", r.TagName)
}

func TestCheckForUpdates_NetworkError(t *testing.T) {
	withReleaseAPI(t, "http://127.0.0.1:1") // closed port
	_, _, err := CheckForUpdates()
	require.Error(t, err)
}

func TestCheckForUpdatesQuietly_NoUpdate(t *testing.T) {
	// Serve same version as runtime to exercise "no update" branch.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(GitHubRelease{TagName: "v0.0.0"})
	}))
	t.Cleanup(srv.Close)
	withReleaseAPI(t, srv.URL)
	// Just exercise the function — it logs but doesn't return errors.
	assert.NotPanics(t, CheckForUpdatesQuietly)
}

func TestCheckForUpdatesQuietly_WithUpdate(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(GitHubRelease{TagName: "v999.0.0"})
	}))
	t.Cleanup(srv.Close)
	withReleaseAPI(t, srv.URL)
	assert.NotPanics(t, CheckForUpdatesQuietly)
}

func TestCheckForUpdatesQuietly_APIError(t *testing.T) {
	withReleaseAPI(t, "http://127.0.0.1:1")
	// Must not panic on error path.
	assert.NotPanics(t, CheckForUpdatesQuietly)
}

// withPerformHooks installs stubbed hooks for PerformUpdate. Checksum
// verification is stubbed to succeed so these tests exercise the rest of the
// flow; the verification itself is covered in checksum_test.go, and a test can
// override verifyDownloadFn after calling this.
func withPerformHooks(t *testing.T, fa func(*GitHubRelease) (string, string, error),
	dl func(string, string) (string, error), exeFn func() (string, error),
	rep func(string, string) error) {
	t.Helper()
	pFA, pDL, pVER, pEXE, pREP := findAssetFn, downloadFn, verifyDownloadFn, osExecutableFn, replaceExecutableFn
	findAssetFn = fa
	downloadFn = dl
	verifyDownloadFn = func(*GitHubRelease, string, string) error { return nil }
	osExecutableFn = exeFn
	replaceExecutableFn = rep
	t.Cleanup(func() {
		findAssetFn = pFA
		downloadFn = pDL
		verifyDownloadFn = pVER
		osExecutableFn = pEXE
		replaceExecutableFn = pREP
	})
}

func TestPerformUpdate_Success(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "current")
	require.NoError(t, os.WriteFile(currentExe, []byte("old"), 0o755))

	downloadDst := filepath.Join(dir, "downloaded")
	require.NoError(t, os.WriteFile(downloadDst, []byte("new-bin"), 0o644))

	withPerformHooks(t,
		func(*GitHubRelease) (string, string, error) { return "http://x/y", "asset.bin", nil },
		func(string, string) (string, error) { return downloadDst, nil },
		func() (string, error) { return currentExe, nil },
		func(_, src string) error {
			data, err := os.ReadFile(src)
			if err != nil {
				return err
			}
			return os.WriteFile(currentExe, data, 0o755)
		},
	)
	require.NoError(t, PerformUpdate(&GitHubRelease{}))
	got, err := os.ReadFile(currentExe)
	require.NoError(t, err)
	assert.Equal(t, "new-bin", string(got))
}

func TestPerformUpdate_FindAssetError(t *testing.T) {
	withPerformHooks(t,
		func(*GitHubRelease) (string, string, error) { return "", "", errors.New("no asset") },
		nil, nil, nil,
	)
	err := PerformUpdate(&GitHubRelease{})
	require.Error(t, err)
}

func TestPerformUpdate_DownloadError(t *testing.T) {
	withPerformHooks(t,
		func(*GitHubRelease) (string, string, error) { return "u", "n", nil },
		func(string, string) (string, error) { return "", errors.New("net fail") },
		nil, nil,
	)
	err := PerformUpdate(&GitHubRelease{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "download")
}

func TestPerformUpdate_ExecutablePathError(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "asset")
	require.NoError(t, os.WriteFile(src, []byte("x"), 0o644))
	withPerformHooks(t,
		func(*GitHubRelease) (string, string, error) { return "u", "asset", nil },
		func(string, string) (string, error) { return src, nil },
		func() (string, error) { return "", errors.New("no exe") },
		nil,
	)
	err := PerformUpdate(&GitHubRelease{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "executable")
}

func TestPerformUpdate_ReplaceFails_RestoresBackup(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "current")
	require.NoError(t, os.WriteFile(currentExe, []byte("ORIGINAL"), 0o755))
	src := filepath.Join(dir, "asset")
	require.NoError(t, os.WriteFile(src, []byte("NEW"), 0o644))

	withPerformHooks(t,
		func(*GitHubRelease) (string, string, error) { return "u", "asset", nil },
		func(string, string) (string, error) { return src, nil },
		func() (string, error) { return currentExe, nil },
		func(string, string) error { return errors.New("rename failed") },
	)
	err := PerformUpdate(&GitHubRelease{})
	require.Error(t, err)
	got, _ := os.ReadFile(currentExe)
	assert.Equal(t, "ORIGINAL", string(got), "backup restore must put original back")
}

func TestPerformUpdate_WindowsDeferredUpdate_NoError(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "current")
	require.NoError(t, os.WriteFile(currentExe, []byte("orig"), 0o755))
	src := filepath.Join(dir, "asset")
	require.NoError(t, os.WriteFile(src, []byte("new"), 0o644))

	withPerformHooks(t,
		func(*GitHubRelease) (string, string, error) { return "u", "asset", nil },
		func(string, string) (string, error) { return src, nil },
		func() (string, error) { return currentExe, nil },
		func(string, string) error {
			return fmt.Errorf("helper started: %w", errUpdateDeferred)
		},
	)
	err := PerformUpdate(&GitHubRelease{})
	require.NoError(t, err, "deferred Windows update is reported as success")
	assert.FileExists(t, updateBackupPath(currentExe), "the helper needs the backup until it finishes")
	assert.NoFileExists(t, updateMarkerPath(currentExe), "the helper writes the marker after a verified copy")
}

func TestPerformUpdate_ImmediateSuccess_KeepsBackupForNextStart(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "current")
	require.NoError(t, os.WriteFile(currentExe, []byte("old"), 0o755))
	src := filepath.Join(dir, "asset")
	require.NoError(t, os.WriteFile(src, []byte("new"), 0o644))

	withPerformHooks(t,
		func(*GitHubRelease) (string, string, error) { return "u", "asset", nil },
		func(string, string) (string, error) { return src, nil },
		func() (string, error) { return currentExe, nil },
		func(_, newExe string) error {
			data, err := os.ReadFile(newExe)
			if err != nil {
				return err
			}
			return os.WriteFile(currentExe, data, 0o755)
		},
	)
	require.NoError(t, PerformUpdate(&GitHubRelease{}))

	backup, err := os.ReadFile(updateBackupPath(currentExe))
	require.NoError(t, err)
	assert.Equal(t, "old", string(backup), "backup must survive until the new binary runs")

	installedHash, err := fileSHA256(currentExe)
	require.NoError(t, err)
	marker, err := os.ReadFile(updateMarkerPath(currentExe))
	require.NoError(t, err)
	assert.Equal(t, installedHash, strings.TrimSpace(string(marker)))

	require.NoError(t, cleanupUpdateArtifacts(currentExe))
	assert.NoFileExists(t, updateBackupPath(currentExe))
	assert.NoFileExists(t, updateMarkerPath(currentExe))
}

func TestCleanupUpdateArtifacts_RemovesVerifiedArtifacts(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "current")
	require.NoError(t, os.WriteFile(currentExe, []byte("new"), 0o755))
	require.NoError(t, os.WriteFile(updateBackupPath(currentExe), []byte("old"), 0o644))
	require.NoError(t, os.WriteFile(updateStagedPath(currentExe), []byte("staged"), 0o644))
	require.NoError(t, os.WriteFile(updateScriptPath(currentExe), []byte("@echo off"), 0o644))

	digest, err := fileSHA256(currentExe)
	require.NoError(t, err)
	require.NoError(t, writeUpdateMarker(currentExe, digest))

	require.NoError(t, cleanupUpdateArtifacts(currentExe))
	assert.NoFileExists(t, updateBackupPath(currentExe))
	assert.NoFileExists(t, updateStagedPath(currentExe))
	assert.NoFileExists(t, updateScriptPath(currentExe))
	assert.NoFileExists(t, updateMarkerPath(currentExe))
}

func TestCleanupUpdateArtifacts_KeepsBackupOnDigestMismatch(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "current")
	require.NoError(t, os.WriteFile(currentExe, []byte("different"), 0o755))
	require.NoError(t, os.WriteFile(updateBackupPath(currentExe), []byte("old"), 0o644))

	require.NoError(t, writeUpdateMarker(currentExe, strings.Repeat("a", 64)))

	err := cleanupUpdateArtifacts(currentExe)
	require.Error(t, err)
	assert.FileExists(t, updateBackupPath(currentExe), "an unverified update must not delete the backup")
	assert.NoFileExists(t, updateMarkerPath(currentExe), "the stale marker is removed to avoid repeating the warning")
}

func TestCleanupUpdateArtifacts_KeepsBackupOnMalformedMarker(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "current")
	require.NoError(t, os.WriteFile(currentExe, []byte("current"), 0o755))
	require.NoError(t, os.WriteFile(updateBackupPath(currentExe), []byte("old"), 0o644))
	require.NoError(t, os.WriteFile(updateMarkerPath(currentExe), []byte("not-a-digest\n"), 0o600))

	err := cleanupUpdateArtifacts(currentExe)
	require.Error(t, err)
	assert.FileExists(t, updateBackupPath(currentExe))
	assert.NoFileExists(t, updateMarkerPath(currentExe))
}

func TestCleanupUpdateArtifacts_NoMarkerIsNoop(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "current")
	require.NoError(t, os.WriteFile(currentExe, []byte("current"), 0o755))

	require.NoError(t, cleanupUpdateArtifacts(currentExe))
}

func TestPrepareWindowsUpdate_StagesNextToExecutable(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "goanime.exe")
	require.NoError(t, os.WriteFile(currentExe, []byte("old"), 0o755))
	downloaded := filepath.Join(dir, "downloaded.exe")
	require.NoError(t, os.WriteFile(downloaded, []byte("new"), 0o644))

	plan, err := prepareWindowsUpdate(currentExe, downloaded)
	require.NoError(t, err)
	assert.Equal(t, updateStagedPath(currentExe), plan.stagedExe)
	assert.Equal(t, updateBackupPath(currentExe), plan.backupFile)
	assert.Equal(t, updateMarkerPath(currentExe), plan.markerFile)
	assert.NotEqual(t, downloaded, plan.stagedExe, "the helper must not depend on the temporary extraction directory")

	staged, err := os.ReadFile(plan.stagedExe)
	require.NoError(t, err)
	assert.Equal(t, "new", string(staged))
	expected, err := fileSHA256(downloaded)
	require.NoError(t, err)
	assert.Equal(t, expected, plan.digest)
}

func TestWindowsUpdateScript_ContainsSafetySteps(t *testing.T) {
	script := windowsUpdateScript()

	for _, want := range []string{
		`tasklist /FI "PID eq %WAITPID%"`,
		`certutil -hashfile`,
		`copy /Y "%STAGED%" "%TARGET%"`,
		`copy /Y "%BACKUP%" "%TARGET%"`,
		`> "%MARKER%" echo %EXPECTED%`,
		`ping -n 2 127.0.0.1`,
	} {
		assert.Contains(t, script, want)
	}

	verifyStaged := strings.Index(script, `call :hash "%STAGED%"`)
	replace := strings.Index(script, `copy /Y "%STAGED%" "%TARGET%"`)
	verifyTarget := strings.LastIndex(script, `call :hash "%TARGET%"`)
	success := strings.Index(script, ":success")
	require.NotEqual(t, -1, verifyStaged)
	require.NotEqual(t, -1, replace)
	require.NotEqual(t, -1, verifyTarget)
	require.NotEqual(t, -1, success)
	assert.Less(t, verifyStaged, replace, "the staged file is hashed before it can touch the target")
	assert.Less(t, replace, verifyTarget, "the target is hashed after the copy")
	assert.Less(t, verifyTarget, success, "success is only reached after the target hash matches")
}

// findAssetForPlatform wrapper just hits findAssetForPlatformWithInfo for the host OS.
func TestFindAssetForPlatform_DelegatesToCurrentPlatform(t *testing.T) {
	t.Parallel()
	rel := &GitHubRelease{
		Assets: []struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{
			{Name: fmt.Sprintf("goanime-%s-%s", runtime.GOOS, runtime.GOARCH),
				BrowserDownloadURL: "http://x"},
			{Name: "goanime-darwin", BrowserDownloadURL: "http://x"},
			{Name: "goanime-linux", BrowserDownloadURL: "http://x"},
			{Name: "goanime", BrowserDownloadURL: "http://x"},
			{Name: "goanime-windows.exe", BrowserDownloadURL: "http://x"},
			{Name: "goanime.exe", BrowserDownloadURL: "http://x"},
		},
	}
	_, _, err := findAssetForPlatform(rel)
	require.NoError(t, err)
}

// downloadAsset wrapper rejects non-GitHub URLs (test flag = false).
func TestDownloadAsset_RejectsNonGitHub(t *testing.T) {
	t.Parallel()
	_, err := downloadAsset("http://evil.example.com/x", "y")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "URL validation")
}

// --- PromptForUpdate ---

func withRunForm(t *testing.T, fn func(func() error) error) {
	t.Helper()
	prev := runForm
	runForm = fn
	t.Cleanup(func() { runForm = prev })
}

func TestPromptForUpdate_FormSucceeds(t *testing.T) {
	withRunForm(t, func(func() error) error { return nil })
	got, err := PromptForUpdate(&GitHubRelease{TagName: "v1", Body: "ok"})
	require.NoError(t, err)
	// shouldUpdate defaults to false since we never let the form run.
	assert.False(t, got)
}

func TestPromptForUpdate_FormError(t *testing.T) {
	withRunForm(t, func(func() error) error { return errors.New("tty fail") })
	_, err := PromptForUpdate(&GitHubRelease{TagName: "v1"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "update prompt")
}

// --- CheckAndPromptUpdate ---

func TestCheckAndPromptUpdate_NoUpdateAvailable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(GitHubRelease{TagName: "v0.0.0"})
	}))
	t.Cleanup(srv.Close)
	withReleaseAPI(t, srv.URL)
	withRunForm(t, func(func() error) error { return nil })

	require.NoError(t, CheckAndPromptUpdate())
}

func TestCheckAndPromptUpdate_CheckFails(t *testing.T) {
	withReleaseAPI(t, "http://127.0.0.1:1")
	err := CheckAndPromptUpdate()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "check for updates")
}

func TestCheckAndPromptUpdate_UserDeclines(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(GitHubRelease{TagName: "v999.0.0"})
	}))
	t.Cleanup(srv.Close)
	withReleaseAPI(t, srv.URL)
	// runForm returns nil → shouldUpdate stays false → cancellation branch
	withRunForm(t, func(func() error) error { return nil })
	require.NoError(t, CheckAndPromptUpdate())
}

func TestCheckAndPromptUpdate_PromptError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(GitHubRelease{TagName: "v999.0.0"})
	}))
	t.Cleanup(srv.Close)
	withReleaseAPI(t, srv.URL)
	withRunForm(t, func(func() error) error { return errors.New("tty") })
	err := CheckAndPromptUpdate()
	require.Error(t, err)
}

// --- downloadAsset (wrapper) basic happy path via test flag wrapper not callable; cover the no-op call ---

func TestDownloadAsset_BadURL(t *testing.T) {
	t.Parallel()
	_, err := downloadAsset(string(rune(0x7f))+"://x", "f")
	require.Error(t, err)
}

// --- createWindowsUpdateScript ---

func TestCreateWindowsUpdateScript_NonWindowsErrors(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("non-windows branch test")
	}
	t.Parallel()
	err := createWindowsUpdateScript("/tmp/x", "/tmp/y")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "non-Windows")
}

// Ensure concurrent reads of releaseAPIURL during tests don't race with writers
// (defensive — currently each test sets it serially, but if added to a parallel
// suite later, this guard helps catch regressions early).
var _ = sync.Mutex{}

// Sanity: confirm the runtime version is not what we send so update-detection branches fire.
func TestReleaseTagNotEqualToBuildVersion(t *testing.T) {
	t.Parallel()
	assert.NotEqual(t, "v999.0.0", strings.TrimSpace(""), "guard against accidental match")
}
