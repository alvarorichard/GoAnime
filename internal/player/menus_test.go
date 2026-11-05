package player

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/upscaler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// menuSeams is a recorder for every seam the menus go through. install swaps
// them in for one test and restores them afterwards; the tests are serial.
type menuSeams struct {
	tty bool

	pickIdx   int
	pickErr   error
	pickCalls int
	gotItems  []tui.PickItem
	gotOpts   tui.PickOptions

	labelsIdx   int
	labelsErr   error
	labelsCalls int
	gotLabels   []string
	gotLabelOpt tui.PickOptions
	labelsSeq   []int // when set, successive answers for the label picker

	confirmAns   bool
	confirmErr   error
	confirmCalls int
	gotConfirm   [4]string

	shaders, gan bool
	modeName     string
	setModes     []upscaler.ShaderMode
	installs     []string
	installErr   error
	ganErr       error
	infos, warns []string
}

func (m *menuSeams) install(t *testing.T) {
	t.Helper()
	oldTTY, oldItems, oldLabels, oldConfirm := stdinIsTerminal, pickItemsFn, pickLabelsFn, confirmFn
	oldName, oldShaders, oldGAN, oldSet := shaderModeNameFn, shadersInstalledFn, ganInstalledFn, setShaderModeFn
	oldInst, oldInstGAN, oldInfo, oldWarn := installShadersFn, installGANShadersFn, infoFn, warnFn
	t.Cleanup(func() {
		stdinIsTerminal, pickItemsFn, pickLabelsFn, confirmFn = oldTTY, oldItems, oldLabels, oldConfirm
		shaderModeNameFn, shadersInstalledFn, ganInstalledFn, setShaderModeFn = oldName, oldShaders, oldGAN, oldSet
		installShadersFn, installGANShadersFn, infoFn, warnFn = oldInst, oldInstGAN, oldInfo, oldWarn
	})

	stdinIsTerminal = func() bool { return m.tty }
	pickItemsFn = func(items []tui.PickItem, o tui.PickOptions) (int, error) {
		m.pickCalls++
		m.gotItems, m.gotOpts = items, o
		return m.pickIdx, m.pickErr
	}
	pickLabelsFn = func(labels []string, o tui.PickOptions) (int, error) {
		i := m.labelsCalls
		m.labelsCalls++
		m.gotLabels, m.gotLabelOpt = labels, o
		if i < len(m.labelsSeq) {
			return m.labelsSeq[i], m.labelsErr
		}
		return m.labelsIdx, m.labelsErr
	}
	confirmFn = func(a, b, c, d string) (bool, error) {
		m.confirmCalls++
		m.gotConfirm = [4]string{a, b, c, d}
		return m.confirmAns, m.confirmErr
	}
	shaderModeNameFn = func() string { return m.modeName }
	shadersInstalledFn = func() bool { return m.shaders }
	ganInstalledFn = func() bool { return m.gan }
	setShaderModeFn = func(mode upscaler.ShaderMode) { m.setModes = append(m.setModes, mode) }
	installShadersFn = func() error { m.installs = append(m.installs, "shaders"); return m.installErr }
	installGANShadersFn = func() error { m.installs = append(m.installs, "gan"); return m.ganErr }
	infoFn = func(msg any, _ ...any) { m.infos = append(m.infos, fmt.Sprint(msg)) }
	warnFn = func(msg any, _ ...any) { m.warns = append(m.warns, fmt.Sprint(msg)) }
}

// ---- downloadMenuItems / downloadMenuHeader / downloadChoiceFor ------------------------------

func labelsOfItems(items []tui.PickItem) []string {
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = it.Label
	}
	return out
}

var downloadItemCases = []struct {
	name   string
	total  int
	status string
	labels []string
	keys   []string
}{
	{"series", 12, "Balanced", []string{"Watch now", "Download this episode", "Download a range...", "Real-time upscale", "Download all 12 episodes..."}, []string{"", "d", "D", "u", "A"}},
	{"two episodes", 2, "Off", []string{"Watch now", "Download this episode", "Download a range...", "Real-time upscale", "Download all 2 episodes..."}, []string{"", "d", "D", "u", "A"}},
	{"a movie has no download-all", 1, "Off", []string{"Watch now", "Download this episode", "Download a range...", "Real-time upscale"}, []string{"", "d", "D", "u"}},
	{"unknown count", 0, "Off", []string{"Watch now", "Download this episode", "Download a range...", "Real-time upscale"}, []string{"", "d", "D", "u"}},
}

