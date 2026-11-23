package v2

import (
	"io"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp/errors"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

// Subkey is an additional public key in an Entity. Subkeys can be used for
// encryption.
type Subkey struct {
	Primary     *Entity
	PublicKey   *packet.PublicKey
	PrivateKey  *packet.PrivateKey
	Bindings    []*packet.VerifiableSignature
	Revocations []*packet.VerifiableSignature
}

func readSubkey(primary *Entity, packets *packet.Reader, pub *packet.PublicKey, priv *packet.PrivateKey) error {
	subKey := Subkey{
		PublicKey:  pub,
		PrivateKey: priv,
		Primary:    primary,
	}

	for {
		p, err := packets.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return errors.StructuralError("subkey signature invalid: " + err.Error())
		}

		sig, ok := p.(*packet.Signature)
		if !ok {
			packets.Unread(p)
			break
		}

		if sig.SigType != packet.SigTypeSubkeyBinding && sig.SigType != packet.SigTypeSubkeyRevocation {
			// Ignore signatures with wrong type
			continue
		}
		switch sig.SigType {
		case packet.SigTypeSubkeyRevocation:
			subKey.Revocations = append(subKey.Revocations, packet.NewVerifiableSig(sig))
		case packet.SigTypeSubkeyBinding:
			subKey.Bindings = append(subKey.Bindings, packet.NewVerifiableSig(sig))
		}
	}
	primary.Subkeys = append(primary.Subkeys, subKey)
	return nil
}

// Serialize serializes the subkey and writes it into writer.
// The includeSecrets flag controls if the secrets should be included in the encoding or not.
func (s *Subkey) Serialize(w io.Writer, includeSecrets bool) error {
	if includeSecrets {
		if err := s.PrivateKey.Serialize(w); err != nil {
			return err
		}
	} else {
		if err := s.PublicKey.Serialize(w); err != nil {
			return err
		}
	}
	for _, revocation := range s.Revocations {
		if err := revocation.Packet.Serialize(w); err != nil {
			return err
		}
	}
	for _, bindingSig := range s.Bindings {
		if err := bindingSig.Packet.Serialize(w); err != nil {
			return err
		}
	}
	return nil
}

// ReSign resigns the latest valid subkey binding signature with the given config.
func (s *Subkey) ReSign(config *packet.Config) error {
	selectedSig, err := s.LatestValidBindingSignature(time.Time{}, config)
	if err != nil {
		return err
	}
	err = selectedSig.SignKey(s.PublicKey, s.Primary.PrivateKey, config)
	if err != nil {
		return err
	}
	if selectedSig.EmbeddedSignature != nil {
		err = selectedSig.EmbeddedSignature.CrossSignKey(s.PublicKey, s.Primary.PrimaryKey,
			s.PrivateKey, config)
		if err != nil {
			return err
		}
	}
	return nil
}

// Verify checks if the subkey is valid by checking:
// - that the key is not revoked
// - that there is valid non-expired binding self-signature
// - that the subkey is not expired
// If date is zero (i.e., date.IsZero() == true) the time checks are not performed.
func (s *Subkey) Verify(date time.Time, config *packet.Config) (selfSig *packet.Signature, err error) {
	selfSig, err = s.LatestValidBindingSignature(date, config)
	if err != nil {
		return nil, err
	}
	if s.Revoked(selfSig, date) {
		return nil, errors.ErrKeyRevoked
	}
	if !date.IsZero() && s.Expired(selfSig, date) {
		return nil, errors.ErrKeyExpired
	}
	return
}

// Expired checks if given the selected self-signature if the subkey is expired.
func (s *Subkey) Expired(selectedSig *packet.Signature, date time.Time) bool {
	return s.PublicKey.KeyExpired(selectedSig, date) || selectedSig.SigExpired(date)
}

