package player

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

type confirmCall struct{ title, desc, yes, no string }

type resumeDialogCase struct {
	name       string
	terminal   bool
	answer     bool
	confirmErr error
	movie      bool
	wantResume bool
	wantErr    string // substring; empty = no error
	wantAsked  bool
}

func resumeDialogCases() []resumeDialogCase {
	return []resumeDialogCase{
		{name: "no terminal fails without asking", terminal: false, answer: true, wantErr: "no terminal"},
		{name: "yes resumes", terminal: true, answer: true, wantResume: true, wantAsked: true},
		{name: "no starts over", terminal: true, answer: false, wantAsked: true},
		{name: "a prompt error is wrapped and never resumes", terminal: true, answer: true, confirmErr: errors.New("tty gone"), wantErr: "error showing dialog", wantAsked: true},
		{name: "movies are asked too", terminal: true, answer: true, movie: true, wantResume: true, wantAsked: true},
	}
}

type askFn func(has func() bool, confirm func(t, d, y, n string) (bool, error), ep, sec int, movie bool) (bool, error)

func resumeDialogFailures(fn askFn) []string {
	var bad []string
	for _, c := range resumeDialogCases() {
		var calls []confirmCall
		confirm := func(t, d, y, n string) (bool, error) {
			calls = append(calls, confirmCall{t, d, y, n})
			return c.answer, c.confirmErr
		}
		got, err := fn(func() bool { return c.terminal }, confirm, 5, 95, c.movie)
		if got != c.wantResume {
			bad = append(bad, fmt.Sprintf("%s: resume=%v, want %v", c.name, got, c.wantResume))
		}
		if (c.wantErr == "") != (err == nil) || (err != nil && !strings.Contains(err.Error(), c.wantErr)) {
			bad = append(bad, fmt.Sprintf("%s: error %v, want %q", c.name, err, c.wantErr))
		}
		if asked := len(calls) > 0; asked != c.wantAsked {
			bad = append(bad, fmt.Sprintf("%s: asked=%v, want %v", c.name, asked, c.wantAsked))
		}
		if c.wantAsked && len(calls) == 1 {
			want := confirmCall{
				resumeTitle(5, 95, c.movie),
				"You can continue watching from where you left off.",
				"Yes, resume", "No, start from beginning",
			}
			if calls[0] != want {
				bad = append(bad, fmt.Sprintf("%s: asked %+v, want %+v", c.name, calls[0], want))
			}
		}
	}
	return bad
}

func TestAskToResume(t *testing.T) {
	if bad := resumeDialogFailures(askToResume); len(bad) > 0 {
		t.Fatal(strings.Join(bad, "\n"))
	}
}

func TestAskToResume_MovieAndEpisodeTitlesDiffer(t *testing.T) {
	if resumeTitle(5, 95, true) == resumeTitle(5, 95, false) {
		t.Fatal("the movie and episode titles must differ, or the isMovie argument is untested")
	}
}

func TestShowResumeDialog_UsesTheConfirmSeam(t *testing.T) {
	prev := confirmFn
	t.Cleanup(func() { confirmFn = prev })
	asked := 0
	confirmFn = func(string, string, string, string) (bool, error) { asked++; return true, nil }
	// Without a terminal (the CI case) the seam must not even be reached.
	got, err := showResumeDialog(1, 10, false)
	if err == nil && (!got || asked != 1) {
		t.Fatalf("with a terminal: got %v asked %d", got, asked)
	}
	if err != nil && asked != 0 {
		t.Fatalf("no terminal but the prompt ran %d times", asked)
	}
}

func TestMeta_AskToResume_CatchesBrokenImplementations(t *testing.T) {
	mutants := map[string]askFn{
		"skips the terminal check": func(_ func() bool, c func(t, d, y, n string) (bool, error), e, s int, m bool) (bool, error) {
			return askToResume(func() bool { return true }, c, e, s, m)
		},
		"inverts the answer": func(h func() bool, c func(t, d, y, n string) (bool, error), e, s int, m bool) (bool, error) {
			r, err := askToResume(h, c, e, s, m)
			if err != nil {
				return r, err
			}
			return !r, nil
		},
		"always resumes": func(h func() bool, c func(t, d, y, n string) (bool, error), e, s int, m bool) (bool, error) {
			_, err := askToResume(h, c, e, s, m)
			return err == nil, err
		},
		"swallows the prompt error": func(h func() bool, c func(t, d, y, n string) (bool, error), e, s int, m bool) (bool, error) {
			r, err := askToResume(h, c, e, s, m)
			if err != nil && strings.Contains(err.Error(), "error showing dialog") {
				return false, nil
			}
			return r, err
		},
		"wrong movie flag": func(h func() bool, c func(t, d, y, n string) (bool, error), e, s int, m bool) (bool, error) {
			return askToResume(h, c, e, s, !m)
		},
		"wrong episode": func(h func() bool, c func(t, d, y, n string) (bool, error), e, s int, m bool) (bool, error) {
			return askToResume(h, c, e+1, s, m)
		},
		"wrong time": func(h func() bool, c func(t, d, y, n string) (bool, error), e, s int, m bool) (bool, error) {
			return askToResume(h, c, e, s+3600, m)
		},
		"swapped buttons": func(h func() bool, c func(t, d, y, n string) (bool, error), e, s int, m bool) (bool, error) {
			return askToResume(h, func(t, d, y, n string) (bool, error) { return c(t, d, n, y) }, e, s, m)
		},
	}
	for name, fn := range mutants {
		t.Run(name, func(t *testing.T) {
			if len(resumeDialogFailures(fn)) == 0 {
				t.Fatalf("mutant %q survived", name)
			}
		})
	}
}
