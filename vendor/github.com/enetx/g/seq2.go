package g

import (
	"context"
	"iter"
	"reflect"
	"slices"

	"github.com/enetx/g/cmp"
	"github.com/enetx/g/constraints"
)

// Seq2 is an iterator over sequences of ordered pairs of values, most commonly ordered key-value pairs.
type Seq2[K, V any] func(yield func(K, V) bool)

// Pull converts the “push-style” iterator sequence seq
// into a “pull-style” iterator accessed by the two functions
// next and stop.
//
// Next returns the next pair in the sequence
// and a boolean indicating whether the pair is valid.
// When the sequence is over, next returns a pair of zero values and false.
// It is valid to call next after reaching the end of the sequence
// or after calling stop. These calls will continue
// to return a pair of zero values and false.
//
// Stop ends the iteration. It must be called when the caller is
// no longer interested in next values and next has not yet
// signaled that the sequence is over (with a false boolean return).
// It is valid to call stop multiple times and when next has
// already returned false.
//
// It is an error to call next or stop from multiple goroutines
// simultaneously.
func (seq Seq2[K, V]) Pull() (func() (K, V, bool), func()) {
	return seq.seq2Pull()
}

// All checks whether all key-value pairs in the iterator satisfy the provided condition.
// This function is useful when you want to determine if all pairs in an iterator
// meet a specific criteria.
//
// Parameters:
// - fn (func(K, V) bool): A function that returns a boolean indicating whether the pair satisfies
// the condition.
//
// Returns:
// - bool: True if all pairs in the iterator satisfy the condition, false otherwise.
//
// Example usage:
//
//	m := g.NewMapOrd[g.String, g.Int]()
//	m.Insert("a", 1)
//	m.Insert("b", 2)
//	allPositive := m.Iter().All(func(_ g.String, v g.Int) bool { return v > 0 })
//
// The resulting allPositive will be true if all values returned by the iterator are positive.
func (seq Seq2[K, V]) All(fn func(K, V) bool) bool {
	all := true

	seq(func(k K, v V) bool {
		if !fn(k, v) {
			all = false
			return false
		}

		return true
	})

	return all
}

// Any checks whether any key-value pair in the iterator satisfies the provided condition.
// This function is useful when you want to determine if at least one pair in an iterator
// meets a specific criteria.
//
// Parameters:
// - fn (func(K, V) bool): A function that returns a boolean indicating whether the pair satisfies
// the condition.
//
// Returns:
// - bool: True if at least one pair in the iterator satisfies the condition, false otherwise.
//
// Example usage:
//
//	m := g.NewMapOrd[g.String, g.Int]()
//	m.Insert("a", 1)
//	m.Insert("b", 2)
//	anyEven := m.Iter().Any(func(_ g.String, v g.Int) bool { return v%2 == 0 })
//
// The resulting anyEven will be true if at least one value returned by the iterator is even.
func (seq Seq2[K, V]) Any(fn func(K, V) bool) bool {
	found := false

	seq(func(k K, v V) bool {
		if fn(k, v) {
			found = true
			return false
		}

		return true
	})

	return found
}

// Keys returns an iterator over the keys of the key-value pairs.
func (seq Seq2[K, V]) Keys() Seq[K] {
	return func(yield func(K) bool) { seq(func(k K, _ V) bool { return yield(k) }) }
}

// Values returns an iterator over the values of the key-value pairs.
func (seq Seq2[K, V]) Values() Seq[V] {
	return func(yield func(V) bool) { seq(func(_ K, v V) bool { return yield(v) }) }
}

// Unzip consumes the sequence and collects the keys and values
// of each pair into two separate slices.
func (seq Seq2[K, V]) Unzip() (Slice[K], Slice[V]) {
	var (
		keys   Slice[K]
		values Slice[V]
	)

	seq(func(k K, v V) bool {
		keys = append(keys, k)
		values = append(values, v)

		return true
	})

	return keys, values
}

