package updater

// VerifyTrusted reports whether sig is a minisign signature of message by one of
// the keys embedded in this build, the same ones that vouch for GoAnime's own
// updates. Other components that fetch executables (the optional cfsolve tool)
// use it so they are never trusted less than the updater is.
func VerifyTrusted(message, sig []byte) error {
	return verifyMinisign(message, sig, trustedPublicKeys)
}
