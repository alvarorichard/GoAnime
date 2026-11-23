// Literal context model selection for the greedy block-splitting path (Q5).
//
// Brotli computes a 6-bit context ID (0–63) for each literal from the two
// preceding bytes. A context map groups these 64 IDs into a smaller number
// of histogram clusters, so each cluster gets its own Huffman tree.
//
// This file implements "static" context clustering: the encoder picks one of
// three hardcoded lookup tables and applies the same clustering to every
// block type. At Q6+ the C reference instead discovers per-block-type
// clustering from the data via histogram merging (BrotliClusterHistograms).
//
// The three static context maps are:
//
//   - staticContextMapSimpleUTF8 (2 clusters): distinguishes ASCII from
//     non-ASCII continuation bytes. Good for typical text.
//   - staticContextMapContinuation (3 clusters): separates UTF-8 lead
//     bytes, continuation bytes, and ASCII. Better when multi-byte
//     sequences are common.
//   - staticContextMapComplexUTF8 (13 clusters): fine-grained
//     classification for large (>= 1 MB) inputs where the overhead of 13
//     histograms per block type is justified by the entropy savings.

package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// maxStaticContexts is the maximum number of literal context clusters
// produced by any static context map (the complex UTF-8 map uses 13).
const maxStaticContexts = 13

// staticContextMapSimpleUTF8 maps 64 context IDs to 2 clusters.
// Contexts 2 and 3 (UTF-8 lead/continuation bytes where p1 is non-ASCII)
// map to cluster 1; all others map to cluster 0.
var staticContextMapSimpleUTF8 = [64]uint32{
	0, 0, 1, 1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
}

// staticContextMapContinuation maps 64 context IDs to 3 clusters.
// Cluster 1 = lead bytes (context 0–1), cluster 2 = continuation bytes
// (context 2–3), cluster 0 = ASCII.
var staticContextMapContinuation = [64]uint32{
	1, 1, 2, 2, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
	0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
}

// staticContextMapComplexUTF8 maps 64 context IDs to 13 clusters using a
// fine-grained character-class scheme. Used only for large inputs (>= 1 MB)
// where the extra histograms are worth the improved entropy coding.
//
// Cluster assignments by character class:
//
//	0 special      → clusters 11, 12
//	4 lf           → cluster 0
//	8 space        → clusters 1, 9
//	! after-space  → cluster 2
//	" quote        → cluster 1
//	% punct        → clusters 8, 3
//	({[ open       → cluster 1
//	}]) close      → cluster 2
//	:; colon       → clusters 8, 4
//	. period       → clusters 8, 7, 4
//	> angle        → clusters 8, 0
//	[0-9] digits   → cluster 3
//	[A-Z] upper    → clusters 5, 10
//	[a-z] lower    → cluster 6
var staticContextMapComplexUTF8 = [64]uint32{
	11, 11, 12, 12,
	0, 0, 0, 0,
	1, 1, 9, 9,
	2, 2, 2, 2,
	1, 1, 1, 1,
	8, 3, 3, 3,
	1, 1, 1, 1,
	2, 2, 2, 2,
	8, 4, 4, 4,
	8, 7, 4, 4,
	8, 0, 0, 0,
	3, 3, 3, 3,
	5, 5, 10, 5,
	5, 5, 10, 5,
	6, 6, 6, 6,
	6, 6, 6, 6,
}

// decideOverLiteralContextModeling analyzes the input data and returns the
// number of literal context clusters and the static context map to use.
//
// Returns numContexts=1 and nil map when context modeling is not beneficial
// (short input, low quality, or insufficient entropy gain). Otherwise returns
// 2, 3, or 13 contexts with the corresponding static map.
func decideOverLiteralContextModeling(
	data []byte, pos, mask, length uint,
	quality int, sizeHint uint,
) (numContexts uint, literalContextMap []uint32) {
	numContexts = 1

	if quality < 5 || length < 64 {
		return
	}

	if shouldUseComplexStaticContextMap(
		data, pos, mask, length, sizeHint,
		&numContexts, &literalContextMap,
	) {
		return
	}

	// Gather bigram data of the UTF-8 byte prefixes. To make the analysis
	// faster we only examine 64-byte strides at every 4 KB interval.
	var bigramHisto [9]uint32
	endPos := pos + length
	for ; pos+64 <= endPos; pos += 4096 {
		strideEnd := pos + 64
		prev := byteCategory(data[pos&mask]) * 3
		for p := pos + 1; p < strideEnd; p++ {
			literal := data[p&mask]
			bigramHisto[prev+byteCategory(literal)]++
			prev = byteCategory(literal) * 3
		}
	}

	chooseContextMap(quality, &bigramHisto, &numContexts, &literalContextMap)
	return
}

// byteCategory classifies a byte into one of three categories for bigram
// analysis: 0 = control/low ASCII (0x00–0x3F), 1 = high ASCII (0x40–0xBF),
// 2 = high bytes (0xC0–0xFF).
func byteCategory(b byte) uint32 {
	// Equivalent to the C code: lut[b >> 6] where lut = {0, 0, 1, 2}.
	return [4]uint32{0, 0, 1, 2}[b>>6]
}

