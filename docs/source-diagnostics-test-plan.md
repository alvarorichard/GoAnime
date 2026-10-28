# Source diagnostics test plan

This plan helps tell a source that is down apart from a bug in GoAnime.

## Diagnostic kinds

Every source failure is classified into one of these kinds, defined in
`internal/scraper/netx/source_diagnostic.go`:

- `SourceUnavailable`: HTTP 521, 522, 523, 524 or 530, a DNS error, a
  connection timeout or an origin that is down. The health check skips it.
- `BlockedOrChallenge`: HTTP 403, 429 or 1020, a captcha or a challenge. The
  health check skips it.
- `ParserBroken`: a 200 OK response without the expected selectors, JSON or
  results. The health check fails.
- `DecryptBroken`: a decrypt or API layer returned data in an invalid format.
  The health check fails.
- `DownloadExpired`: an extracted CDN link returned 403 or 404, and is
  reported as an expired link.
- `InternalBug`: a panic, a nil pointer, an infinite loop or another local
  error. The health check fails.

## Local commands

Run these before opening or updating a pull request:

```bash
go test ./internal/scraper -count=1 -v
go test ./internal/player -count=1 -v
go test -tags sourcehealth -run TestSourceHealthLive -count=1 -v ./internal/scraper
CI=true go test ./... -count=1
go vet ./...
golangci-lint run --timeout=15m
gosec ./...
govulncheck ./...
git diff --check
```

In PowerShell, set the variable with `$env:CI='true'` before the
`go test ./...` line.

## Live health check

`TestSourceHealthLive` runs a known search against every source:

- Anime sources: `naruto`
- Movie and TV sources: `dexter`

Expected outcome:

- Source offline, Cloudflare 521, 522, 523, 524 or 530, DNS error or timeout:
  `t.Skip`.
- Captcha, challenge, or HTTP 403, 429 or 1020: `t.Skip`.
- 200 OK with a broken parser, or no results for the known query: `t.Fatal`.
- Broken decrypt or an internal error: `t.Fatal`.

## Messages in the app and the logs

The user-facing messages look like this:

- `AnimeFire temporarily unavailable: Cloudflare 521/origin down`
- `HiAnime blocked the request: captcha/challenge`
- `Goyabu responded but the parser could not find the expected data: ...`
- `AnimeFire download link expired or was rejected: HTTP 404`

After three consecutive origin or blocking failures, the circuit breaker skips
the source for 10 minutes so that a server that is down is not hammered.

## Discord

GoAnime already has local Discord Rich Presence, which is unrelated to
project health alerts. Posting diagnostics to a Discord channel safely needs a
separate pull request with:

- `DISCORD_WEBHOOK_URL` configured as a GitHub secret.
- A scheduled or manual job that runs
  `go test -tags sourcehealth -run TestSourceHealthLive -count=1 -v ./internal/scraper`.
- A step that posts only the summary of `healthy`, `skipped` and `failed`
  sources, without exposing tokens, cookies or private URLs.

Without that secret, the safe option is to keep the information in the CI log
and the local output.
