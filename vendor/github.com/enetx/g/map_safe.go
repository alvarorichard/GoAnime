package g

import (
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"
)

// MapSafe is a concurrent-safe generic map built on sync.Map.
type MapSafe[K comparable, V any] struct {
	data     sync.Map
	count    atomic.Int64
	structMu sync.RWMutex // coordinates key-presence changes with Clear
}

// NewMapSafe creates a new instance of MapSafe.
func NewMapSafe[K comparable, V any]() *MapSafe[K, V] { return &MapSafe[K, V]{} }

// MapSafeOf creates a new MapSafe from the provided key-value pairs.
// Duplicate keys keep the last-written value, mirroring MapOf / MapOrdOf.
func MapSafeOf[K comparable, V any](pairs ...Pair[K, V]) *MapSafe[K, V] {
	ms := NewMapSafe[K, V]()
	for _, p := range pairs {
		ms.Insert(p.Unpack())
	}

	return ms
}

// Transform applies a transformation function to the MapSafe and returns the result.
func (ms *MapSafe[K, V]) Transform[U any](fn func(*MapSafe[K, V]) U) U {
	return fn(ms)
}

// Iter provides a thread-safe iterator over the MapSafe's key-value pairs.
func (ms *MapSafe[K, V]) Iter() Seq2[K, V] {
	return func(yield func(K, V) bool) {
		ms.data.Range(func(key, value any) bool {
			return yield(key.(K), *(value.(*V)))
		})
	}
}

// Entry returns a SafeEntry for the given key.
func (ms *MapSafe[K, V]) Entry(key K) SafeEntry[K, V] {
	if _, ok := ms.data.Load(key); ok {
		return OccupiedSafeEntry[K, V]{m: ms, key: key}
	}

	return VacantSafeEntry[K, V]{m: ms, key: key}
}

// Keys returns a slice of the MapSafe's keys.
func (ms *MapSafe[K, V]) Keys() Slice[K] {
	keys := NewSlice[K](0, ms.Len())

	ms.data.Range(func(key, _ any) bool {
		keys = append(keys, key.(K))
		return true
	})

	return keys
}

// Values returns a slice of the MapSafe's values.
func (ms *MapSafe[K, V]) Values() Slice[V] {
	values := NewSlice[V](0, ms.Len())

	ms.data.Range(func(_, value any) bool {
		values = append(values, *(value.(*V)))
		return true
	})

	return values
}

// Contains checks if the MapSafe contains the specified key.
func (ms *MapSafe[K, V]) Contains(key K) bool {
	_, ok := ms.data.Load(key)
	return ok
}

// Clone creates a deep copy of the MapSafe.
func (ms *MapSafe[K, V]) Clone() *MapSafe[K, V] {
	res := NewMapSafe[K, V]()

	ms.data.Range(func(key, value any) bool {
		v := *(value.(*V))
		res.data.Store(key, &v)
		res.count.Add(1)
		return true
	})

	return res
}

// Copy performs a deep copy of the source MapSafe's pairs into the current map.
func (ms *MapSafe[K, V]) Copy(src *MapSafe[K, V]) {
	ms.structMu.RLock()
	defer ms.structMu.RUnlock()

	src.data.Range(func(key, value any) bool {
		v := *(value.(*V))
		_, loaded := ms.data.Swap(key, &v)
		if !loaded {
			ms.count.Add(1)
		}

		return true
	})
}

// Remove removes the specified key from the MapSafe and returns the removed value.
func (ms *MapSafe[K, V]) Remove(key K) Option[V] {
	ms.structMu.RLock()
	defer ms.structMu.RUnlock()

	if v, loaded := ms.data.LoadAndDelete(key); loaded {
		ms.count.Add(-1)
		return Some(*(v.(*V)))
	}

	return None[V]()
}

