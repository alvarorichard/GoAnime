package http2

import (
	"context"
	"fmt"
	"maps"
	"net/url"
	"strings"

	"github.com/enetx/http/internal/httpcommon"
	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/http2/hpack"
)

// HPACKRepresentation selects a field's representation. Default allows the
// upstream encoder to reuse an exact table entry or insert a new entry that fits.
type HPACKRepresentation uint8

// HPACKHuffmanMode selects automatic, always or disabled Huffman literal strings.
type HPACKHuffmanMode uint8

const (
	HPACKHuffmanDefault HPACKHuffmanMode = iota
	HPACKHuffmanAlways
	HPACKHuffmanNever
)

// HPACKNameReferencePolicy selects the name entry used by literal fields.
type HPACKNameReferencePolicy uint8

const (
	HPACKNameReferenceDefault HPACKNameReferencePolicy = iota
	// HPACKNameReferenceLast scans static entries, then dynamic entries from
	// newest to oldest. An exact match wins; otherwise the last matching name wins.
	HPACKNameReferenceLast
	// HPACKNameReferenceFirst uses the first matching static name; otherwise
	// it keeps the upstream encoder's most recent dynamic name reference.
	HPACKNameReferenceFirst
)

const (
	HPACKDefault HPACKRepresentation = iota
	// HPACKWithoutIndexing always emits a literal without inserting it.
	HPACKWithoutIndexing
	// HPACKNeverIndex always emits a literal with the never-index instruction.
	HPACKNeverIndex
	// HPACKIncrementalIndexing reuses exact matches, otherwise emits an
	// incremental literal even when the entry exceeds the table capacity.
	// An oversized insertion empties the table, as required by RFC 7541 4.4.
	HPACKIncrementalIndexing
)

// HPACKConfig controls request and trailer encoding without replacing x/net's
// encoder, dynamic table, size updates or Huffman implementation. Callbacks must
// not reenter the transport. Representations override Policy, then Default.
type HPACKConfig struct {
	// DeduplicateTableSizeUpdates ignores SETTINGS that repeat the current
	// encoder capacity, preserving any already pending minimum/maximum updates.
	DeduplicateTableSizeUpdates bool
	Huffman                     HPACKHuffmanMode
	NameReferencePolicy         HPACKNameReferencePolicy
	Default                     HPACKRepresentation
	Representations             map[string]HPACKRepresentation
	Policy                      func(name, value string) HPACKRepresentation
	// IndexingLimit bounds NEW dynamic entries by their HPACK size, including
	// the 32-byte overhead. Existing exact entries can still be referenced.
	// Nil retains the upstream capacity bound. Zero returned by the callback
	// disables new entries. Called under the header lock; must not reenter.
	IndexingLimit func(tableCapacity uint32) uint32
}

func (cc *ClientConn) setFingerprintEncoderTableSize(size uint32) {
	if fp := cc.fingerprint; fp != nil && fp.HPACK != nil && fp.HPACK.DeduplicateTableSizeUpdates && size == cc.henc.MaxDynamicTableSize() {
		return
	}
	cc.henc.SetMaxDynamicTableSize(size)
	if cc.fingerprintTable != nil {
		cc.fingerprintTable.resize(cc.henc.MaxDynamicTableSize())
	}
}

func (c *HPACKConfig) Clone() *HPACKConfig {
	if c == nil {
		return nil
	}
	copy := *c
	copy.Representations = maps.Clone(c.Representations)
	return &copy
}

func (c *HPACKConfig) Validate() error {
	if c == nil {
		return nil
	}
	if c.Default > HPACKIncrementalIndexing {
		return fmt.Errorf("http2: invalid HPACK default representation %d", c.Default)
	}
	if c.Huffman > HPACKHuffmanNever || c.NameReferencePolicy > HPACKNameReferenceFirst {
		return fmt.Errorf("http2: invalid HPACK encoding policy")
	}
	for name, rep := range c.Representations {
		lower, ascii := httpcommon.LowerHeader(name)
		if !ascii || name != lower || !httpguts.ValidHeaderFieldName(strings.TrimPrefix(name, ":")) || rep > HPACKIncrementalIndexing {
			return fmt.Errorf("http2: invalid HPACK representation for %q", name)
		}
	}
	return nil
}

