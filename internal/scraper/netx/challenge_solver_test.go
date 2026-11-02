package netx

import (
	"context"
	"errors"
	"testing"
	"time"
)

// challengeOnlySolver is the Goyabu-grade capability: a full challenge page,
// but no standalone Turnstile widget.
type challengeOnlySolver struct{}

func (challengeOnlySolver) SolveChallenge(context.Context, string, time.Duration, RevealPolicy) (*ChallengeSolveResult, error) {
	return &ChallengeSolveResult{}, nil
}

// widgetSolver answers both, and records the request so the test can see the
// page address reach it unchanged.
type widgetSolver struct {
	challengeOnlySolver
	got   TurnstileRequest
	token string
	err   error
}

func (s *widgetSolver) SolveTurnstile(_ context.Context, r TurnstileRequest, _ time.Duration) (string, error) {
	s.got = r
	return s.token, s.err
}

func resetChallengeSolver(t *testing.T) {
	t.Helper()
	RegisterChallengeSolver(nil)
	t.Cleanup(func() { RegisterChallengeSolver(nil) })
}

func TestTurnstileSolverAvailableNeedsTheWidgetCapability(t *testing.T) {
	resetChallengeSolver(t)

	if TurnstileSolverAvailable() {
		t.Fatal("no solver registered, yet the widget capability reports available")
	}
	if _, ok, err := SolveTurnstileFor(context.Background(), TurnstileRequest{PageURL: "https://x.test/", SiteKey: "k"}, time.Second); ok || err != nil {
		t.Fatalf("no solver: ok=%v err=%v; want a plain cannot-help", ok, err)
	}

	RegisterChallengeSolver(challengeOnlySolver{})
	if TurnstileSolverAvailable() {
		t.Fatal("a page-only solver cannot answer a widget")
	}
	if _, ok, err := SolveTurnstileFor(context.Background(), TurnstileRequest{PageURL: "https://x.test/", SiteKey: "k"}, time.Second); ok || err != nil {
		t.Fatalf("page-only solver: ok=%v err=%v; want a plain cannot-help", ok, err)
	}
}

func TestSolveTurnstileForPassesTheRequestToTheRegisteredSolver(t *testing.T) {
	resetChallengeSolver(t)
	want := TurnstileRequest{
		PageURL: "https://player.test/filme/7?tsv=old",
		SiteKey: "0xKEY",
		Action:  "embed",
		CData:   "cd",
	}
	s := &widgetSolver{token: "0.token"}
	RegisterChallengeSolver(s)

	if !TurnstileSolverAvailable() {
		t.Fatal("a widget solver is registered but the capability reports unavailable")
	}
	token, ok, err := SolveTurnstileFor(context.Background(), want, 10*time.Second)
	if err != nil || !ok || token != "0.token" {
		t.Fatalf("token=%q ok=%v err=%v", token, ok, err)
	}
	if s.got != want {
		t.Fatalf("request = %+v, want %+v", s.got, want)
	}
}

func TestSolveTurnstileForRelaysTheSolverError(t *testing.T) {
	resetChallengeSolver(t)
	boom := errors.New("boom")
	RegisterChallengeSolver(&widgetSolver{err: boom})

	token, ok, err := SolveTurnstileFor(context.Background(), TurnstileRequest{PageURL: "https://x.test/", SiteKey: "k"}, time.Second)
	if !ok {
		t.Fatal("the solver was registered and called; ok must be true")
	}
	if !errors.Is(err, boom) || token != "" {
		t.Fatalf("token=%q err=%v", token, err)
	}
}
