package http2

import (
	"io"
	"sync/atomic"
	"time"
)

var activityEpoch = time.Now()

type connectionActivity struct{ last atomic.Int64 }

func (a *connectionActivity) record() {
	if a != nil {
		a.last.Store(time.Since(activityEpoch).Nanoseconds())
	}
}

func (a *connectionActivity) idleTime() time.Duration {
	return max(0, time.Since(activityEpoch)-time.Duration(a.last.Load()))
}

type activityReader struct {
	reader   io.Reader
	activity *connectionActivity
}

func (r activityReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		r.activity.record()
	}
	return n, err
}

// NetworkIdleTime measures real socket I/O, including partial frames, without
// replacing the net.Conn used for TLS state or connection ownership.
func (cc *ClientConn) NetworkIdleTime() time.Duration { return cc.activity.idleTime() }
