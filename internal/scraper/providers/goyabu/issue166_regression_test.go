// Package scraper — End-to-end regression suite for issue #166.
//
// Issue #166 (reported 2026-04-28): Goyabu episodes were silently failing
// because netx.CheckChallengeDocument's body-text scan included a bare
// strings.Contains(body, "cloudflare") check. Any legitimate Goyabu page
// (footer credit, ToS link, user comment) mentioning the word "cloudflare"
// was misclassified as a Cloudflare challenge, causing GetEpisodeStreamURL
// to return netx.ErrSourceUnavailable before the strategy chain ever ran. The
// resulting "no video source found" / "no valid video URL found" error was
// blind — no log line indicated the false-positive cause.
//
// This file drives the full GetEpisodeStreamURL via httptest with
// realistically shaped Goyabu HTML, covering the regression and every
// real-challenge signal we want to keep firing.
package goyabu

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
)

// realGoyabuEpisodePage returns HTML that mirrors the structure of a real
// Goyabu episode page: WP layout, valid playersData JS, and a footer that
// includes the literal word "cloudflare" — the exact shape that triggered
// the false positive in issue #166.
func realGoyabuEpisodePage(streamURL string) string {
	return fmt.Sprintf(`<!doctype html>
<html lang="pt-BR">
<head>
<meta charset="UTF-8">
<title>Black Clover Dublado — Episódio 5 | Goyabu</title>
<meta name="description" content="Assista Black Clover Dublado Episódio 5 online em HD pelo Goyabu.">
</head>
<body class="single single-post">
<header><nav>Início | Animes | Filmes</nav></header>
<main>
  <article class="episode">
    <h1>Black Clover Dublado Episódio 5</h1>
    <div class="player-area">
      <script type="text/javascript">
        var playersData = [{"blogger_token":"AD6v5dyTESTTOKEN","url":%q}];
      </script>
    </div>
    <section class="episode-list">
      <a href="/?p=37475">Episódio 1</a>
      <a href="/?p=37476">Episódio 2</a>
    </section>
    <section class="comments">
      <p>Usuário 1: muito bom esse episódio</p>
      <p>Usuário 2: alguém sabe quando sai o próximo?</p>
    </section>
  </article>
</main>
<footer>
  <small>© 2026 Goyabu — Todos os direitos reservados.</small>
  <p>Site protegido contra ataques. Entregue via Cloudflare CDN para máxima velocidade.</p>
  <p><a href="/termos">Termos de uso</a> | <a href="/privacidade">Política de privacidade</a></p>
</footer>
</body>
</html>`, streamURL)
}

// goyabuChallengePage returns an HTML body that should trigger
// netx.CheckChallengeDocument via the named signal. The signal name is one of
// the keys defined inside the function and lets a single helper drive
// every challenge variant from the table-driven test below.
func goyabuChallengePage(signal string) string {
	switch signal {
	case "title:just-a-moment":
		return `<html><head><title>Just a moment...</title></head>
<body><p>Verifying you are human. This may take a few seconds.</p></body></html>`
	case "title:attention-required":
		return `<html><head><title>Attention Required! | Cloudflare</title></head>
<body><p>Sorry, you have been blocked.</p></body></html>`
	case "title:access-denied":
		return `<html><head><title>Access denied</title></head><body>403</body></html>`
	case "marker:cf-wrapper":
		return `<html><head><title>x</title></head>
<body><div id="cf-wrapper"><p>Please wait while we check your browser.</p></div></body></html>`
	case "marker:challenge-form":
		return `<html><head><title>x</title></head>
<body><form id="challenge-form" action="/cdn-cgi/l/chk_jschl"></form></body></html>`
	case "marker:cf-error-details":
		return `<html><head><title>x</title></head><body><div id="cf-error-details">Error 1020</div></body></html>`
	case "marker:cf-browser-verification":
		return `<html><head><title>x</title></head>
<body><div class="cf-browser-verification">Checking…</div></body></html>`
	case "phrase:cf-error":
		return `<html><head><title>x</title></head><body><p>cf-error code 1020</p></body></html>`
	case "phrase:checking-your-browser":
		return `<html><head><title>x</title></head>
<body><p>Checking your browser before accessing the site.</p></body></html>`
	case "phrase:ddos-protection":
		return `<html><head><title>x</title></head><body><p>DDoS protection by Cloudflare</p></body></html>`
	case "phrase:enable-cookies":
		return `<html><head><title>x</title></head><body><p>Please enable cookies and reload the page.</p></body></html>`
	case "phrase:performance-security":
		return `<html><head><title>x</title></head><body><p>Performance & security by Cloudflare</p></body></html>`
	case "phrase:ray-id":
		return `<html><head><title>x</title></head><body><p>Ray ID: 7abc123def456789</p></body></html>`
	case "phrase:complete-security-check":
		return `<html><head><title>x</title></head><body><p>Please complete the security check to access goyabu.io.</p></body></html>`
	case "phrase:attention-required-banner":
		return `<html><head><title>x</title></head><body><p>Attention Required! | Cloudflare</p></body></html>`
	}
	panic("unknown signal: " + signal)
}

