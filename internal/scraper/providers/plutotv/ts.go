package plutotv

import (
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

// MPEG-TS plumbing for rebasing a title's timestamps (see rebase.go).
//
// Some titles are cut into ranges that are each a file of their own, and every
// file restarts its clock at about one second. A player that is asked to seek
// looks for a timestamp that no later file ever reaches, so it freezes. Adding
// each range's offset to the timestamps inside its segments puts the title on
// one timeline, which is what a seek needs.

const (
	tsPacket = 188
	tsSync   = 0x47

	// ptsMask is the 33 bits a PTS, DTS and the PCR base are kept in.
	ptsMask = 1<<33 - 1
)

var errNotTS = errors.New("not an MPEG transport stream")

// tsLayout returns the offset of the payload of the packet and, when the
// adaptation field carries a PCR, the offset of it. A packet that has no
// payload answers -1.
func tsLayout(p []byte) (payload, pcr int) {
	pcr = -1
	afc := p[3] >> 4 & 3
	payload = 4
	if afc&2 != 0 {
		n := int(p[4])
		if n > 0 && 5+n <= tsPacket && p[5]&0x10 != 0 && n >= 7 {
			pcr = 6
		}
		payload = 5 + n
	}
	if afc&1 == 0 || payload >= tsPacket {
		return -1, pcr
	}
	return payload, pcr
}

// pesTimes returns the offsets of the PTS and the DTS of the PES header that
// starts at the payload of p, or -1 for the one it does not carry.
func pesTimes(p []byte, payload int) (pts, dts int) {
	pts, dts = -1, -1
	if p[1]&0x40 == 0 { // payload_unit_start_indicator
		return
	}
	b := p[payload:]
	if len(b) < 14 || b[0] != 0 || b[1] != 0 || b[2] != 1 {
		return
	}
	switch id := b[3]; id {
	case 0xBC, 0xBE, 0xBF, 0xF0, 0xF1, 0xF2, 0xF8, 0xFF: // no optional header
		return
	}
	switch b[7] >> 6 {
	case 2:
		pts = payload + 9
	case 3:
		if len(b) >= 19 {
			pts, dts = payload+9, payload+14
		}
	}
	return
}

func readTS(b []byte) int64 {
	return int64(b[0]>>1&7)<<30 | int64(b[1])<<22 | int64(b[2]>>1)<<15 | int64(b[3])<<7 | int64(b[4]>>1)
}

func writeTS(b []byte, v int64) {
	// Each byte takes its own bits of v, so what is above 33 bits is dropped.
	b[0] = b[0]&0xF0 | byte(v>>30&7)<<1 | 1
	b[1] = byte(v >> 22 & 0xFF)
	b[2] = byte(v>>15&0x7F)<<1 | 1
	b[3] = byte(v >> 7 & 0xFF)
	b[4] = byte(v&0x7F)<<1 | 1
}

// checkTS reports whether b is a whole number of transport packets.
func checkTS(b []byte) error {
	if len(b) == 0 || len(b)%tsPacket != 0 {
		return errNotTS
	}
	for i := 0; i < len(b); i += tsPacket {
		if b[i] != tsSync {
			return errNotTS
		}
	}
	return nil
}

// firstTimestamp is the earliest DTS (or PTS where there is none) of the PES
// packets in b, in 90 kHz ticks. b may stop short of a whole packet.
func firstTimestamp(b []byte) (int64, bool) {
	best, found := int64(0), false
	for i := 0; i+tsPacket <= len(b); i += tsPacket {
		p := b[i : i+tsPacket]
		if p[0] != tsSync {
			return 0, false
		}
		payload, _ := tsLayout(p)
		if payload < 0 {
			continue
		}
		pts, dts := pesTimes(p, payload)
		at := dts
		if at < 0 {
			at = pts
		}
		if at < 0 {
			continue
		}
		if v := readTS(p[at:]); !found || v < best {
			best, found = v, true
		}
	}
	return best, found
}

// shiftTimestamps adds delta ticks to every PTS, DTS and PCR in b, in place.
func shiftTimestamps(b []byte, delta int64) error {
	if err := checkTS(b); err != nil {
		return err
	}
	for i := 0; i < len(b); i += tsPacket {
		p := b[i : i+tsPacket]
		payload, pcr := tsLayout(p)
		if pcr >= 0 {
			base := int64(p[pcr])<<25 | int64(p[pcr+1])<<17 | int64(p[pcr+2])<<9 | int64(p[pcr+3])<<1 | int64(p[pcr+4]>>7)
			base = (base + delta) & ptsMask
			p[pcr] = byte(base >> 25)
			p[pcr+1] = byte(base >> 17 & 0xFF)
			p[pcr+2] = byte(base >> 9 & 0xFF)
			p[pcr+3] = byte(base >> 1 & 0xFF)
			p[pcr+4] = byte(base&1)<<7 | p[pcr+4]&0x7F
		}
		if payload < 0 {
			continue
		}
		pts, dts := pesTimes(p, payload)
		if pts >= 0 {
			writeTS(p[pts:], readTS(p[pts:])+delta)
		}
		if dts >= 0 {
			writeTS(p[dts:], readTS(p[dts:])+delta)
		}
	}
	return nil
}

// decryptAES128 decrypts an HLS segment (AES-128, CBC) in place and returns it
// without its PKCS#7 padding. With whole transport packets the padding is
// never the transport's own, so a segment that does not end in valid padding
// is refused rather than guessed at.
func decryptAES128(data, key []byte, iv [16]byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 || len(data)%aes.BlockSize != 0 {
		return nil, errors.New("encrypted segment is not a whole number of blocks")
	}
	cipher.NewCBCDecrypter(block, iv[:]).CryptBlocks(data, data)
	pad := int(data[len(data)-1])
	if pad < 1 || pad > aes.BlockSize || pad > len(data) {
		return nil, errors.New("segment has no valid padding: wrong key")
	}
	for _, v := range data[len(data)-pad:] {
		if int(v) != pad {
			return nil, errors.New("segment has no valid padding: wrong key")
		}
	}
	return data[:len(data)-pad], nil
}

// decryptPrefix decrypts the start of a segment, which needs no padding: CBC
// decrypts block by block from the IV. A trailing partial block is dropped.
func decryptPrefix(data, key []byte, iv [16]byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	data = data[:len(data)/aes.BlockSize*aes.BlockSize]
	if len(data) == 0 {
		return nil, errors.New("no data to decrypt")
	}
	cipher.NewCBCDecrypter(block, iv[:]).CryptBlocks(data, data)
	return data, nil
}
