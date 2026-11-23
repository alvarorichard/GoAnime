package desktop

import "github.com/enetx/surf/profiles"

// ConfigureTCP follows NSS's MAX_FRAGMENT_LENGTH records and ssl_SecureClose's
// close_notify alert on completed TLS connections.
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/security/nss/lib/ssl/ssl3con.c
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/security/nss/lib/ssl/sslsecur.c
func ConfigureTCP(config *profiles.TCPConfig) {
	config.DynamicRecordSizing(false).CloseNotify(true).KeepAlivePolicy(KeepAlive)
}
