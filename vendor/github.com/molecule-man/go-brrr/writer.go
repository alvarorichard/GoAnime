// Streaming brotli writer and public API.

package brrr

import (
	"errors"
	"io"
	"runtime"
	"strconv"
	"sync"

	"github.com/molecule-man/go-brrr/internal/encoder"
)

const minWorkerLevel = 10

const maxLevel = 11

var oneshotCompressors [maxLevel + 1]sync.Pool

type oneshotCompressor struct {
	c      encoder.Compressor
	chunks chunkWriter
}

// Writer compresses data into brotli format.
//
// Callers must Close the Writer to finalize the brotli stream.
//
// At levels 10 and 11, a Writer uses up to two worker goroutines when
// GOMAXPROCS is at least 2. This can reduce latency but uses more memory.
// A Writer does not support concurrent calls.
type Writer struct {
	dst      io.Writer
	err      error
	c        encoder.Compressor
	dicts    []*PreparedDictionary // from WriterOptions, preserved across Reset
	quality  int                   // 0 = one-pass, 1 = two-pass, 2+ = streaming
	lgwin    int
	sizeHint uint // hint for the current stream
	closed   bool
	reused   bool // true after first Reset; suppresses pool release on Close
	parallel bool
}

// NewWriter returns a new Writer compressing data to dst at the given
// quality level. Supported levels are 0 (BestSpeed) through 11
// (BestCompression).
func NewWriter(dst io.Writer, level int) (*Writer, error) {
	return NewWriterOptions(dst, level, WriterOptions{})
}

// NewWriterOptions returns a new Writer compressing data to dst at the given
// quality level with additional tuning options. LGWin range is 10–24; 0
// selects the default (22). Compound dictionaries supplied via opts.Dictionaries
// require level >= 2.
func NewWriterOptions(dst io.Writer, level int, opts WriterOptions) (*Writer, error) {
	return newWriter(dst, level, opts, runtime.GOMAXPROCS(0) >= 2)
}

func newWriter(dst io.Writer, level int, opts WriterOptions, workers bool) (*Writer, error) {
	if err := checkLevel(level); err != nil {
		return nil, err
	}

	lgwin := opts.LGWin
	if lgwin == 0 {
		lgwin = defaultLGWin
	}
	if lgwin < minLGWin || lgwin > maxLGWin {
		return nil, errors.New("brrr: invalid window size: lgwin=" + strconv.Itoa(lgwin) +
			" (must be " + strconv.Itoa(minLGWin) + "–" + strconv.Itoa(maxLGWin) + ")")
	}

	if len(opts.Dictionaries) > encoder.MaxCompoundDicts {
		return nil, encoder.ErrTooManyDicts
	}
	if len(opts.Dictionaries) > 0 && level < 2 {
		return nil, encoder.ErrQualityTooLow
	}

	w := &Writer{
		dst:      dst,
		quality:  level,
		lgwin:    lgwin,
		sizeHint: opts.SizeHint,
		dicts:    opts.Dictionaries,
		parallel: workers && level >= minWorkerLevel,
	}
	w.c = encoder.NewCompressor(w.quality, w.lgwin, w.sizeHint, w.parallel)
	for _, pd := range w.dicts {
		_ = w.c.AttachDictionary(pd.impl)
	}
	if w.parallel {
		// Release the collector if the caller abandons this Writer.
		runtime.SetFinalizer(w, (*Writer).release)
	}
	return w, nil
}

// Compress compresses data at the given quality level and returns the
// brotli-compressed bytes. It is the one-shot counterpart to [Decompress].
// Supported levels are 0 (BestSpeed) through 11 (BestCompression). The exact
// input length is supplied to the encoder as a size hint.
func Compress(data []byte, level int) ([]byte, error) {
	o, err := getOneshotCompressor(level, uint(len(data)))
	if err != nil {
		return nil, err
	}
	err = o.compress(&o.chunks, data)
	if err != nil {
		o.chunks.discard()
		o.drop()
		return nil, err
	}
	out := o.chunks.take()
	oneshotCompressors[level].Put(o)
	return out, nil
}

func getOneshotCompressor(level int, sizeHint uint) (*oneshotCompressor, error) {
	if err := checkLevel(level); err != nil {
		return nil, err
	}
	if v := oneshotCompressors[level].Get(); v != nil {
		o := v.(*oneshotCompressor)
		o.c.ResetSizeHint(sizeHint)
		return o, nil
	}
	o := &oneshotCompressor{
		c:      encoder.NewCompressor(level, defaultLGWin, sizeHint, false),
		chunks: chunkWriter{pool: &encodeChunkPool, size: encodeChunkSize},
	}
	return o, nil
}

func (o *oneshotCompressor) compress(dst io.Writer, data []byte) error {
	if _, err := o.c.Write(dst, data); err != nil {
		return err
	}
	return o.c.Close(dst)
}

func (o *oneshotCompressor) drop() {
	if o.c != nil {
		o.c.Release()
		o.c = nil
	}
}

func checkLevel(level int) error {
	if level < 0 || level > maxLevel {
		return errors.New("brrr: invalid compression level: " + strconv.Itoa(level))
	}
	return nil
}

func (w *chunkWriter) Write(p []byte) (int, error) {
	w.write(p)
	return len(p), nil
}

// Write compresses p and writes it to the underlying writer.
// Data may be buffered internally; call Flush or Close to ensure all
// data is written.
func (w *Writer) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	if w.closed {
		return 0, io.ErrClosedPipe
	}
	n, err := w.c.Write(w.dst, p)
	if err != nil {
		w.err = err
	}
	return n, err
}

// Flush compresses any buffered data and writes it to the underlying
// writer as one or more non-final meta-blocks. Flush does not finalize
// the brotli stream; call Close for that.
func (w *Writer) Flush() error {
	if w.err != nil {
		return w.err
	}
	if w.closed {
		return io.ErrClosedPipe
	}
	w.err = w.c.Flush(w.dst)
	return w.err
}

// Close flushes remaining data, finalizes the brotli stream by writing
// the final empty meta-block, and writes everything to the underlying writer.
// Close does not close the underlying writer.
func (w *Writer) Close() error {
	if w.err != nil {
		return w.err
	}
	if w.closed {
		return nil
	}
	w.closed = true
	w.err = w.c.Close(w.dst)
	if !w.reused {
		w.c.Release()
		w.c = nil
	}
	return w.err
}

// Reset discards the current stream and sets dst for the next stream.
// It clears the size hint. The encoder estimates size from the first Write.
// It keeps compound dictionaries from WriterOptions.
func (w *Writer) Reset(dst io.Writer) {
	w.reset(dst, 0)
}

// ResetWithSizeHint discards the current stream and sets dst for the next stream.
// It uses sizeHint as the expected input size. A zero hint means unknown size.
func (w *Writer) ResetWithSizeHint(dst io.Writer, sizeHint uint) {
	w.reset(dst, sizeHint)
}

func (w *Writer) reset(dst io.Writer, sizeHint uint) {
	w.dst = dst
	w.err = nil
	w.closed = false
	w.reused = true
	w.sizeHint = sizeHint

	if w.c == nil {
		// Compressor was released on a previous Close; re-acquire.
		w.c = encoder.NewCompressor(w.quality, w.lgwin, w.sizeHint, w.parallel)
	} else {
		w.c.ResetSizeHint(w.sizeHint)
	}
	for _, pd := range w.dicts {
		_ = w.c.AttachDictionary(pd.impl)
	}
}

func (w *Writer) release() {
	if w.c != nil {
		w.c.Release()
		w.c = nil
	}
}
