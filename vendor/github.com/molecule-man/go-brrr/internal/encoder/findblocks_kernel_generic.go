//go:build !amd64 || purego

package encoder

// findBlocksDPStep adds insertCost into cost and returns the minimum together
// with the index of its first occurrence.
func findBlocksDPStep(cost, insertCost []float64) (minCost float64, best int) {
	minCost = noMinCost
	for k := range cost {
		cost[k] += insertCost[k]
		if cost[k] < minCost {
			minCost = cost[k]
			best = k
		}
	}
	return minCost, best
}

// findBlocksClamp rebases cost against minCost, clamps it at switchCost and
// sets one bit in sig per clamped histogram.
func findBlocksClamp(cost []float64, sig []byte, minCost, switchCost float64) {
	for k := range cost {
		cost[k] -= minCost
		if cost[k] >= switchCost {
			cost[k] = switchCost
			sig[k>>3] |= 1 << (k & 7)
		}
	}
}

// findBlocksStep runs findBlocksDPStep and findBlocksClamp in a single call.
func findBlocksStep(cost, insertCost []float64, sig []byte, switchCost float64) (minCost float64, best int) {
	minCost, best = findBlocksDPStep(cost, insertCost)
	findBlocksClamp(cost, sig, minCost, switchCost)
	return minCost, best
}

func findBlocksDP(data []uint16, insertCost, cost []float64, switchSignal, blockID []byte, blockSwitchBitcost float64) {
	numHistograms := len(cost)
	bitmapLen := (numHistograms + 7) >> 3

	const prologueLength = 2000
	const prologueMultiplier = 0.07 / 2000

	for byteIx, symbol := range data {
		ix := byteIx * bitmapLen
		insertCostIx := int(symbol) * numHistograms
		switchCost := blockSwitchBitcost

		if byteIx < prologueLength {
			switchCost *= 0.77 + float64(prologueMultiplier*float64(byteIx))
		}

		minCost, best := findBlocksStep(cost,
			insertCost[insertCostIx:insertCostIx+numHistograms],
			switchSignal[ix:], switchCost)
		if minCost < noMinCost {
			blockID[byteIx] = byte(best)
		}
	}
}
