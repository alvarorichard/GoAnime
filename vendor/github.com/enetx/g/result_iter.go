package g

import (
	"context"
	"reflect"

	"github.com/enetx/g/cmp"
	"github.com/enetx/g/constraints"
)

// SeqResult is an iterator over sequences of Result[V] values.
type SeqResult[V any] func(yield func(Result[V]) bool)

// OkSeq wraps a single Ok value into a SeqResult iterator.
// Useful for returning a successful single-element sequence without manually
// constructing the yield function.
//
// Example:
//
//	return OkSeq(42)
func OkSeq[V any](v V) SeqResult[V] {
	return func(yield func(Result[V]) bool) { yield(Ok(v)) }
}

// ErrSeq wraps a single error into a SeqResult iterator.
// Useful for early returns in functions that produce a SeqResult,
// where a plain error needs to be lifted into the sequence type.
//
// Example:
//
//	return ErrSeq[int](errors.New("something went wrong"))
func ErrSeq[V any](err error) SeqResult[V] {
	return func(yield func(Result[V]) bool) { yield(Err[V](err)) }
}

// FlatMap transforms each Ok value into a sequence and flattens the results,
// wrapping each produced element in Ok. The element type may differ from the
// input type. If an Err is encountered, it is passed downstream as-is;
// iteration continues for as long as the consumer keeps accepting values
// (consumer-driven), matching Map.
//
// Example:
//
//	seq.FlatMap(Slice[Int].Iter) // SeqResult[Slice[Int]] -> SeqResult[Int]
func (seq SeqResult[V]) FlatMap[U any](fn func(V) Seq[U]) SeqResult[U] {
	return func(yield func(Result[U]) bool) {
		seq(func(v Result[V]) bool {
			if v.IsErr() {
				return yield(Err[U](v.err))
			}

			cont := true
			fn(v.v)(func(u U) bool {
				if !yield(Ok(u)) {
					cont = false
					return false
				}
				return true
			})

			return cont
		})
	}
}

// Pull converts the “push-style” sequence of Result[V] into a “pull-style” iterator accessed by two functions: next and stop.
//
// The next function returns the next Result[V] in the sequence and a boolean indicating whether the value is valid.
// When the sequence is over, next returns the zero value and false. It is valid to call next after reaching the end
// of the sequence or after calling stop. These calls will continue to return the zero value and false.
//
// The stop function ends the iteration. It must be called when the caller is no longer interested in next values and
// next has not yet signaled that the sequence is over. It is valid to call stop multiple times and after next has
// already returned false.
//
// It is an error to call next or stop from multiple goroutines simultaneously.
func (seq SeqResult[V]) Pull() (func() (Result[V], bool), func()) {
	return Seq[Result[V]](seq).seqPull()
}

// All checks whether all Ok values in the sequence satisfy the provided condition.
//
// If an Err is encountered in the sequence, that Err is immediately returned.
// Otherwise, it returns Ok(true) if all Ok values satisfy the function, or Ok(false) if at least one does not.
func (seq SeqResult[V]) All(fn func(v V) bool) Result[bool] {
	result := Ok(true)

	seq(func(v Result[V]) bool {
		if v.IsErr() {
			result = Err[bool](v.err)
			return false
		}

		if !fn(v.v) {
			result = Ok(false)
			return false
		}

		return true
	})

	return result
}

// Any checks whether any Ok value in the sequence satisfies the provided condition.
//
// If an Err is encountered, that Err is immediately returned.
// Otherwise, it returns Ok(true) if at least one Ok value satisfies the function, or Ok(false) if none do.
func (seq SeqResult[V]) Any(fn func(v V) bool) Result[bool] {
	result := Ok(false)

	seq(func(v Result[V]) bool {
		if v.IsErr() {
			result = Err[bool](v.err)
			return false
		}

		if fn(v.v) {
			result = Ok(true)
			return false
		}

		return true
	})

	return result
}

