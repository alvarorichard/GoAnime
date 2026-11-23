// Ring buffer write support for the streaming encoder (quality >= 2).

package encoder

import "sync"

// ringBufPool recycles large ring buffer allocations to avoid repeated
// multi-megabyte heap allocations and the associated zero-initialization.
// The pool is safe because initRingBuffer and copyInputToRingBuffer
// explicitly zero the prefix, suffix, and trailing bytes that the
// compressor reads beyond the copied input data.
var ringBufPool sync.Pool

// initRingBuffer allocates (or grows) the ring buffer's backing array.
func (e *encodeState) initRingBuffer(buflen uint32) {
	needed := int(2 + buflen + 7)
	if cap(e.ringBufAlloc) >= needed {
		old := e.ringBufAlloc
		e.ringBufAlloc = e.ringBufAlloc[:needed]
		if len(old) < needed {
			clear(e.ringBufAlloc[len(old):])
		}
	} else {
		var newAlloc []byte
		if v := ringBufPool.Get(); v != nil {
			bp := v.(*[]byte)
			if cap(*bp) >= needed {
				newAlloc = (*bp)[:needed]
			}
		}
		if newAlloc == nil {
			newAlloc = make([]byte, needed)
		}
		if e.ringBufAlloc != nil {
			copy(newAlloc, e.ringBufAlloc)
			old := e.ringBufAlloc[:cap(e.ringBufAlloc)]
			ringBufPool.Put(&old)
		}
		e.ringBufAlloc = newAlloc
	}
	e.data = e.ringBufAlloc[2:]
	e.ringBufAlloc[0] = 0
	e.ringBufAlloc[1] = 0
	clear(e.data[buflen : buflen+7])
}

// releaseRingBuffer returns the ring buffer allocation to the pool for
// reuse by future encoders. Called from Writer.Close to avoid repeated
// multi-megabyte allocations in one-shot usage patterns.
func (e *encodeState) releaseRingBuffer() {
	if e.ringBufAlloc != nil {
		buf := e.ringBufAlloc[:cap(e.ringBufAlloc)]
		ringBufPool.Put(&buf)
		e.ringBufAlloc = nil
		e.data = nil
	}
}

// copyInputToRingBuffer copies input bytes into the encoder's ring buffer,
// handling wrap-around, tail mirroring, and the 2-byte prefix.
func (e *encodeState) copyInputToRingBuffer(input []byte) {
	n := uint32(len(input))
	size := e.ringBufSize
	rbMask := size - 1
	tailSize := uint32(1) << e.lgblock

	// Path A: small first write. Allocate only n bytes instead of the full
	// ring buffer when the first input is smaller than the tail/block size.
	if e.ringBufPos == 0 && n < tailSize && cap(e.ringBufAlloc) < int(2+size+tailSize+7) {
		e.ringBufPos = n
		e.initRingBuffer(n)
		copy(e.data[:n], input)
		e.inputPos += uint64(n)
		return
	}

	// Path B: lazy full allocation. Grow to full size on second write or
	// when the first write is >= tailSize.
	if uint32(len(e.ringBufAlloc)) < 2+size+tailSize+7 {
		e.initRingBuffer(size + tailSize)
		e.data[size-2] = 0
		e.data[size-1] = 0
	}

	maskedPos := e.ringBufPos & rbMask

	// Write tail: mirror beginning of ring buffer after main area,
	// so reads that cross the buffer boundary see contiguous data.
	if maskedPos < tailSize {
		p := size + maskedPos
		cnt := min(n, tailSize-maskedPos)
		copy(e.data[p:p+cnt], input[:cnt])
	}

	// Copy data into ring buffer.
	if maskedPos+n <= size {
		// Single contiguous write.
		copy(e.data[maskedPos:maskedPos+n], input)
	} else {
		// Split: fill to end of buffer, then wrap to beginning.
		totalSize := size + tailSize
		first := min(n, totalSize-maskedPos)
		copy(e.data[maskedPos:maskedPos+first], input[:first])
		wrap := size - maskedPos
		copy(e.data[:n-wrap], input[wrap:])
	}

	// Copy last 2 bytes of the ring buffer into the 2-byte prefix area,
	// so backward-reference boundary checks can look 2 bytes before
	// the buffer start without special-casing.
	e.ringBufAlloc[0] = e.data[size-2]
	e.ringBufAlloc[1] = e.data[size-1]

	// Update position, preserving the "not first lap" high bit.
	notFirstLap := e.ringBufPos & (1 << 31)
	const rbPosMask = (1 << 31) - 1
	e.ringBufPos = (e.ringBufPos & rbPosMask) + (n & rbPosMask)
	if notFirstLap != 0 {
		e.ringBufPos |= 1 << 31
	}

	e.inputPos += uint64(n)

	// On the first lap, zero 7 bytes after written data for deterministic
	// hashing (prevents reads of uninitialized memory by LOAD64-style
	// hash functions).
	if e.ringBufPos <= rbMask {
		newPos := e.ringBufPos & rbMask
		clear(e.data[newPos : newPos+7])
	}
}

// ringLimit returns the exclusive source limit for the byte at offset bestLen.
// It returns zero when the current probe crosses the ring end.
func ringLimit(mask, curMasked, bestLen uint) uint {
	lim := mask + 1 - bestLen
	if curMasked >= lim {
		return 0
	}
	return lim
}
