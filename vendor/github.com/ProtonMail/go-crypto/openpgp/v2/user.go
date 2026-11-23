package v2

import (
	"io"
	"time"

	"github.com/ProtonMail/go-crypto/openpgp/errors"
	"github.com/ProtonMail/go-crypto/openpgp/packet"
)

// Identity represents an identity claimed by an Entity and zero or more
// assertions by other entities about that claim.
type Identity struct {
	Primary             *Entity
	Name                string // by convention, has the form "Full Name (comment) <email@example.com>"
	UserId              *packet.UserId
	SelfCertifications  []*packet.VerifiableSignature
	OtherCertifications []*packet.VerifiableSignature
	Revocations         []*packet.VerifiableSignature
}

func readUser(e *Entity, packets *packet.Reader, pkt *packet.UserId) error {
	identity := Identity{
		Primary: e,
		Name:    pkt.Id,
		UserId:  pkt,
	}
	for {
		p, err := packets.NextWithUnsupported()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		unsupportedPacket, unsupported := p.(*packet.UnsupportedPacket)
		sigCandidate := p
		if unsupported {
			sigCandidate = unsupportedPacket.IncompletePacket
		}
		sig, ok := sigCandidate.(*packet.Signature)
		if !ok {
			// sigCandidate is a not a signature packet, reset and stop.
			packets.Unread(p)
			break
		} else if unsupported {
			// sigCandidate is a signature packet but unsupported.
			continue
		}

		if sig.SigType != packet.SigTypeGenericCert &&
			sig.SigType != packet.SigTypePersonaCert &&
			sig.SigType != packet.SigTypeCasualCert &&
			sig.SigType != packet.SigTypePositiveCert &&
			sig.SigType != packet.SigTypeCertificationRevocation {
			// Ignore signatures with wrong type
			continue
		}

		if sig.CheckKeyIdOrFingerprint(e.PrimaryKey) {
			if sig.SigType == packet.SigTypeCertificationRevocation {
				identity.Revocations = append(identity.Revocations, packet.NewVerifiableSig(sig))
			} else {
				identity.SelfCertifications = append(identity.SelfCertifications, packet.NewVerifiableSig(sig))
			}
			e.Identities[pkt.Id] = &identity
		} else {
			identity.OtherCertifications = append(identity.OtherCertifications, packet.NewVerifiableSig(sig))
		}
	}
	return nil
}

// Serialize serializes the user id to the writer.
func (i *Identity) Serialize(w io.Writer) error {
	if err := i.UserId.Serialize(w); err != nil {
		return err
	}
	for _, sig := range i.Revocations {
		if err := sig.Packet.Serialize(w); err != nil {
			return err
		}
	}
	for _, sig := range i.SelfCertifications {
		if err := sig.Packet.Serialize(w); err != nil {
			return err
		}
	}
	for _, sig := range i.OtherCertifications {
		if err := sig.Packet.Serialize(w); err != nil {
			return err
		}
	}
	return nil
}

// Verify checks if the user-id is valid by checking:
// - that a valid self-certification exists and is not expired
// - that user-id has not been revoked at the given point in time
// If date is zero (i.e., date.IsZero() == true) the time checks are not performed.
func (i *Identity) Verify(date time.Time, config *packet.Config) (selfSignature *packet.Signature, err error) {
	if selfSignature, err = i.LatestValidSelfCertification(date, config); err != nil {
		return
	}
	if i.Revoked(selfSignature, date, config) {
		return nil, errors.StructuralError("user-id is revoked")
	}
	return
}

