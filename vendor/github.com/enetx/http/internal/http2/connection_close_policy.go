package http2

import (
	"slices"
	"sync"
	"sync/atomic"
	"time"
)

// ConnectionCloseInfo describes a terminal connection close. Cause is nil for
// local idle/drain closes; read-loop closes retain the original socket error.
type ConnectionCloseInfo struct {
	// Cause is the original terminal error, or nil for local idle/drain closes.
	Cause error
	// ActiveStreams counts streams still owned when the close begins.
	ActiveStreams int
	// PeerGoAway reports whether the peer previously sent a GOAWAY frame.
	PeerGoAway bool
	// ProtocolError reports a connection-level frame parsing error. Its required
	// code takes precedence over the options returned by the policy.
	ProtocolError bool
	// Code is the required code when ProtocolError is true.
	Code ErrCode
	// Detail is the framer's diagnostic for a protocol error, when available.
	Detail string
}

// ConnectionCloseOptions selects an optional final GOAWAY. Existing protocol
// error or graceful-shutdown GOAWAY frames take precedence over this option.
type ConnectionCloseOptions struct {
	// SendGoAway requests a frame when none has already been sent.
	SendGoAway bool
	// Code is the final HTTP/2 error code.
	Code ErrCode
	// LastStreamID identifies the last accepted server-initiated stream.
	LastStreamID uint32
	// DebugData supplies optional diagnostic bytes. They are copied and bounded
	// to fit a minimum-size HTTP/2 frame before the asynchronous write begins.
	DebugData []byte
}

type connectionCloseState struct {
	policyOnce sync.Once
	goAwaySent atomic.Bool
}

func (cc *ClientConn) applyClosePolicy(cause error) {
	if cc.fingerprint == nil || cc.fingerprint.ConnectionClosePolicy == nil {
		return
	}

	cc.closeState.policyOnce.Do(func() {
		cc.mu.Lock()
		info := ConnectionCloseInfo{Cause: cause, ActiveStreams: len(cc.streams), PeerGoAway: cc.goAway != nil}
		cc.mu.Unlock()

		if code, ok := cause.(ConnectionError); ok {
			info.ProtocolError = true
			info.Code = ErrCode(code)
			if detail := cc.fr.ErrorDetail(); detail != nil {
				info.Detail = detail.Error()
			}
		}

		options := cc.fingerprint.ConnectionClosePolicy(info)
		if info.ProtocolError {
			options.SendGoAway = true
			options.Code = info.Code
		}

		if !options.SendGoAway && !cc.closeState.goAwaySent.Load() {
			return
		}

		options.DebugData = slices.Clone(options.DebugData[:min(len(options.DebugData), initialMaxFrameSize-8)])

		done := make(chan struct{})
		go func() {
			defer close(done)
			cc.wmu.Lock()
			defer cc.wmu.Unlock()
			if options.SendGoAway && cc.closeState.goAwaySent.CompareAndSwap(false, true) {
				cc.fr.WriteGoAway(options.LastStreamID&0x7fffffff, options.Code, options.DebugData)
			}

			// Also flush an existing buffered protocol-error GOAWAY instead
			// of appending a misleading second NO_ERROR frame.
			cc.bw.Flush()
		}()
		select {
		case <-done:
		case <-time.After(250 * time.Millisecond):
		}
	})
}
