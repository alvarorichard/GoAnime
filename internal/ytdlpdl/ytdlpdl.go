// Package ytdlpdl drives one yt-dlp download: how it is configured, retried and
// reported on. It knows nothing about where URLs come from or how progress is
// drawn; the caller supplies a Sink.
package ytdlpdl

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/lrstanley/go-ytdlp"
)

// Spec is everything one yt-dlp download is configured with.
type Spec struct {
	URL         string
	Output      string
	Impersonate bool     // present as Chrome, for CDNs that fingerprint TLS
	Headers     []string // "Name:value" request headers
}

// NewCommand configures yt-dlp for spec. It uses yt-dlp's native HLS
// downloader (not ffmpeg) so obfuscated segment extensions (.js, .html, .jpg)
// from CDNs are accepted: ffmpeg's allowed_segment_extensions check rejects them.
func NewCommand(spec Spec) *ytdlp.Command {
	dl := ytdlp.New().
		Output(spec.Output).
		Format("bestvideo+bestaudio/best").
		ConcurrentFragments(24).
		BufferSize("32M").
		FragmentRetries("5").
		Retries("5").
		SocketTimeout(30)
	if spec.Impersonate {
		dl.Impersonate("chrome")
	}
	for _, h := range spec.Headers {
		dl.AddHeaders(h)
	}
	return dl
}

// Engine is what a download needs from yt-dlp: to have it installed, and
// to run one configured download, reporting progress.
type Engine interface {
	Install(ctx context.Context) error
	Run(ctx context.Context, spec Spec, onProgress func(ytdlp.ProgressUpdate)) error
}

// GoEngine is the real engine, over the go-ytdlp library.
type GoEngine struct{}

func (GoEngine) Install(ctx context.Context) error {
	_, err := ytdlp.Install(ctx, nil)
	return err
}

func (GoEngine) Run(ctx context.Context, spec Spec, onProgress func(ytdlp.ProgressUpdate)) error {
	dl := NewCommand(spec)
	if onProgress != nil {
		dl.ProgressFunc(200*time.Millisecond, onProgress)
	}
	// --hls-use-mpegts has no typed method, so it goes as a raw argument.
	_, err := dl.Run(ctx, spec.URL, "--hls-use-mpegts")
	return err
}

// Run runs the download with retries, reporting progress to sink (nil for
// none). Each retry starts the byte counter over and calls onRetry(n) first,
// when it is set.
func Run(ctx context.Context, engine Engine, spec Spec, sink Sink, sleep func(time.Duration), onRetry func(attempt int)) error {
	progress := NewProgress(sink)
	var onProgress func(ytdlp.ProgressUpdate)
	if sink != nil {
		onProgress = progress.Handle
	}
	return RunWithRetry(
		func() error { return engine.Run(ctx, spec, onProgress) },
		MaxRetries,
		sleep,
		func(attempt int) {
			if onRetry != nil {
				onRetry(attempt)
			}
			progress.ResetForRetry()
		},
	)
}

// Outcome turns a finished run into the download's result. yt-dlp rejects
// URLs with unusual extensions (.aspx, ...) as a security measure
// (CVE-2024-38519); those fall back to a direct HTTP download.
func Outcome(runErr error, directHTTP func() error) error {
	if runErr == nil {
		return nil
	}
	if IsUnsafeExtensionError(runErr) {
		return directHTTP()
	}
	return fmt.Errorf("go-ytdlp download failed: %w", runErr)
}

// IsUnsafeExtensionError returns true if the error is yt-dlp's unsafe extension check.
func IsUnsafeExtensionError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "unsafe") && strings.Contains(s, "extension") ||
		strings.Contains(s, "unusual") && strings.Contains(s, "extension") ||
		strings.Contains(s, "is unusual and will be skipped")
}
func IsRetryable(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "network") ||
		strings.Contains(errStr, "temporary") ||
		strings.Contains(errStr, "reset") ||
		strings.Contains(errStr, "refused")
}

// MaxRetries is how many times a retryable yt-dlp failure is retried.
const MaxRetries = 2

// Headers are the request headers that make a CDN accept the download:
// the stored referer and an Origin derived from it (scheme and host only).
// No referer, no headers.
func Headers(referer string) []string {
	if referer == "" {
		return nil
	}
	origin := strings.TrimSuffix(referer, "/")
	if u, err := url.Parse(origin); err == nil && u.Scheme != "" && u.Host != "" {
		origin = u.Scheme + "://" + u.Host
	}
	return []string{"Referer:" + referer, "Origin:" + origin}
}

// RunWithRetry runs fn, retrying up to maxRetries times while the failure is
// retryable (IsRetryable). Before retry number n it waits 2n seconds
// through sleep and calls onRetry(n). It returns the last error, nil when a
// run succeeded.
func RunWithRetry(fn func() error, maxRetries int, sleep func(time.Duration), onRetry func(attempt int)) error {
	var err error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			sleep(time.Duration(attempt*2) * time.Second)
			onRetry(attempt)
		}
		if err = fn(); err == nil {
			return nil
		}
		if attempt == maxRetries || !IsRetryable(err) {
			return err
		}
	}
	return err
}

// Sink is where a download reports how far it got.
type Sink interface {
	AddProgressReceived(delta int64)
	SetProgressTotal(total int64)
	SetProgressPeak(pct float64)
}

// Progress turns yt-dlp's progress updates into sink calls. yt-dlp fetches
// video and audio as separate files and merges them, so received bytes are
// counted per file and the total is the sum of every file's size.
type Progress struct {
	sink      Sink
	lastBytes int64
	lastFile  string
	totals    map[string]int64
}

func NewProgress(sink Sink) *Progress {
	return &Progress{sink: sink, totals: make(map[string]int64)}
}

// ResetForRetry forgets the byte counter of the failed attempt: the next one
// starts again from zero.
func (p *Progress) ResetForRetry() {
	p.lastBytes = 0
	p.lastFile = ""
}

func (p *Progress) Handle(update ytdlp.ProgressUpdate) {
	if update.Status == ytdlp.ProgressStatusPostProcessing || update.Status == ytdlp.ProgressStatusFinished {
		return
	}
	if update.Filename != "" && update.Filename != p.lastFile {
		p.lastFile = update.Filename
		p.lastBytes = 0
	}

	downloaded := int64(update.DownloadedBytes)
	if delta := downloaded - p.lastBytes; delta > 0 {
		p.sink.AddProgressReceived(delta)
		p.lastBytes = downloaded
	}

	switch {
	case update.TotalBytes > 0:
		name := update.Filename
		if name == "" {
			name = "_default"
		}
		p.totals[name] = int64(update.TotalBytes)
		var sum int64
		for _, v := range p.totals {
			sum += v
		}
		p.sink.SetProgressTotal(sum)
	case update.FragmentCount > 0 && update.FragmentIndex > 0:
		// HLS: the size is unknown upfront, so fragment index/count is a
		// monotonically increasing ratio and the bar never goes backward.
		pct := float64(update.FragmentIndex) / float64(update.FragmentCount)
		if pct > 0.99 {
			pct = 0.99 // reserve 100% for actual completion
		}
		p.sink.SetProgressPeak(pct)
	}
}
