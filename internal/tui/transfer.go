package tui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// TransferState is where a download or an upscale stands.
type TransferState int

const (
	TransferRunning TransferState = iota
	TransferDone
	TransferFailed
)

// TaskState is where one part of a batch stands.
type TaskState int

const (
	TaskWaiting TaskState = iota
	TaskActive
	TaskDone
	TaskFailed
)

// TransferTask is one part of a batch, such as one episode of a range.
type TransferTask struct {
	Label   string
	Percent float64
	State   TaskState
}

// Transfer is what a progress screen shows, from the step it belongs to
// down to the latest word from the work.
type Transfer struct {
	Label  string // the step, drawn in capitals: "Download", "Upscale"
	Gloss  string // its Japanese
	Title  string // what is being worked on: the show, the file
	Detail string // which part of it: "Episode 1", "12 episodes"

	Bar     string  // the bar, as a progress.Model from NewTransferBar draws it
	Percent float64 // 0–1, printed beside the bar
	Stats   string  // what the numbers say, from ByteStats or FrameStats
	Tasks   []TransferTask
	Status  string // the latest message from the work
	State   TransferState
}

const (
	// maxTransferBar keeps the bar readable on a very wide terminal.
	maxTransferBar = 64
	// taskBarWidth is the bar of one row in a batch.
	taskBarWidth = 20
	// shownDoneTasks and shownWaitingTasks are the finished and waiting
	// rows a batch lists before it counts the rest.
	shownDoneTasks    = 2
	shownWaitingTasks = 2
)

// NewTransferBar is the bar a progress screen draws: the accent on the
// ground's rule color, in the same heavy line as the panels' bars, without
// its own percentage (RenderTransfer prints it). It keeps progress.Model's
// spring, so the bar glides between updates.
func NewTransferBar() progress.Model {
	theme := NewTheme(true)
	bar := progress.New(
		progress.WithColors(theme.Meter.GetForeground()),
		progress.WithFillCharacters('━', '━'),
		progress.WithoutPercentage(),
		progress.WithWidth(TransferBarWidth(defaultShellWidth)),
	)
	bar.EmptyColor = theme.MeterEmpty.GetForeground()
	return bar
}

// TransferBarWidth is the bar's width on a terminal width cells wide, leaving
// room for the margin and the percentage.
func TransferBarWidth(width int) int {
	return max(min(width-8, maxTransferBar), 10)
}

// RenderTransfer draws t for a terminal width cells wide. It is drawn in
// line with the terminal's output, not on the alternate screen, so the last
// frame stays behind as a record of what happened.
func RenderTransfer(t *Transfer, width int) string {
	theme := NewTheme(true)
	if width <= 0 {
		width = defaultShellWidth
	}

	lines := []string{""}
	head := theme.Header.Render("GOANIME") + "  " + theme.Primary.Render(strings.ToUpper(singleLine(t.Label)))
	if gloss := singleLine(t.Gloss); gloss != "" {
		head += theme.Gloss.Render("  " + gloss)
	}
	lines = append(lines, head, "")

	if title := singleLine(t.Title); title != "" {
		lines = append(lines, " "+theme.Text.Bold(true).Render(title))
	}
	if detail := singleLine(t.Detail); detail != "" {
		lines = append(lines, " "+theme.Muted.Render(detail))
	}
	if len(lines) > 3 {
		lines = append(lines, "")
	}

	lines = append(lines, " "+t.Bar+theme.Text.Bold(true).Render(fmt.Sprintf("%5.0f%%", clampPercent(t.Percent)*100)))
	if stats := singleLine(t.Stats); stats != "" {
		lines = append(lines, " "+theme.Muted.Render(stats))
	}

	if rows := renderTaskRows(&theme, t.Tasks); len(rows) > 0 {
		lines = append(lines, "")
		lines = append(lines, rows...)
	}

	status := singleLine(t.Status)
	switch t.State {
	case TransferDone:
		if status == "" {
			status = "Done"
		}
		lines = append(lines, "", " "+theme.Primary.Render("✓ "+status))
	case TransferFailed:
		if status == "" {
			status = "Failed"
		}
		lines = append(lines, "", " "+theme.Notice.Render("!")+" "+theme.Text.Render(status))
	default:
		if status != "" {
			lines = append(lines, "", " "+theme.Soft.Render(status))
		}
		lines = append(lines, "", " "+theme.Keycap.Render("ctrl+c")+theme.Footer.Render(" cancel"))
	}

	// A last empty line: Bubble Tea leaves the cursor at the start of the
	// view's last line when it quits, and whatever the app prints next
	// would write over the outcome.
	lines = append(lines, "")
	for i, line := range lines {
		lines[i] = ansi.Truncate(line, width, "")
	}
	return strings.Join(lines, "\n")
}

