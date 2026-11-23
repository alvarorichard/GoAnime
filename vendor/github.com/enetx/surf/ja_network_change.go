package surf

import "github.com/enetx/http"

// Snapshot existing proxy children without creating a route or holding proxyMu
// across socket operations. Shared native HTTP/2 pools appear only once.
func (rt *roundtripper) networkTransports() []*http.Transport {
	seen := make(map[*http.Transport]struct{})
	var result []*http.Transport
	var visit func(*roundtripper)
	visit = func(current *roundtripper) {
		for _, transport := range []*http.Transport{current.http1tr, current.http1trFallback} {
			if transport == nil {
				continue
			}
			if _, ok := seen[transport]; !ok {
				seen[transport] = struct{}{}
				result = append(result, transport)
			}
		}

		current.proxyMu.Lock()
		children := make([]*roundtripper, 0, len(current.proxyTransports))
		for _, child := range current.proxyTransports {
			children = append(children, child)
		}
		current.proxyMu.Unlock()
		for _, child := range children {
			visit(child)
		}
	}
	visit(rt)

	return result
}
