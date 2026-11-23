package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// Iterative entropy-based block splitter for the Q10/Q11 slow path.
//
// The greedy block splitter (buildMetaBlockGreedy in block_split.go) processes
// symbols one at a time with irrevocable merge/split decisions. This slow-path
// splitter takes a different approach:
//
//  1. Initialize entropy codes by sampling the data.
//  2. Iterate (3× for Q9, 10× for Q10+):
//     - findBlocks: DP forward pass assigns each symbol to a histogram
//       (block type) minimizing total cost, with a penalty for switching.
//     - remapBlockIDs: Compact non-consecutive block IDs to [0..N).
//     - buildBlockHistograms: Rebuild histograms from assignments.
//  3. clusterBlocks: Hierarchical clustering reduces block types to
//     the format limit (256), producing the final blockSplit.
//
// A single set of functions operates on []uint16, with literal bytes widened
// at the call site.

// Per-category tuning constants for the iterative block splitter.
const (
	maxLiteralHistograms        = 100
	maxCommandHistograms        = 50
	literalBlockSwitchCost      = 28.1
	commandBlockSwitchCost      = 13.5
	distanceBlockSwitchCost     = 14.6
	literalStrideLength         = 70
	commandStrideLength         = 40
	distanceStrideLength        = 40
	symbolsPerLiteralHistogram  = 544
	symbolsPerCommandHistogram  = 530
	symbolsPerDistanceHistogram = core.NumHistogramDistanceSymbols
)

// Algorithm constants for the iterative block splitter.
const (
	histogramsPerBatch         = 64
	clustersPerBatch           = 16
	minLengthForBlockSplitting = 128
	iterMulForRefining         = 2
	minItersForRefining        = 100
)

// Quality threshold: qualities below this use 3 iterations, at or above use 10.
const hqZopflificationQuality = 11

// Below splitOffloadMinCommands, the worker handoffs cost more than the
// command and distance splits and the distance clustering they offload.
const splitOffloadMinCommands = 512

// noMinCost is findBlocks' sentinel starting cost, larger than any real one.
const noMinCost = 1e99

// splitVecParams holds per-category tuning constants for splitByteVector.
type splitVecParams struct {
	symbolsPerHistogram int
	maxHistograms       int
	samplingStride      int
	blockSwitchCost     float64
	quality             int
	alphabetSize        int
}

type splitVecBufs struct {
	svHistograms []uint32
	svBlockIDs   []byte
	svFloat      []float64
	svSwitchSig  []byte
	svNewID      []uint16

	cbHistSymbols   []uint32
	cbAllHistograms []uint32
	cbClusterSizes  []uint32
	cbBatchHist     []uint32
	cbPairs         []histogramPair
	cbTmpHist       []uint32
	cbBatchU32      []uint32
	cbBlockLengths  []uint32
	cbBatchFloat    []float64
	cbBatchTotals   []uint32
	cbClusters      []uint32
	cbNewIndex      []uint32

	sbUint16 []uint16
}

// ---------------------------------------------------------------------------
// Core splitting algorithm — all functions operate on []uint16 symbol arrays.
// ---------------------------------------------------------------------------

// initialEntropyCodes samples the data to create initial histograms. Each
// histogram is seeded from a stride-length window at an evenly-spaced position,
// with a random offset for all but the first histogram.
//
// histograms is a flat buffer of [numHistograms][alphabetSize] uint32 counts.
func initialEntropyCodes(
	data []uint16, histograms []uint32,
	length, stride, numHistograms, alphabetSize int,
) {
	seed := uint32(7)
	blockLength := length / numHistograms

	// Clear all histograms.
	clear(histograms[:numHistograms*alphabetSize])

	for i := range numHistograms {
		pos := length * i / numHistograms
		if i != 0 {
			pos += int(splitRand(&seed) % uint32(blockLength))
		}
		if pos+stride >= length {
			pos = length - stride - 1
		}
		// Add a stride-length window of symbols to this histogram.
		hist := histograms[i*alphabetSize:]
		for j := range stride {
			hist[data[pos+j]]++
		}
	}
}

