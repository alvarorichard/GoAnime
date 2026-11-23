package http2

import (
	"github.com/enetx/http"
	native "github.com/enetx/http/internal/http2"
)

type HPACKConfig = native.HPACKConfig
type HPACKRepresentation = native.HPACKRepresentation
type HPACKHuffmanMode = native.HPACKHuffmanMode
type HPACKNameReferencePolicy = native.HPACKNameReferencePolicy
type PriorityRequest = native.PriorityRequest
type ReceiveWindowUpdateInfo = native.ReceiveWindowUpdateInfo
type ConnectionErrorInfo = native.ConnectionErrorInfo

// ConnectionCloseInfo describes a terminal close before the socket is closed.
type ConnectionCloseInfo = native.ConnectionCloseInfo

// ConnectionCloseOptions selects the optional final GOAWAY frame.
type ConnectionCloseOptions = native.ConnectionCloseOptions

const (
	HPACKDefault              = native.HPACKDefault
	HPACKWithoutIndexing      = native.HPACKWithoutIndexing
	HPACKNeverIndex           = native.HPACKNeverIndex
	HPACKIncrementalIndexing  = native.HPACKIncrementalIndexing
	HPACKHuffmanDefault       = native.HPACKHuffmanDefault
	HPACKHuffmanAlways        = native.HPACKHuffmanAlways
	HPACKHuffmanNever         = native.HPACKHuffmanNever
	HPACKNameReferenceDefault = native.HPACKNameReferenceDefault
	HPACKNameReferenceLast    = native.HPACKNameReferenceLast
	HPACKNameReferenceFirst   = native.HPACKNameReferenceFirst
)

func (t *Transport) fingerprint() *http.HTTP2Fingerprint {
	if t.InitialPeerSettings == nil && t.HealthCheckPayload == nil && t.WriteBufferSize == 0 && t.RequestBodyPolicy == nil && len(t.Settings) == 0 && t.StreamID == 0 && t.ConnectionFlow == 0 &&
		t.PriorityParam.IsZero() && len(t.PriorityFrames) == 0 && t.HPACK == nil &&
		t.HeaderPriority == nil && !t.ChainStreamDependencies && t.StreamReceiveWindow == 0 &&
		t.ReceiveWindowUpdatePolicy == nil && t.ConnectionErrorPolicy == nil && t.ConnectionClosePolicy == nil {
		return nil
	}

	return (&http.HTTP2Fingerprint{
		InitialPeerSettings:       t.InitialPeerSettings,
		HealthCheckPayload:        t.HealthCheckPayload,
		WriteBufferSize:           t.WriteBufferSize,
		RequestBodyPolicy:         t.RequestBodyPolicy,
		Settings:                  t.Settings,
		StreamID:                  t.StreamID,
		ConnectionFlow:            t.ConnectionFlow,
		PriorityParam:             t.PriorityParam,
		PriorityFrames:            t.PriorityFrames,
		HPACK:                     t.HPACK,
		HeaderPriority:            t.HeaderPriority,
		ChainStreamDependencies:   t.ChainStreamDependencies,
		StreamReceiveWindow:       t.StreamReceiveWindow,
		ReceiveWindowUpdatePolicy: t.ReceiveWindowUpdatePolicy,
		ConnectionErrorPolicy:     t.ConnectionErrorPolicy,
		ConnectionClosePolicy:     t.ConnectionClosePolicy,
	}).Clone()
}

// RequestHeaderField is an effective, ordered HTTP/2 request field.
type RequestHeaderField = native.RequestHeaderField

// RequestBodyInfo describes an HTTP/2 upload without reading its body.
type RequestBodyInfo = native.RequestBodyInfo

// RequestBodyOptions controls upload reads and DATA framing.
type RequestBodyOptions = native.RequestBodyOptions
