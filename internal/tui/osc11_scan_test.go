package tui

import "testing"

// osc11Runes returns s as a rune slice; used by every OSC 11 scan test.
func osc11Runes(s string) []rune { return []rune(s) }

func TestFindOSC11ReplySpan(t *testing.T) {
	t.Parallel()
	runes := osc11Runes
	t.Run("bel terminated", func(t *testing.T) {
		t.Parallel()
		s := "\x1b]11;rgb:0000/0000/0000\x07"
		start, end, ok := findOSC11ReplySpan(runes(s))
		if !ok || start != 0 || end != len(runes(s)) {
			t.Fatalf("span = [%d,%d) ok=%v, want [0,%d) ok=true", start, end, ok, len(runes(s)))
		}
	})
	t.Run("st terminated", func(t *testing.T) {
		t.Parallel()
		s := "\x1b]11;rgb:ffff/ffff/ffff\x1b\\"
		start, end, ok := findOSC11ReplySpan(runes(s))
		if !ok || start != 0 || end != len(runes(s)) {
			t.Fatalf("span = [%d,%d) ok=%v, want full span", start, end, ok)
		}
	})
	t.Run("user text only", func(t *testing.T) {
		t.Parallel()
		if _, _, ok := findOSC11ReplySpan(runes("naruto")); ok {
			t.Fatal("plain typing must never match")
		}
	})
	t.Run("lone esc", func(t *testing.T) {
		t.Parallel()
		if _, _, ok := findOSC11ReplySpan(runes("\x1b")); ok {
			t.Fatal("a lone Escape must wait, not match")
		}
	})
	t.Run("alt chord then reply", func(t *testing.T) {
		t.Parallel()
		s := "\x1bx\x1b]11;rgb:0000/0000/0000\x07"
		start, end, ok := findOSC11ReplySpan(runes(s))
		if !ok || start != 2 || end != len(runes(s)) {
			t.Fatalf("span = [%d,%d) ok=%v, want [2,%d)", start, end, ok, len(runes(s)))
		}
	})
	t.Run("arrows before reply", func(t *testing.T) {
		t.Parallel()
		s := "\x1b[A\x1b]11;rgb:0101/0202/0303\x07"
		start, end, ok := findOSC11ReplySpan(runes(s))
		if !ok || start != 3 {
			t.Fatalf("span = [%d,%d) ok=%v, want start=3", start, end, ok)
		}
	})
	t.Run("split reply waits", func(t *testing.T) {
		t.Parallel()
		for _, s := range []string{"\x1b]11;rgb:0000", "\x1b]11;rgb:0000/0000/0000\x1b"} {
			if _, _, ok := findOSC11ReplySpan(runes(s)); ok {
				t.Fatalf("%q is incomplete and must wait", s)
			}
		}
	})
	t.Run("reply plus trailing typing", func(t *testing.T) {
		t.Parallel()
		s := "\x1b]11;rgb:0000/0000/0000\x07na"
		start, end, ok := findOSC11ReplySpan(runes(s))
		if !ok || start != 0 || end != len(runes(s))-2 {
			t.Fatalf("span = [%d,%d) ok=%v, want [0,%d)", start, end, ok, len(runes(s))-2)
		}
	})
	t.Run("leading typing", func(t *testing.T) {
		t.Parallel()
		s := "na\x1b]11;rgb:0000/0000/0000\x07"
		start, end, ok := findOSC11ReplySpan(runes(s))
		if !ok || start != 2 || end != len(runes(s)) {
			t.Fatalf("span = [%d,%d) ok=%v, want [2,%d)", start, end, ok, len(runes(s)))
		}
	})
}

// TestFindOSC11ReplySpan_NeverMistakesTyping is the regression guard for the
// startup freeze: the old reader treated ANY stdin bytes as the terminal's
// answer (ok = len > 0), so the first typed characters were swallowed as a
// "reply". Nothing below may ever match — each of these is user input, a
// key chord, or a truncated header, never a complete OSC 11 answer.
func TestFindOSC11ReplySpan_NeverMistakesTyping(t *testing.T) {
	t.Parallel()
	hostile := []string{
		"naruto",
		"todo poderoso 2003",
		"one piece ep 1000",
		"\x07",                         // lone BEL terminates nothing
		"rgb:ffff/ffff/ffff",            // body without the header
		"]11;rgb:ffff/ffff/ffff\x07",    // header without the opening ESC
		"\x1b",                          // lone Escape key
		"\x1b\x1b",                      // double Escape
		"\x1bx",                         // Alt+x chord
		"\x1b[A",                        // up arrow
		"\x1b[B\x1b[C",                  // arrows
		"\x1b]11",                       // truncated header
		"\x1b]11;",                      // header without body
		"\x1b]11;rgb:0000",              // body without terminator
		"\x1b]10;rgb:ffff/ffff/ffff\x07", // OSC 10 (foreground), not 11
		"\x1b]104\x07",                  // color reset, not 11
		" conclusions \x1b[G",          // stray CSI
		"\r\n",                          // enter
		"\x7f",                          // backspace
		"açaí",                          // multibyte typing
		"フリーレン",                        // CJK typing
	}
	for _, s := range hostile {
		if _, _, ok := findOSC11ReplySpan(osc11Runes(s)); ok {
			t.Fatalf("%q must never match: it is input, not an answer", s)
		}
	}
}

// TestFindOSC11ReplySpan_RandomJunkNeverMatches fuzzes the negative space:
// 2000 deterministic pseudo-random printable streams without ESC must never
// match, so a future loosening of the pattern cannot reintroduce eaten
// keystrokes silently.
func TestFindOSC11ReplySpan_RandomJunkNeverMatches(t *testing.T) {
	t.Parallel()
	seed := uint64(0xC0FFEE)
	next := func() uint64 {
		seed ^= seed << 13
		seed ^= seed >> 7
		seed ^= seed << 17
		return seed
	}
	for i := 0; i < 2000; i++ {
		n := 1 + int(next()%48)
		buf := make([]rune, n)
		for j := range buf {
			// Printable ASCII except ESC.
			buf[j] = rune(0x20 + next()% (0x7E - 0x20))
			if buf[j] == 0x1B {
				buf[j] = 'x'
			}
		}
		if _, _, ok := findOSC11ReplySpan(buf); ok {
			t.Fatalf("random junk %q matched", string(buf))
		}
	}
}

// TestFindOSC11ReplySpan_SplitReads feeds a reply byte by byte like a slow
// terminal would: nothing matches until the terminator lands, then the full
// span matches exactly once.
func TestFindOSC11ReplySpan_SplitReads(t *testing.T) {
	t.Parallel()
	full := osc11Runes("\x1b]11;rgb:0101/0202/0303\x07")
	matched := 0
	for i := 1; i <= len(full); i++ {
		_, _, ok := findOSC11ReplySpan(full[:i])
		if ok {
			matched++
			if i != len(full) {
				t.Fatalf("prefix of %d runes matched before the terminator", i)
			}
		}
	}
	if matched != 1 {
		t.Fatalf("full reply matched %d times, want exactly 1", matched)
	}
}
