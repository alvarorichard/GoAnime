// Greedy block splitter for partitioning symbol streams into typed blocks.
//
// A single concrete type parameterized at construction time handles all three
// symbol categories (literal, command, distance), with a flat []uint32 buffer
// for histograms indexed by [typeIndex * alphabetSize + symbol].

package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// blockSplit records how a symbol stream is partitioned into typed blocks.
// Each block has a type (index into a histogram array) and a length.
type blockSplit struct {
	types    []byte   // block type per block (0-indexed)
	lengths  []uint32 // number of symbols in each block
	numTypes int      // count of distinct types used
}

// splitBufs holds reusable buffers for the Q4 block-splitting path.
// All slices use grow-and-reuse semantics: grown on demand, never shrunk.
// After the first metablock they typically never allocate again.
type splitBufs struct {
	litHistograms  []uint32
	ctxCombined    []uint32
	ctxLastEntropy []float64
	cmdHistograms  []uint32
	distHistograms []uint32
	litTypes       []byte
	litLengths     []uint32
	cmdTypes       []byte
	cmdLengths     []uint32
	distTypes      []byte
	distLengths    []uint32
}

// metaBlockSplit holds the complete partitioning of a metablock into
// block types for literals, commands, and distances, plus their histograms.
type metaBlockSplit struct {
	litHistograms  []uint32 // flat: [numLitTypes*numContexts][core.AlphabetSizeLiteral]
	cmdHistograms  []uint32 // flat: [numCmdTypes][core.AlphabetSizeInsertAndCopyLength]
	distHistograms []uint32 // flat: [numDistTypes][distAlphabetSize]

	// literalContextMap maps (blockType << core.LiteralContextBits | contextID)
	// to a histogram index. Non-nil only when context modeling is active
	// (quality >= 5 and decideOverLiteralContextModeling chose > 1 context).
	literalContextMap []uint32

	// distanceContextMap maps (blockType << core.DistanceContextBits | distContext)
	// to a distance histogram index. Non-nil only in the slow path (Q10+),
	// where histogram clustering produces a non-trivial mapping. The greedy
	// path uses an identity mapping (each block type maps 1:1 to its histograms).
	distanceContextMap []uint32

	litSplit  blockSplit
	cmdSplit  blockSplit
	distSplit blockSplit
}

// blockSplitter is a greedy block splitter for one block category
// (literal, command, or distance).
//
// The core algorithm is identical across all three categories — only the
// alphabet size, minimum block size, and split threshold differ.
type blockSplitter struct {
	// Output.
	split      *blockSplit
	histograms []uint32 // flat: [maxNumTypes+1][alphabetSize]

	// Configuration (set at init, immutable after).
	splitThreshold float64

	lastEntropy      [2]float64
	lastHistogramIdx [2]int

	targetBlockSize  int
	histOff          int // currHistogramIdx * alphabetSize (cached for addSymbol)
	blockSize        int
	alphabetSize     int
	minBlockSize     int
	histogramsSize   int // number of histogram slots in use
	numBlocks        int
	currHistogramIdx int // index of current histogram (0-based)
	mergeLastCount   int
}

// contextBlockSplitter is a context-aware block splitter for literals.
// It extends blockSplitter by maintaining per-context histograms:
// instead of one histogram per block type, it keeps numContexts histograms
// per block type, with split/merge decisions based on total entropy across
// all contexts.
//
// This enables context-aware block splitting at quality >= 5.
//
// The split/merge decisions sum entropy differences across all contexts
// within a block type, so a split only occurs when the total entropy
// reduction across all contexts exceeds the threshold.
type contextBlockSplitter struct {
	// Pointer-containing fields grouped first to minimize GC scan area.
	split       *blockSplit
	histograms  []uint32  // flat: [maxNumTypes * numContexts][alphabetSize]
	lastEntropy []float64 // [2 * numContexts]: per-context entropy for last two block types
	combined    []uint32  // scratch: [2 * numContexts * alphabetSize] for merge evaluation

	splitThreshold   float64
	lastHistogramIdx [2]int // histogram pool index (first context slot) for last two types
	alphabetSize     int
	numContexts      int
	maxBlockTypes    int
	minBlockSize     int
	histogramsSize   int // total histogram slots allocated (maxNumTypes * numContexts)
	targetBlockSize  int
	blockSize        int
	numBlocks        int
	currHistogramIdx int // pool index of first context slot for current candidate
	mergeLastCount   int
}

