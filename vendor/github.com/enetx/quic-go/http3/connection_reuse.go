package http3

import (
	"context"
	"net"
	"slices"

	"github.com/enetx/http"
	"github.com/enetx/quic-go"
)

// HasCachedConnection reports whether address has a confirmed reusable session
// in this context's exact pool partition. It does not dial, resolve aliases,
// reserve a request or invoke the cross-origin reuse policy. The result is a
// snapshot: normal request admission still handles concurrent closure/GOAWAY.
func (t *Transport) HasCachedConnection(ctx context.Context, address string) bool {
	if ctx.Err() != nil {
		return false
	}
	if _, _, err := net.SplitHostPort(address); err != nil {
		return false
	}

	key, _ := t.clientPoolKey(ctx, address)
	t.mutex.Lock()
	defer t.mutex.Unlock()
	client := t.clients[key]
	return !t.closed && client != nil && reusableClient(client)
}

// ConnectionReuseInfo describes a candidate owned by this transport's pool.
// Authority and OriginalAuthority are canonical host:port pairs. Conn remains
// transport-owned and must not be closed or have its accept loops taken over.
type ConnectionReuseInfo struct {
	// Authority is the requested canonical host:port.
	Authority string
	// OriginalAuthority is the host:port used to establish this pooled connection.
	OriginalAuthority string
	// Conn is the transport-owned candidate connection.
	Conn *quic.Conn
	// AllowNetwork is false during the first pass over all candidates. Policies
	// should defer DNS/network lookups until the second pass, where it is true.
	// This lets explicit existing authorizations win before route resolution.
	AllowNetwork bool
}

func nativeClient(client clientConn) *ClientConn {
	switch c := client.(type) {
	case *ClientConn:
		return c
	case *earlyRestartClient:
		return c.current.Load()
	default:
		return nil
	}
}

func hasActiveClientStreams(cl *roundTripperWithCount) bool {
	// Dial publishes clientConn by closing dialing; idle closure can run while
	// the last waiting request has already cancelled but that dial is pending.
	select {
	case <-cl.dialing:
	default:
		return false
	}
	c := nativeClient(cl.clientConn)
	return c != nil && c.rawConn.hasActiveStreams()
}

func reusableClient(cl *roundTripperWithCount) bool {
	select {
	case <-cl.dialing:
	default:
		return false
	}
	if cl.dialErr != nil || cl.conn == nil || cl.conn.Context().Err() != nil {
		return false
	}
	select {
	case <-cl.conn.HandshakeComplete():
	default:
		return false
	}
	c := nativeClient(cl.clientConn)
	return c != nil && c.goAwayCtx.Err() == nil && c.rawConn.context().Err() == nil
}

// Candidates are pinned while callbacks run, so CloseIdleConnections cannot
// close a candidate between authorization and acquisition. Cross-origin uses
// never create extra pool owners: closing, accounting and QPACK stay singular.
func (t *Transport) getClient(ctx context.Context, hostname string, onlyCached bool) (*roundTripperWithCount, bool, error) {
	if t.ConnectionReusePolicy == nil {
		return t.getClientForHost(ctx, hostname, onlyCached)
	}
	key, partition := t.clientPoolKey(ctx, hostname)
	t.mutex.Lock()
	if t.closed || t.clients[key] != nil {
		t.mutex.Unlock()
		return t.getClientForHost(ctx, hostname, onlyCached)
	}
	keys := make([]string, 0, len(t.clients))
	for key := range t.clients {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	type candidate struct {
		key    string
		client *roundTripperWithCount
	}
	candidates := make([]candidate, 0, len(keys))
	for _, key := range keys {
		if cl := t.clients[key]; cl.partition == partition && reusableClient(cl) {
			cl.useCount.Add(1)
			candidates = append(candidates, candidate{key, cl})
		}
	}
	t.mutex.Unlock()
	var selected *roundTripperWithCount
	defer func() {
		for _, candidate := range candidates {
			if candidate.client != selected {
				candidate.client.useCount.Add(-1)
			}
		}
	}()
	for _, allowNetwork := range []bool{false, true} {
		for _, candidate := range candidates {
			if ctx.Err() != nil {
				return nil, false, context.Cause(ctx)
			}
			cl := candidate.client
			if cl.conn.Context().Err() != nil {
				continue
			}
			info := ConnectionReuseInfo{Authority: hostname, OriginalAuthority: cl.authority, Conn: cl.conn, AllowNetwork: allowNetwork}
			if !t.ConnectionReusePolicy(ctx, info) {
				continue
			}
			t.mutex.Lock()
			available := !t.closed && t.clients[key] == nil && t.clients[candidate.key] == cl && reusableClient(cl)
			if available && ctx.Err() == nil {
				selected = cl
			}
			t.mutex.Unlock()
			if selected != nil {
				return selected, true, nil
			}
		}
	}
	if ctx.Err() != nil {
		return nil, false, context.Cause(ctx)
	}
	return t.getClientForHost(ctx, hostname, onlyCached)
}

func (t *Transport) clientPoolKey(ctx context.Context, authority string) (key, partition string) {
	if t.ConnectionPoolPartition == nil {
		return authority, ""
	}
	partition = t.ConnectionPoolPartition(ctx, authority)
	return authority + "\x00" + partition, partition
}

// RoundTripCached uses an existing connection for the initial attempt. An exact
// ErrNoCachedConn leaves Request.Body open and caller-owned; no request was sent.
// All other outcomes transfer body ownership as in RoundTrip. Once a cached
// attempt starts, normal retry rules may establish a dedicated connection.
func (t *Transport) RoundTripCached(req *http.Request) (*http.Response, error) {
	req = t.requestWithPreparationBody(req)
	response, err := t.roundTripOpt(req, RoundTripOpt{OnlyCachedConn: true, allowRetryDial: true})
	if err != nil && err != ErrNoCachedConn && req.Body != nil {
		req.Body.Close()
	}
	return response, err
}

// RoundTripCachedReady uses only a confirmed reusable connection for the first
// attempt. A pending dial, early-only readiness or GOAWAY is an ErrNoCachedConn
// miss without waiting or taking body ownership. Normal retry rules can dial
// after a confirmed initial attempt has started. RoundTripCached retains its
// existing behavior of joining pending cached attempts.
func (t *Transport) RoundTripCachedReady(req *http.Request) (*http.Response, error) {
	req = t.requestWithPreparationBody(req)
	response, err := t.roundTripOpt(req, RoundTripOpt{
		OnlyCachedConn:   true,
		allowRetryDial:   true,
		requireConfirmed: true,
	})
	if err != nil && err != ErrNoCachedConn && req.Body != nil {
		req.Body.Close()
	}
	return response, err
}
