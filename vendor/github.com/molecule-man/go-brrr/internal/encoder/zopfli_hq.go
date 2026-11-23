// Zopfli Q11 entry point: two-pass HQ backward references.
//
// Q11 uses a two-pass approach:
//  1. Match collection: iterate all positions, call h10.findAllMatches,
//     and store all matches in a flat array.
//  2. DP pass 1: run zopfliIterate with a literal-cost model.
//  3. DP pass 2: re-initialize the cost model from pass-1 commands,
//     re-run zopfliIterate for better results.
//
// The second pass produces better commands because the cost model reflects
// actual symbol distributions from the first pass.

package encoder

import (
	"sync/atomic"

	"github.com/molecule-man/go-brrr/internal/core"
)

const (
	feedPublishBatch   = 512
	feedSpinBeforePark = 128
	// Before dictHandoffMinPos, DP stalls reflect initial batches or blocks
	// too small to benefit from a static-dictionary handoff.
	dictHandoffMinPos = 16 * feedPublishBatch
	// dictReclaimLead is the collector lead required to reclaim the static-dictionary search.
	dictReclaimLead = 16 * feedPublishBatch
	maxDictHandoffs = 64
)

const (
	hqJobCollect hqJob = iota
	hqJobSplitCommands
	hqJobSplitDistances
	hqJobClusterDistances
)

// hqMatchesPerByte reserves match space for the parallel collector.
var hqMatchesPerByte uint = 8

type hqCollector struct {
	bufs             *q10Bufs
	hasher           *h10
	compound         *compoundDictionary
	start            chan hqJob
	done             chan struct{}
	cmdSplit         *blockSplit
	distSplit        *blockSplit
	cmds             []command
	splitVecBufs     splitVecBufs
	clusterIn        []uint32
	clusterOut       []uint32
	clusterSymbols   []uint32
	ringbuffer       []byte
	numBytes         uint
	position         uint
	ringBufferMask   uint
	maxBackwardLimit uint
	gap              uint
	storeEnd         uint
	shadowMatches    uint
	clusterInSize    int
	clusterAlphabet  int
	clusterOutSize   int
	quality          int
	distParams       distanceParams
	busy             bool
	lzScratch        [h10MaxNumMatches]backwardMatch
}

type hqJob uint8

// matchFeed shares collected matches with the first DP pass. Buffer growth
// aborts the feed because the DP holds the prior match buffer. A parked DP
// waits for more matches or an abort.
type matchFeed struct {
	wake    chan struct{}
	_       [64]byte
	ready   atomic.Uint64
	aborted atomic.Bool
	parked  atomic.Bool
	dict    dictOwnership
	// dictHandoff enables static-dictionary handoffs between the collector and the DP.
	dictHandoff bool
	_           [64]byte
}

// dictOwnership records static-dictionary handoffs. The collector owns the search first.
// Handoffs alternate owners. The collector records each handoff before it publishes
// that position, so the DP sees all handoffs through each published position.
type dictOwnership struct {
	dpPos   atomic.Uint64
	_       [64]byte
	request atomic.Bool
	_       [64]byte
	count   atomic.Uint32
	at      [maxDictHandoffs]uint64
}

func (d *dictOwnership) reset() {
	d.request.Store(false)
	d.dpPos.Store(0)
	d.count.Store(0)
}

func (f *matchFeed) reset() {
	f.ready.Store(0)
	f.aborted.Store(false)
	f.parked.Store(false)
	f.dict.reset()
	if f.wake == nil {
		f.wake = make(chan struct{}, 1)
	}
	select {
	case <-f.wake:
	default:
	}
}

func (f *matchFeed) publish(n uint) {
	f.ready.Store(uint64(n))
	f.unpark()
}

func (f *matchFeed) abort() {
	f.aborted.Store(true)
	f.unpark()
}

func (f *matchFeed) unpark() {
	if f.parked.Load() && f.parked.CompareAndSwap(true, false) {
		select {
		case f.wake <- struct{}{}:
		default:
		}
	}
}

func (f *matchFeed) wait(i uint) bool {
	return f == nil || f.ready.Load() > uint64(i) || f.waitSlow(i)
}

