package http

import "github.com/enetx/http/internal/http2"

// HTTP2Fingerprint controls initial connection frames, request priorities and
// header encoding and receive-credit batching. A nil value uses upstream defaults.
type HTTP2Fingerprint = http2.Fingerprint
type HTTP2ConnectionErrorInfo = http2.ConnectionErrorInfo

// HTTP2ConnectionCloseInfo describes terminal HTTP/2 connection closure.
type HTTP2ConnectionCloseInfo = http2.ConnectionCloseInfo

// HTTP2ConnectionCloseOptions selects the optional final GOAWAY frame.
type HTTP2ConnectionCloseOptions = http2.ConnectionCloseOptions
type HTTP2ReceiveWindowUpdateInfo = http2.ReceiveWindowUpdateInfo

type HTTP2Setting = http2.Setting
type HTTP2PriorityParam = http2.PriorityParam
type HTTP2PriorityFrame = http2.PriorityFrame
type HTTP2HPACKConfig = http2.HPACKConfig
type HTTP2HPACKRepresentation = http2.HPACKRepresentation
type HTTP2HPACKHuffmanMode = http2.HPACKHuffmanMode
type HTTP2HPACKNameReferencePolicy = http2.HPACKNameReferencePolicy
type HTTP2PriorityRequest = http2.PriorityRequest

const (
	HTTP2HPACKDefault              = http2.HPACKDefault
	HTTP2HPACKWithoutIndexing      = http2.HPACKWithoutIndexing
	HTTP2HPACKNeverIndex           = http2.HPACKNeverIndex
	HTTP2HPACKIncrementalIndexing  = http2.HPACKIncrementalIndexing
	HTTP2HPACKHuffmanDefault       = http2.HPACKHuffmanDefault
	HTTP2HPACKHuffmanAlways        = http2.HPACKHuffmanAlways
	HTTP2HPACKHuffmanNever         = http2.HPACKHuffmanNever
	HTTP2HPACKNameReferenceDefault = http2.HPACKNameReferenceDefault
	HTTP2HPACKNameReferenceLast    = http2.HPACKNameReferenceLast
	HTTP2HPACKNameReferenceFirst   = http2.HPACKNameReferenceFirst
)

// HTTP2RequestHeaderField is an effective, ordered HTTP/2 request field.
type HTTP2RequestHeaderField = http2.RequestHeaderField

// HTTP2RequestBodyInfo describes an HTTP/2 upload without reading its body.
type HTTP2RequestBodyInfo = http2.RequestBodyInfo

// HTTP2RequestBodyOptions controls upload reads and DATA framing.
type HTTP2RequestBodyOptions = http2.RequestBodyOptions
