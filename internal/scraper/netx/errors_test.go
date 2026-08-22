package netx

import (
	"errors"
	"strings"
	"testing"

	"github.com/PuerkitoBio/goquery"
)

// Regression for issue #166 (2026-04-28): a real Goyabu episode page that
// merely mentions the word "cloudflare" in footer text, terms-of-service
// links, or user comments was being misclassified as a Cloudflare challenge,
// short-circuiting GetEpisodeStreamURL before any extraction strategy ran.
// CheckChallengeDocument must only fire on phrases that are unique to actual
// challenge interstitials.
func TestCheckChallengeDocument_DoesNotFalsePositiveOnCloudflareMention(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		html string
	}{
		{
			name: "footer credit mentioning cloudflare",
			html: `<!doctype html>
<html><head><title>Black Clover Episódio 5 — Goyabu</title></head>
<body>
<div id="player-area">
<script>var playersData = {"0":{"player":"https://www.blogger.com/video.g?token=abc"}};</script>
</div>
<footer>Site protegido. Entregue via Cloudflare CDN. © 2026 Goyabu.</footer>
</body></html>`,
		},
		{
			name: "user comment mentioning cloudflare",
			html: `<!doctype html>
<html><head><title>One Piece 1102 — Goyabu</title></head>
<body>
<article>Episode content here</article>
<section class="comments">
<p>Usuário: o site usa cloudflare? tava lento ontem</p>
</section>
</body></html>`,
		},
		{
			name: "terms-of-service link mentioning cloudflare",
			html: `<!doctype html>
<html><head><title>Naruto 25 — Goyabu</title></head>
<body><main>video player</main>
<nav><a href="/tos">Termos</a> — Hospedado com Cloudflare. Veja nossa política.</nav>
</body></html>`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatalf("parse html: %v", err)
			}

			if err := CheckChallengeDocument(doc, "test-source"); err != nil {
				t.Fatalf("expected nil for legitimate page mentioning 'cloudflare', got %v", err)
			}
		})
	}
}

// TestCheckChallengeDocument_EveryAllowlistedPhraseTriggers iterates the
// production challengeBodyPhrases slice. Every phrase must independently
// fire — this guards against accidental removal: if someone deletes a
// phrase from errors.go, the corresponding subtest fails immediately.
func TestCheckChallengeDocument_EveryAllowlistedPhraseTriggers(t *testing.T) {
	t.Parallel()

	if len(challengeBodyPhrases) == 0 {
		t.Fatalf("challengeBodyPhrases is empty — challenge detection has no phrase coverage")
	}

	for _, phrase := range challengeBodyPhrases {
		t.Run(phrase, func(t *testing.T) {
			t.Parallel()

			html := `<html><head><title>x</title></head><body><p>` + phrase + `</p></body></html>`
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				t.Fatalf("parse html: %v", err)
			}

			err = CheckChallengeDocument(doc, "test-source")
			if err == nil {
				t.Fatalf("phrase %q must trigger a challenge, got nil", phrase)
			}
			if !errors.Is(err, ErrSourceUnavailable) {
				t.Fatalf("phrase %q: expected ErrSourceUnavailable, got %v", phrase, err)
			}
		})
	}
}

// TestCheckChallengeDocument_PhraseMatchIsCaseInsensitive locks down that
// the body scan lowercases input. Cloudflare challenge pages occasionally
// vary capitalization across versions; matching must not depend on it.
func TestCheckChallengeDocument_PhraseMatchIsCaseInsensitive(t *testing.T) {
	t.Parallel()

	for _, phrase := range challengeBodyPhrases {
		t.Run(phrase, func(t *testing.T) {
			t.Parallel()

			upper := strings.ToUpper(phrase)
			html := `<html><head><title>x</title></head><body><p>` + upper + `</p></body></html>`
			doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
			if err != nil {
				t.Fatalf("parse html: %v", err)
			}

			if err := CheckChallengeDocument(doc, "test-source"); err == nil {
				t.Fatalf("uppercase phrase %q must trigger, got nil", upper)
			}
		})
	}
}

