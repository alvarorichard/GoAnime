// SPDX-License-Identifier: BSD-3-Clause
package tls

import (
	"bytes"
	"errors"
	"hash"
	"io"

	"golang.org/x/crypto/cryptobyte"
)

// ECHOuterPSKMode controls the public placeholder for an inner session offer.
// Real identities and binders are only sent inside ECH (RFC 9849 section 6.1).
type ECHOuterPSKMode uint8

const (
	ECHOuterPSKOmit ECHOuterPSKMode = iota
	ECHOuterPSKRandomized
)

func (u *UConn) echPSKExtension(ech *echClientContext) (*GenericExtension, *PreSharedKeyCommon, error) {
	e := u.sessionController.pskExtension
	if e == nil || e.Len() == 0 {
		return nil, nil, nil
	}
	common := e.GetPreSharedKeyCommon()
	if common.Session == nil || common.Session.version != VersionTLS13 {
		return nil, nil, errors.New("tls: custom ECH PSK requires an engine-owned TLS 1.3 session")
	}
	suite := cipherSuiteTLS13ByID(common.Session.cipherSuite)
	if suite == nil || len(common.BinderKey) != suite.hash.Size() || len(common.EarlySecret) != suite.hash.Size() {
		return nil, nil, errors.New("tls: invalid custom ECH PSK key material")
	}
	identities := common.Identities
	if ech.innerHello != nil {
		// HRR may update the age or drop a PSK whose hash is incompatible
		// with the selected suite. Keep that engine decision.
		identities = mapSlice(ech.innerHello.pskIdentities, func(i pskIdentity) PskIdentity {
			return PskIdentity{Label: i.label, ObfuscatedTicketAge: i.obfuscatedTicketAge}
		})
	}
	if len(identities) == 0 {
		return nil, &common, nil
	}
	if len(identities) != 1 {
		return nil, nil, errors.New("tls: custom ECH supports one resumption identity")
	}
	binders := [][]byte{make([]byte, suite.hash.Size())}
	data := make([]byte, pskExtLen(identities, binders))
	// readPskIntoBytes can only fail for a short buffer. The exact size was
	// computed from these same identities and binders; success returns io.EOF.
	readPskIntoBytes(data, identities, binders)
	return &GenericExtension{Id: extensionPreSharedKey, Data: data[4:]}, &common, nil
}

// Parse and bind the same raw encoding so unknown extensions remain in the PSK
// transcript. Validation finishes before the caller publishes the inner hello.
func parseECHInnerClientHello(raw []byte, common *PreSharedKeyCommon, previous hash.Hash) (*clientHelloMsg, error) {
	inner := new(clientHelloMsg)
	if !inner.unmarshal(raw) {
		return nil, errors.New("tls: invalid ECH inner ClientHello")
	}

	if err := patchECHPSKBinder(inner, common, previous); err != nil {
		return nil, err
	}

	return inner, nil
}

func patchECHPSKBinder(inner *clientHelloMsg, common *PreSharedKeyCommon, previous hash.Hash) error {
	if common == nil || len(inner.pskIdentities) == 0 {
		return nil
	}
	suite := cipherSuiteTLS13ByID(common.Session.cipherSuite)
	transcript := suite.hash.New()
	if previous != nil {
		transcript = cloneHash(previous, suite.hash)
	}
	truncated, err := inner.marshalWithoutBinders()
	if err != nil {
		return err
	}
	transcript.Write(truncated)
	binder := suite.finishedHash(common.BinderKey, transcript)
	if len(inner.pskBinders) != 1 || len(inner.pskBinders[0]) != len(binder) {
		return errors.New("tls: invalid ECH inner binder placeholder")
	}
	// unmarshal retains slices of original, so this patches both the parsed
	// binder and the exact raw transcript, including unknown extensions.
	copy(inner.pskBinders[0], binder)
	return nil
}

func randomizedECHPSK(data []byte, random io.Reader) ([]byte, error) {
	data = bytes.Clone(data)
	s := cryptobyte.String(data)
	var identities, binders cryptobyte.String
	if !s.ReadUint16LengthPrefixed(&identities) || !s.ReadUint16LengthPrefixed(&binders) || !s.Empty() {
		return nil, errors.New("tls: malformed ECH PSK placeholder")
	}
	for !identities.Empty() {
		var label cryptobyte.String
		var age []byte
		if !identities.ReadUint16LengthPrefixed(&label) || !identities.ReadBytes(&age, 4) {
			return nil, errors.New("tls: malformed ECH PSK identity")
		}
		if _, err := io.ReadFull(random, label); err != nil {
			return nil, err
		}
		if _, err := io.ReadFull(random, age); err != nil {
			return nil, err
		}
	}
	for !binders.Empty() {
		var binder cryptobyte.String
		if !binders.ReadUint8LengthPrefixed(&binder) {
			return nil, errors.New("tls: malformed ECH PSK binder")
		}
		if _, err := io.ReadFull(random, binder); err != nil {
			return nil, err
		}
	}
	return data, nil
}

func (u *UConn) setECHOuterExtensions(inner *clientHelloMsg, ech *echClientContext) (func(), error) {
	original := u.Extensions
	var outerPSK []byte
	if ech.outerPSK == ECHOuterPSKRandomized && len(inner.pskIdentities) > 0 {
		extensions, err := extractRawExtensions(inner)
		if err != nil {
			return nil, err
		}
		for _, e := range extensions {
			if e.extType == extensionPreSharedKey {
				outerPSK, err = randomizedECHPSK(e.data, u.config.rand())
				if err != nil {
					return nil, err
				}
			}
		}
	}
	filtered := make([]TLSExtension, 0, len(original))
	for _, e := range original {
		if _, ok := e.(*QUICTransportParametersExtension); ok && ech.outerQUICTransportParameters != nil {
			filtered = append(filtered, &GenericExtension{Id: extensionQUICTransportParameters, Data: ech.outerQUICTransportParameters})
			continue
		}
		if _, ok := e.(PreSharedKeyExtension); ok {
			if outerPSK != nil {
				filtered = append(filtered, &GenericExtension{Id: extensionPreSharedKey, Data: outerPSK})
			}
			continue
		}
		filtered = append(filtered, e)
	}
	u.Extensions = filtered
	return func() { u.Extensions = original }, nil
}
