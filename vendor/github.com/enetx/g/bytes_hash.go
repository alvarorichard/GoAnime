package g

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
)

// A struct that wraps an Bytes for hashing.
type bhash struct{ bytes Bytes }

// Hash returns a bhash struct wrapping the given Bytes.
func (bs Bytes) Hash() bhash { return bhash{bs} }

// MD5 computes the MD5 hash of the wrapped Bytes and returns the hash as hex-encoded Bytes.
//
// Warning: MD5 is cryptographically broken and must not be used as a security
// primitive (e.g. for passwords, signatures, or integrity against an adversary).
// Use it only for checksums or non-security obfuscation. Prefer SHA256/SHA512.
func (bh bhash) MD5() Bytes { return bh.MD5Raw().Encode().Hex() }

// SHA1 computes the SHA1 hash of the wrapped Bytes and returns the hash as hex-encoded Bytes.
//
// Warning: SHA1 is cryptographically broken and must not be used as a security
// primitive (e.g. for signatures or integrity against an adversary). Use it only
// for checksums or non-security obfuscation. Prefer SHA256/SHA512.
func (bh bhash) SHA1() Bytes { return bh.SHA1Raw().Encode().Hex() }

// SHA256 computes the SHA256 hash of the wrapped Bytes and returns the hash as hex-encoded Bytes.
func (bh bhash) SHA256() Bytes { return bh.SHA256Raw().Encode().Hex() }

// SHA512 computes the SHA512 hash of the wrapped Bytes and returns the hash as hex-encoded Bytes.
func (bh bhash) SHA512() Bytes { return bh.SHA512Raw().Encode().Hex() }

// HMACSHA256 computes the HMAC-SHA256 of the wrapped Bytes using the provided key
// and returns the result as hex-encoded Bytes.
func (bh bhash) HMACSHA256(key Bytes) Bytes { return bh.HMACSHA256Raw(key).Encode().Hex() }

// HMACSHA512 computes the HMAC-SHA512 of the wrapped Bytes using the provided key
// and returns the result as hex-encoded Bytes.
func (bh bhash) HMACSHA512(key Bytes) Bytes { return bh.HMACSHA512Raw(key).Encode().Hex() }

// MD5Raw computes the MD5 hash of the wrapped Bytes and returns the raw digest.
//
// Warning: MD5 is cryptographically broken and must not be used as a security
// primitive. Use it only for checksums or non-security obfuscation.
func (bh bhash) MD5Raw() Bytes { return rawHasher(md5.New(), bh.bytes) }

// SHA1Raw computes the SHA1 hash of the wrapped Bytes and returns the raw digest.
//
// Warning: SHA1 is cryptographically broken and must not be used as a security
// primitive. Use it only for checksums or non-security obfuscation.
func (bh bhash) SHA1Raw() Bytes { return rawHasher(sha1.New(), bh.bytes) }

// SHA256Raw computes the SHA256 hash of the wrapped Bytes and returns the raw digest.
func (bh bhash) SHA256Raw() Bytes { return rawHasher(sha256.New(), bh.bytes) }

// SHA512Raw computes the SHA512 hash of the wrapped Bytes and returns the raw digest.
func (bh bhash) SHA512Raw() Bytes { return rawHasher(sha512.New(), bh.bytes) }

// HMACSHA256Raw computes the HMAC-SHA256 of the wrapped Bytes using the provided key
// and returns the raw digest.
func (bh bhash) HMACSHA256Raw(key Bytes) Bytes {
	return rawHasher(hmac.New(sha256.New, key), bh.bytes)
}

// HMACSHA512Raw computes the HMAC-SHA512 of the wrapped Bytes using the provided key
// and returns the raw digest.
func (bh bhash) HMACSHA512Raw(key Bytes) Bytes {
	return rawHasher(hmac.New(sha512.New, key), bh.bytes)
}

// rawHasher computes the hash of the given Bytes using the specified hash.Hash
// algorithm and returns the raw digest.
func rawHasher(h hash.Hash, bs Bytes) Bytes {
	_, _ = h.Write(bs)
	return h.Sum(nil)
}
