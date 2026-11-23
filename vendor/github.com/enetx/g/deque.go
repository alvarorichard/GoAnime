package g

import (
	"fmt"
	"reflect"
	"slices"

	"github.com/enetx/g/cmp"
	"github.com/enetx/g/f"
)

// Deque is a double-ended queue implemented with a growable ring buffer.
// It provides efficient insertion and removal of elements at both ends.
type Deque[T any] struct {
	data  Slice[T]
	front Int
	len   Int
}

// NewDeque creates a new Deque of the given generic type T with the specified capacity.
// The capacity parameter specifies the initial capacity of the underlying slice.
// If no capacity is provided, an empty Deque with a capacity of 0 is returned.
//
// Parameters:
//
// - capacity ...Int: An optional parameter specifying the initial capacity of the Deque
//
// Returns:
//
// - *Deque[T]: A new Deque of the specified generic type T with the given capacity
//
// Example usage:
//
//	d1 := g.NewDeque[int]()     // Creates an empty Deque of type int
//	d2 := g.NewDeque[int](10)   // Creates an empty Deque with capacity of 10
func NewDeque[T any](capacity ...Int) *Deque[T] {
	cap := Int(0)

	if len(capacity) > 0 {
		cap = capacity[0]
	}

	return &Deque[T]{
		data:  make(Slice[T], cap),
		front: 0,
		len:   0,
	}
}

// DequeOf creates a new Deque containing the provided elements.
func DequeOf[T any](elements ...T) *Deque[T] {
	n := Int(len(elements))
	data := make(Slice[T], n)
	copy(data, elements)

	return &Deque[T]{
		data:  data,
		front: 0,
		len:   n,
	}
}

// Len returns the number of elements in the Deque.
func (dq *Deque[T]) Len() Int {
	return dq.len
}

// IsEmpty returns true if the Deque contains no elements.
func (dq *Deque[T]) IsEmpty() bool {
	return dq.len == 0
}

// Capacity returns the current capacity of the Deque.
func (dq *Deque[T]) Capacity() Int {
	return Int(len(dq.data))
}

// realIndex converts a logical index to the actual index in the ring buffer.
func (dq *Deque[T]) realIndex(index Int) Int {
	return (dq.front + index) % Int(len(dq.data))
}

// grow expands the capacity of the Deque when needed.
func (dq *Deque[T]) grow() {
	oldCap := Int(len(dq.data))
	newCap := oldCap * 2
	if newCap == 0 {
		newCap = 4
	}

	newData := make(Slice[T], newCap)
	dq.copyToContiguous(newData)
	dq.data = newData
	dq.front = 0
}

// copyToContiguous copies the deque's logical elements into dst starting at index 0.
// dst must have at least dq.len capacity.
func (dq *Deque[T]) copyToContiguous(dst Slice[T]) {
	if dq.len == 0 {
		return
	}

	cap := Int(len(dq.data))
	if dq.front+dq.len <= cap {
		// Contiguous region
		copy(dst, dq.data[dq.front:dq.front+dq.len])
	} else {
		// Wrap-around: two copy operations
		firstPart := cap - dq.front
		copy(dst, dq.data[dq.front:])
		copy(dst[firstPart:], dq.data[:dq.len-firstPart])
	}
}

// PushFront adds an element to the front of the Deque.
func (dq *Deque[T]) PushFront(value T) {
	if dq.len == Int(len(dq.data)) {
		dq.grow()
	}

	dq.front = (dq.front - 1 + Int(len(dq.data))) % Int(len(dq.data))
	dq.data[dq.front] = value
	dq.len++
}

// PushBack adds an element to the back of the Deque.
func (dq *Deque[T]) PushBack(value T) {
	if dq.len == Int(len(dq.data)) {
		dq.grow()
	}

	backIndex := dq.realIndex(dq.len)
	dq.data[backIndex] = value
	dq.len++
}

// Extend appends the given values to the back of the Deque, in order.
// It accepts a spread slice too: dq.Extend(sl...).
func (dq *Deque[T]) Extend(values ...T) {
	dq.Reserve(Int(len(values)))
	for _, v := range values {
		dq.PushBack(v)
	}
}

// PopFront removes and returns the first element of the Deque.
// Returns None if the Deque is empty.
func (dq *Deque[T]) PopFront() Option[T] {
	if dq.IsEmpty() {
		return None[T]()
	}

	value := dq.data[dq.front]
	var zero T
	dq.data[dq.front] = zero
	dq.front = (dq.front + 1) % Int(len(dq.data))
	dq.len--

	return Some(value)
}

// PopBack removes and returns the last element of the Deque.
// Returns None if the Deque is empty.
func (dq *Deque[T]) PopBack() Option[T] {
	if dq.IsEmpty() {
		return None[T]()
	}

	dq.len--
	backIndex := dq.realIndex(dq.len)
	value := dq.data[backIndex]
	var zero T
	dq.data[backIndex] = zero

	return Some(value)
}

