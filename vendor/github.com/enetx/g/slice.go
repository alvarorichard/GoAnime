package g

import (
	"fmt"
	"reflect"
	"slices"
	"strings"

	"github.com/enetx/g/cmp"
	"github.com/enetx/g/f"
)

// Slice is a generic alias for a slice.
type Slice[T any] []T

// NewSlice creates a new Slice of the given generic type T with the specified length and
// capacity.
// The size variadic parameter can have zero, one, or two integer values.
// If no values are provided, an empty Slice with a length and capacity of 0 is returned.
// If one value is provided, it sets both the length and capacity of the Slice.
// If two values are provided, the first value sets the length and the second value sets the
// capacity.
//
// Parameters:
//
// - size ...Int: A variadic parameter specifying the length and/or capacity of the Slice
//
// Returns:
//
// - Slice[T]: A new Slice of the specified generic type T with the given length and capacity
//
// Example usage:
//
//	s1 := g.NewSlice[int]()        // Creates an empty Slice of type int
//	s2 := g.NewSlice[int](5)       // Creates an Slice with length and capacity of 5
//	s3 := g.NewSlice[int](3, 10)   // Creates an Slice with length of 3 and capacity of 10
func NewSlice[T any](size ...Int) Slice[T] {
	var (
		length   Int
		capacity Int
	)

	switch {
	case len(size) > 1:
		length, capacity = size[0], size[1]
	case len(size) == 1:
		length, capacity = size[0], size[0]
	}

	return make(Slice[T], length, capacity)
}

// SliceOf creates a new generic slice containing the provided elements.
func SliceOf[T any](slice ...T) Slice[T] { return slice }

// TransformSlice maps a plain Go slice into a Slice through fn. It is the
// exported bridge between stdlib-shaped results and g containers, shared with
// the subpackages (rx uses it for its match groups).
func TransformSlice[T, U any](sl []T, fn func(T) U) Slice[U] {
	if len(sl) == 0 {
		return NewSlice[U]()
	}

	result := make(Slice[U], len(sl))
	for i, v := range sl {
		result[i] = fn(v)
	}

	return result
}

// Transform applies a transformation function to the Slice and returns the result.
func (sl Slice[T]) Transform[U any](fn func(Slice[T]) U) U { return fn(sl) }

// Iter returns an iterator (Seq[T]) for the Slice, allowing for sequential iteration
// over its elements. It is commonly used in combination with higher-order functions,
// such as 'ForEach', to perform operations on each element of the Slice.
//
// Returns:
//
// A Seq[T], which can be used for sequential iteration over the elements of the Slice.
//
// Example usage:
//
//	slice := g.Slice[int]{1, 2, 3, 4, 5}
//	iterator := slice.Iter()
//	iterator.ForEach(func(element int) {
//		// Perform some operation on each element
//		fmt.Println(element)
//	})
//
// The 'Iter' method provides a convenient way to traverse the elements of a Slice
// in a functional style, enabling operations like mapping or filtering.
func (sl Slice[T]) Iter() Seq[T] { return Seq[T](seqFromSlice(sl)) }

// IterReverse returns an iterator (Seq[T]) for the Slice that allows for sequential iteration
// over its elements in reverse order. This method is useful when you need to traverse the elements
// from the end to the beginning.
//
// Returns:
//
// A Seq[T], which can be used for sequential iteration over the elements of the Slice in reverse order.
//
// Example usage:
//
//	slice := g.Slice[int]{1, 2, 3, 4, 5}
//	iterator := slice.IterReverse()
//	iterator.ForEach(func(element int) {
//		// Perform some operation on each element in reverse order
//		fmt.Println(element)
//	})
//
// The 'IterReverse' method enhances the functionality of the Slice by providing an alternative
// way to iterate through its elements, enhancing flexibility in how data within a Slice is accessed and manipulated.
func (sl Slice[T]) IterReverse() Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range slices.Backward(sl) {
			if !yield(v) {
				return
			}
		}
	}
}

