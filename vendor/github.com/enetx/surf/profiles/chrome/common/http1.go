package common

import "github.com/enetx/surf/profiles"

// ConfigureH1 follows Chromium's 1400-byte in-memory header/body merge and
// 16 KiB upload reads. Streaming bodies keep independent header transmission.
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/http/http_stream_parser.cc
func ConfigureH1(config *profiles.HTTP1Config) {
	config.WriteBufferSize(16384).BodyReadSize(16384).HeaderBodyMergeLimit(1400)
}
