package encoder

// Histogram clustering: merges similar histograms to produce a compact context
// map.
//
// Histogram data is stored in flat []uint32 slices indexed by
// [histIdx * alphabetSize + symbol], with parallel arrays for totalCount and
// bitCost.

// histogramPair tracks the cost of merging two histogram clusters.
type clusterBufs struct {
	chClusterSize []uint32
	chClusters    []uint32
	chBitCosts    []float64
	chTotalCounts []uint32
	chSymbols     []uint32
	chTmpHist     []uint32
	chPairs       []histogramPair

	hrNewIndex    []uint32
	hrTmpData     []uint32
	hrTmpBitCosts []float64
	hrTmpTotals   []uint32
}

type histogramPair struct {
	idx1, idx2 uint32
	costCombo  float64 // combined histogram cost
	costDiff   float64 // net cost change from merging (negative = good)
}

// histogramPairIsLess returns true if p1 should rank below p2 in the max-heap
// (i.e., p1 is a worse merge candidate than p2). The heap root (pairs[0])
// holds the best merge — the one with the most negative costDiff.
//
// Tie-breaking by index spread ensures deterministic merge order.
func histogramPairIsLess(p1, p2 *histogramPair) bool {
	if p1.costDiff != p2.costDiff {
		return p1.costDiff > p2.costDiff
	}
	return (p1.idx2 - p1.idx1) > (p2.idx2 - p2.idx1)
}

// clusterCostDiff computes the entropy reduction of the context map when we
// combine two clusters of sizes sizeA and sizeB.
func clusterCostDiff(sizeA, sizeB uint32) float64 {
	sizeC := sizeA + sizeB
	return float64(float64(sizeA)*fastLog2(int(sizeA))) +
		float64(float64(sizeB)*fastLog2(int(sizeB))) -
		float64(float64(sizeC)*fastLog2(int(sizeC)))
}

// histogramSlice returns the sub-slice of the flat histogram array for index idx.
func histogramSlice(data []uint32, idx, alphabetSize int) []uint32 {
	off := idx * alphabetSize
	return data[off : off+alphabetSize]
}

// histogramClear zeroes a histogram.
func histogramClear(h []uint32, alphabetSize int) {
	clear(h[:alphabetSize])
}

// histogramCopy copies src histogram into dst.
func histogramCopy(dst, src []uint32, alphabetSize int) {
	copy(dst[:alphabetSize], src[:alphabetSize])
}

// compareAndPushToQueue evaluates merging two histograms and conditionally adds
// the pair to the priority queue.
//
// The queue is a loose max-heap where only pairs[0] is guaranteed to be the
// maximum (best merge). New pairs that beat the current best replace pairs[0];
// otherwise they are appended.
func compareAndPushToQueue(
	out []uint32, alphabetSize int,
	bitCosts []float64, totalCounts []uint32,
	clusterSize []uint32,
	idx1, idx2 uint32,
	maxNumPairs int,
	pairs []histogramPair,
	numPairs *int,
	tmpHist []uint32,
) {
	if idx1 == idx2 {
		return
	}
	if idx2 < idx1 {
		idx1, idx2 = idx2, idx1
	}

	isGoodPair := false
	p := histogramPair{idx1: idx1, idx2: idx2}
	p.costDiff = 0.5 * clusterCostDiff(clusterSize[idx1], clusterSize[idx2])
	p.costDiff -= bitCosts[idx1]
	p.costDiff -= bitCosts[idx2]

	switch {
	case totalCounts[idx1] == 0:
		p.costCombo = bitCosts[idx2]
		isGoodPair = true
	case totalCounts[idx2] == 0:
		p.costCombo = bitCosts[idx1]
		isGoodPair = true
	default:
		threshold := 1e99
		if *numPairs > 0 {
			threshold = max(0.0, pairs[0].costDiff)
		}
		// Build combined histogram in tmpHist.
		histogramCopy(tmpHist, histogramSlice(out, int(idx1), alphabetSize), alphabetSize)
		histogramAdd(tmpHist, histogramSlice(out, int(idx2), alphabetSize), alphabetSize)
		costCombo := populationCost(tmpHist, alphabetSize)
		if costCombo < threshold-p.costDiff {
			p.costCombo = costCombo
			isGoodPair = true
		}
	}

	if isGoodPair {
		p.costDiff += p.costCombo
		if *numPairs > 0 && histogramPairIsLess(&pairs[0], &p) {
			// Replace the top of the queue if needed.
			if *numPairs < maxNumPairs {
				pairs[*numPairs] = pairs[0]
				*numPairs++
			}
			pairs[0] = p
		} else if *numPairs < maxNumPairs {
			pairs[*numPairs] = p
			*numPairs++
		}
	}
}

