// Zopfli Q10 entry point: single-pass DP shortest path.
//
// This is the backward-reference algorithm for quality 10. It runs a single
// DP pass that finds matches on-the-fly using the H10 binary tree hasher,
// then extracts the optimal command sequence.

package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// h10HashTypeLength is the minimum number of bytes needed for H10 to hash
// a position (4 bytes for the 32-bit hash function).
const h10HashTypeLength = 4

// zopfliComputeShortestPath runs the single-pass DP for Q10.
//
// For each position it calls h10.findAllMatches to get candidate matches,
// then updateNodes to evaluate all (insert, copy, distance) triples.
// On very long matches it calls h10.storeRange for the skipped positions.
//
// Returns the number of commands in the optimal path.
func zopfliComputeShortestPath(numBytes, position uint, ringbuffer []byte, ringBufferMask uint, quality, lgwin int, gap uint, compound *compoundDictionary, distCache []int, hasher *h10, nodes []zopfliNode, bufs *q10Bufs) uint {
	maxBackwardLimit := (uint(1) << lgwin) - core.WindowGap
	maxZopfli := maxZopfliLen(quality)
	var queue startPosQueue
	var sc dcScratch
	hasCompound := compound != nil && compound.numChunks > 0
	// When compound dictionaries are present, LZ matches are written at an
	// offset so that compound-dictionary matches can be placed before them
	// and the two sorted runs merged into matches[0:].
	lzOff := uint(0)
	if hasCompound {
		lzOff = h10MaxNumMatches + 128
	}
	matchesNeeded := 2*(h10MaxNumMatches+64) + int(lzOff)
	if cap(bufs.zMatches) < matchesNeeded {
		bufs.zMatches = make([]backwardMatch, matchesNeeded)
	} else {
		bufs.zMatches = bufs.zMatches[:matchesNeeded]
	}
	matches := bufs.zMatches
	storeEnd := position
	if numBytes >= h10MaxTreeCompLength {
		storeEnd = position + numBytes - h10MaxTreeCompLength + 1
	}

	model := &bufs.zCostModel
	model.init(64, numBytes) // distAlphabetSize=64 for NPOSTFIX=0, NDIRECT=0

	nodes[0].length = 0
	nodes[0].setCost(0)
	model.setFromLiteralCosts(position, ringbuffer, ringBufferMask)

	for i := uint(0); i+h10HashTypeLength-1 < numBytes; i++ {
		pos := position + i
		maxDistance := min(pos, maxBackwardLimit)
		maxLength := numBytes - i

		numFound := hasher.findAllMatches(
			ringbuffer, ringBufferMask, pos, maxLength, maxDistance,
			maxDistance+gap, quality, matches[lzOff:])

		if hasCompound {
			cdMatches := compound.lookupAllMatches(
				ringbuffer, ringBufferMask, pos, 3, maxLength,
				maxDistance, maxBackwardDistance,
				matches[lzOff-64:lzOff],
			)
			if cdMatches > 0 {
				mergeMatches(matches,
					matches[lzOff-64:lzOff-64+cdMatches],
					matches[lzOff:lzOff+numFound])
				numFound += cdMatches
			} else {
				copy(matches, matches[lzOff:lzOff+numFound])
			}
		}

		if numFound > 0 && matches[numFound-1].matchLength() > maxZopfli {
			matches[0] = matches[numFound-1]
			numFound = 1
		}

		skip := updateNodes(nodes, ringbuffer, distCache,
			matches[:numFound], model, &queue,
			numBytes, position, i, ringBufferMask, maxBackwardLimit, gap, compound, numFound, quality, &sc)
		if skip < longCopyQuickStep {
			skip = 0
		}
		if numFound == 1 && matches[0].matchLength() > maxZopfli {
			skip = max(matches[0].matchLength(), skip)
		}
		if skip > 1 {
			// Store the tail of the long copy in the hasher.
			hasher.storeRange(ringbuffer, ringBufferMask, pos+1, min(pos+skip, storeEnd))
			skip--
			for skip > 0 {
				i++
				if i+h10HashTypeLength-1 >= numBytes {
					break
				}
				evaluateNode(nodes, i, position, maxBackwardLimit, gap, distCache, model, &queue)
				skip--
			}
		}
	}

	return computeShortestPathFromNodes(nodes, numBytes)
}

// createZopfliBackwardReferences is the Q10 top-level entry point.
// It allocates the DP node array, runs the shortest path algorithm,
// and extracts commands.
func createZopfliBackwardReferences(numBytes, position uint, ringbuffer []byte, ringBufferMask uint, quality, lgwin int, gap uint, compound *compoundDictionary, distCache []int, hasher *h10, lastInsertLen *uint, commands *[]command, numLiterals *uint, bufs *q10Bufs) {
	maxBackwardLimit := (uint(1) << lgwin) - core.WindowGap
	needed := int(numBytes + 1)
	if cap(bufs.zNodes) < needed {
		bufs.zNodes = make([]zopfliNode, needed)
	} else {
		bufs.zNodes = bufs.zNodes[:needed]
	}
	nodes := bufs.zNodes
	initZopfliNodes(nodes)
	zopfliComputeShortestPath(numBytes, position, ringbuffer, ringBufferMask, quality, lgwin, gap, compound, distCache, hasher, nodes, bufs)
	zopfliCreateCommands(nodes, numBytes, position, maxBackwardLimit, gap, distCache, lastInsertLen, commands, numLiterals)
}
