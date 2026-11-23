// Public package-level constants and option structs.

package brrr

// Compression level constants.
const (
	BestSpeed       = 0
	BestCompression = 11
)

// Window size limits (RFC 7932 Section 9.1).
const (
	minLGWin     = 10
	maxLGWin     = 24
	defaultLGWin = 22
)

// WriterOptions configures advanced tuning knobs for the brotli encoder.
// The compression level is passed positionally to NewWriter / NewWriterOptions.
type WriterOptions struct {
	// Dictionaries are compound dictionary chunks the encoder may reference
	// as backward distances beyond the ring buffer, useful when inputs share
	// content with a known corpus. Build each chunk once with
	// [PrepareDictionary] and share the resulting *PreparedDictionary across
	// any number of Writers. Up to 15 chunks are allowed and compound
	// dictionaries require compression level >= 2. Dictionaries are preserved
	// across Reset.
	Dictionaries []*PreparedDictionary

	// LGWin sets the base-2 logarithm of the sliding window size (10–24).
	// 0 selects the default (22).
	LGWin int

	// SizeHint is the expected total input size in bytes. When set, the
	// encoder uses it to make better decisions about context modeling and
	// hasher selection for large inputs. 0 means unknown; the encoder will
	// auto-estimate from the first Write call.
	//
	// SizeHint is advisory: a lower-than-actual value is safe in both
	// correctness and compression ratio. The encoder transparently
	// promotes its internal hasher to the unbounded variant if buffered
	// input crosses the size-tuned threshold, preserving the bucket state
	// that was learned under the small-hint dispatch.
	//
	// SizeHint applies only to the first stream. [Writer.Reset] clears it.
	// [Writer.ResetWithSizeHint] sets the hint for the next stream.
	SizeHint uint
}

// ReaderOptions configures the brotli decoder.
type ReaderOptions struct {
	// Dictionaries are compound dictionary chunks the decoder will use to
	// resolve backward references beyond the ring buffer. They must match
	// the dictionaries supplied to the encoder. Up to 15 chunks are allowed;
	// each must be non-empty. Dictionaries are preserved across Reset.
	Dictionaries [][]byte
}
