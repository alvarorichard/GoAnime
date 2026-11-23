package surf

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/enetx/http"
	"github.com/enetx/http/http2"
	h3transport "github.com/enetx/surf/internal/http3"
	"github.com/enetx/surf/internal/netchange"
	"github.com/enetx/surf/profiles"
)

type networkState struct {
	config    profiles.NetworkChangeConfig
	mu        sync.Mutex
	closed    bool
	policy    NetworkChangePolicy
	transport http.RoundTripper
	ctx       context.Context
	cancel    context.CancelFunc
	liveness  context.CancelFunc
	sub       *netchange.Subscription
}

func (b *Builder) applyNetworkChanges() error {
	config := profiles.NetworkChangeConfig{}
	if b.configureNetworkChange != nil {
		b.configureNetworkChange(&config)
	}

	if b.networkSettings != nil {
		for _, override := range b.networkSettings.overrides {
			override(&config)
		}
	}

	if config.Delay < 0 || config.Delay > time.Minute {
		return fmt.Errorf("network coalescing delay must be between zero and one minute")
	}

	if config.OnChange == nil {
		b.cli.network = nil
		return nil
	}

	configureNetworkActivity(b.cli.cli.Transport)
	ctx, cancel := context.WithCancel(context.Background())
	state := &networkState{
		config:    config,
		policy:    config.OnChange,
		transport: b.cli.cli.Transport,
		ctx:       ctx,
		cancel:    cancel,
	}

	if config.Enabled {
		sub, err := netchange.Subscribe(config.Delay, func(event netchange.Event) {
			state.change(NetworkChangeInfo{
				AddressChanged: event.Address, LinkChanged: event.Link, RouteChanged: event.Route,
				IdentityChanged:   event.Identity,
				TemporaryIPv6Only: event.TemporaryIPv6Only, IPv6Only: event.IPv6Only,
			})
		}, config.ReadIdentity)
		if err != nil {
			cancel()
			return fmt.Errorf("subscribe to network changes: %w", err)
		}

		state.sub = sub
	}

	b.cli.network = state

	return nil
}

// NetworkChanged immediately applies the built network policy. With no argument
// it describes a stable address/link/route change. Applications that already own
// notifications can disable Automatic and call this hook. Native notifications
// use the configured coalescing delay; explicit calls are already-delivered events.
func (c *Client) NetworkChanged(changes ...NetworkChangeInfo) {
	if c.network == nil {
		return
	}

	if len(changes) == 0 {
		c.network.change(NetworkChangeInfo{AddressChanged: true, LinkChanged: true, RouteChanged: true})
		return
	}

	for _, change := range changes {
		c.network.change(change)
	}
}

func (n *networkState) change(info NetworkChangeInfo) {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.closed {
		return
	}

	action := n.policy(info)
	ctx := n.ctx
	if action.HTTP1LivenessTimeout > 0 {
		if n.liveness != nil {
			n.liveness()
		}
		ctx, n.liveness = context.WithCancel(n.ctx)
	}

	applyNetworkChange(n.transport, ctx, action)
}

func (n *networkState) close() {
	n.mu.Lock()
	n.closed = true
	n.cancel()
	n.mu.Unlock()

	if n.sub != nil {
		n.sub.Close()
	}
}

func applyNetworkChange(transport http.RoundTripper, ctx context.Context, action NetworkChangeAction) {
	switch tr := transport.(type) {
	case *clientHintsTransport:
		applyNetworkChange(tr.Base, ctx, action)
	case *h3transport.Transport:
		if action.ClearFailedAlternatives {
			tr.ClearFailedAlternatives()
		}
		if action.RetireHTTP3 {
			tr.RetireConnections()
		}
		applyNetworkChange(tr.Fallback, ctx, action)
	case *roundtripper:
		for _, tcp := range tr.networkTransports() {
			applyNetworkChange(tcp, ctx, action)
		}
	case *http.Transport:
		maintainHTTP2(tr.HTTP2Connections(), ctx, action)
		if action.RetireHTTP1 {
			tr.RetireHTTP1Connections()
		}
		if action.HTTP1LivenessTimeout > 0 {
			tr.CloseInactiveHTTP1Connections(ctx, action.HTTP1LivenessTimeout)
		}
	case *http2.Transport:
		maintainHTTP2(tr.HTTP2Connections(), ctx, action)
	}
}

func maintainHTTP2(connections []*http.HTTP2Connection, ctx context.Context, action NetworkChangeAction) {
	for _, conn := range connections {
		switch {
		case action.CloseHTTP2:
			conn.Close()
		case action.RetireHTTP2:
			options := http.HTTP2RetirementOptions{GoAway: action.HTTP2GoAwayOnClose}
			if action.HTTP2IdleProbeAfter > 0 && action.HTTP2ProbeTimeout > 0 &&
				conn.IdleTime() >= action.HTTP2IdleProbeAfter {
				probe, cancel := context.WithTimeout(ctx, action.HTTP2ProbeTimeout)
				options.ProbeContext = probe
				options.ProbePayload = action.HTTP2ProbePayload
				conn.Retire(options)
				context.AfterFunc(probe, cancel)
			} else {
				conn.Retire(options)
			}
		}
	}
}

// Configure before the first request; proxy children inherit this through Clone.
func configureNetworkActivity(transport http.RoundTripper) {
	switch tr := transport.(type) {
	case *clientHintsTransport:
		configureNetworkActivity(tr.Base)
	case *h3transport.Transport:
		configureNetworkActivity(tr.Fallback)
	case *roundtripper:
		for _, tcp := range tr.networkTransports() {
			tcp.TrackConnectionActivity = true
		}
	case *http.Transport:
		tr.TrackConnectionActivity = true
	}
}
