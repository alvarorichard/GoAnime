package g

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Option is a generic struct for representing an optional value.
type Option[T any] struct {
	v      T    // Value.
	isSome bool // Indicator of value presence.
}

// Unit represents an empty value.
// Used in contexts where a function needs to return "something" but
// the actual value doesn't matter, only success/failure status.
type Unit struct{}

// Some creates an Option containing a value.
func Some[T any](value T) Option[T] { return Option[T]{v: value, isSome: true} }

// None creates an Option representing no value.
func None[T any]() Option[T] { return Option[T]{isSome: false} }

// OptionOf creates an Option[T] based on the provided value and a boolean flag.
// If ok is true, it returns Some(value).
// Otherwise, it returns None.
func OptionOf[T any](value T, ok bool) Option[T] {
	if ok {
		return Some(value)
	}

	return None[T]()
}

// OptionFromPtr converts a pointer into an Option.
// Returns None if ptr is nil.
func OptionFromPtr[T any](ptr *T) Option[T] {
	if ptr == nil {
		return None[T]()
	}

	return Some(*ptr)
}

// Some returns the contained value of the Option.
//
// WARNING: If the Option is None, this method will return the zero value
// for type T. Always check IsSome() before calling this method, or use safer alternatives
// like Unwrap(), or UnwrapOr().
func (o Option[T]) Some() T { return o.v }

// IsSome returns true if the Option contains a value.
func (o Option[T]) IsSome() bool { return o.isSome }

// IsNone returns true if the Option represents no value.
func (o Option[T]) IsNone() bool { return !o.isSome }

// optionPanic prints caller information for msg to stderr and then panics with msg.
// skip is the number of stack frames between the original caller and runtime.Caller,
// so that the reported file:line and function point at the user's call site rather
// than at this helper.
func optionPanic(skip int, msg string) {
	if pc, file, line, ok := runtime.Caller(skip); ok {
		out := fmt.Sprintf("[%s:%d] [%s] %s", filepath.Base(file), line, runtime.FuncForPC(pc).Name(), msg)
		fmt.Fprintln(os.Stderr, out)
	}

	panic(msg)
}

// Unwrap returns the value held in the Option. If the Option is None, it panics.
func (o Option[T]) Unwrap() T {
	if o.isSome {
		return o.v
	}

	optionPanic(2, "called Option.Unwrap() on a None value")
	panic("unreachable")
}

// UnwrapOr returns the value held in the Option. If the Option is None, it returns the provided default value.
func (o Option[T]) UnwrapOr(value T) T {
	if o.isSome {
		return o.v
	}

	return value
}

// UnwrapOrDefault returns the contained value if Some; otherwise returns the zero value for T.
func (o Option[T]) UnwrapOrDefault() T {
	if o.isSome {
		return o.v
	}

	var zero T
	return zero
}

// Expect returns the value held in the Option. If the Option is None, it panics with the provided message.
func (o Option[T]) Expect(msg string) T {
	if o.isSome {
		return o.v
	}

	optionPanic(2, fmt.Sprintf("Expect() failed: %s", msg))
	panic("unreachable")
}

// Then applies the function fn to the value inside the Option and returns the resulting Option.
// If the Option is None, fn is not called and None is returned.
// The result type may differ from the input type.
func (o Option[T]) Then[U any](fn func(T) Option[U]) Option[U] {
	if o.isSome {
		return fn(o.v)
	}

	return None[U]()
}

// ThenOf applies fn to the value inside the Option and returns a new Option based
// on the returned (U, bool) comma-ok tuple: ok=true yields Some(value), ok=false
// yields None. If the Option is None, fn is not called and None is returned.
// It mirrors Result.ThenOf for the comma-ok idiom.
func (o Option[T]) ThenOf[U any](fn func(T) (U, bool)) Option[U] {
	if o.isSome {
		return OptionOf(fn(o.v))
	}

	return None[U]()
}

// Map applies the function fn to the value inside the Option and returns a new Option
// holding the transformed value. If the Option is None, fn is not called and None is returned.
// Unlike Then, fn returns a plain U (always Some on a Some input) rather than an Option[U].
func (o Option[T]) Map[U any](fn func(T) U) Option[U] {
	if o.isSome {
		return Some(fn(o.v))
	}

	return None[U]()
}

// MapOr applies fn to the contained value if Some and returns the result;
// otherwise returns the provided default value.
func (o Option[T]) MapOr[U any](def U, fn func(T) U) U {
	if o.isSome {
		return fn(o.v)
	}

	return def
}

