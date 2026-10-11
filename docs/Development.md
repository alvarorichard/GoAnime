# Development guide

This guide covers how GoAnime is developed: the branch model, the checks every
change has to pass, how the code is organized, and the Go 1.27 conventions the
code relies on. For a shorter overview of how to contribute, start with
[CONTRIBUTING.md](../CONTRIBUTING.md).

## Contents

- [Branches](#branches)
- [Setting up](#setting-up)
- [Making a change](#making-a-change)
- [Quality checks](#quality-checks)
- [Testing](#testing)
- [Code style](#code-style)
- [Project layout](#project-layout)
- [Adding or removing a source](#adding-or-removing-a-source)
- [Building](#building)
- [Editor setup](#editor-setup)
- [Go 1.27 conventions](#go-127-conventions)
- [Getting help](#getting-help)

## Branches

- `main` holds released code. It only receives merges from `dev` and hotfixes.
- `dev` is where changes are integrated. Every pull request targets `dev`.
- Work happens on short-lived branches created from `dev`: `feature/<name>`
  for new functionality and `bugfix/<name>` for fixes.
- `hotfix/<name>` is reserved for critical fixes that cannot wait for the next
  release and go to `main` directly.

Nothing is committed to `main` directly.

## Setting up

You need Go 1.27 or later and [mpv](https://mpv.io/). Progress tracking uses
SQLite through CGO, so a C compiler is needed for a build that tracks
progress. Without one, GoAnime builds and runs without tracking. See
[BUILD_OPTIONS.md](BUILD_OPTIONS.md).

```bash
git clone https://github.com/alvarorichard/GoAnime.git
cd GoAnime
git checkout dev
go mod download
go build ./cmd/goanime
./goanime
```

## Making a change

```bash
git checkout dev
git pull origin dev
git checkout -b feature/your-feature
# make the change, then:
git commit -m "feat: describe the change"
git push origin feature/your-feature
```

Open the pull request against `dev` after the [quality checks](#quality-checks)
pass locally. A change is merged once every CI check passes and at least one
maintainer has approved it.

### Commit messages

Commits follow [Conventional Commits](https://www.conventionalcommits.org/):

```
type(scope): description

optional body

optional footer
```

The types in use are `feat`, `fix`, `docs`, `style`, `refactor`, `test` and
`chore`:

```
feat(player): add video quality selection
fix(animefire): handle the new episode list markup
docs: document GOANIME_HIANIME_AUDIO
```

### Pull requests

- Use a title in the same format as a commit message.
- Explain what changed and why, and link the issue it fixes.
- Describe how you tested it.
- Call out breaking changes explicitly.
- Include a screenshot or recording when the change affects what the terminal
  UI shows.

## Quality checks

These checks are mandatory. CI runs all of them, and a pull request that fails
any of them is not merged. Every finding has to be fixed or explained in the
pull request description.

| Tool            | What it checks                                  | Install                                                          |
| --------------- | ----------------------------------------------- | ---------------------------------------------------------------- |
| `gofmt`         | Formatting                                      | Included with Go                                                 |
| `go vet`        | Common mistakes                                 | Included with Go                                                 |
| `golangci-lint` | Correctness and resource safety, including staticcheck and every go-critic checker | [Installation guide](https://golangci-lint.run/docs/welcome/install/) |
| `gosec`         | Insecure code patterns                          | `go install github.com/securego/gosec/v2/cmd/gosec@latest`       |
| `govulncheck`   | Known vulnerabilities in reachable dependencies | `go install golang.org/x/vuln/cmd/govulncheck@latest`            |

The linters and their thresholds are configured in
[.golangci.yml](../.golangci.yml). CI pins the exact version of each tool in
[ci.yml](../.github/workflows/ci.yml); if a local run disagrees with CI,
install those versions.

Run them before pushing:

```bash
gofmt -l .                     # must print nothing
go vet ./...
golangci-lint run ./...
gosec -exclude-generated ./...
govulncheck ./...
go test -short -race -count=1 -covermode=atomic -coverprofile=coverage.out ./...
```

On every push and pull request to `main` or `dev`, CI also:

- checks that `go.mod` and `go.sum` are tidy and verified,
- reviews dependency changes,
- runs the tests with the race detector on Linux, macOS and Windows,
- fails if total coverage drops below 66%,
- runs the goroutine leak gate and the JSON decoder fuzzers (see
  [Go 1.27 conventions](#go-127-conventions)),
- builds for every release platform.

CodeQL runs its extended security and quality queries in a separate workflow,
and Dependabot keeps dependencies and actions up to date. The sources are
checked against the live sites once a day by the
[source health workflow](../.github/workflows/source-health.yml), outside the
pull request gate.

## Testing

What a change needs:

- New features need unit tests for the main behavior and its edge cases.
- Bug fixes need a test that fails without the fix.
- Changes to existing behavior or core logic need the affected tests added or
  updated.
- Documentation-only changes need no tests.

When in doubt, add the test. A reviewer will say if it is unnecessary.

```bash
go test -short ./...                        # the suite CI runs
go test -short ./internal/player/           # one package
go test -short -run TestName ./internal/api/
go test -short -cover ./...
```

Tests that talk to the real sites are behind build tags and only run in the
daily source health workflow:

```bash
go test -tags sourcehealth -run TestSourceHealthLive -count=1 -v ./internal/scraper
```

Tests live next to the code they test, in `_test.go` files of the same
package. Suites that span packages are under `test/`. Prefer table-driven
tests when a function has several cases. Tests that wait on timeouts or
retries belong in a `testing/synctest` bubble, described under
[Go 1.27 conventions](#slow-tests-testingsynctest-and-httptestnewtestserver).

## Code style

- Format with `gofmt` and follow standard Go naming.
- Give every exported identifier a doc comment.
- Handle every error, and wrap it with context (`fmt.Errorf("...: %w", err)`)
  when passing it up.
- Keep packages focused. Code that is not part of the public library belongs
  under `internal/`.
- Use comments to explain why the code does something, its edge cases and its
  expected inputs, not to restate what each line does.

## Project layout

```
cmd/goanime/           entry point
pkg/goanime/           public library for search and scraping
internal/api/          search across sources, episode lists, metadata enrichment
internal/api/source/   the source registry
internal/appflow/      application flow from search to playback
internal/scraper/      per-source scrapers, one package per source under providers/
internal/handlers/     flow controllers for media playback
internal/playback/     playback logic per content type (anime, movies, series)
internal/player/       mpv integration, with platform-specific code for Unix and Windows
internal/download/     download workflow (single episodes, ranges, whole series)
internal/downloader/   file and HLS downloads
internal/tracking/     watch progress and history (SQLite with CGO)
internal/discord/      Discord Rich Presence
internal/tui/          terminal UI helpers around go-fuzzyfinder
internal/upscaler/     Anime4K upscaling
internal/updater/      self-update (goanime --update)
internal/models/       shared data types
internal/util/         shared HTTP client, logging, flags and helpers
internal/version/      version information
build/                 build scripts and the Windows installer definition
test/                  test suites that span packages
docs/                  documentation and the README demo
```

The public library is documented in [LIBRARY_USAGE.md](LIBRARY_USAGE.md).

## Adding or removing a source

Every source registers itself in the source registry
(`internal/api/source`). A source describes itself as data through
`Describe() source.Descriptor` and calls `Register` in `init()`. Resolution,
the search fan-out, the circuit breaker and the kill switch all read that
data, so there is no central `switch` to update in the dispatch path.

The procedure, the full list of files to touch and the checklists are in
[ADDING_A_SOURCE.md](ADDING_A_SOURCE.md). Goyabu is the reference
implementation: it is the simplest complete source in the tree.

### The layers

```
internal/scraper/providers/<name>/          leaf HTTP and scraping client
internal/scraper/manager.go                 UnifiedScraper adapter
internal/api/providers/source_providers.go  the source.Source, registered in init()
internal/api/source/                        registry: Register, Resolve, ActiveSources
```

### Adding

Adding a source currently means 12 files and about 20 edit sites. Three of
them are the architecture: the leaf package, `source/kind.go`, and the
provider block in `source_providers.go`. The rest (the `UnifiedScraper`
adapter, `tagging.go`, `tagPattern` in `naming.go`, the five hand-maintained
switches in `enhanced.go`, `source_health.go`, and the `--source` help text)
is migration debt that has not been removed yet. ADDING_A_SOURCE.md lists
every site with its file and line. Use it as a checklist rather than working
from memory: forgetting `naming.go` or `source_health.go` fails silently, not
at compile time.

Capabilities are discovered by type assertion, not by a flag:

- `HasSeasons() bool` (`source.Seasoned`) marks a catalog organized into
  seasons, such as StartFlix and TopCine.
- `Search(ctx, query)` (`source.Searchable`) makes a source searchable. A
  source without it is left out of the search fan-out without any error.
- `WarmUp(ctx) error` (`source.BrowserGated`) is for a source that has to
  drive a browser to pass a bot check. No source implements it today; every
  registered source is plain HTTP.

### Removing

Pick the weakest option that works:

1. At runtime, with no rebuild: `GOANIME_DISABLED_SOURCES=goyabu`. This is the
   right first response to a source that broke overnight.
2. Ship it disabled: set `DefaultDisabled: true` in the descriptor. Users opt
   in with `GOANIME_ENABLED_SOURCES`, and the code and tests stay.
3. Delete it: follow the ordered checklist in
   [ADDING_A_SOURCE.md](ADDING_A_SOURCE.md#removing). It removes references
   before the things they reference, so the build stays green at every step.
   Mind the `ScraperType` `iota`: append or delete at the end, never insert or
   remove in the middle.

After adding or removing a source, run the [quality checks](#quality-checks)
and confirm coverage stays at or above 66%.

## Building

For a local build:

```bash
go build ./cmd/goanime
```

The scripts in `build/` cross-compile and package binaries for each platform.
They use paths relative to `build/`, so run them from there:

```bash
cd build
./buildlinux.sh               # Linux, without SQLite
./buildlinux-with-sqlite.sh   # Linux, with SQLite tracking
./buildmacos.sh               # macOS, amd64 and arm64
./buildwindows.sh             # Windows, from a Unix shell
```

`buildwindows.ps1` does the same for Windows from PowerShell. Release binaries
are built by the [release workflow](../.github/workflows/release.yml) when a
version tag is pushed. Build tags and CGO options are covered in
[BUILD_OPTIONS.md](BUILD_OPTIONS.md).

## Editor setup

Any editor with gopls works. Use [Delve](https://github.com/go-delve/delve)
for debugging.

### VS Code

Create `.vscode/settings.json` (the directory is ignored by git):

```json
{
    "go.formatTool": "gofmt",
    "go.lintTool": "golangci-lint",
    "go.testFlags": ["-v"],
    "editor.formatOnSave": true,
    "editor.codeActionsOnSave": {
        "source.organizeImports": true
    },
    "gopls": {
        "build.buildFlags": ["-tags=sourcehealth"]
    },
    "go.buildTags": "sourcehealth",
    "go.testTags": "sourcehealth"
}
```

The `sourcehealth` tag matters for
`internal/scraper/source_health_live_test.go`, which is behind
`//go:build sourcehealth`. Without it gopls skips the file and reports
"No packages found for open file".

## Go 1.27 conventions

The project targets Go 1.27 and relies on a few of its features in ways that
are easy to undo by accident. These are the rules that keep them working.

### Decoding JSON: use `internal/util/jsonx`, not `encoding/json`

Every provider and API response is decoded through `internal/util/jsonx`,
which wraps `encoding/json/v2`:

```go
// Bytes are needed afterwards (HTML guard, regex fallback, diagnostics):
body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
...
if err := jsonx.Unmarshal(body, &result); err != nil { ... }

// Bytes are only decoded: stream, and bound the read:
if err := jsonx.Decode(resp.Body, 5<<20, &result); err != nil { ... }
```

- `jsonx.Unmarshal` is about 17% faster than `json.Unmarshal` on a realistic
  payload.
- `jsonx.Decode` is about 24% faster than `io.ReadAll` followed by
  `json.Unmarshal`, and allocates about 57% fewer bytes, because the body is
  never buffered.
- Never call `json.NewDecoder(resp.Body).Decode(...)`. It reads without any
  size bound and silently ignores anything after the first JSON value.
  `jsonx.Decode` takes an explicit limit and fails with `jsonx.ErrTooLarge`.

The option set is `DefaultOptionsV1` minus `ReportErrorsWithLegacySemantics`,
the only v1 option that costs real time, so decoding semantics match
`encoding/json` exactly and only the error values differ. Differential tests
and two fuzz targets in `internal/util/jsonx` enforce that, and CI runs the
fuzzers on every push. If you change the option set, run them.

Reproduce the numbers with:

```bash
go test ./internal/util/jsonx/ -bench JSONXVsV1 -benchmem -run XXX
```

### Goroutine leaks are a CI gate

`internal/util/leakcheck` exposes Go 1.27's `goroutineleak` profile as an
assertion. Add a guard to any test that drives concurrent work:

```go
baseline := leakcheck.Count(t)
// ... run the flow ...
leakcheck.AssertNoNewLeaks(t, baseline)
```

Detection only runs when the profile is written, which `leakcheck` handles.
Calling `pprof.Lookup("goroutineleak").Count()` on its own returns a stale
number. The end-to-end gate lives in `test/leak`, and `leakcheck`'s own test
deliberately leaks a goroutine to prove the gate can still fail.

### Slow tests: `testing/synctest` and `httptest.NewTestServer`

Tests that wait on retry backoff or timeouts belong in a synctest bubble,
where the clock is virtual and the wait is instant:

```go
func TestSomethingWithBackoff(t *testing.T) {
    t.Parallel()
    synctest.Test(t, func(t *testing.T) {
        srv := httptest.NewTestServer(t, handler) // in-memory, no TCP port
        client := srv.Client()                    // starts it and fills srv.URL
        ...
    })
}
```

`httptest.NewTestServer` is required inside a bubble: a real loopback
connection is not "durably blocked", so the fake clock would never advance.
`srv.URL` is only populated after the first `srv.Client()` call.

This took the `internal/downloader/hls` tests from 15.1 s to 0.02 s.

A test server whose handler parks, to simulate a slow upstream, must be
released before `srv.Close()` runs, otherwise `Close` waits for it forever.
Prefer `<-r.Context().Done()`. When the client may retry on a fresh context,
also select on a stop channel closed in a `t.Cleanup` registered after
`t.Cleanup(srv.Close)`, since cleanups run in reverse order.

### Modernizers

`go fix ./...` applies the standard modernizers (`wg.Go`, `new(expr)`,
`errors.AsType`, the `slices` and `maps` helpers, `strings.CutLast`, and
others). Review the diff before committing: it is occasionally wrong about
statement lists inside one-line closures.

## Getting help

- Check the other documents in `docs/`.
- Search the [issues](https://github.com/alvarorichard/GoAnime/issues).
- Ask on [Discord](https://discord.gg/FbQuf78D9G) or in
  [Discussions](https://github.com/alvarorichard/GoAnime/discussions).
- Ask for help during review whenever something is unclear.
