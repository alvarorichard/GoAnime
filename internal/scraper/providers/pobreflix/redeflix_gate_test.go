package pobreflix

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"testing"
)

const redeflixGatePage = `<html><form id="rf-gate-form" method="post" action="/filme/603">` +
	`<input type="hidden" name="rf_gate_ticket" value="v1.abc"><div id="rf-turnstile"></div></form>` +
	`<script>turnstile.render(el,{sitekey:"0xKEY",action:"playback_gate",cdata:"cd1"})</script></html>`

func TestParseRedeflixGate(t *testing.T) {
	if tk, ok := parseRedeflixGate([]byte(redeflixGatePage)); !ok || tk != "v1.abc" {
		t.Fatalf("gate not parsed: %q %v", tk, ok)
	}
	if _, ok := parseRedeflixGate([]byte(`<html>var playbackResolveTicket = "x"</html>`)); ok {
		t.Fatal("ticket page mistaken for a gate")
	}
	sk, action, cdata, ok := parseTurnstile([]byte(redeflixGatePage))
	if !ok || sk != "0xKEY" || action != "playback_gate" || cdata != "cd1" {
		t.Fatalf("widget: %q %q %q %v", sk, action, cdata, ok)
	}
}

func TestResolveRedeflixGateWithoutSolver(t *testing.T) {
	clearChallengeSolver(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(redeflixGatePage))
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(srv.Client(), srv.URL, filepath.Join(t.TempDir(), "cache.json"))
	_, err := c.ResolveRedeflix(context.Background(), srv.URL+"/filme/603", AudioDubbed)
	if !errors.Is(err, ErrPlayerSolverMissing) {
		t.Fatalf("want ErrPlayerSolverMissing, got %v", err)
	}
}

func TestResolveRedeflixPassesGate(t *testing.T) {
	solver := &fakeWidgetSolver{}
	registerWidgetSolver(t, solver)
	var posted url.Values
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/playback-resolve.php":
			_, _ = w.Write([]byte(`{"url":"https://cdn.example/v.m3u8"}`))
		case r.Method == http.MethodPost:
			_ = r.ParseForm()
			posted = r.PostForm
			_, _ = w.Write([]byte(`<script>var playbackResolveTicket = "tk1";</script>`))
		default:
			_, _ = w.Write([]byte(redeflixGatePage))
		}
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(srv.Client(), srv.URL, filepath.Join(t.TempDir(), "cache.json"))
	st, err := c.ResolveRedeflix(context.Background(), srv.URL+"/filme/603", AudioDubbed)
	if err != nil || st.URL != "https://cdn.example/v.m3u8" {
		t.Fatalf("stream %+v err %v", st, err)
	}
	if posted.Get("rf_gate_ticket") != "v1.abc" || posted.Get("cf-turnstile-response") != testToken {
		t.Fatalf("form posted: %v", posted)
	}
	if r, _ := solver.lastRequest(); r.SiteKey != "0xKEY" || r.Action != "playback_gate" || r.CData != "cd1" {
		t.Fatalf("solver request: %+v", r)
	}
}