// SortBy applies a custom sorting function to the elements in the iterator
// and returns a new iterator containing the sorted elements.
//
// The sorting function 'fn' should take two arguments, 'a' and 'b', of type Pair[K, V],
// and return a cmp.Ordering: cmp.Less orders 'a' before 'b', cmp.More after it.
//
// Example:
//
//	m := g.NewMapOrd[g.Int, g.String]()
//	m.Insert(6, "bb")
//	m.Insert(0, "dd")
//	m.Insert(1, "aa")
//	m.Insert(5, "xx")
//	m.Insert(2, "cc")
//	m.Insert(3, "ff")
//	m.Insert(4, "zz")
//
//	m.Iter().
//		SortBy(
//			func(a, b g.Pair[g.Int, g.String]) cmp.Ordering {
//				return a.Key.Cmp(b.Key)
//				// return a.Value.Cmp(b.Value)
//			}).
//		Collect().
//		MapOrd[g.Int, g.String]().
//		Print()
//
// Output: MapOrd{0:dd, 1:aa, 2:cc, 3:ff, 4:zz, 5:xx, 6:bb}
//
// The returned iterator is of type Seq2[K, V], which implements the iterator
// interface for further iteration over the sorted elements.
func (seq Seq2[K, V]) SortBy(fn func(a, b Pair[K, V]) cmp.Ordering) Seq2[K, V] {
	buf := seq.seq2ToPairs()
	slices.SortFunc(buf, func(a, b Pair[K, V]) int { return int(fn(a, b)) })

	return seqFromPairs(buf)
}

// SortByKey applies a custom sorting function to the keys in the iterator
// and returns a new iterator containing the sorted elements.
//
// The sorting function 'fn' should take two arguments, 'a' and 'b', of type K,
// and return a cmp.Ordering: cmp.Less orders 'a' before 'b', cmp.More after it.
//
// Example:
//
//	m := g.NewMapOrd[g.Int, g.String]()
//	m.Insert(6, "bb")
//	m.Insert(0, "dd")
//	m.Insert(1, "aa")
//	m.Insert(5, "xx")
//	m.Insert(2, "cc")
//	m.Insert(3, "ff")
//	m.Insert(4, "zz")
//
//	m.Iter().
//		SortByKey(g.Int.Cmp).
//		Collect().
//		MapOrd[g.Int, g.String]().
//		Print()
//
// Output: MapOrd{0:dd, 1:aa, 2:cc, 3:ff, 4:zz, 5:xx, 6:bb}
func (seq Seq2[K, V]) SortByKey(fn func(a, b K) cmp.Ordering) Seq2[K, V] {
	buf := seq.seq2ToPairs()
	slices.SortFunc(buf, func(a, b Pair[K, V]) int { return int(fn(a.Key, b.Key)) })

	return seqFromPairs(buf)
}

// SortByValue applies a custom sorting function to the values in the iterator
// and returns a new iterator containing the sorted elements.
//
// The sorting function 'fn' should take two arguments, 'a' and 'b', of type V,
// and return a cmp.Ordering: cmp.Less orders 'a' before 'b', cmp.More after it.
//
// Example:
//
//	m := g.NewMapOrd[g.Int, g.String]()
//	m.Insert(6, "bb")
//	m.Insert(0, "dd")
//	m.Insert(1, "aa")
//	m.Insert(5, "xx")
//	m.Insert(2, "cc")
//	m.Insert(3, "ff")
//	m.Insert(4, "zz")
//
//	m.Iter().
//		SortByValue(g.String.Cmp).
//		Collect().
//		MapOrd[g.Int, g.String]().
//		Print()
//
// Output: MapOrd{1:aa, 6:bb, 2:cc, 0:dd, 3:ff, 5:xx, 4:zz}
func (seq Seq2[K, V]) SortByValue(fn func(a, b V) cmp.Ordering) Seq2[K, V] {
	buf := seq.seq2ToPairs()
	slices.SortFunc(buf, func(a, b Pair[K, V]) int { return int(fn(a.Value, b.Value)) })

	return seqFromPairs(buf)
}

// Inspect creates a new iterator that wraps around the current iterator
// and allows inspecting each key-value pair as it passes through.
func (seq Seq2[K, V]) Inspect(fn func(k K, v V)) Seq2[K, V] {
	return func(yield func(K, V) bool) { seq(func(k K, v V) bool { fn(k, v); return yield(k, v) }) }
}

