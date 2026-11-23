// Literal bit-cost estimation for Zopfli optimal parsing.
//
// Estimates the per-byte cost (in bits) of encoding each literal in a
// byte stream, using a sliding-window frequency model. The cost model
// guides the Zopfli DP toward commands that compress well.
//
// Two strategies are used depending on the input:
//   - UTF-8 data: a 3-histogram model (one per UTF-8 byte position)
//     captures the strong byte-position correlations in multi-byte encodings.
//   - Non-UTF-8 data: a single 256-entry histogram with a sliding window.

package encoder

// estimateBitCostsForLiterals estimates the Shannon bit-cost for each literal
// byte in data[position..position+numBytes), writing results to
// cost[0..numBytes). The costs are used by the Zopfli cost model to decide
// whether a literal or a backward reference is cheaper.
//
// The histogram parameter is scratch space: 3*256 entries for the UTF-8
// path, or 256 entries for the non-UTF-8 path.
func estimateBitCostsForLiterals(data []byte, position, numBytes, ringBufferMask uint, histogram []uint, cost []float32) {
	if isMostlyUTF8(data, position, ringBufferMask, numBytes) {
		estimateBitCostsForLiteralsUTF8(data, position, numBytes, ringBufferMask, histogram, cost)
	} else {
		estimateBitCostsForLiteralsRaw(data, position, numBytes, ringBufferMask, histogram, cost)
	}
}

// utf8Position returns the byte position within a UTF-8 multi-byte sequence
// (0 = start/ASCII, 1 = byte 2, 2 = byte 3), clamped to clamp.
func utf8Position(last, c, clamp uint) uint {
	if c < 128 {
		return 0
	}
	if c >= 192 {
		return min(1, clamp)
	}
	// Continuation byte: check previous byte to decide.
	if last < 0xE0 {
		return 0
	}
	return min(2, clamp)
}

func nonZeroASCIIRun(data []byte, pos, n uint) uint {
	i := uint(0)
	for ; i+8 <= n; i += 8 {
		w := loadU64LE(data, pos+i)
		if w&0x8080808080808080 != 0 || (w-0x0101010101010101)&^w&0x8080808080808080 != 0 {
			break
		}
	}
	for ; i < n; i++ {
		if c := data[pos+i]; c == 0 || c >= 0x80 {
			break
		}
	}
	return i
}

func contiguousLimit(data []byte, mask uint) uint {
	return min(uint(len(data)), mask+1)
}

func decideMultiByteStatsLevel(data []byte, pos, length, mask uint) uint {
	limit := contiguousLimit(data, mask)
	multiByte := uint(0)
	lastC := uint(0)
	for i := uint(0); i < length; {
		p := (pos + i) & mask
		if p < limit {
			if run := nonZeroASCIIRun(data, p, min(length-i, limit-p)); run > 0 {
				lastC = uint(data[p+run-1])
				i += run
				continue
			}
		}
		c := uint(data[p])
		if utf8Position(lastC, c, 2) != 0 {
			multiByte++
			if multiByte >= 25 {
				return 1
			}
		}
		lastC = c
		i++
	}
	return 0
}

// estimateBitCostsForLiteralsUTF8 estimates per-byte costs using a
// sliding-window UTF-8-aware frequency model with up to 3 histograms
// (one per byte position in a multi-byte sequence).
func estimateBitCostsForLiteralsUTF8(data []byte, pos, length, mask uint, histogram []uint, cost []float32) {
	maxUTF8 := decideMultiByteStatsLevel(data, pos, length, mask)
	windowHalf := uint(495)
	inWindow := min(windowHalf, length)
	var inWindowUTF8 [3]uint

	// Clear histograms: 3 * 256 entries.
	clear(histogram[:3*256])

	// Bootstrap histograms from the initial window.
	lastC := uint(0)
	utf8Pos := uint(0)
	for i := range inWindow {
		c := uint(data[(pos+i)&mask])
		histogram[256*utf8Pos+c]++
		inWindowUTF8[utf8Pos]++
		utf8Pos = utf8Position(lastC, c, maxUTF8)
		lastC = c
	}

	// inWindowUTF8 changes at most once per byte and is far above the 256-entry
	// fastLog2 table, so recomputing it every iteration means a math.Log2 call
	// per byte. Cache one logarithm per UTF-8 position instead.
	var cachedUTF8Count [3]uint
	var cachedUTF8Log [3]float64
	for k := range cachedUTF8Count {
		cachedUTF8Count[k] = inWindowUTF8[k]
		cachedUTF8Log[k] = fastLog2(int(inWindowUTF8[k]))
	}

	var addPrev1, addPrev2 uint
	if windowHalf < length {
		addPrev1 = uint(data[(pos+windowHalf-1)&mask])
		addPrev2 = uint(data[(pos+windowHalf-2)&mask])
	}
	var remPrev1, remPrev2 uint
	var curPrev1, curPrev2 uint

	// Compute bit costs with sliding window.
	for i := range length {
		if i >= windowHalf {
			// Remove a byte in the past.
			utf8Pos2 := utf8Position(remPrev2, remPrev1, maxUTF8)
			gone := uint(data[(pos+i-windowHalf)&mask])
			histogram[256*utf8Pos2+gone]--
			inWindowUTF8[utf8Pos2]--
			remPrev2, remPrev1 = remPrev1, gone
		}
		if i+windowHalf < length {
			// Add a byte in the future.
			utf8Pos2 := utf8Position(addPrev2, addPrev1, maxUTF8)
			added := uint(data[(pos+i+windowHalf)&mask])
			histogram[256*utf8Pos2+added]++
			inWindowUTF8[utf8Pos2]++
			addPrev2, addPrev1 = addPrev1, added
		}

		curUTF8Pos := utf8Position(curPrev2, curPrev1, maxUTF8)
		maskedPos := (pos + i) & mask
		cur := uint(data[maskedPos])
		curPrev2, curPrev1 = curPrev1, cur
		histo := histogram[256*curUTF8Pos+cur]
		if cachedUTF8Count[curUTF8Pos] != inWindowUTF8[curUTF8Pos] {
			cachedUTF8Count[curUTF8Pos] = inWindowUTF8[curUTF8Pos]
			cachedUTF8Log[curUTF8Pos] = fastLog2(int(inWindowUTF8[curUTF8Pos]))
		}
		litCost := cachedUTF8Log[curUTF8Pos] - fastLog2(int(histo))
		litCost += 0.02905
		if litCost < 1.0 {
			litCost = litCost*0.5 + 0.5
		}
		// Make the first bytes more expensive to account for the statistical
		// anomaly at the beginning of the data.
		const prologueLength = 2000
		const multiplier = 0.35 / prologueLength
		if i < prologueLength {
			litCost += 0.35 + float64(multiplier*float64(i))
		}
		cost[i] = float32(litCost)
	}
}

