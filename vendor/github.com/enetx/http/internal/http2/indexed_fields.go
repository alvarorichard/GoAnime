package http2

import (
	"sync"

	"golang.org/x/net/http2/hpack"
)

// Track only membership and eviction of entries the official encoder actually
// inserted. The encoder still owns all indexes and emits all bytes. This bounded
// side table lets an insertion-size policy preserve existing exact references.
type indexedFields struct {
	entries  []hpack.HeaderField
	counts   map[hpack.HeaderField]int
	size     uint64
	capacity uint32
}

func newIndexedFields(capacity uint32) *indexedFields {
	return &indexedFields{counts: make(map[hpack.HeaderField]int), capacity: capacity}
}

func (t *indexedFields) contains(field hpack.HeaderField) bool { return t.counts[field] > 0 }

func (t *indexedFields) resize(capacity uint32) {
	t.capacity = capacity
	for t.size > uint64(capacity) {
		field := t.entries[0]
		t.size -= uint64(field.Size())
		if t.counts[field]--; t.counts[field] == 0 {
			delete(t.counts, field)
		}
		t.entries[0] = hpack.HeaderField{}
		t.entries = t.entries[1:]
	}
}

func (t *indexedFields) add(field hpack.HeaderField) {
	t.entries = append(t.entries, field)
	t.counts[field]++
	t.size += uint64(field.Size())
	t.resize(t.capacity)
}

// RFC 7541 Appendix A defines 61 static entries. Derive exact matches using
// the official decoder instead of copying its table or depending on internals.
var staticHPACKFields = sync.OnceValue(func() map[hpack.HeaderField]bool {
	fields := make(map[hpack.HeaderField]bool, 61)
	for _, field := range staticHPACKEntries() {
		fields[field] = true
	}
	return fields
})

var staticHPACKEntries = sync.OnceValue(func() []hpack.HeaderField {
	decoder := hpack.NewDecoder(4096, nil)
	entries := make([]hpack.HeaderField, 0, 61)
	for index := byte(1); index <= 61; index++ {
		decoded, err := decoder.DecodeFull([]byte{0x80 | index})
		if err == nil && len(decoded) == 1 {
			entries = append(entries, decoded[0])
		}
	}
	return entries
})

var staticHPACKFirstNames = sync.OnceValue(func() map[string]uint64 {
	names := make(map[string]uint64)
	for i, entry := range staticHPACKEntries() {
		if names[entry.Name] == 0 {
			names[entry.Name] = uint64(i + 1)
		}
	}
	return names
})

func (t *indexedFields) nameReference(field hpack.HeaderField) uint64 {
	var reference uint64
	for i, entry := range staticHPACKEntries() {
		if entry.Name == field.Name {
			reference = uint64(i + 1)
			if entry.Value == field.Value {
				return reference
			}
		}
	}
	for i := len(t.entries) - 1; i >= 0; i-- {
		entry := t.entries[i]
		if entry.Name == field.Name {
			reference = uint64(62 + len(t.entries) - 1 - i)
			if entry.Value == field.Value {
				return reference
			}
		}
	}
	return reference
}