// Front returns the first element of the Deque.
// Returns None if the Deque is empty.
func (dq *Deque[T]) Front() Option[T] {
	if dq.IsEmpty() {
		return None[T]()
	}

	return Some(dq.data[dq.front])
}

// Back returns the last element of the Deque.
// Returns None if the Deque is empty.
func (dq *Deque[T]) Back() Option[T] {
	if dq.IsEmpty() {
		return None[T]()
	}

	backIndex := dq.realIndex(dq.len - 1)

	return Some(dq.data[backIndex])
}

// Get retrieves an element at the specified index.
// Index 0 represents the front of the Deque.
// Returns None if the index is out of bounds.
func (dq *Deque[T]) Get(index Int) Option[T] {
	if index < 0 || index >= dq.len {
		return None[T]()
	}

	realIdx := dq.realIndex(index)

	return Some(dq.data[realIdx])
}

// Set sets the element at the specified index, returning the old value as an Option.
// Index 0 represents the front of the Deque.
// Returns None if the index is out of bounds.
func (dq *Deque[T]) Set(index Int, value T) Option[T] {
	if index < 0 || index >= dq.len {
		return None[T]()
	}

	realIdx := dq.realIndex(index)
	old := dq.data[realIdx]
	dq.data[realIdx] = value

	return Some(old)
}

// Insert inserts an element at the specified index.
// Index 0 represents the front of the Deque.
// Panics if the index is out of bounds (index < 0 or index > Len()).
func (dq *Deque[T]) Insert(index Int, value T) {
	if index < 0 || index > dq.len {
		panic(fmt.Sprintf("runtime error: deque index out of range [%d] with length %d", index, dq.len))
	}

	if index == 0 {
		dq.PushFront(value)
		return
	}

	if index == dq.len {
		dq.PushBack(value)
		return
	}

	if index <= dq.len/2 {
		if dq.len == Int(len(dq.data)) {
			dq.grow()
		}

		dq.front = (dq.front - 1 + Int(len(dq.data))) % Int(len(dq.data))
		dq.len++

		for i := range index {
			dq.data[dq.realIndex(i)] = dq.data[dq.realIndex(i+1)]
		}

		dq.data[dq.realIndex(index)] = value
	} else {
		if dq.len == Int(len(dq.data)) {
			dq.grow()
		}

		dq.len++

		for i := dq.len - 1; i > index; i-- {
			dq.data[dq.realIndex(i)] = dq.data[dq.realIndex(i-1)]
		}

		dq.data[dq.realIndex(index)] = value
	}
}

// Remove removes and returns the element at the specified index.
// Returns None if the index is out of bounds.
func (dq *Deque[T]) Remove(index Int) Option[T] {
	if index < 0 || index >= dq.len {
		return None[T]()
	}

	if index == 0 {
		return dq.PopFront()
	}

	if index == dq.len-1 {
		return dq.PopBack()
	}

	realIdx := dq.realIndex(index)
	value := dq.data[realIdx]

	if index <= dq.len/2 {
		for i := index; i > 0; i-- {
			dq.data[dq.realIndex(i)] = dq.data[dq.realIndex(i-1)]
		}
		var zero T
		dq.data[dq.front] = zero
		dq.front = (dq.front + 1) % Int(len(dq.data))
	} else {
		for i := index; i < dq.len-1; i++ {
			dq.data[dq.realIndex(i)] = dq.data[dq.realIndex(i+1)]
		}
		var zero T
		backIdx := dq.realIndex(dq.len - 1)
		dq.data[backIdx] = zero
	}

	dq.len--

	return Some(value)
}

// Clear removes all elements from the Deque.
func (dq *Deque[T]) Clear() {
	clear(dq.data)
	dq.front = 0
	dq.len = 0
}

// Swap swaps the elements at indices i and j.
// Panics if either index is out of bounds.
func (dq *Deque[T]) Swap(i, j Int) {
	if i < 0 || i >= dq.len {
		panic(fmt.Sprintf("runtime error: deque index out of range [%d] with length %d", i, dq.len))
	}

	if j < 0 || j >= dq.len {
		panic(fmt.Sprintf("runtime error: deque index out of range [%d] with length %d", j, dq.len))
	}

	realI := dq.realIndex(i)
	realJ := dq.realIndex(j)

	dq.data[realI], dq.data[realJ] = dq.data[realJ], dq.data[realI]
}