// Fill fills the slice with the specified value.
// This function is useful when you want to create an Slice with all elements having the same
// value.
// This method modifies the original slice in place.
//
// Parameters:
//
// - val T: The value to fill the Slice with.
//
// Example usage:
//
//	slice := g.Slice[int]{0, 0, 0}
//	slice.Fill(5)
//
// The modified slice will now contain: 5, 5, 5.
func (sl Slice[T]) Fill(val T) {
	if len(sl) == 0 {
		return
	}

	if len(sl) > 32 {
		sl[0] = val
		for i := 1; i < len(sl); i <<= 1 {
			copy(sl[i:], sl[:i])
		}
	} else {
		for i := range sl {
			sl[i] = val
		}
	}
}

// Index returns the index of the first occurrence of the specified value in the slice, or -1 if
// not found.
func (sl Slice[T]) Index(val T) Int {
	if f.IsComparable[T]() && reflect.TypeFor[T]().Kind() != reflect.Interface {
		target := any(val)

		for i, v := range sl {
			if any(v) == target {
				return Int(i)
			}
		}

		return -1
	}

	return sl.IndexBy(f.Eqd(val))
}

// IndexBy returns the index of the first element in the slice
// satisfying the predicate function provided by the user.
// It iterates through the slice and applies the predicate to each element.
// If the predicate returns true for an element, it returns the index of that element.
// If no such element is found, it returns -1.
func (sl Slice[T]) IndexBy(fn func(t T) bool) Int { return Int(slices.IndexFunc(sl, fn)) }

// Insert inserts values at the specified index in the slice and modifies the original
// slice.
//
// Panics if the index is out of range. A negative index counts from the end of
// the slice; i == Len() appends at the end.
//
// Parameters:
//
// - i Int: The index at which to insert the new values.
//
// - values ...T: A variadic list of values to insert at the specified index.
//
// Example usage:
//
//	slice := g.Slice[string]{"a", "b", "c", "d"}
//	slice.Insert(2, "e", "f")
//
// The resulting slice will be: ["a", "b", "e", "f", "c", "d"].
func (sl *Slice[T]) Insert(i Int, values ...T) {
	if sl.IsEmpty() {
		if i != 0 {
			boundpanic(i, 0)
		}

		sl.Push(values...)
		return
	}

	sl.Replace(i, i, values...)
}

// Replace replaces the elements of sl[i:j] with the given values,
// and modifies the original slice in place. Replace panics if sl[i:j]
// is not a valid slice of sl.
//
// Parameters:
//
// - i Int: The starting index of the slice to be replaced.
//
// - j Int: The ending index of the slice to be replaced.
//
// - values ...T: A variadic list of values to replace the existing slice.
//
// Example usage:
//
//	slice := g.Slice[string]{"a", "b", "c", "d"}
//	slice.Replace(1, 3, "e", "f")
//
// After the Replace operation, the resulting slice will be: ["a", "e", "f", "d"].
func (sl *Slice[T]) Replace(i, j Int, values ...T) {
	ii, ok := sl.boundReplace(i)
	if !ok {
		boundpanic(i, len(*sl))
	}

	jj, ok := sl.boundReplace(j)
	if !ok {
		boundpanic(j, len(*sl))
	}

	i, j = ii, jj

	if i > j {
		boundpanic(j, len(*sl))
	}

	oldLen := sl.Len()
	removedCount := j - i
	addedCount := Int(len(values))
	newLen := oldLen - removedCount + addedCount

	if i == j {
		if addedCount == 0 {
			return
		}

		if newLen > sl.Cap() {
			newSlice := make(Slice[T], newLen)
			copy(newSlice[:i], (*sl)[:i])
			copy(newSlice[i:i+addedCount], values)
			copy(newSlice[i+addedCount:], (*sl)[i:])
			*sl = newSlice
		} else {
			*sl = (*sl)[:newLen]
			copy((*sl)[i+addedCount:], (*sl)[i:oldLen])
			copy((*sl)[i:], values)
		}

		return
	}

	if newLen > sl.Cap() {
		newSlice := make(Slice[T], newLen)
		copy(newSlice[:i], (*sl)[:i])
		copy(newSlice[i:i+addedCount], values)
		copy(newSlice[i+addedCount:], (*sl)[j:])
		*sl = newSlice
	} else {
		if newLen != oldLen {
			*sl = (*sl)[:newLen]
		}

		if addedCount != removedCount {
			copy((*sl)[i+addedCount:], (*sl)[j:oldLen])
		}

		copy((*sl)[i:], values)
	}
}

