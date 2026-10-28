package tui

import (
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The SOURCES panel once padded the gap between a source name and its
// outcome with plain spaces. On a real terminal those cells show the
// terminal's own background through the paper panel: black bars on dark
// terminals (white ones on light terminals), right where the demo GIF froze
// them mid-search. panelGap paints the padding with the panel background;
// these tests fail on any paper-panel line with an unpainted space.

// unpaintedGaps counts interior runs of space cells rendered with no
// background set. Edge spaces blend with the ground around the panel and
// are harmless; a run with painted cells on BOTH sides shows the
// terminal's own background through the paper panel — the black bars on
// dark terminals (white on light ones) the demo GIF froze mid-search.
func unpaintedGaps(s string) int {
	type seg struct {
		spaces int
		bg     bool
	}
	var segs []seg
	bg := false
	curSpaces, curOther := 0, 0
	flush := func() {
		if curSpaces+curOther > 0 {
			segs = append(segs, seg{spaces: curSpaces, bg: bg})
		}
		curSpaces, curOther = 0, 0
	}
	rest := s
	for rest != "" {
		if strings.HasPrefix(rest, "\x1b[") {
			end := strings.IndexFunc(rest[2:], func(r rune) bool {
				return r >= '@' && r <= '~'
			})
			if end < 0 {
				break
			}
			seq := rest[2 : 2+end]
			final := rest[2+end]
			rest = rest[2+end+1:]
			if final != 'm' {
				continue
			}
			newBG := bg
			params := strings.Split(seq, ";")
			for i := 0; i < len(params); i++ {
				switch params[i] {
				case "0", "":
					newBG = false
				case "49":
					newBG = false
				case "48":
					newBG = true
					// Skip the color payload: 48;2;r;g;b or 48;5;n.
					if i+1 < len(params) && params[i+1] == "2" {
						i += 4
					} else if i+1 < len(params) && params[i+1] == "5" {
						i += 2
					}
				}
			}
			if newBG != bg {
				flush()
				bg = newBG
			}
			continue
		}
		r := []rune(rest)
		rest = string(r[1:])
		if r[0] == ' ' && !bg {
			curSpaces++
		} else {
			if curSpaces > 0 {
				flush()
			}
			curOther++
		}
	}
	flush()
	bad := 0
	for i, sg := range segs {
		if sg.spaces == 0 || sg.bg {
			continue
		}
		// Painted text on both sides means the run reads as a bar.
		hasBefore, hasAfter := false, false
		for _, o := range segs[:i] {
			if o.bg {
				hasBefore = true
			}
		}
		for _, o := range segs[i+1:] {
			if o.bg {
				hasAfter = true
			}
		}
		if hasBefore && hasAfter {
			bad += sg.spaces
		}
	}
	return bad
}

func TestSourcesPanelHasNoUnpaintedGaps(t *testing.T) {
	t.Parallel()

	theme := NewTheme(true)
	const inner = 44
	rows := []sourcesPanelRow{
		{Name: "AnimeFire", Done: true, Count: 5, Elapsed: 100 * time.Millisecond},
		{Name: "AnimesDigital", Elapsed: 200 * time.Millisecond},
		{Name: "StartFlix", Done: true, Count: 0, Elapsed: 100 * time.Millisecond},
		{Name: "TopCine", Failed: true, Elapsed: 300 * time.Millisecond, Reason: "temporarily unavailable (HTTP 503)"},
		{Name: "A deliberately overlong source name that must truncate", Elapsed: time.Second},
	}
	for _, row := range rows {
		for _, line := range sourcePanelLines(&theme, row, inner) {
			assert.Zero(t, unpaintedGaps(line), "unpainted gap in %q", line)
		}
	}

	off := offPanelLine(&theme, models.SearchOffEntry{Name: "Goyabu", Reason: "off by default"}, inner)
	assert.Zero(t, unpaintedGaps(off), "unpainted gap in %q", off)

	panel := renderSourcesPanel(&theme,
		[]sourcesPanelRow{{Name: "AnimeFire", Done: true, Count: 5}},
		[]models.SearchOffEntry{{Name: "Goyabu", Reason: "off by default"}},
		1, 2, 15*time.Second, true, 60, 30)
	require.NotEmpty(t, panel)
	for _, line := range strings.Split(panel, "\n") {
		assert.Zero(t, unpaintedGaps(line), "unpainted gap in %q", line)
	}
}
