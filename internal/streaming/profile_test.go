package streaming

import "testing"

func TestShouldUseNativeHLSDownload(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"hls_playlist", "https://cdn.example.com/master.m3u8", true},
		{"unsafe_extension", "https://cdn.example.com/video.aspx?id=1", true},
		{"plain_mp4", "https://cdn.example.com/video.mp4", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldUseNativeHLSDownload(tt.url); got != tt.want {
				t.Fatalf("ShouldUseNativeHLSDownload(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestShouldUseYtDLPDownload(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"hls", "https://cdn.example.com/master.m3u8", true},
		{"wix", "https://repackager.wixmp.com/video.mp4", true},
		{"blogger", "https://www.blogger.com/video.g?token=abc123", true},
		{"sharepoint", "https://tenant.sharepoint.com/video.aspx", true},
		{"dash", "https://cdn.example.com/manifest.mpd", true},
		{"plain_mp4", "https://cdn.example.com/video.mp4", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ShouldUseYtDLPDownload(tt.url); got != tt.want {
				t.Fatalf("ShouldUseYtDLPDownload(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}

func TestEstimatedProgressSizeBytes(t *testing.T) {
	if got := EstimatedProgressSizeBytes("https://cdn.example.com/master.m3u8"); got != 150*1024*1024 {
		t.Fatalf("EstimatedProgressSizeBytes(HLS) = %d, want %d", got, int64(150*1024*1024))
	}
	if got := EstimatedProgressSizeBytes("https://cdn.example.com/video.mp4"); got != 100*1024*1024 {
		t.Fatalf("EstimatedProgressSizeBytes(MP4) = %d, want %d", got, int64(100*1024*1024))
	}
}

func TestDeriveReferer(t *testing.T) {
	got := DeriveReferer("https://rapid-cloud.co/embed-6/abc123?k=1")
	if got != "https://rapid-cloud.co/" {
		t.Fatalf("DeriveReferer() = %q, want %q", got, "https://rapid-cloud.co/")
	}

	if got := DeriveReferer("not-a-url"); got != "" {
		t.Fatalf("DeriveReferer(invalid) = %q, want empty", got)
	}
}
