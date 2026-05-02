package version

import (
	"fmt"
	"os"
	"strings"

	"github.com/alvarorichard/Goanime/internal/tracking"
)

var (
	Version   = "1.7"
	BuildTime = ""
	Commit    = ""
)

func DisplayVersion() string {
	displayVersion := strings.TrimSpace(Version)
	displayVersion = strings.TrimPrefix(displayVersion, "v")
	displayVersion = strings.TrimPrefix(displayVersion, "V")

	if displayVersion == "" {
		return "unknown"
	}

	return displayVersion
}

func HasVersionArg() bool {
	if len(os.Args) > 1 {
		arg := os.Args[1]
		return arg == "--version" || arg == "-version" || arg == "-v" || arg == "--v" || arg == " version"
	}
	return false
}

func ShowVersion() {
	fmt.Printf("GoAnime v%s", DisplayVersion())
	if tracking.IsCgoEnabled {
		fmt.Println(" (with SQLite tracking)")
	} else {
		fmt.Println(" (without SQLite tracking)")
	}
}
