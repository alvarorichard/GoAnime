//go:build windows

package tui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"testing"
	"time"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/term"
)

// End-to-end probe trial on a REAL console. A child process is spawned with
// CREATE_NEW_CONSOLE, giving it a genuine conhost with a real input queue.
// It injects events with WriteConsoleInputW the way the OS delivers keystrokes,
// runs queryBackgroundWindows against the live queue, and prints verdicts.
// The parent re-executes this same test binary with
// -test.run=TestOsc11E2EChildProcess and GOANIME_OSC11_E2E=<scenario>
// and grades the child's stdout.
//
// Scenario timing is tuned so events land DURING the probe: the crash-then-freeze
// mode exercised here is exactly the WezTerm input-eating bug.

const e2eProbeTimeout = 80 * time.Millisecond

var (
	procAllocConsole = modkernel32.NewProc("AllocConsole")
	procCreateFileW  = modkernel32.NewProc("CreateFileW")
)

// procGetCount wraps GetNumberOfConsoleInputEvents with a value return.
func procGetCount2(h windows.Handle) uint32 {
	var n uint32
	_ = windows.GetNumberOfConsoleInputEvents(h, &n)
	return n
}

// e2eStdInput opens the child console's input buffer via CONIN$ — the only
// reliable handle after AllocConsole, because inherited std handles are pipes.
func e2eStdInput() (windows.Handle, error) {
	name, err := windows.UTF16PtrFromString("CONIN$")
	if err != nil {
		return 0, err
	}
	r1, _, callErr := procCreateFileW.Call(
		uintptr(unsafe.Pointer(name)),
		windows.GENERIC_READ|windows.GENERIC_WRITE,
		windows.FILE_SHARE_READ|windows.FILE_SHARE_WRITE,
		0, // security attrs
		windows.OPEN_EXISTING,
		0, // flags
		0) // template file
	if r1 == ^uintptr(0) { // INVALID_HANDLE_VALUE
		return 0, fmt.Errorf("CreateFileW(CONIN$): %v", callErr)
	}
	return windows.Handle(r1), nil
}

// e2eInjectKey appends one key-down/key-up pair carrying ch, like a real press.
func e2eInjectKey(t *testing.T, h windows.Handle, ch uint16) {
	t.Helper()
	down := synthKey(ch)
	down.Data[4] = 1 // wRepeatCount = 1
	up := synthKey(ch)
	up.Data[0] = 0 // bKeyDown = 0 (release)
	up.Data[4] = 1
	recs := []windowsInputRecord{down, up}
	var wrote uint32
	r1, _, callErr := procWriteConsoleInputW.Call(
		uintptr(h), uintptr(unsafe.Pointer(&recs[0])), uintptr(len(recs)), uintptr(unsafe.Pointer(&wrote)))
	if r1 == 0 || wrote != uint32(len(recs)) {
		t.Fatalf("WriteConsoleInputW failed: r1=%d wrote=%d err=%v", r1, wrote, callErr)
	}
}

// e2eInjectString injects a string as key pairs.
func e2eInjectString(t *testing.T, h windows.Handle, s string) {
	t.Helper()
	for _, c := range s {
		e2eInjectKey(t, h, uint16(c))
	}
}

// e2eQueueChars drains the queue and returns the text carried by key-down
// records, in order.
func e2eQueueChars(t *testing.T, h windows.Handle) string {
	t.Helper()
	var sb strings.Builder
	for {
		if procGetCount2(h) == 0 {
			break
		}
		recs := readAllPendingRecords(h, 4096)
		if len(recs) == 0 {
			break
		}
		for i := range recs {
			if c, ok := recordRune(&recs[i]); ok {
				sb.WriteRune(c)
			}
		}
	}
	return sb.String()
}

// e2eRunningUnderHarness reports whether this process is a spawn target.
func e2eRunningUnderHarness() bool { return os.Getenv("GOANIME_OSC11_E2E") != "" }

