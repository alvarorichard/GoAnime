package quic

import (
	"fmt"
	"time"

	"github.com/enetx/quic-go/internal/ackhandler"
	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/wire"
	"github.com/enetx/quic-go/quicvarint"
)

// PeerAckFrequency requests the peer's legacy Application Data ACK behavior.
type PeerAckFrequency struct {
	AckElicitingThreshold uint64
	MaxAckDelay           time.Duration
	IgnoreOrder           bool
}

// PeerAckFrequencyInfo is a connection-loop snapshot, supplied only after the
// handshake and negotiation of the peer's legacy min_ack_delay parameter.
type PeerAckFrequencyInfo struct {
	CongestionWindow uint64
	MaxDatagramSize  uint64
	SmoothedRTT      time.Duration
	PeerMinAckDelay  time.Duration
	PeerMaxAckDelay  time.Duration
	Current          PeerAckFrequency // Last acknowledged request, initially the protocol defaults.
}

// PeerAckFrequencyPolicy returns a target and whether it needs to be sent.
// It must not block or reenter the connection. Invalid targets fail the local
// connection. Targets require a positive QUIC integer tolerance and a whole
// microsecond delay at least as large as the negotiated peer minimum.
// Delays must fit QUIC's max_ack_delay duration range (16383ms).
// A valid target is required even when send is false, for conservative PTO sizing.
// The engine handles sequencing, one outstanding frame, ACK/loss and PTO bounds.
// A shared policy can run concurrently on different connections.
type PeerAckFrequencyPolicy func(PeerAckFrequencyInfo) (target PeerAckFrequency, send bool)

type peerAckFrequencySender struct {
	policy             PeerAckFrequencyPolicy
	minimum, maximum   time.Duration
	current, target    PeerAckFrequency
	update             bool
	nextSequence       uint64
	outstanding        *peerAckFrequencyTransmission
	lastAckedSequence  uint64
	hasAcked           bool
	latestSentSequence uint64
	hasSent            bool
	unconfirmedDelay   time.Duration
}

type peerAckFrequencyTransmission struct {
	sender   *peerAckFrequencySender
	sequence uint64
	request  PeerAckFrequency
}

func newPeerAckFrequencySender(policy PeerAckFrequencyPolicy, minimum, maximum time.Duration) *peerAckFrequencySender {
	return &peerAckFrequencySender{
		policy: policy, minimum: minimum, maximum: maximum,
		current: PeerAckFrequency{AckElicitingThreshold: 2, MaxAckDelay: maximum},
	}
}

func (s *peerAckFrequencySender) refresh(window, packetSize protocol.ByteCount, rtt time.Duration) error {
	target, update := s.policy(PeerAckFrequencyInfo{
		CongestionWindow: uint64(window), MaxDatagramSize: uint64(packetSize), SmoothedRTT: rtt,
		PeerMinAckDelay: s.minimum, PeerMaxAckDelay: s.maximum, Current: s.current,
	})
	if target.AckElicitingThreshold == 0 || target.AckElicitingThreshold > quicvarint.Max ||
		target.MaxAckDelay < s.minimum || target.MaxAckDelay < 0 || target.MaxAckDelay > protocol.MaxMaxAckDelay ||
		target.MaxAckDelay%time.Microsecond != 0 {
		return fmt.Errorf("invalid peer ACK_FREQUENCY policy result")
	}
	s.target, s.update = target, update
	return nil
}

func (s *peerAckFrequencySender) pending() bool {
	return s != nil && s.update && s.outstanding == nil && s.nextSequence <= quicvarint.Max
}

func (s *peerAckFrequencySender) popFrame(maxLen protocol.ByteCount, version protocol.Version) ackhandler.Frame {
	if !s.pending() {
		return ackhandler.Frame{}
	}
	f := &wire.LegacyAckFrequencyFrame{
		SequenceNumber: s.nextSequence, AckElicitingThreshold: s.target.AckElicitingThreshold,
		RequestMaxAckDelay: s.target.MaxAckDelay, IgnoreOrder: s.target.IgnoreOrder,
	}
	if f.Length(version) > maxLen {
		return ackhandler.Frame{}
	}
	transmission := &peerAckFrequencyTransmission{sender: s, sequence: s.nextSequence, request: s.target}
	s.outstanding = transmission
	s.latestSentSequence, s.hasSent = s.nextSequence, true
	s.nextSequence++
	s.unconfirmedDelay = max(s.unconfirmedDelay, s.target.MaxAckDelay)
	return ackhandler.Frame{Frame: f, Handler: transmission}
}

func (s *peerAckFrequencySender) effectiveDelay() time.Duration {
	// A lost request may have reached the peer. Retain the high-water bound
	// until an ACK for the latest request proves earlier requests superseded.
	return max(s.current.MaxAckDelay, s.target.MaxAckDelay, s.unconfirmedDelay)
}

func (t *peerAckFrequencyTransmission) OnAcked(wire.Frame) {
	s := t.sender
	if !s.hasAcked || t.sequence > s.lastAckedSequence {
		s.current, s.lastAckedSequence, s.hasAcked = t.request, t.sequence, true
	}
	if s.outstanding == t {
		s.outstanding = nil
	}
	if s.current == s.target {
		s.update = false
	}
	if s.hasSent && t.sequence == s.latestSentSequence {
		s.unconfirmedDelay = 0
	}
}

func (t *peerAckFrequencyTransmission) OnLost(wire.Frame) {
	if t.sender.outstanding == t {
		t.sender.outstanding = nil
	}
}

func (c *Conn) configureLegacyAckFrequency(p *wire.TransportParameters) {
	minimum := c.config.LegacyAckFrequencyMinDelay
	if c.perspective == protocol.PerspectiveClient && c.config.ClientHandshake != nil {
		h := c.config.ClientHandshake
		if h.LegacyAckFrequencyMinDelay != 0 {
			minimum = h.LegacyAckFrequencyMinDelay
		}
		p.ParseLegacyAckFrequency = h.PeerAckFrequencyPolicy != nil
	}
	if minimum > 0 {
		p.LegacyMinAckDelay = &minimum
		c.receivedPacketHandler.SetLegacyAckFrequency(minimum)
		c.frameParser.SetLegacyAckFrequency(true)
	}
}

func (c *Conn) configurePeerAckFrequency() {
	if c.perspective != protocol.PerspectiveClient || c.config.ClientHandshake == nil ||
		c.config.ClientHandshake.PeerAckFrequencyPolicy == nil || c.peerParams.LegacyMinAckDelay == nil {
		return
	}
	c.framer.peerAckFrequency = newPeerAckFrequencySender(c.config.ClientHandshake.PeerAckFrequencyPolicy,
		*c.peerParams.LegacyMinAckDelay, c.peerParams.MaxAckDelay)
}

func (c *Conn) refreshPeerAckFrequency(now monotime.Time) error {
	if c.framer == nil {
		return nil
	}
	s := c.framer.peerAckFrequency
	if s == nil {
		return nil
	}
	window, packetSize := c.sentPacketHandler.AckFrequencyState()
	if err := s.refresh(window, packetSize, c.rttStats.SmoothedRTT()); err != nil {
		return err
	}
	if delay := s.effectiveDelay(); delay != c.rttStats.MaxAckDelay() {
		c.sentPacketHandler.SetPeerMaxAckDelay(delay, now)
	}
	return nil
}