// TestCheckChallengeDocument_PhraseInAttributeDoesNotTrigger guarantees
// the scan only inspects rendered text. A phrase that appears only in an
// HTML attribute (data-*, alt, title, aria-label, href) must not fire,
// because doc.Text() — what the body scan operates over — does not
// include attribute values. Without this guard, a footer link like
// `<a href="https://challenge-form.example.com/...">` would create a new
// false-positive class.
func TestCheckChallengeDocument_PhraseInAttributeDoesNotTrigger(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		html string
	}{
		{
			name: "phrase only in href",
			html: `<html><head><title>Goyabu</title></head>
<body><main>conteúdo</main>
<a href="/redirect?to=ddos%20protection%20by%20cloudflare">link</a>
</body></html>`,
		},
		{
			name: "phrase only in data attribute",
			html: `<html><head><title>Goyabu</title></head>
<body><div data-marker="cf-error">conteúdo válido</div></body></html>`,
		},
		{
			name: "phrase only in alt text",
			html: `<html><head><title>Goyabu</title></head>
<body><img alt="ray id: graphic" src="/x.png"><p>episode</p></body></html>`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatalf("parse html: %v", err)
			}

			if err := CheckChallengeDocument(doc, "test-source"); err != nil {
				t.Fatalf("attribute-only phrase must not trigger; got %v", err)
			}
		})
	}
}

// TestCheckChallengeDocument_EmptyAndMinimalDocs covers degenerate input.
// The check must be defensive: nothing in, nothing out — but no panics.
func TestCheckChallengeDocument_EmptyAndMinimalDocs(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		html string
	}{
		{name: "empty body", html: `<html><head><title>x</title></head><body></body></html>`},
		{name: "no body", html: `<html><head><title>x</title></head></html>`},
		{name: "no title", html: `<html><body><p>hello</p></body></html>`},
		{name: "completely empty doc", html: ``},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatalf("parse html: %v", err)
			}

			if err := CheckChallengeDocument(doc, "test-source"); err != nil {
				t.Fatalf("minimal document must not trigger, got %v", err)
			}
		})
	}
}

// TestCheckChallengeDocument_IsIdempotent guards against future state-
// carrying refactors. Calling the function twice on the same document
// must yield the same outcome. If a regression introduces a one-shot
// regex precompile cache or trims the document, this catches it.
func TestCheckChallengeDocument_IsIdempotent(t *testing.T) {
	t.Parallel()

	html := `<html><head><title>Just a moment...</title></head><body>v</body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}

	first := CheckChallengeDocument(doc, "src-a")
	second := CheckChallengeDocument(doc, "src-a")

	if (first == nil) != (second == nil) {
		t.Fatalf("non-idempotent: first=%v second=%v", first, second)
	}
	if first != nil && first.Error() != second.Error() {
		t.Fatalf("error message changed across invocations:\n first=%v\nsecond=%v", first, second)
	}
}

// TestCheckChallengeDocument_PropagatesSourceLabel locks down that the
// reported source label is visible in the error string. Operators rely
// on this to pinpoint which scraper is being throttled when challenges
// spike across the fleet.
func TestCheckChallengeDocument_PropagatesSourceLabel(t *testing.T) {
	t.Parallel()

	html := `<html><head><title>Just a moment...</title></head><body>v</body></html>`
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}

	for _, label := range []string{"Goyabu episode page", "AnimeFire search", "AnimeFire episodes"} {
		err := CheckChallengeDocument(doc, label)
		if err == nil {
			t.Fatalf("expected challenge error, got nil")
		}
		if !strings.Contains(err.Error(), label) {
			t.Fatalf("error message must contain source label %q, got %v", label, err)
		}
	}
}

// Companion test asserting the positive cases still trigger so the relaxed
// matcher does not become permissive.
func TestCheckChallengeDocument_DetectsRealChallenges(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name string
		html string
	}{
		{
			name: "just-a-moment title",
			html: `<html><head><title>Just a moment...</title></head><body>verifying</body></html>`,
		},
		{
			name: "attention-required title",
			html: `<html><head><title>Attention Required! | Cloudflare</title></head><body>blocked</body></html>`,
		},
		{
			name: "cf-wrapper element",
			html: `<html><head><title>x</title></head><body><div id="cf-wrapper">challenge</div></body></html>`,
		},
		{
			name: "challenge-form element",
			html: `<html><head><title>x</title></head><body><form id="challenge-form"></form></body></html>`,
		},
		{
			name: "checking-your-browser body phrase",
			html: `<html><head><title>x</title></head><body><p>Checking your browser before accessing the site.</p></body></html>`,
		},
		{
			name: "ray-id body phrase",
			html: `<html><head><title>x</title></head><body><p>Ray ID: 1234abcd567</p></body></html>`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			doc, err := goquery.NewDocumentFromReader(strings.NewReader(tc.html))
			if err != nil {
				t.Fatalf("parse html: %v", err)
			}

			err = CheckChallengeDocument(doc, "test-source")
			if err == nil {
				t.Fatalf("expected challenge error, got nil")
			}
			if !errors.Is(err, ErrSourceUnavailable) {
				t.Fatalf("expected ErrSourceUnavailable in chain, got %v", err)
			}
		})
	}
}