// Collect returns a collector over the raw Result elements, matching the
// Seq/Seq2 collector idiom: materialize with .Slice() (etc.). Collect itself
// is lazy and does not consume the sequence; the materializer does. Both Ok
// and Err elements flow through as-is; encountering an Err does not stop
// collection. Use TryCollect for the short-circuiting Ok-only variant.
func (seq SeqResult[V]) Collect() collector[Result[V]] {
	return collector[Result[V]]{Seq[Result[V]](seq)}
}

// TryCollect gathers the Ok values from the sequence into a Slice: the first Err
// short-circuits — iteration stops immediately, elements after it are not
// consumed — and that error is returned as Err. An empty sequence yields Ok of
// an empty Slice.
//
// Unlike Collect, which gathers every element (Ok and Err alike) into a
// Slice[Result[V]], TryCollect returns Result[Slice[V]]: either all the
// unwrapped Ok values, or the first error encountered.
func (seq SeqResult[V]) TryCollect() Result[Slice[V]] {
	collection := NewSlice[V]()

	var err error

	seq(func(v Result[V]) bool {
		if v.IsErr() {
			err = v.err
			return false
		}

		collection = append(collection, v.v)

		return true
	})

	if err != nil {
		return Err[Slice[V]](err)
	}

	return Ok(collection)
}

// Count consumes the entire sequence, counting the number of elements it yields.
// Err elements are counted like Ok elements and do not stop the count.
func (seq SeqResult[V]) Count() Int {
	var counter Int
	seq(func(Result[V]) bool {
		counter++
		return true
	})

	return counter
}

// Map transforms each Ok value in the sequence using the given function, returning a new sequence of Result.
// The result type may differ from the input type.
//
// If an Err is encountered, it is passed downstream as-is; iteration continues
// for as long as the consumer keeps accepting values (consumer-driven).
func (seq SeqResult[V]) Map[U any](transform func(V) U) SeqResult[U] {
	return func(yield func(Result[U]) bool) {
		seq(func(v Result[V]) bool {
			if v.IsErr() {
				return yield(Err[U](v.err))
			}
			return yield(Ok(transform(v.v)))
		})
	}
}

// Filter returns a new sequence containing only the Ok elements that satisfy the provided function.
//
// If an Err is encountered, it is yielded downstream as-is; the consumer decides
// whether to continue (consumer-driven).
// Only Ok elements for which fn returns true are yielded downstream as Ok.
func (seq SeqResult[V]) Filter(fn func(V) bool) SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		seq(func(v Result[V]) bool {
			if v.IsErr() {
				return yield(v)
			}
			if fn(v.v) {
				return yield(v)
			}
			return true
		})
	}
}

// Exclude returns a new sequence that excludes Ok elements which satisfy the provided function.
//
// If an Err is encountered, it is yielded downstream as-is (consumer-driven).
// Only Ok elements for which 'fn' returns false are yielded downstream.
func (seq SeqResult[V]) Exclude(fn func(V) bool) SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		seq(func(v Result[V]) bool {
			if v.IsErr() {
				return yield(v)
			}
			if !fn(v.v) {
				return yield(v)
			}
			return true
		})
	}
}

// Dedup removes consecutive duplicates of Ok values from the sequence, returning a new sequence.
//
// If an Err is encountered, it is yielded downstream as-is (consumer-driven).
// Consecutive Ok duplicates (based on equality) are filtered out so only the first occurrence is yielded.
func (seq SeqResult[V]) Dedup() SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		var current V
		hasFirst := false
		comparable := isValueComparable[V]()

		seq(func(v Result[V]) bool {
			if v.IsErr() {
				return yield(v)
			}

			if !hasFirst {
				hasFirst = true
				current = v.v
				return yield(v)
			}

			if comparable {
				if any(current) == any(v.v) {
					return true
				}
			} else {
				if reflect.DeepEqual(current, v.v) {
					return true
				}
			}

			current = v.v
			return yield(v)
		})
	}
}

