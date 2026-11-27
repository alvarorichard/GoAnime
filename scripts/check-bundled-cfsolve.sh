#!/usr/bin/env bash
# Starts a GoAnime build the way a new user does, with an empty config folder,
# and checks what that user gets with no setup at all: GoAnime unpacks the
# cfsolve it carries, byte for byte the binary that was bundled, and that
# solver starts.
#
#   scripts/check-bundled-cfsolve.sh GOANIME_BINARY BUNDLED_SOLVER [--in-image IMAGE]
#
# --in-image (Linux, needs docker) runs the same check inside IMAGE as well, so
# a solver that needs a newer glibc than the image has fails here instead of on
# a user's machine. release.yml requires this for every binary it publishes;
# ci.yml runs it for every binary that carries a solver.
set -euo pipefail

usage() { echo "usage: $0 GOANIME_BINARY BUNDLED_SOLVER [--in-image IMAGE]" >&2; exit 2; }
[ $# -eq 2 ] || [ $# -eq 4 ] || usage
goanime="$1"; bundled="$2"; image=""
if [ $# -eq 4 ]; then
  [ "$3" = "--in-image" ] && [ -n "$4" ] || usage
  image="$4"
fi
[ -x "$goanime" ] || { echo "error: $goanime is not an executable" >&2; exit 1; }
[ -s "$bundled" ] || { echo "error: no bundled solver at $bundled" >&2; exit 1; }

home="$(mktemp -d)"
trap 'rm -rf "$home"' EXIT

case "$(uname -s)" in
  Linux)
    solver="$home/.config/goanime/cfsolve/challenge"
    env -i HOME="$home" PATH=/usr/bin:/bin TERM=dumb "$goanime" --version >/dev/null
    ;;
  Darwin)
    solver="$home/Library/Application Support/goanime/cfsolve/challenge"
    env -i HOME="$home" PATH=/usr/bin:/bin TERM=dumb "$goanime" --version >/dev/null
    ;;
  MINGW* | MSYS* | CYGWIN*)
    # Windows programs need the rest of the environment (SYSTEMROOT, ...);
    # only the folders GoAnime writes to are moved.
    mkdir -p "$home/AppData/Roaming" "$home/AppData/Local"
    solver="$home/AppData/Roaming/goanime/cfsolve/challenge.exe"
    APPDATA="$(cygpath -w "$home/AppData/Roaming")" \
      LOCALAPPDATA="$(cygpath -w "$home/AppData/Local")" \
      GOANIME_TERMINAL=conhost "$goanime" --version >/dev/null
    ;;
  *) echo "error: unsupported system $(uname -s)" >&2; exit 1 ;;
esac

[ -s "$solver" ] || { echo "error: GoAnime did not unpack a solver at $solver" >&2; exit 1; }
cmp -s "$solver" "$bundled" || { echo "error: the unpacked solver is not the bundled one" >&2; exit 1; }
"$solver" --help >/dev/null || { echo "error: the unpacked solver does not start" >&2; exit 1; }
echo "ok: $(basename "$goanime") unpacked its solver and it starts"

if [ -n "$image" ]; then
  dir="$(cd "$(dirname "$goanime")" && pwd)"
  name="$(basename "$goanime")"
  docker run --rm -v "$dir:/goanime:ro" "$image" sh -c '
    set -e
    export HOME="$(mktemp -d)"
    /goanime/'"$name"' --version >/dev/null
    "$HOME/.config/goanime/cfsolve/challenge" --help >/dev/null
  ' || { echo "error: GoAnime or its solver does not start in $image" >&2; exit 1; }
  echo "ok: both start in $image"
fi