// randomSample adds a random-position sample of stride symbols to the given
// histogram. If stride >= length, the entire data is used.
//
// alphabetSize is always passed for API consistency with other histogram functions.
func randomSample(
	data []uint16, sample []uint32,
	seed *uint32,
	length, stride, _ int,
) {
	pos := 0
	if stride >= length {
		stride = length
	} else {
		pos = int(splitRand(seed) % uint32(length-stride+1))
	}
	for j := range stride {
		sample[data[pos+j]]++
	}
}

// refineEntropyCodes iteratively refines histograms via random sampling. Each
// iteration adds a random sample to one histogram in round-robin order,
// progressively improving the entropy code estimates.
func refineEntropyCodes(
	data []uint16, histograms []uint32,
	length, stride, numHistograms, alphabetSize int,
) {
	iters := iterMulForRefining*length/stride + minItersForRefining
	seed := uint32(7)

	// Round up to a multiple of numHistograms so each histogram gets equal samples.
	iters = ((iters + numHistograms - 1) / numHistograms) * numHistograms

	// randomSample only increments, so sampling straight into the target
	// histogram is identical to sampling into scratch and adding it in.
	for iter := range iters {
		k := iter % numHistograms
		hist := histograms[k*alphabetSize : (k+1)*alphabetSize]
		randomSample(data, hist, &seed, length, stride, alphabetSize)
	}
}

// findBlocks assigns a block type to each symbol via dynamic programming cost
// tracking. For each position, the DP computes the coding cost under each
// histogram and tracks per-histogram cost deltas relative to the running
// minimum. When a delta reaches the block-switch cost, the position is marked
// as a potential switch point. A backtrace pass then determines actual block
// boundaries.
//
// The "prologue" region (first 2000 symbols) uses a reduced block-switch cost
// to encourage more splits at the beginning where initial histogram estimates
// may be poor.
//
// Returns the number of blocks (1 + number of block switches).
func findBlocks(
	data []uint16,
	histograms []uint32,
	insertCost []float64,
	cost []float64,
	switchSignal []byte,
	blockID []byte,
	length, numHistograms, alphabetSize int,
	blockSwitchBitcost float64,
) int {
	bitmapLen := (numHistograms + 7) >> 3
	numBlocks := 1

	if numHistograms <= 1 {
		clear(blockID[:length])
		return 1
	}

	// Fill the per-symbol, per-histogram insertion costs.
	// For each histogram j and symbol i:
	//   insertCost[i * numHistograms + j] = log2(totalCount_j) - symbolBitCost(count_j_i)
	//
	// This represents the cost in bits to encode symbol i using histogram j.
	// The reverse loop below writes every element before any read, so no
	// pre-zeroing is needed.
	for j := range numHistograms {
		hist := histograms[j*alphabetSize:]
		var totalCount uint32
		for i := range alphabetSize {
			totalCount += hist[i]
		}
		insertCost[j] = fastLog2(int(totalCount))
	}
	for i := alphabetSize - 1; i >= 0; i-- {
		for j := range numHistograms {
			insertCost[i*numHistograms+j] =
				insertCost[j] - symbolBitCost(int(histograms[j*alphabetSize+i]))
		}
	}

	// DP forward pass: track the minimum cost path through histograms.
	clear(cost[:numHistograms])
	clear(switchSignal[:length*bitmapLen])

	findBlocksDP(data[:length], insertCost, cost[:numHistograms], switchSignal, blockID, blockSwitchBitcost)

	// Backtrace from the last position to determine block boundaries.
	byteIx := length - 1
	ix := byteIx * bitmapLen
	curID := blockID[byteIx]
	for byteIx > 0 {
		mask := byte(1 << (curID & 7))
		byteIx--
		ix -= bitmapLen
		if switchSignal[ix+int(curID>>3)]&mask != 0 {
			if curID != blockID[byteIx] {
				curID = blockID[byteIx]
				numBlocks++
			}
		}
		blockID[byteIx] = curID
	}

	return numBlocks
}

