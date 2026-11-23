package http3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/textproto"
	"sync"
	"time"

	"github.com/enetx/http"
	"github.com/enetx/http/httptrace"

	"github.com/enetx/g"
	"github.com/enetx/quic-go"
	"github.com/enetx/quic-go/qlogwriter"
	"github.com/quic-go/qpack"
)

const (
	// MethodGet0RTT allows a GET request to be sent using 0-RTT.
	// Note that 0-RTT doesn't provide replay protection and should only be used for idempotent requests.
	MethodGet0RTT = "GET_0RTT"
	// MethodHead0RTT allows a HEAD request to be sent using 0-RTT.
	// Note that 0-RTT doesn't provide replay protection and should only be used for idempotent requests.
	MethodHead0RTT = "HEAD_0RTT"
)

const (
	defaultUserAgent              = "quic-go HTTP/3"
	defaultMaxResponseHeaderBytes = 10 * 1 << 20 // 10 MB
)

var errGoAway = errors.New("connection in graceful shutdown")

type errConnUnusable struct{ e error }

func (e *errConnUnusable) Unwrap() error { return e.e }
func (e *errConnUnusable) Error() string { return fmt.Sprintf("http3: conn unusable: %s", e.e.Error()) }

const max1xxResponses = 5 // arbitrary bound on number of informational responses

var defaultQuicConfig = &quic.Config{
	MaxIncomingStreams: -1, // don't allow the server to create bidirectional streams
	KeepAlivePeriod:    10 * time.Second,
}

// ClientConn is an HTTP/3 client doing requests to a single remote server.
type ClientConn struct {
	conn                   *quic.Conn
	rawConn                *rawConn
	initialSettingsWritten chan struct{}
	initialSettingsErr     error           // published by closing initialSettingsWritten
	initialInputReady      <-chan struct{} // buffered peer input has been dispatched; never waits for new data

	decoder fieldDecoder

	// Additional HTTP/3 settings.
	// It is invalid to specify any settings defined by RFC 9114 (HTTP/3) and RFC 9297 (HTTP Datagrams).
	additionalSettings g.MapOrd[uint64, uint64]

	// maxResponseHeaderBytes specifies a limit on how many response bytes are
	// allowed in the server's response header.
	maxResponseHeaderBytes int

	// disableCompression, if true, prevents the Transport from requesting compression with an
	// "Accept-Encoding: gzip" request header when the Request contains no existing Accept-Encoding value.
	// If the Transport requests gzip on its own and gets a gzipped response, it's transparently
	// decoded in the Response.Body.
	// However, if the user explicitly requested gzip it is not automatically uncompressed.
	disableCompression bool

	streamMx     sync.Mutex
	maxStreamID  quic.StreamID // set once a GOAWAY frame is received
	goAwayCtx    context.Context
	goAwayCancel context.CancelFunc

	qlogger qlogwriter.Recorder
	logger  *slog.Logger

	requestWriter  *requestWriter
	prepareHeaders func(*http.Request, *quic.Conn) error
	streamPriority *StreamPriority
	wireConfig     *WireConfig
	controlWriter  *clientControlWriter
}

var _ http.RoundTripper = &ClientConn{}

func newClientConn(
	conn *quic.Conn,
	enableDatagrams bool,
	additionalSettings g.MapOrd[uint64, uint64],
	maxResponseHeaderBytes int,
	disableCompression bool,
	logger *slog.Logger,
	qpackConfig ...*QPACKConfig,
) *ClientConn {
	return newClientConnWithWire(conn, enableDatagrams, additionalSettings, maxResponseHeaderBytes, disableCompression, logger, nil, qpackConfig...)
}

