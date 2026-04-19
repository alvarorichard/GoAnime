package tui

import (
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/colorprofile"
)

// BubbleTeaProgramOptions returns default Bubble Tea options that avoid
// terminal capability probes known to leak raw responses in some terminals.
func BubbleTeaProgramOptions(extra ...tea.ProgramOption) []tea.ProgramOption {
	opts := []tea.ProgramOption{
		tea.WithEnvironment(safeBubbleTeaEnvironment()),
		tea.WithColorProfile(colorprofile.TrueColor),
	}
	return append(opts, extra...)
}

// NewProgram constructs a Bubble Tea program with GoAnime's terminal-safe
// defaults. Use this instead of tea.NewProgram for app-owned progress UIs.
func NewProgram(model tea.Model, extra ...tea.ProgramOption) *tea.Program {
	return tea.NewProgram(model, BubbleTeaProgramOptions(extra...)...)
}

// RunClean runs a TUI action with terminal capability probes suppressed and
// drains delayed terminal responses afterwards.
func RunClean(run func() error) error {
	restoreEnv := suppressBubbleTeaQueries()
	defer func() {
		restoreEnv()
		ResetTerminal()
	}()
	return run()
}

func safeBubbleTeaEnvironment() []string {
	env := os.Environ()
	env = setEnvValue(env, "TERM", "xterm-256color")
	env = setEnvValue(env, "TERM_PROGRAM", "Apple_Terminal")
	env = removeEnvValue(env, "WT_SESSION")
	return env
}

func suppressBubbleTeaQueries() func() {
	type savedEnv struct {
		key     string
		value   string
		present bool
	}
	keys := []string{"TERM", "TERM_PROGRAM", "WT_SESSION"}
	saved := make([]savedEnv, 0, len(keys))
	for _, key := range keys {
		value, present := os.LookupEnv(key)
		saved = append(saved, savedEnv{key: key, value: value, present: present})
	}

	_ = os.Setenv("TERM", "xterm-256color")
	_ = os.Setenv("TERM_PROGRAM", "Apple_Terminal")
	_ = os.Unsetenv("WT_SESSION")

	return func() {
		for _, item := range saved {
			if item.present {
				_ = os.Setenv(item.key, item.value)
			} else {
				_ = os.Unsetenv(item.key)
			}
		}
	}
}

func setEnvValue(env []string, key, value string) []string {
	prefix := key + "="
	for i, item := range env {
		if strings.HasPrefix(item, prefix) {
			env[i] = prefix + value
			return env
		}
	}
	return append(env, prefix+value)
}

func removeEnvValue(env []string, key string) []string {
	prefix := key + "="
	filtered := env[:0]
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			continue
		}
		filtered = append(filtered, item)
	}
	return filtered
}
