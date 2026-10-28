package updater

import (
	"crypto/ed25519"
	"crypto/rand"
	"encoding/base64"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/blake2b"
)

// testSigner produces signatures in the exact byte layout minisign writes.
type testSigner struct {
	id   [minisignKeyIDLen]byte
	priv ed25519.PrivateKey
	pub  ed25519.PublicKey
}

func newTestSigner(t *testing.T) *testSigner {
	t.Helper()
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)
	s := &testSigner{priv: priv, pub: pub}
	_, err = rand.Read(s.id[:])
	require.NoError(t, err)
	return s
}

// publicKey is the base64 line of the matching .pub file.
func (s *testSigner) publicKey() string {
	raw := append([]byte("Ed"), s.id[:]...)
	return base64.StdEncoding.EncodeToString(append(raw, s.pub...))
}

func (s *testSigner) sign(message []byte, alg [2]byte) []byte {
	return s.signWithComment(message, alg, "timestamp:1 file:checksums-sha256.txt")
}

func (s *testSigner) signWithComment(message []byte, alg [2]byte, comment string) []byte {
	signed := message
	if alg == sigAlgPrehashed {
		d := blake2b.Sum512(message)
		signed = d[:]
	}
	sig := ed25519.Sign(s.priv, signed)
	global := ed25519.Sign(s.priv, append(append([]byte{}, sig...), comment...))

	raw := append(append(append([]byte{}, alg[:]...), s.id[:]...), sig...)
	return []byte("untrusted comment: signature from minisign secret key\n" +
		base64.StdEncoding.EncodeToString(raw) + "\n" +
		"trusted comment: " + comment + "\n" +
		base64.StdEncoding.EncodeToString(global) + "\n")
}

// withTestSigner makes a fresh key the only trusted key until cleanup. MUST NOT
// be used with t.Parallel() since it mutates a package var.
func withTestSigner(t *testing.T) *testSigner {
	t.Helper()
	s := newTestSigner(t)
	prev := trustedPublicKeys
	trustedPublicKeys = []string{s.publicKey()}
	t.Cleanup(func() { trustedPublicKeys = prev })
	return s
}

func TestVerifyMinisign(t *testing.T) {
	t.Parallel()
	good := newTestSigner(t)
	other := newTestSigner(t)
	msg := []byte(sha256Hex("bin") + "  goanime-linux-amd64\n")
	keys := []string{good.publicKey()}

	tests := []struct {
		name    string
		message []byte
		sig     []byte
		keys    []string
		wantErr string
	}{
		{name: "prehashed signature", message: msg, sig: good.sign(msg, sigAlgPrehashed), keys: keys},
		{name: "legacy signature", message: msg, sig: good.sign(msg, sigAlgLegacy), keys: keys},
		{
			name:    "manifest altered after signing",
			message: []byte(sha256Hex("evil") + "  goanime-linux-amd64\n"),
			sig:     good.sign(msg, sigAlgPrehashed), keys: keys,
			wantErr: "does not match the manifest",
		},
		{
			name:    "signed by another key",
			message: msg, sig: other.sign(msg, sigAlgPrehashed), keys: keys,
			wantErr: "unknown key",
		},
		{
			name:    "same key id but forged signature",
			message: msg, sig: forgeWithID(good, msg), keys: keys,
			wantErr: "does not match the manifest",
		},
		{
			name:    "no key embedded",
			message: msg, sig: good.sign(msg, sigAlgPrehashed), keys: nil,
			wantErr: "embeds no update-signing key",
		},
		{name: "empty signature", message: msg, sig: nil, keys: keys, wantErr: "malformed"},
		{name: "garbage signature", message: msg, sig: []byte("<html>404</html>"), keys: keys, wantErr: "malformed"},
		{
			name:    "second key of a rotation verifies",
			message: msg, sig: other.sign(msg, sigAlgPrehashed),
			keys: []string{good.publicKey(), other.publicKey()},
		},
		{
			name:    "bad entry does not hide a valid key",
			message: msg, sig: good.sign(msg, sigAlgPrehashed),
			keys: []string{"not-base64!", good.publicKey()},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			err := verifyMinisign(tt.message, tt.sig, tt.keys)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}

// forgeWithID signs with a different key but claims good's key id.
func forgeWithID(good *testSigner, msg []byte) []byte {
	forger := &testSigner{id: good.id}
	forger.pub, forger.priv, _ = ed25519.GenerateKey(rand.Reader)
	return forger.sign(msg, sigAlgPrehashed)
}

func TestVerifyMinisign_TrustedCommentIsAuthenticated(t *testing.T) {
	t.Parallel()
	s := newTestSigner(t)
	msg := []byte("manifest")
	sig := string(s.sign(msg, sigAlgPrehashed))

	swapped := strings.Replace(sig, "file:checksums-sha256.txt", "file:other", 1)
	require.NotEqual(t, sig, swapped)

	err := verifyMinisign(msg, []byte(swapped), []string{s.publicKey()})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "trusted comment")
}

func TestParseMinisignPublicKey(t *testing.T) {
	t.Parallel()
	s := newTestSigner(t)

	for name, text := range map[string]string{
		"bare line": s.publicKey(),
		"pub file":  "untrusted comment: minisign public key ABCD\n" + s.publicKey() + "\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			pk, err := parseMinisignPublicKey(text)
			require.NoError(t, err)
			assert.Equal(t, s.id, pk.id)
			assert.Equal(t, s.pub, pk.key)
		})
	}

	_, err := parseMinisignPublicKey(base64.StdEncoding.EncodeToString([]byte("too short")))
	require.Error(t, err)
}

