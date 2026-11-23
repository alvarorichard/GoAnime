// Fast compressor: q0 (one-pass) and q1 (two-pass).
// Compresses incrementally in fragments of 1<<lgwin, retaining at most one
// fragment in memory; remaining data is emitted on Flush/Close.

package encoder

import (
	"errors"
	"io"
	"sync"
)

var (
	errFastCompoundDict = errors.New("brrr: compound dictionaries require quality >= 2")

	poolOnePassArena = sync.Pool{New: func() any { return new(onePassArena) }}
	poolTwoPassArena = sync.Pool{New: func() any { return new(twoPassArena) }}

	poolFastCompressor = sync.Pool{New: func() any { return new(fastCompressor) }}
)

// fastCompressor implements Compressor for q0 and q1. Input is compressed
// incrementally; at most one 1<<lgwin fragment is buffered in memory.
type fastCompressor struct {
	onePass *onePassArena // non-nil when quality == 0
	twoPass *twoPassArena // non-nil when quality == 1

	buf        []byte // pending input, at most one fragment (1<<lgwin bytes)
	outBuf     []byte // scratch for compressed output
	table      []uint32
	commandBuf []uint32 // q1 only
	literalBuf []byte   // q1 only

	carryBits uint // number of valid sub-byte bits held in carry (0..7)
	quality   int
	lgwin     int

	carry       byte // trailing output bits (carryBits of them) not yet byte-complete
	wroteHeader bool
}

// newFastCompressor returns a Compressor for q0 or q1, acquiring its arena
// from the appropriate pool.
func newFastCompressor(quality, lgwin int) *fastCompressor {
	c := poolFastCompressor.Get().(*fastCompressor)
	c.quality = quality
	c.lgwin = lgwin
	switch quality {
	case 0:
		c.onePass = poolOnePassArena.Get().(*onePassArena)
		c.onePass.initCommandPrefixCodes()
	case 1:
		c.twoPass = poolTwoPassArena.Get().(*twoPassArena)
	}
	return c
}

// Write compresses input incrementally. Full fragments (1<<lgwin bytes) are
// emitted to dst as they fill; at most one fragment is retained in memory, so
// total buffering is bounded regardless of stream length. The final fragment
// is deferred to Flush/Close, where it may be marked last.
func (c *fastCompressor) Write(dst io.Writer, p []byte) (int, error) {
	total := len(p)
	blockSizeLimit := 1 << c.lgwin
	for len(p) > 0 {
		// With nothing buffered and more than a full fragment available,
		// compress straight from p to avoid copying.
		if len(c.buf) == 0 && len(p) > blockSizeLimit {
			if err := c.emitFragment(dst, p[:blockSizeLimit], false); err != nil {
				return total - len(p), err
			}
			p = p[blockSizeLimit:]
			continue
		}
		room := blockSizeLimit - len(c.buf)
		take := min(room, len(p))
		c.buf = append(c.buf, p[:take]...)
		p = p[take:]
		// Emit only when the fragment is full and more input follows, so the
		// last fragment stays buffered for Flush/Close.
		if len(c.buf) == blockSizeLimit && len(p) > 0 {
			if err := c.emitFragment(dst, c.buf, false); err != nil {
				return total - len(p), err
			}
			c.buf = c.buf[:0]
		}
	}
	return total, nil
}

// Flush emits buffered input as a non-final meta-block, then byte-aligns the
// stream so everything written so far is decodable from the bytes on the wire.
// The brotli stream is not finalized.
//
// Between fragments the encoder carries up to seven trailing bits, which is
// correct while the bitstream continues internally. At a flush it is not: those
// bits hold the tail of the data just written, so leaving them behind means a
// streaming consumer decodes a short prefix and blocks. The reference encoder
// injects a byte-padding block for exactly this reason.
func (c *fastCompressor) Flush(dst io.Writer) error {
	if len(c.buf) > 0 {
		if err := c.emitFragment(dst, c.buf, false); err != nil {
			return err
		}
		c.buf = c.buf[:0]
	}
	return c.padToByteBoundary(dst)
}

// Close emits any remaining buffered input as the final meta-block, finalizing
// the brotli stream.
func (c *fastCompressor) Close(dst io.Writer) error {
	if err := c.emitFragment(dst, c.buf, true); err != nil {
		return err
	}
	c.buf = c.buf[:0]
	return nil
}

// ResetSizeHint clears stream state. q0 and q1 ignore the size hint.
// The compressor keeps its arena for reuse.
func (c *fastCompressor) ResetSizeHint(uint) {
	c.buf = c.buf[:0]
	c.carry = 0
	c.carryBits = 0
	c.wroteHeader = false
	if c.quality == 0 {
		c.onePass.initCommandPrefixCodes()
	}
}

// AttachDictionary reports that q0/q1 do not support compound dictionaries.
// Writer is expected to validate this at construction so this is a defensive
// path.
func (c *fastCompressor) AttachDictionary(*PreparedDictionary) error {
	return errFastCompoundDict
}

