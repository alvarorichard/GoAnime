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
// profile.go) and unpacked on first use. Two variables override this:
//
//	GOANIME_CFSOLVE_BIN      path to the cfsolve "challenge" binary
//	GOANIME_CFSOLVE_PROFILE  path to a cfsolve browser profile (.json)
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
	envBin     = "GOANIME_CFSOLVE_BIN"
	envProfile = "GOANIME_CFSOLVE_PROFILE"

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
	profile := os.Getenv(envProfile)
	if profile == "" {
		var err error
		if profile, err = defaultProfilePath(); err != nil {
			util.Debug("cfsolve: cannot prepare the embedded profile", "err", err)
			return nil, false
		}
	}
	if _, err := os.Stat(profile); err != nil {
		util.Debug("cfsolve: profile not found", "profile", profile)
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
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("cfsolve: bad target %q", targetURL)
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
	cmd := exec.CommandContext(ctx, s.Bin, "--page", targetURL, filepath.Join(tmp, "run"), //nolint:gosec // Bin is operator configuration, args are not shell-interpreted
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
			Domain: u.Hostname(), Path: "/", Secure: u.Scheme == "https", HttpOnly: true,
		}},
		UserAgent: ua,
		FinalURL:  targetURL,
	}, nil
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