// remapBlockIDs compacts block IDs to the consecutive range [0..N). Block IDs
// are renumbered in order of first appearance. Returns the number of distinct
// block types.
func remapBlockIDs(blockIDs []byte, length int, newID []uint16, numHistograms int) int {
	const invalidID = 256
	nextID := uint16(0)
	for i := range numHistograms {
		newID[i] = invalidID
	}
	for i := range length {
		if newID[blockIDs[i]] == invalidID {
			newID[blockIDs[i]] = nextID
			nextID++
		}
	}
	for i := range length {
		blockIDs[i] = byte(newID[blockIDs[i]])
	}
	return int(nextID)
}

// buildBlockHistograms rebuilds the flat histogram array from the current
// block type assignments. Each symbol data[i] is added to the histogram
// designated by blockIDs[i].
func buildBlockHistograms(
	data []uint16, blockIDs []byte, histograms []uint32,
	length, numHistograms, alphabetSize int,
) {
	clear(histograms[:numHistograms*alphabetSize])
	for i := range length {
		histograms[int(blockIDs[i])*alphabetSize+int(data[i])]++
	}
}

// ---------------------------------------------------------------------------
// Block clustering
// ---------------------------------------------------------------------------

// clusterBlocks clusters the initial block partition into a compact blockSplit,
// merging similar blocks via histogram clustering. The algorithm:
//
//  1. Compute block lengths from the per-symbol block ID assignments.
//  2. Pre-cluster blocks in batches of 64 using histogramCombine.
//  3. Final clustering across all batches.
//  4. Assign each block to the best final cluster via histogramBitCostDistance.
//  5. Write the compacted result into the blockSplit.
func clusterBlocks(
	split *blockSplit,
	bufs *splitVecBufs,
	data []uint16, blockIDs []byte,
	length, numBlocks, alphabetSize int,
) {
	bufs.cbHistSymbols = growUint32(bufs.cbHistSymbols, numBlocks)
	histogramSymbols := bufs.cbHistSymbols[:numBlocks]
	expectedNumClusters := clustersPerBatch *
		(numBlocks + histogramsPerBatch - 1) / histogramsPerBatch

	bufs.cbAllHistograms = growUint32(bufs.cbAllHistograms, expectedNumClusters*alphabetSize)
	allHistograms := bufs.cbAllHistograms
	allHistogramsSize := 0
	bufs.cbClusterSizes = growUint32(bufs.cbClusterSizes, expectedNumClusters)
	clusterSizes := bufs.cbClusterSizes
	clusterSizesLen := 0
	numClusters := 0

	bufs.cbBatchHist = growUint32(bufs.cbBatchHist, min(numBlocks, histogramsPerBatch)*alphabetSize)
	batchHistograms := bufs.cbBatchHist
	maxNumPairs := histogramsPerBatch * histogramsPerBatch / 2
	bufs.cbPairs = growHistogramPairs(bufs.cbPairs, maxNumPairs+1)
	pairs := bufs.cbPairs
	bufs.cbTmpHist = growUint32(bufs.cbTmpHist, 2*alphabetSize)
	tmpHist := bufs.cbTmpHist

	// Scratch arrays for batch clustering, reused across batches.
	// Combined allocation: sizes + newClusters + symbols + remap (4×64).
	bufs.cbBatchU32 = growUint32(bufs.cbBatchU32, 4*histogramsPerBatch)
	sizes := bufs.cbBatchU32[0*histogramsPerBatch : 1*histogramsPerBatch]
	newClusters := bufs.cbBatchU32[1*histogramsPerBatch : 2*histogramsPerBatch]
	symbols := bufs.cbBatchU32[2*histogramsPerBatch : 3*histogramsPerBatch]
	remap := bufs.cbBatchU32[3*histogramsPerBatch : 4*histogramsPerBatch]
	bufs.cbBlockLengths = growUint32Clear(bufs.cbBlockLengths, numBlocks)
	blockLengths := bufs.cbBlockLengths[:numBlocks]

	// Compute block lengths from per-symbol block ID assignments.
	blockIdx := 0
	for i := range length {
		blockLengths[blockIdx]++
		if i+1 == length || blockIDs[i] != blockIDs[i+1] {
			blockIdx++
		}
	}

	// Pre-cluster blocks in batches.
	// Combined: batchBitCosts reuses cbBatchFloat, batchTotals reuses cbBatchTotals.
	bufs.cbBatchFloat = growFloat64(bufs.cbBatchFloat, histogramsPerBatch)
	bufs.cbBatchTotals = growUint32(bufs.cbBatchTotals, histogramsPerBatch)
	pos := 0
	for i := 0; i < numBlocks; i += histogramsPerBatch {
		numToCombine := min(numBlocks-i, histogramsPerBatch)

		for j := range numToCombine {
			bl := int(blockLengths[i+j])
			hist := batchHistograms[j*alphabetSize : (j+1)*alphabetSize]
			clear(hist)
			for range bl {
				hist[data[pos]]++
				pos++
			}
			newClusters[j] = uint32(j)
			symbols[j] = uint32(j)
			sizes[j] = 1
		}

		// Compute bit costs for the batch histograms. histogramCombine needs
		// bitCosts and totalCounts parallel arrays.
		batchBitCosts := bufs.cbBatchFloat[:numToCombine]
		batchTotalCounts := bufs.cbBatchTotals[:numToCombine]
		for j := range numToCombine {
			hist := batchHistograms[j*alphabetSize : (j+1)*alphabetSize]
			batchBitCosts[j] = populationCost(hist, alphabetSize)
			batchTotalCounts[j] = histogramTotalCount(hist, alphabetSize)
		}

		numNewClusters := histogramCombine(
			batchHistograms, alphabetSize,
			batchBitCosts, batchTotalCounts,
			sizes, symbols[:numToCombine], newClusters[:numToCombine],
			pairs,
			numToCombine, numToCombine, histogramsPerBatch, maxNumPairs,
			tmpHist,
		)

		// Grow allHistograms and clusterSizes if needed.
		needed := (allHistogramsSize + numNewClusters) * alphabetSize
		if needed > len(allHistograms) {
			grown := make([]uint32, needed*2)
			copy(grown, allHistograms[:allHistogramsSize*alphabetSize])
			allHistograms = grown
			bufs.cbAllHistograms = allHistograms
		}
		if clusterSizesLen+numNewClusters > len(clusterSizes) {
			grown := make([]uint32, (clusterSizesLen+numNewClusters)*2)
			copy(grown, clusterSizes[:clusterSizesLen])
			clusterSizes = grown
			bufs.cbClusterSizes = clusterSizes
		}

		for j := range numNewClusters {
			srcIdx := int(newClusters[j])
			copy(
				allHistograms[allHistogramsSize*alphabetSize:(allHistogramsSize+1)*alphabetSize],
				batchHistograms[srcIdx*alphabetSize:(srcIdx+1)*alphabetSize],
			)
			allHistogramsSize++
			clusterSizes[clusterSizesLen] = sizes[srcIdx]
			clusterSizesLen++
			remap[srcIdx] = uint32(j)
		}
		for j := range numToCombine {
			histogramSymbols[i+j] = uint32(numClusters) + remap[symbols[j]]
		}
		numClusters += numNewClusters
	}

	// Final clustering across all pre-clustered batches.
	maxNumPairsFinal := min(64*numClusters, (numClusters/2)*numClusters)
	if maxNumPairsFinal+1 > len(pairs) {
		bufs.cbPairs = growHistogramPairs(bufs.cbPairs, maxNumPairsFinal+1)
		pairs = bufs.cbPairs
	}

	bufs.cbClusters = growUint32(bufs.cbClusters, numClusters)
	clusters := bufs.cbClusters[:numClusters]
	for i := range numClusters {
		clusters[i] = uint32(i)
	}

	// Reuse combined float buffer for allBitCosts (needs numClusters, batch needs 64).
	bufs.cbBatchFloat = growFloat64(bufs.cbBatchFloat, numClusters)
	allBitCosts := bufs.cbBatchFloat[:numClusters]
	// Reuse combined uint32 buffer for allTotalCounts.
	bufs.cbBatchTotals = growUint32(bufs.cbBatchTotals, numClusters)
	allTotalCounts := bufs.cbBatchTotals[:numClusters]
	for i := range numClusters {
		hist := allHistograms[i*alphabetSize : (i+1)*alphabetSize]
		allBitCosts[i] = populationCost(hist, alphabetSize)
		allTotalCounts[i] = histogramTotalCount(hist, alphabetSize)
	}

	numFinalClusters := histogramCombine(
		allHistograms, alphabetSize,
		allBitCosts, allTotalCounts,
		clusterSizes, histogramSymbols, clusters,
		pairs,
		numClusters, numBlocks, maxNumberOfBlockTypes, maxNumPairsFinal,
		tmpHist,
	)

	// Assign each block to its best final cluster.
	const invalidIndex = ^uint32(0)
	bufs.cbNewIndex = growUint32(bufs.cbNewIndex, numClusters)
	newIndex := bufs.cbNewIndex[:numClusters]
	for i := range newIndex {
		newIndex[i] = invalidIndex
	}

	pos = 0
	nextIndex := uint32(0)
	for i := range numBlocks {
		bl := int(blockLengths[i])
		// Build histogram for this block.
		clear(tmpHist[:alphabetSize])
		for range bl {
			tmpHist[data[pos]]++
			pos++
		}

		// Among equally good histograms, prefer the last used one.
		bestOut := histogramSymbols[0]
		if i != 0 {
			bestOut = histogramSymbols[i-1]
		}
		bestBits := histogramBitCostDistance(
			tmpHist[:alphabetSize],
			allHistograms[int(bestOut)*alphabetSize:(int(bestOut)+1)*alphabetSize],
			tmpHist[alphabetSize:2*alphabetSize],
			alphabetSize,
			allBitCosts[bestOut], allTotalCounts[bestOut],
		)

		for j := range numFinalClusters {
			ci := int(clusters[j])
			curBits := histogramBitCostDistance(
				tmpHist[:alphabetSize],
				allHistograms[ci*alphabetSize:(ci+1)*alphabetSize],
				tmpHist[alphabetSize:2*alphabetSize],
				alphabetSize,
				allBitCosts[clusters[j]], allTotalCounts[clusters[j]],
			)
			if curBits < bestBits {
				bestBits = curBits
				bestOut = clusters[j]
			}
		}
		histogramSymbols[i] = bestOut
		if newIndex[bestOut] == invalidIndex {
			newIndex[bestOut] = nextIndex
			nextIndex++
		}
	}

	// Write the final block split, merging consecutive blocks with the same type.
	split.types = growByte(split.types, numBlocks)[:numBlocks]
	split.lengths = growUint32(split.lengths, numBlocks)[:numBlocks]

	var curLength uint32
	splitIdx := 0
	var maxType byte
	for i := range numBlocks {
		curLength += blockLengths[i]
		if i+1 == numBlocks || histogramSymbols[i] != histogramSymbols[i+1] {
			id := byte(newIndex[histogramSymbols[i]])
			split.types[splitIdx] = id
			split.lengths[splitIdx] = curLength
			if id > maxType {
				maxType = id
			}
			curLength = 0
			splitIdx++
		}
	}
	split.types = split.types[:splitIdx]
	split.lengths = split.lengths[:splitIdx]
	split.numTypes = int(maxType) + 1
}

