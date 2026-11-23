package alps

// RequestPreparationError identifies a local, unsent protocol/header-policy failure.
// A protocol race must return it without retrying the body over another route.
type RequestPreparationError struct {
	Err error
}

func (e *RequestPreparationError) Error() string { return e.Err.Error() }
func (e *RequestPreparationError) Unwrap() error { return e.Err }
