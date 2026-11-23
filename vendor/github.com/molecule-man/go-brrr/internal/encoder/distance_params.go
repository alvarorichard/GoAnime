package encoder

import "github.com/molecule-man/go-brrr/internal/core"

// distanceParams holds the distance encoding configuration for a metablock.
//
// Brotli encodes distances using two tunable parameters (RFC 7932 Section 4):
//   - postfixBits (NPOSTFIX, 0–3): number of low-order bits used as postfix
//   - numDirectCodes (NDIRECT): number of direct distance codes
//
// Different parameter combinations produce different distance alphabet sizes
// and encoding efficiencies. The Q10 encoder searches over all valid
// (postfixBits, numDirectCodes) combinations per metablock to minimize total
// distance entropy.
type distanceParams struct {
	postfixBits       uint32 // NPOSTFIX (0–3)
	numDirectCodes    uint32 // NDIRECT
	alphabetSizeMax   uint32 // full distance alphabet size
	alphabetSizeLimit uint32 // effective limit (may differ for large window)
	maxDistance       uint32 // maximum encodable distance
}

// initDistanceParams computes derived distance fields from npostfix and ndirect.
//
// The distance alphabet size and maximum encodable distance are derived from
// the parameter combination. Large-window mode is not yet supported.
func initDistanceParams(npostfix, ndirect uint32) distanceParams {
	alphabetSizeMax := distanceAlphabetSize(npostfix, ndirect, core.MaxDistanceBits)
	maxDistance := ndirect + (1 << (core.MaxDistanceBits + npostfix + 2)) -
		(1 << (npostfix + 2))
	// TODO: add large-window support (BROTLI_LARGE_MAX_DISTANCE_BITS = 62).
	return distanceParams{
		postfixBits:       npostfix,
		numDirectCodes:    ndirect,
		alphabetSizeMax:   alphabetSizeMax,
		alphabetSizeLimit: alphabetSizeMax,
		maxDistance:       maxDistance,
	}
}

// recomputeDistancePrefixes re-encodes distance prefix codes on all commands
// when distance parameters change.
//
// For each command that carries a distance (copy length > 0 and cmdPrefix >= 128),
// the original distance code is reconstructed using origParams, then re-encoded
// with newParams. The command's distPrefix and distExtra fields are updated
// in place.
func recomputeDistancePrefixes(cmds []command, origParams, newParams distanceParams) {
	if origParams.postfixBits == newParams.postfixBits &&
		origParams.numDirectCodes == newParams.numDirectCodes {
		return
	}

	for i := range cmds {
		cmd := &cmds[i]
		if cmd.copyLength() != 0 && cmd.cmdPrefix >= 128 {
			dist := cmd.distanceCode(uint(origParams.numDirectCodes), uint(origParams.postfixBits))
			cmd.distPrefix, cmd.distExtra = prefixEncodeCopyDistance(
				uint(dist),
				uint(newParams.numDirectCodes),
				uint(newParams.postfixBits),
			)
		}
	}
}

// computeDistanceCost evaluates the entropy cost of encoding the distance
// stream with newParams. Returns (cost, ok). ok is false if any distance
// exceeds newParams.maxDistance.
//
// The function builds a histogram of distance prefix codes under newParams,
// then computes the population cost plus the total extra bits.
func computeDistanceCost(
	cmds []command,
	origParams, newParams distanceParams,
	tmpHist []uint32,
) (float64, bool) {
	equalParams := origParams.postfixBits == newParams.postfixBits &&
		origParams.numDirectCodes == newParams.numDirectCodes

	// Clear the scratch histogram.
	clear(tmpHist)

	var extraBits float64
	for i := range cmds {
		cmd := &cmds[i]
		if cmd.copyLength() != 0 && cmd.cmdPrefix >= 128 {
			var distPrefix uint16
			if equalParams {
				distPrefix = cmd.distPrefix
			} else {
				distance := cmd.distanceCode(uint(origParams.numDirectCodes), uint(origParams.postfixBits))
				if distance > newParams.maxDistance {
					return 0, false
				}
				var dExtra uint32
				distPrefix, dExtra = prefixEncodeCopyDistance(
					uint(distance),
					uint(newParams.numDirectCodes),
					uint(newParams.postfixBits),
				)
				_ = dExtra
			}
			tmpHist[distPrefix&0x3FF]++
			extraBits += float64(distPrefix >> 10)
		}
	}

	cost := populationCost(tmpHist, int(newParams.alphabetSizeLimit)) + extraBits
	return cost, true
}

// distanceAlphabetSize computes the distance alphabet size from the encoding
// parameters per RFC 7932 Section 4:
//
//	alphabetSize = NUM_DISTANCE_SHORT_CODES + NDIRECT + (MAXNBITS << (NPOSTFIX + 1))
func distanceAlphabetSize(npostfix, ndirect, maxnbits uint32) uint32 {
	return core.NumDistanceShortCodes + ndirect + (maxnbits << (npostfix + 1))
}
