package http2

import "fmt"

// RequestHeaderField is an effective, ordered request field, including pseudo-fields.
type RequestHeaderField struct {
	Name  string // Lowercase wire name.
	Value string // Value after transport defaults and normalization.
}

// RequestBodyInfo describes an upload without reading its body. Fields belong
// to this call and may be inspected by the policy.
type RequestBodyInfo struct {
	Fields        []RequestHeaderField // Effective ordered wire fields.
	ContentLength int64                // Declared length, or -1 when unknown.
}

// RequestBodyOptions controls upload reads and DATA framing. Zero fields retain
// native behavior. Sizes must be nonnegative and at most 1 MiB. ReadCycleSize
// and InitialReadSize require ReadBufferSize. HeaderBytes accounts for a virtual
// serialized prefix in a buffered source; it never reads or changes the body.
// Short reads remain short: the transport does not wait to fill these sizes.
type RequestBodyOptions struct {
	MaxDataFrameSize      int   // Additional cap, still bounded by peer credit and frame size.
	MaxReadSize           int   // Cap for each fresh source read; short reads do not consume the next read's budget.
	ReadBufferSize        int   // Buffered source boundary and maximum read size.
	ReadCycleSize         int   // Repeating source-read budget, independent of its buffer.
	InitialReadSize       int   // First source-read budget; zero uses ReadCycleSize.
	ResetCycleOnFlowBlock bool  // Restart the cycle after waiting for peer receive credit.
	HeaderBytes           int64 // Virtual prefix consumed before the body's first read.
}

func (o RequestBodyOptions) validate() error {
	for _, size := range []int{o.MaxDataFrameSize, o.MaxReadSize, o.ReadBufferSize, o.ReadCycleSize, o.InitialReadSize} {
		if size < 0 || size > 1<<20 {
			return fmt.Errorf("http2: request body sizes must be between zero and 1 MiB")
		}
	}

	if o.HeaderBytes < 0 || (o.ReadBufferSize == 0 &&
		(o.HeaderBytes != 0 || o.ReadCycleSize != 0 || o.InitialReadSize != 0)) ||
		(o.InitialReadSize != 0 && o.ReadCycleSize == 0) {
		return fmt.Errorf("http2: invalid request body read boundaries")
	}

	return nil
}

// bodyReadPlan tracks source buffer and ReadSegments-style cycle boundaries.
// It advances only by bytes actually returned by the application reader.
type bodyReadPlan struct {
	options RequestBodyOptions
	buffer  int
	cycle   int
}

func newBodyReadPlan(o RequestBodyOptions) bodyReadPlan {
	p := bodyReadPlan{options: o}
	if o.ReadBufferSize != 0 {
		p.buffer = o.ReadBufferSize - int(o.HeaderBytes%int64(o.ReadBufferSize))
	}

	if o.ReadCycleSize != 0 {
		initial := o.InitialReadSize
		if initial == 0 {
			initial = o.ReadCycleSize
		}

		if o.HeaderBytes < int64(initial) {
			p.cycle = initial - int(o.HeaderBytes)
		} else {
			p.cycle = o.ReadCycleSize - int((o.HeaderBytes-int64(initial))%int64(o.ReadCycleSize))
		}
	}

	return p
}

func (p *bodyReadPlan) next(limit int) int {
	if p.buffer != 0 {
		limit = min(limit, p.buffer)
	}

	if p.cycle != 0 {
		limit = min(limit, p.cycle)
	}

	return limit
}

func (p *bodyReadPlan) advance(n int) {
	if p.buffer != 0 {
		p.buffer -= n
		if p.buffer == 0 {
			p.buffer = p.options.ReadBufferSize
		}
	}

	if p.cycle != 0 {
		p.cycle -= n
		if p.cycle == 0 {
			p.cycle = p.options.ReadCycleSize
		}
	}
}
