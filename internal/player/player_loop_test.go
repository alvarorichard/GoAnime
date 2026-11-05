package player

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/discord"
	"github.com/alvarorichard/Goanime/v2/internal/models"
)

var errLoopBoom = errors.New("loop boom")

// loopScript scripts one run of the player menu loop and records what the
// loop did, in order, so a case can assert the exact sequence.
type loopScript struct {
	choices   []string // one per menu call; the last is repeated if the loop outlives it
	menuErr   error    // returned by the first menu call instead of a choice
	pingFails int      // the Nth ping (1-based) fails; 0 = never
	next      error
	previous  error
	selectEp  error

	events []string
	menuIn [][2]string
}

func (s *loopScript) loop() playerLoop {
	pings, menus := 0, 0
	rec := func(f string, a ...any) { s.events = append(s.events, fmt.Sprintf(f, a...)) }
	return playerLoop{
		send: func(_ string, cmd []any) (any, error) {
			if cmd[0] == "quit" {
				rec("quit")
				return nil, nil
			}
			pings++
			rec("ping")
			if s.pingFails > 0 && pings == s.pingFails {
				return nil, errors.New("connection refused")
			}
			return 1.0, nil
		},
		menu: func(_, anime string, ep int) (string, error) {
			rec("menu")
			s.menuIn = append(s.menuIn, [2]string{anime, fmt.Sprint(ep)})
			if s.menuErr != nil {
				return "", s.menuErr
			}
			c := s.choices[min(menus, len(s.choices)-1)]
			menus++
			return c, nil
		},
		next: func(i int, _ []models.Episode, _, _ int, _ *discord.RichPresenceUpdater, _ chan struct{}, _ string) error {
			rec("next:%d", i)
			return s.next
		},
		previous: func(i int, _ []models.Episode, _, _ int, _ *discord.RichPresenceUpdater, _ chan struct{}, _ string) error {
			rec("previous:%d", i)
			return s.previous
		},
		selectEp: func(_ []models.Episode, _, _ int, _ *discord.RichPresenceUpdater, _ chan struct{}, _ string) error {
			rec("select")
			return s.selectEp
		},
		audio:    func(string) { rec("audio") },
		subtitle: func(string) { rec("subtitle") },
		skip: func(_ string, ep *models.Episode) {
			rec("skip:%s", ep.Number)
		},
	}
}

type loopCase struct {
	name   string
	script loopScript
	want   error // compared with errors.Is; nil means the loop must return nil
	events string
}