// ---------------------------------------------------------------------------
// Top-level orchestrators
// ---------------------------------------------------------------------------

// splitByteVector partitions a symbol stream into typed blocks using iterative
// entropy-code refinement and DP block assignment. This is the core of the
// slow-path block splitter.
//
// The algorithm:
//  1. Estimate the initial number of histograms from data length.
//  2. Seed histograms via initialEntropyCodes and refine them.
//  3. Iterate findBlocks → remapBlockIDs → buildBlockHistograms to converge
//     on optimal block assignments.
//  4. Cluster the resulting blocks via clusterBlocks.
func splitByteVector(
	split *blockSplit,
	bufs *splitVecBufs,
	data []uint16, length int,
	p splitVecParams,
) {
	// Estimate the number of histograms; capped at maxHistograms.
	numHistograms := min(length/p.symbolsPerHistogram+1, p.maxHistograms)

	if length == 0 {
		split.numTypes = 1
		return
	}

	if length < minLengthForBlockSplitting {
		split.types = append(split.types, 0)
		split.lengths = append(split.lengths, uint32(length))
		split.numTypes = 1
		return
	}

	// Allocate histograms.
	bufs.svHistograms = growUint32Clear(bufs.svHistograms, numHistograms*p.alphabetSize)
	histograms := bufs.svHistograms[:numHistograms*p.alphabetSize]

	// Seed and refine entropy codes.
	initialEntropyCodes(data, histograms, length, p.samplingStride, numHistograms, p.alphabetSize)
	refineEntropyCodes(data, histograms, length, p.samplingStride, numHistograms, p.alphabetSize)

	// Allocate DP scratch buffers.
	bufs.svBlockIDs = growByte(bufs.svBlockIDs, length)
	blockIDs := bufs.svBlockIDs[:length]
	bitmapLen := (numHistograms + 7) >> 3
	floatNeeded := p.alphabetSize*numHistograms + numHistograms
	bufs.svFloat = growFloat64(bufs.svFloat, floatNeeded)
	insertCost := bufs.svFloat[:p.alphabetSize*numHistograms]
	cost := bufs.svFloat[p.alphabetSize*numHistograms : floatNeeded]
	bufs.svSwitchSig = growByte(bufs.svSwitchSig, length*bitmapLen)
	switchSignal := bufs.svSwitchSig[:length*bitmapLen]
	bufs.svNewID = growUint16(bufs.svNewID, numHistograms)
	newID := bufs.svNewID[:numHistograms]

	// Determine iteration count based on quality.
	iters := 3
	if p.quality >= hqZopflificationQuality {
		iters = 10
	}

	var numBlocks int
	for range iters {
		numBlocks = findBlocks(data,
			histograms, insertCost, cost, switchSignal, blockIDs,
			length, numHistograms, p.alphabetSize, p.blockSwitchCost)
		numHistograms = remapBlockIDs(blockIDs, length, newID, numHistograms)
		buildBlockHistograms(data, blockIDs, histograms, length, numHistograms, p.alphabetSize)
	}

	clusterBlocks(split, bufs, data, blockIDs, length, numBlocks, p.alphabetSize)
}

