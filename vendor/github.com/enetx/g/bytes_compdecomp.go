package g

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"io"
)

type (
	// A struct that wraps Bytes for compression.
	bcompress struct{ bytes Bytes }

	// A struct that wraps Bytes for decompression.
	bdecompress struct{ bytes Bytes }
)

// Compress returns a bcompress struct wrapping the given Bytes.
func (bs Bytes) Compress() bcompress { return bcompress{bs} }

// Decompress returns a bdecompress struct wrapping the given Bytes.
func (bs Bytes) Decompress() bdecompress { return bdecompress{bs} }

// Zlib compresses the wrapped Bytes using the zlib compression algorithm and
// returns the compressed data as Bytes.
func (c bcompress) Zlib() Bytes {
	buffer := new(bytes.Buffer)
	writer := zlib.NewWriter(buffer)

	_, _ = writer.Write(c.bytes)
	_ = writer.Flush()
	_ = writer.Close()

	return buffer.Bytes()
}

// Zlib decompresses the wrapped Bytes using the zlib compression algorithm and
// returns the decompressed data as a Result[Bytes].
func (d bdecompress) Zlib() Result[Bytes] {
	reader, err := zlib.NewReader(bytes.NewReader(d.bytes))
	if err != nil {
		return Err[Bytes](err)
	}

	defer reader.Close()

	buffer := new(bytes.Buffer)
	if _, err := io.Copy(buffer, reader); err != nil {
		return Err[Bytes](err)
	}

	return Ok(Bytes(buffer.Bytes()))
}

// Gzip compresses the wrapped Bytes using the gzip compression format and
// returns the compressed data as Bytes.
func (c bcompress) Gzip() Bytes {
	buffer := new(bytes.Buffer)
	writer := gzip.NewWriter(buffer)

	_, _ = writer.Write(c.bytes)
	_ = writer.Flush()
	_ = writer.Close()

	return buffer.Bytes()
}

// Gzip decompresses the wrapped Bytes using the gzip compression format and
// returns the decompressed data as a Result[Bytes].
func (d bdecompress) Gzip() Result[Bytes] {
	reader, err := gzip.NewReader(bytes.NewReader(d.bytes))
	if err != nil {
		return Err[Bytes](err)
	}

	defer reader.Close()

	buffer := new(bytes.Buffer)
	if _, err := io.Copy(buffer, reader); err != nil {
		return Err[Bytes](err)
	}

	return Ok(Bytes(buffer.Bytes()))
}

// Flate compresses the wrapped Bytes using the flate (deflate) compression
// algorithm and returns the compressed data as Bytes.
// It accepts an optional compression level. If no level is provided, it
// defaults to 7. The level is clamped to the valid flate range [-2, 9] to
// avoid an invalid-level error that would otherwise return a nil writer and
// panic on use.
func (c bcompress) Flate(level ...int) Bytes {
	buffer := new(bytes.Buffer)

	l := 7
	if len(level) != 0 {
		l = level[0]
	}

	if l < flate.HuffmanOnly {
		l = flate.HuffmanOnly
	} else if l > flate.BestCompression {
		l = flate.BestCompression
	}

	writer, _ := flate.NewWriter(buffer, l)

	_, _ = writer.Write(c.bytes)
	_ = writer.Flush()
	_ = writer.Close()

	return buffer.Bytes()
}

// Flate decompresses the wrapped Bytes using the flate (deflate) compression
// algorithm and returns the decompressed data as a Result[Bytes].
func (d bdecompress) Flate() Result[Bytes] {
	reader := flate.NewReader(bytes.NewReader(d.bytes))
	defer reader.Close()

	buffer := new(bytes.Buffer)
	if _, err := io.Copy(buffer, reader); err != nil {
		return Err[Bytes](err)
	}

	return Ok(Bytes(buffer.Bytes()))
}
