package http

import (
	"context"
	"errors"
	"time"
)

var errNoConnectionTraffic = errors.New("http: no connection traffic during liveness check")

// RetireHTTP1Connections excludes current pooled HTTP/1 connections from future
// requests. Idle connections close immediately; active exchanges can finish.
// It neither affects HTTP/2 nor application-owned ClientConns.
func (t *Transport) RetireHTTP1Connections() {
	t.idleMu.Lock()
	var idle []*persistConn
	for conn := range t.trackedHTTP1 {
		conn.retired.Store(true)
		if t.idleLRU.m[conn] != nil {
			idle = append(idle, conn)
		}
	}
	t.idleMu.Unlock()

	for _, conn := range idle {
		conn.close(errCloseIdleConns)
	}
}

func (t *Transport) trackHTTP1(conn *persistConn) {
	if conn.isClientConn {
		return
	}
	t.idleMu.Lock()
	if t.trackedHTTP1 == nil {
		t.trackedHTTP1 = make(map[*persistConn]struct{})
	}
	t.trackedHTTP1[conn] = struct{}{}
	t.idleMu.Unlock()
}

func (t *Transport) untrackHTTP1(conn *persistConn) {
	t.idleMu.Lock()
	delete(t.trackedHTTP1, conn)
	t.idleMu.Unlock()
}

// CloseInactiveHTTP1Connections snapshots active pooled HTTP/1 connections and
// asynchronously closes those with no read or write progress during timeout.
// Idle connections, new connections and application-owned ClientConns are not
// affected. Cancellation stops the check. Nonpositive timeouts disable it.
// TrackConnectionActivity must be enabled before the transport's first use.
func (t *Transport) CloseInactiveHTTP1Connections(ctx context.Context, timeout time.Duration) {
	if !t.TrackConnectionActivity || timeout <= 0 || ctx.Err() != nil {
		return
	}

	t.idleMu.Lock()
	observed := make(map[*persistConn]uint64)
	for conn := range t.trackedHTTP1 {
		if t.idleLRU.m[conn] == nil {
			observed[conn] = conn.traffic.Load()
		}
	}
	t.idleMu.Unlock()
	if len(observed) == 0 {
		return
	}

	go func() {
		timer := time.NewTimer(timeout)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return
		case <-timer.C:
		}

		for conn, traffic := range observed {
			if ctx.Err() != nil {
				return
			}
			t.idleMu.Lock()
			_, owned := t.trackedHTTP1[conn]
			inactive := owned && t.idleLRU.m[conn] == nil && conn.traffic.Load() == traffic
			t.idleMu.Unlock()
			if inactive {
				conn.close(errNoConnectionTraffic)
			}
		}
	}()
}
