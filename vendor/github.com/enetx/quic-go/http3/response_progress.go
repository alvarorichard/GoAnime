package http3

// A received header commits this attempt. Callback errors may wrap a transport
// error code, but that does not turn them into a peer's unsent-request rejection.
type responseReceivedError struct{ err error }

func (e *responseReceivedError) Error() string { return e.err.Error() }
func (e *responseReceivedError) Unwrap() error { return e.err }
