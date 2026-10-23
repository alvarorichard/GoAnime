package tui

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This cascade plays the Japanese gloss decision through the sessions one
// machine really walks on Windows: a font-less box installs GoAnime Glyphs
// on the first run; the terminal already open keeps the glosses hidden until
// it is reopened; and the classic console draws them only once its own
// Properties pick the glyphs font. Every step feeds the next — what one run
// installs, the next run's font probe answers — so the steps share one
// machine and its counters never rewind. The Linux fontconfig cascade of
// #215 lives in japanese_font_linux_test.go.

func TestJapaneseFontCascadeAcrossSessions(t *testing.T) {
	t.Parallel()
	terminal := []string{"WT_SESSION=1"}

	t.Run("Windows Terminal in a font-less Sandbox", func(t *testing.T) {
		t.Parallel()
		machine := fakeMachine{goos: "windows", fonts: fontMissing}

		got := resolveJapanese(terminal, machine.probes())
		assert.False(t, got, "session 1: no installed face has the glosses, so nothing may draw them")
		require.Equal(t, 1, machine.installed, "session 1: the glyphs font installs")
		assert.True(t, machine.asked > 0, "the probe had to answer before anything installed")

		// The install put the font on disk, but this terminal opened before
		// that: its font collection was built at startup and holds no
		// Japanese face for now. Hiding is the honest answer — keeping the
		// glosses would print boxes the open terminal cannot fix.
		machine.fonts = fontPending
		got = resolveJapanese(terminal, machine.probes())
		assert.False(t, got, "session 2: the terminal predates the font, so nothing draws yet")
		require.Equal(t, 1, machine.installed, "session 2: pending is not missing — nothing installs again")

		// Reopened after the install, the terminal enumerates the glyphs
		// and the glosses just draw.
		machine.fonts = fontPresent
		got = resolveJapanese(terminal, machine.probes())
		assert.True(t, got, "session 3: the reopened terminal carries the glyphs")
		assert.Equal(t, 1, machine.installed, "session 3: still a single install across the cascade")
	})

	t.Run("classic console in the same Sandbox", func(t *testing.T) {
		t.Parallel()
		machine := fakeMachine{goos: "windows", classic: true, fonts: fontMissing}

		got := resolveJapanese(nil, machine.probes())
		assert.False(t, got, "session 1: Consolas cannot, the glosses stay off")
		require.Equal(t, 1, machine.installed,
			"session 1: the glyphs font reaches the console's own font list anyway")

		// The font sits on disk now, but the console still draws Consolas —
		// conhost never falls back for a missing glyph. Pending is not
		// missing: the console can only be helped by its Properties, not by
		// a second install.
		machine.fonts = fontPending
		got = resolveJapanese(nil, machine.probes())
		assert.False(t, got, "session 2: the console's font still lacks the glyphs")
		require.Equal(t, 1, machine.installed, "session 2: no install lands on a pending state")

		// The user picked GoAnime Glyphs in the console's Properties, so the
		// coverage probe asks the console's chosen face — and it answers.
		machine.covers = true
		got = resolveJapanese(nil, machine.probes())
		assert.True(t, got, "session 3: the console's own font draws the glosses")
		assert.Equal(t, 1, machine.installed, "session 3: still a single install across the cascade")

		// Whatever the probes said, the env override draws — how a user who
		// insists pushes through a probe that cannot answer.
		got = resolveJapanese([]string{"GOANIME_JAPANESE=1"}, machine.probes())
		assert.True(t, got, "the override wins over the cascade without asking again")
	})
}