func newClientConnWithWire(
	conn *quic.Conn,
	enableDatagrams bool,
	additionalSettings g.MapOrd[uint64, uint64],
	maxResponseHeaderBytes int,
	disableCompression bool,
	logger *slog.Logger,
	wireConfig *WireConfig,
	qpackConfig ...*QPACKConfig,
) *ClientConn {
	var qlogger qlogwriter.Recorder
	if qlogTrace := conn.QlogTrace(); qlogTrace != nil {
		qlogger = qlogTrace.AddProducer()
	}
	c := &ClientConn{
		initialSettingsWritten: make(chan struct{}),
		conn:                   conn,
		additionalSettings:     additionalSettings,
		disableCompression:     disableCompression,
		maxStreamID:            invalidStreamID,
		logger:                 logger,
		qlogger:                qlogger,
		decoder:                qpack.NewDecoder(),
	}
	c.goAwayCtx, c.goAwayCancel = context.WithCancel(context.Background())
	if maxResponseHeaderBytes <= 0 {
		c.maxResponseHeaderBytes = defaultMaxResponseHeaderBytes
	} else {
		c.maxResponseHeaderBytes = maxResponseHeaderBytes
	}
	c.requestWriter = newRequestWriter()
	c.rawConn = newRawConn(
		conn,
		enableDatagrams,
		c.onStreamsEmpty,
		c.handleControlStream,
		qlogger,
		c.logger,
	)
	c.rawConn.strictControlSettings = wireConfig != nil && wireConfig.OnControlFrame != nil
	c.requestWriter.peerMaxFieldSectionSize = c.rawConn.peerFieldSectionLimit
	if len(qpackConfig) > 0 && qpackConfig[0] != nil {
		c.rawConn.clientQPACK = newClientQPACK(c.rawConn, qpackConfig[0], c.initialSettingsWritten,
			qpackFeedbackQueueSize(additionalSettings.Get(1).UnwrapOrDefault(), additionalSettings.Get(7).UnwrapOrDefault()))
		c.requestWriter.encodeQPACK = c.rawConn.clientQPACK.encode
		c.requestWriter.qpackConfig = qpackConfig[0].Clone()
	}
	c.configureWire(wireConfig)
	c.requestWriter.beforeRequestHeaders = c.queuePriorityUpdate
	c.rawConn.configureQPACK(additionalSettings)
	// Send SETTINGS in 0-RTT when available, otherwise stage the control
	// stream before the first post-handshake flight if explicitly requested.
	if c.wireConfig.CoalesceInitialSettings || c.wireConfig.FlushEarlySettings {
		c.startInitialSettings(enableDatagrams, additionalSettings, true)
	} else {
		go c.startInitialSettings(enableDatagrams, additionalSettings, false)
	}

	return c
}

// OpenRequestStream opens a new request stream on the HTTP/3 connection.
func (c *ClientConn) OpenRequestStream(ctx context.Context) (*RequestStream, error) {
	if err := c.preparationFailure(nil); err != nil {
		return nil, err
	}

	if err := c.rawConn.initializePeerSettings(ctx); err != nil {
		return nil, c.preparationFailure(err)
	}

	return c.openRequestStream(ctx, c.requestWriter, make(chan struct{}), c.disableCompression, c.maxResponseHeaderBytes)
}

