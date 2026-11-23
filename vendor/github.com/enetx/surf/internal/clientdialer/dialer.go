// Package clientdialer identifies Surf's default TCP dial functions.
package clientdialer

import (
	"context"
	"net"
	"reflect"
)

type adapter struct {
	dialer *net.Dialer
	policy func(context.Context, string, string, *net.Dialer) (net.Conn, error)
}

// WrapWithPolicy retains default-dialer identity while applying a connection
// policy. The policy owns dialing and must preserve the supplied dialer's DNS,
// address binding, context and application settings.
func WrapWithPolicy(
	dialer *net.Dialer,
	policy func(context.Context, string, string, *net.Dialer) (net.Conn, error),
) func(context.Context, string, string) (net.Conn, error) {
	return adapter{dialer: dialer, policy: policy}.dialContext
}

func (d adapter) dialContext(ctx context.Context, network, address string) (net.Conn, error) {
	if d.policy != nil {
		return d.policy(ctx, network, address, d.dialer)
	}

	return d.dialer.DialContext(ctx, network, address)
}

// IsDefault reports whether dial came from Wrap. DNS address reuse must not
// replace resolution inside an application-supplied dial function.
func IsDefault(dial func(context.Context, string, string) (net.Conn, error)) bool {
	return dial != nil && reflect.ValueOf(dial).Pointer() == reflect.ValueOf(adapter{}.dialContext).Pointer()
}
