package cfsolver

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/updater"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// GoAnime does not make anyone build cfsolve. When no binary is installed it
// fetches the one GoAnime's CI built for the pinned commit (cfsolve.lock) from
// GoAnime's own releases, checks it against the SHA256SUMS published beside it
// (whose minisign signature must verify against the key embedded in GoAnime,
// the same trust the updater applies) and installs it where Discover looks. Every failure is silent to the user:
// Goyabu simply stays out of the search, as it does without the tool.

//go:embed cfsolve.lock
var lockFile string

const (
	envAutoDownload = "GOANIME_CFSOLVE_AUTODOWNLOAD" // "0" turns the download off
	releasesBase    = "https://github.com/alvarorichard/GoAnime/releases/download"
	sumsName        = "SHA256SUMS"
	sigSuffix       = ".minisig"
	retryAfter      = 24 * time.Hour // between failed attempts
	maxBinaryBytes  = 256 << 20
	maxSumsBytes    = 1 << 20
	failureMarker   = ".download-failed"
)

// pinnedSHA is the cfsolve commit named by cfsolve.lock.
func pinnedSHA() string {
	for _, line := range strings.Split(lockFile, "\n") {
		if v, ok := strings.CutPrefix(strings.TrimSpace(line), "sha="); ok {
			return v
		}
	}
	return ""
}

// releaseTag names the GoAnime release that carries the binaries for sha.
func releaseTag(sha string) string {
	if len(sha) > 12 {
		sha = sha[:12]
	}
	return "cfsolve-" + sha
}

// assetName is the release asset for a platform, or false when none is built.
func assetName(goos, goarch string) (string, bool) {
	switch goos + "/" + goarch {
	case "linux/amd64", "linux/arm64", "darwin/amd64", "darwin/arm64":
		return "challenge-" + goos + "-" + goarch, true
	case "windows/amd64":
		return "challenge-windows-amd64.exe", true
	}
	return "", false
}

// Installer downloads the pinned cfsolve binary into Dir.
type Installer struct {
	BaseURL string // releases download root, without the tag
	Tag     string
	Asset   string
	Dir     string
	Client  *http.Client
	Now     func() time.Time
	// Verify checks a minisign signature over the checksum list.
	Verify func(message, sig []byte) error
}

// DefaultInstaller targets this platform's asset for the pinned commit, or
// reports false when the platform has none or the config folder is unknown.
func DefaultInstaller() (*Installer, bool) {
	asset, ok := assetName(runtime.GOOS, runtime.GOARCH)
	sha := pinnedSHA()
	if !ok || sha == "" {
		return nil, false
	}
	cfg, err := os.UserConfigDir()
	if err != nil {
		return nil, false
	}
	return &Installer{
		BaseURL: releasesBase,
		Tag:     releaseTag(sha),
		Asset:   asset,
		Dir:     filepath.Join(cfg, "goanime", "cfsolve"),
		Client:  &http.Client{Timeout: 10 * time.Minute},
		Now:     time.Now,
		Verify:  updater.VerifyTrusted,
	}, true
}

// Install fetches, verifies and installs the binary, returning its path. A
// failed attempt is remembered for retryAfter so a broken release or a missing
// network does not cost a download attempt on every launch.
func (in *Installer) Install(ctx context.Context) (string, error) {
	if err := os.MkdirAll(in.Dir, 0o750); err != nil {
		return "", err
	}
	marker := filepath.Join(in.Dir, failureMarker)
	if st, err := os.Stat(marker); err == nil && in.Now().Sub(st.ModTime()) < retryAfter {
		return "", errors.New("cfsolve: a recent download failed; not retrying yet")
	}
	path, err := in.fetch(ctx)
	if err != nil {
		_ = os.WriteFile(marker, []byte(err.Error()), 0o600)
		_ = os.Chtimes(marker, in.Now(), in.Now())
		return "", err
	}
	_ = os.Remove(marker)
	return path, nil
}