// splitBlock partitions the commands of a metablock into three blockSplits
// (literal, command, distance) using the slow-path iterative splitter.
//
// It extracts three flat symbol arrays from the command stream:
//  1. Literals: bytes from copyLiteralsToByteArray, widened to []uint16.
//  2. Command prefixes: cmd.cmdPrefix for each command.
//  3. Distance prefixes: cmd.distPrefix & 0x3FF for commands with non-zero
//     copy length and cmdPrefix >= 128 (i.e., commands that encode a distance).
//
// Each array is passed to splitByteVector with its category-specific tuning
// constants.
func splitBlock(
	litSplit, cmdSplit, distSplit *blockSplit,
	bufs *q10Bufs,
	cmds []command,
	data []byte, pos, mask uint,
	quality int,
) {
	bufs.sbLiteralBytes = copyLiteralsToByteArrayBuf(cmds, data, pos, mask, bufs.sbLiteralBytes)
	splitLiterals(litSplit, bufs, quality)
	splitCommands(cmdSplit, &bufs.splitVecBufs, cmds, quality)
	splitDistances(distSplit, &bufs.splitVecBufs, cmds, quality)
}

func splitBlockParallel(
	litSplit, cmdSplit, distSplit *blockSplit,
	bufs *q10Bufs,
	cmds []command,
	data []byte, pos, mask uint,
	quality int,
) distanceParams {
	bufs.sbLiteralBytes = copyLiteralsToByteArrayBuf(cmds, data, pos, mask, bufs.sbLiteralBytes)

	cmdWorker := &bufs.hqCollector
	cmdWorker.cmds = cmds
	cmdWorker.cmdSplit = cmdSplit
	cmdWorker.quality = quality
	cmdWorker.begin(hqJobSplitCommands)

	distWorker := &bufs.hqHelper
	distWorker.bufs = bufs
	distWorker.cmds = cmds
	distWorker.distSplit = distSplit
	distWorker.quality = quality
	distWorker.begin(hqJobSplitDistances)

	splitLiterals(litSplit, bufs, quality)
	cmdWorker.wait()
	distWorker.wait()
	return distWorker.distParams
}

