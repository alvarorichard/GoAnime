# QA E2E versioning gate

This guide defines the QA role for GoAnime versioning. The work is split into
two blocks: a deterministic suite that blocks releases, and a live/manual pass
that validates external sources before publishing a version.

## PM + QA acceptance criteria

A version can only be promoted when the deterministic gate passes, the live pass
is classified, and every exception has explicit PM + QA approval.

Release blockers:

- Panic, deadlock, crash, or fatal error during CLI/TUI initialization.
- Argument parsing failure or command routing regression for primary commands.
- Failure in the standard Go suite, lint, vet, security scan, or deterministic
  E2E suite.
- Search, selection, playback/download, or tracking failure when covered by
  mocks, fixtures, or a controlled environment.
- Binary build failure or artifact/checksum generation failure.
- Writes outside temporary directories during automated tests.
- External 200 OK response with broken parser/decrypt logic, nil pointer, panic,
  or internal contract regression.

Can be skipped without blocking release when recorded:

- DNS, timeout, Cloudflare, captcha, rate limit, or external provider outage.
- Discord RPC unavailable or no valid local Discord session.
- Missing `mpv`, `ffmpeg`, GPU, or codec when the test is live/optional.
- Performance differences caused by hardware, network, or external environment.

Safety rule: an external skip is acceptable only when equivalent local coverage
exists for GoAnime's main contract. Internal failures must not be reclassified as
skips to unblock release.

## QA agent

The QA agent is responsible for each version:

- Run the existing unit and integration tests.
- Run the deterministic CLI E2E suite.
- Run static analysis and security checks.
- Run live source diagnostics when network access is available.
- Record failures, external skips, and build evidence.

## Required versioning gate

Minimum commands before creating a tag:

```powershell
go mod verify
go fmt ./...
go vet ./...
golangci-lint run --timeout=15m
gosec ./...
govulncheck ./...
go test -v -race -coverprofile=coverage.out -covermode=atomic ./...
go test -tags e2e -count=1 -v ./test/e2e
go build -v ./cmd/goanime
```

The release workflow must run at least dependency verification, lint, Go tests
with race, deterministic E2E, binary build, and artifact checksums before
publishing. A deterministic failure blocks release.

## Deterministic E2E suite

Command:

```powershell
go test -tags e2e -count=1 -v ./test/e2e
```

Current coverage:

| Area | Covered journey |
| --- | --- |
| CLI | `--help`, `-h`, `--version`, invalid input, and pre-network validation |
| Playback | direct search route and media preferences |
| Anime downloads | single episode, range, all episodes, source, quality, and custom output |
| Movie/TV downloads | movie, TV episode, TV range, and all seasons |
| Upscale | full image parse with output, scale, passes, fast mode, workers, missing `ffmpeg`, and `ffmpeg` shim |
| Release | local build of the real binary used by smoke tests |

The suite avoids real scraping, real downloads, opening `mpv`, and video
processing. This keeps the gate reliable for CI and release.

## Complementary offline tests

In addition to the opt-in E2E package, composed flows now have deterministic
contracts with fakes inside their source packages:

| Package | Flows covered without network or real tools |
| --- | --- |
| `internal/handlers` | download delegation, movie download delegation, update, image upscale, `ffmpeg` validation, series playback, movie/single episode playback, and `ErrBackToSearch` recovery |
| `internal/download` | anime single episode, all with `ErrUserQuit`, range with legacy fallback, 9Anime range, movie redirect, and movie/TV download without interactive selection |
| `internal/tracking` | `!cgo` contract: local tracker does not initialize SQLite, does not create a database, and returns `ErrTrackerNotInited` nil-safely |

These tests should run with `go test ./...` and act as local equivalent coverage
for skips accepted during live QA.

## Live QA pass

Use this when preparing a version or investigating provider failures:

```powershell
go test -tags sourcehealth -run TestSourceHealthLive -count=1 -v ./internal/scraper
go test -tags integration -run TestFlixHQFullFlow -count=1 -v ./internal/api
go test ./pkg/goanime -run Integration -count=1 -v
```

Expected classifications:

- Source outage, Cloudflare, DNS, timeout, captcha, or rate limit: record as
  `skip`/external outage.
- 200 response without expected data, broken parser, or broken decrypt:
  blocking failure.
- Panics, nil pointers, and local errors: blocking failure.
- For a complete release, require at least one healthy anime source and one
  healthy movie/TV source. Otherwise, PM must explicitly approve a degraded
  release in the release note.

Live tests may retry once to confirm external instability. If the failure
continues, QA records it as `external skip` or `blocking failure` according to
the classification above.

## Required evidence

For each release candidate, record in the PR, tag, or release note:

- Tested hash, branch, or tag.
- `go version`, operating system, and architecture.
- CI/release workflow links.
- Summarized output from required commands.
- `coverage.out` or coverage summary.
- Deterministic E2E suite result.
- Live provider report: `healthy`, `external skip`, or `blocking failure`.
- Generated artifacts and checksums.
- Reason for each skip and PM/QA decision for every accepted risk.
- Confirmation that tests did not write outside temporary directories.

## First complete-pass checklist

1. Environment: confirm `go version`, `mpv --version`, and, when applicable,
   `ffmpeg -version`.
2. Local quality: `go fmt ./...`, `go vet ./...`, `golangci-lint run --timeout=15m`.
3. Security: `gosec ./...` and `govulncheck ./...`.
4. Tests: `go test -v -race -coverprofile=coverage.out -covermode=atomic ./...`.
5. Deterministic E2E: `go test -tags e2e -count=1 -v ./test/e2e`.
6. Live sources: run source health and integrations when the network is stable.
7. Build: generate binaries/installer through workflow or build scripts.
8. Evidence: attach commands, status, coverage, and every external skip to the
   PR or release note.

## Exception policy

A release can continue with skips or non-blocking failures only if:

- The risk is classified as an external dependency or optional capability.
- Critical flows remain covered by deterministic local tests.
- There is a follow-up issue or record when applicable.
- PM + QA explicitly approved the exception.

## Next increments

- Add a temporary `mpv` shim to E2E to validate playback handoff without opening
  a real player.
- Add HTTP mock servers to cover search-to-episode-to-stream without depending
  on third-party sites.
- Add a SQLite tracking test with CGO using a temporary database.
- Keep `sourcehealth` as a manual/scheduled job or release-candidate step,
  separate from mandatory PR checks when the outage is external.
