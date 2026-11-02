#!/usr/bin/env bash
# Moves GoAnime to another cfsolve commit: refreshes the embedded browser
# profile from it, rewrites cfsolve.lock and (unless --no-build) builds it.
#
#   scripts/update-cfsolve.sh <sha|HEAD> [--no-build]
#
# Commit the result; the next GoAnime release builds the new commit and attaches
# its binaries. The profile and the binary must come from the same commit,
# which is why one script changes both.
set -euo pipefail

target="${1:?usage: update-cfsolve.sh <sha|HEAD> [--no-build]}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
pkg="$root/internal/scraper/cfsolver"
repo="$(sed -n 's/^repo=//p' "$pkg/cfsolve.lock")"

work="$(mktemp -d)"; trap 'rm -rf "$work"' EXIT
git clone --quiet "$repo" "$work/src"
git -C "$work/src" checkout --quiet "$target"
sha="$(git -C "$work/src" rev-parse HEAD)"

if [ "${2:-}" != "--no-build" ]; then
  (cd "$work/src" && cargo build --release --bin challenge)
fi

rm -rf "$pkg/profile/safari-26-ios" "$pkg/profile/safari-26-ios.json"
cp -r "$work/src/profiles/safari-26-ios" "$work/src/profiles/safari-26-ios.json" "$pkg/profile/"
cp "$work/src/LICENSE" "$pkg/profile/LICENSE-cfsolve"

cat > "$pkg/cfsolve.lock" <<LOCK
# The cfsolve commit GoAnime is built and tested against. Never track upstream's
# HEAD: bump this with scripts/update-cfsolve.sh, which also refreshes the
# embedded profile from the same commit. The release workflow builds this
# commit and attaches the binaries to each GoAnime release.
repo=$repo
sha=$sha
LOCK
echo "cfsolve -> ${sha:0:12}; review 'git diff' and run the live Goyabu test"
