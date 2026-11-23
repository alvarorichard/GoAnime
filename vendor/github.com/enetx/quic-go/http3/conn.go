package http3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"sync/atomic"

	"github.com/enetx/quic-go"
	"github.com/enetx/quic-go/http3/qlog"
	dynamic "github.com/enetx/quic-go/internal/qpack"
	"github.com/enetx/quic-go/qlogwriter"
	"github.com/enetx/quic-go/quicvarint"
)

const maxQuarterStreamID = 1<<60 - 1

// invalidStreamID is a stream ID that is invalid. The first valid stream ID in QUIC is 0.
const invalidStreamID = quic.StreamID(-1)

// rawConn is an HTTP/3 connection.
// It provides HTTP/3 specific functionality by wrapping a quic.Conn,
// in particular handling of unidirectional HTTP/3 streams, SETTINGS and datagrams.
type rawConn struct {
	strictControlSettings bool // Optional extension handlers require SETTINGS first.
	conn                  *quic.Conn
	ctx                   context.Context
	cancel                context.CancelCauseFunc
	generationMx          sync.RWMutex
	qpackDecoder          *dynamic.Decoder
	clientQPACK           *clientQPACK

	logger *slog.Logger

	enableDatagrams bool

	streamMx sync.Mutex
	streams  map[quic.StreamID]*stateTrackingStream
	retiring atomic.Bool

	rcvdControlStr      atomic.Bool
	rcvdQPACKEncoderStr atomic.Bool
	rcvdQPACKDecoderStr atomic.Bool
	controlStrHandler   func(*quic.ReceiveStream, *frameParser) // is called *after* the SETTINGS frame was parsed

	onStreamsEmpty func()

	settings         *Settings
	receivedSettings chan struct{}
	clientSession    *clientSettingsSession
	peerSettings     *clientPeerSettings

	qlogger          qlogwriter.Recorder
	qloggerMx        sync.Mutex
	qloggerProducers int // the open connection, and goroutines and calls that may produce qlog events
}

func newRawConn(
	quicConn *quic.Conn,
	enableDatagrams bool,
	onStreamsEmpty func(),
	controlStrHandler func(*quic.ReceiveStream, *frameParser),
	qlogger qlogwriter.Recorder,
	logger *slog.Logger,
) *rawConn {
	ctx, cancel := context.WithCancelCause(quicConn.Context())
	c := &rawConn{
		conn:              quicConn,
		ctx:               ctx,
		cancel:            cancel,
		logger:            logger,
		enableDatagrams:   enableDatagrams,
		receivedSettings:  make(chan struct{}),
		streams:           make(map[quic.StreamID]*stateTrackingStream),
		qlogger:           qlogger,
		qloggerProducers:  1,
		onStreamsEmpty:    onStreamsEmpty,
		controlStrHandler: controlStrHandler,
	}
	if qlogger != nil {
		context.AfterFunc(ctx, c.RemoveQlogProducer)
	}
	return c
}

func (c *rawConn) context() context.Context {
	if c.ctx != nil {
		return c.ctx
	}
	return c.conn.Context()
}

func (c *rawConn) OpenUniStream() (*quic.SendStream, error) {
	c.generationMx.RLock()
	defer c.generationMx.RUnlock()
	if c.context().Err() != nil {
		return nil, context.Cause(c.context())
	}
	return c.conn.OpenUniStream()
}

func (c *rawConn) retireEarlyData() {
	c.generationMx.Lock()
	c.cancel(quic.Err0RTTRejected)
	c.generationMx.Unlock()
}

// openControlStream opens the control stream and sends the SETTINGS frame.
// It returns the control stream (needed by the server for sending GOAWAY later).
func (c *rawConn) openControlStream(settings *settingsFrame) (*quic.SendStream, error) {
	str, b, err := c.prepareControlStream(settings)
	if err != nil {
		return nil, err
	}
	if _, err := str.Write(b); err != nil {
		return nil, err
	}
	return str, nil
}