// Unique returns a new sequence that contains only the first occurrence of each distinct Ok value.
//
// If an Err is encountered, it is yielded downstream as-is (consumer-driven).
// Future occurrences of a previously seen Ok value are skipped.
func (seq SeqResult[V]) Unique() SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		if isValueComparable[V]() {
			seen := NewSet[any]()

			seq(func(v Result[V]) bool {
				if v.IsErr() {
					return yield(v)
				}

				k := any(v.v)
				if _, ok := seen[k]; !ok {
					seen[k] = Unit{}
					return yield(v)
				}

				return true
			})
		} else {
			var seen Slice[V]

			seq(func(v Result[V]) bool {
				if v.IsErr() {
					return yield(v)
				}

				for _, s := range seen {
					if reflect.DeepEqual(s, v.v) {
						return true
					}
				}

				seen = append(seen, v.v)
				return yield(v)
			})
		}
	}
}

// ForEach applies a function to each Result in the sequence (Ok or Err) without modifying the sequence.
//
// The iteration continues over all elements, passing them to fn for side effects.
func (seq SeqResult[V]) ForEach(fn func(v Result[V])) {
	seq(func(v Result[V]) bool {
		fn(v)
		return true
	})
}

// Range iterates through elements until the given function returns false.
//
// For each element (Ok or Err), fn is called. If fn returns false, iteration stops immediately.
func (seq SeqResult[V]) Range(fn func(v Result[V]) bool) {
	seq(func(v Result[V]) bool {
		return fn(v)
	})
}

// Skip returns a new sequence that skips the first n Ok elements.
//
// If an Err is encountered, it is yielded as-is without consuming the skip
// budget (consumer-driven). Once n Ok elements have been skipped,
// subsequent elements (Ok or Err) are yielded normally.
func (seq SeqResult[V]) Skip(n Int) SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		if n < 0 {
			n = 0
		}

		seq(func(v Result[V]) bool {
			if v.IsErr() {
				return yield(v)
			}
			if n > 0 {
				n--
				return true
			}
			return yield(v)
		})
	}
}

// StepBy creates a new sequence that yields every nth Ok element from the original sequence.
//
// If an Err is encountered, it is yielded downstream as-is (consumer-driven).
// For Ok elements, only every n-th element is yielded.
func (seq SeqResult[V]) StepBy(n Int) SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		if n <= 0 {
			return
		}

		i := Int(0)
		seq(func(v Result[V]) bool {
			if v.IsErr() {
				return yield(v)
			}
			i++
			if (i-1)%n == 0 {
				return yield(v)
			}
			return true
		})
	}
}

// Take returns a new sequence with the first n Ok elements.
// If an Err is encountered, it is yielded downstream as-is (consumer-driven).
// After n Ok elements are yielded, the sequence ends.
func (seq SeqResult[V]) Take(n Int) SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		if n <= 0 {
			return
		}

		seq(func(v Result[V]) bool {
			// Once n Ok values are taken, stop hard — nothing further (Ok or Err)
			// is yielded, even if the source ignores our stop signal.
			if n == 0 {
				return false
			}

			if v.IsErr() {
				return yield(v)
			}

			if !yield(v) {
				return false
			}

			n--

			// Stop tightly once n elements are taken so a well-behaved source is
			// not pulled one extra time.
			return n > 0
		})
	}
}

// Nth returns the nth Ok element (0-indexed) in the sequence.
// If an Err is encountered before reaching the nth element, that Err is returned.
// If there are fewer than n+1 Ok elements, None is returned.
func (seq SeqResult[V]) Nth(n Int) Result[Option[V]] {
	if n < 0 {
		return Ok(None[V]())
	}

	var i Int
	result := Ok(None[V]())
	found := false

	seq(func(v Result[V]) bool {
		if found {
			return false
		}
		if v.IsErr() {
			result = Err[Option[V]](v.err)
			found = true
			return false
		}
		if i == n {
			result = Ok(Some(v.v))
			found = true
			return false
		}
		i++
		return true
	})

	return result
}

