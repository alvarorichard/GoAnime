//go:build windows

package tui

import (
	"os"
	"testing"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

// synthKey builds a pressed-key console record carrying ch.
func synthKey(ch uint16) windowsInputRecord {
	var rec windowsInputRecord
	rec.EventType = windows.KEY_EVENT
	rec.Data[0] = 1 // bKeyDown
	rec.Data[10] = byte(ch)
	rec.Data[11] = byte(ch >> 8)
	return rec
}

// synthRelease builds a key-release record, which carries no text.
func synthRelease() windowsInputRecord {
	var rec windowsInputRecord
	rec.EventType = windows.KEY_EVENT
	return rec
}

// TestPartitionHeldKeepsTyping verifies the core input-safety contract: user
// keystrokes around a terminal reply survive the probe verbatim, and only
// the reply's own records are consumed.
func TestPartitionHeldKeepsTyping(t *testing.T) {
	t.Parallel()
	// User typed "na", the terminal answered, then the user typed "ru".
	var held []windowsInputRecord
	for _, ch := range []uint16{'n', 'a'} {
		held = append(held, synthKey(ch))
	}
	reply := "\x1b]11;rgb:0000/0000/0000\x07"
	for _, r := range reply {
		held = append(held, synthKey(uint16(r)))
	}
	for _, ch := range []uint16{'r', 'u'} {
		held = append(held, synthKey(ch))
	}
	held = append(held, synthRelease()) // noise carries nothing

	var stream []osc11Char
	for i := range held {
		if c, ok := recordRune(&held[i]); ok {
			stream = append(stream, osc11Char{ch: c, recIdx: i})
		}
	}
	body, rest, found := partitionHeld(held, stream)
	if !found {
		t.Fatal("complete reply in view must match")
	}
	if light, ok := parseBackgroundIsLight(body); !ok || light {
		t.Fatalf("body %q must parse as dark", body)
	}
	// Every non-reply record comes back in order: n, a, r, u, release.
	if len(rest) != 5 {
		t.Fatalf("rest has %d records, want 5 (naru + release)", len(rest))
	}
	var text []rune
	for i := range rest {
		if c, ok := recordRune(&rest[i]); ok {
			text = append(text, c)
		}
	}
	if string(text) != "naru" {
		t.Fatalf("restored typing = %q, want %q", string(text), "naru")
	}
}

// TestPartitionHeldIncompleteWaits ensures a split reply is left alone.
func TestPartitionHeldIncompleteWaits(t *testing.T) {
	t.Parallel()
	held := []windowsInputRecord{synthKey('n')}
	for _, r := range "\x1b]11;rgb:00" {
		held = append(held, synthKey(uint16(r)))
	}
	var stream []osc11Char
	for i := range held {
		if c, ok := recordRune(&held[i]); ok {
			stream = append(stream, osc11Char{ch: c, recIdx: i})
		}
	}
	if _, _, found := partitionHeld(held, stream); found {
		t.Fatal("split reply must wait, not match")
	}
}

// TestRecordRune pins the record decoder: only a pressed key with a
// character carries text. Everything else must decode to nothing, or the
// probe would treat control events as reply bytes.
func TestRecordRune(t *testing.T) {
	t.Parallel()
	pressed := synthKey('a')
	if c, ok := recordRune(&pressed); !ok || c != 'a' {
		t.Fatalf("pressed key = %q,%v want 'a',true", c, ok)
	}
	released := synthRelease()
	if _, ok := recordRune(&released); ok {
		t.Fatal("key release must carry no text")
	}
	var mouse windowsInputRecord
	mouse.EventType = windows.MOUSE_EVENT
	mouse.Data[0] = 1
	if _, ok := recordRune(&mouse); ok {
		t.Fatal("mouse events must carry no text")
	}
	var focus windowsInputRecord
	focus.EventType = windows.FOCUS_EVENT
	if _, ok := recordRune(&focus); ok {
		t.Fatal("focus events must carry no text")
	}
	var resize windowsInputRecord
	resize.EventType = windows.WINDOW_BUFFER_SIZE_EVENT
	if _, ok := recordRune(&resize); ok {
		t.Fatal("resize events must carry no text")
	}
	dead := synthKey('a')
	dead.Data[10], dead.Data[11] = 0, 0 // modifier-only press, no character
	if _, ok := recordRune(&dead); ok {
		t.Fatal("a press without a character must carry no text")
	}
}

// TestPartitionHeld_ReplyOnly covers a quiet terminal: nothing typed, just
// the answer. Nothing is restored, the body parses.
func TestPartitionHeld_ReplyOnly(t *testing.T) {
	t.Parallel()
	var held []windowsInputRecord
	for _, r := range "\x1b]11;rgb:ffff/ffff/ffff\x07" {
		held = append(held, synthKey(uint16(r)))
	}
	var stream []osc11Char
	for i := range held {
		if c, ok := recordRune(&held[i]); ok {
			stream = append(stream, osc11Char{ch: c, recIdx: i})
		}
	}
	body, rest, found := partitionHeld(held, stream)
	if !found {
		t.Fatal("complete reply must match")
	}
	if len(rest) != 0 {
		t.Fatalf("nothing typed, rest must be empty, got %d records", len(rest))
	}
	if light, ok := parseBackgroundIsLight(body); !ok || !light {
		t.Fatalf("body %q must parse as light", body)
	}
}

// TestPartitionHeld_PreservesEverythingElse mixes every record kind around
// a reply: mouse, resize, focus, releases and a 300-character paste. The
// probe must hand back every single non-reply record in order.
func TestPartitionHeld_PreservesEverythingElse(t *testing.T) {
	t.Parallel()
	var held []windowsInputRecord
	var mouse windowsInputRecord
	mouse.EventType = windows.MOUSE_EVENT
	held = append(held, mouse)
	paste := ""
	for i := 0; i < 300; i++ {
		paste += string(rune('a' + i%26))
	}
	for _, r := range paste {
		held = append(held, synthKey(uint16(r)))
	}
	held = append(held, synthRelease())
	var resize windowsInputRecord
	resize.EventType = windows.WINDOW_BUFFER_SIZE_EVENT
	held = append(held, resize)
	reply := "\x1b]11;rgb:0000/0000/0000\x1b\\"
	replyStart := len(held)
	for _, r := range reply {
		held = append(held, synthKey(uint16(r)))
	}
	var focus windowsInputRecord
	focus.EventType = windows.FOCUS_EVENT
	held = append(held, focus)

	var stream []osc11Char
	for i := range held {
		if c, ok := recordRune(&held[i]); ok {
			stream = append(stream, osc11Char{ch: c, recIdx: i})
		}
	}
	_, rest, found := partitionHeld(held, stream)
	if !found {
		t.Fatal("complete reply must match")
	}
	wantRest := len(held) - (len(reply))
	if len(rest) != wantRest {
		t.Fatalf("rest has %d records, want %d (everything but the reply)", len(rest), wantRest)
	}
	for i := range rest {
		want := held[i]
		if i >= replyStart {
			want = held[i+len(reply)]
		}
		if rest[i] != want {
			t.Fatalf("record %d not restored verbatim", i)
		}
	}
}

// TestQueryBackgroundWindows_PipeSkipsFast pins the headless contract: with
// piped stdin (no console) the probe skips without probing, fast. It must
// never block a headless run.
func TestQueryBackgroundWindows_PipeSkipsFast(t *testing.T) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("needs piped stdin; running on a real console")
	}
	done := make(chan struct{})
	var light, ok bool
	go func() {
		light, ok = queryBackgroundWindows(osc11Timeout)
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		t.Fatal("probe hung on piped stdin")
	}
	if ok || light {
		t.Fatalf("piped stdin must skip: got light=%v ok=%v", light, ok)
	}
}

// TestPeekConsoleReply_NoConsole pins the grace path headless: nothing
// pending, nothing found, no hang.
func TestPeekConsoleReply_NoConsole(t *testing.T) {
	if term.IsTerminal(int(os.Stdin.Fd())) {
		t.Skip("needs piped stdin; running on a real console")
	}
	h := windows.Handle(os.Stdin.Fd())
	if _, _, found := peekConsoleReply(h); found {
		t.Fatal("empty pipe must never match")
	}
}
