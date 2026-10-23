//go:build windows

package tui

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	windowsTerminalReleaseAPI   = "https://api.github.com/repos/microsoft/terminal/releases/latest"
	windowsTerminalReleasePage  = "https://github.com/microsoft/terminal/releases/latest"
	windowsTerminalDownloadBase = "https://github.com/microsoft/terminal/releases/download/"
	windowsTerminalInstallLimit = 6 * time.Minute
)

var (
	// Release lookup fails fast so an offline launch does not sit on a blank
	// console; the archive itself may take minutes on a slow connection.
	windowsTerminalLookupClient   = &http.Client{Timeout: 15 * time.Second}
	windowsTerminalDownloadClient = &http.Client{Timeout: 5 * time.Minute}
)

type windowsTerminalRelease struct {
	TagName string                 `json:"tag_name"`
	Assets  []windowsTerminalAsset `json:"assets"`
}

// installPortableWindowsTerminal downloads Microsoft's official unpackaged
// Windows Terminal build into the app's per-user data directory and returns
// its wt.exe path. It is the fallback for images that do not ship winget, such
// as Windows Sandbox. The copy is cached, so the download happens once.
func installPortableWindowsTerminal() (string, bool) {
	if wt, ok := portableWindowsTerminalPath(); ok {
		return wt, true
	}
	wt, err := downloadPortableWindowsTerminal()
	if err != nil {
		fmt.Fprintf(os.Stderr, "GoAnime: could not download a private Windows Terminal: %v\n", err)
		return "", false
	}
	return wt, true
}

