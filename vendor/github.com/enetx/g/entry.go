package g

// Entry is a sealed interface representing a view into a single Map entry.
//
// Entry provides an API for in-place manipulation of map entries, enabling
// efficient "get or insert" patterns without redundant lookups.
//
// The interface is sealed to ensure type safety; implementations are limited
// to [OccupiedEntry] (when the key exists) and [VacantEntry] (when the key
// is absent). Use a type switch to access type-specific methods like Get,
// Insert, or Remove.
//
// Common usage patterns:
//
//	// Increment existing value or insert default
//	m.Entry("counter").AndModify(func(v *int) { *v++ }).OrInsert(1)
//
//	// Insert only if absent
//	m.Entry("key").OrInsert(defaultValue)
//
//	// Insert with lazy initialization
//	m.Entry("key").OrInsertWith(func() V { return expensiveComputation() })
//
//	// Type switch for fine-grained control
//	switch e := m.Entry("key").(type) {
//	case OccupiedEntry[string, int]:
//	    fmt.Println("exists:", e.Get())
//	case VacantEntry[string, int]:
//	    e.Insert(42)
//	}
//
// An Entry is a short-lived view of the key state observed by Map.Entry.
// Do not retain it across external insertion or removal of the same key;
// obtain a fresh Entry after structurally changing that key.
type Entry[K comparable, V any] interface {
	sealed()
	Key() K
	OrInsert(value V) V
	OrInsertWith(fn func() V) V
	OrInsertWithKey(fn func(K) V) V
	OrDefault() V
	AndModify(fn func(*V)) Entry[K, V]
}

// OccupiedEntry represents a view into a map entry that is known to be present.
//
// It is typically obtained from Map.Entry(key) when the key already exists.
// OccupiedEntry allows inspecting, modifying, replacing, or removing the value
// associated with the key without performing additional map lookups.
type OccupiedEntry[K comparable, V any] struct {
	m   Map[K, V]
	key K
}

// sealed prevents external implementations of the Entry interface.
func (OccupiedEntry[K, V]) sealed() {}

// Key returns the key of this occupied entry.
func (e OccupiedEntry[K, V]) Key() K { return e.key }

// Get returns the current value associated with the key.
//
// The value is returned by copy, consistent with Go map semantics.
func (e OccupiedEntry[K, V]) Get() V { return e.m[e.key] }

// Insert replaces the value in the map with the provided one
// and returns the previous value.
//
// The key remains present in the map.
func (e OccupiedEntry[K, V]) Insert(value V) V {
	old := e.m[e.key]
	e.m[e.key] = value
	return old
}

// Remove removes the entry from the map and returns the previously stored value.
//
// After this call, the key is no longer present in the map.
func (e OccupiedEntry[K, V]) Remove() V {
	v := e.m[e.key]
	delete(e.m, e.key)
	return v
}

// OrInsert returns the existing value without modifying the map.
//
// For OccupiedEntry, this is equivalent to Get since the key already exists.
func (e OccupiedEntry[K, V]) OrInsert(value V) V { return e.Get() }

// OrInsertWith returns the existing value without invoking the function.
//
// For OccupiedEntry, the function is never called since the key already exists.
func (e OccupiedEntry[K, V]) OrInsertWith(fn func() V) V { return e.Get() }

// OrInsertWithKey returns the existing value without invoking the function.
//
// For OccupiedEntry, the function is never called since the key already exists.
func (e OccupiedEntry[K, V]) OrInsertWithKey(fn func(K) V) V { return e.Get() }

// OrDefault returns the existing value.
//
// For OccupiedEntry, this is equivalent to Get since the key already exists.
func (e OccupiedEntry[K, V]) OrDefault() V { return e.Get() }

// AndModify applies the provided function to the value stored in the map
// and returns the entry for method chaining.
//
// The function receives a pointer to a copy of the value; after modification,
// the updated value is written back to the map.
// The entry must not be used after the same key is externally removed or replaced.
//
// Example:
//
//	m.Entry("count").AndModify(func(v *int) { *v++ }).OrInsert(1)
func (e OccupiedEntry[K, V]) AndModify(fn func(*V)) Entry[K, V] {
	v := e.m[e.key]
	fn(&v)
	e.m[e.key] = v
	return e
}

// VacantEntry represents a view into a map entry that is known to be absent.
//
// It is typically obtained from Map.Entry(key) when the key does not exist.
// VacantEntry allows inserting a value for the key in a controlled manner.
type VacantEntry[K comparable, V any] struct {
	m   Map[K, V]
	key K
}

// sealed prevents external implementations of the Entry interface.
func (VacantEntry[K, V]) sealed() {}

// Key returns the key that would be used for insertion.
func (e VacantEntry[K, V]) Key() K { return e.key }

// Insert inserts the provided value into the map and returns it.
//
// After this call, the key is present in the map with the given value.
func (e VacantEntry[K, V]) Insert(value V) V {
	e.m[e.key] = value
	return value
}

// OrInsert inserts the provided value and returns it.
//
// This is the primary method for inserting values via VacantEntry.
func (e VacantEntry[K, V]) OrInsert(value V) V { return e.Insert(value) }

// OrInsertWith inserts the value returned by the function and returns it.
//
// The function is guaranteed to be called exactly once.
// Use this when computing the default value is expensive.
func (e VacantEntry[K, V]) OrInsertWith(fn func() V) V { return e.Insert(fn()) }

// OrInsertWithKey inserts the value returned by the function and returns it.
//
// The function receives the entry key and is guaranteed to be called exactly once.
// Use this when the default value depends on the key.
func (e VacantEntry[K, V]) OrInsertWithKey(fn func(K) V) V {
	return e.Insert(fn(e.key))
}

// OrDefault inserts the zero value of V into the map and returns it.
//
// This is useful for types where the zero value is a valid initial state,
// such as numeric types (0), slices (nil), or structs with zero defaults.
func (e VacantEntry[K, V]) OrDefault() V {
	var zero V
	return e.Insert(zero)
}

// AndModify does nothing for VacantEntry and returns the entry unchanged.
//
// Since there is no existing value to modify, the function is not called.
// This allows fluent chaining like Entry(k).AndModify(f).OrInsert(v)
// to work correctly regardless of whether the key exists.
func (e VacantEntry[K, V]) AndModify(fn func(*V)) Entry[K, V] { return e }
