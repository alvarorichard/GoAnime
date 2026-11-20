//go:build windows

package diag

import (
	"context"
	"os/exec"
	"strings"
)

func platformInfo(ctx context.Context, o Options) []Row {
	ctx, cancel := context.WithTimeout(ctx, o.Timeout)
	defer cancel()
	out, err := exec.CommandContext(ctx, "cmd", "/c", "ver").Output() // #nosec G204 -- a fixed system command, no user input
	if err != nil {
		return nil
	}
	return []Row{{"Windows", strings.TrimSpace(string(out))}}
}
