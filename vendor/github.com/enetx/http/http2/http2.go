// Package http2 exposes HTTP/2 transports and frame APIs backed by the single
// HTTP/2 engine in github.com/enetx/http. No second protocol engine is compiled.
// The public adapter is maintained for Surf's wire configuration, connection
// APIs and h2c Upgrade support.
//
//go:generate go run gen_aliases.go
package http2

import (
	"io"

	native "github.com/enetx/http/internal/http2"
)

func NewFramer(w io.Writer, r io.Reader) *Framer       { return native.NewFramer(w, r) }
func ReadFrameHeader(r io.Reader) (FrameHeader, error) { return native.ReadFrameHeader(r) }

// SetVerboseLogs configures native HTTP/2 diagnostics before starting any
// transports or servers. GODEBUG=http2debug=1 is the equivalent startup option.
func SetVerboseLogs(enabled bool) { native.VerboseLogs = enabled }
