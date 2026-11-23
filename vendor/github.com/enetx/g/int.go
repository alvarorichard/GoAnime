package g

import (
	"database/sql/driver"
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"

	"github.com/enetx/g/cmp"
	"github.com/enetx/g/constraints"
)

// Int is a wrapper around the int type.
type Int int

// NewInt creates a new Int with the provided int value.
func NewInt[T constraints.Integer | rune | byte](i T) Int { return Int(i) }

// Transform applies a transformation function to the Int and returns the result.
func (i Int) Transform[U any](fn func(Int) U) U { return fn(i) }

// Min returns the minimum of Ints.
func (i Int) Min(b ...Int) Int { return cmp.Min(append(b, i)...) }

// Max returns the maximum of Ints.
func (i Int) Max(b ...Int) Int { return cmp.Max(append(b, i)...) }

// Abs returns the absolute value of the Int.
// Like Go's native arithmetic it wraps on overflow: Abs of math.MinInt is math.MinInt.
// Use CheckedAbs for a guarded variant.
func (i Int) Abs() Int {
	if i < 0 {
		return -i
	}

	return i
}

// Add adds two Ints and returns the result.
// Like Go's native arithmetic it wraps on overflow (two's complement).
// Use CheckedAdd, SaturatingAdd or OverflowingAdd for guarded variants.
func (i Int) Add(b Int) Int { return i + b }

// Neg returns the Int with its sign inverted.
// Like Go's native arithmetic it wraps on overflow: Neg of math.MinInt is math.MinInt.
// Use CheckedNeg for a guarded variant.
func (i Int) Neg() Int { return -i }

// Signum returns the sign of the Int:
// -1 if the Int is negative, 0 if it is zero, and 1 if it is positive.
func (i Int) Signum() Int {
	switch {
	case i < 0:
		return -1
	case i > 0:
		return 1
	default:
		return 0
	}
}

// BigInt returns the Int as a *big.Int.
func (i Int) BigInt() *big.Int { return big.NewInt(i.Int64()) }

// Div divides two Ints and returns the result.
//
// Div panics with a runtime "integer divide by zero" error if b is 0.
// Dividing by zero is treated as a programmer error; guard against a zero
// divisor at the call site. This differs from Float.Div, which follows IEEE
// 754 and yields ±Inf or NaN instead of panicking.
func (i Int) Div(b Int) Int { return i / b }

// Eq checks if two Ints are equal.
func (i Int) Eq(b Int) bool { return i == b }

// Gt checks if the Int is greater than the specified Int.
func (i Int) Gt(b Int) bool { return i > b }

// Gte checks if the Int is greater than or equal to the specified Int.
func (i Int) Gte(b Int) bool { return i >= b }

// Float returns the Int as an Float.
func (i Int) Float() Float { return Float(i) }

// String returns the Int as an String.
func (i Int) String() String { return String(strconv.FormatInt(int64(i), 10)) }

// Std returns the Int as an int.
func (i Int) Std() int { return int(i) }

// Cmp compares two Ints and returns an cmp.Ordering.
func (i Int) Cmp(b Int) cmp.Ordering { return cmp.Cmp(i, b) }

// Int16 returns the Int as an int16.
func (i Int) Int16() int16 { return int16(i) }

// Int32 returns the Int as an int32.
func (i Int) Int32() int32 { return int32(i) }

// Int64 returns the Int as an int64.
func (i Int) Int64() int64 { return int64(i) }

// Int8 returns the Int as an int8.
func (i Int) Int8() int8 { return int8(i) }

// IsZero checks if the Int is 0.
func (i Int) IsZero() bool { return i == 0 }

// IsNegative checks if the Int is negative.
func (i Int) IsNegative() bool { return i < 0 }

// IsPositive reports whether the Int is strictly greater than zero.
// Zero is neither positive nor negative: both
// Int(0).IsPositive() and Int(0).IsNegative() return false. For a
// non-negative check use !i.IsNegative().
func (i Int) IsPositive() bool { return i > 0 }

// Lt checks if the Int is less than the specified Int.
func (i Int) Lt(b Int) bool { return i < b }

// Lte checks if the Int is less than or equal to the specified Int.
func (i Int) Lte(b Int) bool { return i <= b }

// Mul multiplies two Ints and returns the result.
// Like Go's native arithmetic it wraps on overflow (two's complement).
// Use CheckedMul, SaturatingMul or OverflowingMul for guarded variants.
func (i Int) Mul(b Int) Int { return i * b }

// Ne checks if two Ints are not equal.
func (i Int) Ne(b Int) bool { return i != b }

// Rem returns the remainder of the division between the receiver and the input value.
//
// Rem panics with a runtime "integer divide by zero" error if b is 0.
// A zero divisor is treated as a programmer error; guard against it at the
// call site.
func (i Int) Rem(b Int) Int { return i % b }

