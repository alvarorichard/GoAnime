#!/usr/bin/env bash
# Applies GoAnime's temporary patches (third_party/cfsolve/*.patch) to a cfsolve
# checkout, so CI can build cfsolve for platforms upstream does not support yet.
#
#   scripts/apply-cfsolve-patches.sh <path-to-cfsolve-checkout>
#
# A patch the checkout already contains (upstream merged it, and the pin moved
# past it) is skipped with a notice, so the patches retire themselves. A patch
# that neither applies nor is already there fails loudly: the pin moved and the
# patch needs a look.
set -euo pipefail

src="${1:?usage: apply-cfsolve-patches.sh <cfsolve checkout>}"
root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

shopt -s nullglob
for patch in "$root"/third_party/cfsolve/*.patch; do
  name="$(basename "$patch")"
  if git -C "$src" apply --check "$patch" 2>/dev/null; then
    git -C "$src" apply "$patch"
    echo "applied $name"
  elif git -C "$src" apply --check --reverse "$patch" 2>/dev/null; then
    echo "skipped $name: this cfsolve commit already contains it (delete the patch)"
  else
    echo "::error::$name neither applies to nor is contained in this cfsolve commit" >&2
    exit 1
  fi
done
