package http

// HasConnectionGroup reports existing native TCP pool ownership for req's route:
// pooled connections, occupied HTTP/1 connections and pending connection dials.
// A group is not a promise of an immediately reusable socket. The request path
// still owns connection admission, health checks and retries.
//
// This snapshot does not dial, reserve a connection, send a request or touch its
// body. Proxy may be called to select the same pool key as RoundTrip. Custom
// protocol pools are not inspected; use HasHTTP2Connection for reusable H2 state.
func (t *Transport) HasConnectionGroup(req *Request) bool {
	if req == nil || req.URL == nil || req.URL.Host == "" ||
		(req.URL.Scheme != "https" && req.URL.Scheme != "http") ||
		req.Context().Err() != nil || req.Close ||
		t.DisableKeepAlives || t.MaxIdleConnsPerHost < 0 {
		return false
	}

	t.nextProtoOnce.Do(t.onceSetNextProtoDefaults)
	cm, err := t.connectMethodForRequest(&transportRequest{Request: req})
	if err != nil {
		return false
	}

	key := cm.key()
	t.connsPerHostMu.Lock()
	dialing := false
	t.dialsInProgress.all(func(w *wantConn) {
		if w.key == key && w.cancelCtx != nil {
			dialing = true
		}
	})
	t.connsPerHostMu.Unlock()

	if dialing {
		return true
	}

	t.idleMu.Lock()
	defer t.idleMu.Unlock()

	if len(t.idleConn[key]) != 0 {
		return true
	}

	for conn := range t.trackedHTTP1 {
		if conn.cacheKey == key {
			return true
		}
	}

	return false
}