// newBlockSplitter creates a block splitter for a symbol stream of the given
// configuration. split is the output block split structure.
// The caller passes reusable buffers (histBuf, typesBuf, lengthsBuf) which are
// grown if needed. Returns the splitter and the (possibly grown) buffers.
func newBlockSplitter(
	split *blockSplit,
	alphabetSize, minBlockSize int,
	splitThreshold float64,
	numSymbols int,
	histBuf []uint32,
	typesBuf []byte,
	lengthsBuf []uint32,
) (blockSplitter, []uint32, []byte, []uint32) {
	maxNumBlocks := numSymbols/minBlockSize + 1
	// We have to allocate one more histogram than the maximum number of block
	// types for the current histogram when the meta-block is too big.
	maxNumTypes := min(maxNumBlocks, maxNumberOfBlockTypes+1)

	needed := maxNumTypes * alphabetSize
	if cap(histBuf) < needed {
		histBuf = make([]uint32, needed)
	} else {
		histBuf = histBuf[:needed]
		clear(histBuf)
	}

	if cap(typesBuf) < maxNumBlocks {
		typesBuf = make([]byte, maxNumBlocks)
	} else {
		typesBuf = typesBuf[:maxNumBlocks]
	}

	if cap(lengthsBuf) < maxNumBlocks {
		lengthsBuf = make([]uint32, maxNumBlocks)
	} else {
		lengthsBuf = lengthsBuf[:maxNumBlocks]
	}

	split.types = typesBuf
	split.lengths = lengthsBuf
	split.numTypes = 0

	bs := blockSplitter{
		alphabetSize:    alphabetSize,
		minBlockSize:    minBlockSize,
		splitThreshold:  splitThreshold,
		split:           split,
		histograms:      histBuf,
		histogramsSize:  maxNumTypes,
		targetBlockSize: minBlockSize,
	}
	return bs, histBuf, typesBuf, lengthsBuf
}

// newContextBlockSplitter creates a context-aware block splitter for literals.
//
//nolint:unparam // alphabetSize and splitThreshold are always 256/400.0 today; kept as parameters to match C reference API
func newContextBlockSplitter(
	split *blockSplit,
	alphabetSize, numContexts, minBlockSize int,
	splitThreshold float64,
	numSymbols int,
	bufs *splitBufs,
) contextBlockSplitter {
	maxNumBlocks := numSymbols/minBlockSize + 1
	maxBlockTypes := maxNumberOfBlockTypes / numContexts
	maxNumTypes := min(maxNumBlocks, maxBlockTypes+1)

	histSize := maxNumTypes * numContexts * alphabetSize
	bufs.litHistograms = growUint32Clear(bufs.litHistograms, histSize)
	histBuf := bufs.litHistograms
	bufs.ctxLastEntropy = growFloat64(bufs.ctxLastEntropy, 2*numContexts)
	bufs.ctxCombined = growUint32(bufs.ctxCombined, 2*numContexts*alphabetSize)

	if cap(split.types) < maxNumBlocks {
		split.types = make([]byte, maxNumBlocks)
	} else {
		split.types = split.types[:maxNumBlocks]
	}
	if cap(split.lengths) < maxNumBlocks {
		split.lengths = make([]uint32, maxNumBlocks)
	} else {
		split.lengths = split.lengths[:maxNumBlocks]
	}
	split.numTypes = 0

	return contextBlockSplitter{
		alphabetSize:    alphabetSize,
		numContexts:     numContexts,
		maxBlockTypes:   maxBlockTypes,
		minBlockSize:    minBlockSize,
		splitThreshold:  splitThreshold,
		split:           split,
		histograms:      histBuf,
		histogramsSize:  maxNumTypes * numContexts,
		targetBlockSize: minBlockSize,
		lastEntropy:     bufs.ctxLastEntropy,
		combined:        bufs.ctxCombined,
	}
}

