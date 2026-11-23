// SPDX-License-Identifier: BSD-3-Clause
package tls

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"slices"

	"golang.org/x/crypto/cryptobyte"
)

// ECHClientHelloInfo is an owned snapshot of the current outer template.
// ClientHello contains a complete handshake message before ECH encryption.
// ServerName is the private name, while the outer template may carry public_name.
type ECHClientHelloInfo struct {
	ClientHello []byte
	ServerName  string
	NextProtos  []string
	QUIC        bool
	// QUICTransportParameters is the complete engine-owned inner offer.
	// It can differ from the public transport parameters when ECH is used.
	QUICTransportParameters []byte
}

// ECHClientHelloInnerFunc chooses a real inner hello without owning cryptography.
// uTLS retains random generation, key shares, HPKE, transcripts and verification.
type ECHClientHelloInnerFunc func(ECHClientHelloInfo) (*ECHClientHelloSpec, error)

// ECHClientHelloSpec defines the inner cipher order and expanded extension order.
// Outer references must form one contiguous run. The engine appends and binds
// its own resumption extension after those references.
type ECHClientHelloSpec struct {
	OuterPSK     ECHOuterPSKMode
	CipherSuites []uint16
	Extensions   []ECHClientHelloExtension
	// OuterQUICTransportParameters replaces the public QUIC parameters.
	// Nil preserves the template. The inner must retain the engine's full offer
	// as a literal extension when this field is set.
	OuterQUICTransportParameters []byte
}

// ECHClientHelloExtension is a literal extension or a reference to the outer.
// Outer references have no Data and reuse the exact outer bytes. Literal Data
// excludes the extension's four-byte header. ECH itself cannot be compressed.
type ECHClientHelloExtension struct {
	ID    uint16
	Data  []byte
	Outer bool
}