func splitLiterals(litSplit *blockSplit, bufs *q10Bufs, quality int) {
	numLiterals := len(bufs.sbLiteralBytes)
	bufs.sbUint16 = growUint16(bufs.sbUint16, numLiterals)
	symbols := bufs.sbUint16[:numLiterals]
	for i, b := range bufs.sbLiteralBytes {
		symbols[i] = uint16(b)
	}
	splitByteVector(litSplit, &bufs.splitVecBufs, symbols, numLiterals, splitVecParams{
		symbolsPerHistogram: symbolsPerLiteralHistogram,
		maxHistograms:       maxLiteralHistograms,
		samplingStride:      literalStrideLength,
		blockSwitchCost:     literalBlockSwitchCost,
		quality:             quality,
		alphabetSize:        core.AlphabetSizeLiteral,
	})
}

func splitCommands(cmdSplit *blockSplit, bufs *splitVecBufs, cmds []command, quality int) {
	bufs.sbUint16 = growUint16(bufs.sbUint16, len(cmds))
	symbols := bufs.sbUint16[:len(cmds)]
	for i := range cmds {
		symbols[i] = cmds[i].cmdPrefix
	}
	splitByteVector(cmdSplit, bufs, symbols, len(cmds), splitVecParams{
		symbolsPerHistogram: symbolsPerCommandHistogram,
		maxHistograms:       maxCommandHistograms,
		samplingStride:      commandStrideLength,
		blockSwitchCost:     commandBlockSwitchCost,
		quality:             quality,
		alphabetSize:        core.AlphabetSizeInsertAndCopyLength,
	})
}

