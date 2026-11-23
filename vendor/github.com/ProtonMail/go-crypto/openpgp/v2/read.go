// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package openpgp implements high level operations on OpenPGP messages.
package v2

import (
	"bytes"
	"crypto"
	_ "crypto/sha256"
	_ "crypto/sha512"
	"hash"
	"io"
	"io/ioutil"
	"strconv"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/errors"
	"github.com/ProtonMail/go-crypto/openpgp/internal/algorithm"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
	_ "golang.org/x/crypto/sha3"
)

// SignatureType is the armor type for a PGP signature.
var SignatureType = "PGP SIGNATURE"

// MessageType is the armor type for a PGP message.
var MessageType = "PGP MESSAGE"

// readArmored reads an armored block with the given type.
func readArmored(r io.Reader, expectedType string) (body io.Reader, err error) {
	block, err := armor.Decode(r)
	if err != nil {
		return
	}

	if block.Type != expectedType {
		return nil, errors.InvalidArgumentError("expected '" + expectedType + "', got: " + block.Type)
	}

	return block.Body, nil
}

// MessageDetails contains the result of parsing an OpenPGP encrypted and/or
// signed message.
type MessageDetails struct {
	IsEncrypted              bool                  // true if the message was encrypted.
	EncryptedToKeyIds        []uint64              // the list of recipient key ids.
	IsSymmetricallyEncrypted bool                  // true if a passphrase could have decrypted the message.
	DecryptedWith            Key                   // the private key used to decrypt the message, if any.
	DecryptedWithAlgorithm   packet.CipherFunction // Stores the algorithm used to decrypt the message, if any.
	IsSigned                 bool                  // true if the message is signed.
	LiteralData              *packet.LiteralData   // the metadata of the contents
	UnverifiedBody           io.Reader             // the contents of the message.
	CheckRecipients          bool                  // Indicates if the intended recipients should be checked

	SessionKey []byte // Caches the session key if the flag in packet.Config is set to true and a session key was present.

	// If IsSigned is true then the signature candidates will
	// be verified as UnverifiedBody is read. The signature cannot be
	// checked until the whole of UnverifiedBody is read so UnverifiedBody
	// must be consumed until EOF before the data can be trusted. Even if a
	// message isn't signed (or the signer is unknown) the data may contain
	// an authentication code that is only checked once UnverifiedBody has
	// been consumed. Once EOF has been seen, the following fields are
	// valid. (An authentication code failure is reported as a
	// SignatureError error when reading from UnverifiedBody.)
	IsVerified          bool                  // true if the signatures have been verified else false
	SignatureCandidates []*SignatureCandidate // stores state for all signatures of this message
	SignedBy            *Key                  // the issuer key of the fist successfully verified signature, if any found.
	Signature           *packet.Signature     // the first successfully verified signature, if any found.
	// SignatureError is nil if one of the signatures in the message verifies successfully
	// else it points to the last observed signature error.
	// The error of each signature verification can be inspected by iterating trough
	// SignatureCandidates.
	SignatureError error
	// SelectedCandidate points to the signature candidate the SignatureError error stems from or
	// the selected successfully verified signature candidate.
	SelectedCandidate *SignatureCandidate

	decrypted io.ReadCloser
}

// A PromptFunction is used as a callback by functions that may need to decrypt
// a private key, or prompt for a passphrase. It is called with a list of
// acceptable, encrypted private keys and a boolean that indicates whether a
// passphrase is usable. It should either decrypt a private key or return a
// passphrase to try. If the decrypted private key or given passphrase isn't
// correct, the function will be called again, forever. Any error returned will
// be passed up.
type PromptFunction func(keys []Key, symmetric bool) ([]byte, error)

// A keyEnvelopePair is used to store a private key with the envelope that
// contains a symmetric key, encrypted with that key.
type keyEnvelopePair struct {
	key          Key
	encryptedKey *packet.EncryptedKey
}