// reset clears the blockSplit while preserving backing arrays.
func (bs *blockSplit) reset() {
	bs.types = bs.types[:0]
	bs.lengths = bs.lengths[:0]
	bs.numTypes = 0
}

// addSymbol adds the next symbol to the current histogram. When the current
// histogram reaches the target size, decides on merging the block.
func (bs *blockSplitter) addSymbol(symbol int) {
	bs.histograms[bs.histOff+symbol]++
	bs.blockSize++
	if bs.blockSize == bs.targetBlockSize {
		bs.finishBlock(false)
	}
}

// finishBlock makes the merge/split decision for the current block.
func (bs *blockSplitter) finishBlock(isFinal bool) {
	split := bs.split

	if bs.blockSize < bs.minBlockSize {
		bs.blockSize = bs.minBlockSize
	}

	switch {
	case bs.numBlocks == 0:
		// Create first block.
		split.lengths[0] = uint32(bs.blockSize)
		split.types[0] = 0
		bs.lastEntropy[0] = bitsEntropy(bs.histograms[0:bs.alphabetSize])
		bs.lastEntropy[1] = bs.lastEntropy[0]
		bs.numBlocks++
		split.numTypes++
		bs.currHistogramIdx++
		bs.histOff = bs.currHistogramIdx * bs.alphabetSize
		if bs.currHistogramIdx < bs.histogramsSize {
			bs.clearHistogram(bs.currHistogramIdx)
		}
		bs.blockSize = 0

	case bs.blockSize > 0:
		currStart := bs.currHistogramIdx * bs.alphabetSize
		entropy := bitsEntropy(bs.histograms[currStart : currStart+bs.alphabetSize])

		var combinedEntropy [2]float64
		var diff [2]float64

		// Compute combined entropies with last two block types.
		for j := range 2 {
			lastIdx := bs.lastHistogramIdx[j]
			lastStart := lastIdx * bs.alphabetSize
			combinedEntropy[j] = bs.combinedBitsEntropy(currStart, lastStart)
			diff[j] = combinedEntropy[j] - entropy - bs.lastEntropy[j]
		}

		switch {
		case split.numTypes < maxNumberOfBlockTypes &&
			diff[0] > bs.splitThreshold &&
			diff[1] > bs.splitThreshold:
			// Create new block type.
			split.lengths[bs.numBlocks] = uint32(bs.blockSize)
			split.types[bs.numBlocks] = byte(split.numTypes)
			bs.lastHistogramIdx[1] = bs.lastHistogramIdx[0]
			bs.lastHistogramIdx[0] = split.numTypes
			bs.lastEntropy[1] = bs.lastEntropy[0]
			bs.lastEntropy[0] = entropy
			bs.numBlocks++
			split.numTypes++
			bs.currHistogramIdx++
			bs.histOff = bs.currHistogramIdx * bs.alphabetSize
			if bs.currHistogramIdx < bs.histogramsSize {
				bs.clearHistogram(bs.currHistogramIdx)
			}
			bs.blockSize = 0
			bs.mergeLastCount = 0
			bs.targetBlockSize = bs.minBlockSize
		case diff[1] < diff[0]-20.0:
			// Combine this block with second last block.
			split.lengths[bs.numBlocks] = uint32(bs.blockSize)
			split.types[bs.numBlocks] = split.types[bs.numBlocks-2]
			bs.lastHistogramIdx[0], bs.lastHistogramIdx[1] = bs.lastHistogramIdx[1], bs.lastHistogramIdx[0]
			// Update the histogram for the merged block.
			bs.addHistograms(bs.lastHistogramIdx[0], bs.currHistogramIdx)
			bs.lastEntropy[1] = bs.lastEntropy[0]
			bs.lastEntropy[0] = combinedEntropy[1]
			bs.numBlocks++
			bs.blockSize = 0
			bs.clearHistogram(bs.currHistogramIdx)
			bs.mergeLastCount = 0
			bs.targetBlockSize = bs.minBlockSize
		default:
			// Combine this block with last block.
			split.lengths[bs.numBlocks-1] += uint32(bs.blockSize)
			// Update the histogram for the merged block.
			bs.addHistograms(bs.lastHistogramIdx[0], bs.currHistogramIdx)
			bs.lastEntropy[0] = combinedEntropy[0]
			if split.numTypes == 1 {
				bs.lastEntropy[1] = bs.lastEntropy[0]
			}
			bs.blockSize = 0
			bs.clearHistogram(bs.currHistogramIdx)
			bs.mergeLastCount++
			if bs.mergeLastCount > 1 {
				bs.targetBlockSize += bs.minBlockSize
			}
		}
	}

	if isFinal {
		bs.histogramsSize = split.numTypes
		split.types = split.types[:bs.numBlocks]
		split.lengths = split.lengths[:bs.numBlocks]
	}
}