// Chain concatenates this sequence with other sequences, returning a new sequence of Result[V].
//
// The function yields all elements (Ok or Err) from the current sequence, then from each of the provided sequences in order.
// Err elements are yielded like any other element (consumer-driven).
func (seq SeqResult[V]) Chain(seqs ...SeqResult[V]) SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		stopped := false

		for _, seq := range append([]SeqResult[V]{seq}, seqs...) {
			seq(func(v Result[V]) bool {
				if !yield(v) {
					stopped = true
					return false
				}

				return true
			})

			if stopped {
				return
			}
		}
	}
}

// Intersperse inserts the provided Ok separator between each Ok element of the sequence.
//
// If an Err is encountered, it is yielded as-is without a separator (consumer-driven).
// For Ok elements, after the first yield, a separator is inserted before each subsequent Ok value.
func (seq SeqResult[V]) Intersperse(sep V) SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		first := true

		seq(func(v Result[V]) bool {
			if v.IsErr() {
				return yield(v)
			}

			if !first && !yield(Ok(sep)) {
				return false
			}

			first = false
			return yield(v)
		})
	}
}

// Inspect calls fn for every Ok value without changing it.
// Err elements are passed through unchanged (consumer-driven).
func (seq SeqResult[V]) Inspect(fn func(v V)) SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		seq(func(v Result[V]) bool {
			if v.IsErr() {
				return yield(v)
			}
			fn(v.v)
			return yield(v)
		})
	}
}

// Find searches the sequence for the first Ok value that satisfies the provided function.
//
// If an Err is encountered, it returns that Err immediately. If a matching Ok value is found,
// iteration stops and we return Ok(Some(...)). If no matching Ok value is found, it returns Ok(None).
func (seq SeqResult[V]) Find(fn func(V) bool) Result[Option[V]] {
	result := Ok(None[V]())

	seq(func(v Result[V]) bool {
		if v.IsErr() {
			result = Err[Option[V]](v.err)
			return false
		}
		if fn(v.v) {
			result = Ok(Some(v.v))
			return false
		}
		return true
	})

	return result
}

// Context allows the iteration to be controlled with a context.Context.
func (seq SeqResult[V]) Context(ctx context.Context) SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		seq(func(v Result[V]) bool {
			select {
			case <-ctx.Done():
				return false
			default:
				return yield(v)
			}
		})
	}
}

// First returns the first Ok element from the sequence.
// If the sequence is empty or contains only Err values, None is returned.
// If an Err is encountered, that Err is returned.
func (seq SeqResult[V]) First() Result[Option[V]] {
	result := Ok(None[V]())
	found := false

	seq(func(v Result[V]) bool {
		if found {
			return false
		}
		if v.IsErr() {
			result = Err[Option[V]](v.err)
			found = true
			return false
		}
		result = Ok(Some(v.v))
		found = true
		return false
	})

	return result
}

// Last returns the last Ok element from the sequence.
// If the sequence is empty or contains only Err values, None is returned.
// If an Err is encountered, that Err is returned.
func (seq SeqResult[V]) Last() Result[Option[V]] {
	result := Ok(None[V]())

	seq(func(v Result[V]) bool {
		if v.IsErr() {
			result = Err[Option[V]](v.err)
			return false
		}
		result = Ok(Some(v.v))
		return true
	})

	return result
}

// Next extracts the next element from the iterator and advances it.
//
// This method consumes the next element from the iterator and returns it wrapped in an Option.
// The iterator itself is modified to point to the remaining elements.
//
// Returns:
// - Option[Result[V]]: Some(Result[V]) if an element exists, None if the iterator is exhausted.
func (seq *SeqResult[V]) Next() Option[Result[V]] {
	if value, remaining, ok := Seq[Result[V]](*seq).seqNext(); ok {
		*seq = SeqResult[V](remaining)
		return Some(value)
	}

	return None[Result[V]]()
}