// histogramCombine iteratively merges the best histogram pairs until the
// cluster count drops to maxClusters.
//
// The function maintains a loose max-heap of merge candidates. In each
// iteration it takes the best pair from pairs[0], merges the two clusters,
// removes stale pairs, and pushes new pairs formed with the merged cluster.
func histogramCombine(
	out []uint32, alphabetSize int,
	bitCosts []float64, totalCounts []uint32,
	clusterSize, symbols, clusters []uint32,
	pairs []histogramPair,
	numClusters, symbolsSize, maxClusters, maxNumPairs int,
	tmpHist []uint32,
) int {
	costDiffThreshold := 0.0
	minClusterSize := 1
	numPairs := 0

	// Build initial pair queue.
	for idx1 := 0; idx1 < numClusters; idx1++ {
		for idx2 := idx1 + 1; idx2 < numClusters; idx2++ {
			compareAndPushToQueue(out, alphabetSize,
				bitCosts, totalCounts, clusterSize,
				clusters[idx1], clusters[idx2],
				maxNumPairs, pairs, &numPairs, tmpHist)
		}
	}

	for numClusters > minClusterSize {
		if pairs[0].costDiff >= costDiffThreshold {
			costDiffThreshold = 1e99
			minClusterSize = maxClusters
			continue
		}

		// Take the best pair.
		bestIdx1 := pairs[0].idx1
		bestIdx2 := pairs[0].idx2

		// Merge idx2 into idx1.
		histogramAdd(
			histogramSlice(out, int(bestIdx1), alphabetSize),
			histogramSlice(out, int(bestIdx2), alphabetSize),
			alphabetSize,
		)
		bitCosts[bestIdx1] = pairs[0].costCombo
		totalCounts[bestIdx1] += totalCounts[bestIdx2]
		clusterSize[bestIdx1] += clusterSize[bestIdx2]

		// Redirect symbols from idx2 to idx1.
		for i := range symbolsSize {
			if symbols[i] == bestIdx2 {
				symbols[i] = bestIdx1
			}
		}

		// Remove bestIdx2 from clusters list.
		for i := 0; i < numClusters; i++ {
			if clusters[i] == bestIdx2 {
				copy(clusters[i:numClusters-1], clusters[i+1:numClusters])
				break
			}
		}
		numClusters--

		// Remove pairs intersecting the merged pair.
		copyToIdx := 0
		for i := 0; i < numPairs; i++ {
			p := &pairs[i]
			if p.idx1 == bestIdx1 || p.idx2 == bestIdx1 ||
				p.idx1 == bestIdx2 || p.idx2 == bestIdx2 {
				continue
			}
			if histogramPairIsLess(&pairs[0], p) {
				front := pairs[0]
				pairs[0] = *p
				pairs[copyToIdx] = front
			} else {
				pairs[copyToIdx] = *p
			}
			copyToIdx++
		}
		numPairs = copyToIdx

		// Push new pairs formed with the merged cluster.
		for i := 0; i < numClusters; i++ {
			compareAndPushToQueue(out, alphabetSize,
				bitCosts, totalCounts, clusterSize,
				bestIdx1, clusters[i],
				maxNumPairs, pairs, &numPairs, tmpHist)
		}
	}
	return numClusters
}

// histogramBitCostDistance computes the bit cost of assigning one histogram
// to another's cluster. Returns the cost increase from adding histogram to
// candidate.
func histogramBitCostDistance(
	histogram, candidate, tmp []uint32,
	alphabetSize int,
	candidateBitCost float64,
	candidateTotalCount uint32,
) float64 {
	_ = candidateTotalCount // used only in C struct; not needed here
	if histogramTotalCount(histogram, alphabetSize) == 0 {
		return 0.0
	}
	histogramCopy(tmp, histogram, alphabetSize)
	histogramAdd(tmp, candidate, alphabetSize)
	return populationCost(tmp, alphabetSize) - candidateBitCost
}