// clearHistogram zeros the histogram at the given index.
func (bs *blockSplitter) clearHistogram(idx int) {
	start := idx * bs.alphabetSize
	end := start + bs.alphabetSize
	clear(bs.histograms[start:end])
}

// combinedBitsEntropy computes BitsEntropy of the sum of histograms at
// indices a and b (without modifying either). This avoids allocating a
// temporary combined histogram.
func (bs *blockSplitter) combinedBitsEntropy(aStart, bStart int) float64 {
	a := bs.histograms[aStart : aStart+bs.alphabetSize]
	b := bs.histograms[bStart : bStart+bs.alphabetSize]

	var sum int
	var retval float64
	for i, va := range a {
		p := int(va) + int(b[i])
		sum += p
		retval -= float64(float64(p) * fastLog2(p))
	}
	if sum != 0 {
		retval += float64(float64(sum) * fastLog2(sum))
	}
	if retval < float64(sum) {
		retval = float64(sum)
	}
	return retval
}

// addHistograms adds the histogram at srcIdx into the histogram at dstIdx.
func (bs *blockSplitter) addHistograms(dstIdx, srcIdx int) {
	dst := bs.histograms[dstIdx*bs.alphabetSize:]
	src := bs.histograms[srcIdx*bs.alphabetSize:]
	for i := range bs.alphabetSize {
		dst[i] += src[i]
	}
}

// addSymbol adds a literal to the context-specific histogram of the current
// candidate block type.
func (cs *contextBlockSplitter) addSymbol(symbol, context int) {
	idx := (cs.currHistogramIdx + context) * cs.alphabetSize
	cs.histograms[idx+symbol]++
	cs.blockSize++
	if cs.blockSize == cs.targetBlockSize {
		cs.finishBlock(false)
	}
}

