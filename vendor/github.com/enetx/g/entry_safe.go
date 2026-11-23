package g

// SafeEntry is a sealed interface representing a view into a single MapSafe entry.
//
// SafeEntry provides an API for in-place manipulation of concurrent map entries,
// enabling efficient "get or insert" patterns that are safe for concurrent use
// by multiple goroutines.
//
// The interface is sealed to ensure type safety; implementations are limited
// to [OccupiedSafeEntry] (when the key exists) and [VacantSafeEntry] (when the
// key is absent). Use a type switch to access type-specific methods like Get,
// Insert, or Remove.
//
// Concurrency notes:
//   - AndModify uses a compare-and-swap (CAS) loop for atomic updates
//   - AndModify may invoke its callback more than once when CAS retries; the
//     callback must not perform non-idempotent external side effects
//   - VacantSafeEntry stores pending modifications to handle insertion races
//   - All operations are safe for concurrent use without external locking
//
// Common usage patterns:
//
//	// Thread-safe increment or insert (safe for concurrent goroutines)
//	ms.Entry("counter").AndModify(func(v *int) { *v++ }).OrInsert(1)
//
//	// Thread-safe insert only if absent
//	ms.Entry("key").OrInsert(defaultValue)
//
//	// Type switch for fine-grained control
//	switch e := ms.Entry("key").(type) {
//	case OccupiedSafeEntry[string, int]:
//	    fmt.Println("exists:", e.Get())
//	case VacantSafeEntry[string, int]:
//	    e.Insert(42)
//	}
type SafeEntry[K comparable, V any] interface {
	sealed()
	Key() K
	OrInsert(value V) V
	OrInsertWith(fn func() V) V
	OrInsertWithKey(fn func(K) V) V
	OrDefault() V
	AndModify(fn func(*V)) SafeEntry[K, V]
}

// OccupiedSafeEntry represents a view into a concurrent map entry that is known
// to be present.
//
// It is typically obtained from MapSafe.Entry(key) when the key exists.
// All operations on OccupiedSafeEntry are safe for concurrent use.
type OccupiedSafeEntry[K comparable, V any] struct {
	m   *MapSafe[K, V]
	key K
}

// sealed prevents external implementations of the SafeEntry interface.
func (OccupiedSafeEntry[K, V]) sealed() {}

// Key returns the key of this entry.
func (e OccupiedSafeEntry[K, V]) Key() K { return e.key }

// Get returns the current value associated with the key.
//
// If the key is concurrently removed, the zero value of V is returned.
func (e OccupiedSafeEntry[K, V]) Get() V {
	if actual, ok := e.m.data.Load(e.key); ok {
		return *actual.(*V)
	}
	var zero V
	return zero
}

// Insert replaces the value associated with the key and returns the previous value.
//
// The replacement is performed atomically with respect to other map operations.
func (e OccupiedSafeEntry[K, V]) Insert(value V) V {
	e.m.structMu.RLock()
	defer e.m.structMu.RUnlock()

	old, loaded := e.m.data.Swap(e.key, &value)
	if loaded {
		return *old.(*V)
	}

	e.m.count.Add(1)
	var zero V
	return zero
}

// Remove removes the entry from the map and returns the previously stored value.
//
// If the key is concurrently removed, the zero value of V is returned.
func (e OccupiedSafeEntry[K, V]) Remove() V {
	e.m.structMu.RLock()
	defer e.m.structMu.RUnlock()

	if actual, loaded := e.m.data.LoadAndDelete(e.key); loaded {
		e.m.count.Add(-1)
		return *actual.(*V)
	}

	var zero V
	return zero
}

// OrInsert returns the existing value without modifying the map.
// If the key was concurrently removed, the value is inserted as for a vacant entry.
func (e OccupiedSafeEntry[K, V]) OrInsert(value V) V {
	if actual, ok := e.m.data.Load(e.key); ok {
		return *actual.(*V)
	}

	return VacantSafeEntry[K, V]{m: e.m, key: e.key}.OrInsert(value)
}

