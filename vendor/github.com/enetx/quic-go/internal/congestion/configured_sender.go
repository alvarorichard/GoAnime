package congestion

import (
	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/utils"
	"github.com/enetx/quic-go/qlog"
	"github.com/enetx/quic-go/qlogwriter"
)

// congestionUndoState saves the controller before the latest undoable cutback.
type congestionUndoState struct {
	congestionWindow   protocol.ByteCount
	slowStartThreshold protocol.ByteCount
	numAckedPackets    uint64
	rfcCubic           rfcCubic
	// Previous cutback's boundary; configuredSender holds the current boundary.
	largestSentAtLastCutback protocol.PacketNumber
	// Includes ECN events, which cannot be undone by spurious-loss reports.
	numCongestionEvents uint64
}

func (s *congestionUndoState) IsActive() bool {
	return s.numCongestionEvents > 0
}

func (s *congestionUndoState) Reset() {
	*s = congestionUndoState{}
}

type configuredSender struct {
	*renoSender
	connStats *utils.ConnectionStats
	cubic     *Cubic
	clock     Clock

	reno                 bool
	rfcCubic             *rfcCubic
	classicSlowStart     bool
	fixedSegmentSize     protocol.ByteCount
	maxWindowPackets     int64
	pacing               *pacingPolicy
	prr                  prr
	prrEnabled           bool
	recoveryBurstPackets uint32
	recoveryAllowance    protocol.ByteCount
	lastInFlight         protocol.ByteCount
	firstAppLimited      protocol.PacketNumber
	batchingAck          bool
	batchAcked           protocol.ByteCount
	batchPackets         protocol.ByteCount
	batchCwndLimited     bool

	// Retained after recovery, since spurious losses can be reported later.
	undoState congestionUndoState

	initialCongestionWindow    protocol.ByteCount
	initialMaxCongestionWindow protocol.ByteCount
}

var _ SendAlgorithm = &configuredSender{}

// NewCubicSender makes a new cubic sender
func NewCubicSender(
	clock Clock,
	rttStats *utils.RTTStats,
	connStats *utils.ConnectionStats,
	initialMaxDatagramSize protocol.ByteCount,
	reno bool,
	qlogger qlogwriter.Recorder,
	initialWindowPackets ...uint32,
) *configuredSender {
	window := protocol.ByteCount(initialCongestionWindow)
	if len(initialWindowPackets) > 0 && initialWindowPackets[0] != 0 {
		window = protocol.ByteCount(initialWindowPackets[0])
	}
	return newCubicSender(
		clock,
		rttStats,
		connStats,
		reno,
		initialMaxDatagramSize,
		window*initialMaxDatagramSize,
		protocol.MaxCongestionWindowPackets*initialMaxDatagramSize,
		qlogger,
	)
}

func newCubicSender(
	clock Clock,
	rttStats *utils.RTTStats,
	connStats *utils.ConnectionStats,
	reno bool,
	initialMaxDatagramSize,
	initialCongestionWindow,
	initialMaxCongestionWindow protocol.ByteCount,
	qlogger qlogwriter.Recorder,
) *configuredSender {
	c := &configuredSender{
		renoSender:                 newRenoSender(rttStats, initialMaxDatagramSize, initialCongestionWindow, qlogger),
		connStats:                  connStats,
		initialCongestionWindow:    initialCongestionWindow,
		initialMaxCongestionWindow: initialMaxCongestionWindow,
		cubic:                      NewCubic(clock),
		clock:                      clock,
		reno:                       reno,
	}
	c.connStats.CongestionWindow.Store(uint64(c.congestionWindow))
	return c
}

// TimeUntilSend returns when the next packet should be sent.
func (c *configuredSender) timeUntilSend(inFlight protocol.ByteCount) monotime.Time {
	if c.pacing != nil {
		c.lastInFlight = inFlight
		return c.pacing.deadline(c, inFlight, c.clock.Now())
	}
	return c.pacer.TimeUntilSend(c.pacingRate())
}

func (c *configuredSender) HasPacingBudget(now monotime.Time) bool {
	if c.pacing != nil {
		d := c.pacing.deadline(c, c.lastInFlight, now)
		return d.IsZero() || d <= now
	}
	return c.pacer.Budget(now, c.pacingRate()) >= c.maxDatagramSize
}

