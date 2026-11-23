package g

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"math"
	"time"
)

// Scan implements the database/sql.Scanner interface for Option[T].
//
// Behavior:
//   - If src is nil, the Option is set to None (SQL NULL).
//   - If T implements sql.Scanner, its Scan method is used.
//   - If src can be directly assigned to T, it is assigned as-is.
//   - Otherwise, common database type conversions are attempted (e.g., int64 → int, []byte → string).
//
// Supported conversions (common SQL types):
//   - INTEGER    → int, int8, int16, int32, int64, uint*
//   - REAL       → float32, float64
//   - TEXT       → string, []byte
//   - BLOB       → []byte
//   - BOOLEAN    → bool
//   - TIMESTAMP  → time.Time
//
// Driver-owned []byte buffers (BLOB/TEXT) are copied before being stored, so the
// scanned Option keeps a value that is safe to retain across subsequent rows
// (per the database/sql Scanner contract).
//
// Returns an error if the value cannot be converted to T.
func (o *Option[T]) Scan(src any) error {
	if src == nil {
		*o = None[T]()
		return nil
	}

	var v T

	if scanner, ok := any(&v).(sql.Scanner); ok {
		if err := scanner.Scan(src); err != nil {
			return err
		}
		*o = Some(v)
		return nil
	}

	if val, ok := src.(T); ok {
		// Copy driver-owned []byte buffers before storing: database/sql may
		// overwrite the backing array on the next row (database/sql contract).
		if b, isBytes := any(val).([]byte); isBytes {
			val = any(append([]byte(nil), b...)).(T)
		}
		*o = Some(val)
		return nil
	}

	if converted, ok := convertToT[T](src); ok {
		*o = Some(converted)
		return nil
	}

	return fmt.Errorf("Option.Scan: cannot scan %T into %T", src, v)
}

// Value implements the database/sql/driver.Valuer interface for Option[T].
//
// Behavior:
//   - If the Option is None, returns nil (SQL NULL).
//   - If T implements driver.Valuer, its Value method is used.
//   - If the underlying value is already a valid driver.Value type (int64, float64, bool, []byte, string, time.Time), it is returned directly.
//   - Otherwise, safe conversions are applied (int → int64, uint → int64, float32 → float64).
//
// Returns an error if the value cannot be converted to a driver.Value.
func (o Option[T]) Value() (driver.Value, error) {
	if o.IsNone() {
		return nil, nil
	}

	if valuer, ok := any(o.v).(driver.Valuer); ok {
		return valuer.Value()
	}

	switch val := any(o.v).(type) {
	case int64, float64, bool, []byte, string, time.Time:
		return val, nil
	default:
		if converted, ok := convertToDriverValue(val); ok {
			return converted, nil
		}
	}

	return nil, fmt.Errorf("Option.Value: unsupported type %T", o.v)
}

// convertToT attempts to safely convert a source value from database/sql
// into type T, supporting common database types without using reflection.
//
// Only standard Go primitive types are supported:
//   - int64 → int, int8, int16, int32, uint*, uint64 (if fits)
//   - float64 → float32, float64
//   - string / []byte → string
//   - []byte / string → []byte
//   - bool → bool
//   - time.Time → time.Time
//
// Returns the converted value and true on success, otherwise the zero value of T and false.
func convertToT[T any](src any) (T, bool) {
	var zero T

	switch any(zero).(type) {
	case int:
		if i64, ok := src.(int64); ok && fitsInt(i64) {
			return any(int(i64)).(T), true
		}
	case int8:
		if i64, ok := src.(int64); ok && fitsInt8(i64) {
			return any(int8(i64)).(T), true
		}
	case int16:
		if i64, ok := src.(int64); ok && fitsInt16(i64) {
			return any(int16(i64)).(T), true
		}
	case int32:
		if i64, ok := src.(int64); ok && fitsInt32(i64) {
			return any(int32(i64)).(T), true
		}
	case int64:
		if i64, ok := src.(int64); ok {
			return any(i64).(T), true
		}
	case uint:
		if i64, ok := src.(int64); ok && i64 >= 0 {
			return any(uint(i64)).(T), true
		}
	case uint8:
		if i64, ok := src.(int64); ok && i64 >= 0 && i64 <= math.MaxUint8 {
			return any(uint8(i64)).(T), true
		}
	case uint16:
		if i64, ok := src.(int64); ok && i64 >= 0 && i64 <= math.MaxUint16 {
			return any(uint16(i64)).(T), true
		}
	case uint32:
		if i64, ok := src.(int64); ok && i64 >= 0 && i64 <= math.MaxUint32 {
			return any(uint32(i64)).(T), true
		}
	case uint64:
		if i64, ok := src.(int64); ok && i64 >= 0 {
			return any(uint64(i64)).(T), true
		}
	case float32:
		if f64, ok := src.(float64); ok {
			return any(float32(f64)).(T), true
		}
	case float64:
		if f64, ok := src.(float64); ok {
			return any(f64).(T), true
		}
	case string:
		switch v := src.(type) {
		case string:
			return any(v).(T), true
		case []byte:
			return any(string(v)).(T), true
		}
	case []byte:
		switch v := src.(type) {
		case []byte:
			// Copy the driver-owned buffer: database/sql may reuse it on the next row.
			return any(append([]byte(nil), v...)).(T), true
		case string:
			return any([]byte(v)).(T), true
		}
	case bool:
		if b, ok := src.(bool); ok {
			return any(b).(T), true
		}
	case time.Time:
		if t, ok := src.(time.Time); ok {
			return any(t).(T), true
		}
	}

	return zero, false
}

// convertToDriverValue safely converts primitive Go types to a value
// compatible with database/sql driver.Value.
//
// Supported conversions:
//   - int, int8, int16, int32 → int64
//   - uint8, uint16, uint32 → int64
//   - uint, uint64 → int64 (only if <= math.MaxInt64)
//   - float32 → float64
//
// Returns the converted value and true on success, otherwise nil and false.
func convertToDriverValue(val any) (driver.Value, bool) {
	switch v := val.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case uint:
		if uint64(v) <= math.MaxInt64 {
			return int64(v), true
		}
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		if v <= math.MaxInt64 {
			return int64(v), true
		}
	case float32:
		return float64(v), true
	}

	return nil, false
}

// fitsInt returns true if i fits in a Go int.
func fitsInt(i int64) bool { return i >= math.MinInt && i <= math.MaxInt }

// fitsInt8 returns true if i fits in an int8.
func fitsInt8(i int64) bool { return i >= math.MinInt8 && i <= math.MaxInt8 }

// fitsInt16 returns true if i fits in an int16.
func fitsInt16(i int64) bool { return i >= math.MinInt16 && i <= math.MaxInt16 }

// fitsInt32 returns true if i fits in an int32.
func fitsInt32(i int64) bool { return i >= math.MinInt32 && i <= math.MaxInt32 }
