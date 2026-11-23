package g

import "os"

const (
	// ASCIILetters is the set of all ASCII letters (lowercase + uppercase).
	ASCIILetters String = ASCIILowercase + ASCIIUppercase
	// ASCIILowercase is the set of lowercase ASCII letters.
	ASCIILowercase String = "abcdefghijklmnopqrstuvwxyz"
	// ASCIIUppercase is the set of uppercase ASCII letters.
	ASCIIUppercase String = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	// Digits is the set of decimal digit characters.
	Digits String = "0123456789"
	// HexDigits is the set of hexadecimal digit characters (both cases).
	HexDigits String = "0123456789abcdefABCDEF"
	// OctDigits is the set of octal digit characters.
	OctDigits String = "01234567"
	// Punctuation is the set of ASCII punctuation characters.
	Punctuation String = `!"#$%&'()*+,-./:;<=>?@[\]^{|}~` + "`"

	// FileDefault is the default permission mode (0o644) used when writing files.
	FileDefault os.FileMode = 0o644
	// FileCreate is the permission mode (0o666) used when creating files.
	FileCreate os.FileMode = 0o666
	// DirDefault is the default permission mode (0o755) used when creating directories.
	DirDefault os.FileMode = 0o755
	// FullAccess is the permission mode (0o777) granting read, write and execute to everyone.
	FullAccess os.FileMode = 0o777

	// PathSeparator is the OS-specific path separator as a String.
	PathSeparator = String(os.PathSeparator)
)