// StepBy creates a new iterator that iterates over every N-th element of the original iterator.
// This function is useful when you want to skip a specific number of elements between each iteration.
//
// Parameters:
// - n Int: The step size, indicating how many elements to skip between each iteration.
//
// Returns:
// - Seq2[K, V]: A new iterator that produces key-value pairs from the original iterator with a step size of N.
//
// Example usage:
//
//	mapIter := g.MapOrd[string, int]{{"one", 1}, {"two", 2}, {"three", 3}}.Iter()
//	iter := mapIter.StepBy(2)
//	iter.Collect().MapOrd[string, int]().Print()
//
// Output: MapOrd{one:1, three:3}
//
// The resulting iterator will produce key-value pairs from the original iterator with a step size of N.
func (seq Seq2[K, V]) StepBy(n Int) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if n <= 0 {
			return
		}

		index := Int(0)

		seq(func(k K, v V) bool {
			if index%n == 0 {
				if !yield(k, v) {
					return false
				}
			}

			index++

			return true
		})
	}
}

// Chain concatenates the current iterator with other iterators, returning a new iterator.
//
// The function creates a new iterator that combines the elements of the current iterator
// with elements from the provided iterators in the order they are given.
//
// Params:
//
// - seqs ([]Seq2[K, V]): Other iterators to be concatenated with the current iterator.
//
// Returns:
//
// - Seq2[K, V]: A new iterator containing elements from the current iterator and the provided iterators.
//
// Example usage:
//
//	m1 := g.NewMapOrd[int, string]()
//	m1.Insert(1, "a")
//
//	m2 := g.NewMapOrd[int, string]()
//	m2.Insert(2, "b")
//
//	// Concatenating iterators and collecting the result.
//	m1.Iter().Chain(m2.Iter()).Collect().MapOrd[int, string]().Print()
//
// Output: MapOrd{1:a, 2:b}
//
// The resulting iterator will contain elements from both iterators in the specified order.
func (seq Seq2[K, V]) Chain(seqs ...Seq2[K, V]) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		proceed := true

		seq(func(k K, v V) bool {
			if !yield(k, v) {
				proceed = false
				return false
			}

			return true
		})

		if !proceed {
			return
		}

		for _, rest := range seqs {
			rest(func(k K, v V) bool {
				if !yield(k, v) {
					proceed = false
					return false
				}

				return true
			})

			if !proceed {
				return
			}
		}
	}
}

// Count consumes the iterator, counting the number of iterations and returning it.
func (seq Seq2[K, V]) Count() Int {
	count := Int(0)
	seq(func(K, V) bool { count++; return true })

	return count
}

// Skip returns a new iterator skipping the first n elements.
//
// The function creates a new iterator that skips the first n elements of the current iterator
// and returns an iterator starting from the (n+1)th element.
//
// Params:
//
// - n (Int): The number of elements to skip from the beginning of the iterator.
// Negative values are treated as zero.
//
// Returns:
//
// - Seq2[K, V]: An iterator that starts after skipping the first n elements.
//
// Example usage:
//
//	m := g.NewMapOrd[int, string]()
//	m.Insert(1, "a")
//	m.Insert(2, "b")
//	m.Insert(3, "c")
//	m.Insert(4, "d")
//
//	// Skipping the first two elements and collecting the rest.
//	m.Iter().Skip(2).Collect().MapOrd[int, string]().Print()
//
// Output: MapOrd{3:c, 4:d}
//
// The resulting iterator will start after skipping the specified number of elements.
func (seq Seq2[K, V]) Skip(n Int) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if n <= 0 {
			seq(yield)
			return
		}

		count := Int(0)

		seq(func(k K, v V) bool {
			if count < n {
				count++
				return true
			}

			return yield(k, v)
		})
	}
}

// Exclude returns a new iterator excluding elements that satisfy the provided function.
//
// The function creates a new iterator excluding elements from the current iterator
// for which the provided function returns true.
//
// Params:
//
// - fn (func(K, V) bool): The function used to determine exclusion criteria for elements.
//
// Returns:
//
// - Seq2[K, V]: A new iterator excluding elements that satisfy the given condition.
//
// Example usage:
//
//	mo := g.NewMapOrd[int, int]()
//	mo.Insert(1, 1)
//	mo.Insert(2, 2)
//	mo.Insert(3, 3)
//	mo.Insert(4, 4)
//	mo.Insert(5, 5)
//
//	notEven := mo.Iter().
//		Exclude(
//			func(k, v int) bool {
//				return v%2 == 0
//			}).
//		Collect().MapOrd[int, int]()
//	notEven.Print()
//
// Output: MapOrd{1:1, 3:3, 5:5}
//
// The resulting iterator will exclude elements based on the provided condition.
func (seq Seq2[K, V]) Exclude(fn func(K, V) bool) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		seq(func(k K, v V) bool {
			if !fn(k, v) {
				return yield(k, v)
			}

			return true
		})
	}
}