// Release returns the arena and any owned scratch buffers to their pools.
func (c *fastCompressor) Release() {
	if c.onePass != nil {
		poolOnePassArena.Put(c.onePass)
		c.onePass = nil
	}
	if c.twoPass != nil {
		poolTwoPassArena.Put(c.twoPass)
		c.twoPass = nil
	}
	c.buf = c.buf[:0]
	c.carry = 0
	c.carryBits = 0
	c.quality = 0
	c.lgwin = 0
	c.wroteHeader = false
	poolFastCompressor.Put(c)
}

// emitFragment compresses a single fragment (at most 1<<lgwin bytes) and writes
// the resulting whole bytes to dst. Trailing sub-byte bits are retained in
// c.carry and seeded into the next fragment, so consecutive fragments form one
// continuous bitstream. If isLast is true the fragment finalizes the stream and
// the output is byte-aligned.
func (c *fastCompressor) emitFragment(dst io.Writer, block []byte, isLast bool) error {
	// Worst case output: uncompressed meta-block = header + data + padding.
	needed := len(block)*2 + 1024
	if len(c.outBuf) < needed {
		c.outBuf = make([]byte, needed)
	}
	c.outBuf[0] = c.carry
	b := bitWriter{buf: c.outBuf, bitOffset: c.carryBits}

	if !c.wroteHeader {
		// For quality 0-1 the C reference clamps the header lgwin to at
		// least 18 since these modes don't use a sliding window.
		headerLGWin := max(c.lgwin, 18)
		lastBytes, lastBytesBits := encodeWindowBits(headerLGWin)
		b.writeBits(uint(lastBytesBits), uint64(lastBytes))
		c.wroteHeader = true
	}

	// Size and clear the hash table for this fragment, matching the C reference.
	htsize := fastHashTableSize(c.quality, len(block))
	switch c.quality {
	case 0:
		var smallTable32 [1024]uint32
		var table []uint32
		if htsize <= len(smallTable32) {
			table = smallTable32[:htsize]
		} else {
			if len(c.table) < htsize {
				c.table = make([]uint32, htsize)
			}
			table = c.table[:htsize]
		}
		clear(table)
		compressFragmentFast(c.onePass, block, isLast, table, &b)
	case 1:
		bufSize := min(len(block), twoPassBlockSize)
		if len(c.commandBuf) < bufSize {
			c.commandBuf = make([]uint32, bufSize)
		}
		if len(c.literalBuf) < bufSize {
			c.literalBuf = make([]byte, bufSize)
		}
		if len(c.table) < htsize {
			c.table = make([]uint32, htsize)
		}
		table := c.table[:htsize]
		clear(table)
		compressFragmentTwoPass(c.twoPass, block, isLast,
			c.commandBuf[:bufSize], c.literalBuf[:bufSize], table, &b)
	}

	// Emit whole bytes; retain any trailing sub-byte bits for the next fragment.
	n := b.bitOffset / 8
	if n > 0 {
		if _, err := dst.Write(c.outBuf[:n]); err != nil {
			return err
		}
	}
	// Retain any trailing sub-byte bits to seed the next fragment. A last
	// fragment is byte-aligned by compressFragment*, so n*8 == bitOffset and
	// carryBits is 0.
	c.carry = c.outBuf[n]
	c.carryBits = b.bitOffset & 7
	return nil
}

// padToByteBoundary flushes the pending sub-byte carry by appending an empty
// metadata meta-block and zero-padding to the next byte boundary.
//
// The six header bits are ISLAST=0, MNIBBLES=3 (the metadata marker),
// reserved=0, MSKIPBYTES=0 — the value 0x6 written LSB-first, which is what
// InjectBytePaddingBlock emits in the reference encoder's encode.c. Decoders
// skip metadata blocks, so the block costs at most two bytes and leaves the
// stream ready to continue at a byte boundary.
func (c *fastCompressor) padToByteBoundary(dst io.Writer) error {
	if c.carryBits == 0 {
		return nil
	}
	seal := uint32(c.carry) | 0x6<<c.carryBits
	sealBits := c.carryBits + 6 // at most 7+6 = 13, so never more than 2 bytes
	out := [2]byte{byte(seal), byte(seal >> 8)}
	if _, err := dst.Write(out[:(sealBits+7)/8]); err != nil {
		return err
	}
	c.carry = 0
	c.carryBits = 0
	return nil
}

// fastHashTableSize returns the hash table size for a given quality and block
// size, matching the C reference GetHashTable logic.
func fastHashTableSize(quality, blockSize int) int {
	maxTableSize := 1 << 15 // q0
	if quality == 1 {
		maxTableSize = 1 << 17
	}
	htsize := 256
	for htsize < maxTableSize && htsize < blockSize {
		htsize <<= 1
	}
	// Q0 requires odd-bit tables (9, 11, 13, 15).
	if quality == 0 && (htsize&0xAAAAA) == 0 {
		htsize <<= 1
	}
	return htsize
}
