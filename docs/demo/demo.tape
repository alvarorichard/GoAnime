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
Set Height 760
Set Padding 24
Set Framerate 30
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

# Tela 1, the SEARCH home screen. The query box placeholder is the text that
# is always there. setup.sh seeds CONTINUE and RECENT examples, so first
# drop to the continue list and back: it shows the row highlight and its
# details panel without leaving the box behind.
Wait+Screen@10s /title, or/
Sleep 800ms
Down
Sleep 1.2s
Up
Sleep 500ms
Type "momotarou"
Sleep 400ms
Enter

# Results: Momotarou matches first, so Down 2 / Up 2 still lands on it.
Wait+Screen@30s /RESULTS/
Sleep 1.5s
Down 2
Sleep 600ms
Up 2
Sleep 800ms
Enter

# Episodes (tela 5): Momotarou lists three parts. m opens the OPTIONS menu
# for the highlighted episode instead of playing it straight away.
Wait+Screen@60s /EPISODES/
Sleep 1.5s
Down
Sleep 600ms
Up
Sleep 600ms
Type "m"

# Options (tela 6): Watch now is first.
Wait+Screen@30s /Watch now/
Sleep 1.2s
Enter

# NOW PLAYING. Linger so the progress bar motion reads instead of
# flickering past.
Wait+Screen@60s /NOW PLAYING/
Sleep 5s
Type "q"
Wait+Line@15s /^\$\s*$/
Sleep 1.5s

Hide
Type "demo_cleanup"
Enter
Sleep 1s
