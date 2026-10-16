# Sourced by demo.tape before the recording starts; not meant to be run on
# its own. Leaves the shell ready to run the demo against server.go, with
# nothing that can reach a real source or play sound, and seeds the
# throwaway history (seed.go) so the home screen shows CONTINUE and RECENT
# examples. demo_cleanup undoes it.

DEMO_DIR="$(mktemp -d)"
mkdir -p "$DEMO_DIR/bin" "$DEMO_DIR/home" "$DEMO_DIR/tmp"

# GoAnime from this checkout, so the recording shows the code it documents,
# and the stand-in API. Both are built before HOME changes so the Go build
# cache is reused.
go build -o "$DEMO_DIR/bin/goanime" ./cmd/goanime || return 1
go build -o "$DEMO_DIR/server" docs/demo/server.go || return 1

# Every title plays this: a test pattern with no audio track, cut into an
# HLS playlist because the AnimeFire client only accepts a stream URL that
# answers with one.
mkdir -p "$DEMO_DIR/hls"
ffmpeg -loglevel error -f lavfi -i testsrc2=size=320x180:rate=24:duration=120 \
	-an -pix_fmt yuv420p -f hls -hls_time 6 -hls_playlist_type vod \
	-hls_segment_filename "$DEMO_DIR/hls/seg%03d.ts" "$DEMO_DIR/hls/index.m3u8" || return 1

# mpv with audio and video output disabled.
printf '#!/bin/sh\nexec %s "$@" --ao=null --mute=yes --vo=null --force-window=no\n' \
	"$(command -v mpv)" > "$DEMO_DIR/bin/mpv"
chmod +x "$DEMO_DIR/bin/mpv"

"$DEMO_DIR/server" -addr 127.0.0.1:8765 -media "$DEMO_DIR/hls" 2>/dev/null &
DEMO_SERVER_PID=$!
disown
until curl -sf -o /dev/null http://127.0.0.1:8765/anime/namakura-gatana; do sleep 0.2; done

# A throwaway HOME and TMPDIR: the demo leaves no watch history, logs or
# caches behind, and does not show up as Discord presence.
export HOME="$DEMO_DIR/home" TMPDIR="$DEMO_DIR/tmp" PATH="$DEMO_DIR/bin:$PATH"

# Seed the throwaway history (see seed.go): the home screen opens on a
# lived-in CONTINUE list and a few RECENT searches instead of an empty box.
go run docs/demo/seed.go || return 1

# Only the stand-in source: AnimeFire, pointed at server.go.
export GOANIME_ANIMEFIRE_API=http://127.0.0.1:8765
goanime() { command goanime --source animefire "$@"; }

demo_cleanup() {
	kill "$DEMO_SERVER_PID" 2>/dev/null
	rm -rf "$DEMO_DIR"
}

PS1='$ '
