package updater

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sha256Hex returns the digest checksums-sha256.txt would list for data.
func sha256Hex(data string) string {
	sum := sha256.Sum256([]byte(data))
	return hex.EncodeToString(sum[:])
}

// newRelease builds a release with the given {name, url} assets.
func newRelease(tag string, assets ...[2]string) *GitHubRelease {
	rel := &GitHubRelease{TagName: tag}
	for _, a := range assets {
		rel.Assets = append(rel.Assets, struct {
			Name               string `json:"name"`
			BrowserDownloadURL string `json:"browser_download_url"`
		}{Name: a[0], BrowserDownloadURL: a[1]})
	}
	return rel
}

// withFetchChecksums swaps fetchChecksumsFn and restores it on cleanup.
// MUST NOT be used with t.Parallel() since it mutates a package var.
func withFetchChecksums(t *testing.T, fn func(string) ([]byte, error)) {
	t.Helper()
	prev := fetchChecksumsFn
	fetchChecksumsFn = fn
	t.Cleanup(func() { fetchChecksumsFn = prev })
}

// writeTemp writes content to a file in a fresh temp dir and returns its path.
func writeTemp(t *testing.T, name, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))
	return path
}

func TestFindChecksumsAsset(t *testing.T) {
	t.Parallel()
	const url = "https://github.com/alvarorichard/GoAnime/releases/download/v1.2.3/checksums-sha256.txt"

	tests := []struct {
		name    string
		release *GitHubRelease
		wantURL string
		wantErr string
	}{
		{
			name:    "listed among other assets",
			release: newRelease("v1.2.3", [2]string{"goanime-linux-amd64", "https://x/bin"}, [2]string{"checksums-sha256.txt", url}),
			wantURL: url,
		},
		{
			name:    "name differs only in case",
			release: newRelease("v1.2.3", [2]string{"Checksums-SHA256.txt", url}),
			wantURL: url,
		},
		{
			name:    "not published",
			release: newRelease("v1.2.3", [2]string{"goanime-linux-amd64", "https://x/bin"}),
			wantErr: "release v1.2.3 does not publish checksums-sha256.txt",
		},
		{
			name:    "no assets at all",
			release: newRelease("v1.2.3"),
			wantErr: "does not publish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := findChecksumsAsset(tt.release)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantURL, got)
		})
	}
}

func TestFetchChecksums_RejectsURLsOutsideGitHub(t *testing.T) {
	t.Parallel()
	// fetchChecksums runs without test mode, so none of these may reach the
	// network: each is refused by URL validation first.
	for _, url := range []string{
		"http://127.0.0.1:1/checksums-sha256.txt",
		"https://example.com/checksums-sha256.txt",
		"http://github.com/alvarorichard/GoAnime/releases/download/v1/checksums-sha256.txt",
	} {
		t.Run(url, func(t *testing.T) {
			t.Parallel()
			_, err := fetchChecksums(url)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "URL validation failed")
		})
	}
}

func TestFetchChecksumsWithTestFlag(t *testing.T) {
	t.Parallel()
	body := sha256Hex("a") + "  goanime-linux-amd64\n"

	tests := []struct {
		name    string
		handler http.HandlerFunc
		want    string
		wantErr string
	}{
		{
			name:    "returns the file",
			handler: func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte(body)) },
			want:    body,
		},
		{
			name: "accepts a file at the size limit",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(strings.Repeat("x", maxChecksumsBytes)))
			},
			want: strings.Repeat("x", maxChecksumsBytes),
		},
		{
			name: "rejects a file over the size limit",
			handler: func(w http.ResponseWriter, _ *http.Request) {
				_, _ = w.Write([]byte(strings.Repeat("x", maxChecksumsBytes+1)))
			},
			wantErr: "larger than",
		},
		{
			name:    "reports a non-200 status",
			handler: func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusNotFound) },
			wantErr: "status 404",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(tt.handler)
			t.Cleanup(srv.Close)

			got, err := fetchChecksumsWithTestFlag(srv.URL+"/checksums-sha256.txt", true)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, string(got))
		})
	}
}

func TestFetchChecksumsWithTestFlag_ServerUnreachable(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.NotFoundHandler())
	url := srv.URL + "/checksums-sha256.txt"
	srv.Close() // nothing listens on the address any more

	_, err := fetchChecksumsWithTestFlag(url, true)
	require.Error(t, err)
}

