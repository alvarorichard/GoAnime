package profiles

// HTTP1Config describes request buffering for an owned HTTP/1 transport.
// Zero values preserve native buffer/read defaults and header merge decisions.
type HTTP1Config struct {
	// WriteBufferBytes is the outgoing buffer size; zero selects native defaults.
	WriteBufferBytes int
	// BodyReadBytes bounds body reads; zero retains native copy optimizations.
	BodyReadBytes int
	// HeaderBodyMergeBytes limits combined serialized headers and an in-memory
	// body. Nil retains native decisions; zero always separates a nonempty body.
	HeaderBodyMergeBytes *int64
}

// WriteBufferSize sets the HTTP/1 write buffer in bytes. Zero uses the default.
func (c *HTTP1Config) WriteBufferSize(size int) *HTTP1Config {
	c.WriteBufferBytes = size
	return c
}

// BodyReadSize bounds each request body read. Zero keeps native copy fast paths.
func (c *HTTP1Config) BodyReadSize(size int) *HTTP1Config {
	c.BodyReadBytes = size
	return c
}

// HeaderBodyMergeLimit combines a known in-memory body with serialized headers
// only within limit bytes. Zero separates bodies; negative values are invalid.
func (c *HTTP1Config) HeaderBodyMergeLimit(limit int64) *HTTP1Config {
	c.HeaderBodyMergeBytes = &limit
	return c
}
