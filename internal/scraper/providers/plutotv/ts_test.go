package plutotv

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---- a transport stream builder --------------------------------------------

// tsPES builds the packet that starts a PES of the given stream: its header
// carries pts (and dts when dts >= 0), and the packet carries a PCR of pcr
// ticks when pcr >= 0.
func tsPES(pid int, streamID byte, pts, dts, pcr int64) []byte {
	p := bytes.Repeat([]byte{0xFF}, tsPacket)
	p[0], p[1], p[2] = tsSync, 0x40|byte(pid>>8), byte(pid)
	off := 4
	if pcr >= 0 {
		p[3] = 0x30
		p[4], p[5] = 7, 0x10
		p[6], p[7], p[8], p[9] = byte(pcr>>25), byte(pcr>>17), byte(pcr>>9), byte(pcr>>1)
		p[10], p[11] = byte(pcr&1)<<7|0x7E, 0
		off = 12
	} else {
		p[3] = 0x10
	}
	flags, hdr := byte(0x80), 5
	if dts >= 0 {
		flags, hdr = 0xC0, 10
	}
	copy(p[off:], []byte{0, 0, 1, streamID, 0, 0, 0x80, flags, byte(hdr)})
	off += 9
	prefix := byte(0x20)
	if dts >= 0 {
		prefix = 0x30
	}
	b := make([]byte, 5)
	b[0] = prefix
	writeTS(b, pts)
	copy(p[off:], b)
	off += 5
	if dts >= 0 {
		b[0] = 0x10
		writeTS(b, dts)
		copy(p[off:], b)
	}
	return p
}

// tsFiller is a packet that carries data and no timestamps.
func tsFiller(pid int, n byte) []byte {
	p := bytes.Repeat([]byte{n}, tsPacket)
	p[0], p[1], p[2], p[3] = tsSync, byte(pid>>8), byte(pid), 0x10
	return p
}

// tsSegment is a segment: a video PES and an audio PES with fillers between,
// starting its clock at start ticks.
func tsSegment(start int64, fillers int) []byte {
	var b []byte
	b = append(b, tsPES(0x100, 0xE0, start+3003, start, start-9000)...)
	for i := range fillers {
		b = append(b, tsFiller(0x100, byte(i))...)
	}
	b = append(b, tsPES(0x101, 0xC0, start, -1, -1)...)
	b = append(b, tsPES(0x100, 0xE0, start+6006, start+3003, -1)...)
	return b
}

func pad(data []byte) []byte {
	n := aes.BlockSize - len(data)%aes.BlockSize
	return append(append([]byte{}, data...), bytes.Repeat([]byte{byte(n)}, n)...)
}

func encrypt(data, key []byte, iv [16]byte) []byte {
	block, _ := aes.NewCipher(key)
	out := pad(data)
	cipher.NewCBCEncrypter(block, iv[:]).CryptBlocks(out, out)
	return out
}

// ---- the transport stream ---------------------------------------------------

func TestFirstTimestampIsTheEarliestDTSOrPTS(t *testing.T) {
	seg := tsSegment(90000, 3)
	at, ok := firstTimestamp(seg)
	require.True(t, ok)
	assert.EqualValues(t, 90000, at, "the audio PTS and the video DTS, not the video PTS that runs ahead")

	_, ok = firstTimestamp(tsFiller(0x100, 1))
	assert.False(t, ok, "a packet with no PES has no clock")
	_, ok = firstTimestamp(nil)
	assert.False(t, ok)
	_, ok = firstTimestamp(bytes.Repeat([]byte{1}, 400))
	assert.False(t, ok, "not a transport stream")

	// A buffer cut mid-packet, as a probe is: what is whole is read.
	at, ok = firstTimestamp(seg[:len(seg)-100])
	require.True(t, ok)
	assert.EqualValues(t, 90000, at)
}

func TestShiftTimestampsMovesPTSDTSAndPCR(t *testing.T) {
	for _, delta := range []int64{0, 90000, 450_000_000, -45000, 1<<33 - 1} {
		seg := tsSegment(1_000_000, 4)
		before := append([]byte{}, seg...)
		require.NoError(t, shiftTimestamps(seg, delta))

		at, ok := firstTimestamp(seg)
		require.True(t, ok)
		assert.EqualValues(t, (1_000_000+delta)&ptsMask, at, "delta %d", delta)

		// The video PES: PTS and DTS both moved by the same.
		pts, dts := readTS(seg[4+8+9:]), readTS(seg[4+8+14:])
		assert.EqualValues(t, (1_000_000+3003+delta)&ptsMask, pts)
		assert.EqualValues(t, (1_000_000+delta)&ptsMask, dts)
		// The PCR base moved, its extension and reserved bits did not.
		base := int64(seg[6])<<25 | int64(seg[7])<<17 | int64(seg[8])<<9 | int64(seg[9])<<1 | int64(seg[10]>>7)
		assert.EqualValues(t, (1_000_000-9000+delta)&ptsMask, base)
		assert.Equal(t, before[10]&0x7F, seg[10]&0x7F)
		assert.Equal(t, before[11], seg[11])

		// Shifting back is the identity.
		require.NoError(t, shiftTimestamps(seg, -delta))
		assert.Equal(t, before, seg, "delta %d", delta)
	}
}

func TestShiftTimestampsLeavesWhatIsNotATimestampAlone(t *testing.T) {
	filler := tsFiller(0x100, 7)
	seg := append(append([]byte{}, filler...), tsFiller(0x11, 9)...)
	before := append([]byte{}, seg...)
	require.NoError(t, shiftTimestamps(seg, 12345))
	assert.Equal(t, before, seg, "payload bytes that look like a start code are not touched without the start indicator")

	// A PES start whose payload is too short for a header is left as it is.
	short := tsFiller(0x100, 0)
	short[1] |= 0x40
	short[3] = 0x30
	short[4] = 170 // adaptation to the end but 14 bytes
	copy(short[4+1+170:], []byte{0, 0, 1, 0xE0})
	beforeShort := append([]byte{}, short...)
	require.NoError(t, shiftTimestamps(short, 999))
	assert.Equal(t, beforeShort, short)
}

