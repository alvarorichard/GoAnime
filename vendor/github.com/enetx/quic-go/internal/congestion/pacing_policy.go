package congestion

import (
	"math"
	"math/bits"
	"time"

	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
)

type PacingAlgorithm uint8

const (
	PacingNative PacingAlgorithm = iota
	PacingLeakyBucket
	PacingScheduled
)

// PacingConfig selects mechanisms. Gains and thresholds are application policy.
type PacingConfig struct {
	Algorithm                    PacingAlgorithm
	BurstPackets, LumpyPackets   uint32
	LumpyWindowFraction          uint32 // thousandths
	LumpyMinBandwidth            uint64 // bytes per second
	Granularity                  time.Duration
	SlowStartGain, AvoidanceGain uint32 // thousandths
}

type pacingPolicy struct {
	config        PacingConfig
	credit        protocol.ByteCount // signed: sub-granularity sends may incur debt
	capacity      protocol.ByteCount
	updated, next monotime.Time
	burst, lumpy  uint32
	limited       bool
}

// ApplicationLimitedHandler lets a scheduler stop compensating for elapsed
// pacing time when the application, rather than the network, paused sending.
type ApplicationLimitedHandler interface{ OnApplicationLimited() }

func (c *configuredSender) OnApplicationLimited() {
	if c.pacing != nil {
		c.pacing.limited = false
	}
}

func newPacingPolicy(c PacingConfig, mtu protocol.ByteCount, now monotime.Time) *pacingPolicy {
	return &pacingPolicy{config: c, credit: protocol.ByteCount(c.BurstPackets) * mtu,
		capacity: protocol.ByteCount(c.BurstPackets) * mtu, updated: now, burst: c.BurstPackets}
}

func (p *pacingPolicy) rate(c *configuredSender) bandwidth {
	gain := p.config.AvoidanceGain
	if c.InSlowStart() {
		gain = p.config.SlowStartGain
	}
	bw := c.bandwidthEstimate()
	hi, lo := bits.Mul64(uint64(bw), uint64(gain))
	if hi >= 1000 {
		return bandwidth(math.MaxUint64)
	}
	rate, _ := bits.Div64(hi, lo, 1000)
	return bandwidth(rate)
}

func transferTime(bytes protocol.ByteCount, rate bandwidth) time.Duration {
	if rate < bytesPerSecond {
		return time.Duration(math.MaxInt64)
	}
	ns := float64(bytes) * float64(time.Second) / float64(rate/bytesPerSecond)
	return time.Duration(min(ns, float64(math.MaxInt64-1024)))
}

func saturatingMultiply(a, b uint64) uint64 {
	hi, lo := bits.Mul64(a, b)
	if hi != 0 {
		return math.MaxUint64
	}
	return lo
}

func (p *pacingPolicy) leakyFactor(c *configuredSender) uint64 {
	gain := p.config.AvoidanceGain
	if c.InSlowStart() {
		gain = p.config.SlowStartGain
	}
	hi, lo := bits.Mul64(uint64(c.congestionWindow), uint64(gain))
	if hi >= 1000 {
		return math.MaxUint64
	}
	factor, _ := bits.Div64(hi, lo, 1000)
	return max(factor, 1)
}

func (p *pacingPolicy) deadline(c *configuredSender, inFlight protocol.ByteCount, now monotime.Time) monotime.Time {
	if p.config.Algorithm == PacingLeakyBucket {
		if c.rttStats.SmoothedRTT() == 0 {
			return 0
		}
		if p.credit >= c.maxDatagramSize {
			return 0
		}
		ns := saturatingMultiply(uint64(c.maxDatagramSize-p.credit), uint64(c.rttStats.SmoothedRTT())) / p.leakyFactor(c)
		d := time.Duration(min(ns, uint64(math.MaxInt64)))
		if d < p.config.Granularity {
			return 0
		}
		return p.updated.Add(d)
	}
	if p.burst > 0 || p.lumpy > 0 || inFlight == 0 {
		return 0
	}
	if p.next <= now.Add(p.config.Granularity) {
		return 0
	}
	return p.next
}

func (p *pacingPolicy) sent(c *configuredSender, now monotime.Time, prior, size protocol.ByteCount, eliciting bool) {
	if p.config.Algorithm == PacingLeakyBucket {
		if c.rttStats.SmoothedRTT() == 0 {
			p.credit = min(p.capacity, p.credit+p.capacity-size)
			p.updated = now
			return
		}
		elapsed := max(now.Sub(p.updated), 0)
		added := saturatingMultiply(uint64(elapsed), p.leakyFactor(c)) / uint64(c.rttStats.SmoothedRTT())
		p.credit = min(p.capacity, p.credit+protocol.ByteCount(min(added, uint64(protocol.MaxByteCount)))-size)
		p.updated = now
		return
	}
	if !eliciting {
		return
	}
	if prior == 0 && !c.InRecovery() {
		p.burst = min(p.config.BurstPackets, uint32(c.congestionWindow/c.segmentSize()))
	}
	if p.burst > 0 {
		p.burst--
		p.next = 0
		p.limited = false
		return
	}
	if !p.limited || p.lumpy == 0 {
		p.lumpy = max(1, min(p.config.LumpyPackets, uint32(float64(c.congestionWindow)*float64(p.config.LumpyWindowFraction)/1000/float64(c.segmentSize()))))
		if c.bandwidthEstimate()/bytesPerSecond < bandwidth(p.config.LumpyMinBandwidth) || prior+size >= c.congestionWindow {
			p.lumpy = 1
		}
	}
	p.lumpy--
	d := transferTime(size, p.rate(c))
	if p.limited {
		p.next = p.next.Add(d)
	} else {
		p.next = max(p.next.Add(d), now.Add(d))
	}
	p.limited = c.CanSend(prior + size)
}
