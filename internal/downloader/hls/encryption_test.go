package hls

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"encoding/binary"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// encryptCBC seals plain the way an HLS packager does for METHOD=AES-128:
// AES-128-CBC with PKCS#7 padding.
func encryptCBC(t *testing.T, key, iv, plain []byte) []byte {
	t.Helper()
	block, err := aes.NewCipher(key)
	if err != nil {
		t.Fatal(err)
	}
	pad := aes.BlockSize - len(plain)%aes.BlockSize
	buf := append(bytes.Clone(plain), bytes.Repeat([]byte{byte(pad)}, pad)...)
	cipher.NewCBCEncrypter(block, iv).CryptBlocks(buf, buf)
	return buf
}

// sequenceIV is the IV a segment uses when its key gives none: its media
// sequence number as a big-endian 128-bit integer (RFC 8216 §5.2).
func sequenceIV(seq uint64) []byte {
	iv := make([]byte, aes.BlockSize)
	binary.BigEndian.PutUint64(iv[8:], seq)
	return iv
}

// TestAES128SegmentsAreDecrypted pins RFC 8216 AES-128 support: each segment
// is decrypted with the key in force for it — an explicit IV, the implicit
// sequence-number IV after a key rotation, and METHOD=NONE turning
// encryption off again — so the file on disk is the plain stream, never the
// ciphertext.
func TestAES128SegmentsAreDecrypted(t *testing.T) {
	t.Parallel()
	key1 := []byte("0123456789abcdef")
	key2 := []byte("fedcba9876543210")
	iv1 := []byte("IVIVIVIVIVIVIVIV")
	plains := [][]byte{
		bytes.Repeat([]byte("first segment "), 37),
		bytes.Repeat([]byte("second segment, rotated key "), 41),
		[]byte("third segment travels in the clear"),
	}
	const firstSeq = 7
	bodies := map[string][]byte{
		"/hls/k1.bin":  key1,
		"/keys/k2.bin": key2,
		"/hls/seg0.ts": encryptCBC(t, key1, iv1, plains[0]),
		"/hls/seg1.ts": encryptCBC(t, key2, sequenceIV(firstSeq+1), plains[1]),
		"/hls/seg2.ts": plains[2],
		"/hls/index.m3u8": []byte(strings.Join([]string{
			"#EXTM3U", "#EXT-X-VERSION:3", "#EXT-X-TARGETDURATION:4",
			"#EXT-X-MEDIA-SEQUENCE:7",
			`#EXT-X-KEY:METHOD=AES-128,URI="k1.bin",IV=0x` + hex.EncodeToString(iv1),
			"#EXTINF:4.0,", "seg0.ts",
			`#EXT-X-KEY:METHOD=AES-128,URI="/keys/k2.bin"`,
			"#EXTINF:4.0,", "seg1.ts",
			"#EXT-X-KEY:METHOD=NONE",
			"#EXTINF:4.0,", "seg2.ts",
			"#EXT-X-ENDLIST", "",
		}, "\n")),
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, ok := bodies[r.URL.Path]
		if !ok {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(body)
	}))
	t.Cleanup(srv.Close)

	out := filepath.Join(t.TempDir(), "ep.ts")
	if err := DownloadToFileWithClient(context.Background(), srv.Client(), srv.URL+"/hls/index.m3u8", out, nil, nil); err != nil {
		t.Fatalf("download: %v", err)
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if want := bytes.Join(plains, nil); !bytes.Equal(got, want) {
		t.Errorf("file holds %d bytes that are not the plain stream (%d bytes); starts % x", len(got), len(want), got[:min(16, len(got))])
	}
}

// TestUnsupportedEncryptionFailsInsteadOfWritingCiphertext: SAMPLE-AES (and
// any other method) cannot be undone by concatenating segments, so the
// download must fail — letting the caller fall back to yt-dlp — rather than
// report a file of ciphertext as a finished video.
func TestUnsupportedEncryptionFailsInsteadOfWritingCiphertext(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/index.m3u8":
			_, _ = w.Write([]byte("#EXTM3U\n#EXT-X-TARGETDURATION:4\n" +
				`#EXT-X-KEY:METHOD=SAMPLE-AES,URI="skd://key",KEYFORMAT="com.apple.streamingkeydelivery"` + "\n" +
				"#EXTINF:4.0,\nseg0.ts\n#EXT-X-ENDLIST\n"))
		case "/seg0.ts":
			_, _ = w.Write(bytes.Repeat([]byte{0x47}, 188))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	out := filepath.Join(t.TempDir(), "ep.ts")
	err := DownloadToFileWithClient(context.Background(), srv.Client(), srv.URL+"/index.m3u8", out, nil, nil)
	if err == nil {
		t.Fatal("a SAMPLE-AES stream was reported as downloaded")
	}
	if _, statErr := os.Stat(out); statErr == nil {
		t.Error("a file of ciphertext was left behind")
	}
}