func (uconn *UConn) buildECHInner(ech *echClientContext, random []byte) (*clientHelloMsg, error) {
	if err := uconn.MarshalClientHelloNoECH(); err != nil {
		return nil, err
	}
	outer := new(clientHelloMsg)
	if !outer.unmarshal(uconn.HandshakeState.Hello.Raw) {
		return nil, errors.New("tls: malformed outer ClientHello template")
	}
	rawExtensions, err := extractRawExtensions(outer)
	if err != nil {
		return nil, err
	}
	byID := make(map[uint16][]byte, len(rawExtensions))
	outerPositions := make(map[uint16]int, len(rawExtensions))
	// clientHelloMsg.unmarshal already rejects duplicate extension IDs.
	for i, ext := range rawExtensions {
		byID[ext.extType] = ext.data
		outerPositions[ext.extType] = i
	}
	policy := uconn.config.ECHClientHelloInner
	if policy == nil {
		return nil, errors.New("tls: custom ECH inner callback is nil")
	}
	info := ECHClientHelloInfo{
		ClientHello: bytes.Clone(outer.original),
		ServerName:  hostnameInSNI(uconn.config.ServerName),
		NextProtos:  slices.Clone(uconn.config.NextProtos),
		QUIC:        uconn.quic != nil,
	}
	if uconn.quic != nil {
		info.QUICTransportParameters = bytes.Clone(uconn.quic.transportParams)
	}
	spec, err := policy(info)
	if err != nil {
		return nil, fmt.Errorf("tls: build ECH inner: %w", err)
	}
	if spec == nil || len(spec.CipherSuites) == 0 {
		return nil, errors.New("tls: empty ECH inner cipher list")
	}
	if spec.OuterPSK > ECHOuterPSKRandomized {
		return nil, errors.New("tls: invalid ECH outer PSK mode")
	}
	hasCipher := false
	for _, id := range spec.CipherSuites {
		// Legacy cipher identifiers are permitted in a TLS 1.3-only hello.
		// Negotiation is constrained by supported_versions below, not by
		// requiring every advertised cipher to be a TLS 1.3 cipher.
		hasCipher = hasCipher || cipherSuiteTLS13ByID(id) != nil
	}
	if !hasCipher {
		return nil, errors.New("tls: ECH inner lacks a usable cipher suite")
	}
	seen := make(map[uint16]bool, len(spec.Extensions))
	compressed := make([]uint16, 0, len(spec.Extensions))
	extensions := make([]TLSExtension, 0, len(spec.Extensions))
	compressionEnded := false
	lastOuterPosition := -1
	for _, ext := range spec.Extensions {
		if seen[ext.ID] {
			return nil, fmt.Errorf("tls: duplicate ECH inner extension %d", ext.ID)
		}
		seen[ext.ID] = true
		if ext.ID == extensionECHOuterExtensions || ext.ID == extensionPreSharedKey {
			return nil, fmt.Errorf("tls: unsupported literal ECH inner extension %d", ext.ID)
		}
		data := ext.Data
		if ext.Outer {
			if compressionEnded || len(ext.Data) != 0 {
				return nil, errors.New("tls: invalid ECH outer reference run")
			}
			if ext.ID == extensionEncryptedClientHello {
				return nil, errors.New("tls: ECH marker cannot reference the outer")
			}
			var found bool
			data, found = byID[ext.ID]
			if !found {
				return nil, fmt.Errorf("tls: ECH outer reference %d is absent", ext.ID)
			}
			if outerPositions[ext.ID] <= lastOuterPosition {
				return nil, errors.New("tls: ECH outer references changed extension order")
			}
			lastOuterPosition = outerPositions[ext.ID]
			compressed = append(compressed, ext.ID)
		} else if len(compressed) > 0 {
			compressionEnded = true
		}
		extensions = append(extensions, &GenericExtension{Id: ext.ID, Data: bytes.Clone(data)})
	}
	if !seen[extensionEncryptedClientHello] {
		return nil, errors.New("tls: ECH inner lacks its marker")
	}
	psk, common, err := uconn.echPSKExtension(ech)
	if err != nil {
		return nil, err
	}
	if psk != nil {
		extensions = append(extensions, psk)
	}
	raw, err := marshalClientHelloMsg(outer.vers, random, outer.sessionId, spec.CipherSuites, outer.compressionMethods, extensions)
	if err != nil {
		return nil, err
	}
	inner, err := parseECHInnerClientHello(raw, common, ech.innerTranscript)
	if err != nil {
		return nil, err
	}
	if spec.OuterQUICTransportParameters != nil {
		if uconn.quic == nil || slices.Contains(compressed, extensionQUICTransportParameters) ||
			!seen[extensionQUICTransportParameters] || !bytes.Equal(inner.quicTransportParameters, uconn.quic.transportParams) {
			return nil, errors.New("tls: ECH inner must retain the full QUIC transport parameters")
		}
	}
	if !bytes.Equal(inner.encryptedClientHello, []byte{1}) || inner.serverName != hostnameInSNI(uconn.config.ServerName) {
		return nil, errors.New("tls: invalid ECH inner marker or server name")
	}
	if !slices.Contains(inner.supportedVersions, VersionTLS13) {
		return nil, errors.New("tls: ECH inner lacks TLS 1.3")
	}
	for _, version := range inner.supportedVersions {
		if version != VersionTLS13 && !isGREASEUint16(version) {
			return nil, errors.New("tls: ECH inner offers a legacy version")
		}
	}
	// The private keys were generated for the template, not for arbitrary callback bytes.
	if len(inner.keyShares) != len(outer.keyShares) {
		return nil, errors.New("tls: ECH inner changed key shares")
	}
	for i, share := range inner.keyShares {
		if share.group != outer.keyShares[i].group || !bytes.Equal(share.data, outer.keyShares[i].data) {
			return nil, errors.New("tls: ECH inner changed key shares")
		}
	}
	ech.customInner = true
	ech.outerPSK = spec.OuterPSK
	ech.outerQUICTransportParameters = bytes.Clone(spec.OuterQUICTransportParameters)
	ech.innerOuterExts = compressed
	return inner, nil
}

