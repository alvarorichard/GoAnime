// Zopfli optimal parsing: DP core, start-position queue, and helpers.
//
// The Zopfli algorithm finds the globally optimal command sequence for a
// metablock by running a dynamic-programming forward pass over all positions.
// At each position it evaluates all candidate (insert-length, copy-length,
// distance) triples and tracks the minimum-cost path.

package encoder

import (
	"math/bits"

	"github.com/molecule-man/go-brrr/internal/core"
)

// Zopfli quality parameters.
const (
	// zopfliIterateAborted is returned when the match feed was aborted mid-pass.
	zopfliIterateAborted  = ^uint(0)
	zopfliIterateDiverged = ^uint(0) - 1

	// maxZopfliLenQ10 is the maximum copy length for which Q10 evaluates
	// all individual lengths (beyond this, only the maximum match length
	// is tried). Shorter limit = faster but slightly worse compression.
	maxZopfliLenQ10 = 150

	// maxZopfliLenQ11 is the same threshold for Q11.
	maxZopfliLenQ11 = 325

	maxZopfliCandidatesQ11 = 5

	// longCopyQuickStep: when a copy this long is found, skip detailed
	// evaluation of the copied positions (they are unlikely to start
	// new commands).
	longCopyQuickStep = 16384
)

var shortCodeLaneBits = func() (t [2][128]uint16) {
	codes := [2][7]uint16{{9, 7, 5, 0, 4, 6, 8}, {15, 13, 11, 1, 10, 12, 14}}
	for w := range t {
		for m := range t[w] {
			for k, j := range codes[w] {
				t[w][m] |= uint16(m>>k&1) << j
			}
		}
	}
	return t
}()

// posData holds a candidate starting position and its associated state,
// used by startPosQueue to track the best insert-length candidates.
type posData struct {
	pos           uint
	distanceCache [4]int
	costdiff      float32
	cost          float32
}

// startPosQueue maintains the 8 best starting positions ordered by cost
// difference vs. literal-only cost. The DP evaluates candidates from this
// queue at each position to limit the O(n²) insert-length search space.
type startPosQueue struct {
	q   [8]posData
	idx uint
}

type dcHit struct {
	backward, bestLen, length, j uint
}

type dcScan struct {
	distanceCache [4]int
	lo, hi        int
}

type dcScratch struct {
	scans [maxZopfliCandidatesQ11]dcScan
	hits  [maxZopfliCandidatesQ11 * core.NumDistanceShortCodes]dcHit
}

// size returns the number of entries in the queue (at most 8).
func (q *startPosQueue) size() uint {
	return min(q.idx, 8)
}

// push inserts a new entry and restores sorted order by costdiff.
// The queue uses a rotating index so no entries need shifting.
func (q *startPosQueue) push(p *posData) {
	offset := ^q.idx & 7
	q.idx++
	length := q.size()
	// Find the insertion point by scanning for the first element with
	// costdiff >= p.costdiff (cheap float32 comparisons only).
	insertAt := uint(0)
	for insertAt < length-1 {
		next := (offset + insertAt + 1) & 7
		if q.q[next].costdiff >= p.costdiff {
			break
		}
		insertAt++
	}
	// Shift elements [0, insertAt) down by one to make room.
	dst := offset & 7
	for i := uint(0); i < insertAt; i++ {
		src := (offset + i + 1) & 7
		q.q[dst] = q.q[src]
		dst = src
	}
	q.q[(offset+insertAt)&7] = *p
}

// at returns a pointer to the k-th element (0 = best/lowest costdiff).
func (q *startPosQueue) at(k uint) *posData {
	return &q.q[(k-q.idx)&7]
}

// maxZopfliLen returns the maximum Zopfli copy length for the given quality.
func maxZopfliLen(quality int) uint {
	if quality <= 10 {
		return maxZopfliLenQ10
	}
	return maxZopfliLenQ11
}

// maxZopfliCandidates returns the number of start-position queue candidates
// to evaluate per position.
func maxZopfliCandidates(quality int) uint {
	if quality <= 10 {
		return 1
	}
	return maxZopfliCandidatesQ11
}

