package pomfy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"
)

// Meta-tests: tests about the mocks themselves. A fake that drifts from the
// production shapes it doubles for hides real failures behind passing green
// rows, so these assert the RECORDS — the exact answers captured from
// pomfy.my, api.pomfy.stream and the byse CDN on 2026-11-02 — still parse,
// still route, and still refuse what the real server refuses. When any of
// these fail, the fixture has drifted from production and must be recaptured
// (or the client beside it updated) before anything else can be trusted.

// TestMetaCatalogFixtureParsesAndKeys: the recorded catalogo-ids blob must
// produce the id map the client builds from it, with both flagship ids left
// and no id double-booked as movie and series.
func TestMetaCatalogFixtureParsesAndKeys(t *testing.T) {
	var parsed struct {
		Filmes []int64 `json:"filmes"`
		Series []int64 `json:"series"`
	}
	if err := json.Unmarshal([]byte(recordedCatalog), &parsed); err != nil {
		t.Fatalf("recordedCatalog no longer parses: %v", err)
	}
	var filmSet = map[int64]bool{}
	for _, id := range parsed.Filmes {
		filmSet[id] = true
	}
	var saw550, saw1396 bool
	for _, id := range parsed.Series {
		if id == 1396 {
			saw1396 = true
		}
		if filmSet[id] {
			t.Errorf("catalog fixture lists id %d as both movie and series", id)
		}
	}
	saw550 = filmSet[550]
	if !saw550 || !saw1396 {
		t.Fatal("the fixtures' flagship ids (550 movie, 1396 series) must stay listed")
	}
}

// TestMetaWatchSourcesShape: the recorded watch-sources reply carries one
// byse source with an /e/<code> embed plus a flyfile decoy, the exact
// structure the client-side filter drops from.
func TestMetaWatchSourcesShape(t *testing.T) {
	var parsed watchSources
	if err := json.Unmarshal([]byte(recordedWatchSources), &parsed); err != nil {
		t.Fatalf("recordedWatchSources no longer parses: %v", err)
	}
	if parsed.Mode != "direct" {
		t.Errorf("mode = %q, want direct", parsed.Mode)
	}
	if len(parsed.Sources) != 2 {
		t.Fatalf("sources = %d, want both servers as recorded", len(parsed.Sources))
	}
	if parsed.Sources[0].Kind != "byse" || parsed.Sources[1].Kind != "flyfile" {
		t.Errorf("kinds = %q, %q — the recorded order must hold", parsed.Sources[0].Kind, parsed.Sources[1].Kind)
	}
	if !strings.Contains(parsed.Sources[0].EmbedURL, "/e/") {
		t.Errorf("byse embed = %q, must be an /e/<code> URL", parsed.Sources[0].EmbedURL)
	}
	if ByseCode(parsed.Sources[0].EmbedURL) == "" {
		t.Error("the recorded embed must still yield a code")
	}
	if parsed.Sources[0].Label != "Servidor 1" || parsed.Sources[1].Label != "Servidor 2" {
		t.Errorf("labels = %q, %q — server naming is part of the contract",
			parsed.Sources[0].Label, parsed.Sources[1].Label)
	}
}