// Get returns the element at the given index, handling negative indices as counting from the end
// of the slice.
func (sl Slice[T]) Get(index Int) Option[T] {
	i, ok := sl.bound(index)
	if !ok {
		return None[T]()
	}

	return Some(sl[i])
}

// Reverse reverses the order of the elements in the slice.
// This method modifies the original slice in place.
//
// Example usage:
//
// slice := g.Slice[int]{1, 2, 3, 4, 5}
// slice.Reverse()
// fmt.Println(slice)
//
// Output: [5 4 3 2 1].
func (sl Slice[T]) Reverse() { slices.Reverse(sl) }

// SortBy sorts the elements in the slice using the provided comparison function.
// It modifies the original slice in place.
//
// The comparison function should return:
//   - cmp.Less if a should come before b
//   - cmp.Greater if a should come after b
//   - cmp.Equal if a and b are considered equal
//
// The sort is not guaranteed to be stable: the relative order of elements that
// compare equal may change.
//
// Parameters:
//
// - fn func(a, b T) cmp.Ordering: A comparison function reporting the ordering of a relative to b.
//
// Example usage:
//
// sl := NewSlice[int](1, 5, 3, 2, 4)
// sl.SortBy(func(a, b int) cmp.Ordering { return cmp.Cmp(a, b) }) // sorts in ascending order.
func (sl Slice[T]) SortBy(fn func(a, b T) cmp.Ordering) {
	slices.SortFunc(sl, func(a, b T) int { return int(fn(a, b)) })
}

// IsSortedBy checks if the slice is sorted according to the provided comparison function.
//
// The function takes a custom comparison function as an argument and checks if the elements
// are sorted according to the provided logic.
//
// Parameters:
//
// - fn func(a, b T) cmp.Ordering: A comparison function that defines the sort order.
//
// Returns:
//
// - bool: true if the slice is sorted according to the comparison function, false otherwise.
//
// Example usage:
//
//	sl := g.SliceOf(1, 2, 3, 4, 5)
//	sorted := sl.IsSortedBy(func(a, b int) cmp.Ordering { return cmp.Cmp(a, b) }) // returns true
func (sl Slice[T]) IsSortedBy(fn func(a, b T) cmp.Ordering) bool {
	if len(sl) <= 1 {
		return true
	}

	for i := 1; i < len(sl); i++ {
		if fn(sl[i-1], sl[i]).IsGt() {
			return false
		}
	}

	return true
}

// BinarySearch searches a sorted slice for value using the comparator fn and
// returns Some(index) if an equal element is found, or None otherwise. The slice
// must be sorted in ascending order according to fn (see IsSortedBy).
func (sl Slice[T]) BinarySearch(value T, fn func(a, b T) cmp.Ordering) Option[Int] {
	if i, ok := slices.BinarySearchFunc(sl, value, func(a, b T) int { return int(fn(a, b)) }); ok {
		return Some(Int(i))
	}

	return None[Int]()
}

// PartitionPoint returns the index of the first element for which pred returns
// false, assuming the slice is partitioned so that all elements satisfying pred
// come first. If pred is true for every element, it returns the slice length.
// Runs in O(log n).
func (sl Slice[T]) PartitionPoint(pred func(T) bool) Int {
	lo, hi := Int(0), sl.Len()
	for lo < hi {
		mid := (lo + hi) / 2
		if pred(sl[mid]) {
			lo = mid + 1
		} else {
			hi = mid
		}
	}

	return lo
}

// Retain keeps only the elements for which fn returns true, removing the rest
// in place while preserving order. It is the in-place counterpart of Deque.Retain.
func (sl *Slice[T]) Retain(fn func(T) bool) {
	*sl = slices.DeleteFunc(*sl, func(v T) bool { return !fn(v) })
}

// DedupBy removes consecutive elements considered equal by eq, keeping the first
// of each run, in place. It is the eager, in-place counterpart of the lazy
// Seq.Dedup. Only adjacent duplicates are removed, so sort first for a
// global dedup.
func (sl *Slice[T]) DedupBy(eq func(a, b T) bool) {
	*sl = slices.CompactFunc(*sl, eq)
}

