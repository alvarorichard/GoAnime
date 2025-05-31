package tracking

import (
	"fmt"
)

// HandleTrackingNotice displays a notice about tracking availability
func HandleTrackingNotice() {
	if !IsCgoEnabled {
		fmt.Println("Notice: Anime progress tracking disabled (CGO not available)")
		fmt.Println("Episode progress and resume features will not be available.")
		fmt.Println()
	}
}
