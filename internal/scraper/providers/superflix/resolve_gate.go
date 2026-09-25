package superflix

import (
	"context"
	"sync"
)

// contextGate serializes access to a shared resource without making callers
// wait past their own deadline. The zero value is ready to use.
type contextGate struct {
	once  sync.Once
	token chan struct{}
}

func (g *contextGate) lock(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	g.once.Do(func() { g.token = make(chan struct{}, 1) })
	select {
	case g.token <- struct{}{}:
		// If cancellation raced with the acquisition, return the token before
		// reporting cancellation so another caller is never left waiting.
		if err := ctx.Err(); err != nil {
			<-g.token
			return err
		}
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (g *contextGate) unlock() {
	g.once.Do(func() { g.token = make(chan struct{}, 1) })
	select {
	case <-g.token:
	default:
		panic("superflix: unlock of an unlocked contextGate")
	}
}

// keyedGateGroup keeps independent stream resolutions from blocking one
// another, while coalescing concurrent requests for the same movie/episode.
type keyedGateGroup struct {
	mu    sync.Mutex
	gates map[string]*keyedGate
}

type keyedGate struct {
	gate  contextGate
	users int
}

func (g *keyedGateGroup) lock(ctx context.Context, key string) (func(), error) {
	g.mu.Lock()
	if g.gates == nil {
		g.gates = make(map[string]*keyedGate)
	}
	entry := g.gates[key]
	if entry == nil {
		entry = &keyedGate{}
		g.gates[key] = entry
	}
	entry.users++
	g.mu.Unlock()

	if err := entry.gate.lock(ctx); err != nil {
		g.releaseReference(key, entry)
		return nil, err
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			entry.gate.unlock()
			g.releaseReference(key, entry)
		})
	}, nil
}

func (g *keyedGateGroup) releaseReference(key string, entry *keyedGate) {
	g.mu.Lock()
	defer g.mu.Unlock()
	entry.users--
	if entry.users == 0 && g.gates[key] == entry {
		delete(g.gates, key)
	}
}

var streamResolveGates keyedGateGroup
