package common

import "github.com/enetx/http/http2"

// ConnectionClose omits a new GOAWAY on ordinary close. The engine still flushes
// any existing protocol-error GOAWAY before closing the socket.
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/spdy/spdy_session.cc
func ConnectionClose(info http2.ConnectionCloseInfo) http2.ConnectionCloseOptions {
	options := http2.ConnectionCloseOptions{}
	if info.ProtocolError && info.Detail == "DATA frame with stream ID 0" {
		// SpdySession::OnError formats Http2DecoderAdapter::INVALID_STREAM_ID.
		options.DebugData = []byte("Framer error: 1 (INVALID_STREAM_ID).")
	}

	return options
}