func downloadItemFailures(impl func(int, string) []tui.PickItem) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	for _, tc := range downloadItemCases {
		items := impl(tc.total, tc.status)
		var keys []string
		for _, it := range items {
			keys = append(keys, it.Key)
		}
		if !slices.Equal(labelsOfItems(items), tc.labels) || !slices.Equal(keys, tc.keys) {
			f = append(f, fmt.Sprintf("%s: %v %v", tc.name, labelsOfItems(items), keys))
		}
		if len(items) > 3 && items[3].Details != tc.status {
			f = append(f, tc.name+": upscale row does not show the current mode")
		}
		if items[0].Key != "" {
			f = append(f, tc.name+": 'Watch now' must be the plain Enter row")
		}
	}
	return f
}

func TestDownloadMenuItems(t *testing.T) {
	assert.Empty(t, downloadItemFailures(downloadMenuItems))
}

func TestDownloadMenuItemsMeta(t *testing.T) {
	require.Empty(t, downloadItemFailures(downloadMenuItems))
	mutants := map[string]func(int, string) []tui.PickItem{
		"download-all for a single episode": func(n int, s string) []tui.PickItem { return downloadMenuItems(max(n, 2), s) },
		"no download-all": func(n int, s string) []tui.PickItem {
			it := downloadMenuItems(n, s)
			if len(it) > 4 {
				return it[:4]
			}
			return it
		},
		"download first": func(n int, s string) []tui.PickItem {
			it := downloadMenuItems(n, s)
			it[0], it[1] = it[1], it[0]
			return it
		},
		"status hidden": func(n int, _ string) []tui.PickItem { return downloadMenuItems(n, "") },
		"watch gets a key": func(n int, s string) []tui.PickItem {
			it := downloadMenuItems(n, s)
			it[0].Key = "w"
			return it
		},
		"keys swapped": func(n int, s string) []tui.PickItem {
			it := downloadMenuItems(n, s)
			it[1].Key, it[2].Key = it[2].Key, it[1].Key
			return it
		},
		"wrong count in the label": func(n int, s string) []tui.PickItem {
			it := downloadMenuItems(n, s)
			if len(it) > 4 {
				it[4].Label = fmt.Sprintf("Download all %d episodes...", n+1)
			}
			return it
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, downloadItemFailures(m), "mutant %q survived", name)
	}
}

func TestDownloadMenuHeader(t *testing.T) {
	b, c := downloadMenuHeader(5, "Mukashi Banashi")
	assert.Equal(t, "Episode 5 > Options", b)
	assert.Equal(t, "Mukashi Banashi · Episode 5", c)

	b, c = downloadMenuHeader(0, "Movie Title")
	assert.Equal(t, "Playback > Options", b, "no episode number, no episode in the breadcrumb")
	assert.Equal(t, "Movie Title", c)

	b, c = downloadMenuHeader(3, "")
	assert.Equal(t, "Episode 3 > Options", b)
	assert.Empty(t, c, "no title, no context")

	_, c = downloadMenuHeader(2, "Line one\nline two")
	assert.NotContains(t, c, "\n", "the title is cut to a single line")
}

func headerFailures(impl func(int, string) (string, string)) (f []string) {
	for _, tc := range []struct {
		ep           int
		name         string
		wantB, wantC string
	}{
		{5, "Show", "Episode 5 > Options", "Show · Episode 5"},
		{1, "Show", "Episode 1 > Options", "Show · Episode 1"},
		{0, "Show", "Playback > Options", "Show"},
		{4, "", "Episode 4 > Options", ""},
		{0, "", "Playback > Options", ""},
	} {
		if b, c := impl(tc.ep, tc.name); b != tc.wantB || c != tc.wantC {
			f = append(f, fmt.Sprintf("%d/%q: %q %q", tc.ep, tc.name, b, c))
		}
	}
	return f
}

