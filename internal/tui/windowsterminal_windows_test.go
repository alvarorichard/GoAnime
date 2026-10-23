//go:build windows

package tui

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	bundledWindowsTerminalSHA256  = "bf3ef2012f6c44d8340a4c58125acc9498d19b580f9890dc043cdf831852e796"
	bundledWindowsTerminalSize    = 12127272
	bundledWindowsTerminalVersion = "1.25.2733.0"
)

func TestPortableWindowsTerminalPath(t *testing.T) {
	local := t.TempDir()
	t.Setenv("LOCALAPPDATA", local)

	_, ok := portableWindowsTerminalPath()
	assert.False(t, ok, "no copy exists before the first install")

	wt := filepath.Join(local, "GoAnime", "WindowsTerminal", "current", "wt.exe")
	require.NoError(t, os.MkdirAll(filepath.Dir(wt), 0o750))
	require.NoError(t, os.WriteFile(wt, []byte("wt"), 0o600))

	got, ok := portableWindowsTerminalPath()
	require.True(t, ok)
	assert.Equal(t, wt, got)

	got, ok = windowsTerminalPath()
	require.True(t, ok)
	assert.Equal(t, wt, got, "the private copy is used when the app alias is missing")
}

func TestInstalledAliasWinsOverPortableCopy(t *testing.T) {
	local := t.TempDir()
	t.Setenv("LOCALAPPDATA", local)

	alias := filepath.Join(local, "Microsoft", "WindowsApps", "wt.exe")
	require.NoError(t, os.MkdirAll(filepath.Dir(alias), 0o750))
	require.NoError(t, os.WriteFile(alias, []byte("wt"), 0o600))

	portable := filepath.Join(local, "GoAnime", "WindowsTerminal", "current", "wt.exe")
	require.NoError(t, os.MkdirAll(filepath.Dir(portable), 0o750))
	require.NoError(t, os.WriteFile(portable, []byte("wt"), 0o600))

	got, ok := windowsTerminalPath()
	require.True(t, ok)
	assert.Equal(t, alias, got, "a real installation takes precedence over the fallback copy")
}

func TestPortableWindowsTerminalRootRequiresLocalAppData(t *testing.T) {
	t.Setenv("LOCALAPPDATA", "")

	_, err := portableWindowsTerminalRoot()
	require.Error(t, err)
}

func TestInstallPortableWindowsTerminalUsesCachedCopy(t *testing.T) {
	local := t.TempDir()
	t.Setenv("LOCALAPPDATA", local)

	wt := filepath.Join(local, "GoAnime", "WindowsTerminal", "current", "wt.exe")
	require.NoError(t, os.MkdirAll(filepath.Dir(wt), 0o750))
	require.NoError(t, os.WriteFile(wt, []byte("wt"), 0o600))

	got, ok := installPortableWindowsTerminal()
	require.True(t, ok)
	assert.Equal(t, wt, got)
}

// TestBundledWindowsTerminalMatchesAsset is the pin between the archive in
// the tree, the bytes embedded in the executable and the hash recorded in
// assets/README.md. A silent replacement of any of the three fails here.
func TestBundledWindowsTerminalMatchesAsset(t *testing.T) {
	t.Parallel()

	require.Len(t, windowsTerminalBundle, bundledWindowsTerminalSize)
	sum := sha256.Sum256(windowsTerminalBundle)
	assert.Equal(t, bundledWindowsTerminalSHA256, hex.EncodeToString(sum[:]))

	asset, err := os.ReadFile(filepath.Join("assets", "WindowsTerminal_x64.zip"))
	require.NoError(t, err, "the audited archive stays in the tree")
	assetSum := sha256.Sum256(asset)
	assert.Equal(t, bundledWindowsTerminalSHA256, hex.EncodeToString(assetSum[:]),
		"the embedded copy and assets/WindowsTerminal_x64.zip diverged")

	readme, err := os.ReadFile(filepath.Join("assets", "README.md"))
	require.NoError(t, err)
	assert.Contains(t, string(readme), bundledWindowsTerminalSHA256,
		"assets/README.md must record the bundled hash")
	assert.Contains(t, string(readme), "v"+bundledWindowsTerminalVersion,
		"assets/README.md must record the bundled version")
}

func TestBundledWindowsTerminalArchiveStructure(t *testing.T) {
	t.Parallel()

	reader, err := zip.NewReader(bytes.NewReader(windowsTerminalBundle), int64(len(windowsTerminalBundle)))
	require.NoError(t, err)

	root := "terminal-" + bundledWindowsTerminalVersion + "/"
	dest := t.TempDir()
	var (
		hasWT       bool
		hasTerminal bool
		total       uint64
	)
	for _, file := range reader.File {
		assert.True(t, strings.HasPrefix(file.Name, root),
			"unexpected top-level entry %q", file.Name)
		_, err := safeArchivePath(dest, file.Name)
		require.NoError(t, err)
		assert.Zero(t, file.Mode()&os.ModeSymlink, "archive contains the symlink %q", file.Name)
		total += file.UncompressedSize64
		switch file.Name {
		case root + "wt.exe":
			hasWT = true
		case root + "WindowsTerminal.exe":
			hasTerminal = true
		}
	}
	assert.True(t, hasWT, "the bundle must contain wt.exe")
	assert.True(t, hasTerminal, "the bundle must contain WindowsTerminal.exe")
	assert.LessOrEqual(t, total, uint64(windowsTerminalExtractLimit))
}

