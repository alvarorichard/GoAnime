// Package headerorder implements the shared wire-header order for HTTP/1, HTTP/2 and HTTP/3.
package headerorder

import (
	"sort"
	"strings"
)

const (
	HeaderOrderKey  = "Header-Order:"
	PHeaderOrderKey = "PHeader-Order:"
	HeaderCaseKey   = "Header-Case:"
)

func IsMetadataKey(key string) bool {
	return key == HeaderOrderKey || key == PHeaderOrderKey || key == HeaderCaseKey
}

type Field struct {
	Key    string
	Values []string
}

func Indices(names []string) map[string]int {
	if names == nil {
		return nil
	}
	order := make(map[string]int, len(names))
	for i, name := range names {
		order[strings.ToLower(name)] = i
	}
	return order
}

// Sorted returns a fresh slice. It never changes the caller's header values,
// order or exclusion map. Metadata keys never appear in the result.
func Sorted(h map[string][]string, order map[string]int, exclude map[string]bool) []Field {
	fields := make([]Field, 0, len(h))
	for k, values := range h {
		if IsMetadataKey(k) || exclude[k] {
			continue
		}
		fields = append(fields, Field{k, values})
	}
	sort.Slice(fields, func(i, j int) bool {
		a, aok := order[strings.ToLower(fields[i].Key)]
		b, bok := order[strings.ToLower(fields[j].Key)]
		if aok != bok {
			return aok
		}
		if aok && a != b {
			return a < b
		}
		return fields[i].Key < fields[j].Key
	})
	return fields
}
