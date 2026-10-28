package tui

// findOSC11ReplySpan locates a complete OSC 11 background reply in chars:
// ESC ] 1 1 ; <body> (BEL | ESC \).
//
// It returns the span [start, end) covering the reply including its
// terminator. Anything incomplete — a reply split across reads, a lone ESC
// from the Escape key, an Alt+key chord (ESC plus a letter) — reports
// ok=false so the caller keeps waiting instead of mistaking user input for
// a terminal answer.
func findOSC11ReplySpan(chars []rune) (start, end int, ok bool) {
	for i := 0; i < len(chars); i++ {
		if chars[i] != 0x1B {
			continue
		}
		if i+4 >= len(chars) {
			// ESC at the tail: reply header or Esc key, both need more data.
			return 0, 0, false
		}
		if chars[i+1] != ']' || chars[i+2] != '1' || chars[i+3] != '1' || chars[i+4] != ';' {
			// Some other escape (arrows, Alt+key): not our reply.
			continue
		}
		for j := i + 5; j < len(chars); j++ {
			switch chars[j] {
			case 0x07: // BEL terminates the reply.
				return i, j + 1, true
			case 0x1B:
				if j+1 >= len(chars) {
					// Split ST terminator: wait for the backslash.
					return 0, 0, false
				}
				if chars[j+1] == '\\' {
					return i, j + 2, true
				}
				// Stray ESC inside the body: keep scanning conservatively.
			}
		}
		// Header without a terminator yet: the rest is still arriving.
		return 0, 0, false
	}
	return 0, 0, false
}