func (c *configuredSender) maxCongestionWindow() protocol.ByteCount {
	switch {
	case c.maxWindowPackets < 0:
		return protocol.MaxByteCount
	case c.maxWindowPackets > 0:
		return c.segmentSize() * protocol.ByteCount(c.maxWindowPackets)
	default:
		return c.segmentSize() * protocol.MaxCongestionWindowPackets
	}
}

func (c *configuredSender) minCongestionWindow() protocol.ByteCount {
	return c.segmentSize() * minCongestionWindowPackets
}

func (c *configuredSender) onPacketSent(
	sentTime monotime.Time,
	bytesInFlight protocol.ByteCount,
	packetNumber protocol.PacketNumber,
	bytes protocol.ByteCount,
	isRetransmittable bool,
) {
	if c.pacing == nil {
		c.pacer.SentPacket(sentTime, bytes, c.pacingRate())
	}
	c.lastInFlight = bytesInFlight
	if isRetransmittable {
		c.lastInFlight += bytes
	}
	if !isRetransmittable {
		if c.pacing != nil {
			c.pacing.sent(c, sentTime, bytesInFlight, bytes, false)
		}
		return
	}
	if c.prrEnabled && c.InRecovery() {
		c.prr.sent += bytes
	}
	if c.recoveryAllowance > 0 {
		c.recoveryAllowance = max(0, c.recoveryAllowance-bytes)
	}
	if c.rfcCubic != nil && c.rfcCwndLimited(bytesInFlight+bytes) {
		c.firstAppLimited = packetNumber + 1
	}
	c.largestSentPacketNumber = packetNumber
	c.hybridSlowStart.OnPacketSent(packetNumber)
	if c.pacing != nil {
		c.pacing.sent(c, sentTime, bytesInFlight, bytes, true)
	}
}

func (c *configuredSender) CanSend(bytesInFlight protocol.ByteCount) bool {
	c.lastInFlight = bytesInFlight
	if c.prrEnabled && c.InRecovery() {
		return c.prr.canSend(c.congestionWindow, bytesInFlight, c.slowStartThreshold, c.segmentSize())
	}
	return c.recoveryAllowance > 0 || bytesInFlight < c.GetCongestionWindow()
}

func (c *configuredSender) InRecovery() bool {
	return c.largestAckedPacketNumber != protocol.InvalidPacketNumber && c.largestAckedPacketNumber <= c.largestSentAtLastCutback
}

func (c *configuredSender) InSlowStart() bool {
	return c.GetCongestionWindow() < c.slowStartThreshold
}

func (c *configuredSender) MaybeExitSlowStart() {
	if !c.classicSlowStart && c.InSlowStart() &&
		c.hybridSlowStart.ShouldExitSlowStart(c.rttStats.LatestRTT(), c.rttStats.MinRTT(), c.GetCongestionWindow()/c.segmentSize()) {
		// exit slow start
		c.slowStartThreshold = c.congestionWindow
		c.maybeQlogStateChange(qlog.CongestionStateCongestionAvoidance)
	}
}

func (c *configuredSender) onPacketAcked(
	ackedPacketNumber protocol.PacketNumber,
	ackedBytes protocol.ByteCount,
	priorInFlight protocol.ByteCount,
	eventTime monotime.Time,
) {
	c.largestAckedPacketNumber = max(ackedPacketNumber, c.largestAckedPacketNumber)
	if c.rfcCubic != nil && ackedPacketNumber <= c.largestSentAtLastCutback {
		return
	}
	if c.InRecovery() {
		if c.prrEnabled {
			c.prr.acked(ackedBytes)
		}
		return
	}
	if c.rfcCubic != nil {
		limited := ackedPacketNumber < c.firstAppLimited
		if c.batchingAck {
			c.batchAcked += ackedBytes
			c.batchPackets++
			c.batchCwndLimited = c.batchCwndLimited || limited
			return
		}
		c.batchPackets = 1
		c.increaseRFCWindow(ackedBytes, limited, eventTime)
	} else {
		c.maybeIncreaseCwnd(ackedPacketNumber, ackedBytes, priorInFlight, eventTime)
	}
	c.connStats.CongestionWindow.Store(uint64(c.congestionWindow))
	if c.InSlowStart() {
		c.hybridSlowStart.OnPacketAcked(ackedPacketNumber)
	}
}

