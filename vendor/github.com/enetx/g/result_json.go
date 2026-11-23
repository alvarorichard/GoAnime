package g

import (
	"encoding/json/jsontext"
	json "encoding/json/v2"
	"errors"
)

// errResultJSONShape is returned when a Result document is not a JSON object
// with exactly one of the keys "ok" or "err".
var errResultJSONShape = errors.New("g.Result: expected a JSON object with exactly one of the keys \"ok\" or \"err\"")

// MarshalJSON implements the json.Marshaler interface (encoding/json v1) for Result[T].
// The encoding is externally tagged:
// Ok(value) is marshaled as {"ok": <json of value>} and Err(err) is marshaled
// as {"err": "<err.Error()>"}.
//
// If marshaling the contained value fails, that error is returned.
//
// BREAKING: the implementation is backed by encoding/json/v2, which changes some
// edge-case semantics compared to the previous encoding/json implementation:
//   - Ok of a nil slice marshals as {"ok":[]} and Ok of a nil map as {"ok":{}},
//     rather than {"ok":null}.
//   - Strings containing invalid UTF-8 (in the Ok value or the error message)
//     are rejected with an error instead of being silently replaced with U+FFFD.
//
// NOTE: only the error message is serialized. The concrete error type and any
// wrapped errors are lost — after a round trip, errors.Is/errors.As chains no
// longer match.
func (r Result[T]) MarshalJSON() ([]byte, error) {
	if r.IsErr() {
		msg, err := json.Marshal(r.err.Error())
		if err != nil {
			return nil, err
		}

		return append(append([]byte(`{"err":`), msg...), '}'), nil
	}

	v, err := json.Marshal(r.v)
	if err != nil {
		return nil, err
	}

	return append(append([]byte(`{"ok":`), v...), '}'), nil
}

// UnmarshalJSON implements the json.Unmarshaler interface (encoding/json v1) for Result[T].
// It expects the externally tagged encoding produced by MarshalJSON: a JSON
// object with exactly one of the keys "ok" or "err".
//
// BREAKING: the implementation is backed by encoding/json/v2, so duplicate keys
// are rejected — a document that repeats the same key ({"ok":1,"ok":2}) is now
// an unmarshal error instead of the previous encoding/json last-wins semantics.
//
// {"err": "msg"} is unmarshaled as Err(errors.New("msg")); the value must be
// a JSON string. {"ok": <v>} is unmarshaled as Ok with v decoded into T —
// {"ok": null} decodes null into T following the encoding/json/v2 rules
// (zero/nil). Anything else (both keys, neither key, extra keys, duplicate
// keys, a non-object, or JSON null) is an unmarshal error.
//
// NOTE: only the error message survives a round trip. The original error type
// is not restored — the decoded error is a plain errors.New value, so
// errors.Is/errors.As chains against the original error no longer match.
func (r *Result[T]) UnmarshalJSON(data []byte) error {
	var raw map[string]jsontext.Value
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	// JSON null decodes into a nil map without error; reject it explicitly.
	if raw == nil || len(raw) != 1 {
		return errResultJSONShape
	}

	if msg, ok := raw["err"]; ok {
		if string(msg) == "null" {
			return errors.New("g.Result: \"err\" value must be a JSON string, got null")
		}

		var s string
		if err := json.Unmarshal(msg, &s); err != nil {
			return err
		}

		*r = Err[T](errors.New(s))
		return nil
	}

	value, ok := raw["ok"]
	if !ok {
		return errResultJSONShape
	}

	var v T
	if err := json.Unmarshal(value, &v); err != nil {
		return err
	}

	*r = Ok(v)
	return nil
}

// MarshalJSONTo implements the json.MarshalerTo interface (encoding/json/v2) for Result[T].
// encoding/json/v2 prefers this method over MarshalJSON.
// The encoding matches MarshalJSON: Ok(value) is encoded as {"ok": <json of value>}
// and Err(err) is encoded as {"err": "<err.Error()>"}.
func (r Result[T]) MarshalJSONTo(enc *jsontext.Encoder) error {
	if err := enc.WriteToken(jsontext.BeginObject); err != nil {
		return err
	}

	if r.IsErr() {
		if err := enc.WriteToken(jsontext.String("err")); err != nil {
			return err
		}

		if err := enc.WriteToken(jsontext.String(r.err.Error())); err != nil {
			return err
		}
	} else {
		if err := enc.WriteToken(jsontext.String("ok")); err != nil {
			return err
		}

		if err := json.MarshalEncode(enc, r.v); err != nil {
			return err
		}
	}

	return enc.WriteToken(jsontext.EndObject)
}

// UnmarshalJSONFrom implements the json.UnmarshalerFrom interface (encoding/json/v2) for Result[T].
// encoding/json/v2 prefers this method over UnmarshalJSON.
// It expects a JSON object with exactly one member whose key is "ok" or "err",
// as produced by MarshalJSONTo.
//
// Duplicate keys are rejected: the strict single-member contract refuses any
// second object member (a duplicate key included) before the v2 decoder's own
// duplicate-name check even fires, and the encoding/json/v2 decoder itself
// forbids duplicate object member names by default. Presence of both distinct
// keys, extra keys, neither key, a non-object, or JSON null is likewise an
// unmarshal error.
func (r *Result[T]) UnmarshalJSONFrom(dec *jsontext.Decoder) error {
	if dec.PeekKind() != '{' {
		// Surface the underlying syntax error for malformed input; otherwise
		// report the shape violation (null, array, string, number, bool).
		if _, err := dec.ReadToken(); err != nil {
			return err
		}

		return errResultJSONShape
	}

	if _, err := dec.ReadToken(); err != nil { // consume '{'
		return err
	}

	if dec.PeekKind() == '}' {
		return errResultJSONShape // empty object
	}

	name, err := dec.ReadToken()
	if err != nil {
		return err
	}

	var res Result[T]

	switch name.String() {
	case "err":
		if dec.PeekKind() != '"' {
			return errors.New("g.Result: \"err\" value must be a JSON string")
		}

		var s string
		if err := json.UnmarshalDecode(dec, &s); err != nil {
			return err
		}

		res = Err[T](errors.New(s))
	case "ok":
		var v T
		if err := json.UnmarshalDecode(dec, &v); err != nil {
			return err
		}

		res = Ok(v)
	default:
		return errResultJSONShape
	}

	if dec.PeekKind() != '}' {
		return errResultJSONShape // a second member: both keys, extra or duplicate keys
	}

	if _, err := dec.ReadToken(); err != nil { // consume '}'
		return err
	}

	*r = res
	return nil
}
