package g

import (
	"fmt"
	"maps"
	"reflect"
)

// Map is a generic alias for a map.
type Map[K comparable, V any] map[K]V

// NewMap creates a new Map of the specified size or an empty Map if no size is provided.
func NewMap[K comparable, V any](size ...Int) Map[K, V] {
	if len(size) > 0 {
		return make(Map[K, V], size[0])
	}

	return make(Map[K, V])
}

// Transform applies a transformation function to the Map and returns the result.
func (m Map[K, V]) Transform[U any](fn func(Map[K, V]) U) U { return fn(m) }

// Entry returns an Entry for the given key.
func (m Map[K, V]) Entry(key K) Entry[K, V] {
	if _, ok := m[key]; ok {
		return OccupiedEntry[K, V]{m: m, key: key}
	}

	return VacantEntry[K, V]{m: m, key: key}
}

// Iter returns an iterator (Seq2[K, V]) for the Map, allowing for sequential iteration
// over its key-value pairs. It is commonly used in combination with higher-order functions,
// such as 'ForEach', to perform operations on each key-value pair of the Map.
//
// Returns:
//
// - Seq2[K, V], which can be used for sequential iteration over the key-value pairs of the Map.
//
// Example usage:
//
//	myMap := g.Map[string, int]{"one": 1, "two": 2, "three": 3}
//	iterator := myMap.Iter()
//	iterator.ForEach(func(key string, value int) {
//		// Perform some operation on each key-value pair
//		fmt.Printf("%s: %d\n", key, value)
//	})
//
// The 'Iter' method provides a convenient way to traverse the key-value pairs of a Map
// in a functional style, enabling operations like mapping or filtering.
func (m Map[K, V]) Iter() Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for k, v := range m {
			if !yield(k, v) {
				return
			}
		}
	}
}

// Keys returns a slice of the Map's keys.
func (m Map[K, V]) Keys() Slice[K] {
	if m.IsEmpty() {
		return NewSlice[K]()
	}

	keys := make(Slice[K], 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}

	return keys
}

// Values returns a slice of the Map's values.
func (m Map[K, V]) Values() Slice[V] {
	if m.IsEmpty() {
		return NewSlice[V]()
	}

	values := make(Slice[V], 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}

	return values
}

// Contains checks if the Map contains the specified key.
func (m Map[K, V]) Contains(key K) bool {
	_, ok := m[key]
	return ok
}

// Clone creates a new Map that is a copy of the original Map.
func (m Map[K, V]) Clone() Map[K, V] { return maps.Clone(m) }

// Copy copies the source Map's key-value pairs to the target Map.
func (m Map[K, V]) Copy(src Map[K, V]) { maps.Copy(m, src) }

// Remove removes the specified key from the Map and returns the removed value.
func (m Map[K, V]) Remove(key K) Option[V] {
	if v, ok := m[key]; ok {
		delete(m, key)
		return Some(v)
	}

	return None[V]()
}

// Std converts the Map to a regular Go map.
func (m Map[K, V]) Std() map[K]V { return m }

// Eq checks if two Maps are equal.
func (m Map[K, V]) Eq(other Map[K, V]) bool {
	n := len(m)
	if n != len(other) {
		return false
	}
	if n == 0 {
		return true
	}

	comparable := isValueComparable[V]()

	for k, value := range m {
		ovalue, ok := other[k]
		if !ok {
			return false
		}

		if comparable {
			if any(value) != any(ovalue) {
				return false
			}
		} else {
			if !reflect.DeepEqual(value, ovalue) {
				return false
			}
		}
	}

	return true
}

// String returns a string representation of the Map.
func (m Map[K, V]) String() string {
	if len(m) == 0 {
		return "Map{}"
	}

	var b Builder
	b.Grow(Int(len(m)) * 16)
	b.WriteString("Map{")

	first := true
	for k, v := range m {
		if !first {
			b.WriteString(", ")
		}

		first = false
		fmt.Fprint(&b, k)
		b.WriteByte(':')
		fmt.Fprint(&b, v)
	}

	b.WriteString("}")

	return b.String().Std()
}

// Clear removes all key-value pairs from the Map.
func (m Map[K, V]) Clear() { clear(m) }

// IsEmpty checks if the Map is empty.
func (m Map[K, V]) IsEmpty() bool { return len(m) == 0 }

// Get retrieves the value associated with the given key.
func (m Map[K, V]) Get(k K) Option[V] {
	if v, ok := m[k]; ok {
		return Some(v)
	}

	return None[V]()
}

// Len returns the number of key-value pairs in the Map.
func (m Map[K, V]) Len() Int { return Int(len(m)) }

// Ne checks if two Maps are not equal.
func (m Map[K, V]) Ne(other Map[K, V]) bool { return !m.Eq(other) }

// Insert sets the value for the key and returns the previous value if it existed.
func (m Map[K, V]) Insert(key K, value V) Option[V] {
	prev, ok := m[key]
	m[key] = value
	if ok {
		return Some(prev)
	}

	return None[V]()
}

// Print writes the key-value pairs of the Map to the standard output (console)
// and returns the Map unchanged.
func (m Map[K, V]) Print() Map[K, V] { fmt.Print(m); return m }

// Println writes the key-value pairs of the Map to the standard output (console) with a newline
// and returns the Map unchanged.
func (m Map[K, V]) Println() Map[K, V] { fmt.Println(m); return m }

// MapOf creates a Map from the provided key-value pairs.
//
// Example:
//
//	m := g.MapOf(g.PairOf("a", 1), g.PairOf("b", 2))
func MapOf[K comparable, V any](pairs ...Pair[K, V]) Map[K, V] {
	m := NewMap[K, V](Int(len(pairs)))
	for _, p := range pairs {
		m[p.Key] = p.Value
	}

	return m
}