// Join joins the elements in the slice into a single String, separated by the provided separator (if any).
func (sl Slice[T]) Join(sep ...T) String {
	if sl.IsEmpty() {
		return ""
	}

	if s, ok := any(sl).(Slice[Bytes]); ok {
		if len(s) == 0 {
			return ""
		}

		var separator Bytes
		if len(sep) != 0 {
			separator, _ = any(sep[0]).(Bytes)
		}

		total := len(separator) * (len(s) - 1)
		for _, value := range s {
			total += len(value)
		}

		var builder Builder
		builder.Grow(Int(total))
		for i, value := range s {
			if i > 0 {
				builder.Write(separator)
			}
			builder.Write(value)
		}

		return builder.String()
	}

	if s, ok := any(sl).(Slice[String]); ok {
		var separator string
		if len(sep) != 0 {
			if sepStr, ok := any(sep[0]).(String); ok {
				separator = sepStr.Std()
			} else {
				separator = fmt.Sprint(sep[0])
			}
		}

		total := len(separator) * (len(s) - 1)
		for _, str := range s {
			total += len(str)
		}

		var b strings.Builder
		b.Grow(total)

		for i, str := range s {
			if i > 0 {
				b.WriteString(separator)
			}

			b.WriteString(str.Std())
		}

		return String(b.String())
	}

	var separator string
	if len(sep) != 0 {
		separator = fmt.Sprint(sep[0])
	}

	var b strings.Builder

	for i, v := range sl {
		if i > 0 {
			b.WriteString(separator)
		}

		fmt.Fprint(&b, v)
	}

	return String(b.String())
}

// SubSlice returns a new slice containing elements from the current slice between the specified start
// and end indices, with an optional step parameter to define the increment between elements.
// Negative start or end indices count from the end of the slice.
//
// Panics if start or end is out of range after negative-index resolution.
//
// Parameters:
//
// - start (Int): The start index of the range.
//
// - end (Int): The end index of the range.
//
// - step (Int, optional): The increment between elements. Defaults to 1 if not provided.
// If negative, the slice is traversed in reverse order.
//
// Returns:
//
// - Slice[T]: A new slice containing elements from the current slice between the start and end
// indices, with the specified step.
//
// Example usage:
//
//	slice := g.Slice[int]{1, 2, 3, 4, 5, 6, 7, 8, 9}
//	subSlice := slice.SubSlice(1, 7, 2) // Extracts elements 2, 4, 6
//	fmt.Println(subSlice)
//
// Output: [2 4 6].
func (sl Slice[T]) SubSlice(start, end Int, step ...Int) Slice[T] {
	if sl.IsEmpty() {
		return NewSlice[T]()
	}

	_step := Int(1)
	if len(step) > 0 {
		_step = Int(step[0])
	}

	ii, ok := sl.boundsub(start)
	if !ok {
		boundpanic(start, len(sl))
	}

	jj, ok := sl.boundsub(end)
	if !ok {
		boundpanic(end, len(sl))
	}

	start, end = ii, jj

	// For a negative step the iteration starts AT start and moves down, so a
	// start clamped to len(sl) must begin at the last element (s[100:0:-1]
	// starts at the final index, not one past it).
	if _step < 0 && start == sl.Len() {
		start--
	}

	if _step == 1 {
		if start >= end {
			return NewSlice[T]()
		}

		return slices.Clone(sl[start:end])
	}

	if (start >= end && _step > 0) || (start <= end && _step < 0) || _step == 0 {
		return NewSlice[T]()
	}

	var resultSize Int
	if _step > 0 {
		resultSize = (end - start + _step - 1) / _step
	} else {
		resultSize = (start - end + (-_step) - 1) / (-_step)
	}

	slice := make(Slice[T], 0, resultSize)

	if _step > 0 {
		for i := start; i < end; i += _step {
			slice = append(slice, sl[i])
		}
	} else {
		for i := start; i > end; i += _step {
			slice = append(slice, sl[i])
		}
	}

	return slice
}

