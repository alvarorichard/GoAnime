package main

import (
	"fmt"
)

func decodeSourceURLFromAniCli(encoded string) string {
	// Direct implementation of the ani-cli sed script logic
	decoded := ""

	// Split the encoded string into pairs and apply the exact mappings from ani-cli
	for i := 0; i < len(encoded); i += 2 {
		if i+1 >= len(encoded) {
			break
		}

		hexPair := encoded[i : i+2]

		// Exact mappings from ani-cli sed script
		switch hexPair {
		case "79":
			decoded += "A"
		case "7a":
			decoded += "B"
		case "7b":
			decoded += "C"
		case "7c":
			decoded += "D"
		case "7d":
			decoded += "E"
		case "7e":
			decoded += "F"
		case "7f":
			decoded += "G"
		case "70":
			decoded += "H"
		case "71":
			decoded += "I"
		case "72":
			decoded += "J"
		case "73":
			decoded += "K"
		case "74":
			decoded += "L"
		case "75":
			decoded += "M"
		case "76":
			decoded += "N"
		case "77":
			decoded += "O"
		case "68":
			decoded += "P"
		case "69":
			decoded += "Q"
		case "6a":
			decoded += "R"
		case "6b":
			decoded += "S"
		case "6c":
			decoded += "T"
		case "6d":
			decoded += "U"
		case "6e":
			decoded += "V"
		case "6f":
			decoded += "W"
		case "60":
			decoded += "X"
		case "61":
			decoded += "Y"
		case "62":
			decoded += "Z"
		case "59":
			decoded += "a"
		case "5a":
			decoded += "b"
		case "5b":
			decoded += "c"
		case "5c":
			decoded += "d"
		case "5d":
			decoded += "e"
		case "5e":
			decoded += "f"
		case "5f":
			decoded += "g"
		case "50":
			decoded += "h"
		case "51":
			decoded += "i"
		case "52":
			decoded += "j"
		case "53":
			decoded += "k"
		case "54":
			decoded += "l"
		case "55":
			decoded += "m"
		case "56":
			decoded += "n"
		case "57":
			decoded += "o"
		case "48":
			decoded += "p"
		case "49":
			decoded += "q"
		case "4a":
			decoded += "r"
		case "4b":
			decoded += "s"
		case "4c":
			decoded += "t"
		case "4d":
			decoded += "u"
		case "4e":
			decoded += "v"
		case "4f":
			decoded += "w"
		case "40":
			decoded += "x"
		case "41":
			decoded += "y"
		case "42":
			decoded += "z"
		case "08":
			decoded += "0"
		case "09":
			decoded += "1"
		case "0a":
			decoded += "2"
		case "0b":
			decoded += "3"
		case "0c":
			decoded += "4"
		case "0d":
			decoded += "5"
		case "0e":
			decoded += "6"
		case "0f":
			decoded += "7"
		case "00":
			decoded += "8"
		case "01":
			decoded += "9"
		case "15":
			decoded += "-"
		case "16":
			decoded += "."
		case "67":
			decoded += "_"
		case "46":
			decoded += "~"
		case "02":
			decoded += ":"
		case "17":
			decoded += "/"
		case "07":
			decoded += "?"
		case "1b":
			decoded += "#"
		case "63":
			decoded += "["
		case "65":
			decoded += "]"
		case "78":
			decoded += "@"
		case "19":
			decoded += "!"
		case "1c":
			decoded += "$"
		case "1e":
			decoded += "&"
		case "10":
			decoded += "("
		case "11":
			decoded += ")"
		case "12":
			decoded += "*"
		case "13":
			decoded += "+"
		case "14":
			decoded += ","
		case "03":
			decoded += ";"
		case "05":
			decoded += "="
		case "1d":
			decoded += "%"
		default:
			// For unknown hex pairs, keep the original hex
			decoded += hexPair
		}
	}

	// Apply the final transformation
	if decoded != "" {
		decoded = decoded + "/clock.json"
	}

	return decoded
}

func main() {
	// Test with the encoded URL
	encoded := "175948514e4c4f57175b54575b5307515c050f5c0a0c0f0b0f0c0e590a0c0b5b0a0c0f0d0f0b0e0c0a5a0f590a5a0f090e0f0f0a0e0d0e5d0a010f0c0e010e0f0e0a0a5a0e010e080a5a0e000e0f0f0c0f0b0f0a0e010a5a0b0f0b5d0b0c0b0c0b0e0b010e0b0f0e0b5a0b5e0b0a0b090b0d0b080a0c0a590a0c0f0d0f0a0f0c0e0b0e0f0e5a0e0b0f0c0c5e0e0a2a0c0b5b0a0c0e000f0d0e0b0f0c0f080e0b0f0c0a0c0a590a0c0e0a0e0f0f0a0e0b0a0c0b5b0a0c0b0c0b0e0b0c0b0b0a5a0b0e0b090a5a0b0d0b0f0d0a0b0e0b0b0b5b0b0a0b0f0b5b0b0e0b0e0a000b0e0b0e0b0e0d5b0a0c0f5a"

	fmt.Printf("Encoded: %s\n", encoded)
	fmt.Printf("Length: %d\n", len(encoded))

	decoded := decodeSourceURLFromAniCli(encoded)
	fmt.Printf("Decoded: %s\n", decoded)

	// Let's also try to simulate what ani-cli does exactly - first convert every 2 chars
	fmt.Printf("\nFirst few pairs decoded step by step:\n")
	for i := 0; i < 20 && i+1 < len(encoded); i += 2 {
		hexPair := encoded[i : i+2]
		var result string
		switch hexPair {
		case "17":
			result = "/"
		case "59":
			result = "a"
		case "48":
			result = "p"
		case "51":
			result = "i"
		case "4e":
			result = "v"
		case "4c":
			result = "t"
		case "4f":
			result = "w"
		case "57":
			result = "o"
		default:
			result = hexPair
		}
		fmt.Printf("  %s -> %s\n", hexPair, result)
	}
}