// Filter returns a new iterator containing only the elements that satisfy the provided function.
//
// The function creates a new iterator including elements from the current iterator
// for which the provided function returns true.
//
// Params:
//
// - fn (func(K, V) bool): The function used to determine inclusion criteria for elements.
//
// Returns:
//
// - Seq2[K, V]: A new iterator containing elements that satisfy the given condition.
//
// Example usage:
//
//	mo := g.NewMapOrd[int, int]()
//	mo.Insert(1, 1)
//	mo.Insert(2, 2)
//	mo.Insert(3, 3)
//	mo.Insert(4, 4)
//	mo.Insert(5, 5)
//
//	even := mo.Iter().
//		Filter(
//			func(k, v int) bool {
//				return v%2 == 0
//			}).
//		Collect().MapOrd[int, int]()
//	even.Print()
//
// Output: MapOrd{2:2, 4:4}
//
// The resulting iterator will include elements based on the provided condition.
func (seq Seq2[K, V]) Filter(fn func(K, V) bool) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		seq(func(k K, v V) bool {
			if fn(k, v) {
				return yield(k, v)
			}

			return true
		})
	}
}

// FilterByKey returns a new iterator lazily yielding only the pairs whose key
// satisfies the provided predicate; values are not inspected.
//
// It lifts a single-parameter predicate to the pair-wise Filter — composes
// with f.* factories:
//
//	mo.Iter().FilterByKey(f.Eq("host"))
func (seq Seq2[K, V]) FilterByKey(fn func(K) bool) Seq2[K, V] {
	return seq.Filter(func(k K, _ V) bool { return fn(k) })
}

// FilterByValue returns a new iterator lazily yielding only the pairs whose
// value satisfies the provided predicate; keys are not inspected.
//
// It lifts a single-parameter predicate to the pair-wise Filter — composes
// with f.* factories:
//
//	mo.Iter().FilterByValue(f.Gt(10))
func (seq Seq2[K, V]) FilterByValue(fn func(V) bool) Seq2[K, V] {
	return seq.Filter(func(_ K, v V) bool { return fn(v) })
}

// Find searches for an element in the iterator that satisfies the provided function.
//
// The function iterates through the elements of the iterator and returns the first element
// for which the provided function returns true.
//
// Params:
//
// - fn (func(K, V) bool): The function used to test elements for a condition.
//
// Returns:
//
// - Option[Pair[K, V]]: An Option containing the first pair that satisfies the condition; None if not found.
//
// Example usage:
//
//	m := g.NewMapOrd[int, int]()
//	m.Insert(1, 1)
//	f := m.Iter().Find(func(_ int, v int) bool { return v == 1 })
//	if f.IsSome() {
//		print(f.Some().Key)
//	}
//
// The resulting Option may contain the first element that satisfies the condition, or None if not found.
func (seq Seq2[K, V]) Find(fn func(k K, v V) bool) Option[Pair[K, V]] {
	var result Option[Pair[K, V]]

	seq(func(k K, v V) bool {
		if fn(k, v) {
			result = Some(Pair[K, V]{Key: k, Value: v})
			return false
		}

		return true
	})

	return result
}

// ForEach iterates through all elements and applies the given function to each key-value pair.
//
// The function applies the provided function to each key-value pair in the iterator.
//
// Params:
//
// - fn (func(K, V)): The function to be applied to each key-value pair in the iterator.
//
// Example usage:
//
//	m := g.NewMapOrd[int, int]()
//	m.Insert(1, 1)
//	m.Insert(2, 2)
//	m.Insert(3, 3)
//	m.Insert(4, 4)
//	m.Insert(5, 5)
//
//	m.Iter().ForEach(func(key, val int) {
//	    // Process key-value pair
//	})
//
// The provided function will be applied to each key-value pair in the iterator.
func (seq Seq2[K, V]) ForEach(fn func(k K, v V)) {
	seq(func(k K, v V) bool { fn(k, v); return true })
}