// OrInsertWith returns the existing value if present, or inserts the result
// of fn() and returns it.
//
// Note: Due to concurrent access, fn() may be invoked even if another
// goroutine inserts the key between the check and insertion. In this case,
// the result of fn() is discarded and the existing value is returned.
func (e OccupiedSafeEntry[K, V]) OrInsertWith(fn func() V) V {
	if actual, ok := e.m.data.Load(e.key); ok {
		return *actual.(*V)
	}

	return VacantSafeEntry[K, V]{m: e.m, key: e.key}.OrInsertWith(fn)
}

// OrInsertWithKey returns the existing value without invoking the function.
// If the key was concurrently removed, fn is invoked with the key and its
// result inserted as for a vacant entry.
func (e OccupiedSafeEntry[K, V]) OrInsertWithKey(fn func(K) V) V {
	if actual, ok := e.m.data.Load(e.key); ok {
		return *actual.(*V)
	}

	return VacantSafeEntry[K, V]{m: e.m, key: e.key}.OrInsertWithKey(fn)
}

// OrDefault returns the existing value.
// If the key was concurrently removed, the zero value of V is inserted and returned.
func (e OccupiedSafeEntry[K, V]) OrDefault() V {
	if actual, ok := e.m.data.Load(e.key); ok {
		return *actual.(*V)
	}

	var zero V
	return VacantSafeEntry[K, V]{m: e.m, key: e.key}.Insert(zero)
}

// AndModify applies the provided function to the value associated with the key
// and returns the entry.
//
// The modification is performed using a compare-and-swap loop.
// The function receives a pointer to a copy of the value; the updated value
// is written back atomically.
// Under contention, fn may be invoked more than once before CAS succeeds.
// It should only modify the provided value and must not rely on exactly-once
// external side effects.
//
// If the key is concurrently removed, AndModify becomes a no-op.
func (e OccupiedSafeEntry[K, V]) AndModify(fn func(*V)) SafeEntry[K, V] {
	for {
		actual, ok := e.m.data.Load(e.key)
		if !ok {
			return e
		}

		oldPtr := actual.(*V)
		newVal := *oldPtr
		fn(&newVal)

		if e.m.data.CompareAndSwap(e.key, oldPtr, &newVal) {
			return e
		}
	}
}

// VacantSafeEntry represents a view into a concurrent map entry that is known
// to be absent at the time of creation.
//
// It is typically obtained from MapSafe.Entry(key) when the key does not exist.
// All operations on VacantSafeEntry are safe for concurrent use.
//
// The modify field stores a pending modification function from AndModify,
// which will be applied if OrInsert loses a race with another goroutine.
type VacantSafeEntry[K comparable, V any] struct {
	m      *MapSafe[K, V]
	key    K
	modify func(*V)
}

// sealed prevents external implementations of the SafeEntry interface.
func (VacantSafeEntry[K, V]) sealed() {}

// Key returns the key that would be used for insertion.
func (e VacantSafeEntry[K, V]) Key() K { return e.key }

// applyPending applies the pending AndModify function (if any) to the value now
// stored under the key, then returns the resulting stored value.
//
// It is used by the Or* methods when an insert loses the race to a concurrent
// goroutine: the modification registered via AndModify must still be applied to
// the value that actually won the race. If the key has since been removed,
// fallback (the value observed during the failed insert) is returned instead.
func (e VacantSafeEntry[K, V]) applyPending(fallback *V) V {
	if e.modify != nil {
		OccupiedSafeEntry[K, V]{m: e.m, key: e.key}.AndModify(e.modify)
		if val, ok := e.m.data.Load(e.key); ok {
			return *val.(*V)
		}
	}

	return *fallback
}

// Insert inserts the provided value into the map and returns the stored value.
//
// If another goroutine inserts the same key concurrently, the existing value
// is returned instead. Equivalent to OrInsert for VacantSafeEntry.
func (e VacantSafeEntry[K, V]) Insert(value V) V { return e.OrInsert(value) }

