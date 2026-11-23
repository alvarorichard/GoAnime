// Compound dictionary types, hash table construction, and matching.

package encoder

import (
	"errors"
	"unsafe"
)

// MaxCompoundDicts is the upper bound on compound dictionary chunks per
// Writer or Reader.
const MaxCompoundDicts = 15

const dictHashBits = 40

// Errors returned when validating compound dictionary parameters.
var (
	ErrTooManyDicts  = errors.New("brrr: maximum compound dictionaries exceeded")
	ErrEmptyDict     = errors.New("brrr: empty dictionary data")
	ErrQualityTooLow = errors.New("brrr: compound dictionaries require quality >= 2")
)

// PreparedDictionary is an immutable hash table built from caller-provided
// source bytes. It allows the encoder to reference those bytes as if they
// preceded the input stream.
//
// Build a PreparedDictionary once with [PrepareDictionary] and pass it to any
// number of Writers via [WriterOptions.Dictionaries]. A *PreparedDictionary is
// safe to share across goroutines.
type PreparedDictionary struct {
	source      []byte
	slotOffsets []uint32
	heads       []uint16
	items       []uint32
	bucketBits  uint32
	slotBits    uint32
	hashShift   uint32 // 64 - bucketBits, precomputed to avoid per-call arithmetic
	slotMask    uint32 // (1 << slotBits) - 1, precomputed
}

// compoundDictionary holds multiple prepared dictionary chunks that the encoder
// can reference as backward distances beyond the ring buffer.
type compoundDictionary struct {
	chunks       [MaxCompoundDicts]*PreparedDictionary
	chunkSource  [MaxCompoundDicts][]byte
	chunkOffsets [MaxCompoundDicts + 1]uint
	totalSize    uint
	numChunks    int
	// nextHead is a per-encoder write-only prefetch sink written during match
	// search. Living on compoundDictionary (per-encoder) rather than on
	// *PreparedDictionary lets a single prepared dictionary be shared across
	// concurrent Writers without racing on this write.
	nextHead uint16
}

// PrepareDictionary builds an immutable [PreparedDictionary] from the given
// source bytes, suitable for use as a compound dictionary chunk via
// [WriterOptions.Dictionaries]. The returned dictionary may be shared across
// any number of Writers and goroutines.
//
// The returned dictionary keeps a reference to data; the caller must not
// mutate data while any Writer holding the dictionary is still in use.
//
// Returns an error if data is empty.
func PrepareDictionary(data []byte) (*PreparedDictionary, error) {
	if len(data) == 0 {
		return nil, ErrEmptyDict
	}
	return newPreparedDictionary(data), nil
}

