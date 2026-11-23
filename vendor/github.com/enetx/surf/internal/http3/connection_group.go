package h3transport

import "github.com/enetx/http"

// dnsOnlyRace classifies the route already selected for this attempt. Cached
// Alt-Svc, explicit protocol forcing and non-discovery requests retain their
// existing timing. DNS/ECH binding resolution keeps its own privacy boundary.
func (ut *Transport) dnsOnlyRace(req *http.Request) bool {
	if ut.discovery == nil || ut.forceHTTP3 {
		return false
	}

	route, ok := req.Context().Value(alternativeRouteKey{}).(alternativeRoute)
	origin, _ := raceAddress(req)

	return ok && route.origin == origin && route.address != "" && !route.fromAltSvc
}