// computeDistanceShortcut determines which earlier node provides the
// distance cache for this node's position. If the current node introduced
// a new distance (not from the static dictionary and not code 0), its own
// position is the shortcut. Otherwise, it inherits the shortcut from the
// node that started the current command.
func computeDistanceShortcut(nodes []zopfliNode, blockStart, pos, maxBackwardLimit, gap uint) uint32 {
	cLen := uint(nodes[pos].copyLength())
	iLen := uint(nodes[pos].dcodeInsertLength & 0x7FFFFFF)
	dist := uint(nodes[pos].copyDistance())
	if pos == 0 {
		return 0
	}
	if dist+cLen <= blockStart+pos+gap &&
		dist <= maxBackwardLimit+gap &&
		nodes[pos].distanceCode() > 0 {
		return uint32(pos)
	}
	return nodes[pos-cLen-iLen].u // shortcut from the previous command's start
}

// computeDistanceCache walks the shortcut chain to fill distCache[0..3]
// with the four most recent distances at the given position.
func computeDistanceCache(nodes []zopfliNode, pos uint, startingDistCache, distCache []int) {
	idx := 0
	p := uint(nodes[pos].u) // shortcut
	for idx < 4 && p > 0 {
		n := nodes[p]
		distCache[idx] = int(n.distance)
		idx++
		p = uint(nodes[p-uint(n.length&0x1FFFFFF)-uint(n.dcodeInsertLength&0x7FFFFFF)].u)
	}
	sdcIdx := 0
	for ; idx < 4; idx++ {
		distCache[idx] = startingDistCache[sdcIdx]
		sdcIdx++
	}
}

// evaluateNode computes the shortcut for a processed node and pushes it
// to the queue if its cost beats the literal-only cost to that position.
func evaluateNode(nodes []zopfliNode, pos, blockStart, maxBackwardLimit, gap uint, startingDistCache []int, model *zopfliCostModel, queue *startPosQueue) {
	// Save cost before ComputeDistanceCache overwrites the u field.
	nodeCost := nodes[pos].cost()
	nodes[pos].u = computeDistanceShortcut(nodes, blockStart, pos, maxBackwardLimit, gap)
	if nodeCost <= model.getLiteralCosts(0, pos) {
		var pd posData
		pd.pos = pos
		pd.cost = nodeCost
		pd.costdiff = nodeCost - model.getLiteralCosts(0, pos)
		computeDistanceCache(nodes, pos, startingDistCache, pd.distanceCache[:])
		queue.push(&pd)
	}
}

// computeMinimumCopyLength finds the shortest copy length that could
// improve on already-known costs at future positions. This prunes the
// inner DP loop by skipping lengths that cannot possibly be better.
//
// The copy length code uses a staircase of extra bits: every time the
// length crosses a bucket boundary, one extra bit is needed, so the
// minimum achievable cost increases by 1.
func computeMinimumCopyLength(nodes []zopfliNode, pos, numBytes uint, startCost float32) uint {
	minCost := startCost
	length := uint(2)
	nextLenBucket := uint(4)
	nextLenOffset := uint(10)
	for pos+length <= numBytes && nodes[pos+length].cost() <= minCost {
		length++
		if length == nextLenOffset {
			minCost += 1.0
			nextLenOffset += nextLenBucket
			nextLenBucket *= 2
		}
	}
	return length
}

