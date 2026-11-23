package protocol

// CanSwitchVersion reports first-flight compatibility between distinct
// implemented versions. Never infer compatibility with a future version.
// https://www.rfc-editor.org/rfc/rfc9369.html#section-4.1
func CanSwitchVersion(from, to Version) bool {
	return from == Version1 && to == Version2 || from == Version2 && to == Version1
}