func splitDistances(distSplit *blockSplit, bufs *splitVecBufs, cmds []command, quality int) {
	bufs.sbUint16 = growUint16(bufs.sbUint16, len(cmds))
	symbols := bufs.sbUint16[:len(cmds)]
	j := 0
	for i := range cmds {
		cmd := &cmds[i]
		if cmd.copyLength() != 0 && cmd.cmdPrefix >= 128 {
			symbols[j] = cmd.distPrefix & 0x3FF
			j++
		}
	}
	splitByteVector(distSplit, bufs, symbols[:j], j, splitVecParams{
		symbolsPerHistogram: symbolsPerDistanceHistogram,
		maxHistograms:       maxCommandHistograms,
		samplingStride:      distanceStrideLength,
		blockSwitchCost:     distanceBlockSwitchCost,
		quality:             quality,
		alphabetSize:        core.NumHistogramDistanceSymbols,
	})
}

// ---------------------------------------------------------------------------
// Symbol extraction helpers
// ---------------------------------------------------------------------------

// countLiterals sums insertLen across all commands, returning the total number
// of literal bytes in the command stream.
func countLiterals(cmds []command) int {
	total := 0
	for i := range cmds {
		total += int(cmds[i].insertLen)
	}
	return total
}

// copyLiteralsToByteArray extracts literal bytes from the ring buffer into a
// flat array by replaying each command's insertion. The ring buffer is accessed
// via data[pos & mask] where mask = ringBufferSize - 1.
func copyLiteralsToByteArray(cmds []command, data []byte, pos, mask uint) []byte {
	total := countLiterals(cmds)
	literals := make([]byte, total)
	copyLiteralsInto(cmds, data, pos, mask, literals)
	return literals
}

