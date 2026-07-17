package tui

import (
	"io"
	"os"
	"runtime"

	"github.com/charmbracelet/colorprofile"
)

// ResolveColorProfile picks a safe color profile for writer w.
//
// On Windows, colorprofile.Detect returns TrueColor based on the OS build
// alone — even when classic cmd.exe still has VT processing disabled.
// Emitting ANSI then prints raw escape garbage. We only allow color above
// ASCII when VT processing is actually enabled on the target console.
//
// Pure function: pass vtEnabled explicitly so unit tests cover every branch
// without needing a real console.
func ResolveColorProfile(w io.Writer, env []string, vtEnabled bool) colorprofile.Profile {
	p := colorprofile.Detect(w, env)
	if runtime.GOOS != "windows" {
		return p
	}
	// Detect already chose plain text (pipe, NO_COLOR, dumb TERM).
	if p <= colorprofile.ASCII {
		return p
	}
	// Colored Windows console output requires live VT processing.
	if !vtEnabled {
		return colorprofile.ASCII
	}
	return p
}

// ConsoleColorProfile enables VT when possible and returns a safe profile for f.
// Call once at startup (or from InitLogger) before any colored write.
func ConsoleColorProfile(f *os.File) colorprofile.Profile {
	_ = EnableVirtualTerminal()
	return ResolveColorProfile(f, os.Environ(), HasVirtualTerminal(f))
}

// SupportsANSI reports whether it is safe to emit ANSI sequences to f.
// On Windows this means VT processing is active; elsewhere any non-nil file.
func SupportsANSI(f *os.File) bool {
	if f == nil {
		return false
	}
	if runtime.GOOS != "windows" {
		return true
	}
	_ = EnableVirtualTerminal()
	return HasVirtualTerminal(f)
}
