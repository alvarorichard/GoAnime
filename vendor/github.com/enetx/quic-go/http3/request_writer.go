package http3

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"maps"
	"net"
	"strconv"
	"strings"
	"sync"

	"github.com/enetx/http"
	"github.com/enetx/http/httptrace"

	httpcommon "github.com/enetx/http/headerorder"
	"github.com/enetx/quic-go"
	"github.com/enetx/quic-go/http3/qlog"
	"github.com/enetx/quic-go/qlogwriter"
	"github.com/quic-go/qpack"
	"golang.org/x/net/http/httpguts"
	"golang.org/x/net/idna"
)

const bodyCopyBufferSize = 8 * 1024

type requestWriter struct {
	peerMaxFieldSectionSize func() int64
	encodeQPACK             func(quic.StreamID, []qpack.HeaderField) ([]byte, error)
	qpackFields             []qpack.HeaderField
	collectBodyFields       bool
	prepareRequestHeaders   func(io.Writer) (io.Writer, func())
	beforeRequestHeaders    func(*http.Request, quic.StreamID) error
	writeRequestHeaders     func(io.Writer, *http.Request, quic.StreamID, []byte) error
	greaseFrames            func(*http.Request) ([]GreaseFrame, error)
	qpackConfig             *QPACKConfig
	mutex                   sync.Mutex
	encoder                 *qpack.Encoder
	headerBuf               *bytes.Buffer
}

func newRequestWriter() *requestWriter {
	headerBuf := &bytes.Buffer{}
	encoder := qpack.NewEncoder(headerBuf)
	return &requestWriter{
		encoder:   encoder,
		headerBuf: headerBuf,
	}
}

func (w *requestWriter) WriteRequestHeader(wr io.Writer, req *http.Request, gzip bool, streamID quic.StreamID, qlogger qlogwriter.Recorder, prepareBody ...func([]qpack.HeaderField) error) error {
	if w.prepareRequestHeaders != nil {
		var finish func()
		wr, finish = w.prepareRequestHeaders(wr)
		defer finish()
	}

	buf := &bytes.Buffer{}
	if err := w.writeHeaders(buf, req, gzip, streamID, qlogger, prepareBody...); err != nil {
		return err
	}
	if w.writeRequestHeaders != nil {
		if err := w.writeRequestHeaders(wr, req, streamID, buf.Bytes()); err != nil {
			return err
		}
	} else if _, err := wr.Write(buf.Bytes()); err != nil {
		return err
	}
	trace := httptrace.ContextClientTrace(req.Context())
	traceWroteHeaders(trace)
	return nil
}

func (w *requestWriter) writeHeaders(wr io.Writer, req *http.Request, gzip bool, streamID quic.StreamID, qlogger qlogwriter.Recorder, prepareBody ...func([]qpack.HeaderField) error) error {
	w.mutex.Lock()
	defer w.mutex.Unlock()
	defer w.encoder.Close()
	defer w.headerBuf.Reset()
	defer func() {
		clear(w.qpackFields)
		w.qpackFields = w.qpackFields[:0]
		w.collectBodyFields = false
	}()
	w.collectBodyFields = len(prepareBody) > 0 && prepareBody[0] != nil

	var trailers string
	if len(req.Trailer) > 0 {
		keys := make([]string, 0, len(req.Trailer))
		for _, kv := range httpcommon.Sorted(req.Trailer, httpcommon.Indices(req.Trailer[httpcommon.HeaderOrderKey]), nil) {
			if !httpguts.ValidHeaderFieldName(kv.Key) {
				return fmt.Errorf("invalid HTTP trailer name %q", kv.Key)
			}
			if httpguts.ValidTrailerHeader(kv.Key) {
				keys = append(keys, kv.Key)
			}
		}
		trailers = strings.Join(keys, ", ")
	}

	headerFields, err := w.encodeHeaders(req, gzip, trailers, actualContentLength(req), qlogger != nil)
	if err != nil {
		return err
	}
	if w.collectBodyFields {
		if err := prepareBody[0](w.qpackFields); err != nil {
			return &RequestPreparationError{err}
		}
	}
	var grease []byte
	if w.greaseFrames != nil {
		frames, err := w.greaseFrames(req)
		if err != nil {
			return &RequestPreparationError{err}
		}
		grease, err = appendGreaseFrames(frames)
		if err != nil {
			return &RequestPreparationError{err}
		}
	}
	if w.beforeRequestHeaders != nil {
		if err := w.beforeRequestHeaders(req, streamID); err != nil {
			return &RequestPreparationError{err}
		}
	}
	if w.encodeQPACK != nil {
		block, err := w.encodeQPACK(streamID, w.qpackFields)
		if err != nil {
			return &RequestPreparationError{err}
		}
		w.headerBuf.Write(block)
	}
	if len(grease) > 0 {
		recordWireFrames(qlogger, streamID, grease)
		if _, err := wr.Write(grease); err != nil {
			return err
		}
	}

	b := make([]byte, 0, 128)
	b = (&headersFrame{Length: uint64(w.headerBuf.Len())}).Append(b)
	if qlogger != nil {
		qlogCreatedHeadersFrame(qlogger, streamID, len(b)+w.headerBuf.Len(), w.headerBuf.Len(), headerFields)
	}
	if _, err := wr.Write(b); err != nil {
		return err
	}
	_, err = wr.Write(w.headerBuf.Bytes())
	return err
}