// Revoked returns whether the subkey has been revoked by a self-signature.
// Note that third-party revocation signatures are not supported.
func (s *Subkey) Revoked(selfCertification *packet.Signature, date time.Time) bool {
	// Verify revocations first
	for _, revocation := range s.Revocations {
		if revocation.Valid == nil {
			err := s.Primary.PrimaryKey.VerifySubkeyRevocationSignature(revocation.Packet, s.PublicKey)
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
		if *revocation.Valid && (date.IsZero() ||
			!revocation.Packet.SigExpired(date) &&
				(selfCertification == nil ||
					selfCertification.CreationTime.Unix() <= revocation.Packet.CreationTime.Unix())) {
			return true
		}
	}
	return false
}

// Revoke generates a subkey revocation signature (packet.SigTypeSubkeyRevocation) for
// a subkey with the specified reason code and text (RFC4880 section-5.2.3.23).
// If config is nil, sensible defaults will be used.
func (s *Subkey) Revoke(reason packet.ReasonForRevocation, reasonText string, config *packet.Config) error {
	// Check that the subkey is valid (not considering expiration)
	if _, err := s.Verify(time.Time{}, config); err != nil {
		return err
	}

	revSig := createSignaturePacket(s.Primary.PrimaryKey, packet.SigTypeSubkeyRevocation, config)
	revSig.RevocationReason = &reason
	revSig.RevocationReasonText = reasonText

	if err := revSig.RevokeSubkey(s.PublicKey, s.Primary.PrivateKey, config); err != nil {
		return err
	}
	sig := packet.NewVerifiableSig(revSig)
	valid := true
	sig.Valid = &valid
	s.Revocations = append(s.Revocations, sig)
	return nil
}

// LatestValidBindingSignature returns the latest valid self-signature of this subkey
// that is not newer than the provided date.
// Does not consider signatures/embedded signatures that are expired.
// If date is zero (i.e., date.IsZero() == true) the expiration checks are not performed.
// Returns a StructuralError if no valid self-signature is found.
func (s *Subkey) LatestValidBindingSignature(date time.Time, config *packet.Config) (selectedSig *packet.Signature, err error) {
	for sigIdx := len(s.Bindings) - 1; sigIdx >= 0; sigIdx-- {
		sig := s.Bindings[sigIdx]
		if (date.IsZero() || date.Unix() >= sig.Packet.CreationTime.Unix()) &&
			(selectedSig == nil || selectedSig.CreationTime.Unix() < sig.Packet.CreationTime.Unix()) {
			if sig.Valid == nil {
				err := s.Primary.PrimaryKey.VerifyKeySignature(s.PublicKey, sig.Packet)
				if err == nil {
					err = checkSignatureDetails(sig.Packet, date, config)
				}
				valid := err == nil
				sig.Valid = &valid
			}
			mainSigExpired := !date.IsZero() &&
				sig.Packet.SigExpired(date)
			embeddedSigExpired := !date.IsZero() &&
				sig.Packet.FlagSign &&
				sig.Packet.EmbeddedSignature != nil &&
				sig.Packet.EmbeddedSignature.SigExpired(date)
			if *sig.Valid && !mainSigExpired && !embeddedSigExpired {
				selectedSig = sig.Packet
			}
		}
	}
	if selectedSig == nil {
		return nil, errors.StructuralError("no valid binding signature found for subkey")
	}
	return selectedSig, nil
}

// IsPQ returns true if the algorithm is Post-Quantum safe
func (s *Subkey) IsPQ() bool {
	switch s.PublicKey.PubKeyAlgo {
	case packet.PubKeyAlgoMlkem768X25519, packet.PubKeyAlgoMlkem1024X448,
		packet.PubKeyAlgoMldsa65Ed25519, packet.PubKeyAlgoMldsa87Ed448,
		packet.PubKeyAlgoSlhdsaShake128s, packet.PubKeyAlgoSlhdsaShake128f, packet.PubKeyAlgoSlhdsaShake256s:
		return true
	default:
		return false
	}

}
