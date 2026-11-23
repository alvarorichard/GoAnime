// Package cmp provides three-way comparison primitives (Ordering, Cmp, Reverse)
// used by the sorting, heap and min/max APIs of the g library.
package cmp

import "cmp"

// Cmp compares two ordered values and returns the result as an Ordering value.
func Cmp[T cmp.Ordered](x, y T) Ordering { return Ordering(cmp.Compare(x, y)) }

// Reverse compares two ordered values and returns the inverse of Cmp as an Ordering.
func Reverse[T cmp.Ordered](x, y T) Ordering { return Ordering(cmp.Compare(x, y)).Reverse() }
