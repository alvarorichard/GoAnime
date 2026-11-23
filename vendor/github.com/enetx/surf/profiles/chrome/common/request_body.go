package common

import "github.com/enetx/http/http2"

// RequestBody follows SpdyHttpStream::kRequestBodyBufferSize and
// SpdySession::CreateDataBuffer: 16 KiB including the nine-byte DATA header.
// Each UploadDataStream::Read gets the full buffer, including after a short read.
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/spdy/spdy_http_stream.cc
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/spdy/spdy_session.cc
func RequestBody(http2.RequestBodyInfo) http2.RequestBodyOptions {
	return http2.RequestBodyOptions{MaxReadSize: 16375, MaxDataFrameSize: 16375}
}