// finishBlock makes the merge/split decision for the current block.
// The decision is based on the total entropy change summed across all
// contexts, ensuring that splits reflect genuine improvements in the
// overall coding efficiency.
func (cs *contextBlockSplitter) finishBlock(isFinal bool) {
	split := cs.split
	nc := cs.numContexts
	as := cs.alphabetSize

	if cs.blockSize < cs.minBlockSize {
		cs.blockSize = cs.minBlockSize
	}

	switch {
	case cs.numBlocks == 0:
		// Create first block.
		split.lengths[0] = uint32(cs.blockSize)
		split.types[0] = 0
		for i := range nc {
			start := i * as
			cs.lastEntropy[i] = bitsEntropy(cs.histograms[start : start+as])
			cs.lastEntropy[nc+i] = cs.lastEntropy[i]
		}
		cs.numBlocks++
		split.numTypes++
		cs.currHistogramIdx += nc
		if cs.currHistogramIdx < cs.histogramsSize {
			cs.clearHistograms(cs.currHistogramIdx, nc)
		}
		cs.blockSize = 0

	case cs.blockSize > 0:
		// Compute per-context entropies and combined entropies with the
		// last two block types. Sum the diffs to decide on split/merge.
		var diff [2]float64
		var entropyArr [maxStaticContexts]float64
		var combinedEntropyArr [2 * maxStaticContexts]float64
		entropy := entropyArr[:nc]
		combinedEntropy := combinedEntropyArr[:2*nc]

		for i := range nc {
			currStart := (cs.currHistogramIdx + i) * as
			entropy[i] = bitsEntropy(cs.histograms[currStart : currStart+as])

			for j := range 2 {
				jx := j*nc + i
				lastStart := (cs.lastHistogramIdx[j] + i) * as
				combStart := jx * as

				// Build combined histogram: copy current, then add last.
				copy(cs.combined[combStart:combStart+as], cs.histograms[currStart:currStart+as])
				lastHist := cs.histograms[lastStart : lastStart+as]
				for k := range as {
					cs.combined[combStart+k] += lastHist[k]
				}

				combinedEntropy[jx] = bitsEntropy(cs.combined[combStart : combStart+as])
				diff[j] += combinedEntropy[jx] - entropy[i] - cs.lastEntropy[jx]
			}
		}

		switch {
		case split.numTypes < cs.maxBlockTypes &&
			diff[0] > cs.splitThreshold &&
			diff[1] > cs.splitThreshold:
			// Create new block type.
			split.lengths[cs.numBlocks] = uint32(cs.blockSize)
			split.types[cs.numBlocks] = byte(split.numTypes)
			cs.lastHistogramIdx[1] = cs.lastHistogramIdx[0]
			cs.lastHistogramIdx[0] = split.numTypes * nc
			copy(cs.lastEntropy[nc:nc+nc], cs.lastEntropy[:nc])
			copy(cs.lastEntropy[:nc], entropy[:nc])
			cs.numBlocks++
			split.numTypes++
			cs.currHistogramIdx += nc
			if cs.currHistogramIdx < cs.histogramsSize {
				cs.clearHistograms(cs.currHistogramIdx, nc)
			}
			cs.blockSize = 0
			cs.mergeLastCount = 0
			cs.targetBlockSize = cs.minBlockSize

		case diff[1] < diff[0]-20.0:
			// Combine with second-last block.
			split.lengths[cs.numBlocks] = uint32(cs.blockSize)
			split.types[cs.numBlocks] = split.types[cs.numBlocks-2]
			cs.lastHistogramIdx[0], cs.lastHistogramIdx[1] = cs.lastHistogramIdx[1], cs.lastHistogramIdx[0]
			for i := range nc {
				// Replace last block's histograms with the combined version.
				dstStart := (cs.lastHistogramIdx[0] + i) * as
				srcStart := (nc + i) * as // combined_histo[nc + i]
				copy(cs.histograms[dstStart:dstStart+as], cs.combined[srcStart:srcStart+as])
				cs.lastEntropy[nc+i] = cs.lastEntropy[i]
				cs.lastEntropy[i] = combinedEntropy[nc+i]
				cs.clearHistogram(cs.currHistogramIdx + i)
			}
			cs.numBlocks++
			cs.blockSize = 0
			cs.mergeLastCount = 0
			cs.targetBlockSize = cs.minBlockSize

		default:
			// Combine with last block.
			split.lengths[cs.numBlocks-1] += uint32(cs.blockSize)
			for i := range nc {
				dstStart := (cs.lastHistogramIdx[0] + i) * as
				srcStart := i * as // combined_histo[i]
				copy(cs.histograms[dstStart:dstStart+as], cs.combined[srcStart:srcStart+as])
				cs.lastEntropy[i] = combinedEntropy[i]
				if split.numTypes == 1 {
					cs.lastEntropy[nc+i] = cs.lastEntropy[i]
				}
				cs.clearHistogram(cs.currHistogramIdx + i)
			}
			cs.blockSize = 0
			cs.mergeLastCount++
			if cs.mergeLastCount > 1 {
				cs.targetBlockSize += cs.minBlockSize
			}
		}
	}

	if isFinal {
		cs.histogramsSize = split.numTypes * nc
		split.types = split.types[:cs.numBlocks]
		split.lengths = split.lengths[:cs.numBlocks]
	}
}

