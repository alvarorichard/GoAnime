package startflix

import (
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// jsonAny is a stream envelope field whose shape depends on whether the
// upstream runs its URL protection: `stream_urls` is either a plain array of
// URLs or one encrypted string. Decoding eagerly into both shapes breaks one
// or the other, so the raw bytes are kept and decoded on demand.
type jsonAny struct {
	raw []byte
}

// UnmarshalJSON stores the raw bytes of this field.
func (a *jsonAny) UnmarshalJSON(data []byte) error {
	a.raw = append(a.raw[:0], data...)
	return nil
}

// asArray decodes an unquoted plain string-array value, or nil.
func (a *jsonAny) asArray() []string {
	if len(a.raw) == 0 || a.raw[0] != '[' {
		return nil
	}
	var arr []string
	if jsonx.Unmarshal(a.raw, &arr) != nil {
		return nil
	}
	return arr
}

// asString decodes a JSON string value, or "" when this field is not one.
func (a *jsonAny) asString() string {
	if len(a.raw) == 0 || a.raw[0] != '"' {
		return ""
	}
	var s string
	if jsonx.Unmarshal(a.raw, &s) != nil {
		return ""
	}
	return s
}
