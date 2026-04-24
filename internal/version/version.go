package version

import (
	"fmt"
	"os"

	"github.com/alvarorichard/Goanime/internal/tracking"
)

// Version is set via -ldflags at build time by the CI workflow.
// Fallback value is used for local development builds.
var Version = "1.8.2"

// BuildTime and Commit are injected by the CI workflow via -ldflags.
var (
	BuildTime = "unknown"
	Commit    = "unknown"
)

func HasVersionArg() bool {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		return arg == "--version" || arg == "-version" || arg == "-v" || arg == "--v" || arg == " version"
	}
	return false
}

func ShowVersion() {
	fmt.Printf("GoAnime v%s", Version)
	if tracking.IsCgoEnabled {
		fmt.Println(" (with SQLite tracking)")
	} else {
		fmt.Println(" (without SQLite tracking)")
	}
}