// clearHistogram zeros one histogram slot at the given pool index.
func (cs *contextBlockSplitter) clearHistogram(poolIdx int) {
	start := poolIdx * cs.alphabetSize
	clear(cs.histograms[start : start+cs.alphabetSize])
}

// clearHistograms zeros n consecutive histogram slots starting at poolIdx.
func (cs *contextBlockSplitter) clearHistograms(poolIdx, n int) {
	start := poolIdx * cs.alphabetSize
	clear(cs.histograms[start : start+n*cs.alphabetSize])
}

// buildMetaBlockGreedy partitions the commands of a metablock into block types
// using the greedy block splitting algorithm.
//
// When numContexts <= 1, uses simple per-literal block splitting (Q4 path).
// When numContexts > 1, uses context-aware splitting where each literal is
// routed to a context-specific histogram based on the context map lookup.
func buildMetaBlockGreedy(
	ringbuffer []byte, pos, mask uint,
	prevByte, prevByte2 byte,
	numContexts uint, staticContextMap []uint32,
	commands []command,
	bufs *splitBufs, mb *metaBlockSplit,
) {
	// Count total literals.
	var numLiterals int
	for i := range commands {
		numLiterals += int(commands[i].insertLen)
	}

	// Initialize command and distance splitters (unchanged by context modeling).
	var cmdSplitter, distSplitter blockSplitter
	cmdSplitter, bufs.cmdHistograms, bufs.cmdTypes, bufs.cmdLengths =
		newBlockSplitter(&mb.cmdSplit, core.AlphabetSizeInsertAndCopyLength, 1024, 500.0,
			len(commands), bufs.cmdHistograms, bufs.cmdTypes, bufs.cmdLengths)
	distSplitter, bufs.distHistograms, bufs.distTypes, bufs.distLengths =
		newBlockSplitter(&mb.distSplit, 64, 512, 100.0,
			len(commands), bufs.distHistograms, bufs.distTypes, bufs.distLengths)

	if numContexts <= 1 {
		// Clear any stale context map from a previous multi-context metablock
		// so writeMetaBlock does not mistakenly use context-aware encoding.
		mb.literalContextMap = mb.literalContextMap[:0]

		// Single-context path (quality < 5): one histogram per block type.
		var litSplitter blockSplitter
		litSplitter, bufs.litHistograms, bufs.litTypes, bufs.litLengths =
			newBlockSplitter(&mb.litSplit, core.AlphabetSizeLiteral, 512, 400.0,
				numLiterals, bufs.litHistograms, bufs.litTypes, bufs.litLengths)

		for i := range commands {
			cmd := commands[i]
			cmdSplitter.addSymbol(int(cmd.cmdPrefix))
			for j := cmd.insertLen; j != 0; j-- {
				litSplitter.addSymbol(int(ringbuffer[pos&mask]))
				pos++
			}
			copyLen := cmd.copyLength()
			pos += uint(copyLen)
			if copyLen != 0 && cmd.cmdPrefix >= 128 {
				distSplitter.addSymbol(int(cmd.distPrefixCode()))
			}
		}

		litSplitter.finishBlock(true)
		mb.litHistograms = litSplitter.histograms[:litSplitter.histogramsSize*core.AlphabetSizeLiteral]
	} else {
		// Multi-context path (quality >= 5): multiple histograms per block type.
		ctxSplitter := newContextBlockSplitter(
			&mb.litSplit, core.AlphabetSizeLiteral, int(numContexts), 512, 400.0, numLiterals, bufs)
		utf8LUT := uint(core.ContextUTF8) << 9

		for i := range commands {
			cmd := commands[i]
			cmdSplitter.addSymbol(int(cmd.cmdPrefix))
			for j := cmd.insertLen; j != 0; j-- {
				literal := ringbuffer[pos&mask]
				context := staticContextMap[core.ContextLookupTable[utf8LUT+uint(prevByte)]|core.ContextLookupTable[utf8LUT+256+uint(prevByte2)]]
				ctxSplitter.addSymbol(int(literal), int(context))
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
					distSplitter.addSymbol(int(cmd.distPrefixCode()))
				}
			}
		}

		ctxSplitter.finishBlock(true)
		mb.litHistograms = ctxSplitter.histograms[:ctxSplitter.histogramsSize*core.AlphabetSizeLiteral]

		mapStaticContexts(mb, numContexts, staticContextMap)
	}

	cmdSplitter.finishBlock(true)
	distSplitter.finishBlock(true)

	mb.cmdHistograms = cmdSplitter.histograms[:cmdSplitter.histogramsSize*core.AlphabetSizeInsertAndCopyLength]
	mb.distHistograms = distSplitter.histograms[:distSplitter.histogramsSize*64]
}

