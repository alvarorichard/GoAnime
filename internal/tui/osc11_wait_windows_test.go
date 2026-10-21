//go:build windows

package tui

import (
	"math"
	"testing"
	"time"
)

func TestConsoleWaitMilliseconds(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name      string
		remaining time.Duration
		want      uint32
	}{
		{name: "expired deadline", remaining: -time.Millisecond, want: 1},
		{name: "deadline reached", remaining: 0, want: 1},
		{name: "submillisecond remainder", remaining: time.Microsecond, want: 1},
		{name: "normal wait", remaining: 5 * time.Millisecond, want: 5},
		{name: "infinite sentinel", remaining: time.Duration(math.MaxUint32) * time.Millisecond, want: math.MaxUint32 - 1},
		{name: "oversized wait", remaining: time.Duration(math.MaxUint32+10) * time.Millisecond, want: math.MaxUint32 - 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if got := consoleWaitMilliseconds(tt.remaining); got != tt.want {
				t.Fatalf("console wait = %d, want %d", got, tt.want)
			}
		})
	}
}
