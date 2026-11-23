package alps

import (
	"context"

	"github.com/enetx/http"
)

type requestPolicyKey struct{}

// WithRequestPolicy carries a client-owned policy to the selected connection.
// It carries no mutable connection slot and is safe across TCP/QUIC racing.
// The callback receives a private, unsent request and must not read its body.
func WithRequestPolicy(ctx context.Context, apply func(*http.Request, string) error) context.Context {
	return context.WithValue(ctx, requestPolicyKey{}, apply)
}

// CopyRequestPolicy retains the logical request's policy across HTTP adapters
// and redirects. Connection selection remains local to each protocol attempt.
func CopyRequestPolicy(destination, source context.Context) context.Context {
	if apply, present := source.Value(requestPolicyKey{}).(func(*http.Request, string) error); present {
		return WithRequestPolicy(destination, apply)
	}

	return destination
}

// Prepare invokes a request policy only for a matched authenticated origin.
// Missing policy/connection instructions are a no-op, including native TLS.
func Prepare(req *http.Request, snapshot *Snapshot) error {
	apply, _ := req.Context().Value(requestPolicyKey{}).(func(*http.Request, string) error)
	if apply == nil {
		return nil
	}

	if value, found := snapshot.AcceptCH(req.URL); found {
		return apply(req, value)
	}

	return nil
}
