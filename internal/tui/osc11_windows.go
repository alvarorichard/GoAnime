//go:build windows

package tui

import (
	"encoding/binary"
	"math"
	"os"
	"strings"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

// osc11GracePeriod is the extra peek-only window after the main budget for a
// slow terminal reply. It never consumes user input: bytes are taken only
// when a complete OSC 11 answer is in view.
const osc11GracePeriod = 40 * time.Millisecond

// osc11HoldCap bounds how many input records the probe may hold while
// waiting for a reply. Past it the probe gives up and restores everything.
const osc11HoldCap = 1024

var (
	modkernel32            = windows.NewLazySystemDLL("kernel32.dll")
	procPeekConsoleInputW  = modkernel32.NewProc("PeekConsoleInputW")
	procReadConsoleInputW  = modkernel32.NewProc("ReadConsoleInputW")
	procWriteConsoleInputW = modkernel32.NewProc("WriteConsoleInputW")
)

// windowsInputRecord mirrors the Win32 INPUT_RECORD header: WORD EventType
// followed by the 16-byte event union.
type windowsInputRecord struct {
	EventType uint16
	_         uint16
	Data      [16]byte
}

// recordRune decodes one console record to the character it types, if any: a
// pressed key carrying a Unicode character. Anything else (releases, mouse,
// resize, focus, bare modifiers) carries no text.
func recordRune(rec *windowsInputRecord) (rune, bool) {
	if rec.EventType != windows.KEY_EVENT {
		return 0, false
	}
	if binary.LittleEndian.Uint32(rec.Data[0:4]) == 0 { // key release
		return 0, false
	}
	if u := binary.LittleEndian.Uint16(rec.Data[10:12]); u != 0 {
		return rune(u), true
	}
	return 0, false
}

// osc11Char is one decoded character plus the index of the record that
// carried it, so a matched reply can be consumed while every other record
// is handed back to the console untouched.
type osc11Char struct {
	ch     rune
	recIdx int
}

func pendingConsoleEvents(h windows.Handle) uint32 {
	var n uint32
	if err := windows.GetNumberOfConsoleInputEvents(h, &n); err != nil {
		return 0
	}
	return n
}

func peekConsoleRecords(h windows.Handle, maxRecords uint32) []windowsInputRecord {
	if maxRecords == 0 {
		return nil
	}
	if maxRecords > 256 {
		maxRecords = 256
	}
	buf := make([]windowsInputRecord, maxRecords)
	var n uint32
	// #nosec G103 -- Synchronous Win32 call fills this bounded input buffer and count during the call.
	r1, _, _ := procPeekConsoleInputW.Call(uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(maxRecords), uintptr(unsafe.Pointer(&n))) // nosemgrep: go_unsafe_rule-unsafe
	if r1 == 0 {
		return nil
	}
	return buf[:n]
}

func readConsoleRecords(h windows.Handle, maxRecords uint32) []windowsInputRecord {
	if maxRecords == 0 {
		return nil
	}
	if maxRecords > 256 {
		maxRecords = 256
	}
	buf := make([]windowsInputRecord, maxRecords)
	var n uint32
	// #nosec G103 -- Synchronous Win32 call fills this bounded input buffer and count during the call.
	r1, _, _ := procReadConsoleInputW.Call(uintptr(h), uintptr(unsafe.Pointer(&buf[0])), uintptr(maxRecords), uintptr(unsafe.Pointer(&n))) // nosemgrep: go_unsafe_rule-unsafe
	if r1 == 0 {
		return nil
	}
	return buf[:n]
}

// readAllPendingRecords drains whatever is waiting in the console input
// queue, up to recordLimit records.
func readAllPendingRecords(h windows.Handle, recordLimit int) []windowsInputRecord {
	var out []windowsInputRecord
	for len(out) < recordLimit {
		n := pendingConsoleEvents(h)
		if n == 0 {
			break
		}
		batch := readConsoleRecords(h, n)
		if len(batch) == 0 {
			break
		}
		out = append(out, batch...)
	}
	return out
}

func writeConsoleRecords(h windows.Handle, recs []windowsInputRecord) bool {
	for len(recs) > 0 {
		n := len(recs)
		if n > 256 {
			n = 256
		}
		var wrote uint32
		// #nosec G103 -- Synchronous Win32 call reads the bounded record slice and fills its written count.
		r1, _, _ := procWriteConsoleInputW.Call(uintptr(h), uintptr(unsafe.Pointer(&recs[0])), uintptr(n), uintptr(unsafe.Pointer(&wrote)))
		if r1 == 0 || wrote == 0 {
			return false
		}
		recs = recs[wrote:]
	}
	return true
}

// partitionHeld scans the decoded stream for a complete OSC 11 reply. On a
// match it returns the reply body and the held records minus the reply's own
// records, in original order, ready to be written back to the console.
func partitionHeld(held []windowsInputRecord, stream []osc11Char) (body string, rest []windowsInputRecord, found bool) {
	chars := make([]rune, len(stream))
	for i, c := range stream {
		chars[i] = c.ch
	}
	start, end, ok := findOSC11ReplySpan(chars)
	if !ok {
		return "", nil, false
	}
	inReply := make(map[int]bool, end-start)
	for _, c := range stream[start:end] {
		inReply[c.recIdx] = true
	}
	rest = make([]windowsInputRecord, 0, len(held))
	var sb strings.Builder
	for i := range held {
		if inReply[i] {
			continue
		}
		rest = append(rest, held[i])
	}
	for _, c := range stream[start:end] {
		sb.WriteRune(c.ch)
	}
	return sb.String(), rest, true
}

// queryBackgroundWindows asks the console for its background color the way
// the Unix probe does, but without ever losing user input.
//
// A Windows console (conpty included) funnels the terminal's answer and the
// user's keystrokes through the same input queue, so a blind blocking read
// steals whatever the user types first — and on a silent terminal the
// reader outlives its own timeout, eating input until 64 bytes arrive. This
// instead peeks at the queue, consumes records only once they provably form
// the reply, and writes everything else back verbatim before returning, so
// the TUI that starts next sees every keystroke.
func queryBackgroundWindows(timeout time.Duration) (light, ok bool) {
	h := windows.Handle(os.Stdin.Fd())
	var mode uint32
	if err := windows.GetConsoleMode(h, &mode); err != nil {
		return false, false // stdin is a pipe or file, not a console
	}
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return false, false
	}
	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		return false, false
	}
	defer func() { _ = term.Restore(fd, state) }()

	deadline := time.Now().Add(timeout)
	var held []windowsInputRecord
	var stream []osc11Char
	for {
		if time.Now().After(deadline) {
			// Grace for a slow reply: peek only, consume nothing unless a
			// complete answer is already in view.
			if body, rest, found := peekConsoleReply(h); found {
				writeConsoleRecords(h, rest)
				return parseBackgroundIsLight(body)
			}
			writeConsoleRecords(h, held)
			return false, false
		}
		ev, _ := windows.WaitForSingleObject(h, consoleWaitMilliseconds(time.Until(deadline)))
		if ev != windows.WAIT_OBJECT_0 {
			continue
		}
		batch := readAllPendingRecords(h, osc11HoldCap-len(held))
		if len(batch) == 0 {
			continue
		}
		base := len(held)
		held = append(held, batch...)
		for i := range batch {
			if c, ok := recordRune(&batch[i]); ok {
				stream = append(stream, osc11Char{ch: c, recIdx: base + i})
			}
		}
		if len(held) > osc11HoldCap {
			writeConsoleRecords(h, held)
			return false, false
		}
		if body, rest, found := partitionHeld(held, stream); found {
			writeConsoleRecords(h, rest)
			return parseBackgroundIsLight(body)
		}
	}
}

