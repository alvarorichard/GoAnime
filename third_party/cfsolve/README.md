# Temporary cfsolve patches

cfsolve (the tool GoAnime uses to clear Goyabu's Cloudflare gate) does not build
on Windows upstream. `windows-build.patch` is
[chris124567/cfsolve#2](https://github.com/chris124567/cfsolve/pull/2) saved
verbatim after review (67 lines; everything it changes is behind `cfg(windows)`
or a Windows target check, so Linux and macOS binaries are unaffected). The
pull request's author reports it tested on Windows 11: `challenge.exe --solve`
cleared a Cloudflare challenge page 4 times out of 4.

The patch is applied on top of the commit pinned in
`internal/scraper/cfsolver/cfsolve.lock` by
`scripts/apply-cfsolve-patches.sh`. It is **temporary**:

- when upstream merges the pull request and the pin moves past it, the script
  notices the patch is already contained and skips it;
- then delete `windows-build.patch` and this folder.

Shown on a real Windows 11 machine of ours (2026-11-02): the
`challenge-windows-amd64.exe` built by CI cleared Goyabu's gate (HTTP 403, then
200, with a `cf_clearance`). The release workflow builds it in the
`cfsolve-windows` job; `.github/workflows/cfsolve-windows-trial.yml` is the
build-only check that runs when this folder changes.