func (f *matchFeed) waitSlow(i uint) bool {
	for spin := 0; ; spin++ {
		if f.ready.Load() > uint64(i) {
			return true
		}
		if f.aborted.Load() {
			return false
		}
		if spin < feedSpinBeforePark {
			continue
		}
		// Set parked before the re-check: either this goroutine sees the
		// producer's store, or the producer sees parked and sends a wake.
		f.parked.Store(true)
		if f.ready.Load() > uint64(i) || f.aborted.Load() {
			// A wake sent after this point stays buffered; the next park
			// consumes it and re-checks.
			f.parked.Store(false)
			continue
		}
		<-f.wake
		spin = 0
	}
}

// createHqZopfliBackwardReferences is the Q11 top-level entry point.
func createHqZopfliBackwardReferences(numBytes, position uint, ringbuffer []byte, ringBufferMask uint, quality, lgwin int, gap uint, compound *compoundDictionary, distCache []int, hasher *h10, lastInsertLen *uint, commands *[]command, numLiterals *uint, bufs *q10Bufs) {
	maxBackwardLimit := (uint(1) << lgwin) - core.WindowGap
	hasCompound := compound != nil && compound.numChunks > 0
	shadowMatches := uint(0)
	if hasCompound {
		shadowMatches = h10MaxNumMatches + 128
	}
	matchesPerByte := uint(4)
	if bufs.parallel {
		matchesPerByte = hqMatchesPerByte
	}
	matchesSize := matchesPerByte*numBytes + shadowMatches
	storeEnd := position
	if numBytes >= h10MaxTreeCompLength {
		storeEnd = position + numBytes - h10MaxTreeCompLength + 1
	}
	// Reuse preallocated numMatchesArr from bufs.
	if cap(bufs.hqNumMatchesArr) < int(numBytes) {
		bufs.hqNumMatchesArr = make([]uint32, numBytes)
	} else {
		bufs.hqNumMatchesArr = bufs.hqNumMatchesArr[:numBytes]
		clear(bufs.hqNumMatchesArr)
	}
	numMatchesArr := bufs.hqNumMatchesArr

	// Reuse preallocated matches from bufs.
	if cap(bufs.hqMatches) < int(matchesSize) {
		bufs.hqMatches = make([]backwardMatch, matchesSize)
	} else {
		bufs.hqMatches = bufs.hqMatches[:matchesSize]
	}
	// The consumer sees the whole capacity so the producer's reslices are
	// invisible; only a true reallocation (which aborts the feed) matters.
	matches := bufs.hqMatches[:cap(bufs.hqMatches)]

	needed := int(numBytes + 1)
	if cap(bufs.zNodes) < needed {
		bufs.zNodes = make([]zopfliNode, needed)
	} else {
		bufs.zNodes = bufs.zNodes[:needed]
	}
	nodes := bufs.zNodes
	model := &bufs.zCostModel
	model.init(64, numBytes) // distAlphabetSize=64 for NPOSTFIX=0, NDIRECT=0

	passes := 2
	forestUsed := 0
	if quality < hqZopflificationQuality {
		passes = 1
		if numBytes >= longCopyQuickStep {
			forestUsed = int(min(uint64(len(hasher.forest)), 2*uint64(position+numBytes)))
			n := forestUsed + len(hasher.buckets)
			if cap(bufs.hqHasherSnap) < n {
				bufs.hqHasherSnap = make([]uint32, len(hasher.forest)+len(hasher.buckets))
			}
			bufs.hqHasherSnap = bufs.hqHasherSnap[:n]
			copy(bufs.hqHasherSnap, hasher.forest[:forestUsed])
			copy(bufs.hqHasherSnap[forestUsed:], hasher.buckets[:])
		}
	}

	feed := &bufs.hqFeed
	feed.reset()
	col := &bufs.hqCollector
	col.bufs = bufs
	col.hasher = hasher
	col.compound = compound
	col.ringbuffer = ringbuffer
	col.numBytes = numBytes
	col.position = position
	col.ringBufferMask = ringBufferMask
	col.maxBackwardLimit = maxBackwardLimit
	col.gap = gap
	col.storeEnd = storeEnd
	col.shadowMatches = shadowMatches
	col.quality = quality
	feed.dictHandoff = bufs.parallel && quality < hqZopflificationQuality && !hasCompound
	var dict *dictOwnership
	if feed.dictHandoff {
		dict = &feed.dict
	}
	var firstPassFeed *matchFeed
	if bufs.parallel {
		col.begin(hqJobCollect)
		firstPassFeed = feed
	} else {
		col.collect()
		matches = bufs.hqMatches
	}

	// Save original state for the two-pass loop.
	origNumLiterals := *numLiterals
	origLastInsertLen := *lastInsertLen
	origDistCache := [4]int{distCache[0], distCache[1], distCache[2], distCache[3]}
	origNumCommands := len(*commands)
	restore := func() {
		*commands = (*commands)[:origNumCommands]
		*numLiterals = origNumLiterals
		*lastInsertLen = origLastInsertLen
		distCache[0] = origDistCache[0]
		distCache[1] = origDistCache[1]
		distCache[2] = origDistCache[2]
		distCache[3] = origDistCache[3]
	}

	// Phase 2 & 3: Two DP passes. In parallel mode pass 1 overlaps match
	// collection through the feed; pass 2 needs every match and the pass-1
	// commands, so it waits.
	for pass := range passes {
		initZopfliNodes(nodes)
		var f *matchFeed
		if pass == 0 {
			model.setFromLiteralCosts(position, ringbuffer, ringBufferMask)
			f = firstPassFeed
		} else {
			col.wait()
			passCommands := (*commands)[origNumCommands:]
			model.setFromCommands(position, ringbuffer, ringBufferMask,
				passCommands, origLastInsertLen)
		}
		restore()

		result := zopfliIterate(nodes, ringbuffer, distCache, model, numMatchesArr, matches,
			numBytes, position, ringBufferMask, gap, compound, quality, lgwin, f, dict)
		if result == zopfliIterateAborted {
			col.wait()
			matches = bufs.hqMatches
			initZopfliNodes(nodes)
			restore()
			result = zopfliIterate(nodes, ringbuffer, distCache, model, numMatchesArr, matches,
				numBytes, position, ringBufferMask, gap, compound, quality, lgwin, nil, dict)
		}
		if result == zopfliIterateDiverged {
			col.wait()
			hasher.skipDict = false
			copy(hasher.forest[:forestUsed], bufs.hqHasherSnap)
			copy(hasher.buckets[:], bufs.hqHasherSnap[forestUsed:])
			restore()
			createZopfliBackwardReferences(numBytes, position, ringbuffer, ringBufferMask, quality, lgwin, gap, compound, distCache, hasher, lastInsertLen, commands, numLiterals, bufs)
			return
		}

		zopfliCreateCommands(nodes, numBytes, position, maxBackwardLimit, gap, distCache, lastInsertLen, commands, numLiterals)
	}
	col.wait()
	hasher.skipDict = false
}

