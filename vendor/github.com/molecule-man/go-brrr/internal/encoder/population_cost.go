package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// populationCost estimates the bit cost of encoding a histogram using Huffman
// coding, including tree overhead and zero-run-length coding.
//
// Unlike bitsEntropy (which computes pure Shannon entropy), populationCost
// accounts for:
//   - Fixed costs for trivial histograms (1–4 distinct symbols)
//   - Huffman tree encoding overhead (code-length-code histogram)
//   - Zero-run-length encoding via repeat-zero code 17 (RFC 7932 Section 3.5)
//
// histogram is a flat symbol frequency array; dataSize is the number of symbols
// in the alphabet (len(histogram) must be >= dataSize).
func populationCost(histogram []uint32, dataSize int) float64 {
	const (
		oneSymbolHistogramCost   = 12
		twoSymbolHistogramCost   = 20
		threeSymbolHistogramCost = 28
		fourSymbolHistogramCost  = 37
	)

	var totalCount uint32
	for i := range dataSize {
		totalCount += histogram[i]
	}
	if totalCount == 0 {
		return oneSymbolHistogramCost
	}

	// Count distinct non-zero symbols, tracking up to 5.
	var count int
	var s [5]int
	for i := range dataSize {
		if histogram[i] > 0 {
			s[count] = i
			count++
			if count > 4 {
				break
			}
		}
	}

	if count == 1 {
		return oneSymbolHistogramCost
	}
	if count == 2 {
		return twoSymbolHistogramCost + float64(totalCount)
	}
	if count == 3 {
		histo0 := histogram[s[0]]
		histo1 := histogram[s[1]]
		histo2 := histogram[s[2]]
		histoMax := max(histo0, max(histo1, histo2))
		return threeSymbolHistogramCost +
			float64(2*(histo0+histo1+histo2)-histoMax)
	}
	if count == 4 {
		histo := [4]uint32{histogram[s[0]], histogram[s[1]], histogram[s[2]], histogram[s[3]]}
		// Sort descending.
		for i := range 4 {
			for j := i + 1; j < 4; j++ {
				if histo[j] > histo[i] {
					histo[j], histo[i] = histo[i], histo[j]
				}
			}
		}
		h23 := histo[2] + histo[3]
		histoMax := max(h23, histo[0])
		return fourSymbolHistogramCost +
			float64(3*h23+2*(histo[0]+histo[1])-histoMax)
	}

	// General case: compute Shannon entropy of the histogram and simultaneously
	// build a simplified histogram of the code length codes. We use the zero
	// repeat code 17 (alphabetSizeRepeatZeroCodeLength) but not the non-zero
	// repeat code 16.
	var bits float64
	maxDepth := 1
	var depthHisto [core.AlphabetSizeCodeLengths]uint32
	log2total := fastLog2(int(totalCount))

	for i := 0; i < dataSize; {
		if histogram[i] > 0 {
			// Approximate bit cost: -log2(count/total) = log2(total) - log2(count).
			log2p := log2total - fastLog2(int(histogram[i]))
			// Approximate bit depth by rounding.
			depth := int(log2p + 0.5)
			bits += float64(float64(histogram[i]) * log2p)
			if depth > 15 {
				depth = 15
			}
			if depth > maxDepth {
				maxDepth = depth
			}
			depthHisto[depth]++
			i++
		} else {
			// Count run of zeros and encode via zero-run-length codes.
			reps := uint32(1)
			for k := i + 1; k < dataSize && histogram[k] == 0; k++ {
				reps++
			}
			i += int(reps)
			if i == dataSize {
				// Trailing zeros are implicit; no cost.
				break
			}
			if reps < 3 {
				depthHisto[0] += reps
			} else {
				reps -= 2
				for reps > 0 {
					depthHisto[alphabetSizeRepeatZeroCodeLength]++
					// 3 extra bits per repeat-zero code 17.
					bits += 3
					reps >>= 3
				}
			}
		}
	}

	// Estimated encoding cost of the code length code histogram.
	bits += float64(18 + 2*maxDepth)
	// Entropy of the code length code histogram.
	bits += bitsEntropy(depthHisto[:])
	return bits
}
