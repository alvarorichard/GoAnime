package http3

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/enetx/g"
	"github.com/enetx/quic-go"
	"github.com/enetx/quic-go/http3/qlog"
	"github.com/enetx/quic-go/qlogwriter"
	"github.com/enetx/quic-go/quicvarint"
)

// FrameType is the frame type of a HTTP/3 frame
type FrameType uint64

type frame any

var errPriorityUpdateForPush = errors.New("http3: PRIORITY_UPDATE frame for push")

// The maximum length of an encoded HTTP/3 frame header is 16:
// The frame has a type and length field, both QUIC varints (maximum 8 bytes in length)
const frameHeaderLen = 16

type countingByteReader struct {
	quicvarint.Reader
	NumRead int
}

func (r *countingByteReader) ReadByte() (byte, error) {
	b, err := r.Reader.ReadByte()
	if err == nil {
		r.NumRead++
	}
	return b, err
}

func (r *countingByteReader) Read(b []byte) (int, error) {
	n, err := r.Reader.Read(b)
	r.NumRead += n
	return n, err
}

func (r *countingByteReader) Reset() {
	r.NumRead = 0
}

type frameParser struct {
	onUnknownFrame func(ControlFrame) error // Installed only after client SETTINGS.
	r              io.Reader
	streamID       quic.StreamID
	closeConn      func(quic.ApplicationErrorCode, string) error
}

func (p *frameParser) ParseNext(qlogger qlogwriter.Recorder) (frame, error) {
	r := &countingByteReader{Reader: quicvarint.NewReader(p.r)}
	for {
		t, err := quicvarint.Read(r)
		if err != nil {
			return nil, err
		}
		l, err := quicvarint.Read(r)
		if err != nil {
			return nil, err
		}

		switch t {
		case 0x0: // DATA
			if qlogger != nil {
				qlogger.RecordEvent(qlog.FrameParsed{
					StreamID: p.streamID,
					Raw: qlog.RawInfo{
						Length:        int(l) + r.NumRead,
						PayloadLength: int(l),
					},
					Frame: qlog.Frame{Frame: qlog.DataFrame{}},
				})
			}
			return &dataFrame{Length: l}, nil
		case 0x1: // HEADERS
			return &headersFrame{
				Length:    l,
				headerLen: r.NumRead,
			}, nil
		case 0x4: // SETTINGS
			return parseSettingsFrame(r, l, p.streamID, qlogger)
		case 0x3: // unsupported: CANCEL_PUSH
			if qlogger != nil {
				qlogger.RecordEvent(qlog.FrameParsed{
					StreamID: p.streamID,
					Raw:      qlog.RawInfo{Length: r.NumRead + int(l), PayloadLength: int(l)},
					Frame:    qlog.Frame{Frame: qlog.CancelPushFrame{}},
				})
			}
		case 0x5: // unsupported: PUSH_PROMISE
			if qlogger != nil {
				qlogger.RecordEvent(qlog.FrameParsed{
					StreamID: p.streamID,
					Raw:      qlog.RawInfo{Length: r.NumRead + int(l), PayloadLength: int(l)},
					Frame:    qlog.Frame{Frame: qlog.PushPromiseFrame{}},
				})
			}
		case 0x7: // GOAWAY
			return parseGoAwayFrame(r, l, p.streamID, qlogger)
		case 0xf0700: // PRIORITY_UPDATE for a request stream
			return parsePriorityUpdateFrame(r, l, p.streamID, qlogger)
		case 0xf0701: // PRIORITY_UPDATE for a push stream
			return nil, errPriorityUpdateForPush
		case 0xd: // unsupported: MAX_PUSH_ID
			if qlogger != nil {
				qlogger.RecordEvent(qlog.FrameParsed{
					StreamID: p.streamID,
					Raw:      qlog.RawInfo{Length: r.NumRead + int(l), PayloadLength: int(l)},
					Frame:    qlog.Frame{Frame: qlog.MaxPushIDFrame{}},
				})
			}
		case 0x2, 0x6, 0x8, 0x9: // reserved frame types
			if qlogger != nil {
				qlogger.RecordEvent(qlog.FrameParsed{
					StreamID: p.streamID,
					Raw:      qlog.RawInfo{Length: r.NumRead + int(l), PayloadLength: int(l)},
					Frame:    qlog.Frame{Frame: qlog.ReservedFrame{Type: t}},
				})
			}
			p.closeConn(quic.ApplicationErrorCode(ErrCodeFrameUnexpected), "")
			return nil, fmt.Errorf("http3: reserved frame type: %d", t)
		default:
			// unknown frame types
			if qlogger != nil {
				qlogger.RecordEvent(qlog.FrameParsed{
					StreamID: p.streamID,
					Raw:      qlog.RawInfo{Length: r.NumRead + int(l), PayloadLength: int(l)},
					Frame:    qlog.Frame{Frame: qlog.UnknownFrame{Type: t}},
				})
			}
			if p.onUnknownFrame != nil {
				payload := &io.LimitedReader{R: r, N: int64(l)}
				if err := p.onUnknownFrame(ControlFrame{Type: t, Length: l, Payload: payload}); err != nil {
					return nil, err
				}
				if _, err := io.Copy(io.Discard, payload); err != nil {
					return nil, err
				}
				if payload.N != 0 {
					return nil, io.ErrUnexpectedEOF
				}
				r.Reset()
				continue
			}
		}

		// skip over the payload
		if _, err := io.CopyN(io.Discard, r, int64(l)); err != nil {
			return nil, err
		}
		r.Reset()
	}
}