// Eq checks if two MapSafes are equal by deep-comparing their values.
func (ms *MapSafe[K, V]) Eq(other *MapSafe[K, V]) bool {
	n := ms.Len()
	if n != other.Len() {
		return false
	}

	if n == 0 {
		return true
	}

	comparable := isValueComparable[V]()
	equal := true

	ms.data.Range(func(key, value any) bool {
		ovalue, ok := other.data.Load(key)
		if !ok {
			equal = false
			return false
		}

		v1 := *(value.(*V))
		v2 := *(ovalue.(*V))

		if comparable {
			if any(v1) != any(v2) {
				equal = false
				return false
			}
		} else {
			if !reflect.DeepEqual(v1, v2) {
				equal = false
				return false
			}
		}

		return true
	})

	return equal
}

// Get retrieves the value associated with the given key.
func (ms *MapSafe[K, V]) Get(key K) Option[V] {
	if value, ok := ms.data.Load(key); ok {
		return Some(*(value.(*V)))
	}

	return None[V]()
}

// Insert stores the value for the given key.
// Returns Some(previous_value) if the key existed, None if it was newly inserted.
//
// Example:
//
//	ms := NewMapSafe[string, int]()
//	ms.Insert("a", 1)        // None (new key)
//	ms.Insert("a", 2)        // Some(1) (replaced)
//	ms.Get("a").Some()    // 2
func (ms *MapSafe[K, V]) Insert(key K, value V) Option[V] {
	ms.structMu.RLock()
	defer ms.structMu.RUnlock()

	if previous, loaded := ms.data.Swap(key, &value); loaded {
		return Some(*(previous.(*V)))
	}

	ms.count.Add(1)
	return None[V]()
}

// TryInsert inserts value only if the key is absent.
// Returns Some(existing_value) if key already existed (no insert), None if inserted.
//
// Example:
//
//	ms := NewMapSafe[string, int]()
//	ms.TryInsert("a", 1)     // None (inserted)
//	ms.TryInsert("a", 2)     // Some(1) (already existed, not replaced)
//	ms.Get("a").Some()    // 1
func (ms *MapSafe[K, V]) TryInsert(key K, value V) Option[V] {
	ms.structMu.RLock()
	defer ms.structMu.RUnlock()

	if actual, loaded := ms.data.LoadOrStore(key, &value); loaded {
		return Some(*(actual.(*V)))
	}

	ms.count.Add(1)
	return None[V]()
}

// Len returns the number of key-value pairs in the MapSafe.
func (ms *MapSafe[K, V]) Len() Int { return Int(ms.count.Load()) }

// Ne checks if two MapSafes are not equal.
func (ms *MapSafe[K, V]) Ne(other *MapSafe[K, V]) bool { return !ms.Eq(other) }

// Clear removes all key-value pairs from the MapSafe.
func (ms *MapSafe[K, V]) Clear() {
	ms.structMu.Lock()
	defer ms.structMu.Unlock()

	ms.data.Clear()
	ms.count.Store(0)
}

// IsEmpty checks if the MapSafe is empty.
func (ms *MapSafe[K, V]) IsEmpty() bool { return ms.count.Load() == 0 }

// String returns a string representation of the MapSafe.
func (ms *MapSafe[K, V]) String() string {
	var b Builder
	b.Grow(ms.Len() * 16)
	b.WriteString("MapSafe{")

	first := true

	ms.data.Range(func(key, value any) bool {
		if !first {
			b.WriteString(", ")
		}

		first = false

		if vptr, ok := value.(*V); ok && vptr != nil {
			fmt.Fprint(&b, key)
			b.WriteByte(':')
			fmt.Fprint(&b, *vptr)
		} else {
			fmt.Fprint(&b, key)
			b.WriteString(":<invalid>")
		}

		return true
	})

	b.WriteString("}")

	return b.String().Std()
}

// Print writes the MapSafe to standard output.
func (ms *MapSafe[K, V]) Print() *MapSafe[K, V] { fmt.Print(ms); return ms }

// Println writes the MapSafe to standard output with a newline.
func (ms *MapSafe[K, V]) Println() *MapSafe[K, V] { fmt.Println(ms); return ms }
