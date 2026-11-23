package tls

import "crypto/x509"

// Verification appends the peer's certificates to a private candidate pool.
// Sharing or promoting that pool to RootCAs would change caller-owned trust.
func (c *Config) certificateIntermediates() *x509.CertPool {
	if c.CertificateIntermediates != nil {
		return c.CertificateIntermediates.Clone()
	}
	return x509.NewCertPool()
}
