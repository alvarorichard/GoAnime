package qpack

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	base "github.com/quic-go/qpack"
)

type WaitFunc func(context.Context, <-chan struct{}) error

// Decoder belongs to one connection. It accepts encoder instructions serially;
// field sections on different streams may decode concurrently.
type Decoder struct {
	mu                                                   sync.Mutex
	maxCapacity, capacity, size, insertCount, maxBlocked uint64
	table                                                []base.HeaderField // oldest first, absolute base = insertCount-len(table)
	blocked                                              map[uint64]int
	changed                                              chan struct{}
	err                                                  error
	// emit is nonblocking and preserves call order. The HTTP adapter bounds its
	// queue and terminates an overloaded connection instead of retaining data.
	emit          func([]byte) error
	knownReceived uint64
}

func NewDecoder(capacity, blocked uint64, emit func([]byte) error) *Decoder {
	return &Decoder{maxCapacity: capacity, maxBlocked: blocked, changed: make(chan struct{}), blocked: make(map[uint64]int), emit: emit}
}
func (d *Decoder) notify() { close(d.changed); d.changed = make(chan struct{}) }
func (d *Decoder) Close(err error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.err == nil {
		d.err = err
		d.notify()
	}
}
func (d *Decoder) feedback(prefix uint8, flags byte, value uint64) error {
	if d.emit == nil {
		return nil
	}
	if err := d.emit(appendInt(nil, prefix, flags, value)); err != nil {
		d.err = err
		d.notify()
		return err
	}
	return nil
}
func (d *Decoder) at(index uint64) (base.HeaderField, error) {
	first := d.insertCount - uint64(len(d.table))
	if index < first || index >= d.insertCount {
		return base.HeaderField{}, ErrDecompression
	}
	return d.table[index-first], nil
}
func (d *Decoder) relative(baseIndex, index uint64, post bool) (base.HeaderField, uint64, error) {
	var absolute uint64
	if post {
		if index > maxInt-baseIndex {
			return base.HeaderField{}, 0, ErrDecompression
		}
		absolute = baseIndex + index
	} else {
		if index >= baseIndex {
			return base.HeaderField{}, 0, ErrDecompression
		}
		absolute = baseIndex - index - 1
	}
	field, err := d.at(absolute)
	return field, absolute + 1, err
}
func (d *Decoder) evict(target uint64) {
	count := 0
	for d.size > target && count < len(d.table) {
		field := d.table[count]
		d.size -= uint64(len(field.Name) + len(field.Value) + 32)
		count++
	}
	if count > 0 {
		copy(d.table, d.table[count:])
		clear(d.table[len(d.table)-count:])
		d.table = d.table[:len(d.table)-count]
	}
}
func (d *Decoder) insert(field base.HeaderField) error {
	size := uint64(len(field.Name) + len(field.Value) + 32)
	if size > d.capacity || d.insertCount == maxInt {
		return ErrEncoderStream
	}
	d.evict(d.capacity - size)
	d.table = append(d.table, field)
	d.size += size
	d.insertCount++
	d.notify()
	// Eager feedback allows speculative insertions to make progress without
	// requiring a field section to reference them first.
	increment := d.insertCount - d.knownReceived
	d.knownReceived = d.insertCount
	return d.feedback(6, 0, increment)
}

// ReadEncoder processes complete or arbitrarily fragmented encoder instructions.
// EOF/reset on this critical stream is classified by the HTTP/3 adapter.
func (d *Decoder) ReadEncoder(input io.Reader) error {
	reader := bufio.NewReader(input)
	for {
		first, err := reader.ReadByte()
		if err != nil {
			return err
		}
		var field base.HeaderField
		var index uint64
		switch {
		case first&0x80 != 0:
			index, err = readInt(reader, first, 6)
			if err == nil {
				d.mu.Lock()
				if first&0x40 != 0 {
					field, err = staticField(index)
				} else if index >= d.insertCount {
					err = ErrEncoderStream
				} else {
					field, err = d.at(d.insertCount - index - 1)
				}
				d.mu.Unlock()
			}
			if err == nil {
				var b byte
				b, err = reader.ReadByte()
				if err == nil {
					field.Value, err = readString(reader, b, 7, b&0x80 != 0, d.maxCapacity)
				}
			}
		case first&0xc0 == 0x40:
			field.Name, err = readString(reader, first, 5, first&0x20 != 0, d.maxCapacity)
			if err == nil {
				var b byte
				b, err = reader.ReadByte()
				if err == nil {
					field.Value, err = readString(reader, b, 7, b&0x80 != 0, d.maxCapacity)
				}
			}
		case first&0xe0 == 0x20:
			index, err = readInt(reader, first, 5)
			if err == nil {
				d.mu.Lock()
				if index > d.maxCapacity {
					err = ErrEncoderStream
				} else {
					d.capacity = index
					d.evict(index)
				}
				d.mu.Unlock()
			}
			if err == nil {
				continue
			}
		default:
			index, err = readInt(reader, first, 5)
			if err == nil {
				d.mu.Lock()
				if index >= d.insertCount {
					err = ErrEncoderStream
				} else {
					field, err = d.at(d.insertCount - index - 1)
				}
				d.mu.Unlock()
			}
		}
		if err != nil {
			return fmt.Errorf("%w: %w", ErrEncoderStream, err)
		}
		d.mu.Lock()
		if d.err != nil {
			err = d.err
		} else {
			err = d.insert(field)
		}
		d.mu.Unlock()
		if err != nil {
			return err
		}
	}
}
func (d *Decoder) required(encoded uint64) (uint64, error) {
	if encoded == 0 {
		return 0, nil
	}
	entries := d.maxCapacity / 32
	full := 2 * entries
	if full == 0 || encoded > full || d.insertCount > maxInt-entries {
		return 0, ErrDecompression
	}
	maxValue := d.insertCount + entries
	required := maxValue/full*full + encoded - 1
	if required > maxValue {
		if required <= full {
			return 0, ErrDecompression
		}
		required -= full
	}
	if required == 0 {
		return 0, ErrDecompression
	}
	return required, nil
}

