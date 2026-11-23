package http3

import (
	"io"
	"sync"
)

// firstBodyReadReader leaves short reads intact. Empty reads do not consume the
// first-read policy; subsequent nonempty reads use the normal copy buffer.
type firstBodyReadReader struct {
	r    io.Reader
	size int
}

func (r *firstBodyReadReader) Read(b []byte) (int, error) {
	if r.size > 0 {
		b = b[:min(len(b), r.size)]
	}
	n, err := r.r.Read(b)
	if n > 0 || err != nil {
		r.size = 0
	}
	return n, err
}

// Closing a canceled upload must unblock a reader even after response HEADERS
// have arrived. The outer transport can also close the body on an error.
type requestBodyCloser struct {
	io.ReadCloser
	once sync.Once
	err  error
}

func (b *requestBodyCloser) Close() error {
	b.once.Do(func() { b.err = b.ReadCloser.Close() })
	return b.err
}
