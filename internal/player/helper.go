package player

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
)

// Update handles updates to the Bubble Tea model.
//
// This function processes incoming messages (`tea.Msg`) and updates the model's state accordingly.
// It locks the model's mutex to ensure thread safety, especially when modifying shared data like
// `m.received`, `m.totalBytes`, and other stateful properties.
//
// The function processes different message types, including:
//
// 1. `tickMsg`: A periodic message that triggers the progress update. If the download is complete
// (`m.done` is `true`), the program quits. Otherwise, it calculates the percentage of bytes received
// and updates the progress bar, samples the download speed, and schedules the next tick.
//
// 2. `statusMsg`: Updates the status string in the model, which can be used to display custom messages
// to the user, such as "Downloading..." or "Download complete".
//
// 3. `progress.FrameMsg`: Handles frame updates for the progress bar. It delegates the update to the
// internal `progress.Model` and returns any commands necessary to refresh the UI.
//
// 4. `tea.WindowSizeMsg`: Fits the progress bar to the terminal's width.
//
// 5. `tea.KeyMsg`: Responds to key events, such as quitting the program when "Ctrl+C" is pressed.
// If the user requests to quit, the program sets `m.done` to `true` and returns the quit command.
//
// For unhandled message types, it returns the model unchanged.
//
// Returns:
// - Updated `tea.Model` representing the current state of the model.
// - A `tea.Cmd` that specifies the next action the program should perform.
func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch msg := msg.(type) {
	case tickMsg:
		if m.done {
			// Wait a few frames so the terminal redraws before we quit.
			// On error, wait longer so the user can read the status message.
			m.doneFrames++
			maxFrames := 3
			if m.err != nil {
				maxFrames = 80 // ~2 seconds at 25ms/tick
			}
			if m.doneFrames >= maxFrames {
				return m, tea.Quit
			}
			return m, tickCmd()
		}
		m.rate.Observe(m.received, time.Time(msg))
		if m.totalBytes > 0 {
			// Calculate current progress percentage
			currentProgress := float64(m.received) / float64(m.totalBytes)

			// Ensure progress is within valid bounds
			if currentProgress > 0.99 {
				currentProgress = 0.99 // reserve 100% for actual completion
			}

			// Monotonic: never go backward (e.g. when audio file total is added)
			if currentProgress < m.peakPct {
				currentProgress = m.peakPct
			} else {
				m.peakPct = currentProgress
			}

			cmd := m.progress.SetPercent(currentProgress)
			return m, tea.Batch(cmd, tickCmd())
		}
		if m.peakPct > 0 {
			// HLS: use monotonic fragment-based progress when byte totals
			// are unavailable (totalBytes == 0).
			cmd := m.progress.SetPercent(m.peakPct)
			return m, tea.Batch(cmd, tickCmd())
		}
		return m, tickCmd()

	case statusMsg:
		m.status = string(msg)
		return m, nil

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.progress.SetWidth(tui.TransferBarWidth(msg.Width))
		return m, nil

	case progress.FrameMsg:
		var cmd tea.Cmd
		m.progress, cmd = m.progress.Update(msg)
		return m, cmd

	case tea.KeyPressMsg:
		if key.Matches(msg, m.keys.quit) {
			m.done = true
			return m, tea.Quit
		}
		return m, nil

	default:
		return m, nil
	}
}

// View renders the download in the GoAnime theme (tui.RenderTransfer): the
// show and episode, the bar, how much is in and how fast, one row per
// episode of a batch, and the latest status.
//
// When done successfully, the bar is drawn at exactly 100% through ViewAs, so
// the spring animation does not leave it short; on error it stays where it
// stopped.
//
// The download goroutine (downloadAndPlayEpisode) mutates m.done, m.err and
// m.status under m.mu while this render runs on the Bubble Tea loop goroutine,
// so they are read under the same lock — otherwise it's a data race.
func (m *model) View() tea.View {
	m.mu.Lock()
	defer m.mu.Unlock()

	t := tui.Transfer{
		Label:   "Download",
		Gloss:   "ダウンロード",
		Title:   m.title,
		Detail:  m.detail,
		Percent: m.peakPct,
		Status:  m.status,
		Tasks:   m.tasks(),
	}
	switch {
	case m.done && m.err == nil:
		t.State = tui.TransferDone
		t.Bar = m.progress.ViewAs(1.0)
		t.Percent = 1
		t.Stats = tui.DoneStats(max(m.totalBytes, m.received), time.Since(m.started))
	case m.err != nil:
		t.State = tui.TransferFailed
		t.Bar = m.progress.View()
		t.Stats = tui.ByteStats(m.received, m.totalBytes, 0)
	default:
		t.Bar = m.progress.View()
		t.Stats = tui.ByteStats(m.received, m.totalBytes, m.rate.Rate())
	}
	return tea.NewView(tui.RenderTransfer(&t, m.width))
}

// tasks lists the episodes of a batch, in episode order, from the totals and
// received bytes its children report under "episode-N". A single download
// has none.
func (m *model) tasks() []tui.TransferTask {
	if len(m.taskTotals) < 2 {
		return nil
	}
	type entry struct {
		num  int
		task tui.TransferTask
	}
	entries := make([]entry, 0, len(m.taskTotals))
	for id, total := range m.taskTotals {
		num, err := strconv.Atoi(strings.TrimPrefix(id, "episode-"))
		if err != nil {
			continue
		}
		task := tui.TransferTask{Label: fmt.Sprintf("Episode %d", num)}
		received := m.taskReceived[id]
		downloaded, ended := m.taskEnded[id]
		switch {
		case ended && !downloaded:
			task.State = tui.TaskFailed
		case ended, total > 0 && received >= total:
			task.State, task.Percent = tui.TaskDone, 1
		case received > 0:
			task.State = tui.TaskActive
			if total > 0 {
				task.Percent = float64(received) / float64(total)
			}
		}
		entries = append(entries, entry{num: num, task: task})
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].num < entries[j].num })
	tasks := make([]tui.TransferTask, len(entries))
	for i, e := range entries {
		tasks[i] = e.task
	}
	return tasks
}

// tickCmd returns a command that triggers a "tick" every 25 milliseconds.
//
// This function sets up a recurring event (tick) that fires every 25 milliseconds for very smooth updates.
// Each tick sends a `tickMsg` with the current time (`t`) as a message, which can be
// handled by the update function to trigger actions like updating the progress bar.
//
// Returns:
// - A `tea.Cmd` that schedules a tick every 25 milliseconds and sends a `tickMsg`.
func tickCmd() tea.Cmd {
	return tea.Tick(time.Millisecond*25, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