func (cc *ClientConn) writeFingerprintHeader(name, value string) error {
	rep := HPACKDefault
	var config *HPACKConfig
	if fp := cc.fingerprint; fp != nil && fp.HPACK != nil {
		config = fp.HPACK
		rep = config.Default
		if config.Policy != nil {
			rep = config.Policy(name, value)
		}
		if override, ok := config.Representations[name]; ok {
			rep = override
		}
	}
	if rep > HPACKIncrementalIndexing {
		return fmt.Errorf("http2: invalid HPACK representation %d for %q", rep, name)
	}
	field := hpack.HeaderField{Name: name, Value: value}
	var nameReference uint64
	if config != nil && (config.IndexingLimit != nil || config.NameReferencePolicy == HPACKNameReferenceLast) {
		if cc.fingerprintTable == nil {
			cc.fingerprintTable = newIndexedFields(cc.henc.MaxDynamicTableSize())
		}
		if config.IndexingLimit != nil && (rep == HPACKDefault || rep == HPACKIncrementalIndexing) && field.Size() > config.IndexingLimit(cc.henc.MaxDynamicTableSize()) && !cc.fingerprintTable.contains(field) && !staticHPACKFields()[field] {
			rep = HPACKWithoutIndexing
		}
		if config.NameReferencePolicy == HPACKNameReferenceLast {
			nameReference = cc.fingerprintTable.nameReference(field)
		}
	}
	start := cc.hbuf.Len()
	field.Sensitive = rep == HPACKWithoutIndexing || rep == HPACKNeverIndex
	// The official encoder only returns writer/short-write errors. Its writer
	// is this connection's bytes.Buffer, which always accepts the full field.
	cc.henc.WriteField(field)

	return cc.finishFingerprintHeader(start, field, rep, config, nameReference)
}

func skipHPACKSizeUpdates(data []byte) []byte {
	for len(data) > 0 && data[0]&0xe0 == 0x20 {
		first := data[0]
		data = data[1:]
		if first&31 == 31 {
			for len(data) > 0 {
				b := data[0]
				data = data[1:]
				if b&128 == 0 {
					break
				}
			}
		}
	}
	return data
}

// PriorityRequest is a read-only view of the request being assigned a stream.
// The callback runs under the connection's header write lock; it must not reenter
// the transport or mutate Header or URL. Different connections may call it concurrently.
type PriorityRequest struct {
	Context context.Context
	Method  string
	URL     *url.URL
	Header  map[string][]string
}

func (cc *ClientConn) requestFingerprintPriority(req *ClientRequest, cs *clientStream) (PriorityParam, error) {
	priority := cc.fingerprintPriority()
	fp := cc.fingerprint
	if fp != nil && fp.HeaderPriority != nil {
		priority = fp.HeaderPriority(PriorityRequest{req.Context, req.Method, req.URL, req.Header})
	}
	if priority.StreamDep > 0x7fffffff || priority.StreamDep == cs.ID {
		return PriorityParam{}, fmt.Errorf("http2: invalid stream dependency %d for stream %d", priority.StreamDep, cs.ID)
	}
	cc.mu.Lock()
	defer cc.mu.Unlock()
	if fp != nil && fp.ChainStreamDependencies && !priority.IsZero() {
		priority.StreamDep = 0
		for id, active := range cc.streams {
			if id < cs.ID && id > priority.StreamDep && active.headerPrioritySet && active.headerPriority.Weight >= priority.Weight {
				priority.StreamDep = id
			}
		}
	}
	cs.headerPriority, cs.headerPrioritySet = priority, true
	return priority, nil
}
