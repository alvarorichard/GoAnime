package miruro

import (
	"bytes"
	"errors"
	"fmt"
	"math"

	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
)

// miruro.cx is a SvelteKit app, and its __data.json endpoints answer in the
// format SvelteKit serialises load data with (github.com/sveltejs/devalue):
// every value is stored once in a flat JSON array and refers to the others by
// index, data[0] being the root. {"a":1} means "key a holds whatever is at
// index 1". A negative index is a constant (undefined, NaN, ...), and an array
// whose first element is a string is a typed value: ["Date", "..."],
// ["Map", k, v, ...], ["Promise", i].
//
// The watch page streams: its first line holds the nodes, and the slow parts
// of the page (the episode list, the sources) are ["Promise", i] values that
// later lines resolve, one {"type":"chunk","id":N,"data":[...]} per promise,
// where data[i] of the referring array is N.
//
// hydrate turns one flat array back into plain Go values (maps, slices,
// strings, float64, bool, nil), which decodeInto then reads into a struct by
// way of JSON. Unknown typed values become nil rather than an error: the
// fields this scraper reads are all plain, and a new type elsewhere on the page
// must not take search down.

// maxHydrateDepth bounds how deep a document may nest. The pages nest about a
// dozen levels; the limit is there so a hostile or broken payload fails
// instead of recursing without end.
const maxHydrateDepth = 128

var (
	errDevalueCycle = errors.New("devalue: value refers to itself")
	errDevalueDepth = errors.New("devalue: nested too deep")
)

// Constants a reference can be instead of an index.
const (
	devalueUndefined   = -1
	devalueHole        = -2
	devalueNaN         = -3
	devaluePosInfinity = -4
	devalueNegInfinity = -5
	devalueNegZero     = -6
)

// hydrator rebuilds one flat array. resolve is how a ["Promise", i] value
// asks for the chunk that settled it; nil means no chunks (search pages).
type hydrator struct {
	flat    []any
	resolve func(id int) (any, error)
	state   []uint8 // 0 unseen, 1 being built (a cycle if met again), 2 built
	memo    []any
}

func newHydrator(flat []any, resolve func(int) (any, error)) *hydrator {
	return &hydrator{
		flat:    flat,
		resolve: resolve,
		state:   make([]uint8, len(flat)),
		memo:    make([]any, len(flat)),
	}
}

// hydrateFlat rebuilds a node's or a chunk's data. devalue writes a bare
// constant (a lone -1 for undefined) instead of an array when the whole value
// is one, so anything that is not an array is read that way.
func hydrateFlat(data any, resolve func(int) (any, error)) (any, error) {
	flat, ok := data.([]any)
	if !ok {
		idx, isNum := data.(float64)
		if !isNum {
			return nil, fmt.Errorf("devalue: data is %T, not an array", data)
		}
		return constant(int(idx))
	}
	if len(flat) == 0 {
		return nil, errors.New("devalue: empty data array")
	}
	return newHydrator(flat, resolve).value(0, 0)
}

// constant reads a negative reference.
func constant(idx int) (any, error) {
	switch idx {
	case devalueUndefined, devalueHole, devalueNaN, devaluePosInfinity, devalueNegInfinity:
		// None of these has a JSON form, and the fields read here never hold
		// them: absent is the honest translation.
		return nil, nil
	case devalueNegZero:
		return float64(0), nil
	default:
		return nil, fmt.Errorf("devalue: unknown constant %d", idx)
	}
}

// ref reads one reference out of an object or array: an index into flat.
func (h *hydrator) ref(raw any, depth int) (any, error) {
	f, ok := raw.(float64)
	if !ok || f != math.Trunc(f) {
		return nil, fmt.Errorf("devalue: reference %v is not an index", raw)
	}
	return h.value(int(f), depth)
}

func (h *hydrator) value(idx, depth int) (any, error) {
	if idx < 0 {
		return constant(idx)
	}
	if idx >= len(h.flat) {
		return nil, fmt.Errorf("devalue: index %d out of range (%d values)", idx, len(h.flat))
	}
	if depth > maxHydrateDepth {
		return nil, errDevalueDepth
	}
	switch h.state[idx] {
	case 1:
		return nil, errDevalueCycle
	case 2:
		return h.memo[idx], nil
	}
	h.state[idx] = 1
	v, err := h.build(h.flat[idx], depth)
	if err != nil {
		return nil, err
	}
	h.state[idx] = 2
	h.memo[idx] = v
	return v, nil
}

func (h *hydrator) build(raw any, depth int) (any, error) {
	switch v := raw.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, r := range v {
			child, err := h.ref(r, depth+1)
			if err != nil {
				return nil, err
			}
			out[k] = child
		}
		return out, nil
	case []any:
		if len(v) > 0 {
			if tag, ok := v[0].(string); ok {
				return h.typed(tag, v[1:], depth)
			}
		}
		out := make([]any, 0, len(v))
		for _, r := range v {
			child, err := h.ref(r, depth+1)
			if err != nil {
				return nil, err
			}
			out = append(out, child)
		}
		return out, nil
	default:
		// string, float64, bool or nil: a primitive stored in place.
		return v, nil
	}
}