// newPreparedDictionary builds a hash table from source for compound dictionary
// matching.
func newPreparedDictionary(source []byte) *PreparedDictionary {
	sourceSize := uint32(len(source))
	if sourceSize < 8 {
		return &PreparedDictionary{source: source}
	}

	bucketBits := uint32(17)
	slotBits := uint32(7)
	const bucketLimit = uint16(32)

	// Auto-scale to fit dictionary size.
	for 16<<bucketBits < sourceSize && bucketBits < 22 {
		bucketBits++
		slotBits++
	}

	// let's say bucketBits = 17, slotBits = 7, sourceSize = 1 << 20 (1 MiB)
	numSlots := uint32(1) << slotBits             // then it's 128
	numBuckets := uint32(1) << bucketBits         // then it's 131072
	hashShift := 64 - bucketBits                  // then it's 47
	hashMask := ^uint64(0) >> (64 - dictHashBits) // then it's 0xFFFFFFFFFF
	slotMask := numSlots - 1                      // then it's 0b1111111 or 0x7F

	// Step 1: build "bloated" hasher — linked list per hash bucket.
	//
	// If source positions 10, 50, 90 all hash to bucket 7:
	//
	//   bucketHeads[7] = 90
	//   nextBucket[90] = 50
	//   nextBucket[50] = 10
	//   nextBucket[10] = ^uint32(0)   (end of list)
	//
	num := make([]uint16, numBuckets)
	bucketHeads := make([]uint32, numBuckets)
	nextBucket := make([]uint32, sourceSize)

	for i := uint32(0); i+7 < sourceSize; i++ {
		h := (loadU64LE(source, uint(i)) & hashMask) * hashMul64
		key := uint32(h >> hashShift)
		count := num[key]
		if count == 0 {
			nextBucket[i] = ^uint32(0)
		} else {
			nextBucket[i] = bucketHeads[key]
		}
		bucketHeads[key] = i
		count = min(count+1, bucketLimit)
		num[key] = count
	}

	// Step 2: find slot limits and sizes, compute slot offsets.
	d := &PreparedDictionary{
		source:      source,
		slotOffsets: make([]uint32, numSlots),
		heads:       make([]uint16, numBuckets),
		bucketBits:  bucketBits,
		slotBits:    slotBits,
		hashShift:   64 - bucketBits,
		slotMask:    numSlots - 1,
	}
	slotSize := make([]uint32, numSlots)
	slotLimit := make([]uint32, numSlots)
	totalItems := uint32(0)

	for i := range numSlots {
		slotLimit[i] = uint32(bucketLimit)
		for {
			limit := slotLimit[i]
			count := uint32(0)
			overflow := false
			for j := i; j < numBuckets; j += numSlots {
				if count >= 0xFFFF {
					overflow = true
					break
				}
				count += min(uint32(num[j]), limit)
			}
			if !overflow {
				d.slotOffsets[i] = totalItems
				slotSize[i] = count
				totalItems += count
				break
			}
			slotLimit[i]--
		}
	}
	d.items = make([]uint32, totalItems)
	clear(slotSize) // reused as per-slot write cursor in step 3

	// Step 3: transpose bucket chains into slot-contiguous items array.
	// Bucket 7's chain becomes contiguous entries with bit 31 as end marker:
	//
	//   d.items[cursor]   = 90
	//   d.items[cursor+1] = 50
	//   d.items[cursor+2] = 10 | 0x80000000
	for i := range numBuckets {
		slot := i & slotMask
		count := uint32(num[i])
		count = min(count, slotLimit[slot])
		if count == 0 {
			d.heads[i] = 0xFFFF
			continue
		}
		d.heads[i] = uint16(slotSize[slot])
		cursor := slotSize[slot] + d.slotOffsets[slot]
		slotSize[slot] += count
		pos := bucketHeads[i]
		for j := uint32(0); j < count; j++ {
			d.items[cursor] = pos
			cursor++
			pos = nextBucket[pos]
		}
		d.items[cursor-1] |= 0x80000000
	}

	return d
}

// attach appends a prepared dictionary as a chunk.
func (cd *compoundDictionary) attach(pd *PreparedDictionary) error {
	if cd.numChunks == MaxCompoundDicts {
		return ErrTooManyDicts
	}
	idx := cd.numChunks
	cd.totalSize += uint(len(pd.source))
	cd.chunks[idx] = pd
	cd.chunkSource[idx] = pd.source
	cd.chunkOffsets[idx+1] = cd.totalSize
	cd.numChunks++
	return nil
}