func TestDownloadMenuHeaderMeta(t *testing.T) {
	require.Empty(t, headerFailures(downloadMenuHeader))
	mutants := map[string]func(int, string) (string, string){
		"always playback": func(_ int, n string) (string, string) { return "Playback > Options", n },
		"context without the episode": func(e int, n string) (string, string) {
			b, _ := downloadMenuHeader(e, n)
			return b, n
		},
		"context for a nameless show": func(e int, n string) (string, string) {
			b, c := downloadMenuHeader(e, n)
			if n == "" && e > 0 {
				c = fmt.Sprintf(" · Episode %d", e)
			}
			return b, c
		},
		"episode zero shown": func(e int, n string) (string, string) {
			b, c := downloadMenuHeader(e, n)
			if e == 0 {
				b = "Episode 0 > Options"
			}
			return b, c
		},
		"title dropped": func(e int, _ string) (string, string) { return downloadMenuHeader(e, "") },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, headerFailures(m), "mutant %q survived", name)
	}
}

var choiceCases = []struct {
	idx    int
	hasAll bool
	want   int
}{
	{0, true, downloadWatch}, {1, true, downloadThis}, {2, true, downloadRange}, {3, true, downloadUpscale},
	{4, true, downloadAll},
	{0, false, downloadWatch}, {1, false, downloadThis}, {2, false, downloadRange}, {3, false, downloadUpscale},
	{4, false, downloadWatch}, // no such row without download-all
	{-1, true, downloadWatch}, {9, true, downloadWatch}, {-1, false, downloadWatch},
}

func choiceFailures(impl func(int, bool) int) (f []string) {
	for _, tc := range choiceCases {
		if got := impl(tc.idx, tc.hasAll); got != tc.want {
			f = append(f, fmt.Sprintf("idx %d all=%v: got %d want %d", tc.idx, tc.hasAll, got, tc.want))
		}
	}
	return f
}

func TestDownloadChoiceFor(t *testing.T) {
	assert.Empty(t, choiceFailures(downloadChoiceFor))
	assert.Equal(t, []int{0, 1, 2, 3, 4, 5}, []int{downloadBack, downloadThis, downloadRange, downloadUpscale, downloadWatch, downloadAll},
		"the codes the playback loop switches on are fixed")
}

func TestDownloadChoiceForMeta(t *testing.T) {
	require.Empty(t, choiceFailures(downloadChoiceFor))
	mutants := map[string]func(int, bool) int{
		"anything else downloads everything": func(i int, h bool) int {
			if i < 0 || i > 4 {
				return downloadAll
			}
			return downloadChoiceFor(i, h)
		},
		"row 4 downloads all even without the row": func(i int, _ bool) int { return downloadChoiceFor(i, true) },
		"row 4 never downloads all":                func(i int, _ bool) int { return downloadChoiceFor(i, false) },
		"this and range swapped": func(i int, h bool) int {
			switch downloadChoiceFor(i, h) {
			case downloadThis:
				return downloadRange
			case downloadRange:
				return downloadThis
			}
			return downloadChoiceFor(i, h)
		},
		"watch is back": func(i int, h bool) int {
			if i == 0 {
				return downloadBack
			}
			return downloadChoiceFor(i, h)
		},
		"upscale dropped": func(i int, h bool) int {
			if i == 3 {
				return downloadWatch
			}
			return downloadChoiceFor(i, h)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, choiceFailures(m), "mutant %q survived", name)
	}
}

// ---- askForDownload ------------------------------------------------------------------------

func TestAskForDownload(t *testing.T) {
	tests := []struct {
		name  string
		seams menuSeams
		total int
		want  int
		picks int
	}{
		{"no terminal plays", menuSeams{tty: false, pickIdx: 1}, 12, downloadWatch, 0},
		{"enter watches", menuSeams{tty: true, pickIdx: 0}, 12, downloadWatch, 1},
		{"download this", menuSeams{tty: true, pickIdx: 1}, 12, downloadThis, 1},
		{"download a range", menuSeams{tty: true, pickIdx: 2}, 12, downloadRange, 1},
		{"upscale", menuSeams{tty: true, pickIdx: 3}, 12, downloadUpscale, 1},
		{"download all", menuSeams{tty: true, pickIdx: 4}, 12, downloadAll, 1},
		{"back", menuSeams{tty: true, pickErr: tui.ErrPickBack}, 12, downloadBack, 1},
		{"wrapped back", menuSeams{tty: true, pickErr: fmt.Errorf("x: %w", tui.ErrPickBack)}, 12, downloadBack, 1},
		{"a menu failure plays", menuSeams{tty: true, pickErr: errors.New("tty gone")}, 12, downloadWatch, 1},
		{"a movie cannot download all", menuSeams{tty: true, pickIdx: 4}, 1, downloadWatch, 1},
		{"two episodes can be downloaded together", menuSeams{tty: true, pickIdx: 4}, 2, downloadAll, 1},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := tc.seams
			m.modeName = "Quality"
			m.install(t)
			got := askForDownload(5, tc.total, "Mukashi Banashi")
			assert.Equal(t, tc.want, got)
			assert.Equal(t, tc.picks, m.pickCalls, "the menu opens only with a terminal")
		})
	}
}