// Partition separates the sequence into two slices: one containing all Ok values and one containing all errors.
// The iteration continues through all elements, collecting each into the appropriate slice.
func (seq SeqResult[V]) Partition() (Slice[V], Slice[error]) {
	ok := NewSlice[V]()
	err := NewSlice[error]()

	seq(func(v Result[V]) bool {
		if v.IsOk() {
			ok = append(ok, v.v)
		} else {
			err = append(err, v.err)
		}
		return true
	})

	return ok, err
}

// Ok returns a new sequence containing only the Ok values from the original sequence.
// All Err values are filtered out.
func (seq SeqResult[V]) Ok() Seq[V] {
	return Seq[V](func(yield func(V) bool) {
		seq(func(v Result[V]) bool {
			if v.IsOk() {
				return yield(v.v)
			}
			return true
		})
	})
}

// Err returns a new sequence containing only the error values from the original sequence.
// All Ok values are filtered out.
func (seq SeqResult[V]) Err() Seq[error] {
	return Seq[error](func(yield func(error) bool) {
		seq(func(v Result[V]) bool {
			if v.IsErr() {
				return yield(v.err)
			}
			return true
		})
	})
}

// FirstErr returns the first error encountered in the sequence.
// If no error is found, it returns None. The iteration stops at the first error.
func (seq SeqResult[V]) FirstErr() Option[error] {
	result := None[error]()
	found := false

	seq(func(v Result[V]) bool {
		if found {
			return false
		}
		if v.IsErr() {
			result = Some(v.err)
			found = true
			return false
		}
		return true
	})

	return result
}

// FromResultChan converts a channel of Results into a SeqResult iterator.
// It consumes the channel until it's closed, yielding each Result to the iterator.
// This is particularly useful with pool.Stream() for processing task results
// as they complete in real-time.
//
// Example usage with pool.Stream:
//
//	p := pool.New[int]().Limit(10)
//	ch := p.Stream(func() {
//	    for i := range 100 {
//	        p.Go(func() Result[int] {
//	            if i%10 == 0 {
//	                return Err[int](fmt.Errorf("task %d failed", i))
//	            }
//	            return Ok(i * i)
//	        })
//	    }
//	})
//
//	successful, failed := FromResultChan(ch).Partition()
//	fmt.Printf("Successful: %d, Failed: %d\n", successful.Len(), failed.Len())
func FromResultChan[V any](ch <-chan Result[V]) SeqResult[V] {
	return SeqResult[V](seqFromChan(ch))
}

// Fold reduces the sequence to a single value using an accumulator.
// The accumulator type may differ from the element type. The first Err
// short-circuits the iteration and is returned as Err.
func (seq SeqResult[V]) Fold[A any](init A, fn func(acc A, val V) A) Result[A] {
	acc := init
	var err error

	seq(func(r Result[V]) bool {
		if r.IsErr() {
			err = r.err
			return false
		}

		acc = fn(acc, r.v)

		return true
	})

	if err != nil {
		return Err[A](err)
	}

	return Ok(acc)
}

// SumBy maps each Ok value to a numeric value via fn and returns their sum wrapped in Ok.
// The first Err short-circuits: iteration stops and that error is returned as Err[S].
// An empty (or all-consumed) sequence yields Ok of the zero value of S.
func (seq SeqResult[V]) SumBy[S constraints.Number](fn func(V) S) Result[S] {
	var zero S
	return seq.Fold(zero, func(acc S, v V) S { return acc + fn(v) })
}

// ProductBy maps each Ok value to a numeric value via fn and returns their product
// wrapped in Ok. The first Err short-circuits and is returned as Err[S]. An empty
// (or all-consumed) sequence yields Ok of the multiplicative identity, one.
func (seq SeqResult[V]) ProductBy[S constraints.Number](fn func(V) S) Result[S] {
	return seq.Fold(S(1), func(acc S, v V) S { return acc * fn(v) })
}