func (c *ClientConn) openRequestStream(
	ctx context.Context,
	requestWriter *requestWriter,
	reqDone chan struct{},
	disableCompression bool,
	maxHeaderBytes int,
) (*RequestStream, error) {
	if err := c.waitInitialSettings(ctx); err != nil {
		return nil, err
	}
	// RFC 9114 Section 5.2 prohibits opening any new request streams after GOAWAY.
	// The stream ID only identifies requests that were already in flight and might still be processed.
	if c.goAwayCtx.Err() != nil {
		return nil, errGoAway
	}

	openCtx, cancel := context.WithCancelCause(c.rawConn.context())
	stopRequest := context.AfterFunc(ctx, func() { cancel(context.Cause(ctx)) })
	defer stopRequest()
	if ctx.Err() != nil {
		cancel(context.Cause(ctx))
	}
	// A request blocked in OpenStreamSync has no request stream yet, so it is not in flight.
	stop := context.AfterFunc(c.goAwayCtx, func() { cancel(errGoAway) })
	str, err := c.conn.OpenStreamSync(openCtx)
	stop()
	cancel(nil)
	if err != nil {
		if errors.Is(context.Cause(c.rawConn.context()), quic.Err0RTTRejected) {
			return nil, quic.Err0RTTRejected
		}
		if context.Cause(openCtx) == errGoAway {
			return nil, errGoAway
		}
		return nil, maybeReplaceError(err)
	}

	// Check again in case GOAWAY raced with OpenStreamSync.
	if c.goAwayCtx.Err() != nil {
		str.CancelRead(quic.StreamErrorCode(ErrCodeRequestCanceled))
		str.CancelWrite(quic.StreamErrorCode(ErrCodeRequestCanceled))
		return nil, errGoAway
	}

	c.setStreamPriority(ctx, str)
	hstr := c.rawConn.trackStream(str, true)
	if hstr == nil {
		str.CancelRead(quic.StreamErrorCode(ErrCodeRequestCanceled))
		str.CancelWrite(quic.StreamErrorCode(ErrCodeRequestCanceled))
		return nil, errGoAway
	}
	decoder := bindFieldDecoder(c.decoder, c.rawConn, ctx, hstr, c.maxResponseHeaderBytes)
	rsp := &http.Response{}
	trace := httptrace.ContextClientTrace(ctx)
	stream := newRequestStream(
		newStream(hstr, c.rawConn, trace, func(r io.Reader, hf *headersFrame) error {
			hdr, err := decodeTrailers(r, hf, maxHeaderBytes, decoder, c.qlogger, str.StreamID())
			if err != nil {
				return err
			}
			rsp.Trailer = hdr
			return nil
		}, c.qlogger),
		requestWriter,
		reqDone,
		decoder,
		disableCompression,
		maxHeaderBytes,
		rsp,
	)
	if c.wireConfig != nil {
		stream.str.maxDataFrameSize = c.wireConfig.MaxDataFrameSize
		stream.bodyBufferSize = c.wireConfig.RequestBodyBufferSize
		stream.bodyBufferSizePolicy = c.wireConfig.RequestBodyBufferSizePolicy
		stream.bodyFirstReadSizePolicy = c.wireConfig.RequestBodyFirstReadSizePolicy
	}
	stream.priorityUpdater = &PriorityUpdater{request: ctx, connection: c.conn.Context(), done: reqDone,
		send: func(u int8, incremental bool) error { return c.enqueuePriorityUpdate(str.StreamID(), u, incremental) }}
	return stream, nil
}

func (c *ClientConn) handleUnidirectionalStream(str *quic.ReceiveStream) {
	c.rawConn.handleUnidirectionalStream(str, false)
}

func (c *ClientConn) handleControlStream(str *quic.ReceiveStream, fp *frameParser) {
	if c.wireConfig != nil && c.wireConfig.OnControlFrame != nil {
		fp.onUnknownFrame = func(frame ControlFrame) error {
			return c.wireConfig.OnControlFrame(c.conn, frame)
		}
	}
	for {
		f, err := fp.ParseNext(c.qlogger)
		if err != nil {
			if errors.Is(err, errPriorityUpdateForPush) {
				c.rawConn.CloseWithError(quic.ApplicationErrorCode(ErrCodeFrameUnexpected), "")
				return
			}
			_, isStreamError := errors.AsType[*quic.StreamError](err)
			if err == io.EOF || isStreamError {
				c.rawConn.CloseWithError(quic.ApplicationErrorCode(ErrCodeClosedCriticalStream), "")
				return
			}
			c.rawConn.CloseWithError(quic.ApplicationErrorCode(ErrCodeFrameError), "")
			return
		}
		// GOAWAY is the only frame allowed at this point:
		// * unexpected frames are ignored by the frame parser
		// * we don't support any extension that might add support for more frames
		goaway, ok := f.(*goAwayFrame)
		if !ok {
			c.rawConn.CloseWithError(quic.ApplicationErrorCode(ErrCodeFrameUnexpected), "")
			return
		}
		if goaway.StreamID%4 != 0 { // client-initiated, bidirectional streams
			c.rawConn.CloseWithError(quic.ApplicationErrorCode(ErrCodeIDError), "")
			return
		}
		c.streamMx.Lock()
		// the server is not allowed to increase the Stream ID in subsequent GOAWAY frames
		if c.maxStreamID != invalidStreamID && goaway.StreamID > c.maxStreamID {
			c.streamMx.Unlock()
			c.rawConn.CloseWithError(quic.ApplicationErrorCode(ErrCodeIDError), "")
			return
		}
		c.maxStreamID = goaway.StreamID
		c.goAwayCancel()
		c.streamMx.Unlock()

		hasActiveStreams := c.rawConn.hasActiveStreams()
		// immediately close the connection if there are currently no active requests
		if !hasActiveStreams {
			c.CloseWithError(quic.ApplicationErrorCode(ErrCodeNoError), "")
			return
		}
	}
}