func TestAskForDownloadShowsTheRightMenu(t *testing.T) {
	m := &menuSeams{tty: true, pickIdx: 0, modeName: "Quality"}
	m.install(t)
	askForDownload(5, 12, "Mukashi Banashi")

	assert.Equal(t, "Episode 5 > Options", m.gotOpts.Breadcrumb)
	assert.Equal(t, "Mukashi Banashi · Episode 5", m.gotOpts.Context)
	assert.Equal(t, "watch", m.gotOpts.Action)
	assert.Equal(t, "GoAnime - Options", m.gotOpts.WindowTitle)
	require.Len(t, m.gotItems, 5)
	assert.Equal(t, "Quality", m.gotItems[3].Details, "the upscale row shows the active mode")
	assert.Equal(t, "Download all 12 episodes...", m.gotItems[4].Label)
}

// ---- confirmDownloadAll ---------------------------------------------------------------------

func confirmFailures(impl func(int) bool) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	run := func(m *menuSeams, n int) bool {
		oldTTY, oldConfirm := stdinIsTerminal, confirmFn
		defer func() { stdinIsTerminal, confirmFn = oldTTY, oldConfirm }()
		stdinIsTerminal = func() bool { return m.tty }
		confirmFn = func(a, b, c, d string) (bool, error) {
			m.confirmCalls++
			m.gotConfirm = [4]string{a, b, c, d}
			return m.confirmAns, m.confirmErr
		}
		return impl(n)
	}
	yes := &menuSeams{tty: true, confirmAns: true}
	if !run(yes, 12) || yes.gotConfirm[0] != "Download all 12 episodes?" {
		f = append(f, "yes not honoured or title wrong")
	}
	no := &menuSeams{tty: true, confirmAns: false}
	if run(no, 12) {
		f = append(f, "no answered yes")
	}
	failed := &menuSeams{tty: true, confirmAns: true, confirmErr: errors.New("tty gone")}
	if run(failed, 12) {
		f = append(f, "a failed prompt must not confirm")
	}
	headless := &menuSeams{tty: false, confirmAns: false}
	if !run(headless, 12) || headless.confirmCalls != 0 {
		f = append(f, "no terminal: proceeds without asking")
	}
	return f
}

func TestConfirmDownloadAll(t *testing.T) {
	assert.Empty(t, confirmFailures(confirmDownloadAll))

	m := &menuSeams{tty: true, confirmAns: true}
	m.install(t)
	confirmDownloadAll(7)
	assert.Equal(t, [4]string{"Download all 7 episodes?", "This downloads the whole series.", "Yes, download all", "No, back"}, m.gotConfirm)
}

func TestConfirmDownloadAllMeta(t *testing.T) {
	require.Empty(t, confirmFailures(confirmDownloadAll))
	mutants := map[string]func(int) bool{
		"always yes":            func(int) bool { return true },
		"always no":             func(int) bool { return false },
		"an error confirms":     func(n int) bool { return confirmDownloadAll(n) || true },
		"inverted":              func(n int) bool { return !confirmDownloadAll(n) },
		"headless refuses":      func(n int) bool { return stdinIsTerminal() && confirmDownloadAll(n) },
		"wrong count in prompt": func(n int) bool { return confirmDownloadAll(n + 1) },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, confirmFailures(m), "mutant %q survived", name)
	}
}

// ---- askForPlayOffline ----------------------------------------------------------------------