// consoleWaitMilliseconds bounds Win32 waits without allowing a negative
// duration to wrap or a large duration to become the INFINITE sentinel.
func consoleWaitMilliseconds(remaining time.Duration) uint32 {
	ms := remaining.Milliseconds()
	if ms < 1 {
		return 1
	}
	if ms >= math.MaxUint32 {
		return math.MaxUint32 - 1
	}
	return uint32(ms)
}

// peekConsoleReply looks for a complete OSC 11 answer in the pending input
// without removing anything. On a match it consumes exactly the reply (all
// pending records are taken, the non-reply ones handed back) and returns
// the body plus the records to restore.
func peekConsoleReply(h windows.Handle) (body string, rest []windowsInputRecord, found bool) {
	// Bounded peek loop: more input may arrive while peeking.
	giveUp := time.Now().Add(osc11GracePeriod)
	for {
		n := pendingConsoleEvents(h)
		if n == 0 {
			if time.Now().After(giveUp) {
				return "", nil, false
			}
			time.Sleep(5 * time.Millisecond)
			continue
		}
		recs := peekConsoleRecords(h, n)
		if len(recs) == 0 {
			return "", nil, false
		}
		var stream []osc11Char
		for i := range recs {
			if c, ok := recordRune(&recs[i]); ok {
				stream = append(stream, osc11Char{ch: c, recIdx: i})
			}
		}
		chars := make([]rune, len(stream))
		for i, c := range stream {
			chars[i] = c.ch
		}
		if _, _, ok := findOSC11ReplySpan(chars); !ok {
			return "", nil, false // no full answer in view; leave it all
		}
		// A full answer is in view: take everything pending, keep the
		// answer, hand the rest back.
		all := readAllPendingRecords(h, osc11HoldCap)
		var full []osc11Char
		for i := range all {
			if c, ok := recordRune(&all[i]); ok {
				full = append(full, osc11Char{ch: c, recIdx: i})
			}
		}
		return partitionHeld(all, full)
	}
}