func isExtendedConnectRequest(req *http.Request) bool {
	return req.Method == http.MethodConnect && req.Proto != "" && req.Proto != "HTTP/1.1"
}

// copied from net/transport.go
// Modified to support Extended CONNECT:
// Contrary to what the godoc for the http.Request says,
// we do respect the Proto field if the method is CONNECT.
//
// The returned header fields are only set if doQlog is true.
func (w *requestWriter) encodeHeaders(req *http.Request, addGzipHeader bool, trailers string, contentLength int64, doQlog bool) ([]qlog.HeaderField, error) {
	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	host, err := httpguts.PunycodeHostPort(host)
	if err != nil {
		return nil, err
	}
	if !httpguts.ValidHostHeader(host) {
		return nil, errors.New("http3: invalid Host header")
	}

	// http.NewRequest sets this field to HTTP/1.1
	isExtendedConnect := isExtendedConnectRequest(req)
	if isExtendedConnect && !validExtendedConnectProtocol(req.Proto) {
		return nil, fmt.Errorf("invalid request :protocol %q", req.Proto)
	}

	var path string
	if req.Method != http.MethodConnect || isExtendedConnect {
		path = req.URL.RequestURI()
		if !validPseudoPath(path) {
			orig := path
			path = strings.TrimPrefix(path, req.URL.Scheme+"://"+host)
			if !validPseudoPath(path) {
				if req.URL.Opaque != "" {
					return nil, fmt.Errorf("invalid request :path %q from URL.Opaque = %q", orig, req.URL.Opaque)
				} else {
					return nil, fmt.Errorf("invalid request :path %q", orig)
				}
			}
		}
	}

	// Check for any invalid headers and return an error before we
	// potentially pollute our hpack state. (We want to be able to
	// continue to reuse the hpack encoder for future requests)
	for k, vv := range req.Header {
		// Skip validation for special header order keys
		if httpcommon.IsMetadataKey(k) {
			continue
		}
		if !httpguts.ValidHeaderFieldName(k) {
			return nil, fmt.Errorf("invalid HTTP header name %q", k)
		}
		for _, v := range vv {
			if !httpguts.ValidHeaderFieldValue(v) {
				return nil, fmt.Errorf("invalid HTTP header value %q for header %q", v, k)
			}
		}
	}

	order := httpcommon.Indices(req.Header[httpcommon.HeaderOrderKey])
	_, orderedContentLength := order["content-length"]
	sendContentLength := shouldSendReqContentLength(req.Method, contentLength)
	if contentLength == 0 && req.GetBody != nil && orderedContentLength {
		// An explicit slot can frame a replayable zero-byte upload without
		// changing the upstream defaults for requests without ordered metadata.
		sendContentLength = true
	}
	sendContentLength = sendContentLength && !omitZeroContentLength(req.Header, contentLength)
	orderedContentLength = orderedContentLength && sendContentLength
	headers := req.Header
	if orderedContentLength {
		// Only an explicit slot changes the upstream implicit-field order.
		// Framing uses the actual body length, never a supplied header value.
		headers = maps.Clone(req.Header)
		for name := range headers {
			if strings.EqualFold(name, "content-length") {
				delete(headers, name)
			}
		}
		headers["content-length"] = []string{strconv.FormatInt(contentLength, 10)}
	}
	// Share the same deterministic order as HTTP/1 and HTTP/2.
	kvs := httpcommon.Sorted(headers, order, nil)

	enumerateHeaders := func(f func(name, value string)) {
		// Handle pseudo-headers with order support
		pHeaderOrder, hasPHeaderOrder := req.Header[httpcommon.PHeaderOrderKey]

		if hasPHeaderOrder {
			// Follow pseudo header order
			for _, p := range pHeaderOrder {
				switch p {
				case ":authority":
					f(":authority", host)
				case ":method":
					f(":method", req.Method)
				case ":path":
					if req.Method != http.MethodConnect || isExtendedConnect {
						f(":path", path)
					}
				case ":scheme":
					if req.Method != http.MethodConnect || isExtendedConnect {
						f(":scheme", req.URL.Scheme)
					}
				case ":protocol":
					if isExtendedConnect {
						f(":protocol", req.Proto)
					}
				default:
					continue
				}
			}
		} else {
			// Default pseudo-header order
			f(":authority", host)
			f(":method", req.Method)
			if req.Method != http.MethodConnect || isExtendedConnect {
				f(":path", path)
				f(":scheme", req.URL.Scheme)
			}
			if isExtendedConnect {
				f(":protocol", req.Proto)
			}
		}

		if trailers != "" {
			f("trailer", trailers)
		}

		var didUA bool
		for _, kv := range kvs {
			k := kv.Key
			vv := kv.Values

			if strings.EqualFold(k, "host") || (strings.EqualFold(k, "content-length") && !orderedContentLength) {
				// Host is :authority, already sent.
				// Unordered Content-Length is automatic, set below.
				continue
			} else if strings.EqualFold(k, "connection") || strings.EqualFold(k, "proxy-connection") ||
				strings.EqualFold(k, "transfer-encoding") || strings.EqualFold(k, "upgrade") ||
				strings.EqualFold(k, "keep-alive") {
				// Per 8.1.2.2 Connection-Specific Header
				// Fields, don't send connection-specific
				// fields. We have already checked if any
				// are error-worthy so just ignore the rest.
				continue
			} else if strings.EqualFold(k, "user-agent") {
				// Match Go's http1 behavior: at most one
				// User-Agent. If set to nil or empty string,
				// then omit it. Otherwise if not mentioned,
				// include the default (below).
				didUA = true
				if len(vv) < 1 {
					continue
				}
				vv = vv[:1]
				if vv[0] == "" {
					continue
				}

			}

			for _, v := range vv {
				if w.qpackConfig != nil && w.qpackConfig.SplitCookies && strings.EqualFold(k, "cookie") {
					for _, crumb := range strings.Split(v, ";") {
						if crumb = strings.TrimSpace(crumb); crumb != "" {
							f(k, crumb)
						}
					}
				} else {
					f(k, v)
				}
			}
		}

		if sendContentLength && !orderedContentLength {
			f("content-length", strconv.FormatInt(contentLength, 10))
		}
		if addGzipHeader {
			f("accept-encoding", "gzip")
		}
		if !didUA {
			f("user-agent", defaultUserAgent)
		}
	}

	// Count uncompressed fields before mutating the connection's QPACK state.
	hlSize := uint64(0)
	enumerateHeaders(func(name, value string) {
		hlSize += uint64(len(name)) + uint64(len(value)) + 32
	})

	if err := w.checkFieldSectionSize(hlSize); err != nil {
		return nil, err
	}

	trace := httptrace.ContextClientTrace(req.Context())
	traceHeaders := traceHasWroteHeaderField(trace)

	// Header list size is ok. Write the headers.
	var headerFields []qlog.HeaderField
	if doQlog {
		headerFields = make([]qlog.HeaderField, 0, len(req.Header))
	}
	var encodeErr error
	enumerateHeaders(func(name, value string) {
		name = strings.ToLower(name)
		if encodeErr != nil {
			return
		}
		encodeErr = w.writeQPACKField(name, value)
		if traceHeaders {
			traceWroteHeaderField(trace, name, value)
		}
		if doQlog {
			headerFields = append(headerFields, qlog.HeaderField{Name: name, Value: value})
		}
	})

	return headerFields, encodeErr
}

