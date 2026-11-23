package g

import (
	"fmt"
	"reflect"

	"github.com/enetx/g/cmp"
	"github.com/enetx/g/f"
)

// Heap is a generic binary heap data structure that maintains elements in heap order.
// It can be configured as either a min-heap or max-heap based on the comparison function.
type Heap[T any] struct {
	data Slice[T]
	cmp  func(T, T) cmp.Ordering
}

// NewHeap creates a new heap with the given comparison function.
// The comparison function should return:
// - cmp.Less if the first argument should have higher priority
// - cmp.Greater if the second argument should have higher priority
// - cmp.Equal if they have equal priority
//
// NewHeap panics if compareFn is nil, since a nil
// comparison function would otherwise nil-deref on the first Push.
func NewHeap[T any](compareFn func(T, T) cmp.Ordering) *Heap[T] {
	if compareFn == nil {
		panic("g.NewHeap: compareFn cannot be nil")
	}

	return &Heap[T]{
		data: make(Slice[T], 0),
		cmp:  compareFn,
	}
}

// Transform applies a transformation function to the Heap and returns the result.
func (h *Heap[T]) Transform[U any](fn func(*Heap[T]) U) U { return fn(h) }

// Iter returns a non-consuming iterator that yields elements in sorted order.
//
// The iterator creates a clone of the heap and yields elements by repeatedly
// calling Pop() on the clone, ensuring the original heap remains unchanged.
// Elements are yielded in the order determined by the heap's comparison function
// (smallest first for min-heap, largest first for max-heap).
//
// Time complexity: O(n log n) for full iteration
// Space complexity: O(n) for the heap clone
//
// Returns:
//
// - Seq[T]: An iterator that yields elements in sorted order
//
// Example usage:
//
//	heap := g.NewHeap(cmp.Cmp[int])
//	heap.Push(10, 5, 15, 1, 8)
//
//	// Iterate without consuming the original heap
//	heap.Iter().ForEach(func(x int) {
//		fmt.Printf("%d ", x) // Output: 1 5 8 10 15
//	})
//
//	fmt.Printf("Heap still has %d elements\n", heap.Len()) // Output: 5
//
//	// Can be used with other iterator methods
//	// (the Heap materializer requires a comparison function)
//	firstThree := heap.Iter().Take(3).Collect().Heap(cmp.Cmp) // [1, 5, 8]
//	evenNumbers := heap.Iter().Filter(func(x int) bool {
//		return x%2 == 0
//	}).Collect().Heap(cmp.Cmp) // [8, 10]
func (h *Heap[T]) Iter() Seq[T] {
	return func(yield func(T) bool) {
		clone := h.Clone()
		for !clone.IsEmpty() {
			if !yield(clone.Pop().Some()) {
				return
			}
		}
	}
}

// IntoIter returns a consuming iterator that yields elements in sorted order.
//
// This iterator consumes the original heap by repeatedly calling Pop() until
// the heap is empty. After iteration completes (or is stopped early), the
// original heap will be empty. Elements are yielded in the order determined
// by the heap's comparison function (smallest first for min-heap, largest first for max-heap).
//
// Use this method when you want to consume the heap and don't need the original
// data structure afterwards, or when you want to transfer ownership of the elements.
//
// Time complexity: O(n log n) for full iteration
// Space complexity: O(1) - no additional memory allocation
//
// Returns:
//
// - Seq[T]: An iterator that yields elements in sorted order while consuming the heap
//
// Example usage:
//
//	heap := g.NewHeap(cmp.Cmp[int])
//	heap.Push(10, 5, 15, 1, 8)
//
//	// Consume the heap while iterating
//	result := heap.IntoIter().Collect().Heap(cmp.Cmp) // [1, 5, 8, 10, 15]
//
//	fmt.Printf("Heap now has %d elements\n", heap.Len()) // Output: 0
//
//	// Can be stopped early, leaving remaining elements in heap
//	heap2 := g.NewHeap(cmp.Cmp[int])
//	heap2.Push(20, 25, 15, 30)
//
//	heap2.IntoIter().Take(2).ForEach(func(x int) {
//		fmt.Printf("%d ", x) // Output: 15 20
//	})
//	fmt.Printf("Remaining: %d elements\n", heap2.Len()) // Output: 2
func (h *Heap[T]) IntoIter() Seq[T] {
	return func(yield func(T) bool) {
		for !h.IsEmpty() {
			if !yield(h.Pop().Some()) {
				return
			}
		}
	}
}