func TestInstallBundledWindowsTerminal(t *testing.T) {
	local := t.TempDir()
	t.Setenv("LOCALAPPDATA", local)

	wt, err := installBundledWindowsTerminal()
	require.NoError(t, err)

	current := filepath.Join(local, "GoAnime", "WindowsTerminal", "current")
	assert.Equal(t, filepath.Join(current, "wt.exe"), wt)
	for _, name := range []string{"wt.exe", "WindowsTerminal.exe", "NOTICE.html"} {
		assert.FileExists(t, filepath.Join(current, name))
	}

	cached, ok := portableWindowsTerminalPath()
	require.True(t, ok)
	assert.Equal(t, wt, cached, "the next launch finds the installed copy")
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) {
	return 0, errors.New("read failed")
}

func TestDownloadWindowsTerminalAsset(t *testing.T) {
	const (
		assetURL = "https://github.com/microsoft/terminal/releases/download/v1.0/Microsoft.WindowsTerminal_1.0_x64.zip"
		body     = "windows-terminal-archive"
	)

	sum := sha256.Sum256([]byte(body))
	digest := "sha256:" + hex.EncodeToString(sum[:])

	tests := []struct {
		name          string
		asset         windowsTerminalAsset
		status        int
		contentLength int64
		body          io.Reader
		wantErr       string
	}{
		{
			name:          "verified download",
			asset:         windowsTerminalAsset{URL: assetURL, Digest: digest, Size: int64(len(body))},
			status:        http.StatusOK,
			contentLength: int64(len(body)),
			body:          strings.NewReader(body),
		},
		{
			name:          "digest mismatch",
			asset:         windowsTerminalAsset{URL: assetURL, Digest: "sha256:" + strings.Repeat("0", 64), Size: int64(len(body))},
			status:        http.StatusOK,
			contentLength: int64(len(body)),
			body:          strings.NewReader(body),
			wantErr:       "SHA-256",
		},
		{
			name:          "size mismatch",
			asset:         windowsTerminalAsset{URL: assetURL, Size: int64(len(body)) + 1},
			status:        http.StatusOK,
			contentLength: int64(len(body)),
			body:          strings.NewReader(body),
			wantErr:       "expected",
		},
		{
			name:          "GitHub error",
			asset:         windowsTerminalAsset{URL: assetURL},
			status:        http.StatusInternalServerError,
			contentLength: 0,
			body:          strings.NewReader(""),
			wantErr:       "answered 500",
		},
		{
			name:          "content length beyond the cap",
			asset:         windowsTerminalAsset{URL: assetURL},
			status:        http.StatusOK,
			contentLength: windowsTerminalZipLimit + 1,
			body:          strings.NewReader(""),
			wantErr:       "larger than",
		},
		{
			name:          "body read failure",
			asset:         windowsTerminalAsset{URL: assetURL},
			status:        http.StatusOK,
			contentLength: 100,
			body:          io.MultiReader(strings.NewReader("partial"), failingReader{}),
			wantErr:       "download Windows Terminal",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := windowsTerminalDownloadClient
			windowsTerminalDownloadClient = &http.Client{Transport: roundTripFunc(
				func(*http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode:    tt.status,
						Status:        fmt.Sprintf("%d %s", tt.status, http.StatusText(tt.status)),
						ContentLength: tt.contentLength,
						Body:          io.NopCloser(tt.body),
					}, nil
				})}
			t.Cleanup(func() { windowsTerminalDownloadClient = original })

			target := filepath.Join(t.TempDir(), "download.zip")
			require.NoError(t, os.WriteFile(target, nil, 0o600))

			err := downloadWindowsTerminalAsset(context.Background(), tt.asset, target)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			got, readErr := os.ReadFile(target)
			require.NoError(t, readErr)
			assert.Equal(t, body, string(got))
		})
	}
}

func TestDownloadWindowsTerminalAssetRejectsUnexpectedSource(t *testing.T) {
	t.Parallel()

	target := filepath.Join(t.TempDir(), "download.zip")
	for _, rawURL := range []string{
		"http://github.com/x.zip",
		"https://example.com/x.zip",
		"https://github.com.evil.example/x.zip",
		"not a url",
	} {
		err := downloadWindowsTerminalAsset(context.Background(), windowsTerminalAsset{URL: rawURL}, target)
		require.Error(t, err, "URL %q must be rejected", rawURL)
	}

	err := downloadWindowsTerminalAsset(context.Background(), windowsTerminalAsset{
		URL:  "https://github.com/x.zip",
		Size: windowsTerminalZipLimit + 1,
	}, target)
	require.Error(t, err, "a declared size beyond the cap must be rejected before downloading")
}
