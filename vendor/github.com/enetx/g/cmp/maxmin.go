package cmp

import "cmp"

// Min returns the minimum value among the given values. The values must be of a type that implements
// the cmp.Ordered interface for comparison.
func Min[T cmp.Ordered](t ...T) T { return MinBy(Cmp, t...) }

// Max returns the maximum value among the given values. The values must be of a type that implements
// the cmp.Ordered interface for comparison.
func Max[T cmp.Ordered](t ...T) T { return MaxBy(Cmp, t...) }

// MinBy finds the minimum value in the collection t according to the provided comparison function fn.
// It returns the minimum value found.
func MinBy[T any](fn func(x, y T) Ordering, t ...T) T {
	if len(t) == 0 {
		return *new(T)
	}

	m := t[0]

	for _, v := range t[1:] {
		if fn(v, m).IsLt() {
			m = v
		}
	}

	return m
}

// MaxBy finds the maximum value in the collection t according to the provided comparison function fn.
// It returns the maximum value found.
func MaxBy[T any](fn func(x, y T) Ordering, t ...T) T {
	if len(t) == 0 {
		return *new(T)
	}

	m := t[0]

	for _, v := range t[1:] {
		if fn(v, m).IsGt() {
			m = v
		}
	}

	return m
}
