/*
Package brrr implements the Brotli compressed data format (RFC 7932).

Brotli is a lossless compression algorithm that typically produces
better ratios than gzip and zstd at the cost of slower compression.
Decompression is fast at every quality level. Every major browser has
accepted Content-Encoding: br since 2016, which makes brotli a good
match for static web asset pipelines.

# Basic usage

Streaming compression through an io.Writer:

	var buf bytes.Buffer
	w, err := brrr.NewWriter(&buf, 6)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := w.Write(data); err != nil {
		log.Fatal(err)
	}
	if err := w.Close(); err != nil {
		log.Fatal(err)
	}

Streaming decompression through an io.Reader:

	r := brrr.NewReader(src)
	if _, err := io.Copy(dst, r); err != nil {
		log.Fatal(err)
	}

One-shot decompression of a complete in-memory blob:

	out, err := brrr.Decompress(compressed)

# Key types

  - [Writer] implements io.WriteCloser and compresses data written to
    an underlying io.Writer. Close must be called to finalize the
    brotli stream.
  - [Reader] implements io.Reader and decompresses a brotli stream
    read from an underlying io.Reader.
  - [Decompress] is a convenience function for one-shot decompression
    of a complete in-memory blob.

Writer and Reader both expose Reset methods, so the same instance can
be reused across payloads without reallocating per-call buffers. Both
also accept compound dictionaries — useful when the payloads share
content with a known corpus. The encoder takes pre-built
[PreparedDictionary] values via [WriterOptions.Dictionaries], so the
hash table is built once and may be shared across many Writers and
goroutines. The decoder takes raw dictionary bytes via
[ReaderOptions.Dictionaries].

# Choosing a quality level

Quality controls the ratio/speed trade-off on a scale of 0 to 11:

  - Quality 0–1 use fast one-pass and two-pass encoders optimised for
    throughput.
  - Quality 2–9 use a streaming encoder that spends progressively more
    work per byte for better ratios.
  - Quality 10–11 use a Zopfli-style optimal-parsing encoder that
    produces the best ratios but is orders of magnitude slower than
    the lower levels.

The canonical use case for brotli is static asset compression — CSS,
JS, HTML, fonts, WASM — where the data is compressed once at build
time and served many times; quality 11 is the right choice there. For
on-the-fly compression, quality 5–6 gives a good balance of speed and
ratio.

# Compatibility

Output is byte-compatible with the reference implementation at
https://github.com/google/brotli, release v1.2.1. One exception: with a
compound dictionary at quality 3 and 4, this package follows upstream commit
54f74c7, which is not in a release yet. Any conforming brotli decoder can
read streams produced by this package, and this package can decode
any valid brotli stream.
*/
package brrr