// updateNodes is the heart of the Zopfli DP. For each starting position
// in the queue (up to maxZopfliCandidates), it evaluates:
//  1. Distance-cache matches (16 short codes)
//  2. Hash-table matches from findAllMatches
//  3. Updates nodes with better costs
//
// Returns the longest copy length found (used for skip-ahead).
func updateNodes(nodes []zopfliNode, ringbuffer []byte, startingDistCache []int, matches []backwardMatch, model *zopfliCostModel, queue *startPosQueue, numBytes, blockStart, pos, ringBufferMask, maxBackwardLimit, gap uint, compound *compoundDictionary, numMatches uint, quality int, sc *dcScratch) uint {
	curIx := blockStart + pos
	curIxMasked := curIx & ringBufferMask
	maxDistance := min(curIx, maxBackwardLimit)
	maxDistanceGap := maxDistance + gap
	hasCompound := compound != nil && compound.numChunks > 0
	maxLen := numBytes - pos
	maxZopfli := maxZopfliLen(quality)
	maxIters := maxZopfliCandidates(quality)

	// BCE hints: let the compiler prove that mask-derived and
	// pos+len-derived indices are always in bounds.
	_ = ringbuffer[ringBufferMask]
	_ = nodes[numBytes]

	evaluateNode(nodes, pos, blockStart, maxBackwardLimit, gap, startingDistCache, model, queue)

	// Compute minLen from the best queue entry.
	var minLen uint
	{
		pd := queue.at(0)
		minCost := pd.cost + model.getMinCostCmd() +
			model.getLiteralCosts(pd.pos, pos)
		minLen = computeMinimumCopyLength(nodes, pos, numBytes, minCost)
	}

	result := uint(0)

	nodesAtPos := nodes[pos:]
	numScans, numHits := 0, 0
	var back [core.NumDistanceShortCodes]uint

	for k := uint(0); k < maxIters && k < queue.size(); k++ {
		pd := queue.at(k)
		insLen := pos - pd.pos
		insCode := getInsertLenCode(insLen)
		startCostdiff := pd.costdiff
		baseCost := startCostdiff + float32(insertExtra[insCode]) +
			model.getLiteralCosts(0, pos)
		cmdCodes := &cmdCodeLUT[0][insCode]

		dc := &pd.distanceCache
		lo, hi := numHits, numHits
		scanned := false
		for i := range numScans {
			if sc.scans[i].distanceCache == *dc {
				lo, hi = sc.scans[i].lo, sc.scans[i].hi
				scanned = true
				break
			}
		}
		bestLen := minLen - 1
		if !scanned && curIxMasked+bestLen <= ringBufferMask {
			d0, d1 := dc[0], dc[1]
			back[0], back[1], back[2], back[3] = uint(d0), uint(d1), uint(dc[2]), uint(dc[3])
			back[4], back[5], back[6], back[7], back[8], back[9] = uint(d0-1), uint(d0+1), uint(d0-2), uint(d0+2), uint(d0-3), uint(d0+3)
			back[10], back[11], back[12], back[13], back[14], back[15] = uint(d1-1), uint(d1+1), uint(d1-2), uint(d1+2), uint(d1-3), uint(d1+3)
			continuation := ringbuffer[curIxMasked+bestLen]
			cand := shortCodeCandidates(ringbuffer, curIxMasked, bestLen, ringBufferMask, maxDistance, dc, continuation)
			for ; cand != 0 && bestLen < maxLen; cand &= cand - 1 {
				j := uint(bits.TrailingZeros16(cand))
				backward := back[j]
				if backward == 0 || backward > maxDistanceGap {
					continue
				}
				var length uint
				switch {
				case backward <= maxDistance:
					// Regular backward reference. backward <= maxDistance <= curIx,
					// so the reference lies inside the window.
					prevIxMasked := (curIxMasked - backward) & ringBufferMask
					if prevIxMasked+bestLen > ringBufferMask ||
						continuation != ringbuffer[prevIxMasked+bestLen] {
						continue
					}
					length = uint(matchLenAt(ringbuffer, prevIxMasked, curIxMasked, int(maxLen)))
				case hasCompound:
					// Compound dictionary reference.
					d := 0
					offset := maxDistance + 1 + compound.totalSize - 1
					for offset >= backward+compound.chunkOffsets[d+1] {
						d++
					}
					source := compound.chunkSource[d]
					offset = offset - compound.chunkOffsets[d] - backward
					limit := min(compound.chunkOffsets[d+1]-compound.chunkOffsets[d]-offset, maxLen)
					if bestLen >= limit || continuation != source[offset+bestLen] {
						continue
					}
					length = uint(matchLen(
						source[offset:],
						ringbuffer[curIxMasked:],
						int(limit),
					))
				default:
					// Gray area: addressable by decoder but not available here.
					continue
				}

				if length > bestLen {
					sc.hits[numHits] = dcHit{backward, bestLen, length, j}
					numHits++
					bestLen = length
					if curIxMasked+bestLen > ringBufferMask {
						break
					}
					continuation = ringbuffer[curIxMasked+bestLen]
				}
			}
			hi = numHits
			if maxIters > 1 {
				sc.scans[numScans] = dcScan{*dc, lo, hi}
				numScans++
			}
		}
		for _, hit := range sc.hits[lo:hi] {
			distCost := baseCost + model.distanceCost(hit.j)
			jCmdCodes := cmdCodes
			if hit.j == 0 {
				jCmdCodes = &cmdCodeLUT[1][insCode]
			}
			_ = nodesAtPos[hit.length]
			for l := hit.bestLen + 1; l <= hit.length; l++ {
				copyCode := getCopyLenCode(l)
				cmdCode := jCmdCodes[copyCode&31]
				cost := baseCost
				if cmdCode >= 128 {
					cost = distCost
				}
				cost = (cost + copyExtra[copyCode]) + model.commandCost(cmdCode)
				if cost < nodesAtPos[l].cost() {
					updateZopfliNode(&nodesAtPos[l], insLen, l, l, hit.backward, hit.j+1, cost)
					if l > result {
						result = l
					}
				}
			}
		}

		// At higher iterations look only for distance cache matches.
		if k >= 2 {
			continue
		}

		// Phase 2: Hash-table matches.
		matchLen := minLen
		for j := range numMatches {
			match := matches[j]
			dist := uint(match.distance)
			isDictionaryMatch := dist > maxDistanceGap
			distCode := dist + core.NumDistanceShortCodes - 1
			distSymbol, distExtra := prefixEncodeSimpleDistance(distCode)
			distNumExtra := distSymbol >> 10
			distCost := baseCost + float32(distNumExtra) +
				model.distanceCost(uint(distSymbol&0x3FF))

			maxMatchLen := match.matchLength()
			if matchLen < maxMatchLen && (isDictionaryMatch || maxMatchLen > maxZopfli) {
				matchLen = maxMatchLen
			}
			_ = nodesAtPos[maxMatchLen] // BCE: matchLen ≤ maxMatchLen
			if isDictionaryMatch {
				if matchLen <= maxMatchLen {
					lenCode := match.matchLengthCode()
					copyCode := getCopyLenCode(lenCode)
					cmdCode := cmdCodes[copyCode&31]
					cost := distCost + copyExtra[copyCode] +
						model.commandCost(cmdCode)
					if cost < nodesAtPos[matchLen].cost() {
						updateZopfliNode(&nodesAtPos[matchLen], insLen, matchLen, lenCode, dist, 0, cost)
						if matchLen > result {
							result = matchLen
						}
					}
					matchLen++
				}
				continue
			}
			for ; matchLen <= maxMatchLen; matchLen++ {
				copyCode := getCopyLenCode(matchLen)
				cmdCode := cmdCodes[copyCode&31]
				cost := distCost + copyExtra[copyCode] +
					model.commandCost(cmdCode)
				if cost < nodesAtPos[matchLen].cost() {
					updateZopfliNode(&nodesAtPos[matchLen], insLen, matchLen, matchLen, dist, 0, cost)
					if matchLen > result {
						result = matchLen
					}
				}
			}
			_ = distExtra
		}
	}
	return result
}