// mapStaticContexts builds the literal context map for a metablock after
// context-aware block splitting. It applies the chosen static context map
// identically to every block type, mapping each of the 64 context IDs to a
// histogram index offset by the block type's cluster base.
func mapStaticContexts(mb *metaBlockSplit, numContexts uint, staticContextMap []uint32) {
	numTypes := uint(mb.litSplit.numTypes)
	mapSize := numTypes << core.LiteralContextBits
	if cap(mb.literalContextMap) < int(mapSize) {
		mb.literalContextMap = make([]uint32, mapSize)
	} else {
		mb.literalContextMap = mb.literalContextMap[:mapSize]
	}
	for i := range numTypes {
		offset := uint32(i * numContexts)
		for j := range uint(1 << core.LiteralContextBits) {
			mb.literalContextMap[(i<<core.LiteralContextBits)+j] = offset + staticContextMap[j]
		}
	}
}

// optimizeHistograms iterates all histograms in a metaBlockSplit and calls
// optimizeHuffmanCountsForRLE on each one.
func optimizeHistograms(mb *metaBlockSplit, distAlphabetSize int, goodForRLE *[]bool) {
	// Derive histogram counts from slice lengths rather than block type counts.
	// At Q4 and below (no context modeling) these are equal. At Q5+ with
	// context modeling the literal histograms may contain numTypes * numContexts
	// entries — more than numTypes alone.
	numLitHistograms := len(mb.litHistograms) / core.AlphabetSizeLiteral
	for i := range numLitHistograms {
		start := i * core.AlphabetSizeLiteral
		optimizeHuffmanCountsForRLE(mb.litHistograms[start:start+core.AlphabetSizeLiteral], goodForRLE)
	}
	numCmdHistograms := len(mb.cmdHistograms) / core.AlphabetSizeInsertAndCopyLength
	for i := range numCmdHistograms {
		start := i * core.AlphabetSizeInsertAndCopyLength
		optimizeHuffmanCountsForRLE(mb.cmdHistograms[start:start+core.AlphabetSizeInsertAndCopyLength], goodForRLE)
	}
	numDistHistograms := len(mb.distHistograms) / distAlphabetSize
	for i := range numDistHistograms {
		start := i * distAlphabetSize
		optimizeHuffmanCountsForRLE(mb.distHistograms[start:start+distAlphabetSize], goodForRLE)
	}
}