// chooseContextMap selects the best static context map based on the entropy
// of a 3x3 bigram histogram. The histogram counts transitions between three
// byte categories (low ASCII, high ASCII, high bytes).
func chooseContextMap(
	quality int,
	bigramHisto *[9]uint32,
	numContexts *uint,
	literalContextMap *[]uint32,
) {
	// Aggregate into monogram (1-context) and two-prefix (2-context) histograms.
	var monogram [3]uint32
	var twoPrefix [6]uint32
	for i := range 9 {
		monogram[i%3] += bigramHisto[i]
		twoPrefix[i%6] += bigramHisto[i]
	}

	// Compute entropy per symbol for 1, 2, and 3 context models.
	var entropy [4]float64
	entropy[1] = estimateEntropy(monogram[:])
	entropy[2] = estimateEntropy(twoPrefix[:3]) + estimateEntropy(twoPrefix[3:6])
	for i := range 3 {
		entropy[3] += estimateEntropy(bigramHisto[3*i : 3*i+3])
	}

	total := monogram[0] + monogram[1] + monogram[2]
	invTotal := 1.0 / float64(total)
	entropy[1] *= invTotal
	entropy[2] *= invTotal
	entropy[3] *= invTotal

	if quality < 7 {
		// 3-context model is slower to decode; penalize it at lower qualities.
		entropy[3] = entropy[1] * 10
	}

	// If expected savings per symbol are less than 0.2 bits, skip context
	// modeling — the faster decoding speed is more valuable.
	switch {
	case entropy[1]-entropy[2] < 0.2 && entropy[1]-entropy[3] < 0.2:
		*numContexts = 1
	case entropy[2]-entropy[3] < 0.02:
		*numContexts = 2
		*literalContextMap = staticContextMapSimpleUTF8[:]
	default:
		*numContexts = 3
		*literalContextMap = staticContextMapContinuation[:]
	}
}

// shouldUseComplexStaticContextMap checks whether the 13-cluster complex
// context map provides sufficient entropy improvement over no context
// modeling. Only considered for large inputs (sizeHint >= 1 MB).
//
// The analysis samples 64-byte strides at 4 KB intervals. Within each stride
// it computes UTF-8 context IDs and builds per-context histograms of the top
// 5 bits of each literal byte.
func shouldUseComplexStaticContextMap(
	data []byte, startPos, mask, length, sizeHint uint,
	numContexts *uint,
	literalContextMap *[]uint32,
) bool {
	if sizeHint < 1<<20 {
		return false
	}

	endPos := startPos + length
	utf8LUT := core.ContextUTF8 << 9

	// Histograms over the top 5 bits of literal bytes:
	// combined[0..31]  = single (no-context) histogram
	// context[ctx*32..(ctx+1)*32-1] = per-context histograms (13 contexts)
	var combined [32]uint32
	var contextHisto [maxStaticContexts * 32]uint32
	var total uint32

	for pos := startPos; pos+64 <= endPos; pos += 4096 {
		prev2 := data[pos&mask]
		prev1 := data[(pos+1)&mask]
		for p := pos + 2; p < pos+64; p++ {
			literal := data[p&mask]
			ctx := staticContextMapComplexUTF8[core.ContextLookupTable[utf8LUT+int(prev1)]|core.ContextLookupTable[utf8LUT+256+int(prev2)]]
			total++
			combined[literal>>3]++
			contextHisto[ctx*32+uint32(literal>>3)]++
			prev2 = prev1
			prev1 = literal
		}
	}

	entropy1 := estimateEntropy(combined[:])
	var entropy2 float64
	for i := range maxStaticContexts {
		entropy2 += estimateEntropy(contextHisto[i*32 : (i+1)*32])
	}

	invTotal := 1.0 / float64(total)
	entropy1 = float64(entropy1 * invTotal)
	entropy2 = float64(entropy2 * invTotal)

	// Heuristic: skip complex modeling if the contextualized entropy is still
	// above 3.0 bits per 5-bit symbol (poorly compressible) or if the per-symbol
	// savings are less than 0.2 bits.
	if entropy2 > 3.0 || entropy1-entropy2 < 0.2 {
		return false
	}

	*numContexts = maxStaticContexts
	*literalContextMap = staticContextMapComplexUTF8[:]
	return true
}

// estimateEntropy computes the Shannon entropy (total bit cost) of a
// population histogram. Unlike bitsEntropy, this does not clamp to a
// minimum of one bit per symbol — it is used for relative comparisons
// between models, not for actual encoding cost estimates.
func estimateEntropy(population []uint32) float64 {
	var total uint32
	var result float64
	for _, p := range population {
		total += p
		result += float64(float64(p) * fastLog2(int(p)))
	}
	return float64(float64(total)*fastLog2(int(total))) - result
}
