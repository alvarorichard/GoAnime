package player

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInit_ReturnsBatchCommand(t *testing.T) {
	t.Parallel()
	m := &model{progress: progress.New()}
	cmd := m.Init()
	assert.NotNil(t, cmd, "Init must return a non-nil tea.Cmd")
}

func TestTickCmd_ProducesTickMsg(t *testing.T) {
	t.Parallel()
	cmd := tickCmd()
	require.NotNil(t, cmd)
	msg := cmd()
	_, ok := msg.(tickMsg)
	assert.True(t, ok, "tickCmd should emit tickMsg, got %T", msg)
}

func TestUpdate_TickMsg_DoneTriggersQuitAfterMaxFrames(t *testing.T) {
	t.Parallel()
	m := &model{
		progress:   progress.New(),
		done:       true,
		totalBytes: 100,
		received:   100,
	}
	// Push enough ticks to exceed the 3-frame quiet window (no err).
	var lastCmd tea.Cmd
	for range 4 {
		_, lastCmd = m.Update(tickMsg(time.Now()))
	}
	require.NotNil(t, lastCmd)
}

func TestUpdate_TickMsg_ProgressUpdatesPeak(t *testing.T) {
	t.Parallel()
	m := &model{
		progress:   progress.New(),
		totalBytes: 100,
		received:   50,
	}
	_, cmd := m.Update(tickMsg(time.Now()))
	assert.NotNil(t, cmd)
	assert.InDelta(t, 0.5, m.peakPct, 1e-9)
}

func TestUpdate_StatusMsgUpdatesStatus(t *testing.T) {
	t.Parallel()
	m := &model{progress: progress.New()}
	_, cmd := m.Update(statusMsg("downloading…"))
	assert.Nil(t, cmd)
	assert.Equal(t, "downloading…", m.status)
}

func TestUpdate_UnknownMsgIsNoop(t *testing.T) {
	t.Parallel()
	m := &model{progress: progress.New(), status: "x"}
	_, cmd := m.Update(struct{}{})
	assert.Nil(t, cmd)
	assert.Equal(t, "x", m.status)
}

func TestView_RendersStatusAndQuitInstruction(t *testing.T) {
	t.Parallel()
	m := &model{
		progress: progress.New(progress.WithDefaultBlend()),
		status:   "Downloading episode 7",
	}
	v := m.View()
	assert.Contains(t, v.Content, "Downloading episode 7")
	assert.Contains(t, ansi.Strip(v.Content), "ctrl+c  cancel")
}

func TestNewDownloadModel_View(t *testing.T) {
	t.Parallel()
	m := newDownloadModel("[PT-BR] Namakura Gatana", "Episode 1", 664<<20)
	m.received = 412 << 20
	m.peakPct = float64(m.received) / float64(m.totalBytes)

	view := ansi.Strip(m.View().Content)
	assert.Contains(t, view, "DOWNLOAD  ダウンロード")
	assert.Contains(t, view, "[PT-BR] Namakura Gatana")
	assert.Contains(t, view, "Episode 1")
	assert.Contains(t, view, "62%")
	assert.Contains(t, view, "412.0 MB of 664.0 MB")

	_, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 20})
	for _, line := range strings.Split(m.View().Content, "\n") {
		assert.LessOrEqual(t, lipgloss.Width(line), 60)
	}
}

func TestDownloadModel_DoneAndFailed(t *testing.T) {
	t.Parallel()
	done := newDownloadModel("Show", "Episode 2", 100)
	done.received, done.done, done.status = 100, true, "Download complete"
	view := ansi.Strip(done.View().Content)
	assert.Contains(t, view, "100%")
	assert.Contains(t, view, "✓ Download complete")
	assert.Contains(t, view, "100 B in ")
	assert.NotContains(t, view, "ctrl+c")

	failed := newDownloadModel("Show", "Episode 2", 100)
	failed.received, failed.done, failed.err, failed.status = 30, true, errors.New("HTTP 404"), "Download failed"
	view = ansi.Strip(failed.View().Content)
	assert.Contains(t, view, "!  Download failed")
	assert.NotContains(t, view, "✓")
}

func TestDownloadModel_BatchRows(t *testing.T) {
	t.Parallel()
	m := newDownloadModel("Show", "4 episodes", 0)
	for _, ep := range []int{1, 2, 3, 4} {
		m.childProgress(fmt.Sprintf("episode-%d", ep), 100)
	}
	m.childProgress("episode-1", 100).addProgressReceived(100)
	m.childProgress("episode-2", 100).addProgressReceived(40)
	m.childProgress("episode-3", 100).addProgressReceived(10)
	m.endTask("episode-1", true)
	m.endTask("episode-3", false)

	view := ansi.Strip(m.View().Content)
	assert.Contains(t, view, "✓ Episode 1")
	assert.Contains(t, view, "◐ Episode 2")
	assert.Contains(t, view, "40%")
	assert.Contains(t, view, "! Episode 3")
	assert.Contains(t, view, "failed")
	assert.Contains(t, view, "Episode 4")
	assert.Contains(t, view, "waiting")
}

func TestEndTask_SettlesTheEstimate(t *testing.T) {
	t.Parallel()
	m := newDownloadModel("Show", "2 episodes", 0)
	m.childProgress("episode-1", 300)
	m.childProgress("episode-2", 300)
	m.childProgress("episode-1", 300).addProgressReceived(250)

	m.endTask("episode-1", true)

	assert.Equal(t, int64(250), m.taskTotal("episode-1"), "the real size replaces the estimate")
	assert.Equal(t, int64(550), m.totalBytes)

	var none *model
	assert.NotPanics(t, func() { none.endTask("episode-1", true) })
}

func TestDownloadDetail(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Episode 3", downloadDetail("3"))
	assert.Equal(t, "Episode 12", downloadDetail("Episódio 12"))
}

func TestView_DoneSuccessRendersFullBar(t *testing.T) {
	t.Parallel()
	m := &model{
		progress: progress.New(progress.WithDefaultBlend()),
		status:   "Done",
		done:     true,
	}
	v := m.View()
	assert.Contains(t, v.Content, "Done")
}