// renderTaskRows lists a batch's parts: the last finished ones (the rest
// counted), every failed and active one, and the next ones waiting (the rest
// counted).
func renderTaskRows(theme *Theme, tasks []TransferTask) []string {
	if len(tasks) < 2 {
		return nil
	}
	var done, failed, active, waiting []TransferTask
	labelWidth := 0
	for _, task := range tasks {
		labelWidth = max(labelWidth, lipgloss.Width(singleLine(task.Label)))
		switch task.State {
		case TaskDone:
			done = append(done, task)
		case TaskFailed:
			failed = append(failed, task)
		case TaskActive:
			active = append(active, task)
		default:
			waiting = append(waiting, task)
		}
	}
	label := func(task TransferTask) string {
		text := singleLine(task.Label)
		return text + strings.Repeat(" ", labelWidth-lipgloss.Width(text)+2)
	}

	var rows []string
	if len(done) > shownDoneTasks {
		rows = append(rows, theme.Muted.Render(fmt.Sprintf(" ✓ %d more done", len(done)-shownDoneTasks)))
		done = done[len(done)-shownDoneTasks:]
	}
	for _, task := range done {
		rows = append(rows, theme.Muted.Render(" ✓ "+label(task)+strings.Repeat("━", taskBarWidth)+"   done"))
	}
	for _, task := range failed {
		rows = append(rows, theme.Primary.Render(" ! ")+theme.Text.Render(label(task))+
			strings.Repeat(" ", taskBarWidth+3)+theme.Primary.UnsetBold().Render("failed"))
	}
	for _, task := range active {
		rows = append(rows, theme.Primary.UnsetBold().Render(" ◐ ")+theme.Text.Render(label(task))+
			renderBar(taskBarWidth, task.Percent, &theme.Meter, &theme.MeterEmpty)+
			theme.Soft.Render(fmt.Sprintf("%7.0f%%", clampPercent(task.Percent)*100)))
	}
	for i, task := range waiting {
		if i == shownWaitingTasks && len(waiting) > shownWaitingTasks+1 {
			rows = append(rows, theme.Muted.Render(fmt.Sprintf("   %d more waiting", len(waiting)-shownWaitingTasks)))
			break
		}
		rows = append(rows, theme.Muted.Render("   "+label(task)+strings.Repeat(" ", taskBarWidth+3)+"waiting"))
	}
	return rows
}

func clampPercent(p float64) float64 {
	return max(0, min(1, p))
}

// RateMeter estimates how fast a count grows, smoothed so that one burst or
// stall does not swing the time left.
type RateMeter struct {
	last int64
	at   time.Time
	rate float64
}

// Observe records the count reached at now. Samples less than half a second
// after the last one are ignored, so the 25 ms ticks of a progress screen
// do not turn rounding into noise.
func (r *RateMeter) Observe(count int64, now time.Time) {
	if r.at.IsZero() {
		r.last, r.at = count, now
		return
	}
	elapsed := now.Sub(r.at).Seconds()
	if elapsed < 0.5 {
		return
	}
	current := max(float64(count-r.last)/elapsed, 0)
	if r.rate == 0 {
		r.rate = current
	} else {
		r.rate = 0.7*r.rate + 0.3*current
	}
	r.last, r.at = count, now
}

// Rate is the smoothed growth per second; 0 until two samples are in.
func (r *RateMeter) Rate() float64 {
	return r.rate
}

// ByteStats says how much of a download is in: "412.3 MB of 664.0 MB ·
// 8.4 MB/s · 00:29 left". Unknown parts are left out; with nothing known it
// is "".
func ByteStats(received, total int64, rate float64) string {
	var amount string
	switch {
	case total > 0:
		amount = FormatBytes(received) + " of " + FormatBytes(total)
	case received > 0:
		amount = FormatBytes(received)
	default:
		return ""
	}
	return withRate(amount, received, total, rate, FormatBytes(int64(rate))+"/s")
}

// FrameStats says how many frames are done: "1,412 of 3,000 frames ·
// 38 frames/s · 00:41 left".
func FrameStats(done, total int64, rate float64) string {
	var amount string
	switch {
	case total > 0:
		amount = FormatCount(done) + " of " + FormatCount(total) + " frames"
	case done > 0:
		amount = FormatCount(done) + " frames"
	default:
		return ""
	}
	return withRate(amount, done, total, rate, fmt.Sprintf("%.0f frames/s", rate))
}

// withRate adds the speed and the time left to amount, once there is a rate.
func withRate(amount string, done, total int64, rate float64, speed string) string {
	if rate <= 0 {
		return amount
	}
	// Avoid displaying a nonsensical multi-year ETA when the byte/frame rate
	// is positive only due to smoothing but rounds to zero in the UI.
	if rate < 1 && total > done {
		return amount + " · stalled"
	}
	parts := []string{amount, speed}
	if total > done {
		parts = append(parts, FormatClock(float64(total-done)/rate)+" left")
	}
	return strings.Join(parts, " · ")
}

// DoneStats sums up a finished transfer: "664.0 MB in 01:19".
func DoneStats(total int64, elapsed time.Duration) string {
	if total <= 0 {
		return "in " + FormatClock(elapsed.Seconds())
	}
	return FormatBytes(total) + " in " + FormatClock(elapsed.Seconds())
}

// FormatBytes renders a size in binary units: "900 B", "512.0 KB", "3.1 GB".
func FormatBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", max(n, 0))
	}
	value, suffix := float64(n), "B"
	for _, next := range []string{"KB", "MB", "GB", "TB"} {
		if value < unit {
			break
		}
		value /= unit
		suffix = next
	}
	return fmt.Sprintf("%.1f %s", value, suffix)
}

// FormatCount groups thousands: 3000 → "3,000".
func FormatCount(n int64) string {
	s := fmt.Sprintf("%d", n)
	if n < 0 {
		return s
	}
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