// histogramRemap reassigns each input histogram to its best output cluster.
//
// For each input histogram, we find the output cluster with minimum assignment
// cost, then recompute each output cluster by summing all inputs assigned to it.
func histogramRemap(
	in []uint32, inSize int,
	clusters []uint32, numClusters int,
	out []uint32, alphabetSize int,
	bitCosts []float64, totalCounts []uint32,
	symbols []uint32,
	tmpHist []uint32,
) {
	for i := range inSize {
		var bestOut uint32
		if i == 0 {
			bestOut = symbols[0]
		} else {
			bestOut = symbols[i-1]
		}
		bestBits := histogramBitCostDistance(
			histogramSlice(in, i, alphabetSize),
			histogramSlice(out, int(bestOut), alphabetSize),
			tmpHist, alphabetSize,
			bitCosts[bestOut], totalCounts[bestOut],
		)
		for j := range numClusters {
			curBits := histogramBitCostDistance(
				histogramSlice(in, i, alphabetSize),
				histogramSlice(out, int(clusters[j]), alphabetSize),
				tmpHist, alphabetSize,
				bitCosts[clusters[j]], totalCounts[clusters[j]],
			)
			if curBits < bestBits {
				bestBits = curBits
				bestOut = clusters[j]
			}
		}
		symbols[i] = bestOut
	}

	// Recompute each output cluster from raw inputs and symbols.
	for i := range numClusters {
		histogramClear(histogramSlice(out, int(clusters[i]), alphabetSize), alphabetSize)
		totalCounts[clusters[i]] = 0
	}
	for i := range inSize {
		histogramAdd(
			histogramSlice(out, int(symbols[i]), alphabetSize),
			histogramSlice(in, i, alphabetSize),
			alphabetSize,
		)
		totalCounts[symbols[i]] += histogramTotalCount(histogramSlice(in, i, alphabetSize), alphabetSize)
	}
	// Recompute bit costs for active clusters.
	for i := range numClusters {
		bitCosts[clusters[i]] = populationCost(
			histogramSlice(out, int(clusters[i]), alphabetSize), alphabetSize)
	}
}

// histogramReindex renumbers output histograms to consecutive indices [0, N).
//
// On input, symbols[] contains arbitrary indices into out[]. On return,
// symbols'[i] = f(symbols[i]) where f maps the old indices to [0, N), and
// out'[symbols'[i]] = out[symbols[i]].
//
// Returns N, the number of unique output histograms.
func histogramReindex(
	out []uint32, alphabetSize int,
	bitCosts []float64, totalCounts []uint32,
	symbols []uint32, length int,
	bufs *clusterBufs,
) int {
	const invalidIndex = ^uint32(0)
	bufs.hrNewIndex = growUint32(bufs.hrNewIndex, length)
	newIndex := bufs.hrNewIndex[:length]
	for i := range newIndex {
		newIndex[i] = invalidIndex
	}

	var nextIndex uint32
	for i := range length {
		if newIndex[symbols[i]] == invalidIndex {
			newIndex[symbols[i]] = nextIndex
			nextIndex++
		}
	}

	// Reorder out histograms into consecutive positions.
	bufs.hrTmpData = growUint32(bufs.hrTmpData, int(nextIndex)*alphabetSize)
	tmpData := bufs.hrTmpData[:int(nextIndex)*alphabetSize]
	bufs.hrTmpBitCosts = growFloat64(bufs.hrTmpBitCosts, int(nextIndex))
	tmpBitCosts := bufs.hrTmpBitCosts[:nextIndex]
	bufs.hrTmpTotals = growUint32(bufs.hrTmpTotals, int(nextIndex))
	tmpTotalCounts := bufs.hrTmpTotals[:nextIndex]
	ni := uint32(0)
	for i := range length {
		if newIndex[symbols[i]] == ni {
			histogramCopy(
				histogramSlice(tmpData, int(ni), alphabetSize),
				histogramSlice(out, int(symbols[i]), alphabetSize),
				alphabetSize,
			)
			tmpBitCosts[ni] = bitCosts[symbols[i]]
			tmpTotalCounts[ni] = totalCounts[symbols[i]]
			ni++
		}
		symbols[i] = newIndex[symbols[i]]
	}

	for i := uint32(0); i < nextIndex; i++ {
		histogramCopy(
			histogramSlice(out, int(i), alphabetSize),
			histogramSlice(tmpData, int(i), alphabetSize),
			alphabetSize,
		)
		bitCosts[i] = tmpBitCosts[i]
		totalCounts[i] = tmpTotalCounts[i]
	}
	return int(nextIndex)
}

