package http

import (
	"fmt"
	"io"
)

// HTTP1WriteInfo describes the serialized HTTP/1 request before its body is
// copied. HeaderBytes includes the request line and terminating empty line.
type HTTP1WriteInfo struct {
	// HeaderBytes counts the actual serialized request line and header section.
	HeaderBytes int64
	// ContentLength is the declared body size; -1 denotes an unknown length.
	ContentLength int64
	// InMemory is true only for the standard in-memory reader types after
	// unwrapping the transport's own body tracking and io.NopCloser wrappers.
	InMemory bool
	// Chunked reports the actual selected Transfer-Encoding.
	Chunked bool
}

// HTTP1WriteOptions selects request packing without changing field order or
// framing. The zero value retains the transport's native behavior.
type HTTP1WriteOptions struct {
	// FlushHeaders sends buffered headers before copying the body. It cannot
	// suppress a required Expect/Continue or potentially blocking-body flush.
	FlushHeaders bool
	// BodyReadSize bounds each body read. Zero retains native copy fast paths;
	// positive values select explicit buffered copying, including file bodies.
	// The supported range is zero through 1 MiB.
	BodyReadSize int
}

type http1WritePolicyKey struct{}

type headerWriteCounter struct {
	io.Writer
	bytes int64
}

func (c *headerWriteCounter) Write(p []byte) (int, error) {
	n, err := c.Writer.Write(p)
	c.bytes += int64(n)
	return n, err
}

func (c *headerWriteCounter) WriteString(s string) (int, error) {
	n, err := io.WriteString(c.Writer, s)
	c.bytes += int64(n)
	return n, err
}

func (r *Request) countHTTP1Headers(w io.Writer) (*headerWriteCounter, func(HTTP1WriteInfo) HTTP1WriteOptions) {
	policy, _ := r.Context().Value(http1WritePolicyKey{}).(func(HTTP1WriteInfo) HTTP1WriteOptions)
	if policy == nil {
		return nil, nil
	}

	return &headerWriteCounter{Writer: w}, policy
}

func (t *transferWriter) configureHTTP1Write(
	counter *headerWriteCounter,
	policy func(HTTP1WriteInfo) HTTP1WriteOptions,
) error {
	options := policy(HTTP1WriteInfo{
		HeaderBytes: counter.bytes, ContentLength: t.ContentLength,
		InMemory: isKnownInMemoryReader(t.Body), Chunked: chunked(t.TransferEncoding),
	})
	if options.BodyReadSize < 0 || options.BodyReadSize > 1<<20 {
		return fmt.Errorf("http: HTTP/1 body read size must be between zero and 1 MiB")
	}

	t.FlushHeaders = t.FlushHeaders || options.FlushHeaders
	t.bodyReadSize = options.BodyReadSize
	return nil
}