// prepareControlStream reserves the stream and builds SETTINGS without a
// blocking write, allowing a client to stage it from the QUIC send loop.
func (c *rawConn) prepareControlStream(settings *settingsFrame) (*quic.SendStream, []byte, error) {
	c.AddQlogProducer()
	defer c.RemoveQlogProducer()

	str, err := c.OpenUniStream()
	if err != nil {
		return nil, nil, err
	}
	b := make([]byte, 0, 64)
	b = quicvarint.Append(b, streamTypeControlStream)
	b = settings.Append(b)
	if c.qlogger != nil {
		c.qlogger.RecordEvent(qlog.StreamTypeSet{
			Initiator:  qlog.InitiatorLocal,
			StreamID:   str.StreamID(),
			StreamType: qlog.StreamTypeControl,
		})
		sf := qlog.SettingsFrame{
			MaxFieldSectionSize: settings.MaxFieldSectionSize,
			Datagram:            -1,
			ExtendedConnect:     -1,
			Other:               settings.Other.Iter().Collect().Map[uint64, uint64](),
		}
		if settings.Datagram {
			sf.Datagram = 1
		}
		if settings.ExtendedConnect {
			sf.ExtendedConnect = 1
		}
		c.qlogger.RecordEvent(qlog.FrameCreated{
			StreamID: str.StreamID(),
			Raw:      qlog.RawInfo{Length: len(b), PayloadLength: qlog.OmitLength},
			Frame:    qlog.Frame{Frame: sf},
		})
	}
	return str, b, nil
}

func (c *rawConn) TrackStream(str *quic.Stream) *stateTrackingStream {
	return c.trackStream(str, false)
}

func (c *rawConn) trackStream(str *quic.Stream, rejectRetired bool) *stateTrackingStream {
	hstr := newStateTrackingStream(str, c, func(b []byte) error { return c.sendDatagram(str.StreamID(), b) })
	hstr.qpackCancel = c.qpackCancellation(str.StreamID())

	c.streamMx.Lock()
	defer c.streamMx.Unlock()
	if rejectRetired && c.retiring.Load() {
		return nil
	}
	c.streams[str.StreamID()] = hstr
	return hstr
}

func (c *rawConn) UpdateStreamPriority(id quic.StreamID, urgency int8, incremental bool) {
	c.streamMx.Lock()
	str := c.streams[id]
	c.streamMx.Unlock()
	// A PRIORITY_UPDATE can arrive before its request stream. We deliberately ignore such reordered frames.
	if str != nil {
		str.SetPriority(urgency, incremental)
	}
}

func (c *rawConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *rawConn) ConnectionState() quic.ConnectionState {
	return c.conn.ConnectionState()
}

func (c *rawConn) clearStream(id quic.StreamID) {
	c.streamMx.Lock()
	defer c.streamMx.Unlock()

	delete(c.streams, id)
	if len(c.streams) == 0 {
		c.onStreamsEmpty()
	}
}

func (c *rawConn) hasActiveStreams() bool {
	c.streamMx.Lock()
	defer c.streamMx.Unlock()

	return len(c.streams) > 0
}

func (c *rawConn) CloseWithError(code quic.ApplicationErrorCode, msg string) error {
	if errors.Is(context.Cause(c.context()), quic.Err0RTTRejected) {
		return nil
	}
	return c.conn.CloseWithError(code, msg)
}

func (c *rawConn) handleUnidirectionalStream(str *quic.ReceiveStream, isServer bool) {
	c.handleUnidirectionalStreamMode(str, isServer, false)
}

