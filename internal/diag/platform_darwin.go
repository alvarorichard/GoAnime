//go:build darwin

package diag

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

func platformInfo(ctx context.Context, o Options) []Row {
	ctx, cancel := context.WithTimeout(ctx, o.Timeout)
	defer cancel()
	var rows []Row
	if v := run(ctx, "sw_vers", "-productVersion"); v != "" {
		rows = append(rows, Row{"macOS", v})
	}
	if v := run(ctx, "uname", "-r"); v != "" {
		rows = append(rows, Row{"Kernel", v})
	}
	if v := run(ctx, "sysctl", "-n", "hw.memsize"); v != "" {
		if n, err := strconv.ParseUint(v, 10, 64); err == nil {
			rows = append(rows, Row{"Memory", fmt.Sprintf("%.1f GiB total", float64(n)/(1<<30))})
		}
	}
	return rows
}

func run(ctx context.Context, name string, args ...string) string {
	out, err := exec.CommandContext(ctx, name, args...).Output() // #nosec G204 -- fixed system tools, no shell
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