// ReadMessage parses an OpenPGP message that may be signed and/or encrypted.
// The given KeyRing should contain both public keys (for signature
// verification) and, possibly encrypted, private keys for decrypting.
// If config is nil, sensible defaults will be used.
func ReadMessage(r io.Reader, keyring KeyRing, prompt PromptFunction, config *packet.Config) (md *MessageDetails, err error) {
	var p packet.Packet

	var symKeys []*packet.SymmetricKeyEncrypted
	var pubKeys []keyEnvelopePair
	// Integrity protected encrypted packet: SymmetricallyEncrypted or AEADEncrypted
	var edp packet.EncryptedDataPacket
	var packets packet.PacketReader
	if config.StrictPacketSequence() {
		packets = packet.NewCheckReader(r)
	} else {
		packets = packet.NewReader(r)
	}
	md = new(MessageDetails)
	md.IsEncrypted = true
	md.CheckRecipients = config.IntendedRecipients()

	// The message, if encrypted, starts with a number of packets
	// containing an encrypted decryption key. The decryption key is either
	// encrypted to a public key, or with a passphrase. This loop
	// collects these packets.
ParsePackets:
	for {
		p, err = packets.Next()
		if err != nil {
			return nil, err
		}
		switch p := p.(type) {
		case *packet.SymmetricKeyEncrypted:
			// This packet contains the decryption key encrypted with a passphrase.
			md.IsSymmetricallyEncrypted = true
			symKeys = append(symKeys, p)
		case *packet.EncryptedKey:
			// This packet contains the decryption key encrypted to a public key.
			md.EncryptedToKeyIds = append(md.EncryptedToKeyIds, p.KeyId)
			switch p.Algo {
			case packet.PubKeyAlgoRSA, packet.PubKeyAlgoRSAEncryptOnly, packet.PubKeyAlgoElGamal, packet.PubKeyAlgoECDH,
				packet.PubKeyAlgoX25519, packet.PubKeyAlgoX448,
				packet.PubKeyAlgoMlkem768X25519, packet.PubKeyAlgoMlkem1024X448:
				break
			default:
				continue
			}
			if keyring != nil {
				unverifiedEntities := keyring.EntitiesById(p.KeyId)
				for _, unverifiedEntity := range unverifiedEntities {
					// Do not check key expiration to allow decryption of old messages
					keys := unverifiedEntity.DecryptionKeys(p.KeyId, time.Time{}, config)
					for _, key := range keys {
						pubKeys = append(pubKeys, keyEnvelopePair{key, p})
					}
				}
			}
		case *packet.SymmetricallyEncrypted:
			if !p.IntegrityProtected && !config.AllowUnauthenticatedMessages() {
				return nil, errors.UnsupportedError("message is not integrity protected")
			}
			edp = p
			if p.Version == 2 { // SEIPD v2 packet stores the decryption algorithm
				md.DecryptedWithAlgorithm = p.Cipher
			}
			break ParsePackets
		case *packet.AEADEncrypted:
			edp = p
			break ParsePackets
		case *packet.Compressed, *packet.LiteralData, *packet.OnePassSignature, *packet.Signature:
			// This message isn't encrypted.
			if len(symKeys) != 0 || len(pubKeys) != 0 {
				return nil, errors.StructuralError("key material not followed by encrypted message")
			}
			packets.Unread(p)
			md.IsEncrypted = false
			return readSignedMessage(packets, md, keyring, config)
		}
	}

	var candidates []Key
	var decrypted io.ReadCloser

	// Now that we have the list of encrypted keys we need to decrypt at
	// least one of them or, if we cannot, we need to call the prompt
	// function so that it can decrypt a key or give us a passphrase.
FindKey:
	for {
		// See if any of the keys already have a private key available
		candidates = candidates[:0]
		candidateFingerprints := make(map[string]bool)

		for _, pk := range pubKeys {
			if pk.key.PrivateKey == nil {
				continue
			}
			if !pk.key.PrivateKey.Encrypted {
				if len(pk.encryptedKey.Key) == 0 {
					errDec := pk.encryptedKey.Decrypt(pk.key.PrivateKey, config)
					if errDec != nil {
						continue
					}
				}
				// Try to decrypt symmetrically encrypted
				decrypted, err = edp.Decrypt(pk.encryptedKey.CipherFunc, pk.encryptedKey.Key)
				if err != nil && err != errors.ErrKeyIncorrect {
					return nil, err
				}
				if decrypted != nil {
					md.DecryptedWith = pk.key
					if md.DecryptedWithAlgorithm == 0 { // if no SEIPD v2 packet, key packet stores the decryption algorithm
						md.DecryptedWithAlgorithm = pk.encryptedKey.CipherFunc
					}
					if config.RetrieveSessionKey() {
						md.SessionKey = pk.encryptedKey.Key
					}
					break FindKey
				}
			} else {
				fpr := string(pk.key.PublicKey.Fingerprint[:])
				if v := candidateFingerprints[fpr]; v {
					continue
				}
				candidates = append(candidates, pk.key)
				candidateFingerprints[fpr] = true
			}
		}

		if len(candidates) == 0 && len(symKeys) == 0 {
			return nil, errors.ErrKeyIncorrect
		}

		if prompt == nil {
			return nil, errors.ErrKeyIncorrect
		}

		passphrase, err := prompt(candidates, len(symKeys) != 0)
		if err != nil {
			return nil, err
		}

		// Try the symmetric passphrase first
		if len(symKeys) != 0 && passphrase != nil {
			for _, s := range symKeys {
				key, cipherFunc, err := s.Decrypt(passphrase)
				// In v4, on wrong passphrase, session key decryption is very likely to result in an invalid cipherFunc:
				// only for < 5% of cases we will proceed to decrypt the data
				if err == nil {
					decrypted, err = edp.Decrypt(cipherFunc, key)
					if err != nil {
						return nil, err
					}
					if decrypted != nil {
						if md.DecryptedWithAlgorithm == 0 { // if no SEIPD v2 packet, key packet stores the decryption algorithm
							md.DecryptedWithAlgorithm = cipherFunc
						}
						if config.RetrieveSessionKey() {
							md.SessionKey = key
						}
						break FindKey
					}
				}
			}
		}
	}

	md.decrypted = decrypted
	if err := packets.Push(decrypted); err != nil {
		return nil, err
	}
	mdFinal, sensitiveParsingErr := readSignedMessage(packets, md, keyring, config)
	if sensitiveParsingErr != nil {
		return nil, errors.HandleSensitiveParsingError(sensitiveParsingErr, md.decrypted != nil)
	}
	return mdFinal, nil
}