func (c *ClientConn) onStreamsEmpty() {
	c.streamMx.Lock()
	defer c.streamMx.Unlock()

	// The server is performing a graceful shutdown.
	if c.maxStreamID != invalidStreamID || c.rawConn.retiring.Load() {
		c.rawConn.CloseWithError(quic.ApplicationErrorCode(ErrCodeNoError), "")
	}
}

// retire stops admission atomically with stream tracking. Active response bodies
// retain their connection until the last stream ends; no GOAWAY is sent.
func (c *ClientConn) retire() {
	c.rawConn.streamMx.Lock()
	c.rawConn.retiring.Store(true)
	c.goAwayCancel()
	idle := len(c.rawConn.streams) == 0
	c.rawConn.streamMx.Unlock()

	if idle {
		c.rawConn.CloseWithError(quic.ApplicationErrorCode(ErrCodeNoError), "")
	}
}

// RoundTrip executes a request and returns a response
func (c *ClientConn) RoundTrip(req *http.Request) (*http.Response, error) {
	if c.prepareHeaders != nil || c.rawConn.peerSettings != nil {
		req = requestWithPreparationBody(req)
	}

	rsp, err := c.roundTrip(req)
	if err != nil && (c.prepareHeaders != nil || c.rawConn.peerSettings != nil) && req.Body != nil {
		_, local := errors.AsType[*RequestPreparationError](err)
		_, unopened := errors.AsType[*errConnUnusable](err)
		if local || req.Context().Err() != nil || !unopened && !errors.Is(err, quic.Err0RTTRejected) {
			req.Body.Close()
		}
	}

	if err != nil && req.Context().Err() != nil {
		// if the context was canceled, return the context cancellation error
		err = req.Context().Err()
	}
	return rsp, err
}