func (c *rawConn) handleUnidirectionalStreamMode(str *quic.ReceiveStream, isServer, bufferedSettings bool) {
	// HTTP/3 handles this stream, so the peer can replace it as soon as it completes.
	str.ReleaseStreamCredit()
	c.AddQlogProducer()
	defer c.RemoveQlogProducer()

	streamType, err := quicvarint.Read(quicvarint.NewReader(str))
	if err != nil {
		if c.logger != nil {
			c.logger.Debug("reading stream type on stream failed", "stream ID", str.StreamID(), "error", err)
		}
		return
	}
	if c.qlogger != nil {
		c.qlogger.RecordEvent(qlog.StreamTypeSet{
			Initiator:       qlog.InitiatorRemote,
			StreamID:        str.StreamID(),
			StreamType:      qlogStreamType(streamType),
			StreamTypeBytes: streamType,
		})
	}
	// We're only interested in the control stream here.
	switch streamType {
	case streamTypeControlStream:
	case streamTypeQPACKEncoderStream:
		if isFirst := c.rcvdQPACKEncoderStr.CompareAndSwap(false, true); !isFirst {
			c.CloseWithError(quic.ApplicationErrorCode(ErrCodeStreamCreationError), "duplicate QPACK encoder stream")
			return
		}
		if c.qpackDecoder != nil {
			c.handleQPACKStream(str, true)
		}
		return
	case streamTypeQPACKDecoderStream:
		if isFirst := c.rcvdQPACKDecoderStr.CompareAndSwap(false, true); !isFirst {
			c.CloseWithError(quic.ApplicationErrorCode(ErrCodeStreamCreationError), "duplicate QPACK decoder stream")
			return
		}
		if c.qpackDecoder != nil || c.clientQPACK != nil {
			c.handleQPACKStream(str, false)
		}
		return
	case streamTypePushStream:
		if isServer {
			// only the server can push
			c.CloseWithError(quic.ApplicationErrorCode(ErrCodeStreamCreationError), "")
		} else {
			// we never increased the Push ID, so we don't expect any push streams
			c.CloseWithError(quic.ApplicationErrorCode(ErrCodeIDError), "")
		}
		return
	default:
		str.CancelRead(quic.StreamErrorCode(ErrCodeStreamCreationError))
		return
	}
	// Only a single control stream is allowed.
	if isFirstControlStr := c.rcvdControlStr.CompareAndSwap(false, true); !isFirstControlStr {
		c.CloseWithError(quic.ApplicationErrorCode(ErrCodeStreamCreationError), "duplicate control stream")
		return
	}
	if bufferedSettings {
		c.handleControlStreamMode(str, true)
		return
	}

	c.handleControlStream(str)
}

func (c *rawConn) handleControlStream(str *quic.ReceiveStream) {
	c.handleControlStreamMode(str, false)
}

func (c *rawConn) handleControlStreamMode(str *quic.ReceiveStream, bufferedSettings bool) {
	fp := &frameParser{closeConn: c.CloseWithError, r: str, streamID: str.StreamID()}
	if c.strictControlSettings {
		fp.onUnknownFrame = func(ControlFrame) error { return errControlFrameBeforeSettings }
	}
	f, err := fp.ParseNext(c.qlogger)
	if err != nil {
		if errors.Is(err, errReservedPeerSetting) {
			c.CloseWithError(quic.ApplicationErrorCode(ErrCodeSettingsError), err.Error())
			return
		}

		if errors.Is(err, errPriorityUpdateForPush) || errors.Is(err, errControlFrameBeforeSettings) {
			c.CloseWithError(quic.ApplicationErrorCode(ErrCodeMissingSettings), "")
			return
		}
		_, isStreamError := errors.AsType[*quic.StreamError](err)
		if err == io.EOF || isStreamError {
			c.CloseWithError(quic.ApplicationErrorCode(ErrCodeClosedCriticalStream), "")
			return
		}
		c.CloseWithError(quic.ApplicationErrorCode(ErrCodeFrameError), "")
		return
	}
	sf, ok := f.(*settingsFrame)
	if !ok {
		c.CloseWithError(quic.ApplicationErrorCode(ErrCodeMissingSettings), "")
		return
	}
	if code, err := c.applyWirePeerSettings(sf); err != nil {
		c.CloseWithError(quic.ApplicationErrorCode(code), err.Error())
		return
	}
	if c.settings.EnableDatagrams {
		c.AddQlogProducer()
		go func() {
			defer c.RemoveQlogProducer()
			err := c.receiveDatagrams()
			if c.logger != nil {
				c.logger.Debug("receiving datagrams failed", "error", err)
			}
		}()
	}

	if bufferedSettings {
		c.AddQlogProducer()
		go func() {
			defer c.RemoveQlogProducer()
			c.controlStrHandler(str, fp)
		}()
		return
	}

	c.controlStrHandler(str, fp)
}

