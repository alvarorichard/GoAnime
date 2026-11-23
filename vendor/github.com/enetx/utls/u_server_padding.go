package tls

import (
	"encoding/binary"
	"errors"
)

const extensionServerPadding uint16 = 0x12e0

// Validate against the actual negotiated ClientHello, including an accepted
// ECH inner, rather than a template that an application may have overridden.
// The response length must equal the requested uint16; contents are opaque.
// https://github.com/google/boringssl/blob/ea9f3a34c3977144c45d272be81842fd891a0bdb/ssl/extensions.cc
func (hs *clientHandshakeStateTLS13) readServerPadding(message *encryptedExtensionsMsg) error {
	if !message.utls.serverPaddingPresent {
		return nil
	}
	// uTLS retains the exact wire bytes. The native Go construction may keep
	// only parsed fields, so obtain its encoding through the same TLS marshaller.
	raw, err := hs.hello.marshal()
	var extensions []rawExtension
	if err == nil {
		extensions, err = extractRawExtensions(&clientHelloMsg{original: raw})
	}

	if err != nil {
		hs.c.sendAlert(alertInternalError)
		return err
	}
	for _, extension := range extensions {
		if extension.extType != extensionServerPadding {
			continue
		}
		if len(extension.data) != 2 || int(binary.BigEndian.Uint16(extension.data)) != message.utls.serverPaddingLength {
			hs.c.sendAlert(alertDecodeError)
			return errors.New("tls: server padding length does not match request")
		}
		return nil
	}
	hs.c.sendAlert(alertUnsupportedExtension)
	return errors.New("tls: server sent unrequested server padding")
}