// Clone returns a copy of the slice.
func (sl Slice[T]) Clone() Slice[T] {
	if sl.IsEmpty() {
		return NewSlice[T]()
	}

	return slices.Clone(sl)
}

// LastIndex returns the last index of the slice.
func (sl Slice[T]) LastIndex() Int {
	if !sl.IsEmpty() {
		return sl.Len() - 1
	}

	return 0
}

// Eq returns true if the slice is equal to the provided other slice.
func (sl Slice[T]) Eq(other Slice[T]) bool {
	if len(sl) != len(other) {
		return false
	}

	if f.IsComparable[T]() && reflect.TypeFor[T]().Kind() != reflect.Interface {
		for i, v := range sl {
			if any(v) != any(other[i]) {
				return false
			}
		}

		return true
	}

	return sl.EqBy(other, func(x, y T) bool { return reflect.DeepEqual(x, y) })
}

// EqBy reports whether two slices are equal using an equality
// function on each pair of elements. If the lengths are different,
// EqBy returns false. Otherwise, the elements are compared in
// increasing index order, and the comparison stops at the first index
// for which eq returns false.
func (sl Slice[T]) EqBy(other Slice[T], fn func(x, y T) bool) bool {
	return slices.EqualFunc(sl, other, fn)
}

// String returns a string representation of the slice.
func (sl Slice[T]) String() string {
	if len(sl) == 0 {
		return "Slice[]"
	}

	var b Builder
	b.Grow(Int(len(sl)) * 8)
	b.WriteString("Slice[")

	for i, v := range sl {
		if i > 0 {
			b.WriteString(", ")
		}

		fmt.Fprint(&b, v)
	}

	b.WriteString("]")

	return b.String().Std()
}

// Append appends the provided elements to the slice and returns the modified slice.
func (sl Slice[T]) Append(elems ...T) Slice[T] { return append(sl, elems...) }

// AppendUnique appends unique elements from the provided arguments to the current slice.
//
// The function iterates over the provided elements and checks if they are already present
// in the slice. If an element is not already present, it is appended to the slice. The
// resulting slice is returned, containing the unique elements from both the original
// slice and the provided elements.
//
// Parameters:
//
// - elems (...T): A variadic list of elements to be appended to the slice.
//
// Returns:
//
// - Slice[T]: A new slice containing the unique elements from both the original slice
// and the provided elements.
//
// Example usage:
//
//	slice := g.Slice[int]{1, 2, 3, 4, 5}
//	slice = slice.AppendUnique(3, 4, 5, 6, 7)
//	fmt.Println(slice)
//
// Output: [1 2 3 4 5 6 7].
func (sl Slice[T]) AppendUnique(elems ...T) Slice[T] {
	if f.IsComparable[T]() && reflect.TypeFor[T]().Kind() != reflect.Interface {
		set := make(Set[any], len(sl)+len(elems))
		for _, v := range sl {
			set[v] = Unit{}
		}

		for _, elem := range elems {
			if !set.Contains(elem) {
				sl = append(sl, elem)
				set.Insert(elem)
			}
		}

		return sl
	}

	for _, elem := range elems {
		if !sl.Contains(elem) {
			sl = append(sl, elem)
		}
	}

	return sl
}

// Push appends the provided elements to the slice and modifies the original slice.
func (sl *Slice[T]) Push(elems ...T) { *sl = append(*sl, elems...) }

// PushUnique appends unique elements from the provided arguments to the current slice.
//
// The function iterates over the provided elements and checks if they are already present
// in the slice. If an element is not already present, it is appended to the slice.
//
// Parameters:
//
// - elems (...T): A variadic list of elements to be appended to the slice.
//
// Example usage:
//
//	slice := g.Slice[int]{1, 2, 3, 4, 5}
//	slice.PushUnique(3, 4, 5, 6, 7)
//	fmt.Println(slice)
//
// Output: [1 2 3 4 5 6 7].
func (sl *Slice[T]) PushUnique(elems ...T) {
	if f.IsComparable[T]() && reflect.TypeFor[T]().Kind() != reflect.Interface {
		set := make(Set[any], len(*sl)+len(elems))
		for _, v := range *sl {
			set[v] = Unit{}
		}

		for _, elem := range elems {
			if !set.Contains(elem) {
				sl.Push(elem)
				set.Insert(elem)
			}
		}

		return
	}

	for _, elem := range elems {
		if !sl.Contains(elem) {
			sl.Push(elem)
		}
	}
}

