package quic

import (
	"context"
	"os"
	"time"

	"github.com/enetx/quic-go/internal/monotime"
)

// WaitForReadReady waits for another protocol stream to make buffered data
// decodable. It consumes no stream bytes and honors read cancellation, reset,
// connection shutdown and changes to the read deadline. FIN alone does not end
// the wait. Calls must be serialized with Read and Peek.
func (s *ReceiveStream) WaitForReadReady(ctx context.Context, ready <-chan struct{}) error {
	select {
	case s.readOnce <- struct{}{}:
	case <-ctx.Done():
		return context.Cause(ctx)
	}
	defer func() { <-s.readOnce }()
	for {
		s.mutex.Lock()
		var err error
		switch {
		case s.closeForShutdownErr != nil:
			err = s.closeForShutdownErr
		case s.cancelledLocally || s.isRemoteCancellationEffective():
			if s.cancelErr != nil {
				err = s.cancelErr
			} else {
				err = context.Canceled
			}
		}
		deadline := s.deadline
		s.mutex.Unlock()
		if err != nil {
			return err
		}
		if !deadline.IsZero() && !deadline.After(monotime.Now()) {
			return os.ErrDeadlineExceeded
		}
		var timer *time.Timer
		var timeout <-chan time.Time
		if !deadline.IsZero() {
			timer = time.NewTimer(monotime.Until(deadline))
			timeout = timer.C
		}
		select {
		case <-ctx.Done():
			err = context.Cause(ctx)
		case <-timeout:
			err = os.ErrDeadlineExceeded
		case <-ready:
			if timer != nil {
				timer.Stop()
			}
			return nil
		case <-s.readChan:
		}
		if timer != nil {
			timer.Stop()
		}
		if err != nil {
			return err
		}
	}
}

// WaitForReadReady preserves the receive side's cancellation/deadline semantics.
// See [ReceiveStream.WaitForReadReady].
func (s *Stream) WaitForReadReady(ctx context.Context, ready <-chan struct{}) error {
	return s.receiveStr.WaitForReadReady(ctx, ready)
}