func downloadPortableWindowsTerminal() (string, error) {
	root, err := portableWindowsTerminalRoot()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(root, 0o750); err != nil {
		return "", fmt.Errorf("create %s: %w", root, err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), windowsTerminalInstallLimit)
	defer cancel()

	release, err := windowsTerminalReleaseInfo(ctx)
	if err != nil {
		return "", err
	}
	asset, err := selectWindowsTerminalAsset(release.Assets, runtime.GOARCH)
	if err != nil {
		return "", err
	}
	if asset.Size > 0 {
		fmt.Fprintf(os.Stderr, "GoAnime: downloading Windows Terminal %s (%.1f MB)...\n", release.TagName, float64(asset.Size)/(1<<20))
	} else {
		fmt.Fprintf(os.Stderr, "GoAnime: downloading Windows Terminal %s...\n", release.TagName)
	}

	zipFile, err := os.CreateTemp(root, "download-*.zip")
	if err != nil {
		return "", fmt.Errorf("create download file: %w", err)
	}
	zipPath := zipFile.Name()
	defer func() { _ = os.Remove(zipPath) }()
	if err := zipFile.Close(); err != nil {
		return "", fmt.Errorf("prepare download file: %w", err)
	}

	if err := downloadWindowsTerminalAsset(ctx, asset, zipPath); err != nil {
		return "", err
	}

	staging, err := os.MkdirTemp(root, "extract-*")
	if err != nil {
		return "", fmt.Errorf("create extraction directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(staging) }()

	wtDir, err := extractWindowsTerminalZip(zipPath, staging)
	if err != nil {
		return "", err
	}

	current := filepath.Join(root, "current")
	if err := os.RemoveAll(current); err != nil {
		return "", fmt.Errorf("replace previous Windows Terminal copy: %w", err)
	}
	if err := os.Rename(wtDir, current); err != nil {
		return "", fmt.Errorf("activate Windows Terminal: %w", err)
	}
	wt := filepath.Join(current, "wt.exe")
	if info, err := os.Stat(wt); err != nil || info.IsDir() {
		return "", errors.New("private Windows Terminal copy is missing wt.exe")
	}
	fmt.Fprintln(os.Stderr, "GoAnime: Windows Terminal is ready.")
	return wt, nil
}

// portableWindowsTerminalRoot is the per-user directory for the fallback
// build. It lives under GoAnime's own data directory instead of the aliases
// Windows reserves for installed packages.
func portableWindowsTerminalRoot() (string, error) {
	local := os.Getenv("LOCALAPPDATA")
	if local == "" {
		return "", errors.New("LOCALAPPDATA is not set")
	}
	return filepath.Join(local, "GoAnime", "WindowsTerminal"), nil
}

func portableWindowsTerminalPath() (string, bool) {
	root, err := portableWindowsTerminalRoot()
	if err != nil {
		return "", false
	}
	wt := filepath.Join(root, "current", "wt.exe")
	// #nosec G703 -- LOCALAPPDATA is the user's profile root and the joined
	// path is the app's own fixed install directory.
	if info, err := os.Stat(wt); err == nil && !info.IsDir() {
		return wt, true
	}
	return "", false
}

// windowsTerminalReleaseInfo asks the release API for the newest stable build,
// including GitHub's SHA-256 digest. If the API is rate-limited or offline,
// it falls back to the release page redirect and builds the asset URL from the
// tag; the download is still served over TLS by github.com.
func windowsTerminalReleaseInfo(ctx context.Context) (windowsTerminalRelease, error) {
	release, apiErr := fetchWindowsTerminalRelease(ctx)
	if apiErr == nil {
		return release, nil
	}
	release, pageErr := fetchWindowsTerminalReleasePage(ctx)
	if pageErr == nil {
		return release, nil
	}
	return windowsTerminalRelease{}, fmt.Errorf("look up the Windows Terminal release: %w", errors.Join(apiErr, pageErr))
}

func fetchWindowsTerminalRelease(ctx context.Context) (windowsTerminalRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, windowsTerminalReleaseAPI, http.NoBody)
	if err != nil {
		return windowsTerminalRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "GoAnime")
	// #nosec G107 -- fixed GitHub API URL; the parsed assets are downloaded
	// from github.com and checked against GitHub's digest when available.
	resp, err := windowsTerminalLookupClient.Do(req)
	if err != nil {
		return windowsTerminalRelease{}, fmt.Errorf("contact GitHub: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return windowsTerminalRelease{}, fmt.Errorf("GitHub answered %s", resp.Status)
	}

	var release windowsTerminalRelease
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&release); err != nil {
		return windowsTerminalRelease{}, fmt.Errorf("read the GitHub release: %w", err)
	}
	if release.TagName == "" || len(release.Assets) == 0 {
		return windowsTerminalRelease{}, errors.New("GitHub returned an empty Windows Terminal release")
	}
	return release, nil
}

func fetchWindowsTerminalReleasePage(ctx context.Context) (windowsTerminalRelease, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, windowsTerminalReleasePage, http.NoBody)
	if err != nil {
		return windowsTerminalRelease{}, err
	}
	client := &http.Client{
		Timeout: 15 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	// #nosec G107 -- fixed github.com release page; only its redirect target
	// is read, no body is downloaded.
	resp, err := client.Do(req)
	if err != nil {
		return windowsTerminalRelease{}, fmt.Errorf("contact the release page: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	location := resp.Header.Get("Location")
	if location == "" {
		return windowsTerminalRelease{}, fmt.Errorf("release page answered %s without a redirect", resp.Status)
	}
	target, err := url.Parse(location)
	if err != nil {
		return windowsTerminalRelease{}, fmt.Errorf("parse the release redirect: %w", err)
	}
	tag := path.Base(strings.TrimSuffix(target.Path, "/"))
	if !strings.HasPrefix(tag, "v") {
		return windowsTerminalRelease{}, fmt.Errorf("unexpected release tag %q", tag)
	}

	arch, err := windowsTerminalArchitecture(runtime.GOARCH)
	if err != nil {
		return windowsTerminalRelease{}, err
	}
	version := strings.TrimPrefix(tag, "v")
	name := fmt.Sprintf("Microsoft.WindowsTerminal_%s_%s.zip", version, arch)
	return windowsTerminalRelease{
		TagName: tag,
		Assets: []windowsTerminalAsset{{
			Name: name,
			URL:  windowsTerminalDownloadBase + tag + "/" + name,
		}},
	}, nil
}

func downloadWindowsTerminalAsset(ctx context.Context, asset windowsTerminalAsset, target string) error {
	if asset.Size > windowsTerminalZipLimit {
		return fmt.Errorf("archive is larger than %d MB", windowsTerminalZipLimit>>20)
	}
	parsed, err := url.Parse(asset.URL)
	if err != nil || parsed.Scheme != "https" || parsed.Host != "github.com" {
		return fmt.Errorf("unexpected Windows Terminal download URL %q", asset.URL)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, asset.URL, http.NoBody)
	if err != nil {
		return err
	}
	// #nosec G107 -- the URL is restricted to https://github.com above, and
	// GitHub's digest is verified when the release API supplied one.
	resp, err := windowsTerminalDownloadClient.Do(req)
	if err != nil {
		return fmt.Errorf("download Windows Terminal: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download Windows Terminal: GitHub answered %s", resp.Status)
	}
	if resp.ContentLength > windowsTerminalZipLimit {
		return fmt.Errorf("archive is larger than %d MB", windowsTerminalZipLimit>>20)
	}

	out, err := os.OpenFile(target, os.O_WRONLY|os.O_TRUNC, 0o600) // #nosec G304 -- target is the temp file just created by the caller.
	if err != nil {
		return fmt.Errorf("open the download: %w", err)
	}
	hasher := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(out, hasher), io.LimitReader(resp.Body, windowsTerminalZipLimit+1))
	closeErr := out.Close()
	if copyErr != nil {
		return fmt.Errorf("download Windows Terminal: %w", copyErr)
	}
	if closeErr != nil {
		return fmt.Errorf("finish the Windows Terminal download: %w", closeErr)
	}
	if written > windowsTerminalZipLimit {
		return fmt.Errorf("archive is larger than %d MB", windowsTerminalZipLimit>>20)
	}
	if asset.Size > 0 && written != asset.Size {
		return fmt.Errorf("download is %d bytes, expected %d", written, asset.Size)
	}
	if digest := strings.TrimPrefix(asset.Digest, "sha256:"); digest != "" {
		sum := hex.EncodeToString(hasher.Sum(nil))
		if !strings.EqualFold(sum, digest) {
			return errors.New("download failed its SHA-256 check")
		}
	}
	return nil
}