// Cap returns the capacity of the Slice.
func (sl Slice[T]) Cap() Int { return Int(cap(sl)) }

// Contains returns true if the slice contains the provided value.
func (sl Slice[T]) Contains(val T) bool { return sl.Index(val) >= 0 }

// ContainsBy returns true if the slice contains an element that satisfies the provided function fn, false otherwise.
func (sl Slice[T]) ContainsBy(fn func(t T) bool) bool { return sl.IndexBy(fn) >= 0 }

// ContainsAny checks if the Slice contains any element from another Slice.
func (sl Slice[T]) ContainsAny(values ...T) bool {
	if sl.IsEmpty() || len(values) == 0 {
		return false
	}

	if f.IsComparable[T]() && reflect.TypeFor[T]().Kind() != reflect.Interface {
		set := make(Set[any], len(sl))
		for _, v := range sl {
			set[v] = Unit{}
		}

		for _, v := range values {
			if set.Contains(v) {
				return true
			}
		}

		return false
	}

	return slices.ContainsFunc(values, sl.Contains)
}

// ContainsAll checks if the Slice contains all elements from another Slice.
func (sl Slice[T]) ContainsAll(values ...T) bool {
	if sl.IsEmpty() || len(values) == 0 {
		return len(values) == 0
	}

	if f.IsComparable[T]() && reflect.TypeFor[T]().Kind() != reflect.Interface {
		set := make(Set[any], len(sl))
		for _, v := range sl {
			set[v] = Unit{}
		}

		for _, v := range values {
			if !set.Contains(v) {
				return false
			}
		}

		return true
	}

	for _, v := range values {
		if !sl.Contains(v) {
			return false
		}
	}

	return true
}

// Remove removes and returns the element at the specified index.
// Returns None if index is out of bounds.
// Negative indices are supported: -1 refers to the last element, etc.
func (sl *Slice[T]) Remove(index Int) Option[T] {
	if sl.IsEmpty() {
		return None[T]()
	}

	length := sl.Len()

	if index < 0 {
		index += length
	}

	if index < 0 || index >= length {
		return None[T]()
	}

	value := (*sl)[index]
	*sl = append((*sl)[:index], (*sl)[index+1:]...)

	return Some(value)
}

// IsEmpty returns true if the slice is empty.
func (sl Slice[T]) IsEmpty() bool { return len(sl) == 0 }

// First returns the first element of the slice.
func (sl Slice[T]) First() Option[T] { return sl.Get(0) }

// Last returns the last element of the slice.
func (sl Slice[T]) Last() Option[T] { return sl.Get(-1) }

// Ne returns true if the slice is not equal to the provided other slice.
func (sl Slice[T]) Ne(other Slice[T]) bool { return !sl.Eq(other) }

// NeBy reports whether two slices are not equal using an inequality
// function on each pair of elements. If the lengths are different,
// NeBy returns true. Otherwise, the elements are compared in
// increasing index order, and the comparison stops at the first index
// for which fn returns true.
func (sl Slice[T]) NeBy(other Slice[T], fn func(x, y T) bool) bool { return !sl.EqBy(other, fn) }

// Pop removes and returns the last element of the slice.
// It mutates the original slice by removing the last element.
// It returns None if the slice is empty.
func (sl *Slice[T]) Pop() Option[T] {
	if sl.Len() == 0 {
		return None[T]()
	}

	last := (*sl)[sl.Len()-1]
	*sl = (*sl)[:sl.Len()-1]

	return Some(last)
}