func (c *hqCollector) begin(job hqJob) {
	if c.start == nil {
		c.start = make(chan hqJob)
		c.done = make(chan struct{})
		go c.serve(c.start, c.done)
	}
	c.busy = true
	c.start <- job
}

func (c *hqCollector) serve(start <-chan hqJob, done chan<- struct{}) {
	for job := range start {
		switch job {
		case hqJobSplitCommands:
			splitCommands(c.cmdSplit, &c.splitVecBufs, c.cmds, c.quality)
		case hqJobSplitDistances:
			c.distParams = optimizeDistanceParams(c.cmds, &c.bufs.bmTmpHist)
			splitDistances(c.distSplit, &c.splitVecBufs, c.cmds, c.quality)
		case hqJobClusterDistances:
			c.clusterOutSize, c.clusterSymbols = clusterHistograms(c.clusterIn, c.clusterInSize, c.clusterAlphabet,
				maxHistograms, c.clusterOut, &c.bufs.distClusterBufs)
		default:
			c.collect()
		}
		done <- struct{}{}
	}
}

func (c *hqCollector) wait() {
	if c.busy {
		<-c.done
		c.busy = false
	}
}

func (c *hqCollector) stop() {
	c.wait()
	if c.start != nil {
		close(c.start)
		c.start = nil
		c.done = nil
	}
	c.bufs = nil
	c.hasher = nil
	c.compound = nil
	c.ringbuffer = nil
	c.cmds = nil
	c.clusterIn = nil
	c.clusterOut = nil
	c.clusterSymbols = nil
}

