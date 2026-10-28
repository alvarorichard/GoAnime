package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRenderTransfer_Running(t *testing.T) {
	t.Parallel()
	bar := NewTransferBar()
	tr := &Transfer{
		Label: "Download", Gloss: "ダウンロード",
		Title: "[PT-BR] Namakura Gatana", Detail: "Episode 1",
		Bar: bar.ViewAs(0.62), Percent: 0.62,
		Stats:  ByteStats(412<<20, 664<<20, 8.4*(1<<20)),
		Status: "Downloading with yt-dlp",
	}

	lines := strings.Split(ansi.Strip(RenderTransfer(tr, 100)), "\n")

	assert.Equal(t, []string{
		"",
		" GOANIME   DOWNLOAD  ダウンロード",
		"",
		" [PT-BR] Namakura Gatana",
		" Episode 1",
		"",
		" " + strings.Repeat("━", TransferBarWidth(defaultShellWidth)) + "   62%",
		" 412.0 MB of 664.0 MB · 8.4 MB/s · 00:30 left",
		"",
		" Downloading with yt-dlp",
		"",
		"  ctrl+c  cancel",
		"",
	}, lines)
}

func TestRenderTransfer_Outcomes(t *testing.T) {
	t.Parallel()

	done := ansi.Strip(RenderTransfer(&Transfer{Label: "Download", Percent: 1, State: TransferDone, Status: "Download complete"}, 80))
	assert.Contains(t, done, "100%")
	assert.Contains(t, done, "✓ Download complete")
	assert.NotContains(t, done, "ctrl+c", "nothing left to cancel")

	failed := ansi.Strip(RenderTransfer(&Transfer{Label: "Download", Percent: 0.31, State: TransferFailed, Status: "Download failed: HTTP 404"}, 80))
	assert.Contains(t, failed, "31%")
	assert.Contains(t, failed, " !  Download failed: HTTP 404")
	assert.NotContains(t, failed, "ctrl+c")

	assert.Contains(t, ansi.Strip(RenderTransfer(&Transfer{State: TransferDone}, 80)), "✓ Done")
	assert.Contains(t, ansi.Strip(RenderTransfer(&Transfer{State: TransferFailed}, 80)), "Failed")
}

func TestRenderTransfer_FitsTheWidth(t *testing.T) {
	t.Parallel()
	bar := NewTransferBar()
	bar.SetWidth(TransferBarWidth(30))
	tr := &Transfer{
		Label: "Download", Gloss: "ダウンロード", Title: strings.Repeat("A very long title ", 5),
		Bar: bar.ViewAs(0.5), Percent: 0.5, Stats: ByteStats(1<<30, 2<<30, 1<<20),
		Status: strings.Repeat("status ", 20),
		Tasks:  []TransferTask{{Label: "Episode 1", State: TaskActive, Percent: 0.5}, {Label: "Episode 2"}},
	}
	for _, line := range strings.Split(RenderTransfer(tr, 30), "\n") {
		assert.LessOrEqual(t, lipgloss.Width(line), 30)
	}
}

func TestRenderTaskRows(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)

	assert.Empty(t, renderTaskRows(&theme, []TransferTask{{Label: "Episode 1"}}), "a single download lists nothing")

	var tasks []TransferTask
	for i := 1; i <= 12; i++ {
		task := TransferTask{Label: fmt.Sprintf("Episode %d", i)}
		switch {
		case i <= 4:
			task.State, task.Percent = TaskDone, 1
		case i == 5:
			task.State = TaskFailed
		case i <= 7:
			task.State, task.Percent = TaskActive, 0.4
		}
		tasks = append(tasks, task)
	}
	rows := renderTaskRows(&theme, tasks)
	plain := make([]string, len(rows))
	for i, row := range rows {
		plain[i] = strings.TrimRight(ansi.Strip(row), " ")
	}

	require.Len(t, plain, 9)
	assert.Equal(t, " ✓ 2 more done", plain[0])
	assert.True(t, strings.HasPrefix(plain[1], " ✓ Episode 3 "), plain[1])
	assert.True(t, strings.HasSuffix(plain[2], "done"), plain[2])
	assert.True(t, strings.HasPrefix(plain[3], " ! Episode 5 "), plain[3])
	assert.True(t, strings.HasSuffix(plain[3], "failed"), plain[3])
	assert.True(t, strings.HasPrefix(plain[4], " ◐ Episode 6 "), plain[4])
	assert.True(t, strings.HasSuffix(plain[4], "40%"), plain[4])
	assert.True(t, strings.HasSuffix(plain[6], "waiting"), plain[6])
	assert.True(t, strings.HasPrefix(plain[7], "   Episode 9 "), plain[7])
	assert.Equal(t, "   3 more waiting", plain[8])

	// Labels of different lengths still line up their bars.
	assert.Equal(t, strings.Index(plain[1], "━"), strings.Index(plain[4], "━"))
}