func (c *configuredSender) onCongestionEvent(packetNumber protocol.PacketNumber, lostBytes, priorInFlight protocol.ByteCount) {
	if lostBytes > 0 && c.pacing != nil && c.pacing.config.Algorithm == PacingScheduled {
		c.pacing.burst = 0
	}

	// TCP NewReno (RFC6582) says that once a loss occurs, any losses in packets
	// already sent should be treated as a single loss event, since it's expected.
	if packetNumber <= c.largestSentAtLastCutback {
		// Older events are already accounted for in the saved congestion window.
		if c.undoState.IsActive() && packetNumber > c.undoState.largestSentAtLastCutback {
			c.undoState.numCongestionEvents++
		}
		return
	}
	if c.reno || c.rfcCubic != nil {
		c.undoState = congestionUndoState{
			congestionWindow:         c.congestionWindow,
			slowStartThreshold:       c.slowStartThreshold,
			numAckedPackets:          c.numAckedPackets,
			largestSentAtLastCutback: c.largestSentAtLastCutback,
			numCongestionEvents:      1,
		}
		if c.rfcCubic != nil {
			c.undoState.rfcCubic = *c.rfcCubic
		}
	}

	if c.prrEnabled {
		c.prr.lost(priorInFlight)
	}
	c.recoveryAllowance = protocol.ByteCount(c.recoveryBurstPackets) * c.maxDatagramSize
	c.maybeQlogStateChange(qlog.CongestionStateRecovery)

	if c.reno {
		c.congestionWindow = protocol.ByteCount(float64(c.congestionWindow) * renoBeta)
	} else if c.rfcCubic != nil {
		c.congestionWindow = c.rfcCubic.afterLoss(c.congestionWindow, c.segmentSize(), lostBytes == 0)
	} else {
		c.congestionWindow = c.cubic.CongestionWindowAfterPacketLoss(c.congestionWindow)
	}
	if minCwnd := c.minCongestionWindow(); c.congestionWindow < minCwnd {
		c.congestionWindow = minCwnd
	}
	c.connStats.CongestionWindow.Store(uint64(c.congestionWindow))
	c.slowStartThreshold = c.congestionWindow
	c.largestSentAtLastCutback = c.largestSentPacketNumber
	// reset packet count from congestion avoidance mode. We start
	// counting again when we're out of recovery.
	c.numAckedPackets = 0
}

func (c *configuredSender) OnSpuriousLoss(packetNumber protocol.PacketNumber) {
	if !c.undoState.IsActive() {
		return
	}
	// only losses counted for this cutback can undo it
	if packetNumber <= c.undoState.largestSentAtLastCutback || packetNumber > c.largestSentAtLastCutback {
		return
	}
	c.undoState.numCongestionEvents--
	if c.undoState.numCongestionEvents > 0 {
		return
	}

	state := c.undoState
	c.undoState.Reset()
	// Regaining the old window usually takes multiple RTTs, making a spurious-loss report this late unlikely.
	if c.congestionWindow > state.congestionWindow {
		return
	}
	c.congestionWindow = state.congestionWindow
	c.slowStartThreshold = state.slowStartThreshold
	if c.rfcCubic != nil {
		*c.rfcCubic = state.rfcCubic
	}
	c.numAckedPackets = state.numAckedPackets
	c.largestSentAtLastCutback = state.largestSentAtLastCutback
	c.connStats.CongestionWindow.Store(uint64(c.congestionWindow))
	switch {
	case c.InRecovery():
		c.maybeQlogStateChange(qlog.CongestionStateRecovery)
	case c.InSlowStart():
		c.maybeQlogStateChange(qlog.CongestionStateSlowStart)
	default:
		c.maybeQlogStateChange(qlog.CongestionStateCongestionAvoidance)
	}
}