type dataFrame struct {
	Length uint64
}

func (f *dataFrame) Append(b []byte) []byte {
	b = quicvarint.Append(b, 0x0)
	return quicvarint.Append(b, f.Length)
}

type headersFrame struct {
	Length    uint64
	headerLen int // number of bytes read for type and length field
}

func (f *headersFrame) Append(b []byte) []byte {
	b = quicvarint.Append(b, 0x1)
	return quicvarint.Append(b, f.Length)
}

const (
	// SETTINGS_MAX_FIELD_SECTION_SIZE
	settingMaxFieldSectionSize = 0x6
	// Extended CONNECT, RFC 9220
	settingExtendedConnect = 0x8
	// HTTP Datagrams, RFC 9297
	settingDatagram = 0x33
)

type settingsFrame struct {
	MaxFieldSectionSize int64 // SETTINGS_MAX_FIELD_SECTION_SIZE, -1 if not set

	Datagram                        bool                     // HTTP Datagrams, RFC 9297
	ExtendedConnect                 bool                     // Extended CONNECT, RFC 9220
	Other                           g.MapOrd[uint64, uint64] // all settings that we don't explicitly recognize
	datagramSet, extendedConnectSet bool                     // Explicit zero overrides an authenticated prelude.
}

const maxSettingsFrameSize = 8 << 10

func parseSettingsFrame(r *countingByteReader, l uint64, streamID quic.StreamID, qlogger qlogwriter.Recorder) (*settingsFrame, error) {
	if l > maxSettingsFrameSize {
		return nil, fmt.Errorf("unexpected size for SETTINGS frame: %d", l)
	}
	buf := make([]byte, l)
	if _, err := io.ReadFull(r, buf); err != nil {
		if err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}
		return nil, err
	}
	frame := &settingsFrame{MaxFieldSectionSize: -1}
	b := bytes.NewReader(buf)

	datagram := int64(-1)
	extendedConnect := int64(-1)
	for b.Len() > 0 {
		id, err := quicvarint.Read(b)
		if err != nil { // should not happen. We allocated the whole frame already.
			return nil, err
		}
		val, err := quicvarint.Read(b)
		if err != nil { // should not happen. We allocated the whole frame already.
			return nil, err
		}
		if err := validatePeerSetting(id, val); err != nil {
			return nil, err
		}

		switch id {
		case settingMaxFieldSectionSize:
			if frame.MaxFieldSectionSize >= 0 {
				return nil, fmt.Errorf("duplicate setting: %d", id)
			}
			frame.MaxFieldSectionSize = int64(val)
		case settingExtendedConnect:
			if extendedConnect >= 0 {
				return nil, fmt.Errorf("duplicate setting: %d", id)
			}
			extendedConnect = int64(val)
			frame.ExtendedConnect = val == 1
			frame.extendedConnectSet = true
		case settingDatagram:
			if datagram >= 0 {
				return nil, fmt.Errorf("duplicate setting: %d", id)
			}
			datagram = int64(val)
			frame.Datagram = val == 1
			frame.datagramSet = true
		default:
			if frame.Other.Contains(id) {
				return nil, fmt.Errorf("duplicate setting: %d", id)
			}
			if frame.Other == nil {
				frame.Other = g.NewMapOrd[uint64, uint64]()
			}
			frame.Other.Insert(id, val)
		}
	}
	if qlogger != nil {
		qlogger.RecordEvent(qlog.FrameParsed{
			StreamID: streamID,
			Raw: qlog.RawInfo{
				Length:        r.NumRead,
				PayloadLength: int(l),
			},
			Frame: qlog.Frame{Frame: qlog.SettingsFrame{
				MaxFieldSectionSize: frame.MaxFieldSectionSize,
				Datagram:            datagram,
				ExtendedConnect:     extendedConnect,
				Other:               frame.Other.Iter().Collect().Map[uint64, uint64](),
			}},
		})
	}
	return frame, nil
}