// TestMetaMovieTitleFixtureIsDevalue: the assistir payload's node 1 keys
// (id/tipo/title/…) must point into itself the devalue way, and the values
// at those indexes must be the litteral recorded strings.
func TestMetaMovieTitleFixtureIsDevalue(t *testing.T) {
	var parsed struct {
		Nodes []struct {
			Data json.RawMessage `json:"data"`
		} `json:"nodes"`
	}
	if err := json.Unmarshal([]byte(movieTitleFixture), &parsed); err != nil {
		t.Fatalf("movieTitleFixture no longer parses: %v", err)
	}
	if len(parsed.Nodes) != 2 {
		t.Fatalf("nodes = %d, want the layout assistir carries", len(parsed.Nodes))
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(parsed.Nodes[1].Data, &arr); err != nil {
		t.Fatal("node 1 must be a flat devalue array")
	}
	var head map[string]int64
	if json.Unmarshal(arr[0], &head) != nil {
		t.Fatal("entry 0 of the data array must be the id-table")
	}
	if head["tipo"] != 2 || head["title"] != 3 || head["watch"] != 11 {
		t.Errorf("id-table = %+v, the recorded pointer layout must hold", head)
	}
}

// TestMetaFakeByseCDNJudgeLikeProduction drives the fake CDN with raw
// requests the way the browser does, OUTSIDE resolveByse: the verify
// endpoint declines wrong solutions, hands an ok-token for right ones, and
// the playback endpoint only answers bearing that token — the behaviors the
// live network showed on 2026-11-02.
func TestMetaFakeByseCDNJudgeLikeProduction(t *testing.T) {
	cdn := newFakeByseCDN(t, "metaref", 6, sealFixture(t, []byte(`{"sources":[],"tracks":[]}`)))
	base := cdn.srv.URL + "/api/videos/cta12345/embed"

	post := func(path string, payload []byte, token string) (int, string, http.Header) {
		req, err := http.NewRequest(http.MethodPost, base+path, bytes.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("X-Embed-Origin", DefaultBaseHost)
		req.Header.Set("X-Embed-Parent", cdn.srv.URL+"/e/cta12345")
		req.Header.Set("Content-Type", "application/json")
		if token != "" {
			req.Header.Set("X-Captcha-Token", token)
		}
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return resp.StatusCode, buf.String(), resp.Header
	}

	// Wrong solution: forbidden, the recorded rejection shape.
	wrongBody := mustJSON(t, map[string]any{
		"pow_token":   "tok-1",
		"solution":    "0",
		"fingerprint": map[string]any{"token": "t", "viewer_id": "v", "device_id": "d", "confidence": 0.99},
	})
	code, body, _ := post("/captcha/verify", wrongBody, "")
	if code != http.StatusForbidden || !strings.Contains(body, "rejected") {
		t.Fatalf("wrong-solution verify = %d %q, want 403 rejected", code, body)
	}

	// Right solution: ok-token.
	solution, err := solveNonce(context.Background(), "metaref"+":", 6, 5*time.Second)
	if err != nil {
		t.Fatalf("the meta solve failed: %v", err)
	}
	okBody := mustJSON(t, map[string]any{
		"pow_token":   "tok-1",
		"solution":    solution,
		"fingerprint": map[string]any{"token": "meta", "viewer_id": "m", "device_id": "m", "confidence": 0.99},
	})
	code, verifyAnswer, _ := post("/captcha/verify", okBody, "")
	if code != http.StatusOK || !strings.Contains(verifyAnswer, `"status":"ok"`) {
		t.Fatalf("ok-solution verify = %d %q, want 200 ok", code, verifyAnswer)
	}
	var captcha captchaVerify
	if json.Unmarshal([]byte(verifyAnswer), &captcha) == nil {
		if !strings.HasPrefix(captcha.Token, "captcha-ok-") {
			t.Errorf("issued captcha token = %q, want the fake's ok- namespace", captcha.Token)
		}
	} else {
		t.Fatal("verify answer must be the {status,token} JSON")
	}

	// Playback without the issued token: refused before the seal is touched.
	codeNoTok, _, _ := post("/playback", []byte(`{"fingerprint":{"token":"meta"}}`), "")
	if codeNoTok != http.StatusForbidden {
		t.Errorf("playback with no captcha token = %d, want 403", codeNoTok)
	}
	codeTok, sealAnswer, _ := post("/playback", []byte(`{"fingerprint":{"token":"meta"}}`), "captcha-ok-meta")
	if codeTok != http.StatusOK || !strings.Contains(sealAnswer, `"playback"`) {
		t.Errorf("playback with issued token = %d %q, want 200 seal", codeTok, short(sealAnswer))
	}
}

// short shortens seal payloads for failure messages.
func short(s string) string {
	if len(s) > 90 {
		return s[:90] + "…"
	}
	return s
}

// TestMetaFakeCatalogSiteRouteContracts: a fake that answers anything for a
// broken query is worse than no fake — assert the negative paths the site
// holds (missing params, /serie vs /assistir routes) once more, directly.
func TestMetaFakeCatalogSiteRouteContracts(t *testing.T) {
	site := newFakeCatalogSite(t)
	base := site.srv.URL

	get := func(path string) (*http.Response, string) {
		resp, err := http.Get(base + path)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = resp.Body.Close() }()
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(resp.Body)
		return resp, buf.String()
	}

	// the catalog route answers the recorded blob.
	_, body := get("/api/catalogo-ids")
	if body != recordedCatalog {
		t.Errorf("catalog body = %q, want the recorded blob", body)
	}
	// watch-sources without id/tipo is a 400.
	resp, _ := get("/api/watch-sources")
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("watch-sources without params = %d, want 400", resp.StatusCode)
	}
	// assistir and serie payload routes exist only behind /__data.json.
	if code, _ := statusCodeOf(base + "/assistir/550"); code != http.StatusNotFound {
		t.Errorf("bare assistir page = %d, want 404 (the fake only serves the payload)", code)
	}
	if code, _ := statusCodeOf(base + "/assistir/550/__data.json"); code != http.StatusOK {
		t.Errorf("assistir payload = %d, want 200", code)
	}
	if code, _ := statusCodeOf(base + "/serie/1396/__data.json"); code != http.StatusOK {
		t.Errorf("serie payload = %d, want 200", code)
	}
}

func statusCodeOf(url string) (int, string) {
	resp, err := http.Get(url)
	if err != nil {
		return 0, ""
	}
	defer func() { _ = resp.Body.Close() }()
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	return resp.StatusCode, buf.String()
}

// TestMetaFakeTMDBContract guards the TMDB fake's judgments: off-pool keys
// meet 401, over-quota keys meet 429, unknown queries meet empty pages —
// the same answers production gave us.
func TestMetaFakeTMDBContract(t *testing.T) {
	tmdb := newFakeTMDB(t, "known", fakeMedia(1))
	base := tmdb.srv.URL + "/3/search/multi?query=known&page=1&api_key=%s"

	if code := oneStatus(base, "0000ffff0000ffff0000ffff0000fffe"); code != http.StatusUnauthorized {
		t.Errorf("off-pool key = %d, want 401", code)
	}
	tmdb.quota[PublicKeyForTesting(0)] = true
	if code := oneStatus(base, PublicKeyForTesting(0)); code != http.StatusTooManyRequests {
		t.Errorf("quota key = %d, want 429", code)
	}
	tmdb.quota[PublicKeyForTesting(1)] = false
	if code, body := oneStatusBody(base, PublicKeyForTesting(1)); code != http.StatusOK || !strings.Contains(body, `"results":[`) {
		t.Errorf("valid key = %d %q, want 200 with results", code, body)
	}
}

func oneStatus(base, key string) (n int) {
	resp, err := http.Get(fmt.Sprintf(base, key))
	if err != nil {
		return 0
	}
	_ = resp.Body.Close()
	return resp.StatusCode
}

func oneStatusBody(base, key string) (int, string) {
	resp, err := http.Get(fmt.Sprintf(base, key))
	if err != nil {
		return 0, ""
	}
	defer func() { _ = resp.Body.Close() }()
	buf := new(bytes.Buffer)
	_, _ = buf.ReadFrom(resp.Body)
	return resp.StatusCode, buf.String()
}
