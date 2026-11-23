package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// Frequency counting for Huffman code construction.

// blockHistograms groups frequency counts for the three prefix code alphabets
// in a brotli meta-block: literals, insert-and-copy lengths, and distances.
type blockHistograms struct {
	lit  *[core.AlphabetSizeLiteral]uint32
	cmd  *[core.AlphabetSizeInsertAndCopyLength]uint32
	dist *[alphabetSizeDistance]uint32
}

// blockSplitIterator walks through the blocks of a blockSplit sequentially.
type blockSplitIterator struct {
	split     *blockSplit
	idx       int
	blockType int
	remaining int
}

func newBlockSplitIterator(split *blockSplit) blockSplitIterator {
	blockType := 0
	remaining := 0
	if len(split.lengths) > 0 {
		blockType = int(split.types[0])
		remaining = int(split.lengths[0])
	}
	return blockSplitIterator{
		split:     split,
		blockType: blockType,
		remaining: remaining,
	}
}

// tally counts symbols from a single command. It returns the position delta
// (insert length + copy length) and a distance delta (1 if a distance symbol
// was recorded, 0 otherwise). Literals are read from the ring buffer one byte
// at a time using input[pos & mask] to handle wrap-around correctly.
func (h *blockHistograms) tally(input []byte, pos, mask uint, cmd command) (posDelta, distDelta uint) {
	h.cmd[cmd.cmdPrefix]++
	insLen := uint(cmd.insertLen)
	if insLen > 0 {
		basePos := pos & mask
		// Fast path: no ring-buffer wrap. Iterate over a subslice so the
		// compiler can eliminate per-iteration bounds checks on input too.
		if basePos+insLen <= mask {
			for _, b := range input[basePos : basePos+insLen] {
				h.lit[b]++
			}
		} else {
			for j := range insLen {
				h.lit[input[(pos+j)&mask]]++
			}
		}
	}
	copyLen := uint(cmd.copyLength())
	if copyLen != 0 && !cmd.usesLastDistance() {
		h.dist[cmd.distPrefixCode()]++
		distDelta++
	}
	posDelta = uint(cmd.insertLen) + copyLen
	return
}

// next advances to the next symbol. Must be called once per symbol, not
// once per block.
func (it *blockSplitIterator) next() {
	if it.remaining == 0 {
		it.idx++
		if it.idx < len(it.split.lengths) {
			it.blockType = int(it.split.types[it.idx])
			it.remaining = int(it.split.lengths[it.idx])
		}
	}
	it.remaining--
}

// buildHistogramsWithContext rebuilds per-block-type frequency histograms
// from a completed metablock split, applying context modeling to route
// literal and distance symbols into context-specific histograms.
//
// When contextModes is non-nil, each literal is routed to:
//
//	litHistograms[(blockType << core.LiteralContextBits) + core.ContextLookup(mode, p1, p2)]
//
// When contextModes is nil, literals use only the block type as their
// histogram index (equivalent to the non-context path).
//
// Distance symbols are always routed to:
//
//	distHistograms[(blockType << core.DistanceContextBits) + cmd.distanceContext()]
func buildHistogramsWithContext(
	commands []command,
	mb *metaBlockSplit,
	ringbuffer []byte, startPos, mask uint,
	prevByte, prevByte2 byte,
	contextModes []byte,
	distAlphabetSize int,
	litHistograms, cmdHistograms, distHistograms []uint32,
) {
	pos := startPos
	litIter := newBlockSplitIterator(&mb.litSplit)
	cmdIter := newBlockSplitIterator(&mb.cmdSplit)
	distIter := newBlockSplitIterator(&mb.distSplit)

	for i := range commands {
		cmd := commands[i]

		cmdIter.next()
		cmdHistograms[uint(cmdIter.blockType)*core.AlphabetSizeInsertAndCopyLength+uint(cmd.cmdPrefix)]++

		for j := cmd.insertLen; j != 0; j-- {
			litIter.next()
			context := uint(litIter.blockType)
			if contextModes != nil {
				mode := uint(contextModes[context])
				context = (context << core.LiteralContextBits) +
					uint(core.ContextLookup(mode, prevByte, prevByte2))
			}
			literal := ringbuffer[pos&mask]
			litHistograms[context*core.AlphabetSizeLiteral+uint(literal)]++
			prevByte2 = prevByte
			prevByte = literal
			pos++
		}

		copyLen := cmd.copyLength()
		pos += uint(copyLen)
		if copyLen != 0 {
			prevByte2 = ringbuffer[(pos-2)&mask]
			prevByte = ringbuffer[(pos-1)&mask]
			if cmd.cmdPrefix >= 128 {
				distIter.next()
				context := (uint(distIter.blockType) << core.DistanceContextBits) +
					uint(cmd.distanceContext())
				distHistograms[context*uint(distAlphabetSize)+uint(cmd.distPrefixCode())]++
			}
		}
	}
}