// copyLiteralsToByteArrayBuf is like copyLiteralsToByteArray but reuses buf.
func copyLiteralsToByteArrayBuf(cmds []command, data []byte, pos, mask uint, buf []byte) []byte {
	total := countLiterals(cmds)
	buf = growByte(buf, total)
	copyLiteralsInto(cmds, data, pos, mask, buf[:total])
	return buf[:total]
}

func copyLiteralsInto(cmds []command, data []byte, pos, mask uint, literals []byte) {
	dst := 0
	from := pos & mask
	for i := range cmds {
		insertLen := uint(cmds[i].insertLen)
		if from+insertLen > mask {
			// Wrap around the ring buffer boundary.
			head := mask + 1 - from
			copy(literals[dst:], data[from:from+head])
			from = 0
			dst += int(head)
			insertLen -= head
		}
		if insertLen > 0 {
			copy(literals[dst:], data[from:from+insertLen])
			dst += int(insertLen)
		}
		from = (from + insertLen + uint(cmds[i].copyLength())) & mask
	}
}

// splitRand is a simple LCG PRNG used for sampling during entropy code
// initialization. The initial seed should be 7, which gives a loop length
// of 2^29.
func splitRand(seed *uint32) uint32 {
	*seed *= 16807
	return *seed
}

// symbolBitCost returns the approximate bit cost for a symbol count:
// -log2(count) for positive counts, or -2.0 for zero (signaling "absent
// symbol" with a small penalty).
func symbolBitCost(count int) float64 {
	if count == 0 {
		return -2.0
	}
	return fastLog2(count)
}

// ---------------------------------------------------------------------------
// Slice-growth helpers
// ---------------------------------------------------------------------------

// growUint32 returns s with length n, reusing existing capacity when possible.
func growUint32(s []uint32, n int) []uint32 {
	if cap(s) >= n {
		return s[:n]
	}
	return make([]uint32, n)
}

// growUint32Clear returns s with length n zeroed, reusing existing capacity.
func growUint32Clear(s []uint32, n int) []uint32 {
	s = growUint32(s, n)
	clear(s)
	return s
}

// growByte returns s with length n, reusing existing capacity when possible.
func growByte(s []byte, n int) []byte {
	if cap(s) >= n {
		return s[:n]
	}
	return make([]byte, n)
}

// growFloat64 returns s with length n, reusing existing capacity when possible.
func growFloat64(s []float64, n int) []float64 {
	if cap(s) >= n {
		return s[:n]
	}
	return make([]float64, n)
}

// growUint16 returns s with length n, reusing existing capacity when possible.
func growUint16(s []uint16, n int) []uint16 {
	if cap(s) >= n {
		return s[:n]
	}
	return make([]uint16, n)
}

// growHistogramPairs returns s with length n, reusing existing capacity.
func growHistogramPairs(s []histogramPair, n int) []histogramPair {
	if cap(s) >= n {
		return s[:n]
	}
	return make([]histogramPair, n)
}