// findCompoundDictionaryMatch searches a single prepared dictionary for the
// best backward reference match. Two phases: distance cache check, then hash
// chain walk.
func (d *PreparedDictionary) findCompoundMatch(
	data []byte, ringBufferMask uint,
	distCache *[4]uint, cur, maxLength, distanceOffset uint,
	out *hasherSearchResult,
	prefetchSink *uint16,
) {
	sourceSize := uint(len(d.source))
	hashMask := ^uint64(0) >> (64 - dictHashBits)

	// Speculatively load from the next position's heads entry to warm the cache.
	// By the time the next call arrives, the cache line will be in L1.
	// prefetchSink points at a per-encoder slot so concurrent Writers sharing
	// the same dict do not race on this write.
	if sourceSize >= 8 {
		nextCurMasked := (cur + 1) & ringBufferMask
		nh := (loadU64LE(data, nextCurMasked) & hashMask) * hashMul64
		nextKey := uint32(nh >> d.hashShift)
		*prefetchSink = d.heads[nextKey]
	}

	boundary := distanceOffset - sourceSize

	source := d.source
	curMasked := cur & ringBufferMask
	bestScore := out.score
	bestLen := out.len

	// Phase 1: distance cache. distCache holds the last 4 backward references.
	// To match a chunk of this dictionary the distance must lie in
	// (boundary, distanceOffset]. For the typical caller the cache contains
	// in-window (small) distances, so all 4 checks fail; bail out cheaply by
	// skipping the loop when the maximum cached distance is already at or
	// below boundary.
	maxCachedDist := distCache[0]
	if d := distCache[1]; d > maxCachedDist {
		maxCachedDist = d
	}
	if d := distCache[2]; d > maxCachedDist {
		maxCachedDist = d
	}
	if d := distCache[3]; d > maxCachedDist {
		maxCachedDist = d
	}
	if maxCachedDist > boundary {
		for i := range 4 {
			distance := distCache[i]
			if distance <= boundary || distance > distanceOffset {
				continue
			}
			offset := distanceOffset - distance
			limit := min(sourceSize-offset, maxLength)
			ml := uint(matchLen(source[offset:offset+limit], data[curMasked:curMasked+limit], int(limit)))
			if ml >= 2 {
				score := backwardReferenceScoreUsingLastDistance(ml)
				if bestScore < score {
					if i != 0 {
						score -= backwardReferencePenaltyUsingLastDistance(uint(i))
					}
					if bestScore < score {
						bestScore = score
						if ml > bestLen {
							bestLen = ml
						}
						out.len = ml
						out.lenCodeDelta = 0
						out.distance = distance
						out.score = bestScore
					}
				}
			}
		}
	}

	// Short dictionaries have no hash table but can match cached distances.
	if sourceSize < 8 {
		return
	}

	// Raise bestLen floor to 3 so hash chain only accepts length >= 4.
	if bestLen < 3 {
		bestLen = 3
	}

	// Phase 2: hash chain walk.
	h := (loadU64LE(data, curMasked) & hashMask) * hashMul64
	key := uint32(h >> d.hashShift)
	slot := key & d.slotMask
	head := d.heads[key]
	if head == 0xFFFF {
		return
	}
	// Hoist loop-invariant guards out of the inner loop. bestLen only grows
	// (after a successful match), so once any of these is violated, no future
	// iteration can record an improvement — the original code skipped them
	// silently via the inner if, so an early return preserves semantics while
	// avoiding pointless chain traversal. After bestLen grows we re-check
	// before computing the next curProbe.
	if bestLen >= maxLength || curMasked+bestLen > ringBufferMask || sourceSize <= bestLen {
		return
	}
	maxValidOffset := sourceSize - bestLen
	// Match the native byte order of the source probe.
	curProbe := le32(loadU32LE(data, curMasked+bestLen-3))
	// Hoist the per-iteration `offset + bestLen - 3` address computation by
	// folding `bestLen - 3` into a base pointer over source. Inside the loop
	// the probe load becomes a single `MOV` with `[srcProbeBase + offset*1]`
	// instead of the `LEA offset+bestLen; MOV -3(base+lea)` pair the compiler
	// emits when the loop-invariant `bestLen - 3` term is left inside
	// loadU32LE. bestLen is >= 3 here, so the subtraction never underflows.
	sourcePtr := unsafe.Pointer(unsafe.SliceData(source))
	srcProbeBase := unsafe.Add(sourcePtr, bestLen-3)
	// Pin items.data into a single local pointer: the gc compiler otherwise
	// spills d and reloads both items.data and items.len from the slice header
	// on every iteration. The chain is built so every visited index is within
	// d.items, so the bounds check is provably unnecessary.
	itemsPtr := unsafe.Pointer(unsafe.SliceData(d.items))
	for i := d.slotOffsets[slot] + uint32(head); ; i++ {
		item := *(*uint32)(unsafe.Add(itemsPtr, uintptr(i)*4))
		offset := uint(item & 0x7FFFFFFF)
		// Defer `distance := distanceOffset - offset` and the
		// `distance <= maxBackwardDistance` check until after the probe
		// prefilter succeeds. The probe compare fails on most iterations, so
		// hoisting the SUB+CMP+JA out of the per-iteration path saves three
		// instructions per failed probe; the match-found branch pays them
		// instead but runs at most a handful of times per call.
		if offset < maxValidOffset &&
			curProbe == *(*uint32)(unsafe.Add(srcProbeBase, offset)) {
			distance := distanceOffset - offset
			if distance <= maxBackwardDistance {
				limit := min(sourceSize-offset, maxLength)
				ml := uint(matchLen(source[offset:offset+limit], data[curMasked:curMasked+limit], int(limit)))
				if ml >= 4 {
					score := backwardReferenceScore(ml, distance)
					if bestScore < score {
						bestScore = score
						bestLen = ml
						out.len = bestLen
						out.lenCodeDelta = 0
						out.distance = distance
						out.score = bestScore
						if bestLen >= maxLength || curMasked+bestLen > ringBufferMask || sourceSize <= bestLen {
							return
						}
						maxValidOffset = sourceSize - bestLen
						srcProbeBase = unsafe.Add(sourcePtr, bestLen-3)
						curProbe = le32(loadU32LE(data, curMasked+bestLen-3))
					}
				}
			}
		}
		if item&0x80000000 != 0 {
			break
		}
	}
}