// Decode fully validates one field section under the table lock so concurrent
// eviction cannot change its references halfway through decoding. Output size is
// bounded independently from compressed input. wait preserves stream cancellation
// and read-deadline semantics while another QUIC stream supplies table entries.
func (d *Decoder) Decode(ctx context.Context, streamID uint64, block []byte, limit uint64, wait WaitFunc) ([]base.HeaderField, error) {
	if ctx.Err() != nil {
		return nil, context.Cause(ctx)
	}
	reader := bytes.NewReader(block)
	first, err := reader.ReadByte()
	if err != nil {
		return nil, ErrDecompression
	}
	encoded, err := readInt(reader, first, 8)
	if err != nil {
		return nil, ErrDecompression
	}
	first, err = reader.ReadByte()
	if err != nil {
		return nil, ErrDecompression
	}
	delta, err := readInt(reader, first, 7)
	if err != nil {
		return nil, ErrDecompression
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	required, err := d.required(encoded)
	if err != nil {
		return nil, err
	}
	var baseIndex uint64
	if first&0x80 != 0 {
		if delta >= required {
			return nil, ErrDecompression
		}
		baseIndex = required - delta - 1
	} else {
		if delta > maxInt-required {
			return nil, ErrDecompression
		}
		baseIndex = required + delta
	}
	if required > d.insertCount {
		if _, ok := d.blocked[streamID]; !ok && uint64(len(d.blocked)) >= d.maxBlocked {
			return nil, ErrDecompression
		}
		d.blocked[streamID]++
		defer func() {
			d.blocked[streamID]--
			if d.blocked[streamID] == 0 {
				delete(d.blocked, streamID)
			}
		}()
		for required > d.insertCount && d.err == nil {
			changed := d.changed
			d.mu.Unlock()
			if wait != nil {
				err = wait(ctx, changed)
			} else {
				select {
				case <-changed:
				case <-ctx.Done():
					err = context.Cause(ctx)
				}
			}
			d.mu.Lock()
			if err != nil {
				return nil, err
			}
		}
	}
	if d.err != nil {
		return nil, d.err
	}
	var fields []base.HeaderField
	var largest uint64
	for reader.Len() > 0 {
		first, _ := reader.ReadByte()
		var field base.HeaderField
		var reference, index uint64
		switch {
		case first&0x80 != 0:
			index, err = readInt(reader, first, 6)
			if err == nil {
				if first&0x40 != 0 {
					field, err = staticField(index)
				} else {
					field, reference, err = d.relative(baseIndex, index, false)
				}
			}
		case first&0xc0 == 0x40:
			index, err = readInt(reader, first, 4)
			if err == nil {
				if first&0x10 != 0 {
					field, err = staticField(index)
				} else {
					field, reference, err = d.relative(baseIndex, index, false)
				}
			}
			if err == nil {
				var b byte
				b, err = reader.ReadByte()
				if err == nil {
					field.Value, err = readString(reader, b, 7, b&0x80 != 0, limit)
				}
			}
		case first&0xe0 == 0x20:
			field.Name, err = readString(reader, first, 3, first&8 != 0, limit)
			if err == nil {
				var b byte
				b, err = reader.ReadByte()
				if err == nil {
					field.Value, err = readString(reader, b, 7, b&0x80 != 0, limit)
				}
			}
		case first&0xf0 == 0x10:
			index, err = readInt(reader, first, 4)
			if err == nil {
				field, reference, err = d.relative(baseIndex, index, true)
			}
		default:
			index, err = readInt(reader, first, 3)
			if err == nil {
				field, reference, err = d.relative(baseIndex, index, true)
			}
			if err == nil {
				var b byte
				b, err = reader.ReadByte()
				if err == nil {
					field.Value, err = readString(reader, b, 7, b&0x80 != 0, limit)
				}
			}
		}
		if err != nil {
			if errors.Is(err, ErrFieldSectionTooLarge) {
				return nil, err
			}
			return nil, fmt.Errorf("%w: %v", ErrDecompression, err)
		}
		if reference > required {
			return nil, ErrDecompression
		}
		largest = max(largest, reference)
		size := uint64(len(field.Name) + len(field.Value) + 32)
		if size > limit {
			return nil, ErrFieldSectionTooLarge
		}
		limit -= size
		fields = append(fields, field)
	}
	if largest != required {
		return nil, ErrDecompression
	}
	if required > 0 {
		d.knownReceived = max(d.knownReceived, required)
		if err := d.feedback(7, 0x80, streamID); err != nil {
			return nil, err
		}
	}
	return fields, nil
}

// Cancel releases references when the protocol layer abandons a stream.
func (d *Decoder) Cancel(streamID uint64) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.err != nil {
		return d.err
	}
	return d.feedback(6, 0x40, streamID)
}
