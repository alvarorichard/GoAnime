package g

import "github.com/enetx/g/cmp"

// Collect returns a collector over the sequence. Collect itself is lazy and
// does not consume the sequence; only the collector's materializer methods do,
// turning the elements into a concrete container:
//
//	s.Iter().Filter(fn).Collect().Slice()
//	s.Iter().Filter(fn).Collect().Set[int]()
//	s.Iter().Collect().Heap(cmp.Cmp)
//	s.Iter().Collect().Deque()
//
// Set (like Map/MapOrd/MapSafe on the key-value collector) takes the element
// type as an explicit type argument: Go checks constraints at method
// DECLARATION, where the sequence's own type parameter is still `any`, so a
// method of Seq[V any] can never mention Set[V] — the comparable constraint
// must live on the method's own type parameter, and that parameter cannot be
// inferred from zero arguments. The other collectors need no annotations.
func (seq Seq[V]) Collect() collector[V] { return collector[V]{seq} }

// Collect returns a collector over the key-value sequence. Collect itself is
// lazy and does not consume the sequence; only the collector's materializer
// methods do, turning the pairs into a concrete container:
//
//	m.Iter().FilterByKey(fn).Collect().Pairs()
//	m.Iter().FilterByKey(fn).Collect().Map[string, string]()
//	mo.Iter().Collect().MapOrd[string, string]()
func (seq Seq2[K, V]) Collect() collector2[K, V] { return collector2[K, V]{seq} }

// collector materializes a value sequence into containers; build one with
// Seq.Collect.
type collector[V any] struct{ seq Seq[V] }

// collector2 materializes a key-value sequence into containers; build one with
// Seq2.Collect.
type collector2[K, V any] struct{ seq Seq2[K, V] }

// Slice consumes the sequence and returns its elements as a Slice.
func (c collector[V]) Slice() Slice[V] { return c.seq.seqToSlice() }

// Deque consumes the sequence and returns its elements as a Deque, preserving
// order.
func (c collector[V]) Deque() *Deque[V] {
	result := NewDeque[V]()

	c.seq(func(v V) bool {
		result.PushBack(v)
		return true
	})

	return result
}

// Heap consumes the sequence and returns its elements as a Heap ordered by
// compareFn.
func (c collector[V]) Heap(compareFn func(V, V) cmp.Ordering) *Heap[V] {
	result := NewHeap(compareFn)

	c.seq(func(v V) bool {
		result.Push(v)
		return true
	})

	return result
}

// Set consumes the sequence and returns its elements as a Set, deduplicating
// them. The element type is passed explicitly — Collect().Set[int]() — because
// the sequence's own type parameter cannot carry the comparable constraint;
// each element is converted at runtime and a mismatched type argument panics.
func (c collector[V]) Set[W comparable]() Set[W] {
	collection := make(Set[W])

	c.seq(func(v V) bool {
		collection[any(v).(W)] = Unit{}
		return true
	})

	return collection
}

// Pairs consumes the sequence and returns its elements as plain pairs.
func (c collector2[K, V]) Pairs() []Pair[K, V] {
	var result []Pair[K, V]

	c.seq(func(k K, v V) bool {
		result = append(result, Pair[K, V]{Key: k, Value: v})
		return true
	})

	return result
}

// Map consumes the key-value sequence and returns it as a Map; later keys
// overwrite earlier ones. Both types are passed explicitly, mirroring the
// container's own signature — Collect().Map[string, int]() — because the
// sequence's key parameter cannot carry the comparable constraint; the pairs
// are converted at runtime and a mismatched type argument panics.
func (c collector2[K, V]) Map[K2 comparable, V2 any]() Map[K2, V2] {
	collection := NewMap[K2, V2]()

	c.seq(func(k K, v V) bool {
		collection[any(k).(K2)] = any(v).(V2)
		return true
	})

	return collection
}

// MapSafe consumes the key-value sequence and returns it as a thread-safe
// MapSafe; later keys overwrite earlier ones. Both types are passed explicitly
// — Collect().MapSafe[string, int]() — see [collector2.Map].
func (c collector2[K, V]) MapSafe[K2 comparable, V2 any]() *MapSafe[K2, V2] {
	collection := NewMapSafe[K2, V2]()

	c.seq(func(k K, v V) bool {
		collection.Insert(any(k).(K2), any(v).(V2))
		return true
	})

	return collection
}

// MapOrd consumes the key-value sequence and returns it as a MapOrd, keeping
// first-seen key order; a repeated key updates the value in place. Both types
// are passed explicitly — Collect().MapOrd[string, int]() — see
// [collector2.Map].
func (c collector2[K, V]) MapOrd[K2 comparable, V2 any]() MapOrd[K2, V2] {
	collection := NewMapOrd[K2, V2]()
	idx := make(map[K2]int)

	c.seq(func(k K, v V) bool {
		key := any(k).(K2)
		if i, ok := idx[key]; ok {
			collection[i].Value = any(v).(V2)
			return true
		}

		collection = append(collection, Pair[K2, V2]{Key: key, Value: any(v).(V2)})
		idx[key] = len(collection) - 1

		return true
	})

	return collection
}
