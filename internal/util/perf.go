// Package util provides performance profiling utilities
package util

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"charm.land/lipgloss/v2"
)

// PerfEnabled indicates if performance profiling is enabled
var PerfEnabled bool

// PerfMetric represents a single performance measurement
type PerfMetric struct {
	Name      string
	Duration  time.Duration
	StartTime time.Time
	EndTime   time.Time
	Count     int64
	TotalTime time.Duration
}

// PerfTracker tracks performance metrics across the application
type PerfTracker struct {
	mu       sync.RWMutex
	metrics  map[string]*PerfMetric
	started  time.Time
	counters map[string]*int64
}

// Global performance tracker
var (
	globalPerf     *PerfTracker
	globalPerfOnce sync.Once
)

// GetPerfTracker returns the global performance tracker
func GetPerfTracker() *PerfTracker {
	globalPerfOnce.Do(func() {
		globalPerf = &PerfTracker{
			metrics:  make(map[string]*PerfMetric),
			started:  time.Now(),
			counters: make(map[string]*int64),
		}
	})
	return globalPerf
}

// Timer represents an active timing operation
type Timer struct {
	name    string
	start   time.Time
	tracker *PerfTracker
}

// StartTimer starts a new timer for the given operation name
func StartTimer(name string) *Timer {
	if !PerfEnabled {
		return nil
	}
	return &Timer{
		name:    name,
		start:   time.Now(),
		tracker: GetPerfTracker(),
	}
}

// Stop stops the timer and records the duration
func (t *Timer) Stop() time.Duration {
	if t == nil || !PerfEnabled {
		return 0
	}
	duration := time.Since(t.start)
	t.tracker.Record(t.name, duration)
	return duration
}

// Record records a metric with the given name and duration
func (pt *PerfTracker) Record(name string, duration time.Duration) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	metric, exists := pt.metrics[name]
	if !exists {
		metric = &PerfMetric{
			Name:     name,
			Duration: duration,
		}
		pt.metrics[name] = metric
	}

	metric.Count++
	metric.TotalTime += duration

	// Keep track of min/max via Duration field (stores last)
	metric.Duration = duration
}

// IncrementCounter increments a named counter atomically
func (pt *PerfTracker) IncrementCounter(name string) {
	pt.mu.Lock()
	counter, exists := pt.counters[name]
	if !exists {
		var c int64 = 0
		counter = &c
		pt.counters[name] = counter
	}
	pt.mu.Unlock()

	atomic.AddInt64(counter, 1)
}

// GetUptime returns the time since tracking started
func (pt *PerfTracker) GetUptime() time.Duration {
	return time.Since(pt.started)
}

// Performance styles using lipgloss
var (
	perfTitleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true).
			PaddingBottom(1)

	perfHeaderStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ECDC4")).
			Bold(true)

	perfMetricStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#95E1D3"))

	perfValueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFE66D"))

	perfSlowStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true)

	perfFastStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7BED9F"))

	perfSeparatorStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#636E72"))
)