// RotateLeft rotates the Deque in-place such that the first mid elements
// move to the end while the last len - mid elements move to the front.
func (dq *Deque[T]) RotateLeft(mid Int) {
	if dq.len == 0 {
		return
	}

	mid = mid % dq.len
	if mid == 0 {
		return
	}
	if dq.len == Int(len(dq.data)) {
		dq.front = dq.realIndex(mid)
		return
	}

	contiguous := dq.MakeContiguous()

	reverseDequeSlice(contiguous[:mid])
	reverseDequeSlice(contiguous[mid:])
	reverseDequeSlice(contiguous)
}

// RotateRight rotates the Deque in-place such that the first len - k elements
// move to the end while the last k elements move to the front.
func (dq *Deque[T]) RotateRight(k Int) {
	if dq.len == 0 {
		return
	}

	k = k % dq.len
	if k == 0 {
		return
	}

	dq.RotateLeft(dq.len - k)
}

// MakeContiguous rearranges the internal storage of the Deque so that its elements
// are in contiguous memory. Returns a slice that contains all elements.
func (dq *Deque[T]) MakeContiguous() Slice[T] {
	if dq.len == 0 {
		return Slice[T]{}
	}

	if dq.front+dq.len <= Int(len(dq.data)) {
		return dq.data[dq.front : dq.front+dq.len]
	}

	newData := make(Slice[T], len(dq.data))
	dq.copyToContiguous(newData)
	dq.data = newData
	dq.front = 0

	return dq.data[:dq.len]
}

// Clone creates a deep copy of the Deque.
func (dq *Deque[T]) Clone() *Deque[T] {
	newData := make(Slice[T], len(dq.data))
	dq.copyToContiguous(newData)

	return &Deque[T]{
		data:  newData,
		front: 0,
		len:   dq.len,
	}
}

// Iter returns an iterator for the Deque, allowing for sequential iteration
// over its elements from front to back.
func (dq *Deque[T]) Iter() Seq[T] {
	return func(yield func(T) bool) {
		cap := Int(len(dq.data))
		if cap == 0 || dq.len == 0 {
			return
		}

		if dq.front+dq.len <= cap {
			// Contiguous: iterate directly
			for _, v := range dq.data[dq.front : dq.front+dq.len] {
				if !yield(v) {
					return
				}
			}
		} else {
			// Wrap-around: first part from front to end, then from start
			for _, v := range dq.data[dq.front:] {
				if !yield(v) {
					return
				}
			}
			wrapLen := dq.len - (cap - dq.front)
			for _, v := range dq.data[:wrapLen] {
				if !yield(v) {
					return
				}
			}
		}
	}
}

// IterReverse returns an iterator for the Deque that allows for sequential iteration
// over its elements in reverse order (from back to front).
func (dq *Deque[T]) IterReverse() Seq[T] {
	return func(yield func(T) bool) {
		for i := dq.len - 1; i >= 0; i-- {
			value := dq.data[dq.realIndex(i)]
			if !yield(value) {
				return
			}
		}
	}
}

// Reserve ensures that the Deque can hold at least the specified number of elements
// without reallocating. If the current capacity is already sufficient, this is a no-op.
func (dq *Deque[T]) Reserve(additional Int) {
	required := dq.len + additional
	if required <= Int(len(dq.data)) {
		return
	}

	newCap := Int(len(dq.data))
	if newCap == 0 {
		newCap = 4
	}

	for newCap < required {
		newCap *= 2
	}

	newData := make(Slice[T], newCap)
	dq.copyToContiguous(newData)
	dq.data = newData
	dq.front = 0
}

// ShrinkToFit shrinks the capacity of the Deque as much as possible.
func (dq *Deque[T]) ShrinkToFit() {
	if dq.len == 0 {
		dq.data = Slice[T]{}
		dq.front = 0

		return
	}

	if Int(len(dq.data)) == dq.len {
		return
	}

	newData := make(Slice[T], dq.len)
	dq.copyToContiguous(newData)
	dq.data = newData
	dq.front = 0
}

// Contains checks if the Deque contains the specified value.
func (dq *Deque[T]) Contains(value T) bool {
	if dq.len == 0 {
		return false
	}

	cap := Int(len(dq.data))

	if f.IsComparable[T]() && reflect.TypeFor[T]().Kind() != reflect.Interface {
		target := any(value)
		if dq.front+dq.len <= cap {
			for _, v := range dq.data[dq.front : dq.front+dq.len] {
				if any(v) == target {
					return true
				}
			}
		} else {
			for _, v := range dq.data[dq.front:] {
				if any(v) == target {
					return true
				}
			}
			wrapLen := dq.len - (cap - dq.front)
			for _, v := range dq.data[:wrapLen] {
				if any(v) == target {
					return true
				}
			}
		}
	} else {
		for i := Int(0); i < dq.len; i++ {
			if reflect.DeepEqual(dq.data[dq.realIndex(i)], value) {
				return true
			}
		}
	}

	return false
}