// Push adds one or more items to the heap.
func (h *Heap[T]) Push(items ...T) {
	if len(items) == 1 {
		h.data = append(h.data, items[0])
		h.heapifyUp(len(h.data) - 1)
		return
	}

	if len(items) > 1 {
		start := len(h.data)
		h.data = append(h.data, items...)

		// Rebuilding is linear and wins for large batches. For a small batch on
		// an established heap, sift only the appended elements to avoid scanning
		// the entire existing heap.
		if start == 0 || len(items) > start/2 {
			h.heapify()
			return
		}

		for i := start; i < len(h.data); i++ {
			h.heapifyUp(i)
		}
	}
}

// Pop removes and returns the top element from the heap.
// Returns None if the heap is empty.
func (h *Heap[T]) Pop() Option[T] {
	if len(h.data) == 0 {
		return None[T]()
	}

	top := h.data[0]
	last := len(h.data) - 1
	h.data[0] = h.data[last]
	var zero T
	h.data[last] = zero
	h.data = h.data[:last]

	if len(h.data) > 0 {
		h.heapifyDown(0)
	}

	return Some(top)
}

// Peek returns the top element without removing it.
// Returns None if the heap is empty.
func (h *Heap[T]) Peek() Option[T] {
	if len(h.data) == 0 {
		return None[T]()
	}

	return Some(h.data[0])
}

// Contains reports whether the heap contains the given value.
//
// Equality is determined the same way as Slice.Contains: a direct == fast path
// for comparable element types, falling back to reflect.DeepEqual for
// interface-typed or otherwise uncomparable values.
func (h *Heap[T]) Contains(value T) bool { return h.data.Contains(value) }

// Remove removes and returns the element at index i in the heap's backing
// storage. Indices follow the internal heap layout (index 0 is the root);
// use Slice to observe element positions.
//
// Returns None if i is out of range. After removal the heap property is
// restored in O(log n).
func (h *Heap[T]) Remove(i Int) Option[T] {
	n := len(h.data) - 1
	if i < 0 || int(i) > n {
		return None[T]()
	}

	idx := int(i)
	removed := h.data[idx]

	if idx != n {
		h.data[idx] = h.data[n]
	}

	var zero T
	h.data[n] = zero
	h.data = h.data[:n]

	if idx < len(h.data) {
		h.heapifyDown(idx)
		h.heapifyUp(idx)
	}

	return Some(removed)
}

// Fix re-establishes the heap ordering after the element at index i has changed
// its value. It is equivalent to, but less expensive than, removing the element
// at index i and pushing the new value.
//
// Indices follow the internal heap layout (index 0 is the root). Fix is a no-op
// if i is out of range. The cost is O(log n).
func (h *Heap[T]) Fix(i Int) {
	if i < 0 || int(i) >= len(h.data) {
		return
	}

	idx := int(i)
	h.heapifyDown(idx)
	h.heapifyUp(idx)
}

// Len returns the number of elements in the heap.
func (h *Heap[T]) Len() Int {
	return h.data.Len()
}

// IsEmpty returns true if the heap contains no elements.
func (h *Heap[T]) IsEmpty() bool {
	return len(h.data) == 0
}