// authorityAddr returns a given authority (a host/IP, or host:port / ip:port)
// and returns a host:port. The port 443 is added if needed.
func authorityAddr(authority string) (addr string) {
	host, port, err := net.SplitHostPort(authority)
	if err != nil { // authority didn't have a port
		port = "443"
		host = authority
	}
	if a, err := idna.ToASCII(host); err == nil {
		host = a
	}
	// IPv6 address literal, without a port:
	if strings.HasPrefix(host, "[") && strings.HasSuffix(host, "]") {
		return host + ":" + port
	}
	return net.JoinHostPort(host, port)
}

// validPseudoPath reports whether v is a valid :path pseudo-header
// value. It must be either:
//
//	*) a non-empty string starting with '/'
//	*) the string '*', for OPTIONS requests.
//
// For now this is only used a quick check for deciding when to clean
// up Opaque URLs before sending requests from the Transport.
// See golang.org/issue/16847
//
// We used to enforce that the path also didn't start with "//", but
// Google's GFE accepts such paths and Chrome sends them, so ignore
// that part of the spec. See golang.org/issue/19103.
func validPseudoPath(v string) bool {
	return (len(v) > 0 && v[0] == '/') || v == "*"
}

// actualContentLength returns a sanitized version of
// req.ContentLength, where 0 actually means zero (not unknown) and -1
// means unknown.
func actualContentLength(req *http.Request) int64 {
	if req.Body == nil || req.Body == http.NoBody {
		return 0
	}
	if req.ContentLength != 0 {
		return req.ContentLength
	}
	return -1
}