// estimateBitCostsForLiteralsRaw estimates per-byte costs using a single
// 256-entry sliding-window frequency model for non-UTF-8 data.
func estimateBitCostsForLiteralsRaw(data []byte, pos, length, mask uint, histogram []uint, cost []float32) {
	windowHalf := uint(2000)
	inWindow := min(windowHalf, length)

	clear(histogram[:256])

	// Bootstrap histogram.
	for i := range inWindow {
		histogram[data[(pos+i)&mask]]++
	}

	// inWindow is invariant once the window is full, and is far above the
	// 256-entry fastLog2 table, so recomputing it every iteration means a
	// math.Log2 call per byte.
	cachedInWindow := inWindow
	cachedLog := fastLog2(int(inWindow))

	// Compute bit costs with sliding window.
	for i := range length {
		if i >= windowHalf {
			histogram[data[(pos+i-windowHalf)&mask]]--
			inWindow--
		}
		if i+windowHalf < length {
			histogram[data[(pos+i+windowHalf)&mask]]++
			inWindow++
		}
		histo := histogram[data[(pos+i)&mask]]
		if inWindow != cachedInWindow {
			cachedInWindow = inWindow
			cachedLog = fastLog2(int(inWindow))
		}
		litCost := cachedLog - fastLog2(int(histo))
		litCost += 0.029
		if litCost < 1.0 {
			litCost = litCost*0.5 + 0.5
		}
		cost[i] = float32(litCost)
	}
}

// isMostlyUTF8 returns true if at least minUTF8Ratio of the data bytes form
// valid UTF-8 sequences.
func isMostlyUTF8(data []byte, pos, mask, length uint) bool {
	limit := contiguousLimit(data, mask)
	sizeUTF8 := uint(0)
	i := uint(0)
	for i < length {
		p := (pos + i) & mask
		if p < limit {
			if run := nonZeroASCIIRun(data, p, min(length-i, limit-p)); run > 0 {
				sizeUTF8 += run
				i += run
				continue
			}
		}
		bytesRead, isUTF8 := parseAsUTF8(data, p, length-i, mask)
		i += bytesRead
		if isUTF8 {
			sizeUTF8 += bytesRead
		}
	}
	return float64(sizeUTF8) > minUTF8Ratio*float64(length)
}

// parseAsUTF8 attempts to parse a UTF-8 sequence starting at data[pos].
// Returns the number of bytes consumed and whether it was valid UTF-8.
func parseAsUTF8(data []byte, pos, remaining, mask uint) (bytesRead uint, valid bool) {
	b0 := data[pos]

	// ASCII
	if b0&0x80 == 0 {
		if b0 > 0 {
			return 1, true
		}
	}

	// 2-byte UTF-8
	if remaining > 1 {
		b1 := data[(pos+1)&mask]
		if b0&0xE0 == 0xC0 && b1&0xC0 == 0x80 {
			symbol := (uint(b0&0x1F) << 6) | uint(b1&0x3F)
			if symbol > 0x7F {
				return 2, true
			}
		}
	}

	// 3-byte UTF-8
	if remaining > 2 {
		b1 := data[(pos+1)&mask]
		b2 := data[(pos+2)&mask]
		if b0&0xF0 == 0xE0 && b1&0xC0 == 0x80 && b2&0xC0 == 0x80 {
			symbol := (uint(b0&0x0F) << 12) | (uint(b1&0x3F) << 6) | uint(b2&0x3F)
			if symbol > 0x7FF {
				return 3, true
			}
		}
	}

	// 4-byte UTF-8
	if remaining > 3 {
		b1 := data[(pos+1)&mask]
		b2 := data[(pos+2)&mask]
		b3 := data[(pos+3)&mask]
		if b0&0xF8 == 0xF0 && b1&0xC0 == 0x80 && b2&0xC0 == 0x80 && b3&0xC0 == 0x80 {
			symbol := (uint(b0&0x07) << 18) | (uint(b1&0x3F) << 12) |
				(uint(b2&0x3F) << 6) | uint(b3&0x3F)
			if symbol > 0xFFFF && symbol <= 0x10FFFF {
				return 4, true
			}
		}
	}

	return 1, false
}