// SignatureCandidate keeps state about a signature that can be potentially verified.
type SignatureCandidate struct {
	OPSVersion        int
	SigType           packet.SignatureType
	HashAlgorithm     crypto.Hash
	PubKeyAlgo        packet.PublicKeyAlgorithm
	IssuerKeyId       uint64
	IssuerFingerprint []byte
	Salt              []byte // v6 only

	SignedByEntity    *Entity
	SignedBy          *Key              // the key of the signer, if available. (OPS)
	SignatureError    error             // nil if the signature is valid or not checked.
	CorrespondingSig  *packet.Signature // the candidate's signature packet
	Hash, WrappedHash hash.Hash         // hashes for this signature candidate (OPS)
}

func newSignatureCandidate(ops *packet.OnePassSignature) (sigCandidate *SignatureCandidate) {
	sigCandidate = &SignatureCandidate{
		OPSVersion:        ops.Version,
		SigType:           ops.SigType,
		HashAlgorithm:     ops.Hash,
		PubKeyAlgo:        ops.PubKeyAlgo,
		IssuerKeyId:       ops.KeyId,
		IssuerFingerprint: ops.KeyFingerprint,
		Salt:              ops.Salt,
	}
	sigCandidate.Hash, sigCandidate.WrappedHash, sigCandidate.SignatureError = hashForSignature(
		sigCandidate.HashAlgorithm,
		sigCandidate.SigType,
		sigCandidate.Salt,
	)
	return
}

