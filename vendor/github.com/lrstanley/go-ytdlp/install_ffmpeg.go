// Copyright (c) Liam Stanley <liam@liam.sh>. All rights reserved. Use of
// this source code is governed by the MIT license that can be found in
// the LICENSE file.

package ytdlp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"sync"
	"sync/atomic"
)

var (
	ffmpegResolveCache  = atomic.Pointer[ResolvedInstall]{} // Should only be used by [InstallFFmpeg].
	ffmpegInstallLock   sync.Mutex
	ffprobeResolveCache = atomic.Pointer[ResolvedInstall]{} // Should only be used by [InstallFFprobe].
	ffprobeInstallLock  sync.Mutex

	ffmpegBinConfigs = map[string]ffmpegBinConfig{
		"darwin_amd64": {
			ffmpegURL:     "https://evermeet.cx/ffmpeg/getrelease/ffmpeg",
			ffprobeURL:    "https://evermeet.cx/ffmpeg/getrelease/ffprobe",
			ffmpegBinary:  "ffmpeg",
			ffprobeBinary: "ffprobe",
			grouped:       false,
		},
		"darwin_arm64": {
			ffmpegURL:     "https://www.osxexperts.net/ffmpeg80arm.zip",
			ffprobeURL:    "https://www.osxexperts.net/ffprobe80arm.zip",
			ffmpegBinary:  "ffmpeg",
			ffprobeBinary: "ffprobe",
			grouped:       false,
		},
		"linux_amd64": {
			ffmpegURL:     "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linux64-gpl.tar.xz",
			ffmpegBinary:  "ffmpeg",
			ffprobeBinary: "ffprobe",
			grouped:       true,
		},
		"linux_arm64": {
			ffmpegURL:     "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-linuxarm64-gpl.tar.xz",
			ffmpegBinary:  "ffmpeg",
			ffprobeBinary: "ffprobe",
			grouped:       true,
		},
		"windows_amd64": {
			ffmpegURL:     "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-win64-gpl.zip",
			ffmpegBinary:  "ffmpeg.exe",
			ffprobeBinary: "ffprobe.exe",
			grouped:       true,
		},
		"windows_arm": {
			ffmpegURL:     "https://github.com/BtbN/FFmpeg-Builds/releases/download/latest/ffmpeg-master-latest-winarm64-gpl.zip",
			ffmpegBinary:  "ffmpeg.exe",
			ffprobeBinary: "ffprobe.exe",
			grouped:       true,
		},
	}
)

type ffmpegBinConfig struct {
	ffmpegURL     string
	ffprobeURL    string
	ffmpegBinary  string
	ffprobeBinary string
	grouped       bool // true if the download is an archive containing both binaries
}

type InstallFFmpegOptions struct {
	// DisableDownload is a simple toggle to never allow downloading, which would
	// be the same as never calling [InstallFFmpeg] or [MustInstallFFmpeg] in the first place.
	DisableDownload bool

	// DisableSystem is a simple toggle to never allow resolving from the system PATH.
	DisableSystem bool

	// DownloadURL is the exact url to the binary location to download (and store).
	// Leave empty to use GitHub (windows, linux) and evermeet.cx (macos) +
	// auto-detected os/arch.
	DownloadURL string
}

// MustInstallFFmpeg is similar to [InstallFFmpeg], but panics if there is an error.
func MustInstallFFmpeg(ctx context.Context, opts *InstallFFmpegOptions) {
	_, err := InstallFFmpeg(ctx, opts)
	if err != nil {
		panic(err)
	}
}

// InstallFFmpeg will attempt to download and install FFmpeg for the current platform.
// If the binary is already installed or found in the PATH, it will return the resolved
// binary unless [InstallFFmpegOptions.DisableSystem] is set to true. Note that
// downloading of ffmpeg and ffprobe is only supported on a handful of platforms, and so
// it is still recommended to install ffmpeg/ffprobe via other means.
func InstallFFmpeg(ctx context.Context, opts *InstallFFmpegOptions) (*ResolvedInstall, error) {
	if opts == nil {
		opts = &InstallFFmpegOptions{}
	}

	if cached := ffmpegResolveCache.Load(); cached != nil {
		return cached, nil
	}

	config, err := getBinaryConfig(ffmpegBinConfigs)
	if err != nil {
		return nil, err
	}

	ffmpegInstallLock.Lock()
	defer ffmpegInstallLock.Unlock()

	resolved, err := resolveExecutable(ctx, false, opts.DisableSystem, []string{config.ffmpegBinary})
	if err == nil {
		if resolved.Version == "" {
			err = ffGetVersion(ctx, resolved)
			if err != nil {
				return nil, err
			}
		}
		ffmpegResolveCache.Store(resolved)
		return resolved, nil
	}

	if opts.DisableDownload {
		return nil, errors.New("ffmpeg binary not found, and downloading is disabled")
	}

	// Download and install FFmpeg (and FFprobe if archive).
	resolved, err = downloadAndInstallFFmpeg(ctx, opts)
	if err != nil {
		return nil, err
	}

	ffmpegResolveCache.Store(resolved)
	return resolved, nil
}

// MustInstallFFprobe is similar to [InstallFFprobe], but panics if there is an error.
func MustInstallFFprobe(ctx context.Context, opts *InstallFFmpegOptions) {
	_, err := InstallFFprobe(ctx, opts)
	if err != nil {
		panic(err)
	}
}

