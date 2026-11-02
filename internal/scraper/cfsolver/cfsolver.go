// Package cfsolver clears Cloudflare challenge pages through the external
// cfsolve tool (github.com/chris124567/cfsolve), which solves them from plain
// requests with no browser.
//
// cfsolve is a Rust program, so GoAnime does not link it: it runs the "challenge"
// binary when one is installed and registers itself as the process-wide
// netx.ChallengeSolver only then. Without the binary nothing is registered and
// gated sources keep reporting the challenge as blocked.
//
// Nothing needs configuring. The binary is looked for next to the goanime
// executable, in a "cfsolve" folder there, in GoAnime's config folder and on
// PATH. The browser profile it impersonates is embedded in GoAnime (see
// profile.go) and unpacked on first use. GOANIME_CFSOLVE_BIN points at a
// "challenge" binary kept somewhere else.
package cfsolver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

const (
	envBin = "GOANIME_CFSOLVE_BIN"

	clearanceName = "cf_clearance"
)

// Solver runs the cfsolve binary at Bin with the browser profile at Profile.
type Solver struct {
	Bin     string
	Profile string
}

// binaryNames are the file names cfsolve's binary may carry.
var binaryNames = []string{"challenge", "cfsolve"}

// binaryDirs lists where a bundled cfsolve is looked for, most specific first.
func binaryDirs() []string {
	var dirs []string
	if exe, err := os.Executable(); err == nil {
		d := filepath.Dir(exe)
		dirs = append(dirs, d, filepath.Join(d, "cfsolve"))
	}
	if cfg, err := os.UserConfigDir(); err == nil {
		dirs = append(dirs, filepath.Join(cfg, "goanime", "cfsolve"))
	}
	return dirs
}

func findBinary() string {
	if bin := os.Getenv(envBin); bin != "" {
		return bin
	}
	for _, d := range binaryDirs() {
		for _, n := range binaryNames {
			for _, ext := range []string{"", ".exe"} {
				p := filepath.Join(d, n+ext)
				if st, err := os.Stat(p); err == nil && !st.IsDir() {
					return p
				}
			}
		}
	}
	for _, n := range binaryNames {
		if p, err := exec.LookPath(n); err == nil {
			return p
		}
	}
	return ""
}

// Discover builds a Solver from the environment, the usual install folders and
// PATH, or reports false when the binary cannot be found or has no profile.
func Discover() (*Solver, bool) {
	bin := findBinary()
	if bin == "" {
		return nil, false
	}
	if _, err := os.Stat(bin); err != nil {
		util.Debug("cfsolve: binary not found", "bin", bin)
		return nil, false
	}
	// The profile is always GoAnime's own embedded copy, unpacked by us.
	profile, err := defaultProfilePath()
	if err != nil {
		util.Debug("cfsolve: cannot prepare the embedded profile", "err", err)
		return nil, false
	}
	return &Solver{Bin: bin, Profile: profile}, true
}

// Register installs a discovered Solver as the process-wide challenge solver.
// It does nothing, and reports false, when cfsolve is not installed.
func Register() bool {
	s, ok := Discover()
	if !ok {
		return false
	}
	netx.RegisterChallengeSolver(s)
	return true
}

// userAgent reads the User-Agent the profile impersonates. Cloudflare binds
// cf_clearance to it, so the caller has to replay with exactly this value.
func (s *Solver) userAgent() (string, error) {
	raw, err := os.ReadFile(s.Profile)
	if err != nil {
		return "", err
	}
	var p struct {
		Transport struct {
			Headers map[string]string `json:"headers"`
		} `json:"transport"`
	}
	if err := json.Unmarshal(raw, &p); err != nil {
		return "", fmt.Errorf("cfsolve profile: %w", err)
	}
	for k, v := range p.Transport.Headers {
		if strings.EqualFold(k, "user-agent") && v != "" {
			return v, nil
		}
	}
	return "", errors.New("cfsolve profile has no User-Agent")
}

