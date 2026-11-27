#!/usr/bin/env bash
# Installs the cfsolve binary Goyabu needs, where GoAnime looks for it.
#
#   scripts/get-cfsolve.sh           download the binary from GoAnime's latest release and verify it
#   scripts/get-cfsolve.sh --build   build the commit pinned in cfsolve.lock (Rust, cmake, libclang)
#   scripts/get-cfsolve.sh --bundle GOOS GOARCH
#                                    put that platform's binary in internal/scraper/cfsolver/embedded,
#                                    so the next "go build" carries it inside GoAnime (the build/
#                                    scripts call this; a missing binary only warns)
#
# Release binaries are built from cfsolve's latest commit at the time of that
# release, which can differ from the commit in this checkout's
# internal/scraper/cfsolver/cfsolve.lock, and cfsolve can refuse a profile from
# another commit. Use --build to get exactly the commit of the checkout you are on.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
lock="$root/internal/scraper/cfsolver/cfsolve.lock"
sha="$(sed -n 's/^sha=//p' "$lock" | tr -d '\r')"
repo="$(sed -n 's/^repo=//p' "$lock" | tr -d '\r')"
[ -n "$sha" ] || { echo "no sha in $lock" >&2; exit 1; }

case "$(uname -s)" in
  Darwin) cfg="${HOME}/Library/Application Support" ;;
  *)      cfg="${XDG_CONFIG_HOME:-$HOME/.config}" ;;
esac
dest="$cfg/goanime/cfsolve"
bundle=""
if [ "${1:-}" = "--bundle" ]; then
  [ -n "${2:-}" ] && [ -n "${3:-}" ] || { echo "usage: $0 --bundle GOOS GOARCH" >&2; exit 2; }
  bundle=1
  os="$2"; arch="$3"
  dest="$root/internal/scraper/cfsolver/embedded"
fi
mkdir -p "$dest"

if [ "${1:-}" = "--build" ]; then
  for tool in git cargo cmake; do
    command -v "$tool" >/dev/null || { echo "--build needs $tool" >&2; exit 1; }
  done
  work="$(mktemp -d)"; trap 'rm -rf "$work"' EXIT
  git clone --quiet "$repo" "$work/src"
  git -C "$work/src" checkout --quiet "$sha"
  (cd "$work/src" && cargo build --release --bin challenge)
  install -m 0755 "$work/src/target/release/challenge" "$dest/challenge"
  echo "built cfsolve ${sha:0:12} -> $dest/challenge"
  exit 0
fi

if [ -z "$bundle" ]; then
  os="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "$(uname -m)" in
    x86_64|amd64) arch=amd64 ;;
    aarch64|arm64) arch=arm64 ;;
    *) echo "no prebuilt cfsolve for $(uname -m); try --build" >&2; exit 1 ;;
  esac
fi
asset="challenge-$os-$arch"
[ "$os" = windows ] && asset="$asset.exe"
case "$asset" in
  challenge-linux-amd64|challenge-linux-arm64|challenge-darwin-arm64|challenge-darwin-amd64) ;;
  challenge-windows-amd64.exe) [ -n "$bundle" ] || { echo "on Windows GoAnime installs cfsolve itself" >&2; exit 1; } ;;
  *) echo "releases carry no cfsolve for $os/$arch; try --build" >&2; exit 1 ;;
esac
base="https://github.com/alvarorichard/GoAnime/releases/latest/download"

work="$(mktemp -d)"; trap 'rm -rf "$work"' EXIT
curl -fsSL "$base/checksums-sha256.txt" -o "$work/checksums-sha256.txt"
curl -fsSL "$base/checksums-sha256.txt.minisig" -o "$work/checksums-sha256.txt.minisig"
curl -fsSL "$base/$asset" -o "$work/$asset"
# The checksum list is signed with the key GoAnime's updater trusts. Check it
# when minisign is installed; the checksum below still has to match either way.
if command -v minisign >/dev/null; then
  pub="$(grep -oE '"RW[A-Za-z0-9+/=]{50,}"' "$root/internal/updater/minisign.go" | head -1 | tr -d '"')"
  minisign -V -P "$pub" -m "$work/checksums-sha256.txt" -x "$work/checksums-sha256.txt.minisig" >/dev/null \
    || { echo "checksums-sha256.txt signature does not verify" >&2; exit 1; }
else
  echo "note: minisign not installed; skipping the signature check (GoAnime itself always checks it)" >&2
fi
want="$(awk -v a="$asset" '$2 == a || $2 == "*"a {print $1}' "$work/checksums-sha256.txt")"
[ -n "$want" ] || { echo "the latest release lists no $asset (it may predate cfsolve support); try --build" >&2; exit 1; }
if command -v sha256sum >/dev/null; then got="$(sha256sum "$work/$asset" | cut -d' ' -f1)"
else got="$(shasum -a 256 "$work/$asset" | cut -d' ' -f1)"; fi
[ "$got" = "$want" ] || { echo "checksum mismatch for $asset" >&2; exit 1; }
if [ -n "$bundle" ]; then
  # go:embed takes every file in the folder: keep only this platform's binary.
  rm -f "$dest"/challenge-*
  install -m 0644 "$work/$asset" "$dest/$asset"
  echo "bundled $asset from the latest GoAnime release -> $dest/$asset"
  exit 0
fi
install -m 0755 "$work/$asset" "$dest/challenge"
echo "installed $asset from the latest GoAnime release -> $dest/challenge"
