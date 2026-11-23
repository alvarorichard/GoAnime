package http3

import "context"

type connectionStartedKey struct{}

// NotifyConnectionStarted lets a custom Transport.Dial report that DNS and
// socket preparation are complete and a QUIC handshake is starting. Call before
// entering the QUIC dial. Repeated calls are safe; contexts without a transport
// observer are unchanged. Notifications belong to the actual shared pool dial,
// so all PreconnectWithOptions waiters observe the same connection attempt.
func NotifyConnectionStarted(ctx context.Context) {
	if notify, _ := ctx.Value(connectionStartedKey{}).(func()); notify != nil {
		notify()
	}
}
