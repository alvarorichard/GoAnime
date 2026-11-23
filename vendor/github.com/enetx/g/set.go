package g

import "fmt"

// Set is a generic alias for a set implemented using a map.
type Set[T comparable] map[T]Unit

// NewSet creates a new Set of the specified size or an empty Set if no size is provided.
func NewSet[T comparable](size ...Int) Set[T] {
	if len(size) > 0 {
		return make(Set[T], size[0])
	}

	return make(Set[T])
}

// SetOf creates a new generic set containing the provided elements.
func SetOf[T comparable](values ...T) Set[T] {
	set := make(Set[T], len(values))
	for _, v := range values {
		set[v] = Unit{}
	}

	return set
}

// Transform applies a transformation function to the Set and returns the result.
func (s Set[T]) Transform[U any](fn func(Set[T]) U) U { return fn(s) }

// Iter returns an iterator (Seq[T]) for the Set, allowing for sequential iteration
// over its elements. It is commonly used in combination with higher-order functions,
// such as 'ForEach' or 'Map', to perform operations on each element of the Set.
//
// Returns:
//
// A Seq[T], which can be used for sequential iteration over the elements of the Set.
//
// Example usage:
//
//	g.SetOf(1, 2, 3).Iter().ForEach(func(val int) {
//	    fmt.Println(val) // Replace this with the function logic you need.
//	})
//
// The 'Iter' method provides a convenient way to traverse the elements of a Set
// in a functional style, enabling operations like mapping or filtering.
func (s Set[T]) Iter() Seq[T] {
	return func(yield func(T) bool) {
		for v := range s {
			if !yield(v) {
				return
			}
		}
	}
}

// Insert adds the provided elements to the set.
func (s Set[T]) Insert(values ...T) {
	for _, v := range values {
		s[v] = Unit{}
	}
}

// Remove removes the specified value from the Set and returns true if it was present.
func (s Set[T]) Remove(v T) bool {
	if _, ok := s[v]; ok {
		delete(s, v)
		return true
	}

	return false
}

// Len returns the number of values in the Set.
func (s Set[T]) Len() Int { return Int(len(s)) }

// Contains checks if the Set contains the specified value.
func (s Set[T]) Contains(v T) bool {
	_, ok := s[v]
	return ok
}

// ContainsAny checks if the Set contains any of the provided values, matching
// the variadic shape of Slice.ContainsAny and String.ContainsAny.
func (s Set[T]) ContainsAny(values ...T) bool {
	for _, v := range values {
		if _, ok := s[v]; ok {
			return true
		}
	}

	return false
}

// ContainsAll checks if the Set contains all of the provided values, matching
// the variadic shape of Slice.ContainsAll and String.ContainsAll.
func (s Set[T]) ContainsAll(values ...T) bool {
	for _, v := range values {
		if _, ok := s[v]; !ok {
			return false
		}
	}

	return true
}

// Clone creates a new Set that is a copy of the original Set.
func (s Set[T]) Clone() Set[T] {
	if s.IsEmpty() {
		return NewSet[T]()
	}

	clone := make(Set[T], len(s))
	for k := range s {
		clone[k] = Unit{}
	}

	return clone
}

// Intersection returns the intersection of the current set and another set, i.e., elements
// present in both sets.
//
// Parameters:
//
// - other Set[T]: The other set to calculate the intersection with.
//
// Returns:
//
// - Set[T]: A new Set containing the intersection of the two sets.
//
// Example usage:
//
//	s1 := g.SetOf(1, 2, 3, 4, 5)
//	s2 := g.SetOf(4, 5, 6, 7, 8)
//	intersection := s1.Intersection(s2)
//
// The resulting intersection will be: [4, 5].
func (s Set[T]) Intersection(other Set[T]) Set[T] {
	small, big := s, other
	if len(big) < len(small) {
		small, big = big, small
	}

	result := make(Set[T], len(small))

	for v := range small {
		if big.Contains(v) {
			result[v] = Unit{}
		}
	}

	return result
}

// Difference returns the difference between the current set and another set,
// i.e., elements present in the current set but not in the other set.
//
// Parameters:
//
// - other Set[T]: The other set to calculate the difference with.
//
// Returns:
//
// - Set[T]: A new Set containing the difference between the two sets.
//
// Example usage:
//
//	s1 := g.SetOf(1, 2, 3, 4, 5)
//	s2 := g.SetOf(4, 5, 6, 7, 8)
//	diff := s1.Difference(s2)
//
// The resulting diff will be: [1, 2, 3].
func (s Set[T]) Difference(other Set[T]) Set[T] {
	result := make(Set[T], len(s))

	for v := range s {
		if !other.Contains(v) {
			result[v] = Unit{}
		}
	}

	return result
}