// SolveChallenge implements netx.ChallengeSolver. There is no window to reveal,
// so reveal is ignored.
func (s *Solver) SolveChallenge(ctx context.Context, targetURL string, timeout time.Duration, _ netx.RevealPolicy) (*netx.ChallengeSolveResult, error) {
	u, err := url.Parse(targetURL)
	if err != nil || u.Host == "" || u.Scheme != "https" {
		return nil, fmt.Errorf("cfsolve: target must be an https URL, got %q", targetURL)
	}
	ua, err := s.userAgent()
	if err != nil {
		return nil, err
	}

	// cfsolve wants an empty or missing directory for its request log.
	tmp, err := os.MkdirTemp("", "goanime-cfsolve-")
	if err != nil {
		return nil, err
	}
	defer func() { _ = os.RemoveAll(tmp) }()

	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	cmd := exec.CommandContext(ctx, s.Bin, "--page", targetURL, filepath.Join(tmp, "run"), // #nosec G204 -- Bin is the operator-configured tool; no shell; targetURL is a parsed https URL
		"--solve", "--fingerprint", s.Profile)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("cfsolve: %w: %s", err, lastLine(stderr.String()))
	}

	value, ok := parseClearance(stdout.String())
	if !ok {
		return nil, fmt.Errorf("cfsolve: no %s in output: %q", clearanceName, lastLine(stdout.String()))
	}
	return &netx.ChallengeSolveResult{
		Cookies: []*http.Cookie{{
			Name: clearanceName, Value: value,
			// Secure and HttpOnly as Cloudflare sets them. net/http's cookie jar
			// ignores SameSite, so Lax is the safe value and changes nothing.
			Domain: u.Hostname(), Path: "/", Secure: true, HttpOnly: true, SameSite: http.SameSiteLaxMode,
		}},
		UserAgent: ua,
		FinalURL:  targetURL,
	}, nil
}

// TurnstileRequest names a Turnstile widget a page renders from script.
type TurnstileRequest struct {
	PageURL string // the address the widget belongs to; the page is not fetched
	SiteKey string
	Action  string // optional
	CData   string // optional
}

// SolveTurnstile returns a token for one widget without loading its page, so
// the caller keeps the only session that fetched it (a site may bind its
// own challenge to that session's address and User-Agent).
func (s *Solver) SolveTurnstile(ctx context.Context, r TurnstileRequest, timeout time.Duration) (string, error) {
	if r.SiteKey == "" {
		return "", errors.New("cfsolve: a sitekey is required")
	}
	u, err := url.Parse(r.PageURL)
	if err != nil || u.Host == "" || u.Scheme != "https" {
		return "", fmt.Errorf("cfsolve: page must be an https URL, got %q", r.PageURL)
	}
	// These come out of a page the site controls and cfsolve parses its argv with
	// options, so a value shaped like one must never reach it.
	for name, v := range map[string]string{"sitekey": r.SiteKey, "action": r.Action, "cdata": r.CData} {
		if err := checkArg(name, v); err != nil {
			return "", err
		}
	}
	tmp, err := os.MkdirTemp("", "goanime-cfsolve-")
	if err != nil {
		return "", err
	}
	defer func() { _ = os.RemoveAll(tmp) }()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	args := []string{"--page", r.PageURL, filepath.Join(tmp, "run"), "--solve", "--blank-page",
		"--fingerprint", s.Profile, "--sitekey", r.SiteKey}
	if r.Action != "" {
		args = append(args, "--action", r.Action)
	}
	if r.CData != "" {
		args = append(args, "--cdata", r.CData)
	}
	cmd := exec.CommandContext(ctx, s.Bin, args...) // #nosec G204 -- Bin is the operator-configured tool; no shell; every value was checked by checkArg or parsed as an https URL
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cfsolve: %w: %s", err, lastLine(stderr.String()))
	}
	token := lastLine(stdout.String())
	if token == "" || strings.HasPrefix(token, clearanceName+"=") {
		return "", fmt.Errorf("cfsolve: no token in output: %q", token)
	}
	return token, nil
}

// checkArg rejects a value cfsolve could read as an option or that cannot be an
// argument at all.
func checkArg(name, v string) error {
	if strings.HasPrefix(v, "-") || strings.ContainsRune(v, 0) {
		return fmt.Errorf("cfsolve: %s %q cannot be passed as an argument", name, v)
	}
	return nil
}

// parseClearance finds the "cf_clearance=<value>" line cfsolve prints on success.
func parseClearance(out string) (string, bool) {
	for _, line := range strings.Split(out, "\n") {
		line = strings.TrimSpace(line)
		if v, ok := strings.CutPrefix(line, clearanceName+"="); ok && v != "" {
			return v, true
		}
	}
	return "", false
}

func lastLine(s string) string {
	s = strings.TrimSpace(s)
	if i := strings.LastIndexByte(s, '\n'); i >= 0 {
		s = s[i+1:]
	}
	return s
}