// Set sets the value at the specified index in the slice and returns the previous
// value wrapped in Some. If the index is out of bounds, the slice is left unchanged
// and None is returned, mirroring Deque.Set.
// This method modifies the original slice in place. Negative indices count from the end.
//
// Parameters:
//
// - index (Int): The index at which to set the new value.
// - val (T): The new value to be set at the specified index.
//
// Returns:
//
// - Option[T]: The previous value at the index, or None if the index is out of bounds.
//
// Example usage:
//
// slice := g.Slice[int]{1, 2, 3, 4, 5}
// old := slice.Set(2, 99) // Some(3)
// fmt.Println(slice)
//
// Output: [1 2 99 4 5].
func (sl Slice[T]) Set(index Int, val T) Option[T] {
	i, ok := sl.bound(index)
	if !ok {
		return None[T]()
	}

	old := sl[i]
	sl[i] = val

	return Some(old)
}

// Len returns the length of the slice.
func (sl Slice[T]) Len() Int { return Int(len(sl)) }

// Swap swaps the elements at the specified indices in the slice.
// This method modifies the original slice in place.
//
// Panics if either index is out of range. A negative index counts
// from the end of the slice.
//
// Parameters:
//
// - i (Int): The index of the first element to be swapped.
//
// - j (Int): The index of the second element to be swapped.
//
// Example usage:
//
// slice := g.Slice[int]{1, 2, 3, 4, 5}
// slice.Swap(1, 3)
// fmt.Println(slice)
//
// Output: [1 4 3 2 5].
func (sl Slice[T]) Swap(i, j Int) {
	ii, ok := sl.bound(i)
	if !ok {
		boundpanic(i, len(sl))
	}

	jj, ok := sl.bound(j)
	if !ok {
		boundpanic(j, len(sl))
	}

	sl[ii], sl[jj] = sl[jj], sl[ii]
}

// Grow increases the slice's capacity, if necessary, to guarantee space for
// another n elements. After Grow(n), at least n elements can be appended
// to the slice without another allocation. If n is negative or too large to
// allocate the memory, Grow panics.
func (sl Slice[T]) Grow(n Int) Slice[T] { return slices.Grow(sl, n.Std()) }

// Clip removes unused capacity from the slice.
func (sl Slice[T]) Clip() Slice[T] { return slices.Clip(sl) }

// Std returns a new slice with the same elements as the Slice[T].
func (sl Slice[T]) Std() []T { return sl }

// Print writes the elements of the Slice to the standard output (console)
// and returns the Slice unchanged.
func (sl Slice[T]) Print() Slice[T] { fmt.Print(sl); return sl }

// Println writes the elements of the Slice to the standard output (console) with a newline
// and returns the Slice unchanged.
func (sl Slice[T]) Println() Slice[T] { fmt.Println(sl); return sl }

// Unpack assigns values of the slice's elements to the variables passed as pointers.
// If the number of variables passed is greater than the length of the slice,
// the function ignores the extra variables.
//
// Parameters:
//
// - vars (...*T): Pointers to variables where the values of the slice's elements will be stored.
//
// Example:
//
//	slice := g.Slice[int]{1, 2, 3, 4, 5}
//	var a, b, c int
//	slice.Unpack(&a, &b, &c)
//	fmt.Println(a, b, c) // Output: 1 2 3
func (sl Slice[T]) Unpack(vars ...*T) {
	n := min(len(sl), len(vars))

	for i := range n {
		if vars[i] != nil {
			*vars[i] = sl[i]
		}
	}
}

func (sl Slice[T]) bound(i Int) (Int, bool) {
	n := sl.Len()
	if n == 0 {
		return 0, false
	}

	if i < 0 {
		i += n
	}

	if i >= n || i < 0 {
		return 0, false
	}

	return i, true
}

// boundReplace resolves an index for Replace, which may legitimately target
// i == len(sl) to append at the end. Element accessors use bound (strict i < n).
func (sl Slice[T]) boundReplace(i Int) (Int, bool) {
	n := sl.Len()
	if n == 0 {
		return 0, false
	}

	if i < 0 {
		i += n
	}

	if i > n || i < 0 {
		return 0, false
	}

	return i, true
}

func (sl Slice[T]) boundsub(i Int) (Int, bool) {
	n := sl.Len()
	if n == 0 {
		return 0, false
	}

	if i < 0 {
		i += n
	}

	if i > n || i < -1 {
		return 0, false
	}

	return i, true
}

func boundpanic(index Int, length int) {
	panic(fmt.Sprintf("runtime error: slice bounds out of range [%d] with length %d", index, length))
}
