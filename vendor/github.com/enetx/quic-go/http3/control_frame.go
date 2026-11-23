package http3

import (
	"errors"
	"io"
)

var errControlFrameBeforeSettings = errors.New("http3: extension before SETTINGS")

// ControlFrame exposes an otherwise unhandled peer control-stream extension.
// It is valid only while WireConfig.OnControlFrame is running.
type ControlFrame struct {
	// Type is the extension's HTTP/3 frame identifier.
	Type uint64
	// Length is the declared payload length, excluding the frame header.
	Length uint64
	// Payload is bounded to Length bytes. A truncated peer stream can still
	// return EOF early; handlers must enforce framing before publishing state.
	Payload io.Reader
}