func (c *hqCollector) collect() {
	bufs := c.bufs
	hasher := c.hasher
	compound := c.compound
	ringbuffer := c.ringbuffer
	numBytes := c.numBytes
	position := c.position
	ringBufferMask := c.ringBufferMask
	maxBackwardLimit := c.maxBackwardLimit
	gap := c.gap
	storeEnd := c.storeEnd
	shadowMatches := c.shadowMatches
	quality := c.quality
	hasCompound := compound != nil && compound.numChunks > 0
	maxZopfli := maxZopfliLen(quality)
	numMatchesArr := bufs.hqNumMatchesArr
	feed := &bufs.hqFeed
	matches := bufs.hqMatches
	curMatchPos := uint(0)
	nextPublish := uint(0)
	aborted := false
	dpOwnsDict := false
	handoffs := uint32(0)
	// Phase 1: Collect all matches.
	for i := uint(0); i+h10HashTypeLength-1 < numBytes; i++ {
		pos := position + i
		maxDistance := min(pos, maxBackwardLimit)
		maxLength := numBytes - i

		if feed.dictHandoff && handoffs < maxDictHandoffs &&
			(!dpOwnsDict && feed.dict.request.Load() ||
				dpOwnsDict && uint64(i) > feed.dict.dpPos.Load()+dictReclaimLead) {
			feed.dict.at[handoffs] = uint64(i)
			handoffs++
			feed.dict.count.Store(handoffs)
			feed.dict.request.Store(false)
			dpOwnsDict = !dpOwnsDict
			hasher.skipDict = dpOwnsDict
		}

		// Ensure capacity (grow-and-reuse via bufs).
		if curMatchPos+h10MaxNumMatches+shadowMatches > uint(len(matches)) {
			newSize := max(uint(len(matches))*2, curMatchPos+h10MaxNumMatches+shadowMatches)
			if cap(bufs.hqMatches) < int(newSize) {
				if !aborted {
					aborted = true
					feed.abort()
				}
				grown := make([]backwardMatch, newSize)
				copy(grown, matches[:curMatchPos])
				bufs.hqMatches = grown
			} else {
				bufs.hqMatches = bufs.hqMatches[:newSize]
			}
			matches = bufs.hqMatches
		}

		numFound := hasher.findAllMatches(
			ringbuffer, ringBufferMask, pos, maxLength, maxDistance,
			maxDistance+gap, quality, matches[curMatchPos+shadowMatches:])

		if hasCompound {
			cdMatches := compound.lookupAllMatches(
				ringbuffer, ringBufferMask, pos, 3, maxLength,
				maxDistance, maxBackwardDistance,
				matches[curMatchPos+shadowMatches-64:curMatchPos+shadowMatches],
			)
			if cdMatches > 0 {
				lzSlice := c.lzScratch[:numFound]
				copy(lzSlice, matches[curMatchPos+shadowMatches:curMatchPos+shadowMatches+numFound])
				cdSlice := matches[curMatchPos+shadowMatches-64 : curMatchPos+shadowMatches-64+cdMatches]
				mergeMatches(matches[curMatchPos:], cdSlice, lzSlice)
				numFound += cdMatches
			} else {
				copy(matches[curMatchPos:], matches[curMatchPos+shadowMatches:curMatchPos+shadowMatches+numFound])
			}
		}

		curMatchEnd := curMatchPos + numFound
		numMatchesArr[i] = uint32(numFound)

		if numFound > 0 {
			matchLen := matches[curMatchEnd-1].matchLength()
			if matchLen > maxZopfli {
				skip := matchLen - 1
				matches[curMatchPos] = matches[curMatchEnd-1]
				curMatchPos++
				numMatchesArr[i] = 1
				// Store the tail in the hasher.
				hasher.storeRange(ringbuffer, ringBufferMask,
					pos+1, min(pos+matchLen, storeEnd))
				for j := uint(1); j <= skip; j++ {
					if i+j < numBytes {
						numMatchesArr[i+j] = 0
					}
				}
				i += skip
			} else {
				curMatchPos = curMatchEnd
			}
		}
		if !aborted && i+1 >= nextPublish {
			feed.publish(i + 1)
			nextPublish = i + 1 + feedPublishBatch
		}
	}
	if !aborted {
		feed.publish(numBytes)
	}
}