// Map creates a new iterator by applying the given function to each key-value pair.
//
// The function creates a new iterator by applying the provided function to each key-value pair in the iterator.
//
// Params:
//
// - transform (func(K, V) (K2, V2)): The function used to transform each key-value pair
// in the iterator. The key and value types of the result may differ.
//
// Returns:
//
// - Seq2[K2, V2]: A new iterator containing transformed key-value pairs.
//
// Example usage:
//
//	mo := g.NewMapOrd[int, int]()
//	mo.Insert(1, 1)
//	mo.Insert(2, 2)
//	mo.Insert(3, 3)
//	mo.Insert(4, 4)
//	mo.Insert(5, 5)
//
//	momap := mo.Iter().
//		Map(
//			func(k, v int) (int, int) {
//				return k * k, v * v
//			}).
//		Collect().MapOrd[int, int]()
//
//	momap.Print()
//
// Output: MapOrd{1:1, 4:4, 9:9, 16:16, 25:25}
//
// The resulting iterator will contain transformed key-value pairs.
func (seq Seq2[K, V]) Map[K2, V2 any](transform func(K, V) (K2, V2)) Seq2[K2, V2] {
	return func(yield func(K2, V2) bool) {
		seq(func(k K, v V) bool { return yield(transform(k, v)) })
	}
}

// FilterMap applies a function to each key-value pair and filters out None results.
//
// Pairs where the function returns None are filtered out; pairs where it returns
// Some(Pair) are transformed and included in the result. Key and value types may differ
// from the input types.
func (seq Seq2[K, V]) FilterMap[K2, V2 any](fn func(K, V) Option[Pair[K2, V2]]) Seq2[K2, V2] {
	return func(yield func(K2, V2) bool) {
		seq(func(k K, v V) bool {
			if pair, ok := fn(k, v).Option(); ok {
				return yield(pair.Unpack())
			}

			return true
		})
	}
}

// Range iterates through elements until the given function returns false.
//
// The function iterates through the key-value pairs in the iterator, applying the provided function to each pair.
// It continues iterating until the function returns false.
//
// Params:
//
// - fn (func(K, V) bool): The function to be applied to each key-value pair in the iterator.
//
// Example usage:
//
//	m := g.NewMapOrd[int, int]()
//	m.Insert(1, 1)
//	m.Insert(2, 2)
//	m.Insert(3, 3)
//	m.Insert(4, 4)
//	m.Insert(5, 5)
//
//	m.Iter().Range(func(k, v int) bool {
//	    fmt.Println(v) // Replace this with the function logic you need.
//	    return v < 5 // Replace this with the condition for continuing iteration.
//	})
//
// The iteration will stop when the provided function returns false.
func (seq Seq2[K, V]) Range(fn func(k K, v V) bool) {
	seq(fn)
}

// Context allows the iteration to be controlled with a context.Context.
func (seq Seq2[K, V]) Context(ctx context.Context) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if err := ctx.Err(); err != nil {
			return
		}

		seq(func(k K, v V) bool {
			select {
			case <-ctx.Done():
				return false
			default:
				return yield(k, v)
			}
		})
	}
}

// Take returns a new iterator with the first n elements.
// The function creates a new iterator containing the first n elements from the original iterator.
func (seq Seq2[K, V]) Take(n Int) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		if n <= 0 {
			return
		}

		count := Int(0)

		seq(func(k K, v V) bool {
			if count >= n {
				return false
			}

			count++

			return yield(k, v)
		})
	}
}

// First returns the first key-value pair from the sequence.
func (seq Seq2[K, V]) First() Option[Pair[K, V]] {
	var result Option[Pair[K, V]]

	seq(func(k K, v V) bool {
		result = Some(Pair[K, V]{Key: k, Value: v})
		return false
	})

	return result
}

// Last returns the last key-value pair from the sequence.
func (seq Seq2[K, V]) Last() Option[Pair[K, V]] {
	var result Option[Pair[K, V]]

	seq(func(k K, v V) bool {
		result = Some(Pair[K, V]{Key: k, Value: v})
		return true
	})

	return result
}

// Nth returns the nth key-value pair (0-indexed) in the sequence.
func (seq Seq2[K, V]) Nth(n Int) Option[Pair[K, V]] {
	var result Option[Pair[K, V]]

	index := Int(0)

	seq(func(k K, v V) bool {
		if index == n {
			result = Some(Pair[K, V]{Key: k, Value: v})
			return false
		}

		index++

		return true
	})

	return result
}