// Called when we receive an ack. Normal TCP tracks how many packets one ack
// represents, but quic has a separate ack for each packet.
func (c *configuredSender) maybeIncreaseCwnd(
	_ protocol.PacketNumber,
	ackedBytes protocol.ByteCount,
	priorInFlight protocol.ByteCount,
	eventTime monotime.Time,
) {
	// Do not increase the congestion window unless the sender is close to using
	// the current window.
	if !c.isCwndLimited(priorInFlight) {
		c.cubic.OnApplicationLimited()
		c.maybeQlogStateChange(qlog.CongestionStateApplicationLimited)
		return
	}
	if c.congestionWindow >= c.maxCongestionWindow() {
		return
	}
	if c.InSlowStart() {
		// TCP slow start, exponential growth, increase by one for each ACK.
		if c.classicSlowStart {
			c.congestionWindow = min(c.congestionWindow+ackedBytes, c.slowStartThreshold, c.maxCongestionWindow())
		} else {
			c.congestionWindow += c.segmentSize()
			if c.maxWindowPackets != 0 {
				c.congestionWindow = min(c.congestionWindow, c.maxCongestionWindow())
			}
		}
		c.maybeQlogStateChange(qlog.CongestionStateSlowStart)
		return
	}
	// Congestion avoidance
	c.maybeQlogStateChange(qlog.CongestionStateCongestionAvoidance)
	if c.reno {
		// Classic Reno congestion avoidance.
		c.numAckedPackets++
		if c.numAckedPackets >= uint64(c.congestionWindow/c.segmentSize()) {
			c.congestionWindow += c.segmentSize()
			if c.maxWindowPackets != 0 {
				c.congestionWindow = min(c.congestionWindow, c.maxCongestionWindow())
			}
			c.numAckedPackets = 0
		}
	} else {
		c.congestionWindow = min(
			c.maxCongestionWindow(),
			c.cubic.CongestionWindowAfterAck(ackedBytes, c.congestionWindow, c.rttStats.MinRTT(), eventTime),
		)
	}
}

func (c *configuredSender) isCwndLimited(bytesInFlight protocol.ByteCount) bool {
	congestionWindow := c.GetCongestionWindow()
	if bytesInFlight >= congestionWindow {
		return true
	}
	availableBytes := congestionWindow - bytesInFlight
	slowStartLimited := c.InSlowStart() && bytesInFlight > congestionWindow/2
	return slowStartLimited || availableBytes <= maxBurstPackets*c.segmentSize()
}

// OnRetransmissionTimeout is called on an retransmission timeout
func (c *configuredSender) OnRetransmissionTimeout(packetsRetransmitted bool) {
	c.largestSentAtLastCutback = protocol.InvalidPacketNumber
	c.undoState.Reset()
	c.prr = prr{}
	c.recoveryAllowance = 0
	if !packetsRetransmitted {
		return
	}
	c.hybridSlowStart.Restart()
	c.cubic.Reset()
	if c.rfcCubic != nil {
		*c.rfcCubic = rfcCubic{}
	}
	c.slowStartThreshold = c.congestionWindow / 2
	c.congestionWindow = c.minCongestionWindow()
	c.connStats.CongestionWindow.Store(uint64(c.congestionWindow))
}

// OnConnectionMigration is called when the connection is migrated (?)
func (c *configuredSender) OnConnectionMigration() {
	c.hybridSlowStart.Restart()
	c.largestSentPacketNumber = protocol.InvalidPacketNumber
	c.largestAckedPacketNumber = protocol.InvalidPacketNumber
	c.largestSentAtLastCutback = protocol.InvalidPacketNumber
	c.undoState.Reset()
	c.firstAppLimited = 0
	c.cubic.Reset()
	if c.rfcCubic != nil {
		*c.rfcCubic = rfcCubic{}
	}
	c.numAckedPackets = 0
	c.prr = prr{}
	c.recoveryAllowance = 0
	if c.pacing != nil {
		c.pacing = newPacingPolicy(c.pacing.config, c.maxDatagramSize, c.clock.Now())
	}
	c.congestionWindow = c.initialCongestionWindow
	c.connStats.CongestionWindow.Store(uint64(c.congestionWindow))
	c.slowStartThreshold = c.initialMaxCongestionWindow
}

func (c *configuredSender) SetMaxDatagramSize(s protocol.ByteCount) {
	cwndIsMinCwnd := c.congestionWindow == c.minCongestionWindow()
	c.maxDatagramSize = s
	if cwndIsMinCwnd {
		c.congestionWindow = c.minCongestionWindow()
		c.connStats.CongestionWindow.Store(uint64(c.congestionWindow))
	}
	if !c.reno && c.rfcCubic == nil && c.fixedSegmentSize == 0 {
		c.cubic.segmentSize = s
	}
	c.pacer.SetMaxDatagramSize(s)
}