func TestRenderTaskRows_ShowsTheLastWaitingRowInsteadOfCountingOne(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)
	tasks := []TransferTask{{Label: "Episode 1", State: TaskActive}, {Label: "Episode 2"}, {Label: "Episode 3"}, {Label: "Episode 4"}}
	plain := ansi.Strip(strings.Join(renderTaskRows(&theme, tasks), "\n"))
	assert.Contains(t, plain, "Episode 4")
	assert.NotContains(t, plain, "more waiting", "\"1 more waiting\" takes the same row as the episode itself")
}

func TestStats(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "", ByteStats(0, 0, 0))
	assert.Equal(t, "512.0 KB", ByteStats(512<<10, 0, 0), "a download of unknown size still says how much is in")
	assert.Equal(t, "1.0 MB of 4.0 MB", ByteStats(1<<20, 4<<20, 0))
	assert.Equal(t, "1.0 MB of 4.0 MB · 1.0 MB/s · 00:03 left", ByteStats(1<<20, 4<<20, 1<<20))
	assert.Equal(t, "4.0 MB of 4.0 MB · 1.0 MB/s", ByteStats(4<<20, 4<<20, 1<<20), "nothing left, no time left")

	assert.Equal(t, "1,412 of 3,000 frames · 38 frames/s · 00:41 left", FrameStats(1412, 3000, 38))
	assert.Equal(t, "12 frames", FrameStats(12, 0, 0))
	assert.Equal(t, "", FrameStats(0, 0, 0))

	assert.Equal(t, "664.0 MB in 01:19", DoneStats(664<<20, 79*time.Second))
	assert.Equal(t, "in 00:05", DoneStats(0, 5*time.Second))
}

func TestFormatBytes(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "0 B", FormatBytes(-1))
	assert.Equal(t, "900 B", FormatBytes(900))
	assert.Equal(t, "1.0 KB", FormatBytes(1024))
	assert.Equal(t, "412.3 MB", FormatBytes(432330342))
	assert.Equal(t, "3.1 GB", FormatBytes(3328599654))
}

func TestFormatCount(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "0", FormatCount(0))
	assert.Equal(t, "999", FormatCount(999))
	assert.Equal(t, "3,000", FormatCount(3000))
	assert.Equal(t, "1,234,567", FormatCount(1234567))
}

func TestRateMeter(t *testing.T) {
	t.Parallel()
	start := time.Unix(1_700_000_000, 0)
	var r RateMeter

	r.Observe(0, start)
	assert.Zero(t, r.Rate(), "one sample is not a rate")

	r.Observe(1000, start.Add(100*time.Millisecond))
	assert.Zero(t, r.Rate(), "samples closer than half a second are ignored")

	r.Observe(1000, start.Add(time.Second))
	assert.InDelta(t, 1000, r.Rate(), 1e-9)

	r.Observe(1000, start.Add(2*time.Second))
	assert.InDelta(t, 700, r.Rate(), 1e-9, "a stall lowers the rate gradually")

	r.Observe(500, start.Add(3*time.Second))
	assert.InDelta(t, 490, r.Rate(), 1e-9, "a count that goes back reads as no progress")
}

func TestNewTransferBar(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)
	bar := NewTransferBar()
	assert.Equal(t, '━', bar.Full)
	assert.Equal(t, '━', bar.Empty)
	assert.False(t, bar.ShowPercentage, "RenderTransfer prints the percentage")
	assert.Equal(t, theme.Meter.GetForeground(), bar.FullColor)
	assert.Equal(t, theme.MeterEmpty.GetForeground(), bar.EmptyColor)

	assert.Equal(t, 10, TransferBarWidth(5))
	assert.Equal(t, 52, TransferBarWidth(60))
	assert.Equal(t, maxTransferBar, TransferBarWidth(400))
}