func newSignatureCandidateFromSignature(sig *packet.Signature) (sigCandidate *SignatureCandidate, err error) {
	if sig.IssuerKeyId == nil {
		return nil, errors.StructuralError("signature doesn't have an issuer")
	}

	sigCandidate = &SignatureCandidate{
		SigType:           sig.SigType,
		HashAlgorithm:     sig.Hash,
		PubKeyAlgo:        sig.PubKeyAlgo,
		IssuerKeyId:       *sig.IssuerKeyId,
		IssuerFingerprint: sig.IssuerFingerprint,
		Salt:              sig.Salt(),
	}
	sigCandidate.OPSVersion = 3
	if sig.Version == 6 {
		sigCandidate.OPSVersion = sig.Version
	}
	sigCandidate.Hash, sigCandidate.WrappedHash, sigCandidate.SignatureError = hashForSignature(
		sigCandidate.HashAlgorithm,
		sigCandidate.SigType,
		sigCandidate.Salt,
	)
	sigCandidate.CorrespondingSig = sig
	return
}

func (sc *SignatureCandidate) validate() bool {
	if sc.CorrespondingSig == nil {
		return false
	}
	correspondingSig := sc.CorrespondingSig
	invalidV3 := sc.OPSVersion == 3 && correspondingSig.Version == 6
	invalidV6 := sc.OPSVersion == 6 &&
		(correspondingSig.Version != 6 ||
			!bytes.Equal(sc.IssuerFingerprint, correspondingSig.IssuerFingerprint) ||
			!bytes.Equal(sc.Salt, correspondingSig.Salt()))
	return !invalidV3 && !invalidV6 &&
		sc.SigType == correspondingSig.SigType &&
		sc.HashAlgorithm == correspondingSig.Hash &&
		sc.PubKeyAlgo == correspondingSig.PubKeyAlgo &&
		correspondingSig.IssuerKeyId != nil &&
		sc.IssuerKeyId == *correspondingSig.IssuerKeyId
}

// readSignedMessage reads a possibly signed message if mdin is non-zero then
// that structure is updated and returned. Otherwise a fresh MessageDetails is
// used.
func readSignedMessage(packets packet.PacketReader, mdin *MessageDetails, keyring KeyRing, config *packet.Config) (md *MessageDetails, err error) {
	if mdin == nil {
		mdin = new(MessageDetails)
	}
	md = mdin

	var p packet.Packet
	var prevLast bool
FindLiteralData:
	for {
		p, err = packets.Next()
		if err != nil {
			return nil, err
		}
		switch p := p.(type) {
		case *packet.Compressed:
			if err := packets.Push(p.LimitedBodyReader(config.DecompressedMessageSizeLimit())); err != nil {
				return nil, err
			}
		case *packet.OnePassSignature:
			if prevLast {
				return nil, errors.UnsupportedError("nested signature packets")
			}

			if p.IsLast {
				prevLast = true
			}

			sigCandidate := newSignatureCandidate(p)
			md.IsSigned = true
			if keyring != nil {
				keys := keyring.EntitiesById(p.KeyId)
				if len(keys) > 0 {
					sigCandidate.SignedByEntity = keys[0]
				}
			}
			// If a message contains more than one one-pass signature, then the Signature packets bracket the message
			md.SignatureCandidates = append([]*SignatureCandidate{sigCandidate}, md.SignatureCandidates...)
		case *packet.Signature:
			// Old style signature i.e., sig | literal
			sigCandidate, err := newSignatureCandidateFromSignature(p)
			if err != nil {
				return nil, err
			}
			md.IsSigned = true
			if keyring != nil {
				keys := keyring.EntitiesById(sigCandidate.IssuerKeyId)
				if len(keys) > 0 {
					sigCandidate.SignedByEntity = keys[0]
				}
			}
			md.SignatureCandidates = append([]*SignatureCandidate{sigCandidate}, md.SignatureCandidates...)
		case *packet.LiteralData:
			md.LiteralData = p
			break FindLiteralData
		case *packet.EncryptedKey,
			*packet.SymmetricKeyEncrypted,
			*packet.AEADEncrypted,
			*packet.SymmetricallyEncrypted:
			return nil, errors.UnsupportedError("cannot read signed message with encrypted data")
		}
	}

	if md.IsSigned {
		md.UnverifiedBody = &signatureCheckReader{packets, md, config, md.LiteralData.Body}
	} else {
		md.UnverifiedBody = &checkReader{md, packets, false}
	}

	return md, nil
}