func offlineFailures(impl func() bool) (f []string) {
	run := func(m *menuSeams) bool {
		oldTTY, oldLabels := stdinIsTerminal, pickLabelsFn
		defer func() { stdinIsTerminal, pickLabelsFn = oldTTY, oldLabels }()
		stdinIsTerminal = func() bool { return m.tty }
		pickLabelsFn = func(l []string, o tui.PickOptions) (int, error) {
			m.labelsCalls++
			m.gotLabels = l
			return m.labelsIdx, m.labelsErr
		}
		return impl()
	}
	if m := (&menuSeams{tty: true, labelsIdx: 0}); !run(m) || fmt.Sprint(m.gotLabels) != "[Yes No]" {
		f = append(f, "Yes (row 0) must play offline")
	}
	if m := (&menuSeams{tty: true, labelsIdx: 1}); run(m) {
		f = append(f, "No (row 1) must not")
	}
	if m := (&menuSeams{tty: true, labelsErr: errors.New("x")}); run(m) {
		f = append(f, "a failed menu defaults to no")
	}
	if m := (&menuSeams{tty: false, labelsIdx: 0}); run(m) || m.labelsCalls != 0 {
		f = append(f, "no terminal defaults to no, without asking")
	}
	return f
}

func TestAskForPlayOffline(t *testing.T) {
	assert.Empty(t, offlineFailures(askForPlayOffline))
}

func TestAskForPlayOfflineMeta(t *testing.T) {
	require.Empty(t, offlineFailures(askForPlayOffline))
	mutants := map[string]func() bool{
		"always yes": func() bool { return true },
		"always no":  func() bool { return false },
		"inverted":   func() bool { return !askForPlayOffline() },
		"error means yes": func() bool {
			return askForPlayOffline() || (stdinIsTerminal() && func() bool { _, err := pickLabelsFn(nil, tui.PickOptions{}); return err != nil }())
		},
		"headless means yes": func() bool { return !stdinIsTerminal() || askForPlayOffline() },
	}
	for name, m := range mutants {
		assert.NotEmpty(t, offlineFailures(m), "mutant %q survived", name)
	}
}

// ---- the upscaling menu ---------------------------------------------------------------------

func valuesOf(es []upscaleEntry) []string {
	out := make([]string, len(es))
	for i, e := range es {
		out[i] = e.value
	}
	return out
}

func TestUpscaleMenuEntries(t *testing.T) {
	assert.Equal(t, []string{"back", "off", "setup"}, valuesOf(upscaleMenuEntries(false, false)),
		"without shaders only off and setup are offered")
	assert.Equal(t, []string{"back", "off", "setup"}, valuesOf(upscaleMenuEntries(false, true)),
		"GAN shaders do not matter without the base shaders")

	withoutGAN := valuesOf(upscaleMenuEntries(true, false))
	assert.Equal(t, []string{"back", "off", "performance", "fast", "balanced", "quality", "ultra", "separator",
		"advanced_aa", "advanced_bb", "advanced_ca", "setup_gan", "setup"}, withoutGAN)

	withGAN := valuesOf(upscaleMenuEntries(true, true))
	assert.Equal(t, "gan_uul", withGAN[len(withGAN)-2])
	assert.NotContains(t, withGAN, "setup_gan")
	assert.Equal(t, "setup", withGAN[len(withGAN)-1], "setup stays last")
	assert.Equal(t, "back", withGAN[0], "back stays first")
}

func upscaleEntryFailures(impl func(bool, bool) []upscaleEntry) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	want := map[[2]bool]string{
		{false, false}: "back off setup",
		{false, true}:  "back off setup",
		{true, false}:  "back off performance fast balanced quality ultra separator advanced_aa advanced_bb advanced_ca setup_gan setup",
		{true, true}:   "back off performance fast balanced quality ultra separator advanced_aa advanced_bb advanced_ca gan_uul setup",
	}
	for k, w := range want {
		got := impl(k[0], k[1])
		if strings.Join(valuesOf(got), " ") != w {
			f = append(f, fmt.Sprintf("shaders=%v gan=%v: %v", k[0], k[1], valuesOf(got)))
		}
		for _, e := range got {
			if e.label == "" {
				f = append(f, "an entry has no label: "+e.value)
			}
		}
	}
	return f
}

