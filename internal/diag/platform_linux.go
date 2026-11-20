//go:build linux

package diag

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

func platformInfo(_ context.Context, _ Options) []Row {
	var rows []Row
	if name := osReleaseField("PRETTY_NAME"); name != "" {
		rows = append(rows, Row{"Distribution", name})
	}
	if b, err := os.ReadFile("/proc/sys/kernel/osrelease"); err == nil {
		rows = append(rows, Row{"Kernel", strings.TrimSpace(string(b))})
	}
	if b, err := os.ReadFile("/proc/version"); err == nil && strings.Contains(strings.ToLower(string(b)), "microsoft") {
		rows = append(rows, Row{"Virtualization", "WSL"})
	} else if fileExists("/.dockerenv") || fileExists("/run/.containerenv") {
		rows = append(rows, Row{"Virtualization", "container"})
	}
	if total, avail := memInfo(); total > 0 {
		rows = append(rows, Row{"Memory", fmt.Sprintf("%s total, %s available", gib(total), gib(avail))})
	}
	return rows
}

func osReleaseField(key string) string {
	f, err := os.Open("/etc/os-release")
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if v, ok := strings.CutPrefix(sc.Text(), key+"="); ok {
			return strings.Trim(v, `"'`)
		}
	}
	return ""
}

// memInfo returns MemTotal and MemAvailable in bytes.
func memInfo() (total, avail uint64) {
	f, err := os.Open("/proc/meminfo")
	if err != nil {
		return 0, 0
	}
	defer func() { _ = f.Close() }()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		var kb uint64
		switch {
		case strings.HasPrefix(sc.Text(), "MemTotal:"):
			_, _ = fmt.Sscanf(strings.TrimPrefix(sc.Text(), "MemTotal:"), "%d", &kb)
			total = kb << 10
		case strings.HasPrefix(sc.Text(), "MemAvailable:"):
			_, _ = fmt.Sscanf(strings.TrimPrefix(sc.Text(), "MemAvailable:"), "%d", &kb)
			avail = kb << 10
		}
	}
	return total, avail
}

func fileExists(p string) bool { _, err := os.Stat(p); return err == nil }

func gib(b uint64) string { return fmt.Sprintf("%.1f GiB", float64(b)/(1<<30)) }