func wrapHashForSignature(hashFunc hash.Hash, sigType packet.SignatureType) (hash.Hash, error) {
	switch sigType {
	case packet.SigTypeBinary:
		return hashFunc, nil
	case packet.SigTypeText:
		return NewCanonicalTextHash(hashFunc), nil
	}
	return nil, errors.UnsupportedError("unsupported signature type: " + strconv.Itoa(int(sigType)))
}

// hashForSignature returns a pair of hashes that can be used to verify a
// signature. The signature may specify that the contents of the signed message
// should be preprocessed (i.e. to normalize line endings). Thus this function
// returns two hashes. The first, directHash, will feed directly into the signature algorithm.
// The second, wrappedHash, should be used to hash the message itself and performs any needed preprocessing.
func hashForSignature(hashFunc crypto.Hash, sigType packet.SignatureType, sigSalt []byte) (directHash hash.Hash, wrappedHash hash.Hash, err error) {
	if _, ok := algorithm.HashToHashIdWithSha1(hashFunc); !ok {
		return nil, nil, errors.UnsupportedError("unsupported hash function")
	}
	if !hashFunc.Available() {
		return nil, nil, errors.UnsupportedError("hash not available: " + strconv.Itoa(int(hashFunc)))
	}
	h := hashFunc.New()
	if sigSalt != nil {
		h.Write(sigSalt)
	}
	wrappedHash, err = wrapHashForSignature(h, sigType)
	if err != nil {
		return nil, nil, err
	}
	switch sigType {
	case packet.SigTypeBinary:
		return h, wrappedHash, nil
	case packet.SigTypeText:
		return h, wrappedHash, nil
	}
	return nil, nil, errors.UnsupportedError("unsupported signature type: " + strconv.Itoa(int(sigType)))
}

// checkReader wraps an io.Reader from a LiteralData packet. When it sees EOF
// it closes the ReadCloser from any SymmetricallyEncrypted packet to trigger
// MDC checks.
type checkReader struct {
	md      *MessageDetails
	packets packet.PacketReader
	checked bool
}

func (cr *checkReader) Read(buf []byte) (int, error) {
	n, sensitiveParsingError := cr.md.LiteralData.Body.Read(buf)
	if sensitiveParsingError == io.EOF {
		if cr.checked {
			return n, io.EOF
		}
		for {
			_, err := cr.packets.Next()
			if err == io.EOF {
				break
			}
			if err != nil {
				return n, err
			}
		}
		if cr.md.decrypted != nil {
			if sensitiveParsingError := cr.md.decrypted.Close(); sensitiveParsingError != nil {
				return n, errors.HandleSensitiveParsingError(sensitiveParsingError, true)
			}
		}
		cr.checked = true
		return n, io.EOF
	}
	if sensitiveParsingError != nil {
		return n, errors.HandleSensitiveParsingError(sensitiveParsingError, true)
	}
	return n, nil
}

