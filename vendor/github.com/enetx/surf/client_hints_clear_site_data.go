package surf

import (
	"strings"

	"github.com/enetx/http"
)

// Chromium tokenizes exact quoted comma members. The parsed-response gate
// suppresses Accept-CH/Critical-CH independently of clearing's ASCII/trust gates.
// Storage alone does not clear hints; cookies/cache also invalidate opt-ins.
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/url_request/clear_site_data.cc
// https://github.com/chromium/chromium/blob/158.0.8127.18/content/browser/browsing_data/clear_site_data_utils.cc#L176
func clearsClientHints(value string) bool {
	for _, member := range strings.Split(value, ",") {
		switch strings.Trim(member, " \t\r\n\f\v") {
		case `"cookies"`, `"cache"`, `"clientHints"`, `"*"`:
			return true
		}
	}

	return false
}

func (t *clientHintsTransport) observeClearSiteData(req *http.Request, resp *http.Response) bool {
	value := joinedClientHint(resp.Header, "clear-site-data")
	if !clearsClientHints(value) {
		return false
	}

	if !t.clearSiteData || !clientHintTrustworthy(req.URL) {
		return true
	}

	for i := range len(value) {
		if value[i] >= 128 {
			return true
		}
	}

	origin, _ := clientHintOrigin(req.URL)
	t.cache.put(origin, nil)

	return true
}
