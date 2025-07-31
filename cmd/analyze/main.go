package main

import (
	"fmt"
)

func main() {
	// Let's analyze the encoded string character by character
	encoded := "175948514e4c4f57175b54575b5307515c050f5c0a0c0f0b0f0c0e590a0c0b5b0a0c0f0d0f0b0e0c0a5a0f590a5a0f090e0f0f0a0e0d0e5d0a010f0c0e010e0f0e0a0a5a0e010e080a5a0e000e0f0f0c0f0b0f0a0e010a5a0b0f0b5d0b0c0b0c0b0e0b010e0b0f0e0b5a0b5e0b0a0b090b0d0b080a0c0a590a0c0f0d0f0a0f0c0e0b0e0f0e5a0e0b0f0c0c5e0e0a0a0c0b5b0a0c0e000f0d0e0b0f0c0f080e0b0f0c0a0c0a590a0c0e0a0e0f0f0a0e0b0a0c0b5b0a0c0b0c0b0e0b0c0b0b0a5a0b0e0b090a5a0b0d0b0f0d0a0b0e0b0b0b5b0b0a0b0f0b5b0b0e0b0e0a000b0e0b0e0b0e0d5b0a0c0f5a"

	fmt.Printf("Analyzing encoded string length: %d\n", len(encoded))

	// Let's see what unique hex pairs we have
	uniquePairs := make(map[string]bool)
	for i := 0; i < len(encoded); i += 2 {
		if i+1 < len(encoded) {
			pair := encoded[i : i+2]
			uniquePairs[pair] = true
		}
	}

	fmt.Printf("Unique hex pairs found:\n")
	for pair := range uniquePairs {
		fmt.Printf("  %s\n", pair)
	}

	// Let's try a different approach - what if this isn't standard hex encoding?
	// Let's see if this could be some other encoding scheme

	// First few characters: "17 59 48 51 4e 4c 4f 57"
	// Let's check if these look like valid hex values
	fmt.Printf("\nFirst 16 characters as hex pairs:\n")
	for i := 0; i < 16 && i+1 < len(encoded); i += 2 {
		pair := encoded[i : i+2]
		fmt.Printf("  %s\n", pair)
	}
}