func TestUpscaleMenuEntriesMeta(t *testing.T) {
	require.Empty(t, upscaleEntryFailures(upscaleMenuEntries))
	mutants := map[string]func(bool, bool) []upscaleEntry{
		"modes offered without shaders": func(_, g bool) []upscaleEntry { return upscaleMenuEntries(true, g) },
		"modes never offered":           func(_, g bool) []upscaleEntry { return upscaleMenuEntries(false, g) },
		"gan row wrong way round":       func(s, g bool) []upscaleEntry { return upscaleMenuEntries(s, !g) },
		"setup dropped": func(s, g bool) []upscaleEntry {
			e := upscaleMenuEntries(s, g)
			return e[:len(e)-1]
		},
		"back dropped": func(s, g bool) []upscaleEntry { return upscaleMenuEntries(s, g)[1:] },
		"off dropped": func(s, g bool) []upscaleEntry {
			e := upscaleMenuEntries(s, g)
			return append(e[:1:1], e[2:]...)
		},
		"separator dropped": func(s, g bool) []upscaleEntry {
			var out []upscaleEntry
			for _, e := range upscaleMenuEntries(s, g) {
				if e.value != "separator" {
					out = append(out, e)
				}
			}
			return out
		},
		"labels blank": func(s, g bool) []upscaleEntry {
			e := upscaleMenuEntries(s, g)
			e[0].label = ""
			return e
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, upscaleEntryFailures(m), "mutant %q survived", name)
	}
}

func TestUpscaleBreadcrumb(t *testing.T) {
	assert.Equal(t, "Upscaling [Balanced]", upscaleBreadcrumb(true, "Balanced"))
	assert.Equal(t, "Upscaling (install shaders first)", upscaleBreadcrumb(false, "Balanced"), "the mode is moot without shaders")
	assert.NotContains(t, upscaleBreadcrumb(true, "Two\nlines"), "\n")
}

func TestApplyUpscaleChoiceSetsEveryMode(t *testing.T) {
	want := map[string]upscaler.ShaderMode{
		"off": upscaler.ShaderModeOff, "performance": upscaler.ShaderModePerformance, "fast": upscaler.ShaderModeFast,
		"balanced": upscaler.ShaderModeBalanced, "quality": upscaler.ShaderModeQuality, "ultra": upscaler.ShaderModeUltra,
		"advanced_aa": upscaler.ShaderModeAdvancedAA, "advanced_bb": upscaler.ShaderModeAdvancedBB,
		"advanced_ca": upscaler.ShaderModeAdvancedCA, "gan_uul": upscaler.ShaderModeGAN_UUL,
	}
	for choice, mode := range want {
		t.Run(choice, func(t *testing.T) {
			m := &menuSeams{}
			m.install(t)
			again, err := applyUpscaleChoice(choice)
			require.NoError(t, err)
			assert.False(t, again)
			assert.Equal(t, []upscaler.ShaderMode{mode}, m.setModes)
			require.Len(t, m.infos, 1, "the user is told what changed")
			assert.NotEmpty(t, m.infos[0])
			assert.Empty(t, m.installs)
			if choice == "gan_uul" {
				require.Len(t, m.warns, 1, "the heavy GAN mode warns")
			} else {
				assert.Empty(t, m.warns)
			}
		})
	}
	assert.Len(t, upscaleModes, len(want), "every selectable mode is covered here")
}

func TestApplyUpscaleChoiceOtherRows(t *testing.T) {
	t.Run("back and unknown change nothing", func(t *testing.T) {
		for _, c := range []string{"back", "no-such-row", ""} {
			m := &menuSeams{}
			m.install(t)
			again, err := applyUpscaleChoice(c)
			require.NoError(t, err)
			assert.False(t, again)
			assert.Empty(t, m.setModes)
			assert.Empty(t, m.installs)
			assert.Empty(t, m.infos)
		}
	})
	t.Run("the separator shows the menu again", func(t *testing.T) {
		m := &menuSeams{}
		m.install(t)
		again, err := applyUpscaleChoice("separator")
		require.NoError(t, err)
		assert.True(t, again)
		assert.Empty(t, m.setModes)
	})
	t.Run("setup installs the shaders", func(t *testing.T) {
		m := &menuSeams{}
		m.install(t)
		_, err := applyUpscaleChoice("setup")
		require.NoError(t, err)
		assert.Equal(t, []string{"shaders"}, m.installs)
		assert.Len(t, m.infos, 2, "announced before and after")
		assert.Empty(t, m.setModes, "installing does not switch modes")
	})
	t.Run("setup failure is reported", func(t *testing.T) {
		boom := errors.New("offline")
		m := &menuSeams{installErr: boom}
		m.install(t)
		again, err := applyUpscaleChoice("setup")
		assert.False(t, again)
		assert.ErrorIs(t, err, boom)
		assert.Contains(t, err.Error(), "failed to install shaders")
		assert.Len(t, m.infos, 1, "no success message after a failure")
	})
	t.Run("setup_gan installs the GAN shaders", func(t *testing.T) {
		m := &menuSeams{}
		m.install(t)
		again, err := applyUpscaleChoice("setup_gan")
		require.NoError(t, err)
		assert.False(t, again)
		assert.Equal(t, []string{"gan"}, m.installs)
	})
	t.Run("setup_gan failure is reported", func(t *testing.T) {
		boom := errors.New("disk full")
		m := &menuSeams{ganErr: boom}
		m.install(t)
		again, err := applyUpscaleChoice("setup_gan")
		assert.False(t, again)
		assert.ErrorIs(t, err, boom)
		assert.Contains(t, err.Error(), "failed to install GAN shaders")
		assert.Len(t, m.infos, 1)
	})
}