// Chan converts the iterator into a channel, optionally with context(s).
//
// The function converts the key-value pairs from the iterator into a channel, allowing iterative processing
// using channels. It can be used to stream key-value pairs for concurrent or asynchronous operations.
//
// Params:
//
// - ctxs (...context.Context): Optional context(s) that can be used to cancel or set deadlines for the operation.
//
// Returns:
//
// - chan Pair[K, V]: A channel emitting key-value pairs from the iterator.
//
// Example usage:
//
//	m := g.NewMapOrd[int, int]()
//	m.Insert(1, 1)
//	m.Insert(2, 2)
//	m.Insert(3, 3)
//	m.Insert(4, 4)
//	m.Insert(5, 5)
//
//	ctx, cancel := context.WithCancel(context.Background())
//	defer cancel() // Ensure cancellation to avoid goroutine leaks.
//
//	ch := m.Iter().Chan(ctx)
//	for pair := range ch {
//	    // Process key-value pair from the channel
//	}
//
// The function converts the iterator into a channel to allow sequential or concurrent processing of key-value pairs.
func (seq Seq2[K, V]) Chan(ctxs ...context.Context) chan Pair[K, V] {
	ctx := context.Background()
	if len(ctxs) > 0 {
		ctx = ctxs[0]
	}

	ch := make(chan Pair[K, V])

	go func() {
		defer close(ch)

		if err := ctx.Err(); err != nil {
			return
		}

		seq(func(k K, v V) bool {
			select {
			case <-ctx.Done():
				return false
			case ch <- Pair[K, V]{k, v}:
				return true
			}
		})
	}()

	return ch
}

// Next extracts the next key-value pair from the iterator and advances it.
//
// This method consumes the next key-value pair from the iterator and returns them wrapped in an Option.
// The iterator itself is modified to point to the remaining elements.
//
// Returns:
// - Option[Pair[K, V]]: Some(Pair{Key, Value}) if a pair exists, None if the iterator is exhausted.
func (seq *Seq2[K, V]) Next() Option[Pair[K, V]] {
	if key, value, remaining, ok := (*seq).seq2Next(); ok {
		*seq = Seq2[K, V](remaining)
		return Some(Pair[K, V]{Key: key, Value: value})
	}

	return None[Pair[K, V]]()
}

// Fold reduces the sequence of key-value pairs to a single value using an accumulator.
// The accumulator type may differ from the key and value types.
func (seq Seq2[K, V]) Fold[A any](init A, fn func(acc A, k K, v V) A) A {
	seq(func(k K, v V) bool { init = fn(init, k, v); return true })
	return init
}

// SumBy maps each key-value pair to a numeric value via fn and returns the sum of those values,
// visiting pairs in insertion order. An empty sequence yields the zero value of S.
func (seq Seq2[K, V]) SumBy[S constraints.Number](fn func(K, V) S) S {
	var zero S
	return seq.Fold(zero, func(acc S, k K, v V) S { return acc + fn(k, v) })
}

// ProductBy maps each key-value pair to a numeric value via fn and returns their
// product, in insertion order. An empty sequence yields the multiplicative
// identity, one.
func (seq Seq2[K, V]) ProductBy[S constraints.Number](fn func(K, V) S) S {
	return seq.Fold(S(1), func(acc S, k K, v V) S { return acc * fn(k, v) })
}

// FindMap applies fn to each key-value pair in insertion order and returns the
// first Some result, or None if fn returns None for every pair.
func (seq Seq2[K, V]) FindMap[U any](fn func(K, V) Option[U]) Option[U] {
	var result Option[U]

	seq(func(k K, v V) bool {
		if o := fn(k, v); o.IsSome() {
			result = o
			return false
		}

		return true
	})

	return result
}

// TryMap applies a fallible transform to each key-value pair and enters the
// Result pipeline, producing a SeqResult[U]. See [Seq.TryMap] for the full
// contract.
func (seq Seq2[K, V]) TryMap[U any](fn func(K, V) Result[U]) SeqResult[U] {
	return func(yield func(Result[U]) bool) {
		seq(func(k K, v V) bool {
			return yield(fn(k, v))
		})
	}
}

