// SPDX-License-Identifier: BSD-3-Clause
package tls

const keyLogLabelECHInner = "UTLS_ECH_INNER_CLIENT_HELLO"
const keyLogLabelClientEarly = "CLIENT_EARLY_TRAFFIC_SECRET"

// logECHInner records a handshake send attempt, never a speculative build.
// The outer random associates the expanded private hello with its wire record.
// This diagnostic label is emitted only through an explicit KeyLogWriter.
func (c *UConn) logECHInner(outer *clientHelloMsg, ech *echClientContext) error {
	if ech == nil || !ech.customInner || c.config.KeyLogWriter == nil {
		return nil
	}

	// Observe the already-bound transcript without rebuilding the hello.
	return c.config.writeKeyLog(keyLogLabelECHInner, outer.random, ech.innerHello.original)
}
