package httpcommon

import (
	"maps"
	"strconv"
	"strings"

	"github.com/enetx/http/internal/headerorder"
)

const (
	HeaderOrderKey  = "Header-Order:"
	PHeaderOrderKey = "PHeader-Order:"
	HeaderCaseKey   = headerorder.HeaderCaseKey
)

func isHeaderMetadata(key string) bool { return headerorder.IsMetadataKey(key) }

func orderedRequestHeaders(req Request) []headerorder.Field {
	order := headerorder.Indices(req.Header[HeaderOrderKey])
	headers := maps.Clone(req.Header)
	for k := range headers {
		if strings.EqualFold(k, "content-length") {
			delete(headers, k)
		}
	}
	if _, ok := order["content-length"]; ok && shouldSendRequestContentLength(req) {
		headers["content-length"] = []string{strconv.FormatInt(req.ActualContentLength, 10)}
	}
	return headerorder.Sorted(headers, order, nil)
}

func shouldSendRequestContentLength(req Request) bool {
	if omitZeroContentLength(req.Header, req.ActualContentLength) {
		return false
	}
	if req.ActualContentLength == 0 && req.HasReplayBody {
		for _, name := range req.Header[HeaderOrderKey] {
			if strings.EqualFold(name, "content-length") {
				return true
			}
		}
	}
	return shouldSendReqContentLength(req.Method, req.ActualContentLength)
}

// An explicitly empty field opts out only when the sanitized body length is zero.
// It never changes framing for a nonempty or unknown-length body.
func omitZeroContentLength(headers map[string][]string, contentLength int64) bool {
	if contentLength != 0 {
		return false
	}
	for name, values := range headers {
		if strings.EqualFold(name, "content-length") && len(values) == 0 {
			return true
		}
	}
	return false
}

// Return false to retain the upstream pseudo-header defaults when no order is supplied.
func writeOrderedPseudoHeaders(req Request, host, path string, normalConnect bool, protocol string, f func(string, string)) bool {
	order, ok := req.Header[PHeaderOrderKey]
	if !ok {
		return false
	}
	method := req.Method
	if method == "" {
		method = "GET"
	}
	for _, name := range order {
		switch name {
		case ":authority":
			f(name, host)
		case ":method":
			f(name, method)
		case ":path":
			if !normalConnect {
				f(name, path)
			}
		case ":scheme":
			if !normalConnect {
				f(name, req.URL.Scheme)
			}
		}
	}
	// The original fork places Extended CONNECT's :protocol after the ordered fields.
	if protocol != "" {
		f(":protocol", protocol)
	}
	return true
}
