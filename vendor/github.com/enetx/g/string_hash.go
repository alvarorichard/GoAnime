package g

// A struct that wraps an String for hashing.
type shash struct{ str String }

// Hash returns a shash struct wrapping the given String.
func (s String) Hash() shash { return shash{s} }

// MD5 computes the MD5 hash of the wrapped String and returns the hash as hex-encoded String.
//
// WARNING: MD5 is cryptographically broken and is NOT a security primitive.
// Do not use it for passwords, signatures, or integrity against an adversary.
// Use it only for checksums or legacy interop.
func (sh shash) MD5() String { return sh.str.BytesUnsafe().Hash().MD5().StringUnsafe() }

// SHA1 computes the SHA1 hash of the wrapped String and returns the hash as hex-encoded String.
//
// WARNING: SHA1 is cryptographically broken (collision attacks are practical) and
// is NOT a security primitive. Do not use it for signatures or integrity against
// an adversary. Use SHA256/SHA512 instead.
func (sh shash) SHA1() String { return sh.str.BytesUnsafe().Hash().SHA1().StringUnsafe() }

// SHA256 computes the SHA256 hash of the wrapped String and returns the hash as hex-encoded String.
func (sh shash) SHA256() String { return sh.str.BytesUnsafe().Hash().SHA256().StringUnsafe() }

// SHA512 computes the SHA512 hash of the wrapped String and returns the hash as hex-encoded String.
func (sh shash) SHA512() String { return sh.str.BytesUnsafe().Hash().SHA512().StringUnsafe() }

// HMACSHA256 computes the HMAC-SHA256 of the wrapped String using the provided key
// and returns the result as hex-encoded String.
func (sh shash) HMACSHA256(key String) String {
	return sh.str.BytesUnsafe().Hash().HMACSHA256(key.BytesUnsafe()).StringUnsafe()
}

// HMACSHA512 computes the HMAC-SHA512 of the wrapped String using the provided key
// and returns the result as hex-encoded String.
func (sh shash) HMACSHA512(key String) String {
	return sh.str.BytesUnsafe().Hash().HMACSHA512(key.BytesUnsafe()).StringUnsafe()
}

// MD5Raw computes the MD5 hash of the wrapped String and returns the raw digest as Bytes.
//
// WARNING: MD5 is cryptographically broken and is NOT a security primitive.
// Use it only for checksums or legacy interop, never for security.
func (sh shash) MD5Raw() Bytes { return sh.str.BytesUnsafe().Hash().MD5Raw() }

// SHA1Raw computes the SHA1 hash of the wrapped String and returns the raw digest as Bytes.
//
// WARNING: SHA1 is cryptographically broken and is NOT a security primitive.
// Use SHA256/SHA512 instead for any security-sensitive purpose.
func (sh shash) SHA1Raw() Bytes { return sh.str.BytesUnsafe().Hash().SHA1Raw() }

// SHA256Raw computes the SHA256 hash of the wrapped String and returns the raw digest as Bytes.
func (sh shash) SHA256Raw() Bytes { return sh.str.BytesUnsafe().Hash().SHA256Raw() }

// SHA512Raw computes the SHA512 hash of the wrapped String and returns the raw digest as Bytes.
func (sh shash) SHA512Raw() Bytes { return sh.str.BytesUnsafe().Hash().SHA512Raw() }

// HMACSHA256Raw computes the HMAC-SHA256 of the wrapped String using the provided key
// and returns the raw digest as Bytes.
func (sh shash) HMACSHA256Raw(key String) Bytes {
	return sh.str.BytesUnsafe().Hash().HMACSHA256Raw(key.BytesUnsafe())
}

// HMACSHA512Raw computes the HMAC-SHA512 of the wrapped String using the provided key
// and returns the raw digest as Bytes.
func (sh shash) HMACSHA512Raw(key String) Bytes {
	return sh.str.BytesUnsafe().Hash().HMACSHA512Raw(key.BytesUnsafe())
}