// Revoked returns whether the identity has been revoked by a self-signature.
// Note that third-party revocation signatures are not supported.
func (i *Identity) Revoked(selfCertification *packet.Signature, date time.Time, config *packet.Config) bool {
	// Verify revocations first
	for _, revocation := range i.Revocations {
		if selfCertification == nil || // if there is not selfCertification verify revocation
			selfCertification.IssuerKeyId == nil ||
			revocation.Packet.IssuerKeyId == nil ||
			(*selfCertification.IssuerKeyId == *revocation.Packet.IssuerKeyId) { // check matching key id
			if revocation.Valid == nil {
				// Verify revocation signature (not verified yet).
				err := i.Primary.PrimaryKey.VerifyUserIdSignature(i.Name, i.Primary.PrimaryKey, revocation.Packet)
				if err == nil {
					err = checkSignatureDetails(revocation.Packet, date, config)
				}
				valid := err == nil
				revocation.Valid = &valid
			}

			if *revocation.Valid &&
				(date.IsZero() || // Check revocation not expired
					!revocation.Packet.SigExpired(date)) &&
				(selfCertification == nil || // Check that revocation is not older than the selfCertification
					selfCertification.CreationTime.Unix() <= revocation.Packet.CreationTime.Unix()) {
				return true
			}
		}
	}
	return false
}

// ReSign resigns the latest valid self-certification with the given config.
func (i *Identity) ReSign(config *packet.Config) error {
	selectedSig, err := i.LatestValidSelfCertification(config.Now(), config)
	if err != nil {
		return err
	}
	if err = selectedSig.SignUserId(
		i.UserId.Id,
		i.Primary.PrimaryKey,
		i.Primary.PrivateKey,
		config,
	); err != nil {
		return err
	}
	return nil
}

// SignIdentity adds a signature to e, from signer, attesting that identity is
// associated with e. The provided identity must already be an element of
// e.Identities and the private key of signer must have been decrypted if
// necessary.
// If config is nil, sensible defaults will be used.
func (ident *Identity) SignIdentity(signer *Entity, config *packet.Config) error {
	certificationKey, ok := signer.CertificationKey(config.Now(), config)
	if !ok {
		return errors.InvalidArgumentError("no valid certification key found")
	}

	if certificationKey.PrivateKey.Encrypted {
		return errors.InvalidArgumentError("signing Entity's private key must be decrypted")
	}

	if !ok {
		return errors.InvalidArgumentError("given identity string not found in Entity")
	}

	sig := createSignaturePacket(certificationKey.PublicKey, packet.SigTypeGenericCert, config)

	signingUserID := config.SigningUserId()
	if signingUserID != "" {
		if _, ok := signer.Identities[signingUserID]; !ok {
			return errors.InvalidArgumentError("signer identity string not found in signer Entity")
		}
		sig.SignerUserId = &signingUserID
	}

	if err := sig.SignUserId(ident.Name, ident.Primary.PrimaryKey, certificationKey.PrivateKey, config); err != nil {
		return err
	}
	ident.OtherCertifications = append(ident.OtherCertifications, packet.NewVerifiableSig(sig))
	return nil
}

// LatestValidSelfCertification returns the latest valid self-signature of this user-id
// that is not newer than the provided date.
// Does not consider signatures that are expired.
// If date is zero (i.e., date.IsZero() == true) the expiration checks are not performed.
// Returns a StructuralError if no valid self-certification is found.
func (i *Identity) LatestValidSelfCertification(date time.Time, config *packet.Config) (selectedSig *packet.Signature, err error) {
	for sigIdx := len(i.SelfCertifications) - 1; sigIdx >= 0; sigIdx-- {
		sig := i.SelfCertifications[sigIdx]
		if (date.IsZero() || date.Unix() >= sig.Packet.CreationTime.Unix()) && // SelfCertification must be older than date
			(selectedSig == nil || selectedSig.CreationTime.Unix() < sig.Packet.CreationTime.Unix()) { // Newer ones are preferred
			if sig.Valid == nil {
				// Verify revocation signature (not verified yet).
				err = i.Primary.PrimaryKey.VerifyUserIdSignature(i.Name, i.Primary.PrimaryKey, sig.Packet)
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
		return nil, errors.StructuralError("no valid certification signature found for identity")
	}
	return selectedSig, nil
}
