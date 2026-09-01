package version

import (
	"fmt"
	"os"
	"strings"

	"github.com/alvarorichard/Goanime/internal/tracking"
)

// Version is set via -ldflags at build time by the CI workflow.
// Fallback value is used for local development builds.
//
// The release workflow injects this from `${GITHUB_REF#refs/tags/}`, which
// keeps the leading `v` from the git tag (e.g. `v1.8.4`). Code that prints
// the version uses a `v%s` format, so without normalization CI builds log
// `vv1.8.4`. Strip the prefix at init so both injected and fallback values
// are stored without it.
var Version = "1.8.7"

func init() {
	Version = strings.TrimPrefix(Version, "v")
}

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
