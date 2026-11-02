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

Not yet shown on a real Windows machine of ours: the CI build, and the solve
itself. See `.github/workflows/cfsolve-windows-trial.yml`.
