package main

import (
	"fmt"
	"strings"
)

func decodeSourceURL(encoded string) string {
	// Handle the case where the encoded string might contain a colon and port
	parts := strings.Split(encoded, ":")
	mainPart := parts[0]
	port := ""
	if len(parts) > 1 {
		port = ":" + parts[1]
	}

	// This implements the complex hex decoding logic from ani-cli
	decoded := strings.Builder{}

	// Split the encoded string into pairs of characters
	for i := 0; i < len(mainPart); i += 2 {
		if i+1 >= len(mainPart) {
			break
		}

		hexPair := mainPart[i : i+2]

		// Map hex pairs to characters (from ani-cli sed script)
		switch hexPair {
		case "79":
			decoded.WriteString("A")
		case "7a":
			decoded.WriteString("B")
		case "7b":
			decoded.WriteString("C")
		case "7c":
			decoded.WriteString("D")
		case "7d":
			decoded.WriteString("E")
		case "7e":
			decoded.WriteString("F")
		case "7f":
			decoded.WriteString("G")
		case "80":
			decoded.WriteString("H")
		case "81":
			decoded.WriteString("I")
		case "82":
			decoded.WriteString("J")
		case "83":
			decoded.WriteString("K")
		case "84":
			decoded.WriteString("L")
		case "85":
			decoded.WriteString("M")
		case "86":
			decoded.WriteString("N")
		case "87":
			decoded.WriteString("O")
		case "88":
			decoded.WriteString("P")
		case "89":
			decoded.WriteString("Q")
		case "8a":
			decoded.WriteString("R")
		case "8b":
			decoded.WriteString("S")
		case "8c":
			decoded.WriteString("T")
		case "8d":
			decoded.WriteString("U")
		case "8e":
			decoded.WriteString("V")
		case "8f":
			decoded.WriteString("W")
		case "90":
			decoded.WriteString("X")
		case "91":
			decoded.WriteString("Y")
		case "92":
			decoded.WriteString("Z")
		case "93":
			decoded.WriteString("a")
		case "94":
			decoded.WriteString("b")
		case "95":
			decoded.WriteString("c")
		case "96":
			decoded.WriteString("d")
		case "97":
			decoded.WriteString("e")
		case "98":
			decoded.WriteString("f")
		case "99":
			decoded.WriteString("g")
		case "9a":
			decoded.WriteString("h")
		case "9b":
			decoded.WriteString("i")
		case "9c":
			decoded.WriteString("j")
		case "9d":
			decoded.WriteString("k")
		case "9e":
			decoded.WriteString("l")
		case "9f":
			decoded.WriteString("m")
		case "a0":
			decoded.WriteString("n")
		case "a1":
			decoded.WriteString("o")
		case "a2":
			decoded.WriteString("p")
		case "a3":
			decoded.WriteString("q")
		case "a4":
			decoded.WriteString("r")
		case "a5":
			decoded.WriteString("s")
		case "a6":
			decoded.WriteString("t")
		case "a7":
			decoded.WriteString("u")
		case "a8":
			decoded.WriteString("v")
		case "a9":
			decoded.WriteString("w")
		case "aa":
			decoded.WriteString("x")
		case "ab":
			decoded.WriteString("y")
		case "ac":
			decoded.WriteString("z")
		case "ad":
			decoded.WriteString("0")
		case "ae":
			decoded.WriteString("1")
		case "af":
			decoded.WriteString("2")
		case "b0":
			decoded.WriteString("3")
		case "b1":
			decoded.WriteString("4")
		case "b2":
			decoded.WriteString("5")
		case "b3":
			decoded.WriteString("6")
		case "b4":
			decoded.WriteString("7")
		case "b5":
			decoded.WriteString("8")
		case "b6":
			decoded.WriteString("9")
		case "b7":
			decoded.WriteString("+")
		case "b8":
			decoded.WriteString("/")
		case "b9":
			decoded.WriteString("=")
		case "0a":
			decoded.WriteString(".")
		case "0b":
			decoded.WriteString("-")
		case "0c":
			decoded.WriteString("_")
		case "0d":
			decoded.WriteString("~")
		case "0e":
			decoded.WriteString("!")
		case "0f":
			decoded.WriteString("*")
		case "10":
			decoded.WriteString("'")
		case "11":
			decoded.WriteString("(")
		case "12":
			decoded.WriteString(")")
		case "13":
			decoded.WriteString(";")
		case "14":
			decoded.WriteString(":")
		case "15":
			decoded.WriteString("@")
		case "16":
			decoded.WriteString("&")
		case "17":
			decoded.WriteString("=")
		case "18":
			decoded.WriteString("+")
		case "19":
			decoded.WriteString("$")
		case "1a":
			decoded.WriteString(",")
		case "1b":
			decoded.WriteString("/")
		case "1c":
			decoded.WriteString("?")
		case "1d":
			decoded.WriteString("#")
		case "1e":
			decoded.WriteString("[")
		case "1f":
			decoded.WriteString("]")
		case "20":
			decoded.WriteString(" ")
		case "21":
			decoded.WriteString("!")
		case "22":
			decoded.WriteString("\"")
		case "23":
			decoded.WriteString("#")
		case "24":
			decoded.WriteString("$")
		case "25":
			decoded.WriteString("%")
		case "26":
			decoded.WriteString("&")
		case "27":
			decoded.WriteString("'")
		case "28":
			decoded.WriteString("(")
		case "29":
			decoded.WriteString(")")
		case "2a":
			decoded.WriteString("*")
		case "2b":
			decoded.WriteString("+")
		case "2c":
			decoded.WriteString(",")
		case "2d":
			decoded.WriteString("-")
		case "2e":
			decoded.WriteString(".")
		case "2f":
			decoded.WriteString("/")
		case "30":
			decoded.WriteString("0")
		case "31":
			decoded.WriteString("1")
		case "32":
			decoded.WriteString("2")
		case "33":
			decoded.WriteString("3")
		case "34":
			decoded.WriteString("4")
		case "35":
			decoded.WriteString("5")
		case "36":
			decoded.WriteString("6")
		case "37":
			decoded.WriteString("7")
		case "38":
			decoded.WriteString("8")
		case "39":
			decoded.WriteString("9")
		case "3a":
			decoded.WriteString(":")
		case "3b":
			decoded.WriteString(";")
		case "3c":
			decoded.WriteString("<")
		case "3d":
			decoded.WriteString("=")
		case "3e":
			decoded.WriteString(">")
		case "3f":
			decoded.WriteString("?")
		case "40":
			decoded.WriteString("@")
		case "41":
			decoded.WriteString("A")
		case "42":
			decoded.WriteString("B")
		case "43":
			decoded.WriteString("C")
		case "44":
			decoded.WriteString("D")
		case "45":
			decoded.WriteString("E")
		case "46":
			decoded.WriteString("F")
		case "47":
			decoded.WriteString("G")
		case "48":
			decoded.WriteString("H")
		case "49":
			decoded.WriteString("I")
		case "4a":
			decoded.WriteString("J")
		case "4b":
			decoded.WriteString("K")
		case "4c":
			decoded.WriteString("L")
		case "4d":
			decoded.WriteString("M")
		case "4e":
			decoded.WriteString("N")
		case "4f":
			decoded.WriteString("O")
		case "50":
			decoded.WriteString("P")
		case "51":
			decoded.WriteString("Q")
		case "52":
			decoded.WriteString("R")
		case "53":
			decoded.WriteString("S")
		case "54":
			decoded.WriteString("T")
		case "55":
			decoded.WriteString("U")
		case "56":
			decoded.WriteString("V")
		case "57":
			decoded.WriteString("W")
		case "58":
			decoded.WriteString("X")
		case "59":
			decoded.WriteString("Y")
		case "5a":
			decoded.WriteString("Z")
		case "5b":
			decoded.WriteString("[")
		case "5c":
			decoded.WriteString("\\")
		case "5d":
			decoded.WriteString("]")
		case "5e":
			decoded.WriteString("^")
		case "5f":
			decoded.WriteString("_")
		case "60":
			decoded.WriteString("`")
		case "61":
			decoded.WriteString("a")
		case "62":
			decoded.WriteString("b")
		case "63":
			decoded.WriteString("c")
		case "64":
			decoded.WriteString("d")
		case "65":
			decoded.WriteString("e")
		case "66":
			decoded.WriteString("f")
		case "67":
			decoded.WriteString("g")
		case "68":
			decoded.WriteString("h")
		case "69":
			decoded.WriteString("i")
		case "6a":
			decoded.WriteString("j")
		case "6b":
			decoded.WriteString("k")
		case "6c":
			decoded.WriteString("l")
		case "6d":
			decoded.WriteString("m")
		case "6e":
			decoded.WriteString("n")
		case "6f":
			decoded.WriteString("o")
		case "70":
			decoded.WriteString("p")
		case "71":
			decoded.WriteString("q")
		case "72":
			decoded.WriteString("r")
		case "73":
			decoded.WriteString("s")
		case "74":
			decoded.WriteString("t")
		case "75":
			decoded.WriteString("u")
		case "76":
			decoded.WriteString("v")
		case "77":
			decoded.WriteString("w")
		case "78":
			decoded.WriteString("x")
		default:
			// If we don't have a mapping, keep the original hex pair
			decoded.WriteString(hexPair)
		}
	}

	return decoded.String() + port
}

