package http2

import (
	"context"
	"crypto/rand"
)

// startPing registers or joins the payload before returning. Only the actual
// write and acknowledgement wait are asynchronous, so callers cannot race an
// acknowledgement with delayed duplicate registration.
func (cc *ClientConn) startPing(ctx context.Context, written func(), payload *[8]byte) (func() error, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	c := make(chan struct{})
	// Generate a random payload
	var p [8]byte
	joined := false
	for {
		if payload != nil {
			p = *payload
		} else {
			// Go 1.27's crypto/rand.Read always fills p or terminates the process.
			rand.Read(p[:])
		}
		cc.mu.Lock()
		// check for dup before insert
		if existing, found := cc.pings[p]; !found {
			cc.pings[p] = c
			cc.mu.Unlock()
			break
		} else if payload != nil {
			c, joined = existing, true
			cc.mu.Unlock()
			break
		}
		cc.mu.Unlock()
	}
	var pingError error
	errc := make(chan struct{})
	go func() {
		if joined {
			return
		}
		cc.wmu.Lock()
		if pingError = cc.fr.WritePing(false, p); pingError != nil {
			cc.wmu.Unlock()
			close(errc)
			return
		}
		if pingError = cc.bw.Flush(); pingError != nil {
			cc.wmu.Unlock()
			close(errc)
			return
		}
		cc.wmu.Unlock()
		if written != nil {
			written()
		}
	}()
	return func() error {
		select {
		case <-c:
			return nil
		case <-errc:
			return pingError
		case <-ctx.Done():
			return ctx.Err()
		case <-cc.readerDone:
			// connection closed
			return cc.readerErr
		}
	}, nil
}
