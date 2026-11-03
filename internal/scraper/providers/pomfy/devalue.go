package pomfy

import (
	"encoding/json"

	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
)

// SvelteKit ships its server data as "devalue": one flat array where every
// unique value appears once and plain JSON object/array entries point into
// the array by index. /serie/<tmdb>/__data.json answers with
//
//	nodes[1].data = [
//	  {"tvShow": 1},
//	  {"adult": 2, "backdrop_path": 3, …  "seasons": 67, …},
//	   false, "/tsRy63Mu5….jpg", [5], {"id": 6, "name": 8, …}, …
//	]
//
// so a reader walks references: a value that is a number N names data[N] and
// must be resolved recursively; anything non-numeric is literal. This file
// decodes just what the show pages carry — names, seasons, counts — and stays
// garbage-free on shapes it has not seen before (missing keys → zero value,
// mismatch → not-an-object rather than panic).

// devalueGraph holds the flat payload and memoizes resolved reads.
type devalueGraph struct {
	arr       []json.RawMessage
	objectAt  map[int]map[string]int64
	arrayAt   map[int][]int64
	stringAtI map[int]string
	intAtI    map[int]int64
}

func newDevalueGraph(arr []json.RawMessage) *devalueGraph {
	return &devalueGraph{
		arr:       arr,
		objectAt:  make(map[int]map[string]int64),
		arrayAt:   make(map[int][]int64),
		stringAtI: make(map[int]string),
		intAtI:    make(map[int]int64),
	}
}

func (g *devalueGraph) size() int { return len(g.arr) }

// object decodes arr[i] as an index-pointing object; false when it is not one.
func (g *devalueGraph) object(i int64) (map[string]int64, bool) {
	idx := int(i)
	if idx < 0 || idx >= len(g.arr) {
		return nil, false
	}
	if cached, ok := g.objectAt[idx]; ok {
		return cached, cached != nil
	}
	var raw map[string]json.RawMessage
	if err := jsonx.Unmarshal(g.arr[idx], &raw); err != nil {
		g.objectAt[idx] = nil
		return nil, false
	}
	out := make(map[string]int64, len(raw))
	for k, v := range raw {
		var n int64
		if jsonx.Unmarshal(v, &n) == nil {
			out[k] = n
		}
	}
	g.objectAt[idx] = out
	return out, true
}

// array decodes arr[i] as an index-pointing array; nil when it is not one.
func (g *devalueGraph) array(i int64) []int64 {
	idx := int(i)
	if idx < 0 || idx >= len(g.arr) {
		return nil
	}
	if cached, ok := g.arrayAt[idx]; ok {
		return cached
	}
	var raw []json.RawMessage
	if jsonx.Unmarshal(g.arr[idx], &raw) != nil {
		g.arrayAt[idx] = nil
		return nil
	}
	out := make([]int64, 0, len(raw))
	for _, v := range raw {
		var n int64
		if jsonx.Unmarshal(v, &n) == nil {
			out = append(out, n)
		}
	}
	g.arrayAt[idx] = out
	return out
}

// stringRef resolves arr[i] as a literal string; ok=false when it is not one.
func (g *devalueGraph) stringRef(i int64) (string, bool) {
	idx := int(i)
	if idx < 0 || idx >= len(g.arr) {
		return "", false
	}
	if cached, ok := g.stringAtI[idx]; ok {
		return cached, true
	}
	var s string
	if jsonx.Unmarshal(g.arr[idx], &s) != nil {
		g.stringAtI[idx] = ""
		return "", false
	}
	g.stringAtI[idx] = s
	return s, true
}

// integerRef resolves arr[i] as a literal integer (episode counts), 0 otherwise.
func (g *devalueGraph) integerRef(i int64) int64 {
	idx := int(i)
	if idx < 0 || idx >= len(g.arr) {
		return 0
	}
	if cached, ok := g.intAtI[idx]; ok {
		return cached
	}
	var n int64
	if jsonx.Unmarshal(g.arr[idx], &n) != nil {
		n = 0
	}
	g.intAtI[idx] = n
	return n
}