// Union returns a new set containing the unique elements of the current set and the provided
// other set.
//
// Parameters:
//
// - other Set[T]: The other set to create the union with.
//
// Returns:
//
// - Set[T]: A new Set containing the unique elements of the current set and the provided
// other set.
//
// Example usage:
//
//	s1 := g.SetOf(1, 2, 3)
//	s2 := g.SetOf(3, 4, 5)
//	union := s1.Union(s2)
//
// The resulting union set will be: [1, 2, 3, 4, 5].
func (s Set[T]) Union(other Set[T]) Set[T] {
	result := make(Set[T], len(s)+len(other))

	for v := range s {
		result[v] = Unit{}
	}

	for v := range other {
		result[v] = Unit{}
	}

	return result
}

// SymmetricDifference returns the symmetric difference between the current set and another
// set, i.e., elements present in either the current set or the other set but not in both.
//
// Parameters:
//
// - other Set[T]: The other set to calculate the symmetric difference with.
//
// Returns:
//
// - Set[T]: A new Set containing the symmetric difference between the two sets.
//
// Example usage:
//
//	s1 := g.SetOf(1, 2, 3, 4, 5)
//	s2 := g.SetOf(4, 5, 6, 7, 8)
//	symDiff := s1.SymmetricDifference(s2)
//
// The resulting symDiff will be: [1, 2, 3, 6, 7, 8].
func (s Set[T]) SymmetricDifference(other Set[T]) Set[T] {
	result := make(Set[T])

	for v := range s {
		if !other.Contains(v) {
			result[v] = Unit{}
		}
	}

	for v := range other {
		if !s.Contains(v) {
			result[v] = Unit{}
		}
	}

	return result
}

// Subset checks if the current set 's' is a subset of the provided 'other' set.
// A set 's' is a subset of 'other' if all elements of 's' are also elements of 'other'.
//
// Parameters:
//
// - other Set[T]: The other set to compare with.
//
// Returns:
//
// - bool: true if 's' is a subset of 'other', false otherwise.
//
// Example usage:
//
//	s1 := g.SetOf(1, 2, 3)
//	s2 := g.SetOf(1, 2, 3, 4, 5)
//	isSubset := s1.Subset(s2) // Returns true
func (s Set[T]) Subset(other Set[T]) bool {
	if len(s) > len(other) {
		return false
	}

	for v := range s {
		if _, ok := other[v]; !ok {
			return false
		}
	}

	return true
}

// Superset checks if the current set 's' is a superset of the provided 'other' set.
// A set 's' is a superset of 'other' if all elements of 'other' are also elements of 's'.
//
// Parameters:
//
// - other Set[T]: The other set to compare with.
//
// Returns:
//
// - bool: true if 's' is a superset of 'other', false otherwise.
//
// Example usage:
//
//	s1 := g.SetOf(1, 2, 3, 4, 5)
//	s2 := g.SetOf(1, 2, 3)
//	isSuperset := s1.Superset(s2) // Returns true
func (s Set[T]) Superset(other Set[T]) bool { return other.Subset(s) }

// Eq checks if two Sets are equal.
func (s Set[T]) Eq(other Set[T]) bool {
	if len(s) != len(other) {
		return false
	}

	for v := range other {
		if _, ok := s[v]; !ok {
			return false
		}
	}

	return true
}

// Ne checks if two Sets are not equal.
func (s Set[T]) Ne(other Set[T]) bool { return !s.Eq(other) }

// Clear removes all values from the Set.
func (s Set[T]) Clear() { clear(s) }

// IsEmpty checks if the Set is empty.
func (s Set[T]) IsEmpty() bool { return len(s) == 0 }

// String returns a string representation of the Set.
func (s Set[T]) String() string {
	if s.IsEmpty() {
		return "Set{}"
	}

	var b Builder
	b.Grow(Int(len(s)) * 8)
	b.WriteString("Set{")

	first := true
	for v := range s {
		if !first {
			b.WriteString(", ")
		}

		first = false
		fmt.Fprint(&b, v)
	}

	b.WriteString("}")

	return b.String().Std()
}

// Disjoint reports whether the set has no elements in common with other.
// It is the complement of ContainsAny.
func (s Set[T]) Disjoint(other Set[T]) bool {
	small, big := s, other
	if len(big) < len(small) {
		small, big = big, small
	}

	for v := range small {
		if _, ok := big[v]; ok {
			return false
		}
	}

	return true
}

// Print writes the elements of the Set to the standard output (console)
// and returns the Set unchanged.
func (s Set[T]) Print() Set[T] { fmt.Print(s); return s }

// Println writes the elements of the Set to the standard output (console) with a newline
// and returns the Set unchanged.
func (s Set[T]) Println() Set[T] { fmt.Println(s); return s }