// Raw encoding follows RFC 9849 section 6.1 and upstream uTLS PR #432.
// https://github.com/refraction-networking/utls/pull/432
// marshalClientHelloMsg serializes a ClientHello handshake message, header
// included, from its fields and marshaled extensions.
func marshalClientHelloMsg(vers uint16, random, sessionID []byte, cipherSuites []uint16, compressionMethods []byte, exts []TLSExtension) ([]byte, error) {
	var b cryptobyte.Builder
	b.AddUint8(typeClientHello)
	b.AddUint24LengthPrefixed(func(b *cryptobyte.Builder) {
		b.AddUint16(vers)
		b.AddBytes(random)
		b.AddUint8LengthPrefixed(func(b *cryptobyte.Builder) { b.AddBytes(sessionID) })
		b.AddUint16LengthPrefixed(func(b *cryptobyte.Builder) {
			for _, id := range cipherSuites {
				b.AddUint16(id)
			}
		})
		b.AddUint8LengthPrefixed(func(b *cryptobyte.Builder) { b.AddBytes(compressionMethods) })
		b.AddUint16LengthPrefixed(func(b *cryptobyte.Builder) {
			for _, ext := range exts {
				buf := make([]byte, ext.Len())
				n, err := ext.Read(buf)
				if err != nil && !errors.Is(err, io.EOF) {
					b.SetError(err)
					return
				}
				b.AddBytes(buf[:n])
			}
		})
	})
	return b.Bytes()
}

// encodeInnerClientHelloFromRaw returns the EncodedClientHelloInner of inner
// (RFC 9849, section 6.1): its marshaled form with an empty session ID and
// the extensions listed in outerExts, one contiguous run in that order,
// replaced by ech_outer_extensions, then padded.
func encodeInnerClientHelloFromRaw(inner *clientHelloMsg, outerExts []uint16, maxNameLength int) ([]byte, error) {
	s := cryptobyte.String(inner.original)
	var versionAndRandom, sessionID, cipherSuites, compressionMethods []byte
	var extensions cryptobyte.String
	if !s.Skip(4) || !s.ReadBytes(&versionAndRandom, 2+32) ||
		!readUint8LengthPrefixed(&s, &sessionID) ||
		!readUint16LengthPrefixed(&s, &cipherSuites) ||
		!readUint8LengthPrefixed(&s, &compressionMethods) ||
		!s.ReadUint16LengthPrefixed(&extensions) || !s.Empty() {
		return nil, errors.New("tls: malformed inner client hello")
	}

	var b cryptobyte.Builder
	b.AddBytes(versionAndRandom)
	b.AddUint8(0) // the session ID is the outer's
	b.AddUint16LengthPrefixed(func(b *cryptobyte.Builder) { b.AddBytes(cipherSuites) })
	b.AddUint8LengthPrefixed(func(b *cryptobyte.Builder) { b.AddBytes(compressionMethods) })
	b.AddUint16LengthPrefixed(func(b *cryptobyte.Builder) {
		compressed := 0
		for !extensions.Empty() {
			var extType uint16
			var extData cryptobyte.String
			if !extensions.ReadUint16(&extType) || !extensions.ReadUint16LengthPrefixed(&extData) {
				b.SetError(errors.New("tls: malformed inner client hello extensions"))
				return
			}
			if compressed < len(outerExts) && extType == outerExts[compressed] {
				if compressed == 0 {
					b.AddUint16(extensionECHOuterExtensions)
					b.AddUint16LengthPrefixed(func(b *cryptobyte.Builder) {
						b.AddUint8LengthPrefixed(func(b *cryptobyte.Builder) {
							for _, id := range outerExts {
								b.AddUint16(id)
							}
						})
					})
				}
				compressed++
				continue
			}
			if compressed > 0 && compressed < len(outerExts) {
				b.SetError(errors.New("tls: compressed inner client hello extensions are not contiguous"))
				return
			}
			b.AddUint16(extType)
			b.AddUint16LengthPrefixed(func(b *cryptobyte.Builder) { b.AddBytes(extData) })
		}
		if compressed != len(outerExts) {
			b.SetError(errors.New("tls: inner client hello lacks a compressed extension"))
		}
	})
	h, err := b.Bytes()
	if err != nil {
		return nil, err
	}
	paddingLen := maxNameLength + 9
	if inner.serverName != "" {
		paddingLen = max(0, maxNameLength-len(inner.serverName))
	}
	paddingLen += 31 - ((len(h) + paddingLen - 1) % 32)
	return append(h, make([]byte, paddingLen)...), nil
}
