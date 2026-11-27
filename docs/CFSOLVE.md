# cfsolve (Cloudflare gates)

Goyabu sits behind a Cloudflare managed challenge, and Pobreflix's player pages
behind a standalone Cloudflare Turnstile. GoAnime clears both with
[cfsolve](https://github.com/chris124567/cfsolve), a Rust tool that solves the
challenges from plain requests, with no browser. It is an optional external
binary: GoAnime runs it with `os/exec` and does not link it, so building GoAnime
needs no Rust, no cmake and no cgo.

## For users

Nothing to do. Every release build of GoAnime carries the cfsolve binary for its
own platform inside the executable. At start it unpacks it into the config
folder (`~/.config/goanime/cfsolve/` on Linux,
`~/Library/Application Support/goanime/cfsolve/` on macOS,
`%AppData%\goanime\cfsolve\` on Windows) when the file there is not byte for
byte the one it carries, so there is no download and no setup, and an update
never keeps an older solver. The same binary answers both gates, and it comes
from the same cfsolve commit as the browser profile embedded in GoAnime
(`internal/scraper/cfsolver/profile/`), which cfsolve requires.

A release is published only if each of its binaries, started with an empty
config folder, unpacks its solver and that solver starts; the Linux binaries are
also started in AlmaLinux 9, whose glibc (2.34) is the oldest GoAnime itself
runs on (`scripts/check-bundled-cfsolve.sh`).

A build that carries no solver (`go install`, a development build) downloads the
binary attached to its own release into the same folder, after checking the
minisign signature of that release's `checksums-sha256.txt` against the key
embedded in GoAnime and the binary's SHA-256 against that list. A development
build has no release to download from and uses a binary you installed yourself.
Failures are silent; Goyabu then stays out of the search, and Pobreflix keeps
its direct files and StartFlix panel fallback.

Binary lookup order: `GOANIME_CFSOLVE_BIN`, next to `goanime`, a `cfsolve`
folder next to it, the config folder, `PATH`. `GOANIME_CFSOLVE_AUTODOWNLOAD=0`
turns the download off.

## Pobreflix's player gate

Pobreflix lists its movies and episodes as player pages on `streambetter.lat`.
Each page renders a Turnstile widget, and the media URL only appears after the
widget is answered:

1. GoAnime fetches `/filme/<tmdb>` (or `/serie/<tmdb>/<season>/<episode>`) and
   reads the widget's sitekey.
2. cfsolve returns a token without loading the page, so the request keeps the
   session that fetched the widget.
3. GoAnime posts the token to `/api/turnstile-verify`, reloads the page with the
   proof in `?tsv=`, and reads `window.__PLAYER_PARAMS__`.
4. The media proxy is requested with that proof-bearing page as `Referer`; it
   refuses the URL otherwise.

With cfsolve installed, a page that already carries `__PLAYER_PARAMS__` is
played without invoking it, and a title with its own MP4 never needs the player.
Without cfsolve, GoAnime makes no gate requests at all: direct MP4s still play
and the rest takes the previous panel fallback. A catalog
`/api/player?id=<tmdb>` that answers 404 is rebuilt from the TMDB id.
`GOANIME_POBREFLIX_PLAYER_URL` repoints the player host without a release.

## For developers

`git clone`, `go build` and `go test` work with nothing else installed; the
solver tests use a fake binary. To see the gates working:

    scripts/get-cfsolve.sh            # latest release's binary, checksum verified
    scripts/get-cfsolve.sh --build    # or build the commit in cfsolve.lock (Rust, cmake, libclang)
    go test -tags live -run TestGoyabuThroughCfsolveLive -v ./internal/scraper/providers/goyabu
    go test -tags live -run TestLivePobreflixPlayerThroughCfsolve -v ./internal/scraper/providers/pobreflix

The Pobreflix test clears the live gate for *Todo Poderoso* (TMDB 310) and skips
when cfsolve is not installed.

## Which commit ships

CI and release builds use cfsolve's latest commit. `cfsolve-build.yml` reads
upstream's HEAD once per run and builds that commit for every platform, and each
GoAnime build runs `scripts/update-cfsolve.sh <that commit> --no-build` before
`go build`, so the profile it embeds comes from the same commit as the binary it
bundles: cfsolve refuses a profile that names another reference build than its
own, which a new commit can change. The release notes name the commit a release
was built from.

`internal/scraper/cfsolver/cfsolve.lock` and the profile beside it are the copy
committed in the repository, which `go test`, the `build/` scripts and
`scripts/get-cfsolve.sh --build` use. The latest release's binary
(`scripts/get-cfsolve.sh`, and the one the `build/` scripts bundle) can be
refused by a checkout whose lock names another commit than that release was
built from; use `--build` then.

    scripts/update-cfsolve.sh <sha|HEAD>   # moves the lock and the profile together

| Workflow | When | What |
| --- | --- | --- |
| `cfsolve-build.yml` | called by `ci.yml` and `release.yml` | reads upstream's latest commit and builds it for the five platforms. CI keeps each binary in the Actions cache until upstream moves; a release always compiles |
| `ci.yml` | every push and pull request | bundles that build and its profile in the binaries and the Windows installer it uploads, and runs the same new-user check on each binary that carries one; a failed cfsolve build only leaves that binary without it |
| `release.yml` | every release tag | bundles that build and its profile in each binary, checks what a new user gets (`check-bundled-cfsolve.sh`), and adds the binaries to the signed `checksums-sha256.txt` and to the release assets. Required: a platform whose solver is missing or does not start stops the release |
| `source-health.yml` (`goyabu-cfsolve`) | manual | builds upstream's latest commit with that commit's profile, as the next release would ship it, and runs the live Goyabu test |
| `cfsolve-update.yml` | weekly | tries upstream HEAD in a read-only job; if it builds and clears the gate, opens a PR moving the committed lock and profile, and warns when it does not |

If Cloudflare changes a gate, the affected source stops working until cfsolve
is fixed upstream; the next release then ships the fix with no change here.
`cfsolve-update.yml` (weekly) and `goyabu-cfsolve` (on demand) run the live
Goyabu test against upstream's latest commit, so a warning from the first or a
failure of the second means the next release would ship a solver that does not
clear the gate. The Pobreflix live test above checks the
Turnstile path by hand.

## Platforms

Releases carry cfsolve for **linux/amd64, linux/arm64, darwin/arm64,
darwin/amd64 and windows/amd64**, built in CI, and bundle the one for their own
platform inside the GoAnime executable (it is unpacked on first run). The Linux
builds are compiled in AlmaLinux 9, so they need glibc 2.34 at most, as GoAnime
does; built on Ubuntu 24.04, the amd64 one needed 2.38 and did not start on
Ubuntu 22.04, Linux Mint 21, Debian 12 or RHEL 9. `cfsolve-build.yml` checks the
version and starts each Linux binary in that image before it can be bundled.
The Windows build is the only one tested on a real
machine of its kind: on Windows 11 it cleared Goyabu's gate (HTTP 403, then 200,
with a `cf_clearance`). It is compiled with MinGW-w64; upstream merged the
Windows support ([chris124567/cfsolve#2](https://github.com/chris124567/cfsolve/pull/2)),
so GoAnime no longer patches it. The Intel Mac build is cross-compiled on an
Apple Silicon runner and has not been run on real hardware.
On a platform without a build, Goyabu is simply absent unless the network does
not challenge it and you enable it with `GOANIME_ENABLED_SOURCES=goyabu`;
Pobreflix still plays its direct files and falls back to the panel.