// signatureCheckReader wraps an io.Reader from a LiteralData packet and hashes
// the data as it is read. When it sees an EOF from the underlying io.Reader
// it parses and checks a trailing Signature packet and triggers any MDC checks.
type signatureCheckReader struct {
	packets packet.PacketReader
	md      *MessageDetails
	config  *packet.Config
	data    io.Reader
}

func (scr *signatureCheckReader) Read(buf []byte) (int, error) {
	n, sensitiveParsingError := scr.data.Read(buf)

	for _, candidate := range scr.md.SignatureCandidates {
		if candidate.SignatureError == nil && candidate.SignedByEntity != nil {
			candidate.WrappedHash.Write(buf[:n])
		}
	}

	readsDecryptedData := scr.md.decrypted != nil
	if sensitiveParsingError == io.EOF {

		var signatures []*packet.Signature

		// Read all signature packets and discard others.
		for {
			p, sensitiveParsingErrorPacket := scr.packets.Next()
			if sensitiveParsingErrorPacket == io.EOF {
				break
			}
			if sensitiveParsingErrorPacket != nil {
				return n, errors.HandleSensitiveParsingError(sensitiveParsingErrorPacket, readsDecryptedData)
			}
			if sig, ok := p.(*packet.Signature); ok {
				if sig.Version == 5 && scr.md.LiteralData != nil && (sig.SigType == 0x00 || sig.SigType == 0x01) {
					sig.Metadata = scr.md.LiteralData
				}
				signatures = append(signatures, sig)
			}
		}

		// Verify the integrity of the decrypted data before verifying the signatures.
		if scr.md.decrypted != nil {
			sensitiveParsingErrorPacket := scr.md.decrypted.Close()
			if sensitiveParsingErrorPacket != nil {
				return n, errors.HandleSensitiveParsingError(sensitiveParsingErrorPacket, true)
			}
		}

		numberOfOpsSignatures := 0
		for _, candidate := range scr.md.SignatureCandidates {
			if candidate.CorrespondingSig == nil {
				numberOfOpsSignatures++
			}
		}

		if len(signatures) != numberOfOpsSignatures {
			// Cannot handle this case yet with no information about invalid packets, should fail.
			// This case can happen if a known OPS version is used but an unknown signature version.
			noMatchError := errors.StructuralError("number of ops signature candidates does not match the number of signature packets")
			for _, candidate := range scr.md.SignatureCandidates {
				candidate.SignatureError = noMatchError
			}
		} else {
			var sigIndex int
			// Verify all signature candidates.
			for _, candidate := range scr.md.SignatureCandidates {
				if candidate.CorrespondingSig == nil {
					candidate.CorrespondingSig = signatures[sigIndex]
					sigIndex++
				}
				if !candidate.validate() {
					candidate.SignatureError = errors.StructuralError("signature does not match the expected ops data")
				}
				if candidate.SignatureError == nil {
					sig := candidate.CorrespondingSig
					if candidate.SignedByEntity == nil {
						candidate.SignatureError = errors.ErrUnknownIssuer
						scr.md.SignatureError = candidate.SignatureError
					} else {
						// Verify and retrieve signing key at signature creation time
						key, err := candidate.SignedByEntity.signingKeyByIdUsage(sig.CreationTime, candidate.IssuerKeyId, packet.KeyFlagSign, scr.config)
						if err != nil {
							candidate.SignatureError = err
							continue
						} else {
							candidate.SignedBy = &key
						}
						signatureError := key.PublicKey.VerifySignature(candidate.Hash, sig)
						if signatureError == nil {
							signatureError = checkMessageSignatureDetails(&key, sig, scr.config)
						}
						if !scr.md.IsSymmetricallyEncrypted && len(sig.IntendedRecipients) > 0 && scr.md.CheckRecipients && signatureError == nil {
							if !scr.md.IsEncrypted {
								signatureError = errors.SignatureError("intended recipients in non-encrypted message")
							} else {
								// Check signature matches one of the recipients
								signatureError = checkIntendedRecipientsMatch(&scr.md.DecryptedWith, sig)
							}
						}
						candidate.SignatureError = signatureError
					}
				}
			}
		}

		// Check if there is a valid candidate.
		for _, candidate := range scr.md.SignatureCandidates {
			if candidate.SignedBy == nil {
				// Ignore candidates that have no matching key
				continue
			}
			// md.SignatureError points to the last candidate with a key match, if
			// all signature verifications have failed.
			scr.md.SignatureError = candidate.SignatureError
			scr.md.SelectedCandidate = candidate
			if candidate.SignatureError == nil {
				// There is a valid signature.
				scr.md.Signature = candidate.CorrespondingSig
				scr.md.SignedBy = candidate.SignedBy
				break
			}
		}

		if len(scr.md.SignatureCandidates) == 0 {
			scr.md.SignatureError = errors.StructuralError("no signature found")
		}

		if len(scr.md.SignatureCandidates) > 0 && scr.md.SelectedCandidate == nil {
			// No candidate with a matching key present.
			// Just point to the last candidate in this case.
			candidate := scr.md.SignatureCandidates[len(scr.md.SignatureCandidates)-1]
			scr.md.SignatureError = candidate.SignatureError
			scr.md.SelectedCandidate = candidate
		}

		if scr.md.SignatureError == nil && scr.md.Signature == nil {
			scr.md.SignatureError = errors.StructuralError("no matching signature found")
		}

		scr.md.IsVerified = true
		return n, io.EOF
	}

	if sensitiveParsingError != nil {
		return n, errors.HandleSensitiveParsingError(sensitiveParsingError, readsDecryptedData)
	}

	return n, nil
}

