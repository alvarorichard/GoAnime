// Copyright 2011 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package v2

import (
	"bytes"
	"crypto"
	goerrors "errors"
	"hash"
	"io"
	"strconv"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp/armor"
	"github.com/ProtonMail/go-crypto/openpgp/errors"
	"github.com/ProtonMail/go-crypto/openpgp/internal/algorithm"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

// DetachSign signs message with the private key from signer (which must
// already have been decrypted) and writes the signature to w.
// If config is nil, sensible defaults will be used.
func DetachSign(w io.Writer, signers []*Entity, message io.Reader, config *packet.Config) error {
	return detachSign(w, signers, message, packet.SigTypeBinary, config)
}

// DetachSignWithParams signs message with the private key from signer (which must
// already have been decrypted) and writes the signature to the Writer.
// If config is nil, sensible defaults will be used.
func DetachSignWithParams(w io.Writer, signers []*Entity, message io.Reader, params *SignParams) error {
	if params == nil {
		params = &SignParams{}
	}
	sigType := packet.SigTypeBinary
	if params.TextSig {
		sigType = packet.SigTypeText
	}
	return detachSign(w, signers, message, sigType, params.Config)
}

// ArmoredDetachSign signs message with the private key from signer (which
// must already have been decrypted) and writes an armored signature to the Writer.
// If config is nil, sensible defaults will be used.
func ArmoredDetachSign(w io.Writer, signers []*Entity, message io.Reader, params *SignParams) (err error) {
	if params == nil {
		params = &SignParams{}
	}
	sigType := packet.SigTypeBinary
	if params.TextSig {
		sigType = packet.SigTypeText
	}
	return armoredDetachSign(w, signers, message, sigType, params.Config)
}

// DetachSignWriter signs a message with the private key from a signer (which must
// already have been decrypted) and writes the signature to the Writer.
// DetachSignWriter returns a WriteCloser to which the message can be written to.
// The resulting WriteCloser must be closed after the contents of the message have
// been written. If utf8Message is set to true, the line endings of the message are
// canonicalised and the type of the signature will be SigTypeText.
// If config is nil, sensible defaults will be used.
func DetachSignWriter(w io.Writer, signers []*Entity, params *SignParams) (io.WriteCloser, error) {
	if params == nil {
		params = &SignParams{}
	}
	sigType := packet.SigTypeBinary
	if params.TextSig {
		sigType = packet.SigTypeText
	}
	return detachSignWithWriter(w, signers, sigType, params.Config)
}

func armoredDetachSign(w io.Writer, signers []*Entity, message io.Reader, sigType packet.SignatureType, config *packet.Config) (err error) {
	out, err := armor.EncodeWithChecksumOption(w, SignatureType, nil, false)
	if err != nil {
		return
	}
	err = detachSign(out, signers, message, sigType, config)
	if err != nil {
		return
	}
	return out.Close()
}

func detachSign(w io.Writer, signers []*Entity, message io.Reader, sigType packet.SignatureType, config *packet.Config) (err error) {
	ptWriter, err := detachSignWithWriter(w, signers, sigType, config)
	if err != nil {
		return
	}
	_, err = io.Copy(ptWriter, message)
	if err != nil {
		return
	}
	return ptWriter.Close()
}

type detachSignWriter struct {
	signatureWriter io.Writer
	signatures      []*detachSignContext
	config          *packet.Config
}

type detachSignContext struct {
	wrappedHash hash.Hash
	h           hash.Hash
	signer      *packet.PrivateKey
	sig         *packet.Signature
}

func (s detachSignWriter) Write(data []byte) (int, error) {
	for _, signature := range s.signatures {
		if n, err := signature.wrappedHash.Write(data); err != nil {
			return n, err
		}
	}
	return len(data), nil
}

func (s detachSignWriter) Close() error {
	for _, signature := range s.signatures {
		err := signature.sig.Sign(signature.h, signature.signer, s.config)
		if err != nil {
			return err
		}
		err = signature.sig.Serialize(s.signatureWriter)
		if err != nil {
			return err
		}
	}
	return nil
}

func detachSignWithWriter(w io.Writer, signers []*Entity, sigType packet.SignatureType, config *packet.Config) (ptWriter io.WriteCloser, err error) {
	var detachSignContexts []*detachSignContext
	for _, signer := range signers {
		signingKey, ok := signer.SigningKeyById(config.Now(), config.SigningKey(), config)
		if !ok {
			return nil, errors.InvalidArgumentError("no valid signing keys")
		}
		if signingKey.PrivateKey == nil {
			return nil, errors.InvalidArgumentError("signing key doesn't have a private key")
		}
		if signingKey.PrivateKey.Encrypted {
			return nil, errors.InvalidArgumentError("signing key is encrypted")
		}
		candidateHashes := []uint8{
			hashToHashId(crypto.SHA256),
			hashToHashId(crypto.SHA384),
			hashToHashId(crypto.SHA512),
			hashToHashId(crypto.SHA3_256),
			hashToHashId(crypto.SHA3_512),
		}
		defaultHashes := candidateHashes[0:1]
		primarySelfSignature, _ := signer.PrimarySelfSignature(config.Now(), config)
		if primarySelfSignature == nil {
			return nil, errors.StructuralError("signed entity has no valid self-signature")
		}
		preferredHashes := primarySelfSignature.PreferredHash
		if len(preferredHashes) == 0 {
			preferredHashes = defaultHashes
		}
		candidateHashes = intersectPreferences(candidateHashes, preferredHashes)

		var hash crypto.Hash
		if hash, err = selectHash(candidateHashes, config.Hash(), signingKey.PrivateKey); err != nil {
			return
		}

		detachSignCtx := detachSignContext{
			signer: signingKey.PrivateKey,
		}

		detachSignCtx.sig = createSignaturePacket(signingKey.PublicKey, sigType, config)
		detachSignCtx.sig.Hash = hash

		detachSignCtx.h, err = detachSignCtx.sig.PrepareSign(config)
		if err != nil {
			return
		}
		detachSignCtx.wrappedHash, err = wrapHashForSignature(detachSignCtx.h, sigType)
		if err != nil {
			return
		}
		detachSignContexts = append(detachSignContexts, &detachSignCtx)
	}

	return &detachSignWriter{
		signatureWriter: w,
		signatures:      detachSignContexts,
		config:          config,
	}, nil
}

// FileHints contains metadata about encrypted files. This metadata is, itself,
// encrypted. OpenPGP signatures do not include the FileHints in a signature hash and
// thus those fields are not protected against tampering in a signed document.
// The crypto[refresh does not recommend to set the data in file hints.
// See https://www.ietf.org/archive/id/draft-ietf-openpgp-crypto-refresh-12.html#section-5.9.
type FileHints struct {
	// IsUTF8 can be set to hint that the contents are utf8 encoded data.
	IsUTF8 bool
	// FileName hints at the name of the file that should be written.
	FileName string
	// ModTime contains the modification time of the file, or the zero time if not applicable.
	ModTime time.Time
}

type EncryptParams struct {
	// KeyWriter is a Writer to which the encrypted
	// session keys are written to.
	// If nil, DataWriter is used instead.
	KeyWriter io.Writer
	// Hints contains file metadata for the literal data packet.
	// If nil, default is used.
	Hints *FileHints
	// SiningEntities contains the private keys to produce signatures with
	// If nil, no signatures are created.
	Signers []*Entity
	// TextSig indicates if signatures of type SigTypeText should be produced.
	TextSig bool
	// Passwords defines additional passwords that the message should be encrypted to.
	// i.e., for each defined password an additional SKESK packet is written.
	Passwords [][]byte
	// SessionKey provides a session key to be used for encryption.
	// If nil, a random one-time session key is generated.
	SessionKey []byte
	// OutsideSig allows to set a signature that should be included
	// in the message to encrypt.
	// Should only be used for exceptional cases.
	// If nil, ignored.
	OutsideSig []byte
	// EncryptionTime allows to override the time that is used
	// for selecting the encryption key.
	// If EncryptionTime is zero (i.e., EncryptionTime.isZero()) expiration checks
	// are not performed on the encryption key.
	// If nil, the default clock in config is used.
	EncryptionTime *time.Time
	// Config provides the config to be used.
	// If Config is nil, sensible defaults will be used.
	Config *packet.Config
}

// SymmetricallyEncrypt acts like gpg -c: it encrypts a file with a passphrase.
// The resulting WriteCloser must be closed after the contents of the file have
// been written.
// If config is nil, sensible defaults will be used.
func SymmetricallyEncrypt(ciphertext io.Writer, passphrase []byte, hints *FileHints, config *packet.Config) (plaintext io.WriteCloser, err error) {
	return SymmetricallyEncryptWithParams(passphrase, ciphertext, &EncryptParams{
		Hints:  hints,
		Config: config,
	})
}

// SymmetricallyEncryptWithParams acts like SymmetricallyEncrypt but provides more configuration options.
// EncryptParams provides the optional parameters.
// The resulting WriteCloser must be closed after the contents of the file have been written.
func SymmetricallyEncryptWithParams(passphrase []byte, dataWriter io.Writer, params *EncryptParams) (plaintext io.WriteCloser, err error) {
	if params == nil {
		params = &EncryptParams{}
	}
	return symmetricallyEncrypt(passphrase, dataWriter, params)
}

func symmetricallyEncrypt(passphrase []byte, dataWriter io.Writer, params *EncryptParams) (plaintext io.WriteCloser, err error) {
	if params.KeyWriter == nil {
		params.KeyWriter = dataWriter
	}
	if params.Hints == nil {
		params.Hints = &FileHints{}
	}
	if params.SessionKey == nil {
		params.SessionKey, err = packet.SerializeSymmetricKeyEncrypted(params.KeyWriter, passphrase, params.Config)
		defer func() {
			// zero the session key after we are done
			for i := range params.SessionKey {
				params.SessionKey[i] = 0
			}
			params.SessionKey = nil
		}()
	} else {
		err = packet.SerializeSymmetricKeyEncryptedReuseKey(params.KeyWriter, params.SessionKey, passphrase, params.Config)
	}
	if err != nil {
		return
	}
	for _, additionalPassword := range params.Passwords {
		if err = packet.SerializeSymmetricKeyEncryptedReuseKey(params.KeyWriter, params.SessionKey, additionalPassword, params.Config); err != nil {
			return
		}
	}

	config := params.Config
	candidateCompression := []uint8{uint8(config.Compression())}
	cipherSuite := packet.CipherSuite{
		Cipher: config.Cipher(),
		Mode:   config.AEAD().Mode(),
	}
	var candidateHashesPerSignature [][]uint8
	if params.Signers != nil {
		for _, signer := range params.Signers {
			// candidateHashes := []uint8{hashToHashId(config.Hash())}
			// Check what the preferred hashes are for the signing key
			candidateHashes := []uint8{
				hashToHashId(crypto.SHA256),
				hashToHashId(crypto.SHA384),
				hashToHashId(crypto.SHA512),
				hashToHashId(crypto.SHA3_256),
				hashToHashId(crypto.SHA3_512),
			}
			defaultHashes := candidateHashes[0:1]
			primarySelfSignature, _ := signer.PrimarySelfSignature(params.Config.Now(), params.Config)
			if primarySelfSignature == nil {
				return nil, errors.StructuralError("signed entity has no self-signature")
			}
			preferredHashes := primarySelfSignature.PreferredHash
			if len(preferredHashes) == 0 {
				preferredHashes = defaultHashes
			}
			candidateHashes = intersectPreferences(candidateHashes, preferredHashes)
			if len(candidateHashes) == 0 {
				candidateHashes = []uint8{hashToHashId(crypto.SHA256)}
			}
			candidateHashesPerSignature = append(candidateHashesPerSignature, candidateHashes)
		}
	}
	return encryptDataAndSign(dataWriter, params, candidateHashesPerSignature, candidateCompression, config.Cipher(), config.AEAD() != nil, cipherSuite, nil)
}

// intersectPreferences mutates and returns a prefix of a that contains only
// the values in the intersection of a and b. The order of a is preserved.
func intersectPreferences(a []uint8, b []uint8) (intersection []uint8) {
	var j int
	for _, v := range a {
		for _, v2 := range b {
			if v == v2 {
				a[j] = v
				j++
				break
			}
		}
	}

	return a[:j]
}

// intersectCipherSuites mutates and returns a prefix of a that contains only
// the values in the intersection of a and b. The order of a is preserved.
func intersectCipherSuites(a [][2]uint8, b [][2]uint8) (intersection [][2]uint8) {
	var j int
	for _, v := range a {
		for _, v2 := range b {
			if v[0] == v2[0] && v[1] == v2[1] {
				a[j] = v
				j++
				break
			}
		}
	}

	return a[:j]
}

func hashToHashId(h crypto.Hash) uint8 {
	v, ok := algorithm.HashToHashId(h)
	if !ok {
		panic("tried to convert unknown hash")
	}
	return v
}

// EncryptWithParams encrypts a message to a number of recipients and, optionally,
// signs it. The resulting WriteCloser must be closed after the contents of the file have been written.
// The to argument contains recipients that are explicitly mentioned in signatures and encrypted keys,
// whereas the toHidden argument contains recipients that will be hidden and not mentioned.
// Params contains all optional parameters.
func EncryptWithParams(ciphertext io.Writer, to, toHidden []*Entity, params *EncryptParams) (plaintext io.WriteCloser, err error) {
	if params == nil {
		params = &EncryptParams{}
	}
	if params.KeyWriter == nil {
		params.KeyWriter = ciphertext
	}
	return encrypt(to, toHidden, ciphertext, params)
}

// Encrypt encrypts a message to a number of recipients and, optionally, signs
// it. Hints contains optional information, that is also encrypted, that aids
// the recipients in processing the message. The crypto-refresh recommends
// to not set file hints since the data is not included in the signature hash.
// See https://www.ietf.org/archive/id/draft-ietf-openpgp-crypto-refresh-12.html#section-5.9.
// The resulting WriteCloser must be closed after the contents of the file have been written.
// The to argument contains recipients that are explicitly mentioned in signatures and encrypted keys,
// whereas the toHidden argument contains recipients that will be hidden and not mentioned.
// If config is nil, sensible defaults will be used.
func Encrypt(ciphertext io.Writer, to, toHidden []*Entity, signers []*Entity, hints *FileHints, config *packet.Config) (plaintext io.WriteCloser, err error) {
	return EncryptWithParams(ciphertext, to, toHidden, &EncryptParams{
		Signers: signers,
		Hints:   hints,
		Config:  config,
	})
}

// writeAndSign writes the data as a payload package and, optionally, signs
// it. Hints contains optional information, that is also encrypted,
// that aids the recipients in processing the message. The resulting
// WriteCloser must be closed after the contents of the file have been
// written. If config is nil, sensible defaults will be used.
func writeAndSign(payload io.WriteCloser, candidateHashes [][]uint8, signEntities []*Entity, hints *FileHints, sigType packet.SignatureType, intendedRecipients []*packet.Recipient, outsideSig []byte, config *packet.Config) (plaintext io.WriteCloser, err error) {
	var signers []*signatureContext
	var numberOfOutsideSigs int

	if outsideSig != nil {
		outSigPacket, err := parseOutsideSig(outsideSig)
		if err != nil {
			return nil, err
		}
		opsVersion := 3
		if outSigPacket.Version == 6 {
			opsVersion = 6
		}
		opsOutside := &packet.OnePassSignature{
			Version:    opsVersion,
			SigType:    outSigPacket.SigType,
			Hash:       outSigPacket.Hash,
			PubKeyAlgo: outSigPacket.PubKeyAlgo,
			KeyId:      *outSigPacket.IssuerKeyId,
			IsLast:     len(signEntities) == 0,
		}
		sigContext := signatureContext{
			outsideSig: outSigPacket,
		}
		if outSigPacket.Version == 6 {
			opsOutside.KeyFingerprint = outSigPacket.IssuerFingerprint
			sigContext.salt = outSigPacket.Salt()
			opsOutside.Salt = outSigPacket.Salt()
		}
		sigContext.h, sigContext.wrappedHash, err = hashForSignature(outSigPacket.Hash, sigType, sigContext.salt)
		if err != nil {
			return nil, err
		}
		if err := opsOutside.Serialize(payload); err != nil {
			return nil, err
		}
		signers = append([]*signatureContext{&sigContext}, signers...)
		numberOfOutsideSigs = 1
	}

	for signEntityIdx, signEntity := range signEntities {
		if signEntity == nil {
			continue
		}
		signKey, ok := signEntity.SigningKeyById(config.Now(), config.SigningKey(), config)
		if !ok {
			return nil, errors.InvalidArgumentError("no valid signing keys")
		}
		signer := signKey.PrivateKey
		if signer == nil {
			return nil, errors.InvalidArgumentError("no private key in signing key")
		}
		if signer.Encrypted {
			return nil, errors.InvalidArgumentError("signing key must be decrypted")
		}
		sigContext := signatureContext{
			signer: signer,
		}

		if signKey.PrimarySelfSignature == nil {
			return nil, errors.InvalidArgumentError("signing key has no self-signature")
		}
		candidateHashes[signEntityIdx] = intersectPreferences(candidateHashes[signEntityIdx], signKey.PrimarySelfSignature.PreferredHash)
		hash, err := selectHash(candidateHashes[signEntityIdx], config.Hash(), signKey.PrivateKey)
		if err != nil {
			return nil, err
		}
		sigContext.hashType = hash

		var opsVersion = 3
		if signer.Version == 6 {
			opsVersion = signer.Version
		}
		isLast := signEntityIdx == len(signEntities)-1
		ops := &packet.OnePassSignature{
			Version:    opsVersion,
			SigType:    sigType,
			Hash:       hash,
			PubKeyAlgo: signer.PubKeyAlgo,
			KeyId:      signer.KeyId,
			IsLast:     isLast,
		}
		if opsVersion == 6 {
			ops.KeyFingerprint = signer.Fingerprint
			sigContext.salt, err = packet.SignatureSaltForHash(hash, config.Random())
			if err != nil {
				return nil, err
			}
			ops.Salt = sigContext.salt
		}
		if err := ops.Serialize(payload); err != nil {
			return nil, err
		}

		sigContext.h, sigContext.wrappedHash, err = hashForSignature(hash, sigType, sigContext.salt)
		if err != nil {
			return nil, err
		}
		// Prepend since the last signature has to be written first
		signers = append([]*signatureContext{&sigContext}, signers...)
	}

	if signEntities != nil && len(signEntities)+numberOfOutsideSigs != len(signers) {
		return nil, errors.InvalidArgumentError("no valid signing key")
	}

	if hints == nil {
		hints = &FileHints{}
	}

	w := payload
	if signers != nil || numberOfOutsideSigs > 0 {
		// If we need to write a signature packet after the literal
		// data then we need to stop literalData from closing
		// encryptedData.
		w = noOpCloser{w}

	}
	var epochSeconds uint32
	if !hints.ModTime.IsZero() {
		epochSeconds = uint32(hints.ModTime.Unix())
	}
	literalData, err := packet.SerializeLiteral(w, !hints.IsUTF8, hints.FileName, epochSeconds)
	if err != nil {
		return nil, err
	}

	if signers != nil || numberOfOutsideSigs > 0 {
		metadata := &packet.LiteralData{
			Format:   'b',
			FileName: hints.FileName,
			Time:     epochSeconds,
		}
		if hints.IsUTF8 {
			metadata.Format = 'u'
		}
		return signatureWriter{payload, literalData, signers, sigType, config, metadata, intendedRecipients, 0}, nil
	}
	return literalData, nil
}

// encrypt encrypts a message to a number of recipients and, optionally, signs
// it. The resulting WriteCloser must
// be closed after the contents of the file have been written.
func encrypt(
	to, toHidden []*Entity,
	dataWriter io.Writer,
	params *EncryptParams,
) (plaintext io.WriteCloser, err error) {
	if len(to)+len(toHidden) == 0 {
		return nil, errors.InvalidArgumentError("no encryption recipient provided")
	}

	// These are the possible ciphers that we'll use for the message.
	candidateCiphers := []uint8{
		uint8(packet.CipherAES256),
		uint8(packet.CipherAES128),
	}

	// These are the possible hash functions that we'll use for the signature.
	candidateHashes := []uint8{
		hashToHashId(crypto.SHA256),
		hashToHashId(crypto.SHA384),
		hashToHashId(crypto.SHA512),
		hashToHashId(crypto.SHA3_256),
		hashToHashId(crypto.SHA3_512),
	}

	// Prefer GCM if everyone supports it
	candidateCipherSuites := [][2]uint8{
		{uint8(packet.CipherAES256), uint8(packet.AEADModeGCM)},
		{uint8(packet.CipherAES256), uint8(packet.AEADModeEAX)},
		{uint8(packet.CipherAES256), uint8(packet.AEADModeOCB)},
		{uint8(packet.CipherAES128), uint8(packet.AEADModeGCM)},
		{uint8(packet.CipherAES128), uint8(packet.AEADModeEAX)},
		{uint8(packet.CipherAES128), uint8(packet.AEADModeOCB)},
	}

	candidateCompression := []uint8{
		uint8(packet.CompressionNone),
		uint8(packet.CompressionZIP),
		uint8(packet.CompressionZLIB),
	}

	encryptKeys := make([]Key, len(to)+len(toHidden))

	config := params.Config
	// AEAD is used if every key supports it
	aeadSupported := true

	var intendedRecipients []*packet.Recipient
	// Intended Recipient Fingerprint subpacket SHOULD be used when creating a signed and encrypted message
	for _, publicRecipient := range to {
		if config.IntendedRecipients() {
			intendedRecipients = append(intendedRecipients, &packet.Recipient{KeyVersion: publicRecipient.PrimaryKey.Version, Fingerprint: publicRecipient.PrimaryKey.Fingerprint})
		}
	}

	timeForEncryptionKey := config.Now()
	if params.EncryptionTime != nil {
		// Override the time to select the encryption key with the provided one.
		timeForEncryptionKey = *params.EncryptionTime
	}

	allPQ := len(encryptKeys) > 0
	for i, recipient := range append(to, toHidden...) {
		if encryptKeys[i], err = recipient.EncryptionKeyWithError(timeForEncryptionKey, config); err != nil {
			return nil, err
		}

		if !encryptKeys[i].PublicKey.IsPQ() {
			allPQ = false
		}

		primarySelfSignature, _ := recipient.PrimarySelfSignature(timeForEncryptionKey, config)
		if primarySelfSignature == nil {
			return nil, errors.StructuralError("entity without a self-signature")
		}

		if !primarySelfSignature.SEIPDv2 {
			aeadSupported = false
		}

		candidateCiphers = intersectPreferences(candidateCiphers, primarySelfSignature.PreferredSymmetric)
		candidateHashes = intersectPreferences(candidateHashes, primarySelfSignature.PreferredHash)
		candidateCipherSuites = intersectCipherSuites(candidateCipherSuites, primarySelfSignature.PreferredCipherSuites)
		candidateCompression = intersectPreferences(candidateCompression, primarySelfSignature.PreferredCompression)
	}

	// In the event that the intersection of supported algorithms is empty we use the ones
	// labelled as MUST that every implementation supports.
	if len(candidateCiphers) == 0 {
		// https://www.ietf.org/archive/id/draft-ietf-openpgp-crypto-refresh-07.html#section-9.3
		candidateCiphers = []uint8{uint8(packet.CipherAES128)}
	}
	if len(candidateHashes) == 0 {
		// https://www.ietf.org/archive/id/draft-ietf-openpgp-crypto-refresh-07.html#hash-algos
		candidateHashes = []uint8{hashToHashId(crypto.SHA256)}
	}
	if len(candidateCipherSuites) == 0 {
		if allPQ {
			candidateCipherSuites = [][2]uint8{{uint8(packet.CipherAES256), uint8(packet.AEADModeOCB)}}
		} else {
			// https://www.ietf.org/archive/id/draft-ietf-openpgp-crypto-refresh-07.html#section-9.6
			candidateCipherSuites = [][2]uint8{{uint8(packet.CipherAES128), uint8(packet.AEADModeOCB)}}
		}
	}

	cipher := packet.CipherFunction(candidateCiphers[0])
	aeadCipherSuite := packet.CipherSuite{
		Cipher: packet.CipherFunction(candidateCipherSuites[0][0]),
		Mode:   packet.AEADMode(candidateCipherSuites[0][1]),
	}

	// If the cipher specified by config is a candidate, we'll use that.
	configuredCipher := config.Cipher()
	for _, c := range candidateCiphers {
		cipherFunc := packet.CipherFunction(c)
		if cipherFunc == configuredCipher {
			cipher = cipherFunc
			break
		}
	}

	if params.SessionKey == nil {
		if aeadSupported {
			params.SessionKey = make([]byte, aeadCipherSuite.Cipher.KeySize())
		} else {
			params.SessionKey = make([]byte, cipher.KeySize())
		}

		if _, err := io.ReadFull(config.Random(), params.SessionKey); err != nil {
			return nil, err
		}
		defer func() {
			// zero the session key after we are done
			for i := range params.SessionKey {
				params.SessionKey[i] = 0
			}
			params.SessionKey = nil
		}()
	}

	for idx, key := range encryptKeys {
		// hide the keys of the hidden recipients
		hidden := idx >= len(to)
		if err := packet.SerializeEncryptedKeyAEADwithHiddenOption(params.KeyWriter, key.PublicKey, cipher, aeadSupported, params.SessionKey, hidden, config); err != nil {
			return nil, err
		}
	}

	for _, password := range params.Passwords {
		if err = packet.SerializeSymmetricKeyEncryptedAEADReuseKey(params.KeyWriter, params.SessionKey, password, aeadSupported, params.Config); err != nil {
			return nil, err
		}
	}

	var candidateHashesPerSignature [][]uint8
	for range params.Signers {
		candidateHashesPerSignature = append(candidateHashesPerSignature, candidateHashes)
	}
	return encryptDataAndSign(dataWriter, params, candidateHashesPerSignature, candidateCompression, cipher, aeadSupported, aeadCipherSuite, intendedRecipients)
}

func encryptDataAndSign(
	dataWriter io.Writer,
	params *EncryptParams,
	candidateHashes [][]uint8,
	candidateCompression []uint8,
	cipher packet.CipherFunction,
	aeadSupported bool,
	aeadCipherSuite packet.CipherSuite,
	intendedRecipients []*packet.Recipient,
) (plaintext io.WriteCloser, err error) {
	sigType := packet.SigTypeBinary
	if params.TextSig {
		sigType = packet.SigTypeText
	}
	payload, err := packet.SerializeSymmetricallyEncrypted(dataWriter, cipher, aeadSupported, aeadCipherSuite, params.SessionKey, params.Config)
	if err != nil {
		return
	}
	payload, err = handleCompression(payload, candidateCompression, params.Config)
	if err != nil {
		return nil, err
	}
	return writeAndSign(payload, candidateHashes, params.Signers, params.Hints, sigType, intendedRecipients, params.OutsideSig, params.Config)
}

type SignParams struct {
	// Hints contains file metadata for the literal data packet.
	// The crypto-refresh recommends to not set file hints since the data is not included in the signature hash.
	// See https://www.ietf.org/archive/id/draft-ietf-openpgp-crypto-refresh-12.html#section-5.9.
	// If nil, default is used.
	Hints *FileHints
	// TextSig indicates if signatures of type SigTypeText should be produced
	TextSig bool
	// OutsideSig allows to set a signature that should be included
	// in an inline signed message.
	// Should only be used for exceptional cases.
	// If nil, ignored.
	OutsideSig []byte
	// Config provides the config to be used.
	// If Config is nil, sensible defaults will be used.
	Config *packet.Config
}

// SignWithParams signs a message. The resulting WriteCloser must be closed after the
// contents of the file have been written.
// SignParams can contain optional params and can be nil for defaults.
func SignWithParams(output io.Writer, signers []*Entity, params *SignParams) (input io.WriteCloser, err error) {
	if params == nil {
		params = &SignParams{}
	}
	if len(signers) < 1 && params.OutsideSig == nil {
		return nil, errors.InvalidArgumentError("no signer provided")
	}
	var candidateHashesPerSignature [][]uint8
	candidateCompression := []uint8{
		uint8(packet.CompressionNone),
		uint8(packet.CompressionZIP),
		uint8(packet.CompressionZLIB),
	}
	for _, signer := range signers {
		// These are the possible hash functions that we'll use for the signature.
		candidateHashes := []uint8{
			hashToHashId(crypto.SHA256),
			hashToHashId(crypto.SHA384),
			hashToHashId(crypto.SHA512),
			hashToHashId(crypto.SHA3_256),
			hashToHashId(crypto.SHA3_512),
		}
		defaultHashes := candidateHashes[0:1]
		primarySelfSignature, _ := signer.PrimarySelfSignature(params.Config.Now(), params.Config)
		if primarySelfSignature == nil {
			return nil, errors.StructuralError("signed entity has no self-signature")
		}
		preferredHashes := primarySelfSignature.PreferredHash
		if len(preferredHashes) == 0 {
			preferredHashes = defaultHashes
		}
		candidateHashes = intersectPreferences(candidateHashes, preferredHashes)
		if len(candidateHashes) == 0 {
			return nil, errors.StructuralError("cannot sign because signing key shares no common algorithms with candidate hashes")
		}
		candidateHashesPerSignature = append(candidateHashesPerSignature, candidateHashes)
		candidateCompression = intersectPreferences(candidateCompression, primarySelfSignature.PreferredCompression)

	}

	sigType := packet.SigTypeBinary
	if params.TextSig {
		sigType = packet.SigTypeText
	}

	var payload io.WriteCloser
	payload = noOpCloser{output}
	payload, err = handleCompression(payload, candidateCompression, params.Config)
	if err != nil {
		return nil, err
	}
	return writeAndSign(payload, candidateHashesPerSignature, signers, params.Hints, sigType, nil, params.OutsideSig, params.Config)
}

// Sign signs a message. The resulting WriteCloser must be closed after the
// contents of the file have been written. Hints contains optional information
// that aids the recipients in processing the message.
// The crypto-refresh recommends to not set file hints since the data is not included in the signature hash.
// See https://www.ietf.org/archive/id/draft-ietf-openpgp-crypto-refresh-12.html#section-5.9.
// If config is nil, sensible defaults will be used.
func Sign(output io.Writer, signers []*Entity, hints *FileHints, config *packet.Config) (input io.WriteCloser, err error) {
	return SignWithParams(output, signers, &SignParams{
		Config: config,
		Hints:  hints,
	})
}

// signatureWriter hashes the contents of a message while passing it along to
// literalData. When closed, it closes literalData, writes a signature packet
// to encryptedData and then also closes encryptedData.
type signatureWriter struct {
	encryptedData      io.WriteCloser
	literalData        io.WriteCloser
	signatureContexts  []*signatureContext
	sigType            packet.SignatureType
	config             *packet.Config
	metadata           *packet.LiteralData // V5 signatures protect document metadata
	intendedRecipients []*packet.Recipient
	flag               int
}

type signatureContext struct {
	hashType    crypto.Hash
	wrappedHash hash.Hash
	h           hash.Hash
	salt        []byte // v6 only
	signer      *packet.PrivateKey
	outsideSig  *packet.Signature
}

func (s signatureWriter) Write(data []byte) (int, error) {
	for _, ctx := range s.signatureContexts {
		if _, err := ctx.wrappedHash.Write(data); err != nil {
			return 0, err
		}
	}
	switch s.sigType {
	case packet.SigTypeBinary:
		return s.literalData.Write(data)
	case packet.SigTypeText:
		return writeCanonical(s.literalData, data, &s.flag)
	}
	return 0, errors.UnsupportedError("unsupported signature type: " + strconv.Itoa(int(s.sigType)))
}

func (s signatureWriter) Close() error {
	if err := s.literalData.Close(); err != nil {
		return err
	}
	for _, ctx := range s.signatureContexts {
		var sig *packet.Signature
		if ctx.outsideSig != nil {
			// Signature that was supplied outside
			sig = ctx.outsideSig
		} else {
			sig = createSignaturePacket(&ctx.signer.PublicKey, s.sigType, s.config)
			sig.Hash = ctx.hashType
			sig.Metadata = s.metadata
			sig.IntendedRecipients = s.intendedRecipients
			if err := sig.SetSalt(ctx.salt); err != nil {
				return err
			}
			if err := sig.Sign(ctx.h, ctx.signer, s.config); err != nil {
				return err
			}
		}
		if err := sig.Serialize(s.encryptedData); err != nil {
			return err
		}
	}
	return s.encryptedData.Close()
}

func selectHashForSigningKey(config *packet.Config, primary *packet.PublicKey) crypto.Hash {
	acceptableHashes := acceptableHashesToWrite(primary)
	hash, ok := algorithm.HashToHashId(config.Hash())
	if !ok {
		return config.Hash()
	}
	for _, acceptableHashes := range acceptableHashes {
		if acceptableHashes == hash {
			return config.Hash()
		}
	}
	if len(acceptableHashes) > 0 {
		defaultAcceptedHash, ok := algorithm.HashIdToHash(acceptableHashes[0])
		if ok {
			return defaultAcceptedHash
		}
	}
	return config.Hash()
}

func createSignaturePacket(signer *packet.PublicKey, sigType packet.SignatureType, config *packet.Config) *packet.Signature {
	sigLifetimeSecs := config.SigLifetime()
	hash := selectHashForSigningKey(config, signer)
	return &packet.Signature{
		Version:           signer.Version,
		SigType:           sigType,
		PubKeyAlgo:        signer.PubKeyAlgo,
		Hash:              hash,
		CreationTime:      config.Now(),
		IssuerKeyId:       &signer.KeyId,
		IssuerKeyVersion:  uint8(signer.Version),
		IssuerFingerprint: signer.Fingerprint,
		Notations:         config.Notations(),
		SigLifetimeSecs:   &sigLifetimeSecs,
	}
}

// noOpCloser is like an ioutil.NopCloser, but for an io.Writer.
// TODO: we have two of these in OpenPGP packages alone. This probably needs
// to be promoted somewhere more common.
type noOpCloser struct {
	w io.Writer
}

func (c noOpCloser) Write(data []byte) (n int, err error) {
	return c.w.Write(data)
}

func (c noOpCloser) Close() error {
	return nil
}

func handleCompression(compressed io.WriteCloser, candidateCompression []uint8, config *packet.Config) (data io.WriteCloser, err error) {
	data = compressed
	confAlgo := config.Compression()
	if confAlgo == packet.CompressionNone {
		return
	}

	// Set algorithm labelled as MUST as fallback
	// https://www.ietf.org/archive/id/draft-ietf-openpgp-crypto-refresh-07.html#section-9.4
	finalAlgo := packet.CompressionNone
	// if compression specified by config available we will use it
	for _, c := range candidateCompression {
		if uint8(confAlgo) == c {
			finalAlgo = confAlgo
			break
		}
	}

	if finalAlgo != packet.CompressionNone {
		var compConfig *packet.CompressionConfig
		if config != nil {
			compConfig = config.CompressionConfig
		}
		data, err = packet.SerializeCompressed(compressed, finalAlgo, compConfig)
		if err != nil {
			return
		}
	}
	return data, nil
}

// selectHash selects the preferred hash given the candidateHashes and the configuredHash
func selectHash(candidateHashes []byte, configuredHash crypto.Hash, signer *packet.PrivateKey) (hash crypto.Hash, err error) {
	acceptableHashes := acceptableHashesToWrite(&signer.PublicKey)
	candidateHashes = intersectPreferences(acceptableHashes, candidateHashes)

	for _, hashId := range candidateHashes {
		if h, ok := algorithm.HashIdToHash(hashId); ok && h.Available() {
			hash = h
			break
		}
	}

	// If the hash specified by config is a candidate, we'll use that.
	if configuredHash.Available() {
		for _, hashId := range candidateHashes {
			if h, ok := algorithm.HashIdToHash(hashId); ok && h == configuredHash {
				hash = h
				break
			}
		}
	}

	if hash == 0 {
		if len(acceptableHashes) > 0 {
			if h, ok := algorithm.HashIdToHash(acceptableHashes[0]); ok {
				hash = h
			} else {
				return 0, errors.UnsupportedError("no candidate hash functions are compiled in.")
			}
		} else {
			return 0, errors.UnsupportedError("no candidate hash functions are compiled in.")
		}
	}
	return
}

func parseOutsideSig(outsideSig []byte) (outSigPacket *packet.Signature, err error) {
	packets := packet.NewReader(bytes.NewReader(outsideSig))
	p, err := packets.Next()
	if goerrors.Is(err, io.EOF) {
		return nil, errors.ErrUnknownIssuer
	}
	if err != nil {
		return nil, err
	}

	var ok bool
	outSigPacket, ok = p.(*packet.Signature)
	if !ok {
		return nil, errors.StructuralError("non signature packet found")
	}
	if outSigPacket.IssuerKeyId == nil {
		return nil, errors.StructuralError("signature doesn't have an issuer")
	}
	return outSigPacket, nil
}

func acceptableHashesToWrite(singingKey *packet.PublicKey) []uint8 {
	switch singingKey.PubKeyAlgo {
	case packet.PubKeyAlgoEd448, packet.PubKeyAlgoMldsa87Ed448, packet.PubKeyAlgoSlhdsaShake256s:
		return []uint8{
			hashToHashId(crypto.SHA512),
			hashToHashId(crypto.SHA3_512),
		}
	case packet.PubKeyAlgoECDSA, packet.PubKeyAlgoEdDSA:
		if curve, err := singingKey.Curve(); err == nil {
			if curve == packet.Curve448 ||
				curve == packet.CurveNistP521 ||
				curve == packet.CurveBrainpoolP512 {
				return []uint8{
					hashToHashId(crypto.SHA512),
					hashToHashId(crypto.SHA3_512),
				}
			} else if curve == packet.CurveBrainpoolP384 ||
				curve == packet.CurveNistP384 {
				return []uint8{
					hashToHashId(crypto.SHA384),
					hashToHashId(crypto.SHA512),
					hashToHashId(crypto.SHA3_512),
				}
			}
		}
	}
	return []uint8{
		hashToHashId(crypto.SHA256),
		hashToHashId(crypto.SHA384),
		hashToHashId(crypto.SHA512),
		hashToHashId(crypto.SHA3_256),
		hashToHashId(crypto.SHA3_512),
	}
}