// TestIssue166_LegitimateGoyabuPageWithCloudflareMentionStillResolves
// is the regression that would have caught issue #166: a real Goyabu
// episode page that happens to mention "cloudflare" in a non-challenge
// context (footer, ToS, comment) must NOT short-circuit the strategy
// chain. GetEpisodeStreamURL should reach Strategy 4 and return the
// Blogger embed URL.
func TestIssue166_LegitimateGoyabuPageWithCloudflareMentionStillResolves(t *testing.T) {
	t.Parallel()

	// Capture the served stream URL so we can assert it round-trips.
	var streamURL string

	mux := http.NewServeMux()
	mux.HandleFunc("/episode", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		_, _ = fmt.Fprint(w, realGoyabuEpisodePage(streamURL))
	})
	// Strategy 2 will POST here; we deliberately fail the AJAX so the flow
	// reaches Strategy 4 (Blogger embed URL fallback). This mirrors what
	// happens for the user when Goyabu's WP AJAX path is rate-limited but
	// the embed URL is still usable.
	mux.HandleFunc("/wp-admin/admin-ajax.php", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	// playersData[0].url must be a valid absolute http(s) URL because
	// netx.ValidateStreamURL guards on scheme + host. Use the test server's
	// origin so the round-trip succeeds without contacting the internet.
	streamURL = srv.URL + "/blogger-embed?token=AD6v5dyTESTTOKEN"

	client := NewGoyabuClient()
	client.baseURL = srv.URL
	client.maxRetries = 0
	client.retryDelay = 0

	got, err := client.GetEpisodeStreamURL(srv.URL + "/episode")
	if err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
	if got != streamURL {
		t.Fatalf("expected stream URL %q, got %q", streamURL, got)
	}
	if errors.Is(err, netx.ErrSourceUnavailable) {
		t.Fatalf("must NOT be classified as netx.ErrSourceUnavailable")
	}
}

// TestIssue166_AllRealChallengeSignalsStillTrigger guarantees that
// relaxing the body-text scan in errors.go did not regress the actual
// challenge detection. Every signal that should fire still fires.
func TestIssue166_AllRealChallengeSignalsStillTrigger(t *testing.T) {
	t.Parallel()

	signals := []string{
		"title:just-a-moment",
		"title:attention-required",
		"title:access-denied",
		"marker:cf-wrapper",
		"marker:challenge-form",
		"marker:cf-error-details",
		"marker:cf-browser-verification",
		"phrase:cf-error",
		"phrase:checking-your-browser",
		"phrase:ddos-protection",
		"phrase:enable-cookies",
		"phrase:performance-security",
		"phrase:ray-id",
		"phrase:complete-security-check",
		"phrase:attention-required-banner",
	}

	for _, sig := range signals {
		sig := sig
		t.Run(sig, func(t *testing.T) {
			t.Parallel()

			html := goyabuChallengePage(sig)
			mux := http.NewServeMux()
			mux.HandleFunc("/episode", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=UTF-8")
				_, _ = fmt.Fprint(w, html)
			})
			srv := httptest.NewServer(mux)
			defer srv.Close()

			client := NewGoyabuClient()
			client.baseURL = srv.URL
			client.maxRetries = 0
			client.retryDelay = 0

			_, err := client.GetEpisodeStreamURL(srv.URL + "/episode")
			if err == nil {
				t.Fatalf("expected challenge error for signal %q, got nil", sig)
			}
			if !errors.Is(err, netx.ErrSourceUnavailable) {
				t.Fatalf("signal %q: expected netx.ErrSourceUnavailable in chain, got %v", sig, err)
			}
		})
	}
}

