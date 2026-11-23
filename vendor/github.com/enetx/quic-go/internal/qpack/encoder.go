package qpack

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"math"
	"slices"
	"strings"
	"sync"

	base "github.com/quic-go/qpack"
	"golang.org/x/net/http2/hpack"
)

// HuffmanMode selects string encoding without replacing the official Huffman codec.
type HuffmanMode uint8

const (
	HuffmanDefault HuffmanMode = iota // Preserve the official static encoder's always-Huffman encoding.
	HuffmanAuto                       // Huffman only when strictly shorter.
	HuffmanAlways
	HuffmanNever
)

// IndexingMode controls new entries; existing entries can still be referenced.
type IndexingMode uint8

const (
	IndexAll IndexingMode = iota
	IndexNewNames
	IndexNone
)

type EncoderOptions struct {
	DynamicTable           bool
	Huffman                HuffmanMode
	Indexing               IndexingMode
	IndexingPolicy         func(name, value string) bool
	DrainingFraction       float64
	PostBaseReferences     bool
	LiteralInsertNames     bool
	StopOnInsertionFailure bool
	MaxTrackedSections     uint64 // Zero selects the bounded default of 4096 outstanding sections.
}

func (o EncoderOptions) Validate() error {
	if o.Huffman > HuffmanNever || o.Indexing > IndexNone || math.IsNaN(o.DrainingFraction) || o.DrainingFraction < 0 || o.DrainingFraction > 1 || o.MaxTrackedSections > maxInt {
		return errors.New("qpack: invalid encoder options")
	}
	return nil
}

type Field struct {
	Name, Value string
	NeverIndex  bool
}

type encoderEntry struct {
	field      Field
	index      uint64
	references uint64
}

type encoderSection struct {
	required uint64
	indices  []uint64
}

type encoderState struct {
	table                 []encoderEntry // Oldest first; absolute indices never change.
	size, inserted, known uint64
	sections              map[uint64][]encoderSection // FIFO, including multiple sections on one stream.
	tracked               uint64
}

// Encoder belongs to one connection. Encoding and decoder feedback are serialized.
// Outstanding references and unacknowledged insertions cannot be evicted (RFC
// 9204 2.1.1). A full tracking budget falls back to static encoding, without waiting.
type Encoder struct {
	mu                         sync.Mutex
	options                    EncoderOptions
	state                      encoderState
	maximum, capacity, blocked uint64
	configured                 bool
}

func NewEncoder(options EncoderOptions) *Encoder {
	if options.MaxTrackedSections == 0 {
		options.MaxTrackedSections = 4096
	}
	return &Encoder{options: options, state: encoderState{sections: make(map[uint64][]encoderSection)}}
}

// Configure separates the peer SETTINGS maximum (which determines RIC modulo)
// from the encoder's chosen capacity. The HTTP adapter serializes the capacity
// instruction before allowing insertions. SETTINGS is immutable per connection.
func (e *Encoder) Configure(maximum, capacity, blocked uint64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if maximum > maxInt || capacity > maximum || blocked > maxInt {
		return errors.New("qpack: invalid encoder limits")
	}
	if e.configured && (maximum != e.maximum || capacity != e.capacity || blocked != e.blocked) {
		return errors.New("qpack: encoder limits already configured")
	}
	e.maximum, e.capacity, e.blocked, e.configured = maximum, capacity, blocked, true
	return nil
}

// UpdatePeerLimits applies compatible peer limits after remembered 0-RTT
// settings. A nonzero maximum fixes Required Insert Count modulo and cannot
// change. Zero can grow while the table is empty; capacity and blocked-stream
// allowances may increase. Existing entries and acknowledgments are retained.
func (e *Encoder) UpdatePeerLimits(maximum, capacity, blocked uint64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if maximum > maxInt || capacity > maximum || blocked > maxInt ||
		e.configured && (e.maximum != 0 && maximum != e.maximum || capacity < e.capacity || blocked < e.blocked) {
		return errors.New("qpack: incompatible remembered encoder limits")
	}
	e.maximum, e.capacity, e.blocked, e.configured = maximum, capacity, blocked, true
	return nil
}

// ErrEncoderBlocked asks Encode to keep its existing state and emit a static
// section. emit must either accept the entire instruction batch in order or
// reject it without sending any bytes. It must not reenter this Encoder.
var ErrEncoderBlocked = errors.New("qpack: encoder stream backpressure")

