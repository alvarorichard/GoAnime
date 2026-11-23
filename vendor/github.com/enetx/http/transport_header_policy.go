package http

import (
	"context"
	"fmt"
)

type requestHeaderPolicyKey struct{}

// prepareRequestHeaders keeps protocol-specific defaults out of the caller's
// request, and out of subsequent retries that might select another protocol.
func prepareRequestHeaders(req *Request, policy func(*Request, int), protocol int) (*Request, error) {
	clone := *req
	clone.Header = req.Header.Clone()
	policy(&clone, protocol)
	if err := context.Cause(clone.Context()); err != nil {
		return nil, err
	}

	if invalid := validateHeaders(clone.Header); invalid != "" {
		return nil, fmt.Errorf("net/http: invalid header %s", invalid)
	}
	return &clone, nil
}
