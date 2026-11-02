# cfsolve (Goyabu's Cloudflare gate)

Goyabu sits behind a Cloudflare managed challenge. GoAnime clears it with
[cfsolve](https://github.com/chris124567/cfsolve), a Rust tool that solves the
challenge from plain requests, with no browser. It is an optional external
binary: GoAnime runs it with `os/exec` and does not link it, so building GoAnime
needs no Rust, no cmake and no cgo.

## For users

Nothing to do. On first run GoAnime downloads the binary built for the pinned
commit from its own releases (`cfsolve-<sha12>`) into the config folder
(`~/.config/goanime/cfsolve/`), after verifying the minisign signature of
`SHA256SUMS` against the key embedded in GoAnime and the binary's SHA-256 against
that list. Failures are silent; Goyabu then stays out of the search.

Binary lookup order: `GOANIME_CFSOLVE_BIN`, next to `goanime`, a `cfsolve`
folder next to it, the config folder, `PATH`. The browser profile it imitates is
embedded in GoAnime (`internal/scraper/cfsolver/profile/`).

## For developers

`git clone`, `go build` and `go test` work with nothing else installed; the
solver tests use a fake binary. To see Goyabu working:

    scripts/get-cfsolve.sh            # download the CI build and verify it
    scripts/get-cfsolve.sh --build    # or build the pinned commit (Rust, cmake, libclang)
    go test -tags live -run TestGoyabuThroughCfsolveLive -v ./internal/scraper/providers/goyabu

## Pinning and updates

GoAnime never follows cfsolve's HEAD. `internal/scraper/cfsolver/cfsolve.lock`
names one commit, and the embedded profile must come from that same commit.

    scripts/update-cfsolve.sh <sha|HEAD>   # moves the lock and the profile together

| Workflow | When | What |
| --- | --- | --- |
| `cfsolve-build.yml` | the lock changes | builds the pinned commit, signs `SHA256SUMS` (environment `release`, needs approval), publishes release `cfsolve-<sha12>` |
| `source-health.yml` (`goyabu-cfsolve`) | daily | installs the published binary the way a user would and runs the live Goyabu test |
| `cfsolve-update.yml` | weekly | tries upstream HEAD in a read-only job; if it builds and clears the gate, opens a PR moving the pin |

A red daily check on a published pin means Cloudflare changed the gate. The
weekly workflow shows whether upstream already fixed it; merge its PR, approve
the release environment, and users pick up the new binary on their next run.

## Platforms

linux/amd64 is verified. linux/arm64, darwin/arm64 and windows/amd64 are built
best effort by CI and have not been tested; a platform without an asset simply
has no Goyabu. There is no darwin/amd64 build.
