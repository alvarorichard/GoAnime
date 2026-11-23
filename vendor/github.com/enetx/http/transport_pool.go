package http

// HasHTTP2Connection reports whether the normal pool has a reusable HTTP/2
// session for req's route. A session at its stream limit can still report true:
// the normal request path decides whether to queue or create another connection.
// This snapshot neither reserves a stream nor dials, sends, or changes req's body.
// It may call Proxy to select the same route as RoundTrip. Custom HTTP/2 protocol
// implementations and custom connection pools are not inspected.
func (t *Transport) HasHTTP2Connection(req *Request) bool {
	if req == nil || req.URL == nil || req.URL.Host == "" ||
		(req.URL.Scheme != "https" && req.URL.Scheme != "http") ||
		req.Context().Err() != nil || req.Close ||
		t.DisableKeepAlives || t.MaxIdleConnsPerHost < 0 {
		return false
	}

	t.nextProtoOnce.Do(t.onceSetNextProtoDefaults)
	cm, err := t.connectMethodForRequest(&transportRequest{Request: req})
	if err != nil || cm.onlyH1 {
		return false
	}

	return t.hasHTTP2Connection(cm)
}