// TestIssue166_NoFalsePositiveForCommonContentMentions hardens the
// allowlist against a representative set of phrases that real anime
// sites have included in normal page chrome. Each must NOT trigger.
func TestIssue166_NoFalsePositiveForCommonContentMentions(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		body string
	}{
		{
			name: "footer mentioning cloudflare cdn",
			body: `<p>Site servido via Cloudflare CDN para máxima velocidade.</p>`,
		},
		{
			name: "uppercase CLOUDFLARE in slogan",
			body: `<p>POWERED BY CLOUDFLARE WORKERS</p>`,
		},
		{
			name: "user comment mentioning cloudflare incident",
			body: `<p>O site caiu ontem por causa do incidente do Cloudflare.</p>`,
		},
		{
			name: "ToS link mentioning cloudflare",
			body: `<a href="/tos">Hospedado com Cloudflare. Veja a política de privacidade.</a>`,
		},
		{
			name: "credit mentioning cloudflare workers",
			body: `<small>Powered by Cloudflare Workers — © 2026</small>`,
		},
		{
			name: "navigation mentioning cloudflare turnstile only as a label",
			body: `<button>Resolver Cloudflare Turnstile</button>`,
		},
		{
			name: "comment with the word 'cloudflare' but not a challenge",
			body: `<p>cloudflare é uma cdn famosa</p>`,
		},
		{
			name: "title 'cloud' near 'flare' but not 'cloudflare'",
			body: `<p>O cloud da flare é diferente.</p>`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			html := fmt.Sprintf(`<!doctype html><html><head><title>Goyabu — Anime</title></head>
<body><main>conteúdo válido</main><footer>%s</footer></body></html>`, tc.body)

			mux := http.NewServeMux()
			mux.HandleFunc("/episode", func(w http.ResponseWriter, _ *http.Request) {
				w.Header().Set("Content-Type", "text/html; charset=UTF-8")
				_, _ = fmt.Fprint(w, html)
			})
			srv := httptest.NewServer(mux)
			defer srv.Close()

			client := NewGoyabuClient()
			client.baseURL = srv.URL
			client.maxRetries = 0
			client.retryDelay = 0

			// No playersData is intentional — we want to assert the
			// challenge check did not fire. The downstream strategies
			// will fail with a different error, but it must not be
			// netx.ErrSourceUnavailable.
			_, err := client.GetEpisodeStreamURL(srv.URL + "/episode")
			if errors.Is(err, netx.ErrSourceUnavailable) {
				t.Fatalf("legitimate page %q was misclassified as a challenge: %v", tc.name, err)
			}
		})
	}
}

// TestIssue166_DocumentSerializesSourceLabel proves the `source` field is
// not lost in the error chain. Operators use this label in dashboards to
// decide which scraper to disable when challenges spike.
func TestIssue166_DocumentSerializesSourceLabel(t *testing.T) {
	t.Parallel()

	html := goyabuChallengePage("title:just-a-moment")
	mux := http.NewServeMux()
	mux.HandleFunc("/episode", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=UTF-8")
		_, _ = fmt.Fprint(w, html)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewGoyabuClient()
	client.baseURL = srv.URL
	client.maxRetries = 0
	client.retryDelay = 0

	_, err := client.GetEpisodeStreamURL(srv.URL + "/episode")
	if err == nil {
		t.Fatalf("expected challenge error, got nil")
	}
	if !strings.Contains(err.Error(), "Goyabu") {
		t.Fatalf("error does not mention source label 'Goyabu': %v", err)
	}
}
