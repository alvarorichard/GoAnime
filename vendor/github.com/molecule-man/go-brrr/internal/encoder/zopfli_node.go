// Zopfli DP node type and command extraction.
//
// The Zopfli optimal parsing algorithm (Q10/Q11) uses dynamic programming over
// a node array where each node records the best way to reach that position
// from an earlier position. After the forward DP pass, a backward trace
// extracts the optimal command sequence.

package encoder

import (
	"math"

	"github.com/molecule-man/go-brrr/internal/core"
)

// zopfliNode is a single element in the DP array. nodes[i] records the
// cheapest known way to reach position i from some earlier position.
//
// The fields use bit-packing to keep the struct compact (16 bytes), matching
// the C layout:
//   - length: copy length (low 25 bits) | length code modifier (high 7 bits)
//   - distance: backward reference distance
//   - dcodeInsertLength: distance short code+1 (high 5 bits) | insert length (low 27 bits)
//   - u: during forward pass, float32 cost (via Float32bits); during backtrace,
//     uint32 next offset or shortcut pointer
type zopfliNode struct {
	length            uint32
	distance          uint32
	dcodeInsertLength uint32
	u                 uint32
}

// copyLength returns the copy length stored in the low 25 bits.
func (n *zopfliNode) copyLength() uint32 {
	return n.length & 0x1FFFFFF
}

// lengthCode returns the effective length code, adjusting for the modifier
// stored in the high 7 bits.
func (n *zopfliNode) lengthCode() uint32 {
	modifier := n.length >> 25
	return n.copyLength() + 9 - modifier
}

// copyDistance returns the backward reference distance.
func (n *zopfliNode) copyDistance() uint32 {
	return n.distance
}

// distanceCode returns the distance code. If a distance short code was used
// (stored as shortCode+1 in the high 5 bits), it returns shortCode-1.
// Otherwise it returns distance + core.NumDistanceShortCodes - 1.
func (n *zopfliNode) distanceCode() uint32 {
	shortCode := n.dcodeInsertLength >> 27
	if shortCode == 0 {
		return n.copyDistance() + core.NumDistanceShortCodes - 1
	}
	return shortCode - 1
}

// commandLength returns the total command length (insert + copy).
func (n *zopfliNode) commandLength() uint32 {
	return n.copyLength() + (n.dcodeInsertLength & 0x7FFFFFF)
}

// cost interprets the u field as a float32 cost (forward DP pass).
func (n *zopfliNode) cost() float32 {
	return math.Float32frombits(n.u)
}

// setCost stores a float32 cost into the u field.
func (n *zopfliNode) setCost(f float32) {
	n.u = math.Float32bits(f)
}

// initZopfliNodes initializes all nodes with length=1, distance=0,
// dcodeInsertLength=0, cost=infinity.
func initZopfliNodes(nodes []zopfliNode) {
	stub := zopfliNode{
		length:            1,
		distance:          0,
		dcodeInsertLength: 0,
		u:                 math.Float32bits(infinity),
	}
	fillSlice(nodes, stub)
}

func updateZopfliNode(next *zopfliNode, insLen, length, lenCode, dist, shortCode uint, cost float32) {
	next.length = uint32(length | ((length + 9 - lenCode) << 25))
	next.distance = uint32(dist)
	next.dcodeInsertLength = uint32((shortCode << 27) | insLen)
	next.setCost(cost)
}

// computeShortestPathFromNodes backtraces from nodes[numBytes] to nodes[0],
// writing next offsets into the u field of each visited node. Returns the
// number of commands in the optimal path.
func computeShortestPathFromNodes(nodes []zopfliNode, numBytes uint) uint {
	index := numBytes
	numCommands := uint(0)
	// Skip trailing unmatched nodes (length==1, no insert).
	for (nodes[index].dcodeInsertLength&0x7FFFFFF) == 0 && nodes[index].length == 1 {
		index--
	}
	nodes[index].u = math.MaxUint32
	for index != 0 {
		cmdLen := uint(nodes[index].commandLength())
		index -= cmdLen
		nodes[index].u = uint32(cmdLen)
		numCommands++
	}
	return numCommands
}

// zopfliCreateCommands walks the next chain in nodes, creating command
// structs. Updates distCache for non-dictionary backward references.
func zopfliCreateCommands(nodes []zopfliNode, numBytes, blockStart, maxBackwardLimit, gap uint, distCache []int, lastInsertLen *uint, commands *[]command, numLiterals *uint) {
	pos := uint(0)
	offset := nodes[0].u
	for i := uint(0); offset != math.MaxUint32; i++ {
		next := &nodes[pos+uint(offset)]
		copyLength := uint(next.copyLength())
		insertLength := uint(next.dcodeInsertLength & 0x7FFFFFF)
		pos += insertLength
		offset = next.u
		if i == 0 {
			insertLength += *lastInsertLen
			*lastInsertLen = 0
		}

		distance := uint(next.copyDistance())
		lenCode := uint(next.lengthCode())
		dictionaryStart := min(blockStart+pos, maxBackwardLimit)
		isDictionary := distance > dictionaryStart+gap
		distCode := uint(next.distanceCode())

		*commands = append(*commands, newCommand(commandConfig{
			insertLen:    insertLength,
			copyLen:      copyLength,
			copyLenDelta: int(lenCode) - int(copyLength),
			distanceCode: distCode,
		}))

		if !isDictionary && distCode > 0 {
			distCache[3] = distCache[2]
			distCache[2] = distCache[1]
			distCache[1] = distCache[0]
			distCache[0] = int(distance)
		}

		*numLiterals += insertLength
		pos += copyLength
	}
	*lastInsertLen += numBytes - pos
}