// TestVerifyDownload_RejectsSwappedAssetWithMatchingChecksum is the attack the
// signature exists for: a compromised release replaces the binary and the
// checksums file together, so the digest matches but the signature cannot.
func TestVerifyDownload_RejectsSwappedAssetWithMatchingChecksum(t *testing.T) {
	const asset = "goanime-linux-amd64"
	signer := withTestSigner(t)
	genuine := []byte(sha256Hex("genuine") + "  " + asset + "\n")
	signed := signer.sign(genuine, sigAlgPrehashed)

	path := writeTemp(t, asset, "malware")
	attacker := []byte(sha256Hex("malware") + "  " + asset + "\n")
	withFetchChecksums(t, func(url string) ([]byte, error) {
		if strings.HasSuffix(url, ".minisig") {
			return signed, nil // the maintainer's real signature, reused
		}
		return attacker, nil
	})
	release := newRelease("v9.9.9",
		[2]string{checksumsAssetName, "https://x/checksums-sha256.txt"},
		[2]string{checksumsSignatureAssetName, "https://x/checksums-sha256.txt.minisig"})

	err := verifyDownload(release, path, asset)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "signature check failed")
}

func TestVerifyDownload_RefusesEverythingWithoutAKey(t *testing.T) {
	const asset = "goanime-linux-amd64"
	s := newTestSigner(t)
	prev := trustedPublicKeys
	trustedPublicKeys = nil
	t.Cleanup(func() { trustedPublicKeys = prev })

	path := writeTemp(t, asset, "bin")
	manifest := []byte(sha256Hex("bin") + "  " + asset + "\n")
	withFetchChecksums(t, func(url string) ([]byte, error) {
		if strings.HasSuffix(url, ".minisig") {
			return s.sign(manifest, sigAlgPrehashed), nil
		}
		return manifest, nil
	})
	release := newRelease("v9.9.9",
		[2]string{checksumsAssetName, "https://x/checksums-sha256.txt"},
		[2]string{checksumsSignatureAssetName, "https://x/checksums-sha256.txt.minisig"})

	err := verifyDownload(release, path, asset)

	require.Error(t, err)
	assert.ErrorContains(t, err, "embeds no update-signing key")
}

// TestVerifyMinisign_RealMinisign checks interoperability with the actual
// tool, since the helper above only proves the code agrees with itself.
func TestVerifyMinisign_RealMinisign(t *testing.T) {
	bin, err := exec.LookPath("minisign")
	if err != nil {
		t.Skip("minisign is not installed")
	}
	dir := t.TempDir()
	msgPath := filepath.Join(dir, "checksums-sha256.txt")
	msg := []byte(sha256Hex("bin") + "  goanime-linux-amd64\n")
	require.NoError(t, os.WriteFile(msgPath, msg, 0o600))

	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(bin, args...) // #nosec G204 -- fixed binary from LookPath, test-only arguments
		cmd.Dir = dir
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, string(out))
	}
	run("-G", "-W", "-p", "t.pub", "-s", "t.key")
	run("-S", "-s", "t.key", "-m", msgPath)

	pub, err := os.ReadFile(filepath.Join(dir, "t.pub"))
	require.NoError(t, err)
	sig, err := os.ReadFile(msgPath + ".minisig")
	require.NoError(t, err)

	require.NoError(t, verifyMinisign(msg, sig, []string{string(pub)}))
	require.Error(t, verifyMinisign([]byte("tampered"), sig, []string{string(pub)}))
}
