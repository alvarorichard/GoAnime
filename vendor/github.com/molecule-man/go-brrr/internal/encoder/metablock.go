// Metablock encoding for the Brotli format (RFC 7932, section 9.2).
//
// Contains both the MLEN encoding helper and the slow-path metablock
// construction for Q10+ compression. The slow path differs from the greedy
// metablock builder in three ways:
//   1. Distance parameter optimization: searches all valid (NPOSTFIX, NDIRECT)
//      combinations to minimize distance entropy cost.
//   2. DP-based block splitting: uses splitBlock (iterative shortest-path)
//      instead of the greedy single-pass splitter.
//   3. Histogram clustering: clusters literal and distance histograms to
//      reduce Huffman code count, producing context maps.

package encoder

import (
	"math/bits"

	"github.com/molecule-man/go-brrr/internal/core"
)

// maxHistograms is the maximum number of histogram clusters after merging.
// Histogram IDs must fit in one byte.
const maxHistograms = 256

// minUTF8Ratio is the minimum fraction of valid UTF-8 bytes required to
// select UTF-8 context mode over signed context mode.
const minUTF8Ratio = 0.75

// mlenEncoding holds the encoded representation of a metablock length (MLEN).
type mlenEncoding struct {
	bits       uint64 // value bits (length - 1)
	numBits    uint   // total data bits (mnibbles * 4)
	nibbleBits uint64 // 2-bit MNIBBLES field (mnibbles - 4)
}

// encodeMlen encodes a metablock byte length into its MLEN representation.
// Length must be in [1, 1<<24].
func encodeMlen(length int) mlenEncoding {
	lg := bits.Len(uint(length - 1))
	if lg == 0 {
		lg = 1
	}
	mnibbles := uint(max(lg, 16)+3) / 4
	return mlenEncoding{
		bits:       uint64(length - 1),
		numBits:    mnibbles * 4,
		nibbleBits: uint64(mnibbles - 4),
	}
}

// buildMetaBlock constructs a metablock using the slow-path block splitter
// with distance parameter optimization and histogram clustering.
//
// The algorithm proceeds in five phases:
//
//  1. Distance parameter search — tests all valid (NPOSTFIX, NDIRECT)
//     combinations to minimize total distance entropy cost. The commands'
//     distance prefix codes are recomputed with the optimal parameters.
//
//  2. Block splitting — calls splitBlock (iterative DP) to partition the
//     command stream into literal, command, and distance blocks.
//
//  3. Histogram building — builds per-block-type × per-context histograms
//     for all three symbol categories. When literal context modeling is
//     enabled, each literal block type gets 64 histograms (one per context);
//     each distance block type gets 4 histograms (one per distance context).
//
//  4. Literal histogram clustering — merges the literal histograms into at
//     most 256 clusters and produces a literal context map. When context
//     modeling is disabled, the single-context assignments are broadcast
//     across all 64 context slots per block type.
//
//  5. Distance histogram clustering — merges the distance histograms into
//     at most 256 clusters and produces a distance context map.
//
// Returns the chosen distanceParams so the caller can update the encoder's
// distance alphabet size for Huffman encoding.
func buildMetaBlock(
	ringbuffer []byte,
	pos, mask uint,
	quality int,
	prevByte, prevByte2 byte,
	cmds []command,
	literalContextMode byte,
	disableLiteralContextModeling bool,
	mb *metaBlockSplit,
	bufs *q10Bufs,
) distanceParams {
	// Phase 1 and 2: distance parameter search and block splitting.
	mb.litSplit.reset()
	mb.cmdSplit.reset()
	mb.distSplit.reset()
	parallel := bufs.parallel && len(cmds) >= splitOffloadMinCommands
	var bestParams distanceParams
	if parallel {
		bestParams = splitBlockParallel(&mb.litSplit, &mb.cmdSplit, &mb.distSplit, bufs,
			cmds, ringbuffer, pos, mask, quality)
	} else {
		bestParams = optimizeDistanceParams(cmds, &bufs.bmTmpHist)
		splitBlock(&mb.litSplit, &mb.cmdSplit, &mb.distSplit, bufs,
			cmds, ringbuffer, pos, mask, quality)
	}

	// Phase 3: Build histograms with context.
	distAlphabetSize := int(bestParams.alphabetSizeMax)
	literalContextMultiplier := 1
	var contextModes []byte
	if !disableLiteralContextModeling {
		literalContextMultiplier = 1 << core.LiteralContextBits
		bufs.bmContextModes = growByte(bufs.bmContextModes, mb.litSplit.numTypes)
		contextModes = bufs.bmContextModes[:mb.litSplit.numTypes]
		for i := range contextModes {
			contextModes[i] = literalContextMode
		}
	}

	litHistSize := mb.litSplit.numTypes * literalContextMultiplier
	bufs.bmLitHist = growUint32Clear(bufs.bmLitHist, litHistSize*core.AlphabetSizeLiteral)
	litHistograms := bufs.bmLitHist[:litHistSize*core.AlphabetSizeLiteral]

	distHistSize := mb.distSplit.numTypes << core.DistanceContextBits
	bufs.bmDistHist = growUint32Clear(bufs.bmDistHist, distHistSize*distAlphabetSize)
	distHistograms := bufs.bmDistHist[:distHistSize*distAlphabetSize]

	mb.cmdHistograms = growUint32Clear(mb.cmdHistograms, mb.cmdSplit.numTypes*core.AlphabetSizeInsertAndCopyLength)

	buildHistogramsWithContext(cmds, mb, ringbuffer, pos, mask,
		prevByte, prevByte2, contextModes, distAlphabetSize,
		litHistograms, mb.cmdHistograms, distHistograms)

	// Phase 4: Cluster literal histograms.
	distContextMapSize := mb.distSplit.numTypes << core.DistanceContextBits
	bufs.bmDistOutHist = growUint32(bufs.bmDistOutHist, distContextMapSize*distAlphabetSize)
	distOutHistograms := bufs.bmDistOutHist[:distContextMapSize*distAlphabetSize]
	distWorker := &bufs.hqHelper
	if parallel {
		distWorker.bufs = bufs
		distWorker.clusterIn = distHistograms
		distWorker.clusterInSize = distHistSize
		distWorker.clusterAlphabet = distAlphabetSize
		distWorker.clusterOut = distOutHistograms
		distWorker.begin(hqJobClusterDistances)
	}

	litContextMapSize := mb.litSplit.numTypes << core.LiteralContextBits
	mb.literalContextMap = growUint32(mb.literalContextMap, litContextMapSize)

	bufs.bmLitOutHist = growUint32(bufs.bmLitOutHist, litContextMapSize*core.AlphabetSizeLiteral)
	litOutHistograms := bufs.bmLitOutHist[:litContextMapSize*core.AlphabetSizeLiteral]
	litOutSize, litSymbols := clusterHistograms(
		litHistograms, litHistSize, core.AlphabetSizeLiteral,
		maxHistograms, litOutHistograms, &bufs.clusterBufs)
	mb.litHistograms = litOutHistograms[:litOutSize*core.AlphabetSizeLiteral]

	// Build the literal context map from cluster assignments.
	copy(mb.literalContextMap, litSymbols)

	// If context modeling was disabled, the clustering operated on one
	// histogram per block type. Distribute each assignment across all 64
	// context slots for that block type (iterate in reverse to avoid
	// overwriting source values).
	if disableLiteralContextModeling {
		for i := mb.litSplit.numTypes; i > 0; {
			i--
			for j := range 1 << core.LiteralContextBits {
				mb.literalContextMap[(i<<core.LiteralContextBits)+j] = mb.literalContextMap[i]
			}
		}
	}

	// Phase 5: Cluster distance histograms.
	mb.distanceContextMap = growUint32(mb.distanceContextMap, distContextMapSize)
	var distOutSize int
	var distSymbols []uint32
	if parallel {
		distWorker.wait()
		distOutSize, distSymbols = distWorker.clusterOutSize, distWorker.clusterSymbols
	} else {
		distOutSize, distSymbols = clusterHistograms(
			distHistograms, distHistSize, distAlphabetSize,
			maxHistograms, distOutHistograms, &bufs.distClusterBufs)
	}
	mb.distHistograms = distOutHistograms[:distOutSize*distAlphabetSize]

	copy(mb.distanceContextMap, distSymbols)

	return bestParams
}

