// streaming_compressor.go — Compressor implementation for q>=2 streaming
// encoders (encoderArena, encoderSplit). The shared Write/Flush/Close logic
// previously lived in writer.go and is now owned by the encoder.

package encoder

import "io"

// Compile-time assertion that both q>=2 encoders satisfy Compressor.
var (
	_ Compressor = (*encoderArena)(nil)
	_ Compressor = (*encoderSplit)(nil)
)

// streamWrite drives input through a streaming encoder, emitting output to dst
// when the ring buffer fills.
func streamWrite(se streamEncoder, dst io.Writer, p []byte) (int, error) {
	se.updateSizeHint(uint(len(p)))
	written := len(p)
	for len(p) > 0 {
		remaining := se.remainingInputBlockSize()
		if remaining == 0 {
			out := se.encodeData(false, false)
			if len(out) > 0 {
				if _, err := dst.Write(out); err != nil {
					return 0, err
				}
			}
			continue
		}
		chunk := min(uint(len(p)), remaining)
		se.copyInputToRingBuffer(p[:chunk])
		p = p[chunk:]
	}
	return written, nil
}

// streamFlush emits a non-final meta-block followed by a byte-padding metadata
// block (ISLAST=0, MNIBBLES=11, reserved=0, MSKIPBYTES=00) so the output is
// byte-aligned for downstream readers.
func streamFlush(se streamEncoder, dst io.Writer) error {
	out := se.encodeData(false, true)
	if len(out) > 0 {
		if _, err := dst.Write(out); err != nil {
			return err
		}
	}
	lastBytes, lastBytesBits := se.trailingBits()
	seal := uint32(lastBytes)
	sealBits := uint(lastBytesBits)
	se.clearTrailingBits()
	seal |= 0x6 << sealBits
	sealBits += 6

	var padding [3]byte
	padding[0] = byte(seal)
	if sealBits > 8 {
		padding[1] = byte(seal >> 8)
	}
	if sealBits > 16 {
		padding[2] = byte(seal >> 16)
	}
	n := (sealBits + 7) / 8
	_, err := dst.Write(padding[:n])
	return err
}

// streamClose finalizes the brotli stream by writing the last meta-block and
// any trailing sub-byte bits.
func streamClose(se streamEncoder, dst io.Writer) error {
	out := se.encodeData(true, false)
	if len(out) > 0 {
		if _, err := dst.Write(out); err != nil {
			return err
		}
	}
	lastBytes, lastBytesBits := se.trailingBits()
	if lastBytesBits > 0 {
		var trailing [2]byte
		trailing[0] = byte(lastBytes)
		trailing[1] = byte(lastBytes >> 8)
		n := (lastBytesBits + 7) / 8
		if _, err := dst.Write(trailing[:n]); err != nil {
			return err
		}
	}
	return nil
}

// AttachDictionary attaches a compound dictionary chunk. Shared by both
// streaming encoder variants via the embedded encodeState.
func (c *encoderCore) AttachDictionary(pd *PreparedDictionary) error {
	return c.attachDictionary(pd)
}

// encoderArena Compressor surface. ResetSizeHint/Write/Flush/Close are defined here
// (rather than on encoderCore) so method resolution dispatches to
// encoderArena.reset and encoderArena.encodeData rather than the embedded
// encodeState's reset.

func (e *encoderArena) Write(dst io.Writer, p []byte) (int, error) {
	return streamWrite(e, dst, p)
}

func (e *encoderArena) Flush(dst io.Writer) error {
	return streamFlush(e, dst)
}

func (e *encoderArena) Close(dst io.Writer) error {
	return streamClose(e, dst)
}

func (e *encoderArena) ResetSizeHint(sizeHint uint) {
	e.reset(e.quality, e.lgwin, sizeHint)
}

// Release returns the encoder to its pool. The encoder must not be used
// after Release.
func (e *encoderArena) Release() {
	e.releaseBuffers()
	poolEncoderArena.Put(e)
}

// encoderSplit Compressor surface.

func (e *encoderSplit) Write(dst io.Writer, p []byte) (int, error) {
	return streamWrite(e, dst, p)
}

func (e *encoderSplit) Flush(dst io.Writer) error {
	return streamFlush(e, dst)
}

func (e *encoderSplit) Close(dst io.Writer) error {
	return streamClose(e, dst)
}

func (e *encoderSplit) ResetSizeHint(sizeHint uint) {
	e.reset(e.quality, e.lgwin, sizeHint)
}

// Release returns the encoder to its pool. The encoder must not be used
// after Release.
func (e *encoderSplit) Release() {
	e.releaseBuffers()
	poolEncoderSplit.Put(e)
}
