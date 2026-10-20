package startflix

import (
	"context"
	"encoding/base64"
	"fmt"
	"math"
	"strings"

	"github.com/tetratelabs/wazero"
)

// decryptVIDSRC decrypts the fallback source's protected stream_urls through
// its WASM decryptor exactly the way the player's vsdec.js does in the
// browser (kept byte-for-byte so future upstream rotations keep working):
//
//	ciphertext = base64(nonce[12] || data)
//	wasm: alloc(len) -> ptr; memory[ptr…ptr+len] = ciphertext;
//	      outLen = decrypt(ptr, len); *memory[ptr+12 … ptr+12+outLen]
//
// The decryptor module embeds one key per ~5-minute window, so it is fetched
// at resolution time — the same wazero call is reused for offline fixtures in
// the tests.
func decryptVIDSRC(wasmBytes []byte, encB64 string) ([]string, error) {
	enc, err := base64.StdEncoding.DecodeString(encB64)
	if err != nil || len(enc) < 12 {
		return nil, fmt.Errorf("vidsrc fallback: unreadable stream_urls blob (%d bytes)", len(enc))
	}

	ctx := context.Background()
	r := wazero.NewRuntime(ctx)
	defer func() { _ = r.Close(ctx) }()

	mod, err := r.Instantiate(ctx, wasmBytes)
	if err != nil {
		return nil, fmt.Errorf("vidsrc fallback: decryptor module failed to load: %w", err)
	}
	alloc, decrypt, mem := mod.ExportedFunction("alloc"), mod.ExportedFunction("decrypt"), mod.ExportedMemory("memory")
	if alloc == nil || decrypt == nil || mem == nil {
		return nil, fmt.Errorf("vidsrc fallback: decryptor lacks alloc/decrypt/memory exports")
	}

	ptr, err := alloc.Call(ctx, uint64(len(enc)))
	if err != nil || len(ptr) == 0 {
		return nil, fmt.Errorf("vidsrc fallback: decryptor alloc failed: %w", err)
	}
	address := ptr[0]
	if address > math.MaxUint32-12 {
		return nil, fmt.Errorf("vidsrc fallback: decryptor allocation exceeds memory address range")
	}
	base := uint32(address)
	if !mem.Write(base, enc) {
		return nil, fmt.Errorf("vidsrc fallback: decryptor memory write failed")
	}
	out, err := decrypt.Call(ctx, uint64(base), uint64(len(enc)))
	if err != nil || len(out) == 0 {
		return nil, fmt.Errorf("vidsrc fallback: decryptor decrypt failed: %w", err)
	}
	length := out[0]
	if length > math.MaxUint32 {
		return nil, fmt.Errorf("vidsrc fallback: decryptor output exceeds memory address range")
	}
	// The plaintext starts after the 12-byte nonce the browser's TextDecoder
	// slice skips (vsdec.js decryptWith: memory.buffer, ptr + 12, outLen).
	plain, ok := mem.Read(base+12, uint32(length))
	if !ok {
		return nil, fmt.Errorf("vidsrc fallback: decryptor memory read failed")
	}
	return cleanStreams(strings.Split(string(plain), "\n")), nil
}
