//go:build !windows

package tui

import "time"

// queryBackgroundWindows is the Windows console-queue probe
// (osc11_windows.go): queryBackgroundIsLight dispatches on runtime.GOOS, and
// outside GOOS == "windows" it always falls through to the raw-read probe,
// so this stub only exists for the call site to compile. It answers nothing
// by design and is never called.
func queryBackgroundWindows(time.Duration) (light, ok bool) { return false, false }