func loopCases() []loopCase {
	return []loopCase{
		{"dead mpv returns to the download options without a menu",
			loopScript{choices: []string{"quit"}, pingFails: 1}, ErrBackToDownloadOptions, "ping"},
		{"a menu error quits mpv and goes back",
			loopScript{menuErr: errLoopBoom}, ErrBackToDownloadOptions, "ping,menu,quit"},
		{"download options quits mpv and goes back",
			loopScript{choices: []string{"download_options"}}, ErrBackToDownloadOptions, "ping,menu,quit"},
		{"quit quits mpv and ends the session",
			loopScript{choices: []string{"quit"}}, ErrUserQuit, "ping,menu,quit"},
		{"change quits mpv and asks for another anime",
			loopScript{choices: []string{"change"}}, ErrChangeAnime, "ping,menu,quit"},
		{"next hands over the following index and ends on success",
			loopScript{choices: []string{"next"}}, nil, "ping,menu,next:3"},
		{"next at a boundary stays in the menu",
			loopScript{choices: []string{"next", "quit"}, next: errStayInPlayerMenu}, ErrUserQuit,
			"ping,menu,next:3,ping,menu,quit"},
		{"next failing ends the loop with that error",
			loopScript{choices: []string{"next", "quit"}, next: errLoopBoom}, errLoopBoom, "ping,menu,next:3"},
		{"previous hands over the preceding index and ends on success",
			loopScript{choices: []string{"previous"}}, nil, "ping,menu,previous:1"},
		{"previous at a boundary stays in the menu",
			loopScript{choices: []string{"previous", "download_options"}, previous: errStayInPlayerMenu}, ErrBackToDownloadOptions,
			"ping,menu,previous:1,ping,menu,quit"},
		{"previous failing ends the loop with that error",
			loopScript{choices: []string{"previous", "quit"}, previous: errLoopBoom}, errLoopBoom, "ping,menu,previous:1"},
		{"select ends on success",
			loopScript{choices: []string{"select"}}, nil, "ping,menu,select"},
		{"select cancelled stays in the menu",
			loopScript{choices: []string{"select", "quit"}, selectEp: errStayInPlayerMenu}, ErrUserQuit,
			"ping,menu,select,ping,menu,quit"},
		{"select failing ends the loop with that error",
			loopScript{choices: []string{"select", "quit"}, selectEp: errLoopBoom}, errLoopBoom, "ping,menu,select"},
		{"audio subtitle and skip keep the loop going",
			loopScript{choices: []string{"audio", "subtitle", "skip", "quit"}}, ErrUserQuit,
			"ping,menu,audio,ping,menu,subtitle,ping,menu,skip:7,ping,menu,quit"},
		{"an unknown choice is ignored",
			loopScript{choices: []string{"nonsense", "quit"}}, ErrUserQuit, "ping,menu,ping,menu,quit"},
		{"mpv dying while in the menu ends the loop without another quit",
			loopScript{choices: []string{"audio", "quit"}, pingFails: 2}, ErrBackToDownloadOptions,
			"ping,menu,audio,ping"},
	}
}

// loopFailures runs every case through run and reports each violation.
func loopFailures(run func(pl playerLoop, index int, ep *models.Episode) error) []string {
	var bad []string
	for _, c := range loopCases() {
		s := c.script
		ep := &models.Episode{Number: "7"}
		err := run(s.loop(), 2, ep)
		if got := strings.Join(s.events, ","); got != c.events {
			bad = append(bad, fmt.Sprintf("%s: events %q, want %q", c.name, got, c.events))
		}
		if c.want == nil && err != nil || c.want != nil && !errors.Is(err, c.want) {
			bad = append(bad, fmt.Sprintf("%s: error %v, want %v", c.name, err, c.want))
		}
	}
	return bad
}

func realRun(pl playerLoop, index int, ep *models.Episode) error {
	return runPlayerMenu(pl, "sock", nil, index, 5, 1, 2, nil, nil, ep)
}