func main() {
	// Test decoding one of the URLs from the API response
	encoded := "175948514e4c4f57175b54575b5307515c050f5c0a0c0f0b0f0c0e590a0c0b5b0a0c0f0d0f0b0e0c0a5a0f590a5a0f090e0f0f0a0e0d0e5d0a010f0c0e010e0f0e0a0a5a0e010e080a5a0e000e0f0f0c0f0b0f0a0e010a5a0b0f0b5d0b0c0b0c0b0e0b010e0b0f0e0b5a0b5e0b0a0b090b0d0b080a0c0a590a0c0f0d0f0a0f0c0e0b0e0f0e5a0e0b0f0c0c5e0e0a0a0c0b5b0a0c0e000f0d0e0b0f0c0f080e0b0f0c0a0c0a590a0c0e0a0e0f0f0a0e0b0a0c0b5b0a0c0b0c0b0e0b0c0b0b0a5a0b0e0b090a5a0b0d0b0f0d0a0b0e0b0b0b5b0b0a0b0f0b5b0b0e0b0e0a000b0e0b0e0b0e0d5b0a0c0f5a"

	fmt.Printf("Encoded: %s\n", encoded)
	fmt.Printf("Length: %d\n", len(encoded))

	decoded := decodeSourceURL(encoded)
	fmt.Printf("Decoded: %s\n", decoded)
}