// PrintReport prints a beautiful performance report
func (pt *PerfTracker) PrintReport() {
	if !PerfEnabled {
		return
	}

	pt.mu.RLock()
	defer pt.mu.RUnlock()

	var report strings.Builder

	report.WriteString("\n")
	report.WriteString(perfSeparatorStyle.Render(strings.Repeat("═", 80)))
	report.WriteString("\n")
	report.WriteString(perfTitleStyle.Render("⚡ PERFORMANCE REPORT"))
	report.WriteString("\n")
	report.WriteString(perfSeparatorStyle.Render(strings.Repeat("─", 80)))
	report.WriteString("\n\n")

	// Uptime
	report.WriteString(perfHeaderStyle.Render("📊 Session Info"))
	report.WriteString("\n")
	fmt.Fprintf(&report, "   Uptime: %s\n", perfValueStyle.Render(pt.GetUptime().Round(time.Millisecond).String()))
	report.WriteString("\n")

	// Sort metrics by total time (slowest first)
	type metricEntry struct {
		name   string
		metric *PerfMetric
	}
	var entries []metricEntry
	for name, metric := range pt.metrics {
		entries = append(entries, metricEntry{name, metric})
	}
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].metric.TotalTime > entries[j].metric.TotalTime
	})

	// Print timing metrics
	if len(entries) > 0 {
		report.WriteString(perfHeaderStyle.Render("⏱️  Timing Metrics"))
		report.WriteString("\n")
		fmt.Fprintf(&report, "   %-40s %10s %8s %12s\n",
			perfMetricStyle.Render("Operation"),
			perfMetricStyle.Render("Total"),
			perfMetricStyle.Render("Count"),
			perfMetricStyle.Render("Avg"))
		report.WriteString(perfSeparatorStyle.Render("   " + strings.Repeat("─", 74)))
		report.WriteString("\n")

		for _, entry := range entries {
			avgTime := time.Duration(0)
			if entry.metric.Count > 0 {
				avgTime = entry.metric.TotalTime / time.Duration(entry.metric.Count)
			}

			// Color based on performance
			totalStr := entry.metric.TotalTime.Round(time.Millisecond).String()
			avgStr := avgTime.Round(time.Millisecond).String()

			if entry.metric.TotalTime > 5*time.Second {
				totalStr = perfSlowStyle.Render(totalStr)
			} else if entry.metric.TotalTime < 500*time.Millisecond {
				totalStr = perfFastStyle.Render(totalStr)
			} else {
				totalStr = perfValueStyle.Render(totalStr)
			}

			if avgTime > 2*time.Second {
				avgStr = perfSlowStyle.Render(avgStr)
			} else if avgTime < 200*time.Millisecond {
				avgStr = perfFastStyle.Render(avgStr)
			} else {
				avgStr = perfValueStyle.Render(avgStr)
			}

			// Truncate long names
			name := entry.name
			if len(name) > 38 {
				name = name[:35] + "..."
			}

			fmt.Fprintf(&report, "   %-40s %10s %8d %12s\n",
				perfMetricStyle.Render(name),
				totalStr,
				entry.metric.Count,
				avgStr)
		}
		report.WriteString("\n")
	}

	// Print counters
	if len(pt.counters) > 0 {
		report.WriteString(perfHeaderStyle.Render("🔢 Counters"))
		report.WriteString("\n")
		for name, counter := range pt.counters {
			value := atomic.LoadInt64(counter)
			fmt.Fprintf(&report, "   %-40s %s\n",
				perfMetricStyle.Render(name),
				perfValueStyle.Render(fmt.Sprintf("%d", value)))
		}
		report.WriteString("\n")
	}

	// Summary
	report.WriteString(perfSeparatorStyle.Render(strings.Repeat("─", 80)))
	report.WriteString("\n")

	// Calculate totals
	var totalTime time.Duration
	var totalOps int64
	for _, metric := range pt.metrics {
		totalTime += metric.TotalTime
		totalOps += metric.Count
	}

	report.WriteString(perfHeaderStyle.Render("📈 Summary"))
	report.WriteString("\n")
	fmt.Fprintf(&report, "   Total Operations: %s\n", perfValueStyle.Render(fmt.Sprintf("%d", totalOps)))
	fmt.Fprintf(&report, "   Total Time Tracked: %s\n", perfValueStyle.Render(totalTime.Round(time.Millisecond).String()))

	if totalOps > 0 {
		avgPerOp := totalTime / time.Duration(totalOps)
		fmt.Fprintf(&report, "   Average per Operation: %s\n", perfValueStyle.Render(avgPerOp.Round(time.Millisecond).String()))
	}

	report.WriteString(perfSeparatorStyle.Render(strings.Repeat("═", 80)))
	report.WriteString("\n")

	fmt.Print(report.String())
}

// PerfCount increments a performance counter
func PerfCount(name string) {
	if !PerfEnabled {
		return
	}
	GetPerfTracker().IncrementCounter(name)
}
