package updater

import (
	"bytes"
	"crypto/ed25519"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"golang.org/x/crypto/blake2b"
)

// checksumsSignatureAssetName is the minisign signature every release
// publishes for checksums-sha256.txt.
const checksumsSignatureAssetName = checksumsAssetName + ".minisig"

// trustedPublicKeys are the minisign public keys (the base64 line of a .pub
// file) whose signatures the updater accepts. The matching secret keys are
// kept offline by the maintainer, so a compromised GitHub account or release
// cannot produce an accepted manifest. Listing two keys lets a rotation ship
// in one release: the new key is added while the old one still signs.
//
// With no key listed every update is refused, never accepted unchecked.
var trustedPublicKeys = []string{
	"RWSy2CebiZzHgMdnf1JU4Ik3QlgVTQ+gpMKjbdE/dovVSgqaSM1N/5D0", // minisign public key 80C79C899B27D8B2
}

var (
	errNoTrustedKeys = errors.New("this build embeds no update-signing key, so updates cannot be verified")

	// minisign signature algorithms: the message is signed directly ("Ed") or
	// after a BLAKE2b-512 pre-hash ("ED"), the minisign default.
	sigAlgLegacy    = [2]byte{'E', 'd'}
	sigAlgPrehashed = [2]byte{'E', 'D'}
)

const (
	minisignKeyIDLen    = 8
	minisignTrustedMark = "trusted comment:"
)

// minisignPublicKey is a decoded minisign public key.
type minisignPublicKey struct {
	id  [minisignKeyIDLen]byte
	key ed25519.PublicKey
}

// parseMinisignPublicKey decodes a key given either as the bare base64 line or
// as a whole .pub file (with its "untrusted comment:" header).
func parseMinisignPublicKey(text string) (minisignPublicKey, error) {
	var pk minisignPublicKey

	raw, err := base64.StdEncoding.DecodeString(minisignBase64Line(text))
	if err != nil {
		return pk, fmt.Errorf("public key is not valid base64: %w", err)
	}
	if len(raw) != 2+minisignKeyIDLen+ed25519.PublicKeySize || raw[0] != 'E' || raw[1] != 'd' {
		return pk, errors.New("public key is not a minisign Ed25519 key")
	}

	copy(pk.id[:], raw[2:2+minisignKeyIDLen])
	pk.key = ed25519.PublicKey(bytes.Clone(raw[2+minisignKeyIDLen:]))
	return pk, nil
}

// minisignBase64Line returns the first line that is not a comment.
func minisignBase64Line(text string) string {
	for _, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && !strings.HasPrefix(line, "untrusted comment:") {
			return line
		}
	}
	return ""
}

// verifyMinisign reports whether sig is a valid minisign signature of message
// by one of keys. Both the signature over the message and the global signature
// over the trusted comment must verify: an attacker must not be able to swap
// the comment of a genuine signature.
func verifyMinisign(message, sig []byte, keys []string) error {
	if len(keys) == 0 {
		return errNoTrustedKeys
	}

	lines := strings.Split(strings.ReplaceAll(string(sig), "\r\n", "\n"), "\n")
	if len(lines) < 4 || !strings.HasPrefix(lines[0], "untrusted comment:") ||
		!strings.HasPrefix(lines[2], minisignTrustedMark) {
		return errors.New("signature file is malformed")
	}

	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(lines[1]))
	if err != nil || len(raw) != 2+minisignKeyIDLen+ed25519.SignatureSize {
		return errors.New("signature is malformed")
	}
	global, err := base64.StdEncoding.DecodeString(strings.TrimSpace(lines[3]))
	if err != nil || len(global) != ed25519.SignatureSize {
		return errors.New("global signature is malformed")
	}

	alg := [2]byte{raw[0], raw[1]}
	keyID := raw[2 : 2+minisignKeyIDLen]
	signature := raw[2+minisignKeyIDLen:]
	trusted := strings.TrimPrefix(lines[2], minisignTrustedMark)
	trusted = strings.TrimPrefix(trusted, " ")

	signed := message
	switch alg {
	case sigAlgLegacy:
	case sigAlgPrehashed:
		digest := blake2b.Sum512(message)
		signed = digest[:]
	default:
		return fmt.Errorf("unsupported signature algorithm %q", alg[:])
	}

	// Compare against every configured key, and report only after all were
	// tried so a malformed entry cannot hide a valid one.
	var keyErrs []error
	for _, text := range keys {
		pk, err := parseMinisignPublicKey(text)
		if err != nil {
			keyErrs = append(keyErrs, err)
			continue
		}
		if !bytes.Equal(pk.id[:], keyID) {
			continue
		}
		if !ed25519.Verify(pk.key, signed, signature) {
			return errors.New("signature does not match the manifest")
		}
		globalMsg := append(bytes.Clone(signature), trusted...)
		if !ed25519.Verify(pk.key, globalMsg, global) {
			return errors.New("trusted comment signature does not match")
		}
		return nil
	}

	if err := errors.Join(keyErrs...); err != nil {
		return fmt.Errorf("signed by an unknown key (and a configured key is invalid): %w", err)
	}
	return errors.New("signed by an unknown key")
}
