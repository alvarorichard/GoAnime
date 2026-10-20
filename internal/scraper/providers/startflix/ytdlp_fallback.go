package startflix

import (
	"bytes"
	"context"
	"flag"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
)

// ytdlpFallbackAllowed reports whether the generic yt-dlp fallback may run.
// Under `go test` it is off by default: fixtures use real host names
// (vidsrcme.su) that would otherwise send yt-dlp to the network and stall
// the suite for the per-host timeout. GOANIME_YTDLP_FALLBACK_TEST=1 opts
// the fallback back in for the dedicated fallback test.
func ytdlpFallbackAllowed() bool {
	if os.Getenv("GOANIME_YTDLP_FALLBACK_TEST") == "1" {
		return true
	}
	return flag.Lookup("test.v") == nil && flag.Lookup("test.run") == nil
}

var (
	ytdlpPathOnce sync.Once
	ytdlpPath     string
)

// ytdlpBinary returns the yt-dlp binary path, or "" when it is not on PATH.
// Cached: LookPath hits the filesystem on every call.
func ytdlpBinary() string {
	ytdlpPathOnce.Do(func() {
		if p, err := exec.LookPath("yt-dlp"); err == nil {
			ytdlpPath = p
		} else {
			util.Debug("StartFlix yt-dlp fallback unavailable: yt-dlp not on PATH")
		}
	})
	return ytdlpPath
}

// ytdlpTimeout bounds one yt-dlp extraction. Dead hosts (DNS fail) answer in
// ~seconds; a live host that needs extracting takes a few more. Each host
// runs in parallel, so this is per-host, not total.
const ytdlpTimeout = 20 * time.Second

// resolveViaYtDlp asks yt-dlp for the direct media URL behind a player page
// GoAnime has no dedicated resolver for (dood-like file hosts, fembed
// mirrors, etc.). It returns the first http(s) URL yt-dlp prints, or "".
//
// Best effort by design: any failure (no binary, timeout, unsupported,
// dead host) is just a miss — the caller keeps the original NoStreamError
// so the user still learns which servers were listed.
func resolveViaYtDlp(ctx context.Context, playerURL, referer string) string {
	args := ytdlpArgs(playerURL, referer)
	if args == nil {
		return ""
	}
	bin := ytdlpBinary()
	if bin == "" {
		return ""
	}
	pctx, cancel := context.WithTimeout(ctx, ytdlpTimeout)
	defer cancel()
	// #nosec G204 -- bin comes only from LookPath("yt-dlp"); no shell is used,
	// and ytdlpArgs validates the URL and separates it from command options.
	cmd := exec.CommandContext(pctx, bin, args...)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		util.Debug("StartFlix yt-dlp fallback missed", "url", playerURL[:min(len(playerURL), 80)], "err", err)
		return ""
	}
	for line := range strings.Lines(strings.TrimSpace(out.String())) {
		if u := strings.TrimSpace(line); strings.HasPrefix(u, "https://") || strings.HasPrefix(u, "http://") {
			return u
		}
	}
	return ""
}

// ytdlpArgs accepts only remote player pages and keeps scraped values out
// of yt-dlp's option parser, including a referer beginning with a dash.
func ytdlpArgs(playerURL, referer string) []string {
	u, err := url.Parse(playerURL)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return nil
	}
	args := []string{
		"--no-warnings", "--no-playlist",
		"--get-url", "-f", "best[ext=mp4]/best",
	}
	if referer != "" {
		args = append(args, "--referer="+referer)
	}
	return append(args, "--", playerURL)
}

// ytdlpResolveFn extracts the direct URL for one player; a seam so the
// fallback test drives ResolveStream without a yt-dlp binary or network.
var ytdlpResolveFn = resolveViaYtDlp

// tryYtDlpFallback runs yt-dlp over the players no dedicated resolver
// handles and returns the first direct URL it extracts, or "". It runs the
// hosts in parallel: the first to answer wins, the rest are abandoned with
// the context.
func tryYtDlpFallback(ctx context.Context, players []Player) *Stream {
	if len(players) == 0 || !ytdlpFallbackAllowed() {
		return nil
	}
	type hit struct {
		stream *Stream
	}
	winner := make(chan hit, 1)
	fctx, cancel := context.WithCancel(ctx)
	defer cancel()
	var wg sync.WaitGroup
	for _, p := range players {
		u := p.URL
		ref := p.referer
		if ref == "" {
			ref = p.URL
		}
		host := u
		if i := strings.Index(u, "://"); i >= 0 {
			host = u[i+3:]
			if j := strings.IndexByte(host, '/'); j >= 0 {
				host = host[:j]
			}
		}
		wg.Go(func() {
			if direct := ytdlpResolveFn(fctx, u, ref); direct != "" {
				util.Debug("StartFlix yt-dlp fallback hit", "host", host)
				select {
				case winner <- hit{stream: &Stream{URL: direct, Referer: ref, Host: host}}:
					cancel()
				default:
				}
			}
		})
	}
	go func() { wg.Wait(); close(winner) }()
	for h := range winner {
		if h.stream != nil {
			return h.stream
		}
	}
	return nil
}
