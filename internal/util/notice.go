package util

import "github.com/alvarorichard/Goanime/v2/internal/tui"

// NoticeOnNextPicker makes the next list screen (the episode list, a season
// menu) show text above its list, once. It is for a source that has something
// to tell the user about the list they are about to see: text printed to the
// terminal first is hidden behind the screen the moment it opens.
var NoticeOnNextPicker = tui.NoticeOnNextPicker