// Sub subtracts two Ints and returns the result.
// Like Go's native arithmetic it wraps on overflow (two's complement).
// Use CheckedSub, SaturatingSub or OverflowingSub for guarded variants.
func (i Int) Sub(b Int) Int { return i - b }

// Binary returns the Int as a binary string, zero-padded to a minimum width of
// 8 characters (the sign counts toward the width for negative values).
func (i Int) Binary() String {
	var storage [65]byte
	digits := strconv.AppendInt(storage[:0], int64(i), 2)
	if len(digits) >= 8 {
		return String(digits)
	}

	var padded [8]byte
	start := 8 - len(digits)
	if digits[0] == '-' {
		padded[0] = '-'
		start++
		for j := 1; j < start; j++ {
			padded[j] = '0'
		}
		copy(padded[start:], digits[1:])
	} else {
		for j := 0; j < start; j++ {
			padded[j] = '0'
		}
		copy(padded[start:], digits)
	}

	return String(padded[:])
}

// Hex returns the Int as a hexadecimal string.
func (i Int) Hex() String { return String(strconv.FormatInt(int64(i), 16)) }

// Octal returns the Int as an octal string.
func (i Int) Octal() String { return String(strconv.FormatInt(int64(i), 8)) }

// Uint returns the Int as a uint.
func (i Int) Uint() uint { return uint(i) }

// Uint16 returns the Int as a uint16.
func (i Int) Uint16() uint16 { return uint16(i) }

// Uint32 returns the Int as a uint32.
func (i Int) Uint32() uint32 { return uint32(i) }

// Uint64 returns the Int as a uint64.
func (i Int) Uint64() uint64 { return uint64(i) }

// Uint8 returns the Int as a uint8.
func (i Int) Uint8() uint8 { return uint8(i) }

// bytesFromInt converts Int to Bytes using the given byte order.
// For BE: removes leading zeros while preserving the sign bit.
// For LE: removes trailing zeros while preserving the sign bit.
func bytesFromInt(i Int, order binary.ByteOrder) Bytes {
	var buf [8]byte
	order.PutUint64(buf[:], uint64(i))

	switch order {
	case binary.BigEndian:
		start := 0
		for start < 7 && buf[start] == 0 {
			start++
		}

		if i >= 0 && buf[start]&0x80 != 0 {
			start--
		}

		if i < 0 && start > 0 && buf[start]&0x80 == 0 {
			start--
		}

		return Bytes(buf[start:])
	case binary.LittleEndian:
		end := 8
		for end > 1 && buf[end-1] == 0 {
			end--
		}

		if i >= 0 && buf[end-1]&0x80 != 0 {
			end++
		}

		if i < 0 && end < 8 && buf[end-1]&0x80 == 0 {
			end++
		}

		return Bytes(buf[:end])
	}

	return Bytes(buf[:])
}

// BytesBE converts the Int to Bytes in BigEndian order.
// Leading zero bytes are removed while preserving the sign bit for negative numbers.
func (i Int) BytesBE() Bytes {
	return bytesFromInt(i, binary.BigEndian)
}

// BytesLE converts the Int to Bytes in LittleEndian order.
// Trailing zero bytes are removed while preserving the sign bit for negative numbers
func (i Int) BytesLE() Bytes {
	return bytesFromInt(i, binary.LittleEndian)
}

// Print writes the value of the Int to the standard output (console)
// and returns the Int unchanged.
func (i Int) Print() Int { fmt.Print(i); return i }

// Println writes the value of the Int to the standard output (console) with a newline
// and returns the Int unchanged.
func (i Int) Println() Int { fmt.Println(i); return i }

// Scan implements the database/sql.Scanner interface for g.Int.
//
// Behavior:
//   - If src is nil, the value is set to 0 (SQL NULL).
//   - If src is an int64 (common SQL INTEGER type), it is assigned.
//   - Otherwise, an error is returned.
//
// Supported SQL types (common):
//   - INTEGER → int64
//
// Notes:
//   - This allows g.Int to be used directly with database/sql and compatible drivers.
func (i *Int) Scan(src any) error {
	if src == nil {
		*i = 0
		return nil
	}

	if i64, ok := src.(int64); ok {
		*i = Int(i64)
		return nil
	}

	return fmt.Errorf("g.Int.Scan: cannot scan %T into g.Int", src)
}

// Value implements the database/sql/driver.Valuer interface for g.Int.
//
// Behavior:
//   - Returns the underlying int64 value, ready for database insertion.
//   - Always returns a value compatible with SQL INTEGER type.
func (i Int) Value() (driver.Value, error) { return int64(i), nil }