// InstallFFprobe will attempt to download and install FFprobe for the current platform.
// If the binary is already installed or found in the PATH, it will return the resolved
// binary unless [InstallFFmpegOptions.DisableSystem] is set to true. Note that
// downloading of ffmpeg and ffprobe is only supported on a handful of platforms, and so
// it is still recommended to install ffmpeg/ffprobe via other means.
func InstallFFprobe(ctx context.Context, opts *InstallFFmpegOptions) (*ResolvedInstall, error) {
	if opts == nil {
		opts = &InstallFFmpegOptions{}
	}

	if cached := ffprobeResolveCache.Load(); cached != nil {
		return cached, nil
	}

	config, err := getBinaryConfig(ffmpegBinConfigs)
	if err != nil {
		return nil, err
	}

	if config.grouped {
		// If the binaries are grouped, we need to use the ffmpeg install lock,
		// since the 1 artifact (for either) will contain both ffmpeg and ffprobe.
		ffmpegInstallLock.Lock()
		defer ffmpegInstallLock.Unlock()
	} else {
		ffprobeInstallLock.Lock()
		defer ffprobeInstallLock.Unlock()
	}

	resolved, err := resolveExecutable(ctx, false, opts.DisableSystem, []string{config.ffprobeBinary})
	if err == nil {
		if resolved.Version == "" {
			err = ffGetVersion(ctx, resolved)
			if err != nil {
				return nil, err
			}
		}
		ffprobeResolveCache.Store(resolved)
		return resolved, nil
	}

	if opts.DisableDownload {
		return nil, errors.New("ffprobe binary not found, and downloading is disabled")
	}

	// Download and install FFprobe (and FFmpeg if archive).
	resolved, err = downloadAndInstallFFprobe(ctx, opts)
	if err != nil {
		return nil, err
	}

	ffprobeResolveCache.Store(resolved)
	return resolved, nil
}

func downloadAndInstallFFmpeg(ctx context.Context, opts *InstallFFmpegOptions) (*ResolvedInstall, error) {
	config, err := getBinaryConfig(ffmpegBinConfigs)
	if err != nil {
		return nil, err
	}

	cacheDir, err := createCacheDir(ctx)
	if err != nil {
		return nil, err
	}

	downloadURL := opts.DownloadURL
	if downloadURL == "" {
		downloadURL = config.ffmpegURL
	}

	destPath := filepath.Join(cacheDir, config.ffmpegBinary)

	if config.grouped {
		// Download and extract archive.
		err = downloadAndExtractFilesFromArchive(ctx, downloadURL, cacheDir, []string{config.ffmpegBinary, config.ffprobeBinary})
		if err != nil {
			return nil, fmt.Errorf("failed to download and extract ffmpeg archive: %w", err)
		}
	} else if isArchiveURL(downloadURL) {
		// Download and extract single binary from archive.
		err = downloadAndExtractFilesFromArchive(ctx, downloadURL, cacheDir, []string{config.ffmpegBinary})
		if err != nil {
			return nil, fmt.Errorf("failed to download and extract ffmpeg: %w", err)
		}
	} else {
		// Download single binary.
		destPath, err = downloadFile(ctx, downloadURL, cacheDir, destPath, 0o700)
		if err != nil {
			return nil, fmt.Errorf("failed to download ffmpeg: %w", err)
		}
	}

	return &ResolvedInstall{
		Executable: destPath,
		FromCache:  false,
		Downloaded: true,
	}, nil
}

func downloadAndInstallFFprobe(ctx context.Context, opts *InstallFFmpegOptions) (*ResolvedInstall, error) {
	config, err := getBinaryConfig(ffmpegBinConfigs)
	if err != nil {
		return nil, err
	}

	cacheDir, err := createCacheDir(ctx)
	if err != nil {
		return nil, err
	}

	destPath := filepath.Join(cacheDir, config.ffprobeBinary)

	downloadURL := opts.DownloadURL
	if downloadURL == "" {
		if config.grouped {
			downloadURL = config.ffmpegURL // Use ffmpeg URL for grouped archive.
		} else {
			downloadURL = config.ffprobeURL
		}
	}

	switch {
	case config.grouped:
		// Download and extract archive (contains both ffmpeg and ffprobe).
		err = downloadAndExtractFilesFromArchive(ctx, downloadURL, cacheDir, []string{config.ffmpegBinary, config.ffprobeBinary})
		if err != nil {
			return nil, fmt.Errorf("failed to download and extract ffprobe archive: %w", err)
		}
	case isArchiveURL(downloadURL):
		// Download and extract single binary from archive.
		err = downloadAndExtractFilesFromArchive(ctx, downloadURL, cacheDir, []string{config.ffprobeBinary})
		if err != nil {
			return nil, fmt.Errorf("failed to download and extract ffprobe: %w", err)
		}
	default:
		// Download single binary.
		destPath, err = downloadFile(ctx, downloadURL, cacheDir, "", 0o700)
		if err != nil {
			return nil, fmt.Errorf("failed to download ffprobe: %w", err)
		}
	}

	return &ResolvedInstall{
		Executable: destPath,
		FromCache:  false,
		Downloaded: true,
	}, nil
}

var ffmpegVersionRegex = regexp.MustCompile(`^(?:ffmpeg|ffprobe) version ([^ ]+) .*`)

func ffGetVersion(ctx context.Context, r *ResolvedInstall) error {
	var stdout bytes.Buffer

	cmd := exec.CommandContext(ctx, r.Executable, "-version") //nolint:gosec
	cmd.Stdout = &stdout
	applySyscall(cmd, false)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("unable to run %s to verify version: %w", r.Executable, err)
	}

	version := ffmpegVersionRegex.FindStringSubmatch(stdout.String())
	if len(version) < 2 {
		return fmt.Errorf("unable to parse %s version from output", r.Executable)
	}

	r.Version = version[1]
	debug(ctx, "resolved version", "binary", r.Executable, "version", r.Version)
	return nil
}
