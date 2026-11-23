package g

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Result is a generic struct for representing a result value along with an error.
type Result[T any] struct {
	v   T     // Value.
	err error // Associated error.
}

// Ok returns a new Result[T] containing the given value.
func Ok[T any](value T) Result[T] { return Result[T]{v: value} }

// Err returns a new Result[T] containing the given error.
func Err[T any](err error) Result[T] {
	if err == nil {
		err = errors.New("g.Err called with a nil error")
	}

	return Result[T]{err: err}
}

// ResultOf returns a new Result[T] based on the provided value and error.
// If err is not nil, it returns an Err Result.
// Otherwise, it returns an Ok Result.
func ResultOf[T any](value T, err error) Result[T] {
	if err != nil {
		return Err[T](err)
	}

	return Ok(value)
}

// Ok returns the value held in the Result.
//
// WARNING: If the Result contains an error, this method will return the zero value
// for type T. Always check IsOk() before calling this method, or use safer
// alternatives like Result(), UnwrapOr(), or UnwrapOrDefault(). (Unwrap() and
// Expect() are NOT safe alternatives — they panic on an Err value.)
func (r Result[T]) Ok() T { return r.v }

// Err returns the error held in the Result. If the result is Ok, it returns nil.
func (r Result[T]) Err() error { return r.err }

// IsOk returns true if the Result contains a value (no error).
func (r Result[T]) IsOk() bool { return r.err == nil }

// IsErr returns true if the Result contains an error.
func (r Result[T]) IsErr() bool { return r.err != nil }

// Result returns the value and error, conforming to the standard Go multi-value return pattern.
func (r Result[T]) Result() (T, error) {
	if r.IsOk() {
		return r.v, nil
	}

	var zero T
	return zero, r.err
}

// resultPanic prints caller information for msg to stderr and then panics with v.
// skip is the number of stack frames between the original caller and runtime.Caller,
// so that the reported file:line and function point at the user's call site rather
// than at this helper.
func resultPanic(skip int, msg string, v any) {
	if pc, file, line, ok := runtime.Caller(skip); ok {
		out := fmt.Sprintf("[%s:%d] [%s] %s", filepath.Base(file), line, runtime.FuncForPC(pc).Name(), msg)
		fmt.Fprintln(os.Stderr, out)
	}

	panic(v)
}

// Unwrap returns the value held in the Result. If the Result is Err, it panics.
func (r Result[T]) Unwrap() T {
	if r.IsOk() {
		return r.v
	}

	resultPanic(2, fmt.Sprintf("called Result.Unwrap() on an Err value: %v", r.err), r.err)
	panic("unreachable")
}

// UnwrapOr returns the value held in the Result. If the Result is Err, it returns the provided default value.
func (r Result[T]) UnwrapOr(value T) T {
	if r.IsOk() {
		return r.v
	}

	return value
}

// UnwrapOrDefault returns the contained value if Ok, otherwise returns the zero value for T.
func (r Result[T]) UnwrapOrDefault() T {
	if r.IsOk() {
		return r.v
	}

	var zero T
	return zero
}

// Expect returns the value held in the Result. If the Result is Err, it panics with the provided message.
func (r Result[T]) Expect(msg string) T {
	if r.IsOk() {
		return r.v
	}

	out := fmt.Sprintf("Expect() failed: %s: %v", msg, r.err)
	resultPanic(2, out, out)
	panic("unreachable")
}

// Then applies a function to the contained value (if Ok) and returns the resulting Result.
// If the Result is Err, fn is not called and the error is propagated.
// The result type may differ from the input type.
func (r Result[T]) Then[U any](fn func(T) Result[U]) Result[U] {
	if r.IsOk() {
		return fn(r.v)
	}

	return Err[U](r.err)
}

// Map applies a function to the contained value (if Ok) and returns a new Result
// holding the transformed value. If the Result is Err, fn is not called and the
// error is propagated. Unlike Then, fn returns a plain U rather than a Result[U].
func (r Result[T]) Map[U any](fn func(T) U) Result[U] {
	if r.IsOk() {
		return Ok(fn(r.v))
	}

	return Err[U](r.err)
}

// MapOr applies fn to the contained value if Ok and returns the result;
// otherwise returns the provided default value.
func (r Result[T]) MapOr[U any](def U, fn func(T) U) U {
	if r.IsOk() {
		return fn(r.v)
	}

	return def
}

// MapOrElse applies fn to the contained value if Ok and returns the result;
// otherwise computes the default from the error via defFn.
func (r Result[T]) MapOrElse[U any](defFn func(error) U, fn func(T) U) U {
	if r.IsOk() {
		return fn(r.v)
	}

	return defFn(r.err)
}

// Inspect calls fn with the contained value if the Result is Ok, then returns
// the Result unchanged. If the Result is Err, fn is not called. It is intended
// for side effects (logging, debugging) within a chain and never mutates the Result.
func (r Result[T]) Inspect(fn func(T)) Result[T] {
	if r.IsOk() {
		fn(r.v)
	}

	return r
}

