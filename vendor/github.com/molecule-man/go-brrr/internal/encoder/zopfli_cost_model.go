// Zopfli bit-cost estimation model.
//
// The cost model estimates how many bits a command will take in the final
// compressed stream. The Zopfli DP uses these estimates to find the command
// sequence that minimizes total bit cost.
//
// Two initialization modes:
//   - From literal costs: uses per-byte frequency-based estimates for literals
//     and log2-based approximations for command/distance symbols. Used as the
//     initial cost model for both Q10 and Q11.
//   - From commands: builds histograms from a previous command sequence and
//     derives Shannon bit-costs. Used in Q11's second DP pass to refine
//     the cost model based on actual symbol distributions.

package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// maxEffectiveDistAlphabetSize is the maximum distance alphabet size for
// the Zopfli algorithm: calculateDistanceCodeLimit(maxAllowedDistance, 3, 120).
const maxEffectiveDistAlphabetSize = core.NumHistogramDistanceSymbols

// infinity is the sentinel cost used to mark unreachable positions.
const infinity = float32(1.7e38)

// zopfliCostModel estimates the bit cost of commands for the Zopfli DP.
type zopfliCostModel struct {
	// Pointer-containing fields grouped first (reduces GC scan area).
	costDist              []float32
	literalCosts          []float32 // cumulative; length = numBytes + 2
	costCmd               [core.AlphabetSizeInsertAndCopyLength]float32
	literalHistograms     [3 * 256]uint // scratch for estimateBitCostsForLiterals
	arena                 zopfliCostModelArena
	distanceHistogramSize uint
	numBytes              uint
	minCostCmd            float32
}

// zopfliCostModelArena holds temporary histograms for setFromCommands,
// avoiding allocation in the hot path.
type zopfliCostModelArena struct {
	histogramLiteral [core.AlphabetSizeLiteral]uint32
	histogramCmd     [core.AlphabetSizeInsertAndCopyLength]uint32
	histogramDist    [maxEffectiveDistAlphabetSize]uint32
	costLiteral      [core.AlphabetSizeLiteral]float32
}

// init allocates the cost model's variable-length slices.
func (m *zopfliCostModel) init(distAlphabetSize, numBytes uint) {
	m.numBytes = numBytes
	m.distanceHistogramSize = distAlphabetSize
	if cap(m.literalCosts) < int(numBytes+2) {
		m.literalCosts = make([]float32, numBytes+2)
	} else {
		m.literalCosts = m.literalCosts[:numBytes+2]
	}
	if cap(m.costDist) < int(distAlphabetSize) {
		m.costDist = make([]float32, distAlphabetSize)
	} else {
		m.costDist = m.costDist[:distAlphabetSize]
	}
}

// setFromLiteralCosts initializes the cost model from per-byte literal
// cost estimates. Command and distance costs use rough log2-based
// approximations.
func (m *zopfliCostModel) setFromLiteralCosts(position uint, ringbuffer []byte, ringBufferMask uint) {
	literalCosts := m.literalCosts
	costDist := m.costDist
	costCmd := &m.costCmd
	numBytes := m.numBytes

	estimateBitCostsForLiterals(ringbuffer, position, numBytes, ringBufferMask,
		m.literalHistograms[:], literalCosts[1:])

	// Build cumulative literal costs with Kahan summation for float32 accuracy.
	literalCosts[0] = 0.0
	var literalCarry float32
	for i := range numBytes {
		literalCarry += literalCosts[i+1]
		literalCosts[i+1] = literalCosts[i] + literalCarry
		literalCarry -= literalCosts[i+1] - literalCosts[i]
	}

	for i := range costCmd {
		costCmd[i] = float32(fastLog2(11 + i))
	}
	for i := range costDist {
		costDist[i] = float32(fastLog2(20 + i))
	}
	m.minCostCmd = float32(fastLog2(11))
}