// An explicitly empty field opts out only when the sanitized body length is zero.
// It never changes framing for a nonempty or unknown-length body.
func omitZeroContentLength(headers http.Header, contentLength int64) bool {
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

// shouldSendReqContentLength reports whether the http2.Transport should send
// a "content-length" request header. This logic is basically a copy of the net/http
// transferWriter.shouldSendContentLength.
// The contentLength is the corrected contentLength (so 0 means actually 0, not unknown).
// -1 means unknown.
func shouldSendReqContentLength(method string, contentLength int64) bool {
	if contentLength > 0 {
		return true
	}
	if contentLength < 0 {
		return false
	}
	// For zero bodies, whether we send a content-length depends on the method.
	// It also kinda doesn't matter for http2 either way, with END_STREAM.
	switch method {
	case "POST", "PUT", "PATCH":
		return true
	default:
		return false
	}
}

// WriteRequestTrailer writes HTTP trailers to the stream.
// It should be called after the request body has been fully written.
func (w *requestWriter) WriteRequestTrailer(wr io.Writer, req *http.Request, streamID quic.StreamID, qlogger qlogwriter.Recorder) error {
	var size uint64
	for _, kv := range httpcommon.Sorted(req.Trailer, httpcommon.Indices(req.Trailer[httpcommon.HeaderOrderKey]), nil) {
		if httpguts.ValidTrailerHeader(kv.Key) {
			for _, value := range kv.Values {
				size += uint64(len(kv.Key)) + uint64(len(value)) + 32
			}
		}
	}
	if err := w.checkFieldSectionSize(size); err != nil {
		return err
	}
	if w.encodeQPACK != nil {
		var fields []qpack.HeaderField
		var logged []qlog.HeaderField
		for _, kv := range httpcommon.Sorted(req.Trailer, httpcommon.Indices(req.Trailer[httpcommon.HeaderOrderKey]), nil) {
			if !httpguts.ValidHeaderFieldName(kv.Key) {
				return fmt.Errorf("invalid HTTP trailer name %q", kv.Key)
			}
			if !httpguts.ValidTrailerHeader(kv.Key) {
				continue
			}
			name := strings.ToLower(kv.Key)
			for _, value := range kv.Values {
				if !httpguts.ValidHeaderFieldValue(value) {
					return fmt.Errorf("invalid HTTP trailer value %q for header %q", value, kv.Key)
				}
				fields = append(fields, qpack.HeaderField{Name: name, Value: value})
				if qlogger != nil {
					logged = append(logged, qlog.HeaderField{Name: name, Value: value})
				}
			}
		}
		if len(fields) == 0 {
			return nil
		}
		block, err := w.encodeQPACK(streamID, fields)
		if err != nil {
			return err
		}
		frame := (&headersFrame{Length: uint64(len(block))}).Append(nil)
		if qlogger != nil {
			qlogCreatedHeadersFrame(qlogger, streamID, len(frame)+len(block), len(block), logged)
		}
		_, err = wr.Write(append(frame, block...))
		return err
	}
	_, err := writeTrailers(wr, req.Trailer, streamID, qlogger)
	return err
}

func (w *requestWriter) checkFieldSectionSize(size uint64) error {
	if w.peerMaxFieldSectionSize != nil {
		if limit := w.peerMaxFieldSectionSize(); limit >= 0 && size > uint64(limit) {
			return &RequestPreparationError{fmt.Errorf("http3: request field section size %d exceeds peer limit %d", size, limit)}
		}
	}
	return nil
}
