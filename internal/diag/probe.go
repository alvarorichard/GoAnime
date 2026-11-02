package diag

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"syscall"
	"time"
)

// Probe is a provider's public homepage to check.
type Probe struct{ Name, URL string }

// Result classifies how a provider answered, in words a person can act on.
type Result struct {
	Name, URL string
	Status    int           // HTTP status, 0 when there was no response
	Elapsed   time.Duration // time to the response headers
	FinalURL  string        // where redirects ended, when different
	Outcome   string        // ok, challenge, blocked, down, slow, dns, refused, tls, error
	Detail    string        // a short reason
}

// ProbeAll checks every probe concurrently (at most four at a time) with a
// per-request timeout and calls report with each result as it arrives. It
// reads only the response headers and never sends credentials.
func ProbeAll(ctx context.Context, client *http.Client, probes []Probe, timeout time.Duration, report func(Result)) {
	if timeout <= 0 {
		timeout = 8 * time.Second
	}
	sem := make(chan struct{}, 4)
	var wg sync.WaitGroup
	var mu sync.Mutex
	for _, p := range probes {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			res := probeOne(ctx, client, p, timeout)
			mu.Lock()
			report(res)
			mu.Unlock()
		}()
	}
	wg.Wait()
}

func probeOne(ctx context.Context, client *http.Client, p Probe, timeout time.Duration) Result {
	res := Result{Name: p.Name, URL: p.URL}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, p.URL, http.NoBody)
	if err != nil {
		res.Outcome, res.Detail = "error", err.Error()
		return res
	}
	req.Header.Set("User-Agent", "GoAnime-diagnostics")
	start := time.Now()
	resp, err := client.Do(req)
	res.Elapsed = time.Since(start).Round(time.Millisecond)
	if err != nil {
		res.Outcome, res.Detail = classifyError(err)
		return res
	}
	defer func() { _ = resp.Body.Close() }()
	res.Status = resp.StatusCode
	if resp.Request != nil && resp.Request.URL != nil {
		if final := resp.Request.URL.String(); !sameOrigin(final, p.URL) {
			res.FinalURL = final
		}
	}
	res.Outcome, res.Detail = classifyResponse(resp)
	return res
}

func classifyResponse(resp *http.Response) (outcome, detail string) {
	cloudflare := strings.Contains(strings.ToLower(resp.Header.Get("Server")), "cloudflare")
	switch {
	case resp.Header.Get("Cf-Mitigated") != "":
		return "challenge", "Cloudflare challenge (" + resp.Header.Get("Cf-Mitigated") + ")"
	case cloudflare && (resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusServiceUnavailable):
		return "challenge", "Cloudflare answered HTTP " + http.StatusText(resp.StatusCode)
	case resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusTooManyRequests:
		return "blocked", http.StatusText(resp.StatusCode)
	case resp.StatusCode >= 500:
		return "down", http.StatusText(resp.StatusCode)
	case resp.StatusCode >= 200 && resp.StatusCode < 400:
		return "ok", ""
	}
	return "error", http.StatusText(resp.StatusCode)
}

func classifyError(err error) (outcome, detail string) {
	var dnsErr *net.DNSError
	var tlsErr *tls.CertificateVerificationError
	var netErr net.Error
	switch {
	case errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &netErr) && netErr.Timeout()):
		return "slow", "no answer before the timeout"
	case errors.As(err, &dnsErr):
		return "dns", "the name does not resolve"
	case errors.Is(err, syscall.ECONNREFUSED):
		return "refused", "connection refused"
	case errors.As(err, &tlsErr):
		return "tls", "certificate not accepted"
	}
	return "error", shortErr(err)
}

func shortErr(err error) string {
	var u *url.Error
	if errors.As(err, &u) {
		err = u.Err
	}
	s := err.Error()
	if len(s) > 120 {
		s = s[:120] + "…"
	}
	return s
}

func sameOrigin(a, b string) bool {
	ua, errA := url.Parse(a)
	ub, errB := url.Parse(b)
	if errA != nil || errB != nil {
		return true
	}
	return strings.EqualFold(ua.Host, ub.Host)
}