// setFromCommands builds histograms from a previous command sequence,
// then derives Shannon bit-costs. Used in Q11's second DP pass.
func (m *zopfliCostModel) setFromCommands(position uint, ringbuffer []byte, ringBufferMask uint, commands []command, lastInsertLen uint) {
	arena := &m.arena

	// Clear histograms.
	arena.histogramLiteral = [core.AlphabetSizeLiteral]uint32{}
	arena.histogramCmd = [core.AlphabetSizeInsertAndCopyLength]uint32{}
	clear(arena.histogramDist[:m.distanceHistogramSize])

	pos := position - lastInsertLen
	for i := range commands {
		insLength := uint(commands[i].insertLen)
		copyLength := uint(commands[i].copyLength())
		distCode := commands[i].distPrefix & 0x3FF
		cmdCode := commands[i].cmdPrefix

		arena.histogramCmd[cmdCode]++
		if cmdCode >= 128 {
			arena.histogramDist[distCode]++
		}

		for j := range insLength {
			arena.histogramLiteral[ringbuffer[(pos+j)&ringBufferMask]]++
		}

		pos += insLength + copyLength
	}

	setCost(arena.histogramLiteral[:], arena.costLiteral[:], core.AlphabetSizeLiteral, true)
	setCost(arena.histogramCmd[:], m.costCmd[:], core.AlphabetSizeInsertAndCopyLength, false)
	setCost(arena.histogramDist[:m.distanceHistogramSize], m.costDist, m.distanceHistogramSize, false)

	minCostCmd := infinity
	for i := range m.costCmd {
		if m.costCmd[i] < minCostCmd {
			minCostCmd = m.costCmd[i]
		}
	}
	m.minCostCmd = minCostCmd

	// Build cumulative literal costs with Kahan summation.
	literalCosts := m.literalCosts
	var literalCarry float32
	literalCosts[0] = 0.0
	for i := range m.numBytes {
		literalCarry += arena.costLiteral[ringbuffer[(position+i)&ringBufferMask]]
		literalCosts[i+1] = literalCosts[i] + literalCarry
		literalCarry -= literalCosts[i+1] - literalCosts[i]
	}
}

// commandCost returns the estimated bit cost for the given command symbol.
func (m *zopfliCostModel) commandCost(cmdCode uint16) float32 {
	return m.costCmd[cmdCode]
}

// distanceCost returns the estimated bit cost for the given distance symbol.
func (m *zopfliCostModel) distanceCost(distCode uint) float32 {
	return m.costDist[distCode]
}

// getLiteralCosts returns the cumulative literal cost for positions [from, to).
func (m *zopfliCostModel) getLiteralCosts(from, to uint) float32 {
	return m.literalCosts[to] - m.literalCosts[from]
}

// getMinCostCmd returns the cached minimum cost across all command codes.
func (m *zopfliCostModel) getMinCostCmd() float32 {
	return m.minCostCmd
}

// setCost converts a histogram to Shannon bit-costs.
// Missing symbols get fastLog2(missingSym) + 2 as their cost.
//
// For literal histograms (isLiteral=true), missing symbols do not
// contribute to the missing symbol count (they are never emitted).
func setCost(histogram []uint32, cost []float32, histogramSize uint, isLiteral bool) {
	sum := uint(0)
	for i := range histogramSize {
		sum += uint(histogram[i])
	}
	log2sum := float32(fastLog2(int(sum)))

	missingSymbolSum := sum
	if !isLiteral {
		for i := range histogramSize {
			if histogram[i] == 0 {
				missingSymbolSum++
			}
		}
	}
	missingSymbolCost := float32(fastLog2(int(missingSymbolSum))) + 2

	for i := range histogramSize {
		if histogram[i] == 0 {
			cost[i] = missingSymbolCost
			continue
		}
		// Shannon bits for this symbol.
		cost[i] = log2sum - float32(fastLog2(int(histogram[i])))
		// Cannot be coded with less than 1 bit.
		if cost[i] < 1 {
			cost[i] = 1
		}
	}
}
