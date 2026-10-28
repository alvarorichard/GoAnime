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

// localArchiveName cleans an archive entry's name and refuses anything that is
// not strictly inside the directory it is extracted to: "..", absolute paths,
// drive-relative names and, on Windows, reserved device names. A release zip is
// remote input.
func localArchiveName(name string) (string, error) {
	if name == "" {
		return "", errors.New("archive contains an empty path")
	}
	native := filepath.FromSlash(name)
	if !filepath.IsLocal(native) {
		return "", fmt.Errorf("archive contains an unsafe path %q", name)
	}
	clean := filepath.Clean(native)
	if clean == "." {
		return "", fmt.Errorf("archive contains an unsafe path %q", name)
	}
	return clean, nil
}

// safeArchivePath joins an archive entry to dest, or refuses it when it would
// escape dest. Extraction itself goes through an os.Root, which confines every
// file operation to dest even if this check were wrong.
func safeArchivePath(dest, name string) (string, error) {
	clean, err := localArchiveName(name)
	if err != nil {
		return "", err
	}
	return filepath.Join(dest, clean), nil
}

// extractWindowsTerminalZip unpacks the official archive into dest and returns
// the directory holding wt.exe, which is the archive's top-level folder. The
// caller owns dest and removes it after activating the copy.
func extractWindowsTerminalZip(zipPath, dest string) (string, error) {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return "", fmt.Errorf("open the Windows Terminal archive: %w", err)
	}
	defer func() { _ = reader.Close() }()
	return extractWindowsTerminalReader(&reader.Reader, dest)
}

func extractWindowsTerminalReader(reader *zip.Reader, dest string) (string, error) {
	root, err := os.OpenRoot(dest)
	if err != nil {
		return "", fmt.Errorf("open %s: %w", dest, err)
	}
	defer func() { _ = root.Close() }()

	var wtDir string
	var total int64
	for _, file := range reader.File {
		name, err := localArchiveName(file.Name)
		if err != nil {
			return "", err
		}
		if file.FileInfo().IsDir() {
			if err := root.MkdirAll(name, 0o750); err != nil {
				return "", fmt.Errorf("create %s: %w", name, err)
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
		if err := root.MkdirAll(filepath.Dir(name), 0o750); err != nil {
			return "", fmt.Errorf("create %s: %w", filepath.Dir(name), err)
		}
		written, err := extractFile(file, root, name, windowsTerminalExtractLimit-total)
		if err != nil {
			return "", err
		}
		total += written
		if wtDir == "" && strings.EqualFold(filepath.Base(name), "wt.exe") {
			wtDir = filepath.Join(dest, filepath.Dir(name))
		}
	}
	if wtDir == "" {
		return "", errors.New("archive does not contain wt.exe")
	}
	return wtDir, nil
}

// installWindowsTerminalArchive extracts an archive under root and makes it the
// active private copy, replacing any previous one. Extraction goes to a
// staging directory first, so a failed download or a full disk never leaves a
// half-written current copy behind.
func installWindowsTerminalArchive(reader *zip.Reader, root string) (string, error) {
	if err := os.MkdirAll(root, 0o750); err != nil {
		return "", fmt.Errorf("create %s: %w", root, err)
	}
	staging, err := os.MkdirTemp(root, "extract-*")
	if err != nil {
		return "", fmt.Errorf("create extraction directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(staging) }()

	wtDir, err := extractWindowsTerminalReader(reader, staging)
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
	return wt, nil
}

// extractFile copies one entry while enforcing the remaining extraction
// budget. The header's uncompressed size is not trusted: a malformed archive
// could understate it and expand on disk.
func extractFile(file *zip.File, root *os.Root, name string, limit int64) (int64, error) {
	source, err := file.Open()
	if err != nil {
		return 0, fmt.Errorf("read %s: %w", file.Name, err)
	}
	defer func() { _ = source.Close() }()

	out, err := root.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return 0, fmt.Errorf("create %s: %w", name, err)
	}
	written, copyErr := io.Copy(out, io.LimitReader(source, limit+1))
	closeErr := out.Close()
	if copyErr != nil {
		return written, fmt.Errorf("extract %s: %w", file.Name, copyErr)
	}
	if closeErr != nil {
		return written, fmt.Errorf("close %s: %w", name, closeErr)
	}
	if written > limit {
		return written, errors.New("archive expands beyond the size limit")
	}
	return written, nil
}