// MapOrElse applies fn to the contained value if Some and returns the result;
// otherwise computes and returns the default lazily via defFn.
func (o Option[T]) MapOrElse[U any](defFn func() U, fn func(T) U) U {
	if o.isSome {
		return fn(o.v)
	}

	return defFn()
}

// Inspect calls fn with the contained value if the Option is Some, then returns
// the Option unchanged. If the Option is None, fn is not called. It is intended
// for side effects (logging, debugging) within a chain and never mutates the Option.
func (o Option[T]) Inspect(fn func(T)) Option[T] {
	if o.isSome {
		fn(o.v)
	}

	return o
}

// Filter returns Some(value) if the Option is Some and the predicate returns true.
// Otherwise, it returns None.
func (o Option[T]) Filter(pred func(T) bool) Option[T] {
	if o.isSome && pred(o.v) {
		return o
	}

	return None[T]()
}

// Or returns the Option if it contains a value.
// Otherwise, it returns the provided alternative Option.
func (o Option[T]) Or(other Option[T]) Option[T] {
	if o.isSome {
		return o
	}

	return other
}

// OrElse returns the Option if it contains a value.
// Otherwise, it calls fn and returns its result.
func (o Option[T]) OrElse(fn func() Option[T]) Option[T] {
	if o.isSome {
		return o
	}

	return fn()
}

// And returns None if the Option is None, otherwise returns other.
// It is the eager counterpart of Then (which calls a function instead).
func (o Option[T]) And[U any](other Option[U]) Option[U] {
	if o.isSome {
		return other
	}

	return None[U]()
}

// Xor returns Some if exactly one of the two Options is Some, otherwise None.
func (o Option[T]) Xor(other Option[T]) Option[T] {
	if o.isSome == other.isSome {
		return None[T]()
	}

	if o.isSome {
		return o
	}

	return other
}

// IsSomeAnd returns true if the Option is Some
// and the predicate returns true for the contained value.
func (o Option[T]) IsSomeAnd(pred func(T) bool) bool {
	return o.isSome && pred(o.v)
}

// Insert inserts the given value into the Option,
// replacing any existing value, and returns a pointer
// to the inserted value.
func (o *Option[T]) Insert(value T) *T {
	o.v = value
	o.isSome = true

	return &o.v
}

// GetOrInsert inserts the given value if the Option is None,
// and returns a pointer to the contained value.
// If the Option already contains a value, it is left unchanged.
func (o *Option[T]) GetOrInsert(value T) *T {
	if !o.isSome {
		o.v = value
		o.isSome = true
	}
	return &o.v
}

// GetOrInsertWith inserts a value computed by fn if the Option is None,
// and returns a pointer to the contained value.
// The function fn is evaluated lazily.
func (o *Option[T]) GetOrInsertWith(fn func() T) *T {
	if !o.isSome {
		o.v = fn()
		o.isSome = true
	}

	return &o.v
}

// Take takes the value out of the Option, leaving None in its place.
// It returns Some(value) if the Option was Some,
// otherwise returns None.
func (o *Option[T]) Take() Option[T] {
	if !o.isSome {
		return None[T]()
	}

	val := o.v
	var zero T

	o.v = zero
	o.isSome = false

	return Some(val)
}

// Replace replaces the contained value with the given value,
// returning the old value as an Option.
// If the Option was None, it inserts the value and returns None.
func (o *Option[T]) Replace(value T) Option[T] {
	old := o.Take()
	o.v = value
	o.isSome = true

	return old
}

// Ptr returns a pointer to the contained value if Some.
// Otherwise, it returns nil.
func (o Option[T]) Ptr() *T {
	if o.isSome {
		return &o.v
	}

	return nil
}

// Option returns the contained value and a boolean reporting whether the Option is Some,
// conforming to the standard Go comma-ok pattern.
// If the Option is None, it returns the zero value for T and false.
func (o Option[T]) Option() (T, bool) {
	if o.IsSome() {
		return o.Some(), true
	}

	var zero T
	return zero, false
}

// IsNoneOr returns true if the Option is None or the predicate returns true
// for the contained value. It is the complement of IsSomeAnd.
func (o Option[T]) IsNoneOr(pred func(T) bool) bool {
	return !o.isSome || pred(o.v)
}

// String returns a string representation of the Option.
// If the Option contains a value, it returns a string in the format "Some(value)".
// Otherwise, it returns "None".
func (o Option[T]) String() string {
	if o.isSome {
		return fmt.Sprintf("Some(%v)", o.v)
	}

	return "None"
}
