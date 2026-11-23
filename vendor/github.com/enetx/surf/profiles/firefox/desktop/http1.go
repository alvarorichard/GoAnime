package desktop

import "github.com/enetx/surf/profiles"

// ConfigureH1 follows the buffered header/body transaction stream and the
// network.buffer.cache.size preference consumed by nsIOService.
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/protocol/http/nsHttpTransaction.cpp
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/modules/libpref/init/all.js
func ConfigureH1(config *profiles.HTTP1Config) {
	config.WriteBufferSize(32768).BodyReadSize(32768)
}
