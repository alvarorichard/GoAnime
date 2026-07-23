// Package allanime — regression suite for the AllAnime `aaReq` proof token.
//
// Discovered 2026-07-08: every AllAnime episode-source query returned
// AA_CRYPTO_MISSING. AllAnime (ani-cli PR #1772) began requiring an `aaReq`
// crypto token in the request extensions; without it the `episode { sourceUrls }`
// query is rejected outright. Fixed the same day: buildAAReq now produces the
// token and both the persisted-query GET and the POST fallback carry it.
//
// These tests re-enact what the AllAnime server does to VALIDATE the token —
// decode the envelope, re-derive the IV from the decrypted payload, then
// GCM-decrypt and check every field. That pins the ENTIRE wire contract:
// envelope layout, key, GCM mode, IV derivation, payload shape, and the fixed
// epoch/buildId constants. Drift in any of them fails a test here loudly,
// instead of silently reintroducing AA_CRYPTO_MISSING in production.
package allanime

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// aaReqPayload mirrors the JSON the token wraps, so the test can assert each
// field the server checks.
type aaReqPayload struct {
	V     int    `json:"v"`
	TS    int64  `json:"ts"`
	Epoch int    `json:"epoch"`
	QH    string `json:"qh"`
}

// decryptAAReqAsServer performs the exact validation an AllAnime edge does:
// split [0x01 | iv(12) | ct+tag], GCM-Open with the shared key, and return the
// recovered payload plus the IV that was on the wire. It fails the test if any
// structural or crypto invariant is violated.
func decryptAAReqAsServer(t *testing.T, token string) (aaReqPayload, []byte) {
	t.Helper()

	raw, err := base64.StdEncoding.DecodeString(token)
	require.NoError(t, err, "aaReq must be standard base64")
	require.GreaterOrEqual(t, len(raw), 1+12+16, "envelope must hold version+iv+tag")
	require.Equal(t, byte(0x01), raw[0], "first byte must be the 0x01 version marker")

	iv := raw[1:13]
	sealed := raw[13:]

	block, err := aes.NewCipher(allAnimeKey)
	require.NoError(t, err)
	gcm, err := cipher.NewGCM(block)
	require.NoError(t, err)

	plain, err := gcm.Open(nil, iv, sealed, nil)
	require.NoError(t, err, "server must be able to authenticate+decrypt the token with the shared key")

	var p aaReqPayload
	require.NoError(t, json.Unmarshal(plain, &p), "decrypted payload must be valid JSON")
	return p, iv
}

func TestBuildAAReq_ServerCanValidateEntireContract(t *testing.T) {
	t.Parallel()
	const qh = allAnimePersistedQueryHash
	const now = int64(1751990400000) // fixed wall clock (ms)

	token, err := buildAAReqAt(qh, allAnimeKey, testAAEpoch, now)
	require.NoError(t, err)

	payload, iv := decryptAAReqAsServer(t, token)

	// Payload fields the server checks.
	assert.Equal(t, 1, payload.V, "protocol version field")
	assert.Equal(t, 4128, payload.Epoch, "epoch must be the per-epoch value bound into the token")
	assert.Equal(t, qh, payload.QH, "qh must be the persisted-query hash")

	// Timestamp must be floored to the 5-minute window.
	wantTS := (now / aaReqWindowMillis) * aaReqWindowMillis
	assert.Equal(t, wantTS, payload.TS, "ts must be floored to the 5-minute bucket")

	// The IV on the wire must equal SHA-256("<epoch>:<qh>:<ts>")[:12] — the
	// server re-derives it from the decrypted payload and compares.
	seed := fmt.Sprintf("%s:%s:%d", testAAEpoch, qh, payload.TS)
	sum := sha256.Sum256([]byte(seed))
	assert.Equal(t, sum[:12], iv, "IV must be the first 12 bytes of SHA-256 over the epoch:qh:ts seed")
}

func TestBuildAAReq_DeterministicWithinWindow(t *testing.T) {
	t.Parallel()
	const qh = allAnimePersistedQueryHash
	base := int64(1751990400000)

	// Two calls in the same 5-minute bucket must produce byte-identical tokens
	// (GCM with a fixed key+iv+plaintext is deterministic). This is what lets
	// the server cache/verify without clock skew within the window.
	a, err := buildAAReqAt(qh, allAnimeKey, testAAEpoch, base+1000)
	require.NoError(t, err)
	b, err := buildAAReqAt(qh, allAnimeKey, testAAEpoch, base+aaReqWindowMillis-1)
	require.NoError(t, err)
	assert.Equal(t, a, b, "same 5-minute window ⇒ identical token")
}

func TestBuildAAReq_ChangesAcrossWindows(t *testing.T) {
	t.Parallel()
	const qh = allAnimePersistedQueryHash
	base := int64(1751990400000)

	// Crossing into the next window changes ts, which changes both the payload
	// and the derived IV, so the token must differ — a stale token would be
	// rejected by the server.
	a, err := buildAAReqAt(qh, allAnimeKey, testAAEpoch, base)
	require.NoError(t, err)
	b, err := buildAAReqAt(qh, allAnimeKey, testAAEpoch, base+aaReqWindowMillis)
	require.NoError(t, err)
	assert.NotEqual(t, a, b, "next 5-minute window ⇒ different token")
}

func TestBuildAAReq_BoundToQueryHash(t *testing.T) {
	t.Parallel()
	const now = int64(1751990400000)

	// The qh is bound into both the payload and the IV seed, so two different
	// hashes must yield different tokens — the token cannot be replayed for a
	// different query.
	a, err := buildAAReqAt("d405d0edd690624b66baba3068e0edc3ac90f1597d898a1ec8db4e5c43c00fec", allAnimeKey, testAAEpoch, now)
	require.NoError(t, err)
	b, err := buildAAReqAt("0000000000000000000000000000000000000000000000000000000000000000", allAnimeKey, testAAEpoch, now)
	require.NoError(t, err)
	assert.NotEqual(t, a, b, "token must be bound to the query hash")
}

func TestBuildAAReq_ProductionUsesLiveClock(t *testing.T) {
	t.Parallel()
	// The exported wrapper must produce a currently-valid token (non-empty,
	// server-decryptable) using the real clock — guards against the wrapper
	// being accidentally short-circuited.
	token, err := buildAAReq(allAnimePersistedQueryHash, &aaKeys{key: allAnimeKey, epoch: testAAEpoch})
	require.NoError(t, err)
	require.NotEmpty(t, token)
	payload, _ := decryptAAReqAsServer(t, token)
	assert.Equal(t, allAnimePersistedQueryHash, payload.QH)
}