// VerifyDetachedSignature takes a signed file and a detached signature and
// returns the signature packet and the entity the signature was signed by,
// if any, and a possible signature verification error.
// If the signer isn't known, ErrUnknownIssuer is returned.
func VerifyDetachedSignature(keyring KeyRing, signed, signature io.Reader, config *packet.Config) (sig *packet.Signature, signer *Entity, err error) {
	return verifyDetachedSignature(keyring, signed, signature, config)
}

// VerifyDetachedSignatureReader takes a signed file and a detached signature and
// returns message details struct similar to the ReadMessage function.
// Once all data is read from md.UnverifiedBody the detached signature is verified.
// If a verification error occurs it is stored in md.SignatureError
// If the signer isn't known, ErrUnknownIssuer is returned.
func VerifyDetachedSignatureReader(keyring KeyRing, signed, signature io.Reader, config *packet.Config) (md *MessageDetails, err error) {
	return verifyDetachedSignatureReader(keyring, signed, signature, config)
}

// VerifyArmoredDetachedSignature performs the same actions as
// VerifyDetachedSignature but expects the signature to be armored.
func VerifyArmoredDetachedSignature(keyring KeyRing, signed, signature io.Reader, config *packet.Config) (sig *packet.Signature, signer *Entity, err error) {
	body, err := readArmored(signature, SignatureType)
	if err != nil {
		return
	}

	return VerifyDetachedSignature(keyring, signed, body, config)
}

func verifyDetachedSignature(keyring KeyRing, signed, signature io.Reader, config *packet.Config) (sig *packet.Signature, signer *Entity, err error) {
	md, err := verifyDetachedSignatureReader(keyring, signed, signature, config)
	if err != nil {
		return nil, nil, err
	}
	_, err = io.Copy(ioutil.Discard, md.UnverifiedBody)
	if err != nil {
		return nil, nil, err
	}
	if md.SignatureError != nil {
		return nil, nil, md.SignatureError
	}
	return md.Signature, md.SignedBy.Entity, nil
}

