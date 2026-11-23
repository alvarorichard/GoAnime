package brrr

import "github.com/molecule-man/go-brrr/internal/encoder"

// PreparedDictionary is a compound dictionary chunk built once and shared
// across many Writers. See [WriterOptions.Dictionaries]. It is an opaque
// handle; build one with [PrepareDictionary].
type PreparedDictionary struct {
	impl *encoder.PreparedDictionary
}

// PrepareDictionary builds an immutable [PreparedDictionary] from the given
// source bytes, suitable for use as a compound dictionary chunk via
// [WriterOptions.Dictionaries]. The returned dictionary may be shared across
// any number of Writers and goroutines.
//
// The returned dictionary keeps a reference to data; the caller must not
// mutate data while any Writer holding the dictionary is still in use.
//
// Returns an error if data is empty.
func PrepareDictionary(data []byte) (*PreparedDictionary, error) {
	impl, err := encoder.PrepareDictionary(data)
	if err != nil {
		return nil, err
	}
	return &PreparedDictionary{impl: impl}, nil
}
