package http2

import (
	"context"
	"errors"
	"time"
)

// Retire optionally probes before closing an idle connection. Retirement prevents
// new reservations immediately, independently of socket scheduling.
func (cc *ClientConn) Retire(ctx context.Context, payload *[8]byte, goAway bool) {
	if goAway {
		cc.retirementGoAway.Store(true)
	}
	cc.SetDoNotReuse()
	if ctx == nil {
		cc.closeIfIdle()
		return
	}

	wait, err := cc.startPing(ctx, cc.closeIfIdle, payload)
	go func() {
		if err == nil {
			err = wait()
		}

		cc.closeIfIdle()
		if err != nil && errors.Is(ctx.Err(), context.DeadlineExceeded) {
			cc.closeForError(err)
		}
	}()
}

// Connections snapshots each pooled connection once. Callers may perform
// maintenance after the pool lock has been released.
func (t *Transport) Connections() []*ClientConn {
	t.connPool.mu.Lock()
	defer t.connPool.mu.Unlock()

	result := make([]*ClientConn, 0, len(t.connPool.keys))
	for conn := range t.connPool.keys {
		result = append(result, conn)
	}

	return result
}

// The successful drain path can optionally send a final GOAWAY. Error closure
// retains its own error handling; it must not append a misleading NO_ERROR.
func (cc *ClientConn) closeRetiredConn() {
	if cc.retirementGoAway.Load() {
		cc.retirementCloseOnce.Do(func() {
			done := make(chan struct{})
			go func() {
				defer close(done)
				cc.wmu.Lock()
				defer cc.wmu.Unlock()
				if cc.closeState.goAwaySent.CompareAndSwap(false, true) {
					cc.fr.WriteGoAway(0, ErrCodeNo, nil)
				}
				cc.bw.Flush()
			}()
			select {
			case <-done:
			case <-time.After(250 * time.Millisecond):
			}
		})
	}
	cc.closeConn()
}

// HasConnection observes a reusable pooled session without reserving a stream.
// Concurrency limits do not hide a live session; normal request admission remains
// responsible for its queue/dial decision. No connection or stream is created.
func (t *Transport) HasConnection(scheme, authority string) bool {
	addr := authorityAddr(scheme, authority)
	t.connPool.mu.Lock()
	defer t.connPool.mu.Unlock()

	for _, conn := range t.connPool.conns[addr] {
		conn.mu.Lock()
		usable := !conn.singleUse && conn.isUsableLocked()
		conn.mu.Unlock()
		if usable {
			return true
		}
	}

	return false
}
