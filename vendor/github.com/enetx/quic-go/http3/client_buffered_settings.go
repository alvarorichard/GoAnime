package http3

import (
	"github.com/enetx/quic-go"
	"github.com/enetx/quic-go/quicvarint"
)

// Browser event loops process received control data before the next application
// send. Dispatch a complete buffered SETTINGS prelude without waiting for bytes
// that have not arrived. Partial/late control streams keep the normal reader.
// https://github.com/google/quiche/blob/132bfd90ea370faeacee616c70d10210bb12ba66/quiche/quic/core/http/quic_spdy_session.cc
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/third_party/rust/neqo-http3/src/connection.rs
func bufferedControlSettings(str *quic.ReceiveStream) bool {
	read := func(offset int) (uint64, int, bool) {
		var prefix [24]byte // Stream type, frame type and length: three QUIC varints.
		if _, err := str.TryPeek(prefix[:offset+1]); err != nil {
			return 0, 0, false
		}

		end := offset + 1<<(prefix[offset]>>6)
		if _, err := str.TryPeek(prefix[:end]); err != nil {
			return 0, 0, false
		}

		value, _, err := quicvarint.Parse(prefix[offset:end])
		return value, end, err == nil
	}

	typ, offset, ok := read(0)
	if !ok || typ != uint64(streamTypeControlStream) {
		return false
	}

	typ, offset, ok = read(offset)
	if !ok || typ != 4 {
		return false
	}

	length, offset, ok := read(offset)
	if !ok || length > 64<<10 {
		return false // Bound speculative work; the ordinary parser owns larger frames.
	}

	_, err := str.TryPeek(make([]byte, offset+int(length)))
	return err == nil
}

func (c *ClientConn) dispatchBufferedInitialStreams() {
	// Bound one synchronous dispatch even if a peer opens a large stream range.
	for range 64 {
		str, err := c.conn.TryAcceptUniStream()
		if err != nil {
			return
		}

		if bufferedControlSettings(str) {
			c.rawConn.handleUnidirectionalStreamMode(str, false, true)
		} else {
			go c.handleUnidirectionalStream(str)
		}
	}
}
