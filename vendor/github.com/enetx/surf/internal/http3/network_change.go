package h3transport

// RetireConnections removes current H3 sessions and coalescing authorizations
// from admission. Active response streams finish on their original connection.
// Tickets, confirmed support, RTT and negotiated-version history remain intact.
func (ut *Transport) RetireConnections() {
	if ut.racer != nil {
		ut.racer.ResetPreparations(false)
	}

	if ut.http3tr != nil {
		ut.http3tr.RetireConnections()
	}

	if ut.coalescer != nil {
		ut.coalescer.mu.Lock()
		clear(ut.coalescer.peers)
		ut.coalescer.mu.Unlock()
	}
}

// ClearFailedAlternatives discards failure exclusions without deleting valid
// advertisements, ECH requirements or TTL metadata. Late old attempts cannot
// repopulate the exclusions cleared by this notification.
func (ut *Transport) ClearFailedAlternatives() {
	if ut.racer != nil {
		ut.racer.ResetPreparations(true)
	}

	if ut.discovery != nil {
		ut.discovery.mu.Lock()
		ut.discovery.failureGeneration++
		clear(ut.discovery.failures)
		ut.discovery.mu.Unlock()
	}
}