func (c *rawConn) sendDatagram(streamID quic.StreamID, b []byte) error {
	// TODO: this creates a lot of garbage and an additional copy
	data := make([]byte, 0, len(b)+8)
	quarterStreamID := uint64(streamID / 4)
	data = quicvarint.Append(data, uint64(streamID/4))
	data = append(data, b...)
	if c.qlogger != nil {
		c.qlogger.RecordEvent(qlog.DatagramCreated{
			QuarterStreamID: quarterStreamID,
			Raw: qlog.RawInfo{
				Length:        len(data),
				PayloadLength: len(b),
			},
		})
	}
	return c.conn.SendDatagram(data)
}

func (c *rawConn) receiveDatagrams() error {
	for {
		b, err := c.conn.ReceiveDatagram(c.context())
		if err != nil {
			return err
		}
		quarterStreamID, n, err := quicvarint.Parse(b)
		if err != nil {
			c.CloseWithError(quic.ApplicationErrorCode(ErrCodeDatagramError), "")
			return fmt.Errorf("could not read quarter stream id: %w", err)
		}
		if c.qlogger != nil {
			c.qlogger.RecordEvent(qlog.DatagramParsed{
				QuarterStreamID: quarterStreamID,
				Raw: qlog.RawInfo{
					Length:        len(b),
					PayloadLength: len(b) - n,
				},
			})
		}
		if quarterStreamID > maxQuarterStreamID {
			c.CloseWithError(quic.ApplicationErrorCode(ErrCodeDatagramError), "")
			return fmt.Errorf("invalid quarter stream id: %w", err)
		}
		streamID := quic.StreamID(4 * quarterStreamID)
		c.streamMx.Lock()
		dg, ok := c.streams[streamID]
		c.streamMx.Unlock()
		if !ok {
			continue
		}
		dg.enqueueDatagram(b[n:])
	}
}

// ReceivedSettings returns a channel that is closed once the peer's SETTINGS frame was received.
// Settings can be optained from the Settings method after the channel was closed.
func (c *rawConn) ReceivedSettings() <-chan struct{} { return c.receivedSettings }

// Settings returns the settings received on this connection.
// It is only valid to call this function after the channel returned by ReceivedSettings was closed.
func (c *rawConn) Settings() *Settings { return c.settings }

// AddQlogProducer registers a goroutine or call that may produce qlog events.
// It returns false once the qlogger has been closed.
func (c *rawConn) AddQlogProducer() bool {
	c.qloggerMx.Lock()
	defer c.qloggerMx.Unlock()

	if c.qloggerProducers == 0 {
		return false
	}
	c.qloggerProducers++
	return true
}

// RemoveQlogProducer closes the qlogger when the last producer is removed.
func (c *rawConn) RemoveQlogProducer() {
	c.qloggerMx.Lock()
	defer c.qloggerMx.Unlock()

	if c.qloggerProducers == 0 {
		return
	}
	c.qloggerProducers--
	if c.qloggerProducers == 0 {
		// Closing might block until the trace was written. No producer can be added anymore.
		go c.qlogger.Close()
	}
}