func shortCodeCandidates(ringbuffer []byte, curIxMasked, bestLen, ringBufferMask, maxDistance uint, dc *[4]int, continuation byte) uint16 {
	d0, d1, d2, d3 := dc[0], dc[1], dc[2], dc[3]
	md := int(maxDistance)
	if d0 < 4 || d0 > md-3 || d1 < 4 || d1 > md-3 || d2 < 1 || d2 > md || d3 < 1 || d3 > md {
		return 0xFFFF
	}
	w0 := (curIxMasked - uint(d0) - 4) & ringBufferMask
	w1 := (curIxMasked - uint(d1) - 4) & ringBufferMask
	if max(w0, w1)+bestLen+7 > ringBufferMask {
		return 0xFFFF
	}
	c := uint64(continuation) * 0x0101010101010101
	cand := shortCodeLaneBits[0][equalLanes(loadU64LE(ringbuffer, w0+bestLen)^c)] |
		shortCodeLaneBits[1][equalLanes(loadU64LE(ringbuffer, w1+bestLen)^c)]
	if p := (curIxMasked-uint(d2))&ringBufferMask + bestLen; p <= ringBufferMask && ringbuffer[p] == continuation {
		cand |= 1 << 2
	}
	if p := (curIxMasked-uint(d3))&ringBufferMask + bestLen; p <= ringBufferMask && ringbuffer[p] == continuation {
		cand |= 1 << 3
	}
	return cand
}

func equalLanes(x uint64) uint {
	z := ^((x&0x7F7F7F7F7F7F7F7F + 0x7F7F7F7F7F7F7F7F) | x) & 0x8080808080808080
	return uint((z >> 7) * 0x0102040810204080 >> 57)
}

