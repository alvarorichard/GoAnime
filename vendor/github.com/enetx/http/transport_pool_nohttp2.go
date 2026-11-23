//go:build nethttpomithttp2

package http

func (*Transport) hasHTTP2Connection(connectMethod) bool { return false }

func (*Transport) http2Connections() []*HTTP2Connection { return nil }
