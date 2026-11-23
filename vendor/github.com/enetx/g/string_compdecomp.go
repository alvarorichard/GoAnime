package g

type (
	// A struct that wraps a String for compression.
	compress struct{ str String }

	// A struct that wraps a String for decompression.
	decompress struct{ str String }
)

// Compress returns a compress struct wrapping the given String.
func (s String) Compress() compress { return compress{s} }

// Decompress returns a decompress struct wrapping the given String.
func (s String) Decompress() decompress { return decompress{s} }

// Zlib compresses the wrapped String with zlib; a zero-copy delegate to the
// canonical Bytes implementation.
func (c compress) Zlib() String { return c.str.BytesUnsafe().Compress().Zlib().StringUnsafe() }

// Zlib decompresses the wrapped String with zlib.
func (d decompress) Zlib() Result[String] {
	return d.str.BytesUnsafe().Decompress().Zlib().Map(Bytes.StringUnsafe)
}

// Gzip compresses the wrapped String with gzip; a zero-copy delegate to the
// canonical Bytes implementation.
func (c compress) Gzip() String { return c.str.BytesUnsafe().Compress().Gzip().StringUnsafe() }

// Gzip decompresses the wrapped String with gzip.
func (d decompress) Gzip() Result[String] {
	return d.str.BytesUnsafe().Decompress().Gzip().Map(Bytes.StringUnsafe)
}

// Flate compresses the wrapped String with flate (deflate); a zero-copy
// delegate to the canonical Bytes implementation. It accepts an optional
// compression level, defaulting to 7 and clamping to the valid range [-2, 9].
func (c compress) Flate(level ...int) String {
	return c.str.BytesUnsafe().Compress().Flate(level...).StringUnsafe()
}

// Flate decompresses the wrapped String with flate (deflate).
func (d decompress) Flate() Result[String] {
	return d.str.BytesUnsafe().Decompress().Flate().Map(Bytes.StringUnsafe)
}