func (f *settingsFrame) Append(b []byte) []byte {
	b = quicvarint.Append(b, 0x4)
	var l int
	if f.MaxFieldSectionSize > 0 {
		l += quicvarint.Len(settingMaxFieldSectionSize) + quicvarint.Len(uint64(f.MaxFieldSectionSize))
	}
	for id, val := range f.Other.Iter() {
		l += quicvarint.Len(id) + quicvarint.Len(val)
	}
	if f.Datagram {
		l += quicvarint.Len(settingDatagram) + quicvarint.Len(1)
	}
	if f.ExtendedConnect {
		l += quicvarint.Len(settingExtendedConnect) + quicvarint.Len(1)
	}
	b = quicvarint.Append(b, uint64(l))
	if f.MaxFieldSectionSize > 0 {
		b = quicvarint.Append(b, settingMaxFieldSectionSize)
		b = quicvarint.Append(b, uint64(f.MaxFieldSectionSize))
	}
	if f.Datagram {
		b = quicvarint.Append(b, settingDatagram)
		b = quicvarint.Append(b, 1)
	}
	if f.ExtendedConnect {
		b = quicvarint.Append(b, settingExtendedConnect)
		b = quicvarint.Append(b, 1)
	}
	for id, val := range f.Other.Iter() {
		b = quicvarint.Append(b, id)
		b = quicvarint.Append(b, val)
	}
	return b
}

type goAwayFrame struct {
	StreamID quic.StreamID
}

func parseGoAwayFrame(r *countingByteReader, l uint64, streamID quic.StreamID, qlogger qlogwriter.Recorder) (*goAwayFrame, error) {
	frame := &goAwayFrame{}
	startLen := r.NumRead
	id, err := quicvarint.Read(r)
	if err != nil {
		return nil, err
	}
	if r.NumRead-startLen != int(l) {
		return nil, errors.New("GOAWAY frame: inconsistent length")
	}
	frame.StreamID = quic.StreamID(id)
	if qlogger != nil {
		qlogger.RecordEvent(qlog.FrameParsed{
			StreamID: streamID,
			Raw:      qlog.RawInfo{Length: r.NumRead, PayloadLength: int(l)},
			Frame:    qlog.Frame{Frame: qlog.GoAwayFrame{StreamID: frame.StreamID}},
		})
	}
	return frame, nil
}

func (f *goAwayFrame) Append(b []byte) []byte {
	b = quicvarint.Append(b, 0x7)
	b = quicvarint.Append(b, uint64(quicvarint.Len(uint64(f.StreamID))))
	return quicvarint.Append(b, uint64(f.StreamID))
}

// PRIORITY_UPDATE, RFC 9218
type priorityUpdateFrame struct {
	ElementID          uint64
	PriorityFieldValue string
}

const maxPriorityUpdateFrameSize = 4 << 10

func parsePriorityUpdateFrame(r *countingByteReader, l uint64, streamID quic.StreamID, qlogger qlogwriter.Recorder) (*priorityUpdateFrame, error) {
	if l > maxPriorityUpdateFrameSize {
		return nil, fmt.Errorf("unexpected size for PRIORITY_UPDATE frame: %d", l)
	}
	buf := make([]byte, l)
	if _, err := io.ReadFull(r, buf); err != nil {
		if err == io.ErrUnexpectedEOF {
			return nil, io.EOF
		}
		return nil, err
	}
	id, n, err := quicvarint.Parse(buf)
	if err != nil {
		return nil, err
	}
	frame := &priorityUpdateFrame{
		ElementID:          id,
		PriorityFieldValue: string(buf[n:]),
	}
	if qlogger != nil {
		qlogger.RecordEvent(qlog.FrameParsed{
			StreamID: streamID,
			Raw:      qlog.RawInfo{Length: r.NumRead, PayloadLength: int(l)},
			Frame: qlog.Frame{Frame: qlog.PriorityUpdateFrame{
				StreamID:           quic.StreamID(frame.ElementID),
				PriorityFieldValue: frame.PriorityFieldValue,
			}},
		})
	}
	return frame, nil
}