// ContainsAny checks if the Deque contains any element from the provided values.
func (dq *Deque[T]) ContainsAny(values ...T) bool {
	if dq.len == 0 || len(values) == 0 {
		return false
	}

	return slices.ContainsFunc(values, dq.Contains)
}

// ContainsAll checks if the Deque contains all of the provided values.
func (dq *Deque[T]) ContainsAll(values ...T) bool {
	if len(values) == 0 {
		return true
	}

	if dq.len == 0 {
		return false
	}

	for _, v := range values {
		if !dq.Contains(v) {
			return false
		}
	}

	return true
}

// Index returns the index of the first occurrence of the specified value,
// or -1 if not found.
func (dq *Deque[T]) Index(value T) Int {
	if f.IsComparable[T]() && reflect.TypeFor[T]().Kind() != reflect.Interface {
		target := any(value)
		for i := Int(0); i < dq.len; i++ {
			if any(dq.data[dq.realIndex(i)]) == target {
				return i
			}
		}
	} else {
		for i := Int(0); i < dq.len; i++ {
			if reflect.DeepEqual(dq.data[dq.realIndex(i)], value) {
				return i
			}
		}
	}

	return -1
}

// BinarySearch searches for a value in a sorted Deque using binary search.
// Returns the index where the value is found, or where it should be inserted,
// and a boolean reporting whether the value was found.
func (dq *Deque[T]) BinarySearch(value T, fn func(T, T) cmp.Ordering) (Int, bool) {
	left, right := Int(0), dq.len
	for left < right {
		mid := (left + right) / 2
		result := fn(dq.data[dq.realIndex(mid)], value)

		switch result {
		case cmp.Less:
			left = mid + 1
		case cmp.Greater:
			right = mid
		case cmp.Equal:
			return mid, true
		}
	}

	return left, false
}

func reverseDequeSlice[T any](values Slice[T]) {
	for left, right := 0, len(values)-1; left < right; left, right = left+1, right-1 {
		values[left], values[right] = values[right], values[left]
	}
}

// Transform applies a transformation function to the Deque and returns the result.
func (dq *Deque[T]) Transform[U any](fn func(*Deque[T]) U) U { return fn(dq) }

// String returns a string representation of the Deque.
func (dq *Deque[T]) String() string {
	if dq.IsEmpty() {
		return "Deque[]"
	}

	var b Builder
	b.Grow(dq.len * 8)
	b.WriteString("Deque[")

	cap := Int(len(dq.data))
	first := true

	writeElem := func(v T) {
		if !first {
			b.WriteString(", ")
		}
		first = false
		fmt.Fprint(&b, v)
	}

	if dq.front+dq.len <= cap {
		for _, v := range dq.data[dq.front : dq.front+dq.len] {
			writeElem(v)
		}
	} else {
		for _, v := range dq.data[dq.front:] {
			writeElem(v)
		}
		wrapLen := dq.len - (cap - dq.front)
		for _, v := range dq.data[:wrapLen] {
			writeElem(v)
		}
	}

	b.WriteString("]")

	return b.String().Std()
}

// Eq checks if two Deques are equal.
func (dq *Deque[T]) Eq(other *Deque[T]) bool {
	if dq == other {
		return true
	}

	if dq == nil || other == nil {
		return false
	}

	if dq.len != other.len {
		return false
	}

	if f.IsComparable[T]() && reflect.TypeFor[T]().Kind() != reflect.Interface {
		for i := Int(0); i < dq.len; i++ {
			if any(dq.data[dq.realIndex(i)]) != any(other.data[other.realIndex(i)]) {
				return false
			}
		}
	} else {
		for i := Int(0); i < dq.len; i++ {
			a := dq.data[dq.realIndex(i)]
			b := other.data[other.realIndex(i)]
			if !reflect.DeepEqual(a, b) {
				return false
			}
		}
	}

	return true
}

// Ne checks if two Deques are not equal.
func (dq *Deque[T]) Ne(other *Deque[T]) bool { return !dq.Eq(other) }

// Retain keeps only the elements specified by the predicate.
func (dq *Deque[T]) Retain(predicate func(T) bool) {
	writePos := Int(0)

	for i := Int(0); i < dq.len; i++ {
		value := dq.data[dq.realIndex(i)]
		if predicate(value) {
			if writePos != i {
				dq.data[dq.realIndex(writePos)] = value
			}
			writePos++
		}
	}

	var zero T
	for i := writePos; i < dq.len; i++ {
		dq.data[dq.realIndex(i)] = zero
	}

	dq.len = writePos
}

// Print writes the elements of the Deque to the standard output (console)
// and returns the Deque unchanged.
func (dq *Deque[T]) Print() *Deque[T] { fmt.Print(dq); return dq }

// Println writes the elements of the Deque to the standard output (console) with a newline
// and returns the Deque unchanged.
func (dq *Deque[T]) Println() *Deque[T] { fmt.Println(dq); return dq }
