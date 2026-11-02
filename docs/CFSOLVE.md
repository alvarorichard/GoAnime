# cfsolve (Goyabu's Cloudflare gate)

Goyabu sits behind a Cloudflare managed challenge. GoAnime clears it with
[cfsolve](https://github.com/chris124567/cfsolve), a Rust tool that solves the
challenge from plain requests, with no browser. It is an optional external
binary: GoAnime runs it with `os/exec` and does not link it, so building GoAnime
needs no Rust, no cmake and no cgo.

## For users

Nothing to do. On first run a release build of GoAnime downloads the cfsolve
binary attached to **its own release** into the config folder
(`~/.config/goanime/cfsolve/`), after checking the minisign signature of that
release's `checksums-sha256.txt` against the key embedded in GoAnime and the
binary's SHA-256 against that list. Failures are silent; Goyabu then stays out
of the search.

Because the binary comes from the same release as the GoAnime that runs it, it
always matches the browser profile embedded in GoAnime
(`internal/scraper/cfsolver/profile/`). A development build has no release to
download from and uses a binary you installed yourself.

Binary lookup order: `GOANIME_CFSOLVE_BIN`, next to `goanime`, a `cfsolve`
folder next to it, the config folder, `PATH`. `GOANIME_CFSOLVE_AUTODOWNLOAD=0`
turns the download off.

## For developers

`git clone`, `go build` and `go test` work with nothing else installed; the
solver tests use a fake binary. To see Goyabu working:

    scripts/get-cfsolve.sh            # latest release's binary, checksum verified
    scripts/get-cfsolve.sh --build    # or build the pinned commit (Rust, cmake, libclang)
    go test -tags live -run TestGoyabuThroughCfsolveLive -v ./internal/scraper/providers/goyabu

## Pinning and updates

GoAnime never follows cfsolve's HEAD. `internal/scraper/cfsolver/cfsolve.lock`
names one commit, and the embedded profile must come from that same commit.

    scripts/update-cfsolve.sh <sha|HEAD>   # moves the lock and the profile together

| Workflow | When | What |
| --- | --- | --- |
| `release.yml` (job `cfsolve`) | every release tag | builds the pinned commit, adds the binaries to the signed `checksums-sha256.txt` and to the release assets. Optional: a failure never blocks a release |
| `source-health.yml` (`goyabu-cfsolve`) | manual | installs the latest release's binary the way a user would and runs the live Goyabu test |
| `cfsolve-update.yml` | weekly | tries upstream HEAD in a read-only job; if it builds and clears the gate, opens a PR moving the pin |

If Cloudflare changes the gate, Goyabu stops working until cfsolve is updated.
The weekly workflow shows whether upstream already fixed it; merge its PR and
the next release carries the new binary and profile together.

## Platforms

Releases carry cfsolve for **linux/amd64, linux/arm64 and darwin/arm64**, built
in CI (not exercised on real machines of each kind). There is **no Windows
build**: cfsolve uses Unix-only code there (`tokio::signal::unix`,
`clock_gettime`), and the upstream pull request that adds it
([chris124567/cfsolve#2](https://github.com/chris124567/cfsolve/pull/2)) is not
merged. There is no Intel Mac build either. On those platforms Goyabu is simply
absent unless the network does not challenge it and you enable it with
`GOANIME_ENABLED_SOURCES=goyabu`.
