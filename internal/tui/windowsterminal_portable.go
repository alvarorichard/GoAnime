package tui

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const (
	// The official unpackaged archive is around 12 MB; the caps leave room
	// for future releases while refusing a runaway download or a zip bomb.
	windowsTerminalZipLimit     = 64 << 20
	windowsTerminalExtractLimit = 512 << 20
)

// windowsTerminalAsset is one file attached to a Windows Terminal GitHub
// release. Digest is GitHub's "sha256:<hex>" asset digest when the release
// API answered, and empty when only the release page could be reached.
type windowsTerminalAsset struct {
	Name   string `json:"name"`
	URL    string `json:"browser_download_url"`
	Digest string `json:"digest"`
	Size   int64  `json:"size"`
}

// windowsTerminalArchitecture maps Go's architecture to the suffix Microsoft
// uses in the unpackaged archive names.
func windowsTerminalArchitecture(goarch string) (string, error) {
	switch goarch {
	case "amd64":
		return "x64", nil
	case "arm64":
		return "arm64", nil
	case "386":
		return "x86", nil
	default:
		return "", fmt.Errorf("no unpackaged Windows Terminal build for %s", goarch)
	}
}

// selectWindowsTerminalAsset picks the unpackaged stable archive for arch. The
// release also publishes a Windows 10 preinstall kit and raw MSIX bundles;
// the name filter keeps those out.
func selectWindowsTerminalAsset(assets []windowsTerminalAsset, goarch string) (windowsTerminalAsset, error) {
	arch, err := windowsTerminalArchitecture(goarch)
	if err != nil {
		return windowsTerminalAsset{}, err
	}
	suffix := "_" + arch + ".zip"
	for _, asset := range assets {
		if strings.HasPrefix(asset.Name, "Microsoft.WindowsTerminal_") &&
			strings.HasSuffix(asset.Name, suffix) &&
			asset.URL != "" {
			return asset, nil
		}
	}
	return windowsTerminalAsset{}, fmt.Errorf("no Windows Terminal %s archive in the release", arch)
}

// safeArchivePath joins an archive entry to dest and refuses anything that
// would escape it. A release zip is remote input, so "..", absolute paths and
// drive-relative names never become files outside the install directory.
func safeArchivePath(dest, name string) (string, error) {
	if name == "" {
		return "", errors.New("archive contains an empty path")
	}
	clean := filepath.Clean(filepath.FromSlash(name))
	if clean == "." || clean == ".." || filepath.IsAbs(clean) {
		return "", fmt.Errorf("archive contains an unsafe path %q", name)
	}
	target := filepath.Join(dest, clean)
	rel, err := filepath.Rel(dest, target)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("archive contains an unsafe path %q", name)
	}
	return target, nil
}

// extractWindowsTerminalZip unpacks the official archive into dest and returns
// the directory holding wt.exe, which is the archive's top-level folder. The
// caller owns dest and removes it after activating the copy.
func extractWindowsTerminalZip(zipPath, dest string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("open Windows Terminal archive: %w", err)
	}
	defer func() { _ = reader.Close() }()

	var wtDir string
	var total int64
	for _, file := range reader.File {
		target, err := safeArchivePath(dest, file.Name)
		if err != nil {
			return "", err
		}
		if file.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o750); err != nil {
				return "", fmt.Errorf("create %s: %w", target, err)
			}
			continue
		}
		if file.Mode()&os.ModeSymlink != 0 {
			return "", fmt.Errorf("archive contains a symlink %q", file.Name)
		}
		if file.UncompressedSize64 > windowsTerminalExtractLimit ||
			total+int64(file.UncompressedSize64) > windowsTerminalExtractLimit {
			return "", errors.New("archive expands beyond the size limit")
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
			return "", fmt.Errorf("create %s: %w", filepath.Dir(target), err)
		}
		written, err := extractFile(file, target, windowsTerminalExtractLimit-total)
		if err != nil {
			return "", err
		}
		total += written
		if wtDir == "" && strings.EqualFold(filepath.Base(target), "wt.exe") {
			wtDir = filepath.Dir(target)
		}
	}
	if wtDir == "" {
		return "", errors.New("archive does not contain wt.exe")
	}
	return wtDir, nil
}

// extractFile copies one entry while enforcing the remaining extraction
// budget. The header's uncompressed size is not trusted: a malformed archive
// could understate it and expand on disk.
func extractFile(file *zip.File, target string, limit int64) (int64, error) {
	source, err := file.Open()
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", file.Name, err)
	}
	defer func() { _ = source.Close() }()

	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) // #nosec G304 -- target is validated by safeArchivePath.
	if err != nil {
		return 0, fmt.Errorf("create %s: %w", target, err)
	}
	written, copyErr := io.Copy(out, io.LimitReader(source, limit+1))
	closeErr := out.Close()
	if copyErr != nil {
		return written, fmt.Errorf("extract %s: %w", file.Name, copyErr)
	}
	if closeErr != nil {
		return written, fmt.Errorf("close %s: %w", target, closeErr)
	}
	if written > limit {
		return written, errors.New("archive expands beyond the size limit")
	}
	return written, nil
}
