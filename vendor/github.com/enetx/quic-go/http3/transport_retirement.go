package http3

import "context"

// RetireConnections removes current connections from request lookup. Existing
// streams finish normally and their connections then close. Idle connections
// and pending handshakes close immediately. It does not send GOAWAY, clear TLS
// tickets or prevent future use of the transport. Connections obtained through
// NewClientConn are application-owned and are not included.
func (t *Transport) RetireConnections() {
	t.mutex.Lock()
	if t.closed {
		t.mutex.Unlock()
		return
	}

	clients := t.clients
	t.clients = nil
	for _, client := range clients {
		if t.retired == nil {
			t.retired = make(map[*roundTripperWithCount]struct{})
		}
		t.retired[client] = struct{}{}
	}
	t.mutex.Unlock()

	for _, client := range clients {
		select {
		case <-client.dialing:
			if client.conn != nil && client.conn.Context().Err() == nil {
				t.retireReadyClient(client)
				continue
			}
		default:
		}

		client.cancel()
		go func() {
			client.Close()
			t.forgetRetiredClient(client)
		}()
	}
}

func (t *Transport) retireReadyClient(client *roundTripperWithCount) {
	if restarting, ok := client.clientConn.(*earlyRestartClient); ok {
		restarting.retire()
	} else if conn := nativeClient(client.clientConn); conn != nil {
		conn.retire()
	} else {
		client.Close()
	}

	context.AfterFunc(client.conn.Context(), func() { t.forgetRetiredClient(client) })
}

func (t *Transport) forgetRetiredClient(client *roundTripperWithCount) {
	t.mutex.Lock()
	delete(t.retired, client)
	t.mutex.Unlock()
}
