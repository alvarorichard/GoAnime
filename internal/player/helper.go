package player

import (
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
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
// and updates the progress bar. It then schedules the next tick.
//
// 2. `statusMsg`: Updates the status string in the model, which can be used to display custom messages
// to the user, such as "Downloading..." or "Download complete".
//
// 3. `progress.FrameMsg`: Handles frame updates for the progress bar. It delegates the update to the
// internal `progress.Model` and returns any commands necessary to refresh the UI.
//
// 4. `tea.KeyMsg`: Responds to key events, such as quitting the program when "Ctrl+C" is pressed.
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
			return m, tea.Quit
		}
		if m.totalBytes > 0 {
			// Calculate current progress percentage with smoothing
			currentProgress := float64(m.received) / float64(m.totalBytes)

			// Ensure progress is within valid bounds
			if currentProgress > 1.0 {
				currentProgress = 1.0
			}
			if currentProgress < 0.0 {
				currentProgress = 0.0
			}

			// Update the progress bar with the new percentage
			cmd := m.progress.SetPercent(currentProgress)
			return m, tea.Batch(cmd, tickCmd())
		}
		return m, tickCmd()

	case statusMsg:
		m.status = string(msg)
		// Force a small progress refresh when status changes
		cmd := m.progress.SetPercent(float64(m.received) / max(1, float64(m.totalBytes)))
		return m, tea.Batch(cmd)

	case progress.FrameMsg:
		var cmd tea.Cmd
		var newModel tea.Model
		newModel, cmd = m.progress.Update(msg)
		m.progress = newModel.(progress.Model)
		return m, cmd

	case tea.KeyMsg:
		if key.Matches(msg, m.keys.quit) {
			m.done = true
			return m, tea.Quit
		}
		return m, nil

	default:
		return m, nil
	}
}

// View renders the Bubble Tea model
// View renders the user interface for the Bubble Tea model.
//
// This function generates the visual output that is displayed to the user. It includes the status message,
// the progress bar, and a quit instruction. The layout is formatted with padding for proper alignment.
//
// Steps:
// 1. Adds padding to each line using spaces.
// 2. Styles the status message (m.status) with an orange color (#FFA500).
// 3. Displays the progress bar using the progress model.
// 4. Shows a message instructing the user to press "Ctrl+C" to quit.
//
// Returns:
// - A formatted string that represents the UI for the current state of the model.
func (m *model) View() string {
	// Creates padding spaces for consistent layout
	pad := strings.Repeat(" ", padding)

	// Styles the status message with an orange color
	statusStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))

	// Returns the UI layout: status message, progress bar, and quit instruction
	return "\n" +
		pad + statusStyle.Render(m.status) + "\n\n" + // Render the styled status message
		pad + m.progress.View() + "\n\n" + // Render the progress bar
		pad + "Press Ctrl+C to quit" // Show quit instruction
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

// Provide a small helper for avoiding div by zero
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
