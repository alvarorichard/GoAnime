package congestion

import (
	"math"
	"time"

	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/qlog"
)

// rfcCubic keeps RFC 9438's curve and Reno-friendly estimate in bytes, with
// whole-segment growth and retained fractional ACK credit. This variant follows:
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/third_party/rust/neqo-transport/src/cc/cubic.rs
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/third_party/rust/neqo-transport/src/cc/classic_cc.rs
type rfcCubic struct {
	epoch                    monotime.Time
	wMax, k, wEst, renoAcked float64
	acked                    protocol.ByteCount
}

func (c *rfcCubic) afterAck(cwnd, acked, mss protocol.ByteCount, rtt time.Duration, now monotime.Time) protocol.ByteCount {
	w, segment := float64(cwnd), float64(mss)
	c.renoAcked += float64(acked)
	if c.epoch.IsZero() {
		c.epoch, c.wEst, c.renoAcked = now, w, float64(acked)
		c.k = 0
		if c.wMax > w {
			c.k = math.Cbrt((c.wMax - w) / segment / 0.4)
		} else {
			c.wMax = w
		}
	}
	t := max(now.Sub(c.epoch), 0).Seconds() + rtt.Seconds() - c.k
	target := min(max(math.FMA(0.4*t*t*t, segment, c.wMax), w), 1.5*w)
	const alpha = 3.0 * (1.0 - 0.7) / (1.0 + 0.7)
	if n := math.Floor(alpha * c.renoAcked / w); n > 0 {
		c.wEst = math.FMA(n, segment, c.wEst)
		c.renoAcked -= n * w / alpha
	}
	target = max(target, c.wEst)
	needed := protocol.ByteCount(min(segment*w/max(target-w, 1), float64(protocol.MaxByteCount-1024)))
	needed = max(needed, 1)
	// Credit accumulated under a slower curve earns at most one increment
	// when the curve accelerates; the newly ACKed bytes are applied in full.
	credit := min(c.acked, needed) + acked
	c.acked = credit % needed
	return cwnd + min((credit/needed)*mss, protocol.MaxByteCount-cwnd)
}

func (c *rfcCubic) afterLoss(cwnd, mss protocol.ByteCount, ecn bool) protocol.ByteCount {
	if float64(cwnd+mss) < c.wMax {
		c.wMax = float64(cwnd) * 0.85
	} else {
		c.wMax = float64(cwnd)
	}
	c.epoch = 0
	factor := protocol.ByteCount(70)
	if ecn {
		factor = 85 // RFC 8511, distinct from a loss-induced cutback.
	}
	c.acked = c.acked/100*factor + c.acked%100*factor/100
	return cwnd/100*factor + cwnd%100*factor/100
}

// AckBatcher lets a controller apply one growth calculation per received ACK
// frame, while the packet handler retains ownership of packet bookkeeping.
type AckBatcher interface {
	BeginAck()
	EndAck(now monotime.Time)
}

func (c *configuredSender) BeginAck() {
	c.batchingAck = c.rfcCubic != nil
	c.batchAcked = 0
	c.batchPackets = 0
	c.batchCwndLimited = false
}

func (c *configuredSender) EndAck(now monotime.Time) {
	if !c.batchingAck {
		return
	}
	c.batchingAck = false
	if c.batchAcked > 0 {
		c.increaseRFCWindow(c.batchAcked, c.batchCwndLimited, now)
		c.connStats.CongestionWindow.Store(uint64(c.congestionWindow))
		if c.InSlowStart() {
			c.hybridSlowStart.OnPacketAcked(c.largestAckedPacketNumber)
		}
	}
}

func (c *configuredSender) increaseRFCWindow(acked protocol.ByteCount, limited bool, now monotime.Time) {
	if !limited {
		c.rfcCubic.epoch = 0
		c.maybeQlogStateChange(qlog.CongestionStateApplicationLimited)
		return
	}
	if c.InSlowStart() {
		growth := acked
		if !c.classicSlowStart {
			growth = c.batchPackets * c.segmentSize()
		}
		c.congestionWindow = min(c.congestionWindow+growth, c.slowStartThreshold)
		c.maybeQlogStateChange(qlog.CongestionStateSlowStart)
	}
	if !c.InSlowStart() {
		c.maybeQlogStateChange(qlog.CongestionStateCongestionAvoidance)
		c.congestionWindow = c.rfcCubic.afterAck(c.congestionWindow, acked, c.segmentSize(), c.rttStats.MinRTT(), now)
	}
	c.congestionWindow = min(c.congestionWindow, c.maxCongestionWindow())
}

// Remember utilization when packets are sent, rather than treating a delayed
// ACK of a formerly full window as evidence that the application was idle.
func (c *configuredSender) rfcCwndLimited(inFlight protocol.ByteCount) bool {
	if inFlight >= c.congestionWindow {
		return true
	}
	if c.InSlowStart() {
		return inFlight >= c.congestionWindow/2
	}
	return inFlight+2*c.segmentSize() >= c.congestionWindow
}
