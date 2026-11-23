package http

import (
	"strconv"
	"strings"

	"github.com/enetx/http/internal/headerorder"
	"golang.org/x/net/http/httpguts"
)

// HeaderOrderKey is a magic key for ResponseWriter.Header map keys
// that, if present, defines a header order that will be used to
// write the headers onto wire. The order of the list defined how the headers
// will be sorted. A defined key goes before an undefined key.
//
// This is the only way to specify some order, because maps don't
// have a a stable iteration order. If no order is given, headers will
// be sorted lexicographically.
//
// According to RFC-2616 it is good practice to send general-header fields
// first, followed by request-header or response-header fields and ending
// with entity-header fields.
//
// For outgoing requests, a Content-Length slot can also frame a replayable empty
// upload with Content-Length: 0 when Request.GetBody is non-nil. This does not
// change the defaults for requests without such a slot or unknown-length bodies.
// An explicitly nil Content-Length value slice still suppresses the zero field.
const HeaderOrderKey = "Header-Order:"

// PHeaderOrderKey is a magic key for setting http2 pseudo header order.
// If the header is nil it will use regular GoLang header order.
// Valid fields are :authority, :method, :path, :scheme
const PHeaderOrderKey = "PHeader-Order:"

// HeaderCaseKey selects exact HTTP/1 wire spellings for listed field names.
// Other names retain canonical spelling. Invalid names are ignored. This
// metadata is never sent and has no effect on HTTP/2 or HTTP/3 field names.
const HeaderCaseKey = headerorder.HeaderCaseKey

func isHeaderMetadata(key string) bool { return headerorder.IsMetadataKey(key) }

func http1HeaderNames(h Header) func(string) string {
	if len(h[HeaderCaseKey]) == 0 {
		return CanonicalHeaderKey
	}
	cases := make(map[string]string, len(h[HeaderCaseKey]))
	for _, name := range h[HeaderCaseKey] {
		if httpguts.ValidHeaderFieldName(name) {
			cases[strings.ToLower(name)] = name
		}
	}
	return func(name string) string {
		if exact, ok := cases[strings.ToLower(name)]; ok {
			return exact
		}
		return CanonicalHeaderKey(name)
	}
}

// Computed request framing and explicit ordered lengths use a private copy.
// The shared sorter leaves the caller's headers and exclusion map untouched.
func orderedHeaders(h Header, exclude map[string]bool, cl int64, isRequest bool) []headerorder.Field {
	order := headerorder.Indices(h[HeaderOrderKey])
	_, orderedLength := order["content-length"]
	orderedLength = orderedLength && cl >= 0
	omitLength := isRequest && omitZeroContentLength(h, cl)
	copyLengthFields := orderedLength
	if isRequest && !copyLengthFields {
		for name := range h {
			if strings.EqualFold(name, "content-length") {
				copyLengthFields = true
				break
			}
		}
	}
	if copyLengthFields {
		h = h.Clone()
		if isRequest {
			for name := range h {
				if strings.EqualFold(name, "content-length") {
					delete(h, name)
				}
			}
		}
		if orderedLength && !omitLength {
			h["content-length"] = []string{strconv.FormatInt(cl, 10)}
		}
	}
	return headerorder.Sorted(h, order, exclude)
}

// omitZeroContentLength honors an explicitly empty field only for a body whose
// sanitized length is exactly zero. Positive and unknown lengths retain framing.
func omitZeroContentLength(h Header, contentLength int64) bool {
	if contentLength != 0 {
		return false
	}
	for name, values := range h {
		if strings.EqualFold(name, "content-length") && len(values) == 0 {
			return true
		}
	}
	return false
}
