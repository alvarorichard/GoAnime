#!/usr/bin/env bash
# Signs a published release's checksums-sha256.txt with the offline minisign
# key and uploads the signature the updater requires.
#
#   build/sign-release.sh v2.0.0 [path/to/goanime.key]
#
# The secret key never leaves the maintainer's machine; CI cannot sign.
set -euo pipefail

tag="${1:?usage: sign-release.sh <tag> [secret-key]}"
key="${2:-$HOME/.minisign/goanime.key}"
repo="alvarorichard/GoAnime"

work="$(mktemp -d)"
trap 'rm -rf "$work"' EXIT

gh release download "$tag" --repo "$repo" --pattern checksums-sha256.txt --dir "$work"
minisign -S -s "$key" -m "$work/checksums-sha256.txt" \
  -t "goanime $tag file:checksums-sha256.txt"
gh release upload "$tag" --repo "$repo" "$work/checksums-sha256.txt.minisig"

echo "Signed $tag. Verify with:"
echo "  gh release download $tag --repo $repo -p 'checksums-sha256.txt*' && minisign -Vm checksums-sha256.txt -P <public key>"
