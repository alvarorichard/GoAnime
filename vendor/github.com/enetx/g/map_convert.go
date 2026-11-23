package g

// Conversions between the three map types live here as free functions rather
// than methods. As methods they weld Map, MapOrd and MapSafe into one
// instantiation cluster: any package touching one of them would compile the
// methods of all three (and everything those methods mention, transitively).
// As free functions each conversion costs only the packages that call it.

// MapOrdFromMap converts a standard Map to an ordered Map.
func MapOrdFromMap[K comparable, V any](m Map[K, V]) MapOrd[K, V] {
	mo := make(MapOrd[K, V], 0, len(m))
	for k, v := range m {
		mo = append(mo, Pair[K, V]{Key: k, Value: v})
	}

	return mo
}

// MapSafeFromMap converts a standard Map to a thread-safe Map.
func MapSafeFromMap[K comparable, V any](m Map[K, V]) *MapSafe[K, V] {
	ms := NewMapSafe[K, V]()
	for k, v := range m {
		ms.Insert(k, v)
	}

	return ms
}

// MapFromMapOrd converts an ordered Map to a standard Map.
func MapFromMapOrd[K comparable, V any](mo MapOrd[K, V]) Map[K, V] {
	m := NewMap[K, V](mo.Len())
	for _, p := range mo {
		m[p.Key] = p.Value
	}

	return m
}

// MapSafeFromMapOrd converts an ordered Map to a thread-safe Map.
func MapSafeFromMapOrd[K comparable, V any](mo MapOrd[K, V]) *MapSafe[K, V] {
	ms := NewMapSafe[K, V]()
	for _, p := range mo {
		ms.Insert(p.Unpack())
	}

	return ms
}

// MapFromMapSafe converts the MapSafe to a standard Map by taking a snapshot of its
// current key-value pairs.
//
// The returned Map is an independent, non-thread-safe copy; subsequent
// mutations to the MapSafe are not reflected in it.
func MapFromMapSafe[K comparable, V any](ms *MapSafe[K, V]) Map[K, V] {
	m := NewMap[K, V](ms.Len())

	ms.data.Range(func(key, value any) bool {
		m[key.(K)] = *(value.(*V))
		return true
	})

	return m
}

// MapOrdFromMapSafe converts the MapSafe to an ordered Map by taking a snapshot of its
// current key-value pairs.
//
// Because MapSafe does not track insertion order, the order of the returned
// MapOrd is unspecified. The returned MapOrd is an independent, non-thread-safe
// copy; subsequent mutations to the MapSafe are not reflected in it.
func MapOrdFromMapSafe[K comparable, V any](ms *MapSafe[K, V]) MapOrd[K, V] {
	mo := NewMapOrd[K, V](ms.Len())

	ms.data.Range(func(key, value any) bool {
		mo = append(mo, Pair[K, V]{Key: key.(K), Value: *(value.(*V))})
		return true
	})

	return mo
}
