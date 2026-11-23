// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package v2

import (
	goerrors "errors"
	"fmt"
	"io"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/errors"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

// PublicKeyType is the armor type for a PGP public key.
var PublicKeyType = "PGP PUBLIC KEY BLOCK"

// PrivateKeyType is the armor type for a PGP private key.
var PrivateKeyType = "PGP PRIVATE KEY BLOCK"

// An Entity represents the components of an OpenPGP key: a primary public key
// (which must be a signing key), one or more identities claimed by that key,
// and zero or more subkeys, which may be encryption keys.
type Entity struct {
	PrimaryKey       *packet.PublicKey
	PrivateKey       *packet.PrivateKey
	Identities       map[string]*Identity // indexed by Identity.Name
	Revocations      []*packet.VerifiableSignature
	DirectSignatures []*packet.VerifiableSignature // Direct-key self signature of the PrimaryKey (contains primary key properties in v6)}
	Subkeys          []Subkey
}

// A Key identifies a specific public key in an Entity. This is either the
// Entity's primary key or a subkey.
type Key struct {
	Entity               *Entity
	PrimarySelfSignature *packet.Signature // might be nil, if not verified
	PublicKey            *packet.PublicKey
	PrivateKey           *packet.PrivateKey
	SelfSignature        *packet.Signature // might be nil, if not verified
}

// A KeyRing provides access to public and private keys.
type KeyRing interface {
	// KeysById returns the set of keys that have the given key id.
	// KeysById does not perform any signature validations and verification of the returned keys.
	KeysById(id uint64) []Key
	// EntitiesById returns the set of entities that contain a key with the given key id.
	// EntitiesById does not perform any signature validations and verification of the returned keys.
	EntitiesById(id uint64) []*Entity
}

// PrimaryIdentity returns a valid non-revoked Identity while preferring
// identities marked as primary, or the latest-created identity, in that order.
// Returns an nil for both return values if there is no valid primary identity.
func (e *Entity) PrimaryIdentity(date time.Time, config *packet.Config) (*packet.Signature, *Identity) {
	var primaryIdentityCandidates []*Identity
	var primaryIdentityCandidatesSelfSigs []*packet.Signature
	for _, identity := range e.Identities {
		selfSig, err := identity.Verify(date, config) // identity must be valid at date
		if err == nil {
			// verification is successful
			primaryIdentityCandidates = append(primaryIdentityCandidates, identity)
			primaryIdentityCandidatesSelfSigs = append(primaryIdentityCandidatesSelfSigs, selfSig)
		}
	}
	if len(primaryIdentityCandidates) == 0 {
		return nil, nil
	}
	primaryIdentity := -1
	for idx := range primaryIdentityCandidates {
		if primaryIdentity == -1 ||
			shouldPreferIdentity(primaryIdentityCandidatesSelfSigs[primaryIdentity],
				primaryIdentityCandidatesSelfSigs[idx]) {
			primaryIdentity = idx
		}
	}
	return primaryIdentityCandidatesSelfSigs[primaryIdentity], primaryIdentityCandidates[primaryIdentity]
}

func shouldPreferIdentity(existingId, potentialNewId *packet.Signature) bool {
	// Prefer identities that are marked as primary
	if existingId.IsPrimaryId != nil && *existingId.IsPrimaryId &&
		!(potentialNewId.IsPrimaryId != nil && *potentialNewId.IsPrimaryId) {
		return false
	}
	if !(existingId.IsPrimaryId != nil && *existingId.IsPrimaryId) &&
		potentialNewId.IsPrimaryId != nil && *potentialNewId.IsPrimaryId {
		return true
	}
	// after that newer creation time
	return potentialNewId.CreationTime.Unix() >= existingId.CreationTime.Unix()
}

// EncryptionKey returns the best candidate Key for encrypting a message to the
// given Entity.
func (e *Entity) EncryptionKey(now time.Time, config *packet.Config) (Key, bool) {
	encryptionKey, err := e.EncryptionKeyWithError(now, config)
	return encryptionKey, err == nil
}

// EncryptionKeyWithError returns the best candidate Key for encrypting a message to the
// given Entity.
// Provides an error if the function fails to find an encryption key.
func (e *Entity) EncryptionKeyWithError(now time.Time, config *packet.Config) (Key, error) {
	// The primary key has to be valid at time now
	primarySelfSignature, err := e.VerifyPrimaryKey(now, config)
	if err != nil { // primary key is not valid
		return Key{}, errors.ErrEncryptionKeySelection{
			PrimaryKeyId:  e.PrimaryKey.KeyIdString(),
			PrimaryKeyErr: err,
		}
	}

	if err := checkKeyRequirements(e.PrimaryKey, config); err != nil {
		// The primary key produces weak signatures
		return Key{}, errors.ErrEncryptionKeySelection{
			PrimaryKeyId:  e.PrimaryKey.KeyIdString(),
			PrimaryKeyErr: err,
		}
	}

	// Iterate the keys to find the newest, unexpired one
	var latestSelectionError *errors.ErrEncryptionKeySelection
	candidateSubkey := -1
	var maxTime time.Time
	var selectedSubkeySelfSig *packet.Signature
	for i, subkey := range e.Subkeys {
		subkeyErr := func(encSelectionErr error) *errors.ErrEncryptionKeySelection {
			subkeyKeyId := subkey.PublicKey.KeyIdString()
			return &errors.ErrEncryptionKeySelection{
				PrimaryKeyId:      e.PrimaryKey.KeyIdString(),
				EncSelectionKeyId: &subkeyKeyId,
				EncSelectionErr:   encSelectionErr,
			}

		}
		// Verify the subkey signature.
		subkeySelfSig, err := subkey.Verify(now, config) // subkey has to be valid at time now
		if err != nil {
			latestSelectionError = subkeyErr(err)
			continue
		}
		// Check the algorithm and key flags.
		if !isValidEncryptionKey(subkeySelfSig, subkey.PublicKey.PubKeyAlgo, config) {
			continue
		}
		// Check if the key fulfils the requirements
		if err := checkKeyRequirements(subkey.PublicKey, config); err != nil {
			latestSelectionError = subkeyErr(err)
			continue
		}
		preferNewSubkey := maxTime.IsZero() ||
			subkeySelfSig.CreationTime.After(maxTime) ||
			(subkeySelfSig.CreationTime.Equal(maxTime) && subkey.IsPQ())
		if preferNewSubkey {
			candidateSubkey = i
			selectedSubkeySelfSig = subkeySelfSig
			maxTime = subkeySelfSig.CreationTime
		}
	}

	if candidateSubkey != -1 {
		subkey := &e.Subkeys[candidateSubkey]
		return Key{
			Entity:               subkey.Primary,
			PrimarySelfSignature: primarySelfSignature,
			PublicKey:            subkey.PublicKey,
			PrivateKey:           subkey.PrivateKey,
			SelfSignature:        selectedSubkeySelfSig,
		}, nil
	}

	// If we don't have any subkeys for encryption and the primary key
	// is marked as OK to encrypt with, then we can use it.
	if isValidEncryptionKey(primarySelfSignature, e.PrimaryKey.PubKeyAlgo, config) {
		return Key{
			Entity:               e,
			PrimarySelfSignature: primarySelfSignature,
			PublicKey:            e.PrimaryKey,
			PrivateKey:           e.PrivateKey,
			SelfSignature:        primarySelfSignature,
		}, nil
	}

	if latestSelectionError == nil {
		latestSelectionError = &errors.ErrEncryptionKeySelection{
			PrimaryKeyId:    e.PrimaryKey.KeyIdString(),
			EncSelectionErr: goerrors.New("no encryption-capable key found (no key flags or invalid algorithm)"),
		}
	}

	return Key{}, latestSelectionError
}

// DecryptionKeys returns all keys that are available for decryption, matching the keyID when given
// If date is zero (i.e., date.IsZero() == true) the time checks are not performed,
// which should be proffered to decrypt older messages.
// If id is 0 all decryption keys are returned.
// This is useful to retrieve keys for session key decryption.
func (e *Entity) DecryptionKeys(id uint64, date time.Time, config *packet.Config) (keys []Key) {
	primarySelfSignature, err := e.PrimarySelfSignature(date, config)
	if err != nil { // primary key is not valid
		return
	}
	for _, subkey := range e.Subkeys {
		subkeySelfSig, err := subkey.LatestValidBindingSignature(date, config)
		if err == nil &&
			(config.AllowDecryptionWithSigningKeys() || isValidEncryptionKey(subkeySelfSig, subkey.PublicKey.PubKeyAlgo, config)) &&
			(id == 0 || subkey.PublicKey.KeyId == id) {
			keys = append(keys, Key{subkey.Primary, primarySelfSignature, subkey.PublicKey, subkey.PrivateKey, subkeySelfSig})
		}
	}
	if config.AllowDecryptionWithSigningKeys() || isValidEncryptionKey(primarySelfSignature, e.PrimaryKey.PubKeyAlgo, config) {
		keys = append(keys, Key{e, primarySelfSignature, e.PrimaryKey, e.PrivateKey, primarySelfSignature})
	}
	return
}

// CertificationKey return the best candidate Key for certifying a key with this
// Entity.
func (e *Entity) CertificationKey(now time.Time, config *packet.Config) (Key, bool) {
	return e.CertificationKeyById(now, 0, config)
}

// CertificationKeyById return the Key for key certification with this
// Entity and keyID.
func (e *Entity) CertificationKeyById(now time.Time, id uint64, config *packet.Config) (Key, bool) {
	key, err := e.signingKeyByIdUsage(now, id, packet.KeyFlagSign, config)
	return key, err == nil
}

// SigningKey return the best candidate Key for signing a message with this
// Entity.
func (e *Entity) SigningKey(now time.Time, config *packet.Config) (Key, bool) {
	return e.SigningKeyById(now, 0, config)
}

// SigningKeyById return the Key for signing a message with this
// Entity and keyID.
func (e *Entity) SigningKeyById(now time.Time, id uint64, config *packet.Config) (Key, bool) {
	key, err := e.signingKeyByIdUsage(now, id, packet.KeyFlagSign, config)
	return key, err == nil
}

func (e *Entity) signingKeyByIdUsage(now time.Time, id uint64, flags int, config *packet.Config) (Key, error) {
	primarySelfSignature, err := e.VerifyPrimaryKey(now, config)
	if err != nil {
		return Key{}, err
	}

	if err = checkKeyRequirements(e.PrimaryKey, config); err != nil {
		// The primary key produces weak signatures
		return Key{}, err
	}

	// Iterate the keys to find the newest, unexpired one.
	candidateSubkey := -1
	var maxTime time.Time
	var selectedSubkeySelfSig *packet.Signature
	for idx, subkey := range e.Subkeys {
		subkeySelfSig, err := subkey.Verify(now, config)
		if err == nil &&
			(flags&packet.KeyFlagCertify == 0 || isValidCertificationKey(subkeySelfSig, subkey.PublicKey.PubKeyAlgo, config)) &&
			(flags&packet.KeyFlagSign == 0 || isValidSigningKey(subkeySelfSig, subkey.PublicKey.PubKeyAlgo, config)) &&
			checkKeyRequirements(subkey.PublicKey, config) == nil &&
			(maxTime.IsZero() ||
				subkeySelfSig.CreationTime.After(maxTime) ||
				(subkeySelfSig.CreationTime.Equal(maxTime) && subkey.IsPQ())) &&
			(id == 0 || subkey.PublicKey.KeyId == id) {
			candidateSubkey = idx
			maxTime = subkeySelfSig.CreationTime
			selectedSubkeySelfSig = subkeySelfSig
		}
	}

	if candidateSubkey != -1 {
		subkey := &e.Subkeys[candidateSubkey]
		return Key{
			Entity:               subkey.Primary,
			PrimarySelfSignature: primarySelfSignature,
			PublicKey:            subkey.PublicKey,
			PrivateKey:           subkey.PrivateKey,
			SelfSignature:        selectedSubkeySelfSig,
		}, nil
	}

	// If we don't have any subkeys for signing and the primary key
	// is marked as OK to sign with, then we can use it.
	if (flags&packet.KeyFlagCertify == 0 || isValidCertificationKey(primarySelfSignature, e.PrimaryKey.PubKeyAlgo, config)) &&
		(flags&packet.KeyFlagSign == 0 || isValidSigningKey(primarySelfSignature, e.PrimaryKey.PubKeyAlgo, config)) &&
		(id == 0 || e.PrimaryKey.KeyId == id) {
		return Key{
			Entity:               e,
			PrimarySelfSignature: primarySelfSignature,
			PublicKey:            e.PrimaryKey,
			PrivateKey:           e.PrivateKey,
			SelfSignature:        primarySelfSignature,
		}, nil
	}

	// No keys with a valid Signing Flag or no keys matched the id passed in
	return Key{}, errors.StructuralError("no valid signing or verifying key found")
}

// Revoked returns whether the entity has any direct key revocation signatures.
// Note that third-party revocation signatures are not supported.
// Note also that Identity and Subkey revocation should be checked separately.
func (e *Entity) Revoked(now time.Time) bool {
	// Verify revocations first
	for _, revocation := range e.Revocations {
		if revocation.Valid == nil {
			err := e.PrimaryKey.VerifyRevocationSignature(revocation.Packet)
			valid := err == nil
			revocation.Valid = &valid
		}
		if *revocation.Valid &&
			(revocation.Packet.RevocationReason == nil ||
				*revocation.Packet.RevocationReason == packet.Unknown ||
				*revocation.Packet.RevocationReason == packet.NoReason ||
				*revocation.Packet.RevocationReason == packet.KeyCompromised) {
			// If the key is compromised, the key is considered revoked even before the revocation date.
			return true
		}
		if *revocation.Valid &&
			!revocation.Packet.SigExpired(now) {
			return true
		}
	}
	return false
}

// EncryptPrivateKeys encrypts all non-encrypted keys in the entity with the same key
// derived from the provided passphrase. Public keys and dummy keys are ignored,
// and don't cause an error to be returned.
func (e *Entity) EncryptPrivateKeys(passphrase []byte, config *packet.Config) error {
	var keysToEncrypt []*packet.PrivateKey
	// Add entity private key to encrypt.
	if e.PrivateKey != nil && !e.PrivateKey.Dummy() && !e.PrivateKey.Encrypted {
		keysToEncrypt = append(keysToEncrypt, e.PrivateKey)
	}

	// Add subkeys to encrypt.
	for _, sub := range e.Subkeys {
		if sub.PrivateKey != nil && !sub.PrivateKey.Dummy() && !sub.PrivateKey.Encrypted {
			keysToEncrypt = append(keysToEncrypt, sub.PrivateKey)
		}
	}
	return packet.EncryptPrivateKeys(keysToEncrypt, passphrase, config)
}

// DecryptPrivateKeys decrypts all encrypted keys in the entity with the given passphrase.
// Avoids recomputation of similar s2k key derivations. Public keys and dummy keys are ignored,
// and don't cause an error to be returned.
func (e *Entity) DecryptPrivateKeys(passphrase []byte) error {
	var keysToDecrypt []*packet.PrivateKey
	// Add entity private key to decrypt.
	if e.PrivateKey != nil && !e.PrivateKey.Dummy() && e.PrivateKey.Encrypted {
		keysToDecrypt = append(keysToDecrypt, e.PrivateKey)
	}

	// Add subkeys to decrypt.
	for _, sub := range e.Subkeys {
		if sub.PrivateKey != nil && !sub.PrivateKey.Dummy() && sub.PrivateKey.Encrypted {
			keysToDecrypt = append(keysToDecrypt, sub.PrivateKey)
		}
	}
	return packet.DecryptPrivateKeys(keysToDecrypt, passphrase)
}

// EntityList contains one or more Entities.
type EntityList []*Entity

// KeysById returns the set of keys that have the given key id.
// KeysById does not perform any key validation, and the self-signature
// fields in the returned key structs are nil.
func (el EntityList) KeysById(id uint64) (keys []Key) {
	for _, e := range el {
		if id == 0 || e.PrimaryKey.KeyId == id {
			keys = append(keys, Key{e, nil, e.PrimaryKey, e.PrivateKey, nil})
		}

		for _, subKey := range e.Subkeys {
			if id == 0 || subKey.PublicKey.KeyId == id {
				keys = append(keys, Key{subKey.Primary, nil, subKey.PublicKey, subKey.PrivateKey, nil})
			}
		}
	}
	return
}

// EntitiesById returns the entities that contain a key with the given key id.
func (el EntityList) EntitiesById(id uint64) (entities []*Entity) {
	for _, e := range el {
		if id == 0 || e.PrimaryKey.KeyId == id {
			entities = append(entities, e)
			continue
		}

		for _, subKey := range e.Subkeys {
			if id == 0 || subKey.PublicKey.KeyId == id {
				entities = append(entities, e)
				continue
			}
		}
	}
	return
}

// ReadArmoredKeyRing reads one or more public/private keys from an armor keyring file.
func ReadArmoredKeyRing(r io.Reader) (EntityList, error) {
	block, err := armor.Decode(r)
	if err == io.EOF {
		return nil, errors.InvalidArgumentError("no armored data found")
	}
	if err != nil {
		return nil, err
	}
	if block.Type != PublicKeyType && block.Type != PrivateKeyType {
		return nil, errors.InvalidArgumentError("expected public or private key block, got: " + block.Type)
	}

	return ReadKeyRing(block.Body)
}

// ReadKeyRing reads one or more public/private keys. Unsupported keys are
// ignored as long as at least a single valid key is found.
func ReadKeyRing(r io.Reader) (el EntityList, err error) {
	packets := packet.NewReader(r)
	var lastUnsupportedError error

	for {
		var e *Entity
		e, err = ReadEntity(packets)
		if err != nil {
			// TODO: warn about skipped unsupported/unreadable keys
			if _, ok := err.(errors.UnsupportedError); ok {
				lastUnsupportedError = err
				err = readToNextPublicKey(packets)
			} else if _, ok := err.(errors.StructuralError); ok {
				// Skip unreadable, badly-formatted keys
				lastUnsupportedError = err
				err = readToNextPublicKey(packets)
			}
			if err == io.EOF {
				err = nil
				break
			}
			if err != nil {
				el = nil
				break
			}
		} else {
			el = append(el, e)
		}
	}

	if len(el) == 0 && err == nil {
		err = lastUnsupportedError
	}
	return
}

// readToNextPublicKey reads packets until the start of the entity and leaves
// the first packet of the new entity in the Reader.
func readToNextPublicKey(packets *packet.Reader) (err error) {
	var p packet.Packet
	for {
		p, err = packets.Next()
		if err == io.EOF {
			return
		} else if err != nil {
			if _, ok := err.(errors.UnsupportedError); ok {
				continue
			}
			return
		}

		if pk, ok := p.(*packet.PublicKey); ok && !pk.IsSubkey {
			packets.Unread(p)
			return
		}
	}
}

// ReadEntity reads an entity (public key, identities, subkeys etc) from the
// given Reader.
func ReadEntity(packets *packet.Reader) (*Entity, error) {
	e := new(Entity)
	e.Identities = make(map[string]*Identity)

	p, err := packets.Next()
	if err != nil {
		return nil, err
	}

	var ok bool
	if e.PrimaryKey, ok = p.(*packet.PublicKey); !ok {
		if e.PrivateKey, ok = p.(*packet.PrivateKey); !ok {
			packets.Unread(p)
			return nil, errors.StructuralError("first packet was not a public/private key")
		}
		e.PrimaryKey = &e.PrivateKey.PublicKey
	}

	if !e.PrimaryKey.PubKeyAlgo.CanSign() {
		return nil, errors.StructuralError("primary key cannot be used for signatures")
	}
	var ignoreSigs bool
EachPacket:
	for {
		p, err := packets.NextWithUnsupported()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		var unsupported bool
		if unsupportedPacket, ok := p.(*packet.UnsupportedPacket); ok {
			unsupported = true
			p = unsupportedPacket.IncompletePacket
		}

		// Handle unsupported keys
		switch p.(type) {
		case *packet.PublicKey, *packet.PrivateKey:
			if unsupported {
				// Skip following signature packets
				ignoreSigs = true
			}
		case *packet.Signature:
			if ignoreSigs {
				continue
			}
		default:
			ignoreSigs = false
		}
		// Unsupported packages are handled continue
		// if the packet is unsupported
		if unsupported {
			continue
		}

		switch pkt := p.(type) {
		case *packet.UserId:
			err := readUser(e, packets, pkt)
			if err != nil {
				return nil, err
			}
		case *packet.Signature:
			if pkt.SigType == packet.SigTypeKeyRevocation {
				e.Revocations = append(e.Revocations, packet.NewVerifiableSig(pkt))
			} else if pkt.SigType == packet.SigTypeDirectSignature {
				e.DirectSignatures = append(e.DirectSignatures, packet.NewVerifiableSig(pkt))
			}
			// Else, ignoring the signature as it does not follow anything
			// we would know to attach it to.
		case *packet.PrivateKey:
			if !pkt.IsSubkey {
				packets.Unread(p)
				break EachPacket
			}
			err = readSubkey(e, packets, &pkt.PublicKey, pkt)
			if err != nil {
				return nil, err
			}
		case *packet.PublicKey:
			if !pkt.IsSubkey {
				packets.Unread(p)
				break EachPacket
			}
			err = readSubkey(e, packets, pkt, nil)
			if err != nil {
				return nil, err
			}
		default:
			// we ignore unknown packets
		}
	}

	if len(e.Identities) == 0 && e.PrimaryKey.Version < 6 {
		return nil, errors.StructuralError("v4 entity without any identities")
	}

	if e.PrimaryKey.Version == 6 && len(e.DirectSignatures) == 0 {
		return nil, errors.StructuralError("v6 entity without a  direct-key signature")
	}
	return e, nil
}

// SerializePrivate serializes an Entity, including private key material, but
// excluding signatures from other entities, to the given Writer.
// Identities and subkeys are re-signed in case they changed since NewEntry.
// If config is nil, sensible defaults will be used.
func (e *Entity) SerializePrivate(w io.Writer, config *packet.Config) (err error) {
	if e.PrivateKey.Dummy() {
		return errors.ErrDummyPrivateKey("dummy private key cannot re-sign identities")
	}
	return e.serializePrivate(w, config, true)
}

// SerializePrivateWithoutSigning serializes an Entity, including private key
// material, but excluding signatures from other entities, to the given Writer.
// Self-signatures of identities and subkeys are not re-signed. This is useful
// when serializing GNU dummy keys, among other things.
// If config is nil, sensible defaults will be used.
func (e *Entity) SerializePrivateWithoutSigning(w io.Writer, config *packet.Config) (err error) {
	return e.serializePrivate(w, config, false)
}

func (e *Entity) serializePrivate(w io.Writer, config *packet.Config, reSign bool) (err error) {
	if e.PrivateKey == nil {
		return goerrors.New("openpgp: private key is missing")
	}
	err = e.PrivateKey.Serialize(w)
	if err != nil {
		return
	}
	for _, revocation := range e.Revocations {
		if err = revocation.Packet.Serialize(w); err != nil {
			return err
		}
	}
	for _, directSignature := range e.DirectSignatures {
		if err = directSignature.Packet.Serialize(w); err != nil {
			return err
		}
	}
	for _, ident := range e.Identities {
		if reSign {
			if err = ident.ReSign(config); err != nil {
				return err
			}
		}
		if err = ident.Serialize(w); err != nil {
			return err
		}
	}
	for _, subkey := range e.Subkeys {
		if reSign {
			if err := subkey.ReSign(config); err != nil {
				return err
			}
		}
		if err = subkey.Serialize(w, true); err != nil {
			return err
		}
	}
	return nil
}

// Serialize writes the public part of the given Entity to w, including
// signatures from other entities. No private key material will be output.
func (e *Entity) Serialize(w io.Writer) error {
	if err := e.PrimaryKey.Serialize(w); err != nil {
		return err
	}
	for _, revocation := range e.Revocations {
		if err := revocation.Packet.Serialize(w); err != nil {
			return err
		}
	}
	for _, directSignature := range e.DirectSignatures {
		err := directSignature.Packet.Serialize(w)
		if err != nil {
			return err
		}
	}
	for _, ident := range e.Identities {
		if err := ident.Serialize(w); err != nil {
			return err
		}
	}
	for _, subkey := range e.Subkeys {
		if err := subkey.Serialize(w, false); err != nil {
			return err
		}
	}
	return nil
}

// Revoke generates a key revocation signature (packet.SigTypeKeyRevocation) with the
// specified reason code and text (RFC4880 section-5.2.3.23).
// If config is nil, sensible defaults will be used.
func (e *Entity) Revoke(reason packet.ReasonForRevocation, reasonText string, config *packet.Config) error {
	revSig := createSignaturePacket(e.PrimaryKey, packet.SigTypeKeyRevocation, config)
	revSig.RevocationReason = &reason
	revSig.RevocationReasonText = reasonText

	if err := revSig.RevokeKey(e.PrimaryKey, e.PrivateKey, config); err != nil {
		return err
	}
	sig := packet.NewVerifiableSig(revSig)
	valid := true
	sig.Valid = &valid
	e.Revocations = append(e.Revocations, sig)
	return nil
}

// SignIdentity adds a signature to e, from signer, attesting that identity is
// associated with e. The provided identity must already be an element of
// e.Identities and the private key of signer must have been decrypted if
// necessary.
// If config is nil, sensible defaults will be used.
func (e *Entity) SignIdentity(identity string, signer *Entity, config *packet.Config) error {
	ident, ok := e.Identities[identity]
	if !ok {
		return errors.InvalidArgumentError("given identity string not found in Entity")
	}
	return ident.SignIdentity(signer, config)
}

// LatestValidDirectSignature returns the latest valid direct key-signature of the entity.
func (e *Entity) LatestValidDirectSignature(date time.Time, config *packet.Config) (selectedSig *packet.Signature, err error) {
	for sigIdx := len(e.DirectSignatures) - 1; sigIdx >= 0; sigIdx-- {
		sig := e.DirectSignatures[sigIdx]
		if (date.IsZero() || date.Unix() >= sig.Packet.CreationTime.Unix()) &&
			(selectedSig == nil || selectedSig.CreationTime.Unix() < sig.Packet.CreationTime.Unix()) {
			if sig.Valid == nil {
				err := e.PrimaryKey.VerifyDirectKeySignature(sig.Packet)
				if err == nil {
					err = checkSignatureDetails(sig.Packet, date, config)
				}
				valid := err == nil
				sig.Valid = &valid
			}
			if *sig.Valid && (date.IsZero() || !sig.Packet.SigExpired(date)) {
				selectedSig = sig.Packet
			}
		}
	}
	if selectedSig == nil {
		return nil, errors.StructuralError("no valid direct key signature found")
	}
	return selectedSig, nil
}

// PrimarySelfSignature searches the entity for the self-signature that stores key preferences.
// For V4 keys, returns the self-signature of the primary identity, and the identity.
// For V6 keys, returns the latest valid direct-key self-signature, and no identity (nil).
// This self-signature is to be used to check the key expiration,
// algorithm preferences, and so on.
func (e *Entity) PrimarySelfSignature(date time.Time, config *packet.Config) (primarySig *packet.Signature, err error) {
	if e.PrimaryKey.Version == 6 {
		return e.LatestValidDirectSignature(date, config)
	}
	primarySig, _ = e.PrimaryIdentity(date, config)
	if primarySig == nil {
		return nil, errors.StructuralError("no primary identity found")
	}
	return
}

// VerifyPrimaryKey checks if the primary key is valid by checking:
// - that the primary key is has not been revoked at the given date,
// - that there is valid non-expired self-signature,
// - that the primary key is not expired given its self-signature.
// If date is zero (i.e., date.IsZero() == true) the time checks are not performed.
func (e *Entity) VerifyPrimaryKey(date time.Time, config *packet.Config) (*packet.Signature, error) {
	primarySelfSignature, err := e.PrimarySelfSignature(date, config)
	if err != nil {
		return nil, goerrors.New("no valid self signature found")
	}
	// check for key revocation signatures
	if e.Revoked(date) {
		return nil, errors.ErrKeyRevoked
	}

	if !date.IsZero() && (e.PrimaryKey.KeyExpired(primarySelfSignature, date) || // primary key has expired
		primarySelfSignature.SigExpired(date)) { // self-signature has expired
		return primarySelfSignature, errors.ErrKeyExpired
	}

	if e.PrimaryKey.Version != 6 && len(e.DirectSignatures) > 0 {
		// check for expiration time in direct signatures (for V6 keys, the above already did so)
		primaryDirectKeySignature, _ := e.LatestValidDirectSignature(date, config)
		if primaryDirectKeySignature != nil &&
			(!date.IsZero() && e.PrimaryKey.KeyExpired(primaryDirectKeySignature, date)) {
			return primarySelfSignature, errors.ErrKeyExpired
		}
	}
	return primarySelfSignature, nil
}

func (k *Key) IsPrimary() bool {
	if k.PrimarySelfSignature == nil || k.SelfSignature == nil {
		return k.PublicKey == k.Entity.PrimaryKey
	}
	return k.PrimarySelfSignature == k.SelfSignature
}

func checkKeyRequirements(usedKey *packet.PublicKey, config *packet.Config) error {
	algo := usedKey.PubKeyAlgo
	if config.RejectPublicKeyAlgorithm(algo) {
		return errors.WeakAlgorithmError("public key algorithm " + string(algo))
	}
	switch algo {
	case packet.PubKeyAlgoRSA, packet.PubKeyAlgoRSASignOnly:
		length, err := usedKey.BitLength()
		if err != nil || length < config.MinimumRSABits() {
			return errors.WeakAlgorithmError(fmt.Sprintf("minimum rsa length is %d got %d", config.MinimumRSABits(), length))
		}
	case packet.PubKeyAlgoECDH, packet.PubKeyAlgoEdDSA, packet.PubKeyAlgoECDSA:
		curve, err := usedKey.Curve()
		if err != nil || config.RejectCurve(curve) {
			return errors.WeakAlgorithmError("elliptic curve " + curve)
		}
		if usedKey.Version == 6 && (curve == packet.Curve25519 || curve == packet.Curve448) {
			// Implementations MUST NOT accept or generate v6 key material using the deprecated OIDs.
			return errors.StructuralError("v6 key uses legacy elliptic curve " + curve)
		}
	}
	return nil
}

func isValidSigningKey(signature *packet.Signature, algo packet.PublicKeyAlgorithm, config *packet.Config) bool {
	if !algo.CanSign() {
		return false
	}

	if signature.FlagsValid {
		return signature.FlagSign
	}

	return config.AllowAllKeyFlagsWhenMissing()
}

func isValidCertificationKey(signature *packet.Signature, algo packet.PublicKeyAlgorithm, config *packet.Config) bool {
	if !algo.CanSign() {
		return false
	}

	if signature.FlagsValid {
		return signature.FlagCertify
	}

	return config.AllowAllKeyFlagsWhenMissing()
}

func isValidEncryptionKey(signature *packet.Signature, algo packet.PublicKeyAlgorithm, config *packet.Config) bool {
	if !algo.CanEncrypt() {
		return false
	}

	if signature.FlagsValid {
		return signature.FlagEncryptCommunications || signature.FlagEncryptStorage
	}

	return config.AllowAllKeyFlagsWhenMissing()
}
