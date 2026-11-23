// Package f provides predicate helpers and combinators (f.Eq, f.Gt, f.Contains, ...)
// for use with iterator methods such as Filter and Exclude.
package f

import (
	"cmp"
	"reflect"
	"regexp"
	"strings"

	"github.com/enetx/g/constraints"
)

// Id returns its argument unchanged. It is the identity function, useful
// wherever a transform is required but the value should pass through as-is;
// it pairs with CounterBy for occurrence counting:
//
//	words.Iter().CounterBy(f.Id)
func Id[T any](t T) T { return t }

// IsComparable reports whether the type T is comparable at the type level.
// It uses reflect.TypeFor[T]() without requiring a value, making it suitable
// for use in generic code where the type is known at compile time.
// The result is determined solely by the type and does not depend on any runtime value.
func IsComparable[T any]() bool { return reflect.TypeFor[T]().Comparable() }

// IsComparableValue reports whether the concrete value v is comparable.
// Unlike IsComparable, which checks at the type level, IsComparableValue
// inspects the actual runtime value, making it suitable for filtering
// or checking dynamic values of type 'any'.
func IsComparableValue(v any) bool { return reflect.ValueOf(v).Comparable() }

// IsZero is a generic function designed to check if a value is considered zero.
func IsZero[T cmp.Ordered](v T) bool { return v == *new(T) }

// IsEven is a generic function that checks if the provided integer is even.
func IsEven[T constraints.Integer](i T) bool { return i%2 == 0 }

// IsOdd is a generic function that checks if the provided integer is odd.
func IsOdd[T constraints.Integer](i T) bool { return i%2 != 0 }

// Match returns a function that checks whether a string or []byte matches a given regular expression.
func Match[T ~string | ~[]byte](t *regexp.Regexp) func(T) bool {
	return func(s T) bool {
		if reflect.TypeFor[T]().Kind() == reflect.String {
			return t.MatchString(reflect.ValueOf(s).String())
		}

		return t.Match(reflect.ValueOf(s).Bytes())
	}
}

// Contains returns a function that checks whether a string or []byte contains a given substring.
func Contains[T ~string | ~[]byte](t T) func(T) bool {
	target := string(t)
	return func(s T) bool {
		return strings.Contains(string(s), target)
	}
}

// ContainsAnyChars returns a function that checks whether a string contains any of the characters from a given set.
func ContainsAnyChars[T ~string | ~[]byte](t T) func(T) bool {
	chars := string(t)
	return func(s T) bool {
		return strings.ContainsAny(string(s), chars)
	}
}

// StartsWith returns a function that checks whether a string starts with a given prefix.
func StartsWith[T ~string | ~[]byte](t T) func(T) bool {
	prefix := string(t)
	return func(s T) bool {
		return strings.HasPrefix(string(s), prefix)
	}
}

// EndsWith returns a function that checks whether a string ends with a given suffix.
func EndsWith[T ~string | ~[]byte](t T) func(T) bool {
	suffix := string(t)
	return func(s T) bool {
		return strings.HasSuffix(string(s), suffix)
	}
}

// Eq returns a comparison function that evaluates to true when a value is equal to the provided threshold.
func Eq[T comparable](t T) func(T) bool {
	return func(s T) bool {
		return s == t
	}
}

// Ne returns a comparison function that evaluates to true when a value is not equal to the provided threshold.
func Ne[T comparable](t T) func(T) bool {
	return func(s T) bool {
		return s != t
	}
}

// Eqd returns a comparison function that evaluates to true when a value is deeply equal to the provided threshold.
func Eqd[T any](t T) func(T) bool {
	return func(s T) bool {
		return reflect.DeepEqual(t, s)
	}
}

// Ned returns a comparison function that evaluates to true when a value is not deeply equal to the provided threshold.
func Ned[T any](t T) func(T) bool {
	return func(s T) bool {
		return !reflect.DeepEqual(t, s)
	}
}

// Gt returns a comparison function that evaluates to true when a value is greater than the threshold.
func Gt[T cmp.Ordered](t T) func(T) bool {
	return func(s T) bool {
		return s > t
	}
}

// Gte returns a comparison function that evaluates to true when a value is greater than or equal to the threshold.
func Gte[T cmp.Ordered](t T) func(T) bool {
	return func(s T) bool {
		return s >= t
	}
}

// Lt returns a comparison function that evaluates to true when a value is less than the threshold.
func Lt[T cmp.Ordered](t T) func(T) bool {
	return func(s T) bool {
		return s < t
	}
}

// Lte returns a comparison function that evaluates to true when a value is less than or equal to the threshold.
func Lte[T cmp.Ordered](t T) func(T) bool {
	return func(s T) bool {
		return s <= t
	}
}

// Not returns a predicate that negates the result of the provided predicate.
func Not[T any](fn func(T) bool) func(T) bool {
	return func(s T) bool {
		return !fn(s)
	}
}

// And returns a predicate that evaluates to true only when all of the provided predicates do.
// It short-circuits on the first predicate that returns false. With no predicates it returns true.
func And[T any](fns ...func(T) bool) func(T) bool {
	return func(s T) bool {
		for _, fn := range fns {
			if !fn(s) {
				return false
			}
		}

		return true
	}
}

// Or returns a predicate that evaluates to true when any of the provided predicates does.
// It short-circuits on the first predicate that returns true. With no predicates it returns false.
func Or[T any](fns ...func(T) bool) func(T) bool {
	return func(s T) bool {
		for _, fn := range fns {
			if fn(s) {
				return true
			}
		}

		return false
	}
}