// Encode produces one complete field section. State changes are committed only
// after emit accepts all instructions. A rejected batch cannot leave references
// to insertions that were never sent. Native static encoding needs no emit callback.
func (e *Encoder) Encode(streamID uint64, fields []Field, emit func([]byte) error) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	if streamID > maxInt {
		return nil, errors.New("qpack: invalid stream ID")
	}
	if err := e.options.Validate(); err != nil {
		return nil, err
	}
	if !e.options.DynamicTable || e.capacity < 32 || e.state.tracked >= e.options.MaxTrackedSections {
		return encodeStatic(fields, e.options.Huffman), nil
	}
	// Only table storage is copied. The section map is read-only until commit.
	working := e.state
	working.table = slices.Clone(e.state.table)
	tx := encoderTransaction{state: &working, encoder: e, refs: make(map[uint64]bool)}
	tx.canBlock = e.canBlock(streamID)
	tx.draining = tx.drainingIndex()
	initialBase := working.inserted
	reps := make([]fieldRepresentation, 0, len(fields))
	for _, field := range fields {
		reps = append(reps, tx.encodeField(field))
	}
	baseIndex := tx.required
	if e.options.PostBaseReferences {
		baseIndex = initialBase
	}
	block := appendSectionPrefix(nil, tx.required, baseIndex, e.maximum/32)
	for _, rep := range reps {
		block = appendRepresentation(block, rep, baseIndex, e.options.Huffman)
	}
	if len(tx.instructions) > 0 {
		if emit == nil {
			return encodeStatic(fields, e.options.Huffman), nil
		}
		if err := emit(tx.instructions); err != nil {
			if errors.Is(err, ErrEncoderBlocked) {
				return encodeStatic(fields, e.options.Huffman), nil
			}
			return nil, err
		}
	}
	if tx.required > 0 {
		indices := make([]uint64, 0, len(tx.refs))
		for index := range tx.refs {
			indices = append(indices, index)
		}
		slices.Sort(indices)
		working.sections[streamID] = append(working.sections[streamID], encoderSection{tx.required, indices})
		working.tracked++
	}
	e.state = working
	return block, nil
}

func (e *Encoder) canBlock(id uint64) bool {
	var count uint64
	for stream, sections := range e.state.sections {
		for _, section := range sections {
			if section.required > e.state.known {
				if stream == id {
					return true
				}
				count++
				break
			}
		}
	}
	return count < e.blocked
}

type encoderTransaction struct {
	state            *encoderState
	encoder          *Encoder
	refs             map[uint64]bool
	required         uint64
	draining         uint64
	canBlock         bool
	instructions     []byte
	insertionBlocked bool
}

func (t *encoderTransaction) reference(index uint64) {
	if !t.refs[index] {
		t.refs[index] = true
		for i := range t.state.table {
			if t.state.table[i].index == index {
				t.state.table[i].references++
				break
			}
		}
	}
	t.required = max(t.required, index+1)
}

func (t *encoderTransaction) usable(index uint64) bool {
	return index >= t.draining && (t.canBlock || index < t.state.known)
}

func (t *encoderTransaction) drainingIndex() uint64 {
	space := t.encoder.capacity - t.state.size
	wanted := uint64(float64(t.encoder.capacity) * t.encoder.options.DrainingFraction)
	for _, entry := range t.state.table {
		if space >= wanted {
			return entry.index
		}
		space += fieldSize(entry.field)
	}
	return t.state.inserted
}

// insert first checks all eviction constraints, including references acquired
// earlier in this same section. Encoder-instruction name references are protected
// until their instruction has been constructed. No partial eviction is possible.
func (t *encoderTransaction) insert(field Field, name match, duplicate bool) (uint64, bool) {
	size := fieldSize(field)
	if size > t.encoder.capacity || t.state.inserted == maxInt {
		return 0, false
	}
	remaining, drop := t.state.size, 0
	for remaining > t.encoder.capacity-size {
		if drop >= len(t.state.table) {
			return 0, false
		}
		entry := t.state.table[drop]
		if entry.references > 0 || entry.index >= t.state.known || (name.found && !name.static && entry.index >= name.index) {
			return 0, false
		}
		remaining -= fieldSize(entry.field)
		drop++
	}
	if duplicate {
		t.instructions = appendInt(t.instructions, 5, 0, t.state.inserted-name.index-1)
	} else if name.found && !t.encoder.options.LiteralInsertNames {
		index, flags := name.index, byte(0x80)
		if name.static {
			flags |= 0x40
		} else {
			index = t.state.inserted - index - 1
		}
		t.instructions = appendInt(t.instructions, 6, flags, index)
		t.instructions = appendString(t.instructions, field.Value, 7, 0, 0x80, t.encoder.options.Huffman)
	} else {
		t.instructions = appendString(t.instructions, field.Name, 5, 0x40, 0x20, t.encoder.options.Huffman)
		t.instructions = appendString(t.instructions, field.Value, 7, 0, 0x80, t.encoder.options.Huffman)
	}
	if drop > 0 {
		copy(t.state.table, t.state.table[drop:])
		clear(t.state.table[len(t.state.table)-drop:])
		t.state.table = t.state.table[:len(t.state.table)-drop]
	}
	index := t.state.inserted
	stored := Field{Name: strings.Clone(field.Name), Value: strings.Clone(field.Value)}
	t.state.table = append(t.state.table, encoderEntry{field: stored, index: index})
	t.state.size = remaining + size
	t.state.inserted++
	return index, true
}

