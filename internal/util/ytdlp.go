package util

import (
	"os/exec"
	"strings"
	"sync"
)

var (
	ytdlpImpersonateOnce   sync.Once
	ytdlpImpersonateResult bool
)

// YtdlpCanImpersonate returns true if yt-dlp supports the "chrome"
// impersonate target (requires curl_cffi). The result is cached.
func YtdlpCanImpersonate() bool {
	ytdlpImpersonateOnce.Do(func() {
		out, err := exec.Command("yt-dlp", "--list-impersonate-targets").CombinedOutput()
		if err != nil {
			Debugf("yt-dlp impersonation check failed: %v", err)
			return
		}
		for _, line := range strings.Split(string(out), "\n") {
			lower := strings.ToLower(line)
			if strings.Contains(lower, "chrome") && !strings.Contains(lower, "unavailable") {
				ytdlpImpersonateResult = true
				Debugf("yt-dlp chrome impersonation is available")
				return
			}
		}
		Debugf("yt-dlp chrome impersonation is NOT available (curl_cffi missing)")
	})
	return ytdlpImpersonateResult
}