func optimizeDistanceParams(cmds []command, tmpHistBuf *[]uint32) distanceParams {
	origParams := initDistanceParams(0, 0)
	bestParams := origParams
	bestCost := 1e99
	checkOrig := true

	*tmpHistBuf = growUint32(*tmpHistBuf, int(origParams.alphabetSizeMax))
	tmpHist := *tmpHistBuf
	var ndirectMSB uint32

	for npostfix := uint32(0); npostfix <= 3; npostfix++ {
		for ; ndirectMSB < 16; ndirectMSB++ {
			ndirect := ndirectMSB << npostfix
			candidate := initDistanceParams(npostfix, ndirect)

			if npostfix == origParams.postfixBits &&
				ndirect == origParams.numDirectCodes {
				checkOrig = false
			}

			if int(candidate.alphabetSizeLimit) > len(tmpHist) {
				*tmpHistBuf = growUint32(*tmpHistBuf, int(candidate.alphabetSizeLimit))
				tmpHist = *tmpHistBuf
			}

			cost, ok := computeDistanceCost(cmds, origParams, candidate,
				tmpHist[:candidate.alphabetSizeLimit])
			if !ok || cost > bestCost {
				break
			}
			bestCost = cost
			bestParams = candidate
		}
		if ndirectMSB > 0 {
			ndirectMSB--
		}
		ndirectMSB /= 2
	}

	if checkOrig {
		if int(origParams.alphabetSizeLimit) > len(tmpHist) {
			*tmpHistBuf = growUint32(*tmpHistBuf, int(origParams.alphabetSizeLimit))
			tmpHist = *tmpHistBuf
		}
		cost, _ := computeDistanceCost(cmds, origParams, origParams,
			tmpHist[:origParams.alphabetSizeLimit])
		if cost < bestCost {
			bestParams = origParams
		}
	}

	recomputeDistancePrefixes(cmds, origParams, bestParams)
	return bestParams
}

// chooseContextMode selects the literal context mode for the slow-path
// metablock builder. For Q10+ data that is mostly UTF-8, core.ContextUTF8 is
// used; otherwise core.ContextSigned is selected.
func chooseContextMode(quality int, data []byte, pos, mask, length uint) byte {
	if quality >= 10 && !isMostlyUTF8(data, pos, mask, length) {
		return core.ContextSigned
	}
	return core.ContextUTF8
}