func (t *encoderTransaction) tryInsert(field Field, name match, duplicate bool) (uint64, bool) {
	if t.insertionBlocked {
		return 0, false
	}
	index, ok := t.insert(field, name, duplicate)
	if !ok && t.encoder.options.StopOnInsertionFailure {
		t.insertionBlocked = true
	}
	return index, ok
}

type match struct {
	index                uint64
	found, static, exact bool
}

func (t *encoderTransaction) lookup(field Field) match {
	static := lookupStatic(field)
	if static.exact {
		return static
	}
	name := static
	for i := len(t.state.table) - 1; i >= 0; i-- {
		entry := t.state.table[i]
		if entry.field.Name != field.Name {
			continue
		}
		// New-name indexing uses a usable name match instead of inserting a new
		// value. Filter unacknowledged entries when no blocking budget remains.
		if t.encoder.options.Indexing == IndexNewNames && !t.canBlock && entry.index >= t.state.known {
			continue
		}
		if entry.field.Value == field.Value {
			return match{index: entry.index, found: true, exact: true}
		}
		if !name.found {
			name = match{index: entry.index, found: true}
		}
	}
	return name
}

func (t *encoderTransaction) encodeField(field Field) fieldRepresentation {
	m := t.lookup(field)
	if field.NeverIndex {
		return literalRepresentation(field, lookupStatic(field))
	}
	if m.found && m.static && m.exact {
		return fieldRepresentation{field: field, indexed: true, name: m}
	}
	indexable := t.encoder.options.Indexing != IndexNone && (t.encoder.options.IndexingPolicy == nil || t.encoder.options.IndexingPolicy(field.Name, field.Value))
	if m.found && !m.static && m.exact {
		if t.usable(m.index) {
			t.reference(m.index)
			return fieldRepresentation{field: field, indexed: true, name: m}
		}
		if m.index < t.draining && t.canBlock && indexable {
			if index, ok := t.tryInsert(field, m, true); ok {
				t.reference(index)
				return fieldRepresentation{field: field, indexed: true, name: match{index: index, found: true}}
			}
		}
		// A failed exact match cannot be reused as a name. Prefer a static name;
		// otherwise use a literal, allowing the old entry to drain.
		m = lookupStaticName(field.Name)
		if !m.found {
			return literalRepresentation(field, match{})
		}
	}
	if t.canBlock && indexable && (t.encoder.options.Indexing != IndexNewNames || !m.found) {
		if index, ok := t.tryInsert(field, m, false); ok {
			t.reference(index)
			return fieldRepresentation{field: field, indexed: true, name: match{index: index, found: true}}
		}
	}
	if m.found && !m.static {
		if !t.usable(m.index) {
			m = lookupStaticName(field.Name)
		} else {
			t.reference(m.index)
		}
	}
	return literalRepresentation(field, m)
}

func fieldSize(field Field) uint64 { return uint64(len(field.Name)) + uint64(len(field.Value)) + 32 }

type fieldRepresentation struct {
	field   Field
	name    match
	indexed bool
}

func literalRepresentation(field Field, name match) fieldRepresentation {
	return fieldRepresentation{field: field, name: name}
}

var encoderStaticTable = sync.OnceValue(func() []base.HeaderField {
	fields := make([]base.HeaderField, 99) // RFC 9204 Appendix A; values come from the official codec.
	for i := range fields {
		var err error
		fields[i], err = staticField(uint64(i))
		if err != nil {
			panic(fmt.Sprintf("qpack: official static table index %d: %v", i, err))
		}
	}
	return fields
})

