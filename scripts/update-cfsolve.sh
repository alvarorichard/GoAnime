#!/usr/bin/env bash
# Moves GoAnime to another cfsolve commit: refreshes the embedded browser
# profile from it, rewrites cfsolve.lock and (unless --no-build) builds it.
#
#   scripts/update-cfsolve.sh <sha|HEAD> [--no-build]
#
# CI and release builds run it with upstream's latest commit before "go build"
# (.github/workflows/cfsolve-build.yml), so what they ship always follows
# upstream. Commit its result to move the copy in the repository, which
# "go test", local builds and get-cfsolve.sh --build use. The profile and the
# binary must come from the same commit, which is why one script changes both.
set -euo pipefail

target="${1:?usage: update-cfsolve.sh <sha|HEAD> [--no-build]}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
pkg="$root/internal/scraper/cfsolver"
# A Windows checkout may have given the lock CRLF line endings.
repo="$(sed -n 's/^repo=//p' "$pkg/cfsolve.lock" | tr -d '\r')"

work="$(mktemp -d)"; trap 'rm -rf "$work"' EXIT
# No CRLF conversion: the profile must be byte for byte the one upstream ships.
git clone --quiet --config core.autocrlf=false "$repo" "$work/src"
git -C "$work/src" checkout --quiet "$target"
sha="$(git -C "$work/src" rev-parse HEAD)"

if [ "${2:-}" != "--no-build" ]; then
  (cd "$work/src" && cargo build --release --bin challenge)
fi

rm -rf "$pkg/profile/safari-26-ios" "$pkg/profile/safari-26-ios.json"
cp -r "$work/src/profiles/safari-26-ios" "$work/src/profiles/safari-26-ios.json" "$pkg/profile/"
cp "$work/src/LICENSE" "$pkg/profile/LICENSE-cfsolve"

cat > "$pkg/cfsolve.lock" <<LOCK
# The cfsolve commit the profile in this folder comes from. CI and release
# builds move both to upstream's latest commit before building
# (.github/workflows/cfsolve-build.yml); "go test", local builds and
# scripts/get-cfsolve.sh --build use what is committed here. Change it with
# scripts/update-cfsolve.sh, which moves the sha and the profile together.
repo=$repo
sha=$sha
LOCK
echo "cfsolve -> ${sha:0:12}; review 'git diff' and run the live Goyabu test"