// TakeWhile yields key-value pairs while the predicate returns true, stopping at the first false.
func (seq Seq2[K, V]) TakeWhile(fn func(K, V) bool) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		seq(func(k K, v V) bool {
			if !fn(k, v) {
				return false
			}

			return yield(k, v)
		})
	}
}

// SkipWhile skips key-value pairs while the predicate returns true, then yields the rest.
func (seq Seq2[K, V]) SkipWhile(fn func(K, V) bool) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		skipping := true

		seq(func(k K, v V) bool {
			if skipping && fn(k, v) {
				return true
			}

			skipping = false

			return yield(k, v)
		})
	}
}

// Dedup returns a new iterator that removes consecutive pairs with duplicate keys,
// keeping the first pair of each run; values are not inspected. If the sequence is
// sorted by key, all keys will be unique. It mirrors [Seq.Dedup], comparing keys
// with == when K is a comparable type and falling back to reflect.DeepEqual
// otherwise.
//
// Example usage:
//
//	mo := g.MapOrd[g.Int, g.String]{{1, "a"}, {1, "b"}, {2, "c"}}
//	mo.Iter().Dedup().Collect().MapOrd[g.Int, g.String]().Print()
//
// Output: MapOrd{1:a, 2:c}
func (seq Seq2[K, V]) Dedup() Seq2[K, V] {
	eq := func(a, b K) bool { return reflect.DeepEqual(a, b) }
	if isValueComparable[K]() {
		eq = func(a, b K) bool { return any(a) == any(b) }
	}

	return func(yield func(K, V) bool) {
		var prev K

		first := true

		seq(func(k K, v V) bool {
			if first || !eq(prev, k) {
				prev = k
				first = false

				return yield(k, v)
			}

			return true
		})
	}
}

// Unique returns a new iterator yielding only the first pair for each distinct key;
// later pairs with an already-seen key are dropped and values are not inspected.
// It mirrors [Seq.Unique]. The keys are tracked in a map, so a key type that is
// not hashable at runtime (e.g. a slice, map, function, or a struct containing
// one) panics with "hash of unhashable type".
//
// Example usage:
//
//	mo := g.MapOrd[g.Int, g.String]{{1, "a"}, {2, "b"}, {1, "c"}}
//	mo.Iter().Unique().Collect().MapOrd[g.Int, g.String]().Print()
//
// Output: MapOrd{1:a, 2:b}
func (seq Seq2[K, V]) Unique() Seq2[K, V] {
	return func(yield func(K, V) bool) {
		seen := make(map[any]struct{})

		seq(func(k K, v V) bool {
			key := any(k)
			if _, exists := seen[key]; !exists {
				seen[key] = struct{}{}
				return yield(k, v)
			}

			return true
		})
	}
}

// MaxBy returns the maximum pair in the sequence using the provided comparison
// function, mirroring [Seq.MaxBy]. It returns None if the sequence is empty.
func (seq Seq2[K, V]) MaxBy(fn func(a, b Pair[K, V]) cmp.Ordering) Option[Pair[K, V]] {
	var max Pair[K, V]

	found := false

	seq(func(k K, v V) bool {
		pair := Pair[K, V]{Key: k, Value: v}
		if !found || fn(max, pair) == cmp.Less {
			max = pair
			found = true
		}

		return true
	})

	return OptionOf(max, found)
}

// MinBy returns the minimum pair in the sequence using the provided comparison
// function, mirroring [Seq.MinBy]. It returns None if the sequence is empty.
func (seq Seq2[K, V]) MinBy(fn func(a, b Pair[K, V]) cmp.Ordering) Option[Pair[K, V]] {
	var min Pair[K, V]

	found := false

	seq(func(k K, v V) bool {
		pair := Pair[K, V]{Key: k, Value: v}
		if !found || fn(pair, min) == cmp.Less {
			min = pair
			found = true
		}

		return true
	})

	return OptionOf(min, found)
}

// Reduce aggregates the pairs of the sequence using the provided function,
// mirroring [Seq.Reduce]. The first pair is used as the initial accumulator
// value. It returns None if the sequence is empty.
func (seq Seq2[K, V]) Reduce(fn func(a, b Pair[K, V]) Pair[K, V]) Option[Pair[K, V]] {
	var acc Pair[K, V]

	first := true

	seq(func(k K, v V) bool {
		pair := Pair[K, V]{Key: k, Value: v}
		if first {
			acc = pair
			first = false
		} else {
			acc = fn(acc, pair)
		}

		return true
	})

	return OptionOf(acc, !first)
}

