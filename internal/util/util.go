package util

import (
	"errors"
	"flag"
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/version"
	"github.com/manifoldco/promptui"
)

var (
	IsDebug          bool
	minNameLength    = 4
	ErrHelpRequested = errors.New("help requested") // Custom error for help
)

// ErrorHandler returns a string with the error message, if debug mode is enabled, it will return the full error with details.
func ErrorHandler(err error) string {
	if IsDebug {
		return fmt.Sprintf("%+v", err)
	} else {
		return fmt.Sprintf("%v -- run the program with -debug to see details", err)
	}
}

// Helper prints the beautiful help message
func Helper() {
	ShowBeautifulHelp()
}

// FlagParser parses the -flags and returns the anime name
func FlagParser() (string, error) {
	// Define flags
	debug := flag.Bool("debug", false, "enable debug mode")
	help := flag.Bool("help", false, "show help message")
	altHelp := flag.Bool("h", false, "show help message")
	versionFlag := flag.Bool("version", false, "show version information")

	// Parse the flags early before any manipulation of os.Args
	flag.Parse()

	// Set debug mode based on flag (set unconditionally for consistency)
	IsDebug = *debug

	if *versionFlag || version.HasVersionArg() {
		version.ShowVersion()
		return "", ErrHelpRequested // Signal version instead of exiting
	}

	if *help || *altHelp {
		Helper()
		return "", ErrHelpRequested // Signal help instead of exiting
	}

	if *debug {
		fmt.Println("--- Debug mode is enabled ---")
	}

	// If the user has provided an anime name as an argument, we use it.
	var animeName string
	if len(flag.Args()) > 0 {
		animeName = strings.Join(flag.Args(), " ")
		// Check if it has some flags and remove them
		if strings.Contains(animeName, "-") {
			animeName = strings.Split(animeName, "-")[0]
		}
		fmt.Println("Anime name:", animeName)
		if len(animeName) < minNameLength {
			return "", fmt.Errorf("anime name must have at least %d characters, you entered: %v", minNameLength, animeName)
		}
		return TreatingAnimeName(animeName), nil
	}
	animeName, err := getUserInput("Enter anime name")
	return TreatingAnimeName(animeName), err
}

// getUserInput prompts the user for input the anime name and returns it
func getUserInput(label string) (string, error) {
	prompt := promptui.Prompt{
		Label: label,
	}

	animeName, err := prompt.Run()
	if err != nil {
		return "", err
	}
	if len(animeName) < minNameLength {
		return "", fmt.Errorf("anime name must have at least %d characters, you entered: %v", minNameLength, animeName)
	}
	return animeName, nil
}

// TreatingAnimeName removes special characters and spaces from the anime name.
func TreatingAnimeName(animeName string) string {
	loweredName := strings.ToLower(animeName)
	return strings.ReplaceAll(loweredName, " ", "-")
}