func TestParseChecksums(t *testing.T) {
	t.Parallel()
	a, b := sha256Hex("a"), sha256Hex("b")

	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr string
	}{
		{
			name:  "sha256sum text mode",
			input: a + "  goanime-linux-amd64\n" + b + "  goanime-windows-amd64.zip\n",
			want:  map[string]string{"goanime-linux-amd64": a, "goanime-windows-amd64.zip": b},
		},
		{
			name:  "sha256sum binary mode",
			input: a + " *goanime-darwin-universal\n",
			want:  map[string]string{"goanime-darwin-universal": a},
		},
		{
			name:  "CRLF line endings and blank lines",
			input: "\r\n" + a + "  goanime-linux-amd64\r\n\r\n" + b + "  goanime-linux-arm64\r\n",
			want:  map[string]string{"goanime-linux-amd64": a, "goanime-linux-arm64": b},
		},
		{
			name:  "uppercase digest is normalized",
			input: strings.ToUpper(a) + "  goanime-linux-amd64\n",
			want:  map[string]string{"goanime-linux-amd64": a},
		},
		{
			name:  "identical duplicate entries",
			input: a + "  goanime-linux-amd64\n" + a + "  goanime-linux-amd64\n",
			want:  map[string]string{"goanime-linux-amd64": a},
		},
		{
			name:    "conflicting duplicate entries",
			input:   a + "  goanime-linux-amd64\n" + b + "  goanime-linux-amd64\n",
			wantErr: "conflicting checksums for goanime-linux-amd64",
		},
		{
			name:    "digest too short",
			input:   a[:63] + "  goanime-linux-amd64\n",
			wantErr: "malformed line 1",
		},
		{
			name:    "digest not hex",
			input:   strings.Repeat("z", 64) + "  goanime-linux-amd64\n",
			wantErr: "malformed line 1",
		},
		{
			name:    "line without a file name",
			input:   a + "  goanime-linux-amd64\n" + b + "\n",
			wantErr: "malformed line 2",
		},
		{
			name:    "empty file",
			input:   "",
			wantErr: "lists no checksums",
		},
		{
			name:    "only blank lines",
			input:   "\n\r\n  \n",
			wantErr: "lists no checksums",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := parseChecksums([]byte(tt.input))
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFileSHA256(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		path    func(t *testing.T) string
		want    string
		wantErr bool
	}{
		{
			name: "known vector",
			path: func(t *testing.T) string { return writeTemp(t, "abc", "abc") },
			want: "ba7816bf8f01cfea414140de5dae2223b00361a396177a9cb410ff61f20015ad",
		},
		{
			name: "empty file",
			path: func(t *testing.T) string { return writeTemp(t, "empty", "") },
			want: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
		},
		{
			name:    "missing file",
			path:    func(t *testing.T) string { return filepath.Join(t.TempDir(), "missing") },
			wantErr: true,
		},
		{
			name:    "path traversal",
			path:    func(*testing.T) string { return "../outside" },
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := fileSHA256(tt.path(t))
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestVerifyAssetChecksum(t *testing.T) {
	t.Parallel()
	const asset = "goanime-linux-amd64"
	path := writeTemp(t, asset, "new-bin")

	tests := []struct {
		name    string
		path    string
		sums    map[string]string
		wantErr string
	}{
		{
			name: "digest matches",
			path: path,
			sums: map[string]string{asset: sha256Hex("new-bin")},
		},
		{
			name:    "digest differs",
			path:    path,
			sums:    map[string]string{asset: sha256Hex("tampered")},
			wantErr: "checksum mismatch for goanime-linux-amd64",
		},
		{
			name:    "asset not listed",
			path:    path,
			sums:    map[string]string{"goanime-linux-arm64": sha256Hex("new-bin")},
			wantErr: "has no entry for goanime-linux-amd64",
		},
		{
			name:    "downloaded file missing",
			path:    filepath.Join(t.TempDir(), "gone"),
			sums:    map[string]string{asset: sha256Hex("new-bin")},
			wantErr: "failed to hash goanime-linux-amd64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := verifyAssetChecksum(tt.path, asset, tt.sums)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestVerifyDownload(t *testing.T) {
	const (
		asset        = "goanime-linux-amd64"
		checksumsURL = "https://github.com/alvarorichard/GoAnime/releases/download/v9.9.9/checksums-sha256.txt"
	)
	path := writeTemp(t, asset, "new-bin")
	published := newRelease("v9.9.9", [2]string{asset, "https://x/bin"}, [2]string{checksumsAssetName, checksumsURL})

	tests := []struct {
		name      string
		release   *GitHubRelease
		fetch     func(string) ([]byte, error)
		wantFetch bool
		wantErr   string
	}{
		{
			name:      "download matches the published checksum",
			release:   published,
			fetch:     func(string) ([]byte, error) { return []byte(sha256Hex("new-bin") + "  " + asset + "\n"), nil },
			wantFetch: true,
		},
		{
			name:    "release without a checksums file",
			release: newRelease("v9.9.9", [2]string{asset, "https://x/bin"}),
			fetch:   func(string) ([]byte, error) { return nil, errors.New("must not be called") },
			wantErr: "does not publish checksums-sha256.txt",
		},
		{
			name:      "checksums file cannot be downloaded",
			release:   published,
			fetch:     func(string) ([]byte, error) { return nil, errors.New("connection reset") },
			wantFetch: true,
			wantErr:   "failed to download checksums-sha256.txt: connection reset",
		},
		{
			name:      "checksums file is malformed",
			release:   published,
			fetch:     func(string) ([]byte, error) { return []byte("<html>rate limited</html>"), nil },
			wantFetch: true,
			wantErr:   "malformed line 1",
		},
		{
			name:      "download does not match",
			release:   published,
			fetch:     func(string) ([]byte, error) { return []byte(sha256Hex("tampered") + "  " + asset + "\n"), nil },
			wantFetch: true,
			wantErr:   "checksum mismatch",
		},
		{
			name:      "asset missing from the checksums file",
			release:   published,
			fetch:     func(string) ([]byte, error) { return []byte(sha256Hex("new-bin") + "  goanime-linux-arm64\n"), nil },
			wantFetch: true,
			wantErr:   "has no entry for goanime-linux-amd64",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var fetchedURL string
			withFetchChecksums(t, func(url string) ([]byte, error) {
				fetchedURL = url
				return tt.fetch(url)
			})

			err := verifyDownload(tt.release, path, asset)

			if tt.wantFetch {
				assert.Equal(t, checksumsURL, fetchedURL, "must fetch the release's own checksums file")
			} else {
				assert.Empty(t, fetchedURL, "must not fetch anything")
			}
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

func TestPerformUpdate_VerificationFailureInstallsNothing(t *testing.T) {
	dir := t.TempDir()
	currentExe := filepath.Join(dir, "current")
	require.NoError(t, os.WriteFile(currentExe, []byte("ORIGINAL"), 0o755))
	downloaded := filepath.Join(dir, "downloaded")
	require.NoError(t, os.WriteFile(downloaded, []byte("TAMPERED"), 0o644))

	replaced := false
	withPerformHooks(t,
		func(*GitHubRelease) (string, string, error) { return "u", "goanime-linux-amd64", nil },
		func(string, string) (string, error) { return downloaded, nil },
		func() (string, error) { return currentExe, nil },
		func(string, string) error { replaced = true; return nil },
	)

	release := newRelease("v9.9.9")
	var gotRelease *GitHubRelease
	var gotPath, gotAsset string
	verifyDownloadFn = func(r *GitHubRelease, path, asset string) error {
		gotRelease, gotPath, gotAsset = r, path, asset
		return errors.New("checksum mismatch for goanime-linux-amd64")
	}

	err := PerformUpdate(release)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "nothing was installed")
	assert.Contains(t, err.Error(), "checksum mismatch")
	assert.Same(t, release, gotRelease)
	assert.Equal(t, downloaded, gotPath, "verifies the downloaded file")
	assert.Equal(t, "goanime-linux-amd64", gotAsset)
	assert.False(t, replaced, "the executable must not be replaced")
	got, readErr := os.ReadFile(currentExe)
	require.NoError(t, readErr)
	assert.Equal(t, "ORIGINAL", string(got))
	assert.NoFileExists(t, currentExe+".backup", "no backup is made before verification passes")
}

func TestPerformUpdate_VerifiesAgainstPublishedChecksums(t *testing.T) {
	const asset = "goanime-linux-amd64"

	tests := []struct {
		name      string
		published string // content whose digest the checksums file lists
		wantErr   bool
		wantExe   string
	}{
		{name: "matching download is installed", published: "NEW", wantExe: "NEW"},
		{name: "mismatching download is refused", published: "SOMETHING ELSE", wantErr: true, wantExe: "ORIGINAL"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			currentExe := filepath.Join(dir, "current")
			require.NoError(t, os.WriteFile(currentExe, []byte("ORIGINAL"), 0o755))
			downloaded := filepath.Join(dir, "downloaded")
			require.NoError(t, os.WriteFile(downloaded, []byte("NEW"), 0o644))

			withPerformHooks(t,
				func(*GitHubRelease) (string, string, error) { return "u", asset, nil },
				func(string, string) (string, error) { return downloaded, nil },
				func() (string, error) { return currentExe, nil },
				func(_, src string) error {
					data, err := os.ReadFile(src)
					if err != nil {
						return err
					}
					return os.WriteFile(currentExe, data, 0o755)
				},
			)
			verifyDownloadFn = verifyDownload // the real check, fed by the mock below
			withFetchChecksums(t, func(string) ([]byte, error) {
				return []byte(sha256Hex(tt.published) + "  " + asset + "\n"), nil
			})

			err := PerformUpdate(newRelease("v9.9.9", [2]string{checksumsAssetName, "https://x/checksums"}))

			if tt.wantErr {
				require.Error(t, err)
				assert.Contains(t, err.Error(), "checksum mismatch")
			} else {
				require.NoError(t, err)
			}
			got, readErr := os.ReadFile(currentExe)
			require.NoError(t, readErr)
			assert.Equal(t, tt.wantExe, string(got))
		})
	}
}