// TestOsc11E2EChildProcess is the child payload; never selected in the parent.
func TestOsc11E2EChildProcess(t *testing.T) {
	scenario := os.Getenv("GOANIME_OSC11_E2E")
	if scenario == "" {
		return // parent: not our run
	}
	os.Unsetenv("GOANIME_OSC11_E2E")

	if r1, _, err := procAllocConsole.Call(); r1 == 0 {
		t.Fatalf("AllocConsole: %v", err)
	}
	h, err := e2eStdInput()
	if err != nil {
		t.Fatalf("std input: %v", err)
	}
	if !term.IsTerminal(int(h)) {
		t.Fatal("new console stdin is not a terminal")
	}
	// Point Go's stdin at the real console handle.
	os.Stdin = os.NewFile(uintptr(h), "stdin")

	runProbe := func() (light, ok bool) {
		done := make(chan struct{})
		go func() {
			defer close(done)
			light, ok = queryBackgroundWindows(e2eProbeTimeout)
		}()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("probe hung far beyond its own budget")
		}
		return light, ok
	}

	switch scenario {
	case "silent":
		// Terminal never answers. User types DURING the probe (30ms in).
		// After the timeout, the queue must still hold the typing.
		go func() {
			time.Sleep(30 * time.Millisecond)
			e2eInjectString(t, h, "an")
		}()
		light, ok := runProbe()
		if ok {
			fmt.Fprintln(os.Stdout, "SILENT-BAD-ANSWER")
			os.Exit(1)
		}
		_ = light
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if got := e2eQueueChars(t, h); got == "an" {
				fmt.Fprintln(os.Stdout, "SILENT-OK")
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
		fmt.Fprintf(os.Stdout, "SILENT-LOST queue=%q\n", e2eQueueChars(t, h))
		os.Exit(1)

	case "reply":
		// User types FIRST (10ms in), then the terminal answers (30ms
		// in) — both inside the probe window. The probe must consume only
		// the reply and leave the typing queued.
		go func() {
			time.Sleep(10 * time.Millisecond)
			e2eInjectString(t, h, "xy")
		}()
		go func() {
			time.Sleep(30 * time.Millisecond)
			e2eInjectString(t, h, "\x1b]11;rgb:0000/0000/0000\x07")
		}()
		light, ok := runProbe()
		if !ok {
			fmt.Fprintln(os.Stdout, "REPLY-NOT-PARSED")
			os.Exit(1)
		}
		if light {
			fmt.Fprintln(os.Stdout, "REPLY-BAD-LIGHT")
			os.Exit(1)
		}
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if got := e2eQueueChars(t, h); got == "xy" {
				fmt.Fprintln(os.Stdout, "REPLY-OK")
				return
			}
			time.Sleep(5 * time.Millisecond)
		}
		fmt.Fprintln(os.Stdout, "REPLY-LOST-TYPING")
		os.Exit(1)

	default:
		fmt.Fprintf(os.Stdout, "UNKNOWN-SCENARIO %s\n", scenario)
		os.Exit(1)
	}
}

// runE2EChild re-executes this binary as the child and grades its stdout.
func runE2EChild(t *testing.T, scenario, expect string) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=TestOsc11E2EChildProcess", "-test.timeout=60s")
	cmd.Env = append(os.Environ(), "GOANIME_OSC11_E2E="+scenario)
	// DETACHED_PROCESS: the child starts with NO console, so its
	// AllocConsole call creates a fresh conhost of its own. Stdout/stderr
	// stay piped so the parent can read the verdict markers.
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: windows.DETACHED_PROCESS}
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("child %q failed: %v\n%s", scenario, err, out)
	}
	if !strings.Contains(string(out), expect) {
		t.Fatalf("child %q: want %q in output, got:\n%s", scenario, expect, out)
	}
}

// TestOsc11E2E_SilentKeepsTyping: on a silent console, user typing during the
// probe survives in the queue — the WezTerm startup freeze regression.
func TestOsc11E2E_SilentKeepsTyping(t *testing.T) {
	runE2EChild(t, "silent", "SILENT-OK")
}

// TestOsc11E2E_ReplyParsedAndTypingKept: when the terminal answers, the probe
// parses the reply (dark) and the user's later keystrokes remain queued.
func TestOsc11E2E_ReplyParsedAndTypingKept(t *testing.T) {
	runE2EChild(t, "reply", "REPLY-OK")
}
