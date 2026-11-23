package encoder

func fillSlice[T any](s []T, v T) {
	if len(s) == 0 {
		return
	}
	s[0] = v
	for i := 1; i < len(s); i *= 2 {
		copy(s[i:], s[:i])
	}
}