func lookupStatic(field Field) match {
	var name match
	for i, entry := range encoderStaticTable() {
		if entry.Name != field.Name {
			continue
		}
		m := match{index: uint64(i), found: true, static: true, exact: entry.Value == field.Value}
		if m.exact {
			return m
		}
		if !name.found {
			name = m
		}
	}
	return name
}

func lookupStaticName(name string) match {
	m := lookupStatic(Field{Name: name})
	m.exact = false
	return m
}

func appendString(dst []byte, value string, prefix uint8, flags, huffmanFlag byte, mode HuffmanMode) []byte {
	length := hpack.HuffmanEncodeLength(value)
	if mode != HuffmanNever && (mode != HuffmanAuto || length < uint64(len(value))) {
		dst = appendInt(dst, prefix, flags|huffmanFlag, length)
		return hpack.AppendHuffmanString(dst, value)
	}
	dst = appendInt(dst, prefix, flags, uint64(len(value)))
	return append(dst, value...)
}

func appendSectionPrefix(dst []byte, required, baseIndex, maxEntries uint64) []byte {
	encoded := uint64(0)
	if required > 0 {
		encoded = required%(2*maxEntries) + 1
	}
	dst = appendInt(dst, 8, 0, encoded)
	if baseIndex < required {
		return appendInt(dst, 7, 0x80, required-baseIndex-1)
	}
	return appendInt(dst, 7, 0, baseIndex-required)
}

func appendRepresentation(dst []byte, rep fieldRepresentation, baseIndex uint64, mode HuffmanMode) []byte {
	m, field := rep.name, rep.field
	if rep.indexed {
		if m.static {
			return appendInt(dst, 6, 0xc0, m.index)
		}
		if m.index < baseIndex {
			return appendInt(dst, 6, 0x80, baseIndex-m.index-1)
		}
		return appendInt(dst, 4, 0x10, m.index-baseIndex)
	}
	never := byte(0)
	if field.NeverIndex {
		never = 0x20
	}
	if m.found {
		if m.static {
			dst = appendInt(dst, 4, 0x50|never, m.index)
		} else if m.index < baseIndex {
			dst = appendInt(dst, 4, 0x40|never, baseIndex-m.index-1)
		} else {
			dst = appendInt(dst, 3, never>>2, m.index-baseIndex)
		}
	} else {
		dst = appendString(dst, field.Name, 3, 0x20|never>>1, 0x08, mode)
	}
	return appendString(dst, field.Value, 7, 0, 0x80, mode)
}

func encodeStatic(fields []Field, mode HuffmanMode) []byte {
	out := []byte{0, 0}
	for _, field := range fields {
		m := lookupStatic(field)
		out = appendRepresentation(out, fieldRepresentation{field: field, name: m, indexed: m.exact && !field.NeverIndex}, 0, mode)
	}
	return out
}

// ReadDecoder accepts arbitrarily fragmented feedback. Section ACKs release one
// FIFO section; cancellation releases all sections without advancing Known Received
// Count. Invalid acknowledgments and increments are connection errors (RFC 9204 4.4).
func (e *Encoder) ReadDecoder(input io.Reader) error {
	r := bufio.NewReader(input)
	for {
		first, err := r.ReadByte()
		if err != nil {
			return err
		}
		prefix := uint8(6)
		if first&0x80 != 0 {
			prefix = 7
		}
		value, err := readInt(r, first, prefix)
		if err != nil {
			return errors.Join(ErrDecoderStream, err)
		}
		if err := e.feedback(first, value); err != nil {
			return err
		}
	}
}

func (e *Encoder) feedback(first byte, value uint64) error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if first&0xc0 == 0 {
		if value == 0 || value > e.state.inserted-e.state.known {
			return ErrDecoderStream
		}
		e.state.known += value
		return nil
	}
	sections := e.state.sections[value]
	if first&0x80 != 0 {
		if len(sections) == 0 {
			return ErrDecoderStream
		}
		e.state.known = max(e.state.known, sections[0].required)
		e.release(sections[0])
		sections[0] = encoderSection{}
		sections = sections[1:]
	} else {
		for _, section := range sections {
			e.release(section)
		}
		sections = nil
	}
	if len(sections) == 0 {
		delete(e.state.sections, value)
	} else {
		e.state.sections[value] = sections
	}
	return nil
}

func (e *Encoder) release(section encoderSection) {
	for _, index := range section.indices {
		for i := range e.state.table {
			if e.state.table[i].index == index {
				e.state.table[i].references--
				break
			}
		}
	}
	e.state.tracked--
}
