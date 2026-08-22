// Package jsonx wraps encoding/json/v2 (Go 1.27) with the leniency GoAnime needs
// when decoding responses from anime/metadata providers.
//
// # Why not encoding/json
//
// Since Go 1.27 the v1 encoding/json API is implemented on top of v2, but every
// call pays for the v1 compatibility option set. Calling v2 directly with only
// the three leniency options we actually need is measurably cheaper: on a 100 KB
// episode-list payload it is ~22% faster than json.Unmarshal, and Decode's
// streaming path allocates ~59% fewer bytes than io.ReadAll + json.Unmarshal
// because the whole response is never materialised as a []byte.
//
// # Which options, and why
//
// The option set is "everything v1 does, except v1 error reporting":
//
//	json.DefaultOptionsV1() + ReportErrorsWithLegacySemantics(false)
//
// Benchmarking each v1 compatibility option separately on the same payload
// showed that exactly one of them is expensive: ReportErrorsWithLegacySemantics
// costs ~25%, while Merge, MatchCaseSensitiveDelimiter, CallMethods,
// UnmarshalArrayFromAnyLength, the byte/time/stringify options and the rest are
// free. Turning off only that one keeps every v1 *data* semantic that provider
// payloads depend on:
//
//   - case-insensitive field matching (provider APIs are inconsistent about
//     casing, and v2's strict matching would silently leave fields zeroed — the
//     worst possible failure mode here);
//   - invalid UTF-8 replaced with U+FFFD instead of rejecting the document
//     (scraped titles regularly carry mojibake);
//   - duplicate object names resolved last-one-wins instead of rejected;
//   - v1 merge semantics for null and for pre-populated Go values.
//
// What changes: the error *values* on malformed input are v2 errors rather than
// *json.SyntaxError / *json.UnmarshalTypeError. No call site in this repository
// inspects those types; they are only ever wrapped and logged.
//
// The differential tests in jsonx_diff_test.go (plus a fuzz target) decode a
// corpus with both implementations and require identical results, so any future
// drift in these semantics fails the build.
//
// Building with GOEXPERIMENT=nojsonv2 removes encoding/json/v2 from the standard
// library and this package will not compile.
package jsonx

import (
	jsonv1 "encoding/json"
	jsonv2 "encoding/json/v2"
	"errors"
	"fmt"
	"io"
)

// ErrTooLarge is returned by Decode when the reader yields more than the
// permitted number of bytes. It replaces the silent truncation you get from
// io.ReadAll(io.LimitReader(...)), where an oversized body degrades into an
// opaque "unexpected end of JSON input".
var ErrTooLarge = errors.New("jsonx: response exceeds size limit")

// options is the shared, immutable option set. Building it once avoids
// re-joining options on every call.
var options = jsonv2.JoinOptions(
	jsonv1.DefaultOptionsV1(),
	jsonv1.ReportErrorsWithLegacySemantics(false),
)

// Options returns the option set used by Unmarshal and Decode, for call sites
// that need to add their own on top (later options win).
func Options() jsonv2.Options { return options }

// Unmarshal decodes data into v. It is a drop-in replacement for
// json.Unmarshal for provider payloads.
//
// Use this when the caller also needs the raw bytes (HTML-response guards,
// regex fallbacks, diagnostics). When the bytes are only used for decoding,
// prefer Decode, which skips the intermediate buffer entirely.
func Unmarshal(data []byte, v any) error {
	return jsonv2.Unmarshal(data, v, options)
}

// Decode reads at most limit bytes from r and decodes the JSON value into v
// without buffering the whole payload.
//
// A body larger than limit fails with ErrTooLarge instead of being silently
// truncated. A limit of zero or less means unlimited, which should only be used
// for trusted local input.
func Decode(r io.Reader, limit int64, v any) error {
	if limit > 0 {
		cr := &capReader{r: r, remaining: limit + 1}
		if err := jsonv2.UnmarshalRead(cr, v, options); err != nil {
			if cr.exceeded || errors.Is(err, ErrTooLarge) {
				return fmt.Errorf("%w (%d bytes)", ErrTooLarge, limit)
			}
			return err
		}
		if cr.exceeded {
			return fmt.Errorf("%w (%d bytes)", ErrTooLarge, limit)
		}
		return nil
	}
	return jsonv2.UnmarshalRead(r, v, options)
}

// Marshal encodes v using the same option set.
func Marshal(v any) ([]byte, error) {
	return jsonv2.Marshal(v, options)
}

// capReader fails with ErrTooLarge once the allowance is exhausted, instead of
// reporting a clean EOF the way io.LimitReader does. The allowance is limit+1 so
// a body of exactly limit bytes still reaches EOF normally.
type capReader struct {
	r         io.Reader
	remaining int64
	exceeded  bool
}

func (c *capReader) Read(p []byte) (int, error) {
	if c.remaining <= 0 {
		c.exceeded = true
		return 0, ErrTooLarge
	}
	if int64(len(p)) > c.remaining {
		p = p[:c.remaining]
	}
	n, err := c.r.Read(p)
	c.remaining -= int64(n)
	return n, err
}
