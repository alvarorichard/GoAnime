// Package streaming classifies stream URLs and selects the appropriate download strategy.
package streaming

import (
	"net/url"
	"strings"
)

// IsBloggerProxyURL reports whether the URL points to the local Blogger proxy.
func IsBloggerProxyURL(u string) bool {
	lower := strings.ToLower(u)
	return strings.Contains(lower, "127.0.0.1") && strings.Contains(lower, "blogger_proxy")
}

// LooksLikeHLS reports whether the URL appears to point to an HLS playlist.
func LooksLikeHLS(u string) bool {
	lower := strings.ToLower(u)
	return strings.Contains(lower, ".m3u8") ||
		strings.Contains(lower, "m3u8") ||
		strings.Contains(lower, "/hls/")
}

// IsDASH reports whether the URL appears to point to a DASH manifest.
func IsDASH(u string) bool {
	return strings.Contains(strings.ToLower(u), ".mpd")
}

// HasUnsafeExtension reports whether the URL has an extension yt-dlp treats as unsafe.
func HasUnsafeExtension(u string) bool {
	path := u
	if before, _, ok := strings.Cut(u, "?"); ok {
		path = before
	}

	lower := strings.ToLower(path)
	for _, ext := range []string{".aspx", ".asp", ".php", ".jsp", ".cgi"} {
		if strings.HasSuffix(lower, ext) {
			return true
		}
	}
	return false
}

// IsWixmpURL reports whether the URL uses a Wix/WixMP CDN.
func IsWixmpURL(u string) bool {
	lower := strings.ToLower(u)
	return strings.Contains(lower, "wixmp.com") || strings.Contains(lower, "repackager.wixmp.com")
}

// IsSharePointURL reports whether the URL uses a SharePoint CDN.
func IsSharePointURL(u string) bool {
	return strings.Contains(strings.ToLower(u), "sharepoint.com")
}

// IsBloggerStreamURL reports whether the URL points to a Blogger-hosted stream page.
func IsBloggerStreamURL(u string) bool {
	return strings.Contains(strings.ToLower(u), "blogger.com")
}

// IsMirrorStreamURL reports whether the URL points to a mirrored stream endpoint.
func IsMirrorStreamURL(u string) bool {
	lower := strings.ToLower(u)
	return strings.Contains(lower, "allanime.pro") ||
		strings.Contains(lower, "allanime") ||
		strings.Contains(lower, "allmanga")
}

// NeedsContentLengthEstimate reports whether the URL is known to omit Content-Length frequently.
func NeedsContentLengthEstimate(u string) bool {
	lower := strings.ToLower(u)
	return IsSharePointURL(lower) ||
		IsWixmpURL(lower) ||
		IsDASH(lower) ||
		LooksLikeHLS(lower) ||
		IsMirrorStreamURL(lower) ||
		IsBloggerStreamURL(lower) ||
		strings.Contains(lower, "animefire") ||
		strings.Contains(lower, "animesfire")
}

// ShouldUseNativeHLSDownload reports whether native HLS is the safest first attempt.
func ShouldUseNativeHLSDownload(u string) bool {
	return LooksLikeHLS(u) || HasUnsafeExtension(u)
}

// ShouldUseYtDLPDownload reports whether yt-dlp should handle the stream download path.
func ShouldUseYtDLPDownload(u string) bool {
	lower := strings.ToLower(u)
	return LooksLikeHLS(lower) ||
		IsWixmpURL(lower) ||
		IsMirrorStreamURL(lower) ||
		IsBloggerStreamURL(lower) ||
		IsSharePointURL(lower) ||
		IsDASH(lower)
}

// EstimatedProgressSizeBytes returns a conservative progress estimate for streams that
// do not expose a stable Content-Length.
func EstimatedProgressSizeBytes(u string) int64 {
	if LooksLikeHLS(u) || IsWixmpURL(u) || IsDASH(u) {
		return 150 * 1024 * 1024
	}
	return 100 * 1024 * 1024
}

// DeriveReferer returns the origin form of a URL, suitable as a Referer fallback.
func DeriveReferer(u string) string {
	parsed, err := url.Parse(u)
	if err != nil {
		return ""
	}
	if parsed.Scheme == "" || parsed.Host == "" {
		return ""
	}
	return parsed.Scheme + "://" + parsed.Host + "/"
}
