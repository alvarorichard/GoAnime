# Source Diagnostics Test Plan

This checklist helps separate provider downtime from GoAnime bugs.

## Goal

- `SourceUnavailable`: 521, 522, 523, 524, 530, DNS errors, connection timeouts, or origin outages should skip the health check.
- `BlockedOrChallenge`: 403, 429, 1020, captcha, or challenge pages should skip the health check.
- `ParserBroken`: 200 OK responses without expected selectors, JSON, or results should fail the health check.
- `DecryptBroken`: decrypt/API responses with invalid payloads should fail the health check.
- `DownloadExpired`: extracted CDN links returning 403/404 should be diagnosed as expired links.
- `InternalBug`: panics, nil pointers, infinite loops, or local logic errors should fail.

## Local Commands

Run these commands before opening or updating the PR:

```powershell
go test ./internal/scraper -count=1 -v
go test ./internal/player -count=1 -v
go test -tags sourcehealth -run TestSourceHealthLive -count=1 -v ./internal/scraper
$env:CI='true'; go test ./... -count=1
go vet ./...
golangci-lint run --timeout=15m
gosec ./...
govulncheck ./...
git diff --check
```

## Live Health Check

`TestSourceHealthLive` runs a known search for each provider:

- Anime/general providers: `naruto`
- Movie/TV providers: `dexter`

Expected behavior:

- Source offline, Cloudflare 521/522/523/524/530, DNS, or timeout: `t.Skip`.
- Captcha, challenge, 403/429/1020: `t.Skip`.
- 200 OK with a broken parser or zero results for a known query: `t.Fatal`.
- Broken decrypt or internal app error: `t.Fatal`.

## App Logs

Expected messages:

- `FlixHQ temporarily unavailable: Cloudflare 521/origin down`
- `SFlix blocked the request: captcha/challenge`
- `Goyabu responded, but the parser did not find the expected data`
- `Download link expired or was denied: HTTP 404`

After 3 consecutive origin/block failures, the circuit breaker skips the source for 10 minutes to avoid hammering a dead provider.

## Discord

The project already has local Discord Rich Presence, but that is not the same as project health alerts. To publish diagnostics to a Discord channel safely, use a separate PR with:

- `DISCORD_WEBHOOK_URL` configured as a GitHub secret.
- A scheduled or manual job that runs `go test -tags sourcehealth -run TestSourceHealthLive -count=1 -v ./internal/scraper`.
- A step that sends only a summary of `healthy`, `skipped`, and `failed` sources, without exposing tokens, cookies, or private URLs.

Without that secret configured, the safe option is to keep diagnostics in CI logs and local debug output.