func verifyDetachedSignatureReader(keyring KeyRing, signed, signature io.Reader, config *packet.Config) (md *MessageDetails, err error) {
	var p packet.Packet
	md = &MessageDetails{
		IsEncrypted:     false,
		CheckRecipients: false,
		IsSigned:        true,
	}

	packets := packet.NewReader(signature)
	for {
		p, err = packets.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		sig, ok := p.(*packet.Signature)
		if !ok {
			continue
		}
		candidate, err := newSignatureCandidateFromSignature(sig)
		if err != nil {
			return nil, err
		}
		md.SignatureCandidates = append(md.SignatureCandidates, candidate)

		keys := keyring.EntitiesById(candidate.IssuerKeyId)
		if len(keys) > 0 {
			candidate.SignedByEntity = keys[0]
		}
	}

	if len(md.SignatureCandidates) == 0 {
		return nil, errors.ErrUnknownIssuer
	}
	md.UnverifiedBody = &signatureCheckReader{packets, md, config, signed}
	return md, nil
}

// checkSignatureDetails verifies the metadata of the signature.
// It checks the following:
//   - Hash function should not be invalid according to
//     config.RejectHashAlgorithms.
//   - Check signature notations.
//   - Signature is not expired (unless a zero time is passed to
//     explicitly ignore expiration).
func checkSignatureDetails(signature *packet.Signature, now time.Time, config *packet.Config) error {
	if config.RejectHashAlgorithm(signature.Hash) {
		return errors.SignatureError("insecure hash algorithm: " + signature.Hash.String())
	}

	for _, notation := range signature.Notations {
		if notation.IsCritical && !config.KnownNotation(notation.Name) {
			return errors.SignatureError("unknown critical notation: " + notation.Name)
		}
	}

	if !now.IsZero() && signature.SigExpired(now) {
		return errors.ErrSignatureExpired
	}
	return nil
}

// checkMessageSignatureDetails verifies the metadata of the signature.
// It checks the criteria of checkSignatureDetails for the message
// signature and all relevant binding signatures.
// In addition, the message signature hash algorithm is checked against
// config.RejectMessageHashAlgorithms.
// Finally, the signature must be newer than the verification key.
func checkMessageSignatureDetails(verifiedKey *Key, signature *packet.Signature, config *packet.Config) error {
	if config.RejectMessageHashAlgorithm(signature.Hash) {
		return errors.SignatureError("insecure message hash algorithm: " + signature.Hash.String())
	}

	if signature.CreationTime.Unix() < verifiedKey.PublicKey.CreationTime.Unix() {
		return errors.ErrSignatureOlderThanKey
	}

	sigsToCheck := []*packet.Signature{signature, verifiedKey.PrimarySelfSignature}
	if !verifiedKey.IsPrimary() {
		sigsToCheck = append(sigsToCheck, verifiedKey.SelfSignature, verifiedKey.SelfSignature.EmbeddedSignature)
	}
	var errs []error
	for _, sig := range sigsToCheck {
		var time time.Time
		if sig == signature {
			time = config.Now()
		} else {
			time = signature.CreationTime
		}
		if err := checkSignatureDetails(sig, time, config); err != nil {
			errs = append(errs, err)
		}
	}

	// Return errors.ErrSignatureExpired last as gopenpgp might ignore it.
	for _, err := range errs {
		if err != errors.ErrSignatureExpired {
			return err
		}
	}
	if len(errs) != 0 {
		return errs[0]
	}
	return nil
}

// checkIntendedRecipientsMatch checks if the fingerprint of the primary key matching the decryption key
// is found in the signature's intended recipients list.
func checkIntendedRecipientsMatch(decryptionKey *Key, sig *packet.Signature) error {
	match := false
	for _, recipient := range sig.IntendedRecipients {
		if bytes.Equal(recipient.Fingerprint, decryptionKey.Entity.PrimaryKey.Fingerprint) {
			match = true
			break
		}
	}
	if !match {
		return errors.SignatureError("intended recipients in the signature does not match the decryption key")
	}
	return nil
}
