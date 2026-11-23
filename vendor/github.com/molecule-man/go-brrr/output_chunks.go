package brrr

import "sync"

const (
	encodeChunkSize = 64 << 10
	decodeChunkSize = 272 << 10
)

var (
	encodeChunkPool = sync.Pool{New: func() any { b := make([]byte, encodeChunkSize); return &b }}
	decodeChunkPool = sync.Pool{New: func() any { b := make([]byte, decodeChunkSize); return &b }}
)

type chunkWriter struct {
	pool   *sync.Pool
	chunks []*[]byte
	size   int
	n      int
}

func (w *chunkWriter) write(p []byte) {
	for len(p) > 0 {
		if w.n == len(w.chunks)*w.size {
			w.chunks = append(w.chunks, w.pool.Get().(*[]byte))
		}
		copied := copy((*w.chunks[len(w.chunks)-1])[w.n%w.size:], p)
		w.n += copied
		p = p[copied:]
	}
}

func (w *chunkWriter) take() []byte {
	if w.n == 0 {
		w.discard()
		return nil
	}
	out := make([]byte, w.n)
	for i, chunk := range w.chunks {
		copy(out[i*w.size:], *chunk)
	}
	w.discard()
	return out
}

func (w *chunkWriter) discard() {
	for i, chunk := range w.chunks {
		w.pool.Put(chunk)
		w.chunks[i] = nil
	}
	w.chunks = w.chunks[:0]
	w.n = 0
}
