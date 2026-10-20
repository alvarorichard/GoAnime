package startflix

import (
	"encoding/base64"
	"encoding/binary"
	"math"
	"testing"

	"github.com/stretchr/testify/require"
)

// An upstream module can return i64 values even though its memory uses
// 32-bit addresses. Neither pointers nor lengths may silently wrap.
func TestVIDSRCDecryptRejectsInvalidMemoryRanges(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		ptr  uint64
		out  uint64
		want string
	}{
		{name: "pointer truncation", ptr: 1 << 32, want: "allocation exceeds"},
		{name: "nonce offset overflow", ptr: math.MaxUint32, want: "allocation exceeds"},
		{name: "failed ciphertext write", ptr: 65530, want: "memory write failed"},
		{name: "length truncation", out: 1 << 32, want: "output exceeds"},
		{name: "read beyond memory", out: 65536, want: "memory read failed"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			enc := base64.StdEncoding.EncodeToString(make([]byte, 12))
			_, err := decryptVIDSRC(constantVIDSRCModule(tc.ptr, tc.out), enc)
			require.ErrorContains(t, err, tc.want)
		})
	}
}

// constantVIDSRCModule builds a tiny module with a single memory page and
// alloc/decrypt exports that return the selected positive i64 constants.
func constantVIDSRCModule(ptr, out uint64) []byte {
	module := []byte("\x00asm\x01\x00\x00\x00")
	section := func(id byte, body []byte) {
		module = append(module, id)
		module = binary.AppendUvarint(module, uint64(len(body)))
		module = append(module, body...)
	}
	section(1, []byte{2, 0x60, 1, 0x7e, 1, 0x7e, 0x60, 2, 0x7e, 0x7e, 1, 0x7e})
	section(3, []byte{2, 0, 1})
	section(5, []byte{1, 0, 1})
	section(7, []byte("\x03\x05alloc\x00\x00\x07decrypt\x00\x01\x06memory\x02\x00"))
	code := []byte{2}
	for _, value := range []uint64{ptr, out} {
		body := binary.AppendUvarint([]byte{0, 0x42}, value)
		body = append(body, 0x0b)
		code = binary.AppendUvarint(code, uint64(len(body)))
		code = append(code, body...)
	}
	section(10, code)
	return module
}
