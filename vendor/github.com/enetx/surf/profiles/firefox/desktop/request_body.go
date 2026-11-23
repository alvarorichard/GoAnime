package desktop

import (
	"strings"

	"github.com/enetx/http/http2"
)

// RequestBody follows the buffered transaction, ReadSegments budget and DATA
// chunk size. Connection establishment consumes the preparatory transaction
// before AddStream primes each real request with a 2048-byte read.
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/protocol/http/Http2Session.cpp
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/protocol/http/Http2StreamBase.cpp
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/protocol/http/nsHttpTransaction.cpp
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/modules/libpref/init/all.js
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/modules/libpref/init/StaticPrefList.yaml
func RequestBody(info http2.RequestBodyInfo) http2.RequestBodyOptions {
	const buffer, chunk = 32768, 16000
	cycle := buffer / (chunk + 8) * (chunk + 8)
	initial := 2048

	return http2.RequestBodyOptions{
		ResetCycleOnFlowBlock: true,
		MaxDataFrameSize:      chunk,
		ReadBufferSize:        buffer,
		ReadCycleSize:         cycle,
		InitialReadSize:       initial,
		HeaderBytes:           transactionHeaderBytes(info.Fields),
	}
}

// transactionHeaderBytes counts nsHttpTransaction's virtual HTTP/1 request
// prefix. TE is added to the request head after this prefix has been sealed.
// HTTP/2 splits cookies; the source transaction retains one joined Cookie line.
func transactionHeaderBytes(fields []http2.RequestHeaderField) int64 {
	var method, path, authority string
	var size, cookieBytes int
	cookies := 0
	for _, field := range fields {
		switch field.Name {
		case ":method":
			method = field.Value
		case ":path":
			path = field.Value
		case ":authority":
			authority = field.Value
		case "te":
		case "cookie":
			cookieBytes += len(field.Value)
			cookies++
		default:
			if !strings.HasPrefix(field.Name, ":") {
				size += len(field.Name) + 2 + len(field.Value) + 2
			}
		}
	}

	if cookies != 0 {
		size += len("Cookie: ") + cookieBytes + 2*(cookies-1) + 2
	}

	if path == "" {
		path = authority // Ordinary CONNECT has no :path.
	}

	size += len(method) + 1 + len(path) + len(" HTTP/1.1\r\n")
	size += len("Host: ") + len(authority) + 2
	size += len("Connection: keep-alive\r\n\r\n")
	return int64(size)
}