// zopfliIterate runs the DP over pre-collected matches (Q11 path).
func zopfliIterate(nodes []zopfliNode, ringbuffer []byte, distCache []int, model *zopfliCostModel, numMatches []uint32, matches []backwardMatch, numBytes, position, ringBufferMask, gap uint, compound *compoundDictionary, quality, lgwin int, feed *matchFeed, dict *dictOwnership) uint {
	maxBackwardLimit := (uint(1) << lgwin) - core.WindowGap
	maxZopfli := maxZopfliLen(quality)
	var queue startPosQueue
	var dictScratch [h10MaxNumMatches + maxStaticDictMatchLen + 1]backwardMatch
	var sc dcScratch
	curMatchPos := uint(0)
	handoffs := uint32(0)
	dpOwnsDict := false
	nextDPPos := uint(0)

	nodes[0].length = 0
	nodes[0].setCost(0)

	for i := uint(0); i+3 < numBytes; i++ {
		if feed != nil && dict != nil {
			if i >= nextDPPos {
				dict.dpPos.Store(uint64(i))
				nextDPPos = i + feedPublishBatch
			}
			if !dpOwnsDict && i >= dictHandoffMinPos && feed.ready.Load() <= uint64(i) && !dict.request.Load() {
				dict.request.Store(true)
			}
		}
		if !feed.wait(i) {
			return zopfliIterateAborted
		}
		if dict != nil {
			for n := dict.count.Load(); handoffs < n && dict.at[handoffs] <= uint64(i); handoffs++ {
				dpOwnsDict = !dpOwnsDict
			}
		}
		cur := matches[curMatchPos:]
		n := uint(numMatches[i])
		if dpOwnsDict {
			bestLen := uint(1)
			if n > 0 {
				bestLen = cur[n-1].matchLength()
			}
			pos := position + i
			if numDict := staticDictBackwardMatches(ringbuffer, pos&ringBufferMask, bestLen, numBytes-i,
				min(pos, maxBackwardLimit)+gap, dictScratch[h10MaxNumMatches:]); numDict > 0 {
				cur = dictScratch[h10MaxNumMatches-n:]
				copy(cur, matches[curMatchPos:curMatchPos+n])
				n += uint(numDict)
			}
		}
		skip := updateNodes(nodes, ringbuffer, distCache,
			cur, model, &queue,
			numBytes, position, i, ringBufferMask, maxBackwardLimit, gap, compound, n, quality, &sc)
		if skip < longCopyQuickStep {
			skip = 0
		} else if quality < hqZopflificationQuality &&
			(numMatches[i] != 1 || matches[curMatchPos].matchLength() < skip) {
			return zopfliIterateDiverged
		}
		curMatchPos += uint(numMatches[i])
		if numMatches[i] == 1 && matches[curMatchPos-1].matchLength() > maxZopfli {
			skip = max(matches[curMatchPos-1].matchLength(), skip)
		}
		if skip > 1 {
			skip--
			for skip > 0 {
				i++
				if i+3 >= numBytes {
					break
				}
				if !feed.wait(i) {
					return zopfliIterateAborted
				}
				evaluateNode(nodes, i, position, maxBackwardLimit, gap, distCache, model, &queue)
				curMatchPos += uint(numMatches[i])
				skip--
			}
		}
	}
	return computeShortestPathFromNodes(nodes, numBytes)
}

// mergeMatches merges two sorted backward-match slices (sorted by match length)
// into dst, which must have room for len(src1)+len(src2) entries.
// Ties in length are broken by preferring the shorter distance.
func mergeMatches(dst, src1, src2 []backwardMatch) {
	i, j, k := 0, 0, 0
	for i < len(src1) && j < len(src2) {
		l1 := src1[i].matchLength()
		l2 := src2[j].matchLength()
		if l1 < l2 || (l1 == l2 && src1[i].distance < src2[j].distance) {
			dst[k] = src1[i]
			i++
		} else {
			dst[k] = src2[j]
			j++
		}
		k++
	}
	for ; i < len(src1); i++ {
		dst[k] = src1[i]
		k++
	}
	for ; j < len(src2); j++ {
		dst[k] = src2[j]
		k++
	}
}
