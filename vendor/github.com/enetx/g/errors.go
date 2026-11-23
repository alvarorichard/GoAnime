package g

import "errors"

var (
	// ErrInvalidBinaryLength is returned by binary decoding when the input length is not a multiple of 8.
	ErrInvalidBinaryLength = errors.New("binary string length must be multiple of 8")
	// ErrInvalidBinaryDigit is returned by binary decoding when the input contains characters other than '0' and '1'.
	ErrInvalidBinaryDigit = errors.New("binary string must contain only '0' and '1'")

	// ErrParseInt is returned when a String cannot be parsed as an integer.
	ErrParseInt = errors.New("invalid integer")
	// ErrParseBigInt is returned when a String cannot be parsed as a big integer.
	ErrParseBigInt = errors.New("invalid big integer")
	// ErrParseFloat is returned when a String cannot be parsed as a float.
	ErrParseFloat = errors.New("invalid float")
	// ErrParseBool is returned when a String cannot be parsed as a bool.
	ErrParseBool = errors.New("invalid bool")
	// ErrParseUint is returned when a String cannot be parsed as an unsigned integer.
	ErrParseUint = errors.New("invalid unsigned integer")
	// ErrParseComplex is returned when a String cannot be parsed as a complex number.
	ErrParseComplex = errors.New("invalid complex number")
)

type wrappedError struct {
	msg  string
	errs []error
}

func (e *wrappedError) Error() string   { return e.msg }
func (e *wrappedError) Unwrap() []error { return e.errs }
