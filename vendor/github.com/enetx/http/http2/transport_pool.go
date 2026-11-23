package http2

import "github.com/enetx/http"

// HasHTTP2Connection reports whether the shared native pool has a reusable
// HTTP/2 session for req. It does not dial, send, reserve a stream, or touch the
// body. A busy session can report true. Custom ConnPool implementations report
// false because their ownership and admission rules are not observable here.
func (t *Transport) HasHTTP2Connection(req *http.Request) bool {
	if t.ConnPool != nil || req == nil || req.URL == nil ||
		(req.URL.Scheme == "http" && !t.AllowHTTP) {
		return false
	}

	return t.init().HasHTTP2Connection(req)
}

// HTTP2Connections snapshots the shared native pool without dialing, reserving
// streams or sending frames. Custom ConnPool implementations are not inspected.
func (t *Transport) HTTP2Connections() []*http.HTTP2Connection {
	if t.ConnPool != nil {
		return nil
	}

	return t.init().HTTP2Connections()
}