// FindMap applies fn to each Ok value and returns the first Some result wrapped in
// Ok; None if fn returns None for every value. The first Err short-circuits and is
// returned as Err.
func (seq SeqResult[V]) FindMap[U any](fn func(V) Option[U]) Result[Option[U]] {
	result := Ok(None[U]())

	seq(func(v Result[V]) bool {
		if v.IsErr() {
			result = Err[Option[U]](v.err)
			return false
		}

		if o := fn(v.v); o.IsSome() {
			result = Ok(o)
			return false
		}

		return true
	})

	return result
}

// Reduce aggregates Ok values using the provided function:
// the first Err short-circuits and is returned as Err; an empty sequence yields Ok(None);
// otherwise Ok(Some(accumulated)).
func (seq SeqResult[V]) Reduce(fn func(a, b V) V) Result[Option[V]] {
	var (
		acc   V
		first = true
		err   error
	)

	seq(func(r Result[V]) bool {
		if r.IsErr() {
			err = r.err
			return false
		}

		if first {
			acc, first = r.v, false
		} else {
			acc = fn(acc, r.v)
		}

		return true
	})

	if err != nil {
		return Err[Option[V]](err)
	}

	return Ok(OptionOf(acc, !first))
}

// Scan accumulates Ok values, yielding the initial value followed by every
// intermediate accumulator state. The accumulator type may differ from the
// element type. An Err is passed downstream as-is without touching the
// accumulator; iteration continues for as long as the consumer keeps
// accepting values (consumer-driven).
func (seq SeqResult[V]) Scan[A any](init A, fn func(acc A, val V) A) SeqResult[A] {
	return func(yield func(Result[A]) bool) {
		if !yield(Ok(init)) {
			return
		}

		acc := init

		seq(func(r Result[V]) bool {
			if r.IsErr() {
				return yield(Err[A](r.err))
			}

			acc = fn(acc, r.v)

			return yield(Ok(acc))
		})
	}
}

// FilterMap transforms each Ok value with fn and keeps only the Some results,
// changing the element type from V to U.
//
// If an Err is encountered, it is passed downstream as-is (Err[U]); the consumer
// decides whether to continue (consumer-driven). For an Ok value, fn is applied:
// Some(u) is yielded as Ok(u), None drops the element.
func (seq SeqResult[V]) FilterMap[U any](fn func(V) Option[U]) SeqResult[U] {
	return func(yield func(Result[U]) bool) {
		seq(func(v Result[V]) bool {
			if v.IsErr() {
				return yield(Err[U](v.err))
			}

			if u, ok := fn(v.v).Option(); ok {
				return yield(Ok(u))
			}

			return true
		})
	}
}

// TakeWhile yields Ok values while fn returns true, stopping at the first Ok
// value for which fn returns false.
//
// If an Err is encountered, it is passed downstream as-is and does not stop the
// taking (consumer-driven); only a failing predicate on an Ok value ends it.
func (seq SeqResult[V]) TakeWhile(fn func(V) bool) SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		seq(func(v Result[V]) bool {
			if v.IsErr() {
				return yield(v)
			}

			if !fn(v.v) {
				return false
			}

			return yield(v)
		})
	}
}

// SkipWhile skips Ok values while fn returns true, then yields every remaining
// element.
//
// If an Err is encountered, it is passed downstream as-is regardless of the
// skipping phase (consumer-driven); the skipping predicate is evaluated only on
// Ok values.
func (seq SeqResult[V]) SkipWhile(fn func(V) bool) SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		skipping := true

		seq(func(v Result[V]) bool {
			if v.IsErr() {
				return yield(v)
			}

			if skipping && fn(v.v) {
				return true
			}

			skipping = false

			return yield(v)
		})
	}
}