// clusterHistograms merges similar histograms and produces a mapping from input
// histogram indices to output cluster indices.
//
// The algorithm operates in two passes:
//  1. Process inputs in batches of 64, clustering each batch independently
//  2. Combine batch results into a global clustering with at most maxHistograms
//     output clusters
//
// Finally, histogramRemap reassigns each input to its best cluster, and
// histogramReindex renumbers the output for consecutive indexing.
//
// in:  flat histogram data [inSize][alphabetSize]
// out: flat output buffer, must be at least [inSize][alphabetSize]
// maxHistograms: target cluster count (e.g. 256)
// Returns (outSize, symbols) where outSize is the number of output histograms
// and symbols maps each input histogram to its output index.
func clusterHistograms(
	in []uint32,
	inSize, alphabetSize int,
	maxHistograms int,
	out []uint32,
	bufs *clusterBufs,
) (outSize int, symbols []uint32) {
	bufs.chClusterSize = growUint32(bufs.chClusterSize, inSize)
	clusterSize := bufs.chClusterSize[:inSize]
	bufs.chClusters = growUint32(bufs.chClusters, inSize)
	clusters := bufs.chClusters[:inSize]
	bufs.chBitCosts = growFloat64(bufs.chBitCosts, inSize)
	bitCosts := bufs.chBitCosts[:inSize]
	bufs.chTotalCounts = growUint32(bufs.chTotalCounts, inSize)
	totalCounts := bufs.chTotalCounts[:inSize]
	bufs.chSymbols = growUint32(bufs.chSymbols, inSize)
	symbols = bufs.chSymbols[:inSize]
	bufs.chTmpHist = growUint32(bufs.chTmpHist, alphabetSize)
	tmpHist := bufs.chTmpHist[:alphabetSize]

	const maxInputHistograms = 64
	pairsCapacity := maxInputHistograms * maxInputHistograms / 2
	bufs.chPairs = growHistogramPairs(bufs.chPairs, pairsCapacity+1)
	pairs := bufs.chPairs

	fillSlice(clusterSize[:inSize], 1)

	// Initialize output histograms and compute initial bit costs.
	for i := range inSize {
		histogramCopy(
			histogramSlice(out, i, alphabetSize),
			histogramSlice(in, i, alphabetSize),
			alphabetSize,
		)
		bitCosts[i] = populationCost(histogramSlice(in, i, alphabetSize), alphabetSize)
		totalCounts[i] = histogramTotalCount(histogramSlice(in, i, alphabetSize), alphabetSize)
		symbols[i] = uint32(i)
	}

	// First pass: cluster in batches of maxInputHistograms.
	numClusters := 0
	for i := 0; i < inSize; i += maxInputHistograms {
		numToCombine := min(inSize-i, maxInputHistograms)
		for j := range numToCombine {
			clusters[numClusters+j] = uint32(i + j)
		}
		numNewClusters := histogramCombine(
			out, alphabetSize,
			bitCosts, totalCounts,
			clusterSize, symbols[i:], clusters[numClusters:],
			pairs,
			numToCombine, numToCombine, maxHistograms, pairsCapacity,
			tmpHist,
		)
		numClusters += numNewClusters
	}

	// Second pass: combine all batch results. Limit total pairs to avoid
	// quadratic blowup.
	maxNumPairs := min(64*numClusters, (numClusters/2)*numClusters)
	if maxNumPairs+1 > len(pairs) {
		bufs.chPairs = growHistogramPairs(bufs.chPairs, maxNumPairs+1)
		pairs = bufs.chPairs
	}
	numClusters = histogramCombine(
		out, alphabetSize,
		bitCosts, totalCounts,
		clusterSize, symbols, clusters,
		pairs,
		numClusters, inSize, maxHistograms, maxNumPairs,
		tmpHist,
	)

	// Find the optimal map from original histograms to the final clusters.
	histogramRemap(in, inSize, clusters, numClusters,
		out, alphabetSize, bitCosts, totalCounts, symbols, tmpHist)

	// Convert the context map to a canonical form with consecutive indices.
	outSize = histogramReindex(out, alphabetSize, bitCosts, totalCounts, symbols, inSize, bufs)
	return outSize, symbols
}
