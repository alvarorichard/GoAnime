// Package goanime exposes public client APIs and shared sentinel errors.
package goanime

import "github.com/alvarorichard/Goanime/internal/scraper/netx"

// ErrSourceUnavailable indicates that the selected upstream source is
// temporarily unavailable or blocked.
var ErrSourceUnavailable = netx.ErrSourceUnavailable
