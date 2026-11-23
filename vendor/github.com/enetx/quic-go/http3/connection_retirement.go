package http3

import "context"

func retiredClient(cl *roundTripperWithCount) bool {
	select {
	case <-cl.dialing:
	default:
		return false
	}
	if cl.conn != nil && cl.conn.Context().Err() != nil {
		return true
	}
	client := nativeClient(cl.clientConn)
	return client != nil && client.goAwayCtx.Err() != nil
}

// A drained owner can still carry active response bodies. Remove it from new
// lookup, but retain lifecycle ownership until the connection ends or Close is
// called. Idle cleanup separately checks the actual active stream count.
func (t *Transport) retireClientLocked(cl *roundTripperWithCount) {
	if t.clients[cl.poolKey] == cl {
		delete(t.clients, cl.poolKey)
	}
	if t.closed || cl.conn == nil || cl.conn.Context().Err() != nil {
		return
	}
	if t.retired == nil {
		t.retired = make(map[*roundTripperWithCount]struct{})
	}
	if _, exists := t.retired[cl]; exists {
		return
	}
	t.retired[cl] = struct{}{}
	context.AfterFunc(cl.conn.Context(), func() {
		t.mutex.Lock()
		delete(t.retired, cl)
		t.mutex.Unlock()
	})
}