// OrInsert inserts the provided value and returns the stored value.
//
// If another goroutine inserts the same key concurrently (the insert "loses
// the race"), the existing value is used instead. In this case, if a pending
// modification was registered via AndModify, it is applied atomically to the
// existing value before returning.
//
// This ensures that chained calls like Entry(k).AndModify(f).OrInsert(v)
// behave correctly under concurrent access: under contended insertion the
// modification is applied to the winning value.
//
// Edge case: if the key is also removed by another goroutine between the lost
// insert and the modify, the modification has nothing to apply to and the
// pre-modify value observed during the failed insert is returned. In that
// narrow window the modification can be lost.
func (e VacantSafeEntry[K, V]) OrInsert(value V) V {
	e.m.structMu.RLock()
	defer e.m.structMu.RUnlock()

	actual, loaded := e.m.data.LoadOrStore(e.key, &value)
	if !loaded {
		e.m.count.Add(1)
	}

	if loaded {
		return e.applyPending(actual.(*V))
	}

	return *actual.(*V)
}

// OrInsertWith inserts the value returned by the function and returns the
// stored value.
//
// Note: Due to lock-free implementation, fn() is evaluated before the atomic
// insertion. If another goroutine inserts the key concurrently, the result
// of fn() may be discarded and the existing value is returned instead.
func (e VacantSafeEntry[K, V]) OrInsertWith(fn func() V) V {
	if actual, ok := e.m.data.Load(e.key); ok {
		return e.applyPending(actual.(*V))
	}

	e.m.structMu.RLock()
	defer e.m.structMu.RUnlock()

	actual, loaded := e.m.data.LoadOrStore(e.key, new(fn()))
	if !loaded {
		e.m.count.Add(1)
	}

	if loaded {
		return e.applyPending(actual.(*V))
	}

	return *actual.(*V)
}

// OrInsertWithKey inserts the value returned by the function and returns the
// stored value.
//
// Note: Due to lock-free implementation, fn() is evaluated before the atomic
// insertion. If another goroutine inserts the key concurrently, the result
// of fn() may be discarded and the existing value is returned instead.
func (e VacantSafeEntry[K, V]) OrInsertWithKey(fn func(K) V) V {
	if actual, ok := e.m.data.Load(e.key); ok {
		return e.applyPending(actual.(*V))
	}

	e.m.structMu.RLock()
	defer e.m.structMu.RUnlock()

	actual, loaded := e.m.data.LoadOrStore(e.key, new(fn(e.key)))
	if !loaded {
		e.m.count.Add(1)
	}

	if loaded {
		return e.applyPending(actual.(*V))
	}

	return *actual.(*V)
}

// OrDefault inserts the zero value of V into the map and returns the stored value.
//
// If another goroutine inserts the same key concurrently and a pending
// modification was registered via AndModify, it is applied atomically
// to the existing value before returning.
func (e VacantSafeEntry[K, V]) OrDefault() V {
	var zero V
	return e.OrInsert(zero)
}

// AndModify registers a modification function to be applied to the value.
//
// If the key was concurrently inserted by another goroutine since this
// VacantSafeEntry was created, the modification is applied immediately
// via OccupiedSafeEntry.AndModify.
//
// Otherwise, the function is stored and will be applied later by OrInsert
// if it loses the race to insert the key. This ensures that the pattern
// Entry(k).AndModify(f).OrInsert(v) correctly increments existing values
// even under heavy concurrent access.
//
// Returns the appropriate SafeEntry for method chaining.
func (e VacantSafeEntry[K, V]) AndModify(fn func(*V)) SafeEntry[K, V] {
	if _, ok := e.m.data.Load(e.key); ok {
		return OccupiedSafeEntry[K, V]{m: e.m, key: e.key}.AndModify(fn)
	}

	return VacantSafeEntry[K, V]{m: e.m, key: e.key, modify: fn}
}
