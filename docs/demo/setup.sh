# Sourced by demo.tape before the recording starts; not meant to be run on
# its own. Leaves the shell ready to run the demo against server.go, with
# nothing that can reach a real source or play sound. demo_cleanup undoes it.

DEMO_DIR="$(mktemp -d)"
mkdir -p "$DEMO_DIR/bin" "$DEMO_DIR/home" "$DEMO_DIR/tmp"

# GoAnime from this checkout, so the recording shows the code it documents,
# and the stand-in API. Both are built before HOME changes so the Go build
# cache is reused.
go build -o "$DEMO_DIR/bin/goanime" ./cmd/goanime || return 1
go build -o "$DEMO_DIR/server" docs/demo/server.go || return 1

# Every title plays this: a test pattern with no audio track.
ffmpeg -loglevel error -f lavfi -i testsrc2=size=320x180:rate=24:duration=120 \
	-an -pix_fmt yuv420p "$DEMO_DIR/demo.mp4" || return 1

# mpv with audio and video output disabled.
printf '#!/bin/sh\nexec %s "$@" --ao=null --mute=yes --vo=null --force-window=no\n' \
	"$(command -v mpv)" > "$DEMO_DIR/bin/mpv"
chmod +x "$DEMO_DIR/bin/mpv"

"$DEMO_DIR/server" -addr 127.0.0.1:8765 -media "$DEMO_DIR/demo.mp4" 2>/dev/null &
DEMO_SERVER_PID=$!
disown
until curl -sf -o /dev/null http://127.0.0.1:8765/anime/namakura-gatana; do sleep 0.2; done

# A throwaway HOME and TMPDIR: the demo leaves no watch history, logs or
# caches behind, and does not show up as Discord presence.
export HOME="$DEMO_DIR/home" TMPDIR="$DEMO_DIR/tmp" PATH="$DEMO_DIR/bin:$PATH"

# Only the stand-in source: AnimeFire, pointed at server.go.
export GOANIME_ANIMEFIRE_API=http://127.0.0.1:8765
goanime() { command goanime --source animefire "$@"; }

demo_cleanup() {
	kill "$DEMO_SERVER_PID" 2>/dev/null
	rm -rf "$DEMO_DIR"
}

PS1='$ '