// MaxBy returns the maximum Ok value according to fn, mirroring the short-circuit
// terminals: the first Err stops iteration and is returned as Err; a sequence with
// no Ok values yields Ok(None).
func (seq SeqResult[V]) MaxBy(fn func(V, V) cmp.Ordering) Result[Option[V]] {
	var best V
	has := false

	result := Ok(None[V]())

	seq(func(v Result[V]) bool {
		if v.IsErr() {
			result = Err[Option[V]](v.err)
			return false
		}

		if !has || fn(best, v.v).IsLt() {
			best = v.v
			has = true
		}

		return true
	})

	if result.IsErr() {
		return result
	}

	if has {
		return Ok(Some(best))
	}

	return Ok(None[V]())
}

// MinBy returns the minimum Ok value according to fn, mirroring the short-circuit
// terminals: the first Err stops iteration and is returned as Err; a sequence with
// no Ok values yields Ok(None).
func (seq SeqResult[V]) MinBy(fn func(V, V) cmp.Ordering) Result[Option[V]] {
	var best V
	has := false

	result := Ok(None[V]())

	seq(func(v Result[V]) bool {
		if v.IsErr() {
			result = Err[Option[V]](v.err)
			return false
		}

		if !has || fn(v.v, best).IsLt() {
			best = v.v
			has = true
		}

		return true
	})

	if result.IsErr() {
		return result
	}

	if has {
		return Ok(Some(best))
	}

	return Ok(None[V]())
}

// Flatten flattens one or more levels of nested slices/arrays inside each Ok
// value, yielding the leaf elements as Ok. Err elements are passed downstream
// as-is (consumer-driven). It mirrors Seq.Flatten and, like it, relies on
// reflection: only leaves assignable to V are yielded.
func (seq SeqResult[V]) Flatten() SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		emit := func(v V) bool { return yield(Ok(v)) }

		seq(func(v Result[V]) bool {
			if v.IsErr() {
				return yield(v)
			}

			return flattenValue(v.v, emit)
		})
	}
}

// SortBy consumes the sequence, sorts the Ok values with fn, and re-emits them
// in order as a SeqResult. Being a sort, it is eager: the whole sequence is
// buffered first. The first Err short-circuits — buffering stops and only that
// Err is yielded downstream.
func (seq SeqResult[V]) SortBy(fn func(a, b V) cmp.Ordering) SeqResult[V] {
	return func(yield func(Result[V]) bool) {
		items := NewSlice[V]()

		var err error

		seq(func(v Result[V]) bool {
			if v.IsErr() {
				err = v.err
				return false
			}

			items = append(items, v.v)

			return true
		})

		if err != nil {
			yield(Err[V](err))
			return
		}

		items.SortBy(fn)

		for _, v := range items {
			if !yield(Ok(v)) {
				return
			}
		}
	}
}

// CounterBy counts how many Ok values map to each key produced by fn, returning
// the tally as plain pairs in first-seen key order (convert with MapOrd[K, Int]
// if map access is needed). It is a short-circuit terminal: the first Err stops
// the count and is returned as Err. (A lazy SeqResult of the tally is
// impossible here — it would instantiate SeqResult with a type built from V and
// hit an instantiation cycle; returning MapOrd would weld SeqResult to the map
// cluster.)
func (seq SeqResult[V]) CounterBy[K comparable](fn func(V) K) Result[[]Pair[K, Int]] {
	order := NewSlice[K]()
	counts := NewMap[K, Int]()

	var err error

	seq(func(v Result[V]) bool {
		if v.IsErr() {
			err = v.err
			return false
		}

		k := fn(v.v)
		if !counts.Contains(k) {
			order.Push(k)
		}

		counts[k]++

		return true
	})

	if err != nil {
		return Err[[]Pair[K, Int]](err)
	}

	result := make([]Pair[K, Int], 0, order.Len())
	for _, k := range order {
		result = append(result, Pair[K, Int]{Key: k, Value: counts[k]})
	}

	return Ok(result)
}
