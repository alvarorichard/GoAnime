//go:build !nethttpomithttp2

package http

func (t *Transport) hasHTTP2Connection(cm connectMethod) bool {
	t.idleMu.Lock()
	defer t.idleMu.Unlock()

	for _, conn := range t.idleConn[cm.key()] {
		pooled, ok := conn.alt.(http2RoundTripper)
		if ok && !conn.isBroken() &&
			pooled.t.HasConnection(cm.targetScheme, cm.targetAddr) {
			return true
		}
	}

	return false
}

func (t *Transport) http2Connections() []*HTTP2Connection {
	if t.h2Transport == nil {
		return nil
	}

	connections := t.h2Transport.Connections()
	result := make([]*HTTP2Connection, 0, len(connections))
	for _, conn := range connections {
		result = append(result, &HTTP2Connection{
			close: conn.Close,
			retire: func(options HTTP2RetirementOptions) {
				conn.Retire(options.ProbeContext, options.ProbePayload, options.GoAway)
			},
			ping: conn.Ping,
			idle: conn.NetworkIdleTime,
		})
	}

	return result
}