func TestShiftTimestampsRefusesWhatIsNotTS(t *testing.T) {
	for name, b := range map[string][]byte{
		"empty":       nil,
		"cut":         tsSegment(0, 1)[:300],
		"out of sync": append([]byte{0}, tsSegment(0, 1)...),
		"text":        []byte("#EXTM3U\n"),
	} {
		assert.ErrorIs(t, shiftTimestamps(b, 1), errNotTS, name)
	}
}

// Whatever the stream holds, shifting and shifting back changes nothing, and a
// shift moves the first timestamp by exactly the delta.
func TestPropertyShiftIsReversible(t *testing.T) {
	rnd := rand.New(rand.NewSource(7))
	for range 400 {
		var seg []byte
		for range 1 + rnd.Intn(30) {
			switch rnd.Intn(3) {
			case 0:
				seg = append(seg, tsPES(0x100+rnd.Intn(3), byte(0xC0+rnd.Intn(32)), rnd.Int63n(ptsMask), rnd.Int63n(ptsMask), rnd.Int63n(ptsMask))...)
			case 1:
				seg = append(seg, tsPES(0x100, 0xE0, rnd.Int63n(ptsMask), -1, -1)...)
			default:
				seg = append(seg, tsFiller(0x100+rnd.Intn(8), byte(rnd.Intn(256)))...)
			}
		}
		orig := append([]byte{}, seg...)
		delta := rnd.Int63n(1<<34) - 1<<33
		require.NoError(t, shiftTimestamps(seg, delta))
		require.NoError(t, shiftTimestamps(seg, -delta))
		require.Equal(t, orig, seg)
	}
}

// ---- the cipher -------------------------------------------------------------

func TestDecryptAES128(t *testing.T) {
	key := []byte("0123456789abcdef")
	iv := [16]byte{15: 9}
	plain := tsSegment(0, 5)

	got, err := decryptAES128(encrypt(plain, key, iv), key, iv)
	require.NoError(t, err)
	assert.Equal(t, plain, got, "the padding is gone")

	_, err = decryptAES128(encrypt(plain, key, iv), []byte("fedcba9876543210"), iv)
	assert.Error(t, err, "a wrong key is told by its padding")
	_, err = decryptAES128(encrypt(plain, key, iv)[:30], key, iv)
	assert.Error(t, err, "not whole blocks")
	_, err = decryptAES128(nil, key, iv)
	assert.Error(t, err)
	_, err = decryptAES128(encrypt(plain, key, iv), []byte("short"), iv)
	assert.Error(t, err, "not a key")
}

func TestDecryptPrefixNeedsNoPaddingAndIgnoresAPartialBlock(t *testing.T) {
	key := []byte("0123456789abcdef")
	iv := [16]byte{15: 3}
	plain := tsSegment(5000, 40)
	enc := encrypt(plain, key, iv)

	got, err := decryptPrefix(append([]byte{}, enc[:4097]...), key, iv)
	require.NoError(t, err)
	assert.Equal(t, plain[:4096], got)
	_, err = decryptPrefix([]byte{1, 2, 3}, key, iv)
	assert.Error(t, err)
}

// The padding is the only thing that tells a wrong key from a right one, so
// each way it can be wrong is built by hand: encrypted as it is, not padded.
func TestDecryptAES128RefusesEveryBadPadding(t *testing.T) {
	key := []byte("0123456789abcdef")
	iv := [16]byte{15: 1}
	block, _ := aes.NewCipher(key)
	raw := func(last ...byte) []byte {
		plain := bytes.Repeat([]byte{0x11}, 16)
		copy(plain[16-len(last):], last)
		cipher.NewCBCEncrypter(block, iv[:]).CryptBlocks(plain, plain)
		return plain
	}
	for name, enc := range map[string][]byte{
		"zero":                raw(0),
		"over the block size": raw(17),
		"far over":            raw(0xC8),
		"one wrong byte":      raw(0x03, 0x03, 0x02, 0x03)[:16], // …03 03 02 03: a pad of 3 whose bytes differ
	} {
		_, err := decryptAES128(append([]byte{}, enc...), key, iv)
		assert.Error(t, err, name)
	}
	got, err := decryptAES128(raw(4, 4, 4, 4), key, iv)
	require.NoError(t, err)
	assert.Len(t, got, 12, "a pad of four is four bytes")
	got, err = decryptAES128(raw(1), key, iv)
	require.NoError(t, err)
	assert.Len(t, got, 15)
}

// A segment of the size the service serves at its best rung, about 1.7 MB.
func benchSegment() ([]byte, []byte, [16]byte) {
	key, iv := []byte("0123456789abcdef"), [16]byte{15: 1}
	return encrypt(tsSegment(90000, 9000), key, iv), key, iv
}

func BenchmarkDecryptSegment(b *testing.B) {
	enc, key, iv := benchSegment()
	buf := make([]byte, len(enc))
	b.SetBytes(int64(len(enc)))
	for b.Loop() {
		copy(buf, enc)
		if _, err := decryptAES128(buf, key, iv); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkShiftSegment(b *testing.B) {
	enc, key, iv := benchSegment()
	plain, _ := decryptAES128(append([]byte{}, enc...), key, iv)
	b.SetBytes(int64(len(plain)))
	for b.Loop() {
		if err := shiftTimestamps(plain, 540000); err != nil {
			b.Fatal(err)
		}
	}
}