func applyFailures(impl func(string) (bool, error)) (f []string) {
	defer func() {
		if r := recover(); r != nil {
			f = append(f, fmt.Sprint("panic: ", r))
		}
	}()
	run := func(m *menuSeams, c string) (bool, error) {
		oldSet, oldInst, oldGAN, oldInfo, oldWarn := setShaderModeFn, installShadersFn, installGANShadersFn, infoFn, warnFn
		defer func() {
			setShaderModeFn, installShadersFn, installGANShadersFn, infoFn, warnFn = oldSet, oldInst, oldGAN, oldInfo, oldWarn
		}()
		setShaderModeFn = func(md upscaler.ShaderMode) { m.setModes = append(m.setModes, md) }
		installShadersFn = func() error { m.installs = append(m.installs, "shaders"); return m.installErr }
		installGANShadersFn = func() error { m.installs = append(m.installs, "gan"); return m.ganErr }
		infoFn = func(msg any, _ ...any) { m.infos = append(m.infos, fmt.Sprint(msg)) }
		warnFn = func(msg any, _ ...any) { m.warns = append(m.warns, fmt.Sprint(msg)) }
		return impl(c)
	}
	if m := (&menuSeams{}); true {
		again, err := run(m, "ultra")
		if again || err != nil || fmt.Sprint(m.setModes) != fmt.Sprint([]upscaler.ShaderMode{upscaler.ShaderModeUltra}) || len(m.infos) != 1 {
			f = append(f, fmt.Sprintf("ultra: %v %v %v", again, err, m.setModes))
		}
	}
	if m := (&menuSeams{}); true {
		run(m, "gan_uul")
		if len(m.warns) != 1 || len(m.setModes) != 1 {
			f = append(f, "gan_uul must set the mode and warn once")
		}
	}
	if m := (&menuSeams{}); true {
		if again, _ := run(m, "separator"); !again || len(m.setModes) != 0 {
			f = append(f, "separator")
		}
	}
	if m := (&menuSeams{}); true {
		if _, err := run(m, "setup"); err != nil || fmt.Sprint(m.installs) != "[shaders]" {
			f = append(f, "setup")
		}
	}
	if m := (&menuSeams{installErr: errors.New("x")}); true {
		if _, err := run(m, "setup"); err == nil {
			f = append(f, "setup failure swallowed")
		}
	}
	if m := (&menuSeams{ganErr: errors.New("x")}); true {
		if _, err := run(m, "setup_gan"); err == nil {
			f = append(f, "setup_gan failure swallowed")
		}
	}
	if m := (&menuSeams{}); true {
		run(m, "back")
		if len(m.setModes)+len(m.installs)+len(m.infos) != 0 {
			f = append(f, "back changed something")
		}
	}
	return f
}

func TestApplyUpscaleChoiceMeta(t *testing.T) {
	require.Empty(t, applyFailures(applyUpscaleChoice))
	mutants := map[string]func(string) (bool, error){
		"separator closes the menu": func(c string) (bool, error) {
			a, e := applyUpscaleChoice(c)
			if c == "separator" {
				return false, e
			}
			return a, e
		},
		"everything re-shows the menu": func(c string) (bool, error) {
			_, e := applyUpscaleChoice(c)
			return true, e
		},
		"install errors swallowed": func(c string) (bool, error) {
			a, e := applyUpscaleChoice(c)
			if strings.HasPrefix(c, "setup") {
				return a, nil
			}
			return a, e
		},
		"gan without the warning": func(c string) (bool, error) {
			if c == "gan_uul" {
				return applyUpscaleChoice("ultra")
			}
			return applyUpscaleChoice(c)
		},
		"nothing is applied": func(c string) (bool, error) {
			return c == "separator", nil
		},
		"setup installs the GAN shaders": func(c string) (bool, error) {
			if c == "setup" {
				return applyUpscaleChoice("setup_gan")
			}
			return applyUpscaleChoice(c)
		},
		"mode off by one": func(c string) (bool, error) {
			if c == "ultra" {
				return applyUpscaleChoice("quality")
			}
			return applyUpscaleChoice(c)
		},
	}
	for name, m := range mutants {
		assert.NotEmpty(t, applyFailures(m), "mutant %q survived", name)
	}
}

