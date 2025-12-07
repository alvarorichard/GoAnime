package version

import (
	"fmt"
	"os"

	"github.com/alvarorichard/Goanime/internal/tracking"
)

const (
	Version = "1.5"
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
