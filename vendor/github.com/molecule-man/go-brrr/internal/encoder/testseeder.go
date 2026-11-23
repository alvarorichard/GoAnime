// Test-only seeders that poke streaming encoder state directly. Used by
// integration tests at the public-API layer that need to fast-forward the
// stream position to exercise the 32-bit wrap edge case without compressing
// gigabytes of input.

package encoder

// SeedStreamPosForTest advances the streaming encoder's stream-position fields
// to pos so the next Write straddles the 32-bit wrap boundary. The ring
// buffer is left empty (the encoder only references data subsequently
// written), and lgwin caps distances so references stay within the
// in-metablock region.
//
// Returns false if c is a fast (q0/q1) compressor, which does not have a
// streaming position concept.
func SeedStreamPosForTest(c Compressor, pos uint64) bool {
	var es *encodeState
	switch enc := c.(type) {
	case *encoderArena:
		es = &enc.encodeState
	case *encoderSplit:
		es = &enc.encodeState
	default:
		return false
	}
	es.inputPos = pos
	es.lastProcessedPos = pos
	es.lastFlushPos = pos
	es.ringBufPos = uint32(pos & uint64(es.mask))
	return true
}