func (in *Installer) get(ctx context.Context, name string) (*http.Response, error) {
	u := in.BaseURL + "/" + in.Tag + "/" + name
	if !strings.HasPrefix(u, "https://") && !strings.HasPrefix(u, "http://127.0.0.1") {
		return nil, fmt.Errorf("cfsolve: refusing non-https download %q", u)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := in.Client.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()
		return nil, fmt.Errorf("cfsolve: %s: %s", name, resp.Status)
	}
	return resp, nil
}

// readAll returns a small release file, capped at limit bytes.
func (in *Installer) readAll(ctx context.Context, name string, limit int64) ([]byte, error) {
	resp, err := in.get(ctx, name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(b)) > limit {
		return nil, fmt.Errorf("cfsolve: %s is too large", name)
	}
	return b, nil
}

// expectedSum returns the checksum the signed SHA256SUMS lists for the asset. A
// list whose signature does not verify is rejected before it is even parsed, so
// a swapped release cannot vouch for its own binary.
func (in *Installer) expectedSum(ctx context.Context) (string, error) {
	sums, err := in.readAll(ctx, sumsName, maxSumsBytes)
	if err != nil {
		return "", err
	}
	sig, err := in.readAll(ctx, sumsName+sigSuffix, maxSumsBytes)
	if err != nil {
		return "", err
	}
	if in.Verify == nil {
		return "", errors.New("cfsolve: no signature verifier configured")
	}
	if err := in.Verify(sums, sig); err != nil {
		return "", fmt.Errorf("cfsolve: %s signature: %w", sumsName, err)
	}
	sc := bufio.NewScanner(bytes.NewReader(sums))
	for sc.Scan() {
		f := strings.Fields(sc.Text())
		if len(f) == 2 && strings.TrimPrefix(f[1], "*") == in.Asset {
			if _, err := hex.DecodeString(f[0]); err != nil || len(f[0]) != sha256.Size*2 {
				return "", fmt.Errorf("cfsolve: malformed checksum for %s", in.Asset)
			}
			return strings.ToLower(f[0]), nil
		}
	}
	return "", fmt.Errorf("cfsolve: %s lists no %s", sumsName, in.Asset)
}

func (in *Installer) fetch(ctx context.Context) (string, error) {
	want, err := in.expectedSum(ctx)
	if err != nil {
		return "", err
	}
	resp, err := in.get(ctx, in.Asset)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	tmp, err := os.CreateTemp(in.Dir, "challenge-*.part")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.Remove(tmp.Name()) }() // gone after the rename; cleans up every failure

	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(tmp, h), io.LimitReader(resp.Body, maxBinaryBytes+1))
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return "", err
	}
	if n > maxBinaryBytes {
		return "", errors.New("cfsolve: download exceeds the size limit")
	}
	if got := hex.EncodeToString(h.Sum(nil)); got != want {
		return "", fmt.Errorf("cfsolve: checksum mismatch for %s", in.Asset)
	}
	if err := os.Chmod(tmp.Name(), 0o700); err != nil { // #nosec G302 -- an executable needs its x bit; 0700 is the tightest mode that can run
		return "", err
	}
	final := filepath.Join(in.Dir, "challenge")
	if runtime.GOOS == "windows" {
		final += ".exe"
	}
	if err := os.Rename(tmp.Name(), final); err != nil {
		return "", err
	}
	return final, nil
}

// AutoDownloadEnabled reports whether GoAnime may fetch cfsolve by itself.
func AutoDownloadEnabled() bool { return os.Getenv(envAutoDownload) != "0" }

// InstallInBackground installs the pinned cfsolve when none is present and
// registers it, without ever blocking or reporting to the user. It returns
// immediately; Goyabu joins the search from the moment the install finishes.
func InstallInBackground() {
	if !AutoDownloadEnabled() {
		return
	}
	in, ok := DefaultInstaller()
	if !ok {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), in.Client.Timeout)
		defer cancel()
		path, err := in.Install(ctx)
		if err != nil {
			util.Debug("cfsolve: not installed", "err", err)
			return
		}
		util.Debug("cfsolve: installed", "path", path)
		Register()
	}()
}