// Slice returns a slice containing all elements in the heap.
// The order is not guaranteed to be sorted.
func (h *Heap[T]) Slice() Slice[T] {
	result := make(Slice[T], len(h.data))
	copy(result, h.data)

	return result
}

// Clear removes all elements from the heap and releases the backing array,
// allowing the previously held elements to be garbage collected.
func (h *Heap[T]) Clear() {
	h.data = nil
}

// Clone creates a deep copy of the heap.
func (h *Heap[T]) Clone() *Heap[T] {
	return &Heap[T]{
		data: h.data.Clone(),
		cmp:  h.cmp,
	}
}

// Eq checks if two Heaps are equal.
//
// Heaps are considered equal if they yield the same elements in the same
// iteration order (the sorted order produced by Iter), regardless of the
// internal layout of their backing storage. The comparison functions
// themselves are not compared; each heap is drained using its own ordering.
func (h *Heap[T]) Eq(other *Heap[T]) bool {
	if h == other {
		return true
	}

	if h == nil || other == nil {
		return false
	}

	if h.Len() != other.Len() {
		return false
	}

	a, b := h.Clone(), other.Clone()

	if f.IsComparable[T]() && reflect.TypeFor[T]().Kind() != reflect.Interface {
		for !a.IsEmpty() {
			if any(a.Pop().Some()) != any(b.Pop().Some()) {
				return false
			}
		}
	} else {
		for !a.IsEmpty() {
			if !reflect.DeepEqual(a.Pop().Some(), b.Pop().Some()) {
				return false
			}
		}
	}

	return true
}

// Ne checks if two Heaps are not equal.
func (h *Heap[T]) Ne(other *Heap[T]) bool { return !h.Eq(other) }

// heapify transforms the entire data slice into a valid heap.
func (h *Heap[T]) heapify() {
	for i := len(h.data)/2 - 1; i >= 0; i-- {
		h.heapifyDown(i)
	}
}

// heapifyUp maintains heap property by moving element up.
func (h *Heap[T]) heapifyUp(idx int) {
	for idx > 0 {
		parent := (idx - 1) / 2

		if h.cmp(h.data[idx], h.data[parent]) != cmp.Less {
			break
		}

		h.data[idx], h.data[parent] = h.data[parent], h.data[idx]
		idx = parent
	}
}

// heapifyDown maintains heap property by moving element down.
func (h *Heap[T]) heapifyDown(idx int) {
	for {
		smallest := idx
		left := 2*idx + 1
		right := 2*idx + 2

		if left < len(h.data) && h.cmp(h.data[left], h.data[smallest]) == cmp.Less {
			smallest = left
		}

		if right < len(h.data) && h.cmp(h.data[right], h.data[smallest]) == cmp.Less {
			smallest = right
		}

		if smallest == idx {
			break
		}

		h.data[idx], h.data[smallest] = h.data[smallest], h.data[idx]
		idx = smallest
	}
}

// String returns a string representation of the heap.
func (h *Heap[T]) String() string {
	if len(h.data) == 0 {
		return "Heap[]"
	}

	var b Builder
	b.Grow(Int(len(h.data)) * 8)
	b.WriteString("Heap[")

	for i, v := range h.data {
		if i > 0 {
			b.WriteString(", ")
		}

		fmt.Fprint(&b, v)
	}

	b.WriteString("]")

	return b.String().Std()
}

// Print writes the elements of the Heap to the standard output (console)
// and returns the Heap unchanged.
func (h *Heap[T]) Print() *Heap[T] { fmt.Print(h); return h }

// Println writes the elements of the Heap to the standard output (console) with a newline
// and returns the Heap unchanged.
func (h *Heap[T]) Println() *Heap[T] { fmt.Println(h); return h }

// HeapOf creates a new Heap with the given comparison function containing the provided elements.
func HeapOf[T any](compareFn func(T, T) cmp.Ordering, values ...T) *Heap[T] {
	h := NewHeap(compareFn)
	h.Push(values...)

	return h
}