// typed reads ["Tag", args...]. The tags handled are the ones devalue emits
// for values a page can plausibly hold; the rest come back as nil.
func (h *hydrator) typed(tag string, args []any, depth int) (any, error) {
	switch tag {
	case "Date", "BigInt", "URL", "URLSearchParams":
		if len(args) == 1 {
			if s, ok := args[0].(string); ok {
				return s, nil
			}
		}
		return nil, nil
	case "RegExp":
		if len(args) >= 1 {
			if s, ok := args[0].(string); ok {
				return s, nil
			}
		}
		return nil, nil
	case "Object":
		// A boxed primitive: new String("x") and friends.
		if len(args) == 1 {
			return h.ref(args[0], depth+1)
		}
		return nil, nil
	case "Set":
		out := make([]any, 0, len(args))
		for _, r := range args {
			child, err := h.ref(r, depth+1)
			if err != nil {
				return nil, err
			}
			out = append(out, child)
		}
		return out, nil
	case "Map":
		out := make(map[string]any, len(args)/2)
		for i := 0; i+1 < len(args); i += 2 {
			key, err := h.ref(args[i], depth+1)
			if err != nil {
				return nil, err
			}
			val, err := h.ref(args[i+1], depth+1)
			if err != nil {
				return nil, err
			}
			out[fmt.Sprint(key)] = val
		}
		return out, nil
	case "null":
		// An object with a null prototype: key, then a reference, pairwise.
		out := make(map[string]any, len(args)/2)
		for i := 0; i+1 < len(args); i += 2 {
			key, ok := args[i].(string)
			if !ok {
				return nil, fmt.Errorf("devalue: null-prototype key %v is not a string", args[i])
			}
			val, err := h.ref(args[i+1], depth+1)
			if err != nil {
				return nil, err
			}
			out[key] = val
		}
		return out, nil
	case "Promise":
		if len(args) != 1 || h.resolve == nil {
			return nil, nil
		}
		id, err := h.ref(args[0], depth+1)
		if err != nil {
			return nil, err
		}
		n, ok := id.(float64)
		if !ok {
			return nil, fmt.Errorf("devalue: promise id %v is not a number", id)
		}
		return h.resolve(int(n))
	default:
		return nil, nil
	}
}

// decodeInto reads a hydrated value into a struct, the way encoding/json
// would read the same value written as plain JSON.
func decodeInto(v, out any) error {
	b, err := jsonx.Marshal(v)
	if err != nil {
		return err
	}
	return jsonx.Unmarshal(b, out)
}

// dataResponse is a parsed __data.json answer.
type dataResponse struct {
	Type     string // "data", "redirect" or "error"
	Location string // where a redirect points
	Nodes    []any  // one hydrated value per route node; nil for a skipped one
	Path     string // the page the data belongs to, after any redirect (set by getData)
}

type wireNode struct {
	Type string `json:"type"`
	Data any    `json:"data"`
}

type wireHead struct {
	Type     string     `json:"type"`
	Location string     `json:"location"`
	Nodes    []wireNode `json:"nodes"`
	Error    any        `json:"error"`
}

type wireChunk struct {
	Type  string `json:"type"`
	ID    int    `json:"id"`
	Data  any    `json:"data"`
	Error any    `json:"error"`
}

// parseDataResponse reads a __data.json body: one JSON document per line,
// the head first and then one chunk per promise.
func parseDataResponse(body []byte) (dataResponse, error) {
	lines := bytes.Split(bytes.TrimSpace(body), []byte("\n"))
	if len(lines) == 0 || len(bytes.TrimSpace(lines[0])) == 0 {
		return dataResponse{}, errors.New("empty response")
	}

	var head wireHead
	if err := jsonx.Unmarshal(bytes.TrimSpace(lines[0]), &head); err != nil {
		return dataResponse{}, fmt.Errorf("malformed data response: %w", err)
	}
	switch head.Type {
	case "redirect":
		if head.Location == "" {
			return dataResponse{}, errors.New("redirect without a location")
		}
		return dataResponse{Type: head.Type, Location: head.Location}, nil
	case "error":
		return dataResponse{}, fmt.Errorf("the page answered with an error: %v", head.Error)
	case "data":
	default:
		return dataResponse{}, fmt.Errorf("unknown data response type %q", head.Type)
	}

	// Chunks are kept raw and hydrated on first use: a chunk may itself hold
	// a promise that a later line settles.
	chunks := make(map[int]wireChunk)
	for _, line := range lines[1:] {
		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}
		var c wireChunk
		if err := jsonx.Unmarshal(line, &c); err != nil {
			return dataResponse{}, fmt.Errorf("malformed chunk: %w", err)
		}
		if c.Type == "chunk" {
			chunks[c.ID] = c
		}
	}
	settled := make(map[int]any)
	pending := make(map[int]bool)
	var resolve func(id int) (any, error)
	resolve = func(id int) (any, error) {
		if v, ok := settled[id]; ok {
			return v, nil
		}
		c, ok := chunks[id]
		if !ok || c.Error != nil {
			// Never settled, or rejected: the value is simply absent.
			return nil, nil
		}
		if pending[id] {
			return nil, errDevalueCycle
		}
		// Never cleared: once settled the value is memoised above, and a
		// failure ends the whole parse.
		pending[id] = true
		v, err := hydrateFlat(c.Data, resolve)
		if err != nil {
			return nil, fmt.Errorf("chunk %d: %w", id, err)
		}
		settled[id] = v
		return v, nil
	}

	out := dataResponse{Type: head.Type, Nodes: make([]any, len(head.Nodes))}
	for i, n := range head.Nodes {
		if n.Type != "data" {
			continue // "skip": the layout node a client already has
		}
		v, err := hydrateFlat(n.Data, resolve)
		if err != nil {
			return dataResponse{}, fmt.Errorf("node %d: %w", i, err)
		}
		out.Nodes[i] = v
	}
	return out, nil
}

// nodeWith returns the first node that is an object holding key: the page
// node, whichever position the layout leaves it in.
func (r dataResponse) nodeWith(key string) (map[string]any, bool) {
	for _, n := range r.Nodes {
		if m, ok := n.(map[string]any); ok {
			if _, has := m[key]; has {
				return m, true
			}
		}
	}
	return nil, false
}
