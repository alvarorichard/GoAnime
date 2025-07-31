package main

import (
	"fmt"
	"strconv"
)

func decodeHexToASCII(encoded string) string {
	decoded := ""

	// Split the encoded string into pairs of characters
	for i := 0; i < len(encoded); i += 2 {
		if i+1 >= len(encoded) {
			break
		}

		hexPair := encoded[i : i+2]

		// Convert hex pair to integer
		if val, err := strconv.ParseInt(hexPair, 16, 64); err == nil {
			decoded += string(rune(val))
		} else {
			decoded += hexPair // Keep as is if conversion fails
		}
	}

	return decoded
}

func main() {
	// Test with the encoded URL
	encoded := "175948514e4c4f57175b54575b5307515c050f5c0a0c0f0b0f0c0e590a0c0b5b0a0c0f0d0f0b0e0c0a5a0f590a5a0f090e0f0f0a0e0d0e5d0a010f0c0e010e0f0e0a0a5a0e010e080a5a0e000e0f0f0c0f0b0f0a0e010a5a0b0f0b5d0b0c0b0c0b0e0b010e0b0f0e0b5a0b5e0b0a0b090b0d0b080a0c0a590a0c0f0d0f0a0f0c0e0b0e0f0e5a0e0b0f0c0c5e0e0a0a0c0b5b0a0c0e000f0d0e0b0f0c0f080e0b0f0c0a0c0a590a0c0e0a0e0f0f0a0e0b0a0c0b5b0a0c0b0c0b0e0b0c0b0b0a5a0b0e0b090a5a0b0d0b0f0d0a0b0e0b0b0b5b0b0a0b0f0b5b0b0e0b0e0a000b0e0b0e0b0e0d5b0a0c0f5a"

	fmt.Printf("Encoded: %s\n", encoded)
	fmt.Printf("Length: %d\n", len(encoded))

	decoded := decodeHexToASCII(encoded)
	fmt.Printf("Decoded to ASCII: %s\n", decoded)

	// Let's also check a few specific hex pairs to see what they become
	fmt.Printf("\nFirst few hex pairs decoded:\n")
	for i := 0; i < 20 && i+1 < len(encoded); i += 2 {
		hexPair := encoded[i : i+2]
		if val, err := strconv.ParseInt(hexPair, 16, 64); err == nil {
			fmt.Printf("  %s -> %c (ASCII %d)\n", hexPair, rune(val), val)
		}
	}
}
