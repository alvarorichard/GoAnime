package g

import (
	"encoding/json/jsontext"
	json "encoding/json/v2"
)

// jsonNull is the JSON literal returned when marshaling a None Option.
// It is a package-level value to avoid allocating a new []byte on every None marshal.
//
// NOTE: callers (encoding/json) must not mutate the returned slice; the standard
// library treats Marshaler output as read-only, so sharing this backing array is safe.
var jsonNull = []byte("null")

// MarshalJSON implements the json.Marshaler interface (encoding/json v1) for Option[T].
// Some(value) is marshaled as the JSON representation of value.
// None is marshaled as null.
//
// BREAKING: the implementation is backed by encoding/json/v2, which changes some
// edge-case semantics compared to the previous encoding/json implementation:
//   - Some of a nil slice marshals as [] and Some of a nil map marshals as {},
//     rather than null.
//   - Strings containing invalid UTF-8 are rejected with an error instead of
//     being silently replaced with U+FFFD.
func (o Option[T]) MarshalJSON() ([]byte, error) {
	if o.IsNone() {
		return jsonNull, nil
	}

	return json.Marshal(o.v)
}

// UnmarshalJSON implements the json.Unmarshaler interface (encoding/json v1) for Option[T].
// JSON null is unmarshaled as None.
// Any other valid JSON value is unmarshaled as Some(value).
//
// BREAKING: the implementation is backed by encoding/json/v2, which is stricter
// than the previous encoding/json implementation: duplicate object keys inside
// the value are rejected, struct field names match case-sensitively, and strings
// containing invalid UTF-8 are rejected.
func (o *Option[T]) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*o = None[T]()
		return nil
	}

	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}

	*o = Some(v)
	return nil
}

// MarshalJSONTo implements the json.MarshalerTo interface (encoding/json/v2) for Option[T].
// encoding/json/v2 prefers this method over MarshalJSON.
// Some(value) is encoded as the JSON representation of value; None is encoded as null.
//
// Because None and JSON null share one representation, nested Options collapse:
// Some(None) marshals to null and unmarshals back as None. Wrap the inner
// Option in a struct (or use Result) when the distinction must survive a round trip.
func (o Option[T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	if o.IsNone() {
		return enc.WriteToken(jsontext.Null)
	}

	return json.MarshalEncode(enc, o.v)
}

// UnmarshalJSONFrom implements the json.UnmarshalerFrom interface (encoding/json/v2) for Option[T].
// encoding/json/v2 prefers this method over UnmarshalJSON.
// JSON null is decoded as None; any other value is decoded into T as Some(value).
func (o *Option[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if dec.PeekKind() == 'n' {
		if _, err := dec.ReadToken(); err != nil {
			return err
		}

		*o = None[T]()
		return nil
	}

	var v T
	if err := json.UnmarshalDecode(dec, &v); err != nil {
		return err
	}

	*o = Some(v)
	return nil
}
