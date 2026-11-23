package quic

import "github.com/enetx/quic-go/internal/ackhandler"

type AckPolicy = ackhandler.AckPolicy
type AckReorderingMode = ackhandler.AckReorderingMode
type AckECNMode = ackhandler.AckECNMode

const (
	AckReorderingNative    = ackhandler.AckReorderingNative
	AckReorderingImmediate = ackhandler.AckReorderingImmediate
	AckReorderingWindow    = ackhandler.AckReorderingWindow
	AckECNNative           = ackhandler.AckECNNative
	AckECNTransitions      = ackhandler.AckECNTransitions
	AckECNCountOnly        = ackhandler.AckECNCountOnly
)