// ThenOf applies a function to the contained value (if Ok) and returns a new Result
// based on the returned (U, error) tuple. If the Result is Err, fn is not called
// and the error is propagated.
func (r Result[T]) ThenOf[U any](fn func(T) (U, error)) Result[U] {
	if r.IsOk() {
		return ResultOf(fn(r.v))
	}

	return Err[U](r.err)
}

// MapErr transforms the error in an Err Result by applying a function to it.
// It is useful for custom error handling, like replacing one error with another.
// If the Result is Ok, it does nothing.
func (r Result[T]) MapErr(fn func(error) error) Result[T] {
	if r.IsErr() {
		return Err[T](fn(r.err))
	}

	return r
}

// Option converts a Result into an Option.
// If the Result is Ok, it returns Some(value).
// If the Result is Err, it returns None.
func (r Result[T]) Option() Option[T] {
	if r.IsOk() {
		return Some(r.v)
	}

	return None[T]()
}

// String returns a string representation of the Result.
func (r Result[T]) String() string {
	if r.IsOk() {
		return fmt.Sprintf("Ok(%v)", r.v)
	}

	return fmt.Sprintf("Err(%v)", r.err)
}

// ErrIs reports whether the error in Result matches target (using errors.Is).
// Returns false if Result is Ok.
func (r Result[T]) ErrIs(target error) bool { return errors.Is(r.err, target) }

// ErrAs finds the first error in Result's error chain that matches target,
// and if so, sets target to that error value and returns true (using errors.As).
// Returns false if Result is Ok.
func (r Result[T]) ErrAs(target any) bool { return errors.As(r.err, target) }

// ErrSource returns the underlying error wrapped by the Result's error, if any.
// Returns None if Result is Ok or if the error doesn't wrap another error.
func (r Result[T]) ErrSource() Option[error] {
	if source := errors.Unwrap(r.err); source != nil {
		return Some(source)
	}

	return None[error]()
}

// Wrap wraps the error in Result with additional context error.
// Both errors are preserved in the chain and accessible via ErrIs.
// If Result is Ok, returns unchanged.
func (r Result[T]) Wrap(err error) Result[T] {
	if r.IsErr() {
		return Err[T](fmt.Errorf("%w: %w", err, r.err))
	}

	return r
}

// Or returns the Result if it is Ok, otherwise returns the provided alternative Result.
func (r Result[T]) Or(other Result[T]) Result[T] {
	if r.IsOk() {
		return r
	}

	return other
}

// OrElse returns the Result if it is Ok, otherwise calls fn with the error and returns its result.
func (r Result[T]) OrElse(fn func(error) Result[T]) Result[T] {
	if r.IsOk() {
		return r
	}

	return fn(r.err)
}

// And returns the receiver's error if it is Err, otherwise returns other.
// It is the eager counterpart of Then (which calls a function instead).
func (r Result[T]) And[U any](other Result[U]) Result[U] {
	if r.IsErr() {
		return Err[U](r.err)
	}

	return other
}

// ErrOption returns the contained error as an Option: Some(err) if the Result is
// Err, or None if it is Ok. It is the error-side counterpart of Option, which
// returns the Ok value as an Option.
func (r Result[T]) ErrOption() Option[error] {
	if r.IsErr() {
		return Some(r.err)
	}

	return None[error]()
}

// IsOkAnd returns true if the Result is Ok and the predicate returns true for the contained value.
func (r Result[T]) IsOkAnd(pred func(T) bool) bool {
	return r.IsOk() && pred(r.v)
}

// IsErrAnd returns true if the Result is Err and the predicate returns true for the contained error.
func (r Result[T]) IsErrAnd(pred func(error) bool) bool {
	return r.IsErr() && pred(r.err)
}

// InspectErr calls fn with the contained error if the Result is Err, then returns
// the Result unchanged. It is the error-side counterpart of Inspect.
func (r Result[T]) InspectErr(fn func(error)) Result[T] {
	if r.IsErr() {
		fn(r.err)
	}

	return r
}

// UnwrapErr returns the contained error. If the Result is Ok, it panics.
func (r Result[T]) UnwrapErr() error {
	if r.IsErr() {
		return r.err
	}

	out := fmt.Sprintf("called Result.UnwrapErr() on an Ok value: %v", r.v)
	resultPanic(2, out, out)
	panic("unreachable")
}

// OkOr converts an Option into a Result: Ok(value) if Some, otherwise
// Err(err).
//
// It is a free function rather than an Option method: as a method every
// Option[T] instantiation would also instantiate Result[T] (and everything
// Result's methods mention), even in packages that never convert between the
// two.
func OkOr[T any](o Option[T], err error) Result[T] {
	if o.isSome {
		return Ok(o.v)
	}

	return Err[T](err)
}

// OkOrElse converts an Option into a Result: Ok(value) if Some, otherwise
// Err(fn()). See [OkOr] for why this is not a method.
func OkOrElse[T any](o Option[T], fn func() error) Result[T] {
	if o.isSome {
		return Ok(o.v)
	}

	return Err[T](fn())
}