func (c *ClientConn) roundTrip(req *http.Request) (*http.Response, error) {
	early := c.pendingEarlyRequest(req)
	// Immediately send out this request, if this is a 0-RTT request.
	switch req.Method {
	case MethodGet0RTT:
		// don't modify the original request
		reqCopy := *req
		req = &reqCopy
		req.Method = http.MethodGet
	case MethodHead0RTT:
		// don't modify the original request
		reqCopy := *req
		req = &reqCopy
		req.Method = http.MethodHead
	}
	if !early {
		// wait for the handshake to complete
		select {
		case <-c.conn.HandshakeComplete():
		case <-c.rawConn.context().Done():
			if err := c.handshakeCancellationError(); err != nil {
				return nil, err
			}
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
		if c.initialInputReady != nil {
			select {
			case <-c.initialInputReady:
			case <-c.rawConn.context().Done():
				return nil, &errConnUnusable{e: c.preparationFailure(context.Cause(c.rawConn.context()))}
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}
		if err := c.preparationFailure(nil); err != nil {
			return nil, requestInitializationError(err)
		}

		if err := c.rawConn.initializePeerSettings(req.Context()); err != nil {
			return nil, requestInitializationError(c.preparationFailure(err))
		}
	}
	var err error
	req, err = c.prepareRequestHeaders(req, early)
	if err != nil {
		return nil, err
	}

	// It is only possible to send an Extended CONNECT request once the SETTINGS were received.
	// See section 3 of RFC 8441.
	if isExtendedConnectRequest(req) {
		connCtx := c.rawConn.context()
		// wait for the server's SETTINGS frame to arrive
		select {
		case <-c.rawConn.ReceivedSettings():
		case <-connCtx.Done():
			if errors.Is(context.Cause(connCtx), quic.Err0RTTRejected) {
				return nil, &errConnUnusable{e: quic.Err0RTTRejected}
			}
			return nil, context.Cause(connCtx)
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
		if !c.rawConn.Settings().EnableExtendedConnect {
			return nil, errors.New("http3: server didn't enable Extended CONNECT")
		}
	}

	c.rawConn.AddQlogProducer()
	defer c.rawConn.RemoveQlogProducer()

	reqDone := make(chan struct{})
	str, err := c.openRequestStream(
		req.Context(),
		c.requestWriter,
		reqDone,
		c.disableCompression,
		c.maxResponseHeaderBytes,
	)
	if err != nil {
		return nil, &errConnUnusable{e: err}
	}

	// Request Cancellation:
	// This go routine keeps running even after RoundTripOpt() returns.
	// It is shut down when the application is done processing the body.
	done := make(chan struct{})
	go func() {
		defer close(done)
		select {
		case <-req.Context().Done():
			str.CancelWrite(quic.StreamErrorCode(ErrCodeRequestCanceled))
			str.CancelRead(quic.StreamErrorCode(ErrCodeRequestCanceled))
		case <-reqDone:
		}
	}()

	rsp, err := c.doRequest(req, str)
	if err != nil { // if any error occurred
		close(reqDone)
		<-done
		return nil, err
	}
	return rsp, nil
}

// ReceivedSettings returns a channel that is closed once the server's HTTP/3 settings were received.
// The settings can be obtained from [ClientConn.Settings] after the channel is closed.
func (c *ClientConn) ReceivedSettings() <-chan struct{} {
	return c.rawConn.ReceivedSettings()
}

// Settings returns the HTTP/3 settings for this connection.
// It is only valid to call this method after the channel returned by [ClientConn.ReceivedSettings] is closed.
func (c *ClientConn) Settings() *Settings {
	return c.rawConn.Settings()
}

// CloseWithError closes the connection with the given error code and message.
// It is invalid to call this function after the connection was closed.
func (c *ClientConn) CloseWithError(code quic.ApplicationErrorCode, msg string) error {
	return c.conn.CloseWithError(code, msg)
}

// Context returns a context that is cancelled when the connection is closed.
func (c *ClientConn) Context() context.Context {
	return c.conn.Context()
}

// LocalAddr returns the local address of the underlying QUIC connection.
func (c *ClientConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

// RemoteAddr returns the remote address of the underlying QUIC connection.
func (c *ClientConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// cancelingReader reads from the io.Reader.
// It cancels writing on the stream if any error other than io.EOF occurs.
type cancelingReader struct {
	r   io.Reader
	str *RequestStream
}

func (r *cancelingReader) Read(b []byte) (int, error) {
	n, err := r.r.Read(b)
	if err != nil && err != io.EOF {
		r.str.CancelWrite(quic.StreamErrorCode(ErrCodeRequestCanceled))
	}
	return n, err
}

func (c *ClientConn) sendRequestBody(str *RequestStream, body io.ReadCloser, contentLength int64) error {
	body = &requestBodyCloser{ReadCloser: body}
	stop := context.AfterFunc(str.Context(), func() { body.Close() })
	defer stop()
	defer body.Close()
	size := str.bodyBufferSize
	if size == 0 {
		size = bodyCopyBufferSize
	}
	buf := make([]byte, size)
	var sr io.Reader = &cancelingReader{str: str, r: body}
	if str.bodyFirstReadSize > 0 {
		sr = &firstBodyReadReader{r: sr, size: str.bodyFirstReadSize}
	}
	if contentLength == -1 {
		_, err := io.CopyBuffer(str, sr, buf)
		return err
	}

	// make sure we don't send more bytes than the content length
	n, err := io.CopyBuffer(str, io.LimitReader(sr, contentLength), buf)
	if err != nil {
		return err
	}
	var extra int64
	extra, err = io.CopyBuffer(io.Discard, sr, buf)
	n += extra
	if n > contentLength {
		str.CancelWrite(quic.StreamErrorCode(ErrCodeRequestCanceled))
		return fmt.Errorf("http: ContentLength=%d with Body length %d", contentLength, n)
	}
	return err
}

func (c *ClientConn) doRequest(req *http.Request, str *RequestStream) (*http.Response, error) {
	trace := httptrace.ContextClientTrace(req.Context())
	var sendingReqFailed bool
	str.closeAfterHeaders = (req.Body == nil || req.Body == http.NoBody) && len(req.Trailer) == 0
	if err := str.sendRequestHeader(req); err != nil {
		traceWroteRequest(trace, err)
		if _, local := errors.AsType[*RequestPreparationError](err); local {
			str.CancelRead(quic.StreamErrorCode(ErrCodeRequestCanceled))
			str.CancelWrite(quic.StreamErrorCode(ErrCodeRequestCanceled))
			return nil, err
		}
		if c.logger != nil {
			c.logger.Debug("error writing request", "error", err)
		}
		sendingReqFailed = true
	}
	if !sendingReqFailed {
		if req.Body == nil {
			traceWroteRequest(trace, nil)
			str.Close()
		} else {
			// send the request body asynchronously
			go func() {
				defer str.Close()
				contentLength := int64(-1)
				// According to the documentation for http.Request.ContentLength,
				// a value of 0 with a non-nil Body is also treated as unknown content length.
				if req.ContentLength > 0 {
					contentLength = req.ContentLength
				}
				err := c.sendRequestBody(str, req.Body, contentLength)
				traceWroteRequest(trace, err)
				if err != nil {
					if c.logger != nil {
						c.logger.Debug("error writing request", "error", err)
					}
					return
				}

				if len(req.Trailer) > 0 {
					if err := str.sendRequestTrailer(req); err != nil {
						if c.logger != nil {
							c.logger.Debug("error writing trailers", "error", err)
						}
					}
				}
			}()
		}
	}

	// copy from net/http: support 1xx responses
	var num1xx int // number of informational 1xx headers received
	var res *http.Response
	for {
		var err error
		res, err = str.ReadResponse()
		if err != nil {
			if num1xx > 0 {
				return nil, &responseReceivedError{err}
			}
			return nil, err
		}
		if c.wireConfig != nil && c.wireConfig.OnResponseHeaders != nil {
			if err := c.wireConfig.OnResponseHeaders(req, res); err != nil {
				str.CancelRead(quic.StreamErrorCode(ErrCodeRequestCanceled))
				str.CancelWrite(quic.StreamErrorCode(ErrCodeRequestCanceled))
				return nil, &responseReceivedError{err}
			}
		}
		resCode := res.StatusCode
		is1xx := 100 <= resCode && resCode <= 199
		// treat 101 as a terminal status, see https://github.com/golang/go/issues/26161
		is1xxNonTerminal := is1xx && resCode != http.StatusSwitchingProtocols
		if is1xxNonTerminal {
			num1xx++
			if num1xx > max1xxResponses {
				str.CancelRead(quic.StreamErrorCode(ErrCodeExcessiveLoad))
				str.CancelWrite(quic.StreamErrorCode(ErrCodeExcessiveLoad))
				return nil, errors.New("http3: too many 1xx informational responses")
			}
			if err := traceGot1xxResponse(trace, resCode, textproto.MIMEHeader(res.Header)); err != nil {
				str.CancelRead(quic.StreamErrorCode(ErrCodeRequestCanceled))
				str.CancelWrite(quic.StreamErrorCode(ErrCodeRequestCanceled))
				return nil, &responseReceivedError{err}
			}
			if resCode == http.StatusContinue {
				traceGot100Continue(trace)
			}
			continue
		}
		break
	}
	connState := c.conn.ConnectionState().TLS
	res.TLS = &connState
	res.Request = req
	return res, nil
}

// RawClientConn is a low-level HTTP/3 client connection.
// It allows the application to take control of the stream accept loops,
// giving the application the ability to handle streams originating from the server.
type RawClientConn struct {
	*ClientConn
}

// HandleUnidirectionalStream handles an incoming unidirectional stream.
func (c *RawClientConn) HandleUnidirectionalStream(str *quic.ReceiveStream) {
	c.rawConn.handleUnidirectionalStream(str, false)
}

// HandleBidirectionalStream handles an incoming bidirectional stream.
func (c *ClientConn) HandleBidirectionalStream(str *quic.Stream) {
	// According to RFC 9114, the server is not allowed to open bidirectional streams.
	c.rawConn.CloseWithError(
		quic.ApplicationErrorCode(ErrCodeStreamCreationError),
		fmt.Sprintf("server opened bidirectional stream %d", str.StreamID()),
	)
}