// ---- handleUpscaleFromMenu --------------------------------------------------------------------

func TestHandleUpscaleFromMenu(t *testing.T) {
	t.Run("no terminal", func(t *testing.T) {
		m := &menuSeams{tty: false}
		m.install(t)
		err := handleUpscaleFromMenu()
		require.Error(t, err)
		assert.Contains(t, err.Error(), "no TTY")
		assert.Zero(t, m.labelsCalls)
	})
	t.Run("picking a mode applies it", func(t *testing.T) {
		m := &menuSeams{tty: true, shaders: true, modeName: "Off", labelsIdx: 5} // back off performance fast balanced [quality]
		m.install(t)
		require.NoError(t, handleUpscaleFromMenu())
		assert.Equal(t, []upscaler.ShaderMode{upscaler.ShaderModeQuality}, m.setModes)
		assert.Equal(t, "Upscaling [Off]", m.gotLabelOpt.Breadcrumb)
		assert.Equal(t, "GoAnime - Upscale", m.gotLabelOpt.WindowTitle)
	})
	t.Run("without shaders the menu says so", func(t *testing.T) {
		m := &menuSeams{tty: true, shaders: false, labelsIdx: 0}
		m.install(t)
		require.NoError(t, handleUpscaleFromMenu())
		assert.Equal(t, "Upscaling (install shaders first)", m.gotLabelOpt.Breadcrumb)
		assert.Equal(t, []string{"Back", "Off (no upscaling)", "Setup shaders (download)"}, m.gotLabels)
	})
	t.Run("back, cancel and a bad index leave quietly", func(t *testing.T) {
		for _, tc := range []menuSeams{
			{tty: true, shaders: true, labelsErr: tui.ErrPickBack},
			{tty: true, shaders: true, labelsErr: tui.ErrPickCancelled},
			{tty: true, shaders: true, labelsIdx: 99},
			{tty: true, shaders: true, labelsIdx: -1},
			{tty: true, shaders: false, labelsIdx: 3}, // exactly one past the last of three rows
		} {
			m := tc
			m.install(t)
			assert.NoError(t, handleUpscaleFromMenu())
			assert.Empty(t, m.setModes)
		}
	})
	t.Run("a menu failure is reported", func(t *testing.T) {
		boom := errors.New("tty gone")
		m := &menuSeams{tty: true, shaders: true, labelsErr: boom}
		m.install(t)
		err := handleUpscaleFromMenu()
		assert.ErrorIs(t, err, boom)
		assert.Contains(t, err.Error(), "cancelled")
	})
	t.Run("the separator row asks again instead of ending", func(t *testing.T) {
		// shaders on, no GAN: back off performance fast balanced quality ultra [separator]=7 ... then Ultra=6
		m := &menuSeams{tty: true, shaders: true, modeName: "Off", labelsSeq: []int{7, 7, 6}}
		m.install(t)
		require.NoError(t, handleUpscaleFromMenu())
		assert.Equal(t, 3, m.labelsCalls, "two separators, then a real choice")
		assert.Equal(t, []upscaler.ShaderMode{upscaler.ShaderModeUltra}, m.setModes)
	})
	t.Run("installing the shaders ends the menu", func(t *testing.T) {
		m := &menuSeams{tty: true, shaders: false, labelsIdx: 2} // back off [setup]
		m.install(t)
		require.NoError(t, handleUpscaleFromMenu())
		assert.Equal(t, []string{"shaders"}, m.installs)
		assert.Equal(t, 1, m.labelsCalls)
	})
	t.Run("an install failure is returned", func(t *testing.T) {
		boom := errors.New("offline")
		m := &menuSeams{tty: true, shaders: false, labelsIdx: 2, installErr: boom}
		m.install(t)
		assert.ErrorIs(t, handleUpscaleFromMenu(), boom)
	})
}
