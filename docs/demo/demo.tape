# Records demo.gif, the animation shown at the top of the README.
#
# Requirements: vhs (https://github.com/charmbracelet/vhs), Go, mpv, ffmpeg
# and curl. GoAnime itself is built from this checkout. Run from the
# repository root:
#
#   vhs docs/demo/demo.tape
#
# The recording does not touch any real source. setup.sh points GoAnime at
# server.go, a local stand-in for the AnimeFire API whose catalog holds only
# animated shorts in the public domain. Every title plays a test pattern
# with no audio track, in an mpv started with audio and video output
# disabled, so nothing is heard or shown outside the terminal.

Output docs/demo/demo.gif

Set Shell bash
Set WindowBar Colorful
Set FontSize 18
Set Width 1200
Set Height 640
Set Padding 24
Set Framerate 24
Set TypingSpeed 80ms

Hide
Type "source docs/demo/setup.sh && clear"
Enter
Wait+Line@120s /^\$\s*$/
Show

Sleep 500ms
Type "goanime"
Sleep 300ms
Enter

Wait+Screen@10s /anime\/movie name/
Sleep 800ms
Type "namakura"
Sleep 400ms
Enter

# Results.
Wait+Screen@30s /Search > Results/
Sleep 1.5s
Down 2
Sleep 600ms
Up 2
Sleep 800ms
Enter

# Play online instead of downloading.
Wait+Screen@60s /play online/
Sleep 1.5s
Type "G"
Sleep 1s
Enter

# The film is playing in mpv; the terminal shows the player controls.
# Exit, the last entry, stops playback and quits.
Wait+Screen@60s /Player Controls/
Sleep 3s
Type "G"
Sleep 800ms
Enter
Wait+Line@15s /^\$\s*$/
Sleep 1.5s

Hide
Type "demo_cleanup"
Enter
Sleep 1s
