package pomfy

import (
	"context"
	"encoding/json"
)

// Test helpers the packages ABOVE this one (the api/dispatch layers) use to
// stand the full Pomfy flow up offline. The pomfy package's own tests have
// their own, stronger fixtures; these export only the three pieces the glue
// tests cannot otherwise reach: a real seal, a real solve and the same
// judge the fake verify endpoints use. Kept out of test-only build tags
// because they live beside the exported ForTest constructors the repo
// already ships (NewClientForTest and kin).

// SealForTest seals plaintext the way the byse API answers /embed/playback:
// unpadded base64url fields, the key at parts 9 and 22 (1-based), version
// "9", and the whole thing wrapped as {"playback": {...}} so a fake CDN can
// pour it into its response verbatim.
func SealForTest(plaintext []byte) json.RawMessage {
	return json.RawMessage(mustMarshalSeal(sealForTestFn(plaintext)))
}

func mustMarshalSeal(seal playbackSeal) []byte {
	// The seal carries string fields only: marshaling cannot fail.
	raw, _ := json.Marshal(map[string]any{"playback": seal})
	return raw
}

// SolveForTest scans for a nonce meeting the difficulty on prefix:s, the
// exact oracle the fake byse verify endpoints reuse. It balances out fast:
// tests choose difficulties of 6 or less.
func SolveForTest(prefix string, difficulty int) (string, error) {
	return solveNonce(context.Background(), prefix+":", difficulty, powBudgetMost)
}

// CheckForTest is the judge one raw step gets: does prefix+solution meet the
// difficulty the production server applies? (prefix carries its own colon.)
func CheckForTest(prefix, solution string, difficulty int) bool {
	return checkNonce(prefix, solution, difficulty)
}