// Scan accumulates the pairs of the sequence using a function, yielding the
// initial value followed by each intermediate accumulator state, mirroring
// [Seq.Scan]. The accumulator type may differ from the key and value types.
//
// Example usage:
//
//	mo := g.MapOrd[g.String, g.Int]{{"a", 1}, {"b", 2}, {"c", 3}}
//	sums := mo.Iter().Scan(0, func(acc int, _ g.String, v g.Int) int {
//		return acc + int(v)
//	})
//	// sums will yield: 0, 1, 3, 6
func (seq Seq2[K, V]) Scan[A any](init A, fn func(acc A, k K, v V) A) Seq[A] {
	return func(yield func(A) bool) {
		if !yield(init) {
			return
		}

		acc := init

		seq(func(k K, v V) bool {
			acc = fn(acc, k, v)
			return yield(acc)
		})
	}
}

// Intersperse inserts the provided separator pair between each consecutive pair
// of elements in the original iterator, mirroring [Seq.Intersperse].
func (seq Seq2[K, V]) Intersperse(sep Pair[K, V]) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		first := true

		seq(func(k K, v V) bool {
			if !first && !yield(sep.Unpack()) {
				return false
			}

			first = false

			return yield(k, v)
		})
	}
}

// MapTo transforms each pair into a single value using the given function,
// returning a value sequence of the results. It complements [Seq2.Map], which
// keeps the pair shape.
func (seq Seq2[K, V]) MapTo[T any](fn func(K, V) T) Seq[T] {
	return func(yield func(T) bool) {
		seq(func(k K, v V) bool { return yield(fn(k, v)) })
	}
}

// ── iterator core (key-value sequences), ported from github.com/enetx/iter (MIT) ──

// seq2Next extracts the first key-value pair from the sequence and returns the remaining sequence.
// Returns (key, value, remainingSeq, true) if a pair exists, or (zeroK, zeroV, nil, false) if empty.
//
// Example:
//
//	s := Map[int, string]{1: "a", 2: "b", 3: "c"}.Iter()
//	k, v, rest, ok := s.seq2Next()
//	// k = 1, v = "a", ok = true (order not guaranteed for maps)
//	// rest yields remaining pairs
//
//	k2, v2, rest2, ok2 := rest.seq2Next()
//	// k2 = 2, v2 = "b", ok2 = true
//	// rest2 yields remaining pairs
func (seq Seq2[K, V]) seq2Next() (K, V, Seq2[K, V], bool) {
	next, stop := seq.seq2Pull()

	firstK, firstV, ok := next()
	if !ok {
		stop()
		var zeroK K
		var zeroV V
		return zeroK, zeroV, nil, false
	}

	// The remaining sequence continues from the same pull iterator, so the source
	// is walked exactly once. This makes Next O(1) per element and correct for
	// non-deterministic sources (e.g. maps), at the cost of the remaining
	// sequence being single-use.
	remaining := func(yield func(K, V) bool) {
		defer stop()
		for {
			k, v, ok := next()
			if !ok {
				return
			}
			if !yield(k, v) {
				return
			}
		}
	}

	return firstK, firstV, remaining, true
}

// seq2Pull converts a push-style iterator (Seq2) to a pull-style iterator.
func (seq Seq2[K, V]) seq2Pull() (next func() (K, V, bool), stop func()) {
	return iter.Pull2(iter.Seq2[K, V](seq))
}

// seq2ToPairs collects all key-value pairs into a slice of Pair structs.
func (seq Seq2[K, V]) seq2ToPairs() []Pair[K, V] {
	out := make([]Pair[K, V], 0)
	seq(func(k K, v V) bool {
		out = append(out, Pair[K, V]{k, v})
		return true
	})
	return out
}

// seqFromPairs creates a Seq2 from a slice of key-value pairs.
//
// Example:
//
//	pairs := []Pair[int, string]{{1, "a"}, {2, "b"}}
//	s := seqFromPairs(pairs)
func seqFromPairs[K, V any](pairs []Pair[K, V]) Seq2[K, V] {
	return func(yield func(K, V) bool) {
		for _, p := range pairs {
			if !yield(p.Unpack()) {
				return
			}
		}
	}
}