func TestRunPlayerMenu(t *testing.T) {
	setTestLogger(t)
	if bad := loopFailures(realRun); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

func TestRunPlayerMenu_PassesNameAndEpisodeToTheMenu(t *testing.T) {
	setTestLogger(t)
	s := loopScript{choices: []string{"audio", "quit"}}
	SetAnimeName("Frieren", 1)
	_ = runPlayerMenu(s.loop(), "sock", nil, 0, 9, 0, 0, nil, nil, &models.Episode{})
	if len(s.menuIn) != 2 {
		t.Fatalf("menu called %d times, want 2", len(s.menuIn))
	}
	for _, in := range s.menuIn {
		if in != [2]string{"Frieren", "9"} {
			t.Errorf("menu got %v, want [Frieren 9]", in)
		}
	}
}

func TestPlayerAnimeName(t *testing.T) {
	SetAnimeName("Stored", 1)
	if got := playerAnimeName(nil); got != "Stored" {
		t.Errorf("no updater: got %q, want Stored", got)
	}
	mk := func(a *models.Anime) *discord.RichPresenceUpdater {
		paused := false
		var mu sync.Mutex
		return discord.NewRichPresenceUpdater(a, &paused, &mu, time.Second, 0, "", MpvSendCommand)
	}
	if got := playerAnimeName(mk(&models.Anime{Name: "Live"})); got != "Live" {
		t.Errorf("updater with a name: got %q, want Live", got)
	}
	if got := playerAnimeName(mk(&models.Anime{})); got != "Stored" {
		t.Errorf("updater with empty name: got %q, want Stored", got)
	}
	if got := playerAnimeName(mk(nil)); got != "Stored" {
		t.Errorf("updater without anime: got %q, want Stored", got)
	}
}

func TestRealPlayerLoop_WiresEveryHandler(t *testing.T) {
	pl := realPlayerLoop()
	if pl.send == nil || pl.menu == nil || pl.next == nil || pl.previous == nil ||
		pl.selectEp == nil || pl.audio == nil || pl.subtitle == nil || pl.skip == nil {
		t.Fatalf("a handler is nil: %+v", pl)
	}
}

// The menu offers exactly the values the loop acts on, so none is dead.
func TestRunPlayerMenu_HandlesEveryMenuValue(t *testing.T) {
	setTestLogger(t)
	handled := map[string]bool{}
	for _, c := range loopCases() {
		for _, ch := range c.script.choices {
			handled[ch] = true
		}
	}
	for _, withTracks := range []bool{false, true} {
		for _, movie := range []bool{false, true} {
			for _, o := range playerMenuOptions(movie, withTracks, withTracks) {
				if !handled[o.Value] {
					t.Errorf("menu value %q has no loop case", o.Value)
				}
			}
		}
	}
}

func TestMeta_RunPlayerMenu_CatchesBrokenLoops(t *testing.T) {
	setTestLogger(t)
	wrap := func(f func(pl playerLoop) playerLoop) func(pl playerLoop, i int, ep *models.Episode) error {
		return func(pl playerLoop, i int, ep *models.Episode) error { return realRun(f(pl), i, ep) }
	}
	mutants := map[string]func(pl playerLoop, i int, ep *models.Episode) error{
		"next gets the wrong index": func(pl playerLoop, i int, ep *models.Episode) error {
			return realRun(pl, i+1, ep)
		},
		"never pings (stale mpv shows the menu)": wrap(func(pl playerLoop) playerLoop {
			orig := pl.send
			pl.send = func(s string, c []any) (any, error) {
				if c[0] == "get_property" {
					return 1.0, nil
				}
				return orig(s, c)
			}
			return pl
		}),
		"quit forgets to stop mpv": wrap(func(pl playerLoop) playerLoop {
			orig := pl.send
			pl.send = func(s string, c []any) (any, error) {
				if c[0] == "quit" {
					return nil, nil
				}
				return orig(s, c)
			}
			return pl
		}),
		"swallows the next error": wrap(func(pl playerLoop) playerLoop {
			orig := pl.next
			pl.next = func(i int, e []models.Episode, m, a int, u *discord.RichPresenceUpdater, st chan struct{}, s string) error {
				if err := orig(i, e, m, a, u, st, s); errors.Is(err, errLoopBoom) {
					return errStayInPlayerMenu
				}
				return nil
			}
			return pl
		}),
		"treats stay as a result": wrap(func(pl playerLoop) playerLoop {
			orig := pl.selectEp
			pl.selectEp = func(e []models.Episode, m, a int, u *discord.RichPresenceUpdater, st chan struct{}, s string) error {
				if err := orig(e, m, a, u, st, s); errors.Is(err, errStayInPlayerMenu) {
					return nil
				}
				return nil
			}
			return pl
		}),
		"audio and subtitle swapped": wrap(func(pl playerLoop) playerLoop {
			pl.audio, pl.subtitle = pl.subtitle, pl.audio
			return pl
		}),
		"skip gets no episode": func(pl playerLoop, i int, _ *models.Episode) error {
			return realRun(pl, i, &models.Episode{Number: "?"})
		},
		"menu error is fatal": wrap(func(pl playerLoop) playerLoop {
			orig := pl.menu
			pl.menu = func(s, a string, e int) (string, error) {
				c, err := orig(s, a, e)
				if err != nil { //nolint:nilerr // deliberate mutant: swallows the menu error
					return "quit", nil
				}
				return c, nil
			}
			return pl
		}),
	}
	for name, run := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(loopFailures(run)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}