// findAllCompoundMatches searches this prepared dictionary for all matches
// at the given position, returning them sorted by strictly increasing length.
func (d *PreparedDictionary) findAllCompoundMatches(
	data []byte, ringBufferMask, curIx, minLength, maxLength, distanceOffset, maxDistance uint,
	matches []backwardMatch,
) uint {
	sourceSize := uint(len(d.source))
	if sourceSize < 8 {
		return 0
	}

	hashMask := ^uint64(0) >> (64 - dictHashBits)

	source := d.source
	curIxMasked := curIx & ringBufferMask
	bestLen := minLength

	h := (loadU64LE(data, curIxMasked) & hashMask) * hashMul64
	key := uint32(h >> d.hashShift)
	slot := key & d.slotMask
	head := d.heads[key]
	if head == 0xFFFF {
		return 0
	}
	matchLimit := uint(len(matches))
	found := uint(0)

	for i := d.slotOffsets[slot] + uint32(head); ; i++ {
		item := d.items[i]
		offset := uint(item & 0x7FFFFFFF)
		distance := distanceOffset - offset
		limit := min(sourceSize-offset, maxLength)
		if distance <= maxDistance &&
			curIxMasked+bestLen <= ringBufferMask &&
			bestLen < limit &&
			data[curIxMasked+bestLen] == source[offset+bestLen] {
			ml := uint(matchLen(source[offset:], data[curIxMasked:], int(limit)))
			if ml > bestLen {
				bestLen = ml
				matches[found] = newBackwardMatch(distance, ml)
				found++
				if found == matchLimit {
					break
				}
			}
		}
		if item&0x80000000 != 0 {
			break
		}
	}
	return found
}

// lookupAllCompoundMatches searches all chunks in the compound dictionary for
// matches, returning them sorted by strictly increasing length.
func (cd *compoundDictionary) lookupAllMatches(
	data []byte, ringBufferMask, curIx, minLength, maxLength, maxRingBufferDistance, maxDistance uint,
	matches []backwardMatch,
) uint {
	baseOffset := maxRingBufferDistance + 1 + cd.totalSize - 1
	totalFound := uint(0)
	ml := minLength
	for d := range cd.numChunks {
		totalFound += cd.chunks[d].findAllCompoundMatches(
			data, ringBufferMask, curIx, ml, maxLength,
			baseOffset-cd.chunkOffsets[d], maxDistance,
			matches[totalFound:])
		if totalFound > 0 {
			ml = matches[totalFound-1].matchLength()
		}
		if totalFound >= uint(len(matches)) {
			break
		}
	}
	return totalFound
}

// lookupCompoundDictionaryMatch iterates all chunks in the compound dictionary,
// calling findCompoundDictionaryMatch for each. The single-chunk path is split
// out so the compiler emits a tail call without a loop prologue, eliminating
// per-call register save/restore around the call. Most callers register only
// one PreparedDictionary, so this is the overwhelmingly common case.
func (cd *compoundDictionary) lookupMatch(
	data []byte, ringBufferMask uint,
	distCache *[4]uint, cur, maxLength, maxRingBufferDistance uint,
	sr *hasherSearchResult,
) {
	baseOffset := maxRingBufferDistance + 1 + cd.totalSize - 1
	if cd.numChunks == 1 {
		cd.chunks[0].findCompoundMatch(
			data, ringBufferMask,
			distCache, cur, maxLength,
			baseOffset, sr,
			&cd.nextHead)
		return
	}
	for i := range cd.numChunks {
		cd.chunks[i].findCompoundMatch(
			data, ringBufferMask,
			distCache, cur, maxLength,
			baseOffset-cd.chunkOffsets[i], sr,
			&cd.nextHead)
	}
}
