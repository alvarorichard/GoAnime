package util

import (
	"errors"
	"flag"
	"fmt"
	"github.com/manifoldco/promptui"
	"os"
	"strings"
)

var (
	IsDebug       bool
	minNameLength = 4
)

// ErrorHandler returns a string with the error message, if debug mode is enabled, it will return the full error with details.
func ErrorHandler(err error) string {
	if IsDebug {
		return fmt.Sprintf("%+v", err)
	} else {
		return fmt.Sprintf("%v -- run the program with -debug to see details", err)
	}
}

// Helper prints the help message
func Helper() {
	fmt.Print(`	Usage:
	goanime 
	goanime [options]
	goanime [options] [anime name] (don't use - in the anime name, use spaces instead, e.g: "one piece" instead of "one-piece")

	Options:
	   -debug: run the program in debug mode, which will show more details about errors and other information.
	   -help; -h; show this help message.
	`)
}

// FlagParser parses the -flags and returns the anime name
func FlagParser() (string, error) {
	// Define flags
	debug := flag.Bool("debug", false, "enable debug mode")
	help := flag.Bool("help", false, "show help message")
	altHelp := flag.Bool("h", false, "show help message")

	// Parse the flags early before any manipulation of os.Args
	flag.Parse()

	if *help || *altHelp {
		Helper()
		os.Exit(0)
	}

	IsDebug = *debug
	if *debug {
		fmt.Println("--- Debug mode is enabled ---")
	}
	//If the user has provided an anime name as an argument, we use it.
	var animeName string
	if len(flag.Args()) > 0 {
		animeName = strings.Join(flag.Args(), " ")
		// Check if it has some flags and remove them
		if strings.Contains(animeName, "-") {
			animeName = strings.Split(animeName, "-")[0]
		}
		fmt.Println("Anime name:", animeName)
		if len(animeName) < minNameLength {
			return "", errors.New(fmt.Sprintf("Anime name must have at least %d characters, you entered: %v", minNameLength, animeName))
		}
	} else {
		animeName, err := getUserInput("Enter anime name")
		if err != nil {
			return animeName, err

		}
	}
	return TreatingAnimeName(animeName), nil
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
		return "", errors.New(fmt.Sprintf("Anime name must have at least %d characters, you entered: %v", minNameLength, animeName))
	}
	return animeName, nil
}

// TreatingAnimeName removes special characters and spaces from the anime name.
func TreatingAnimeName(animeName string) string {
	loweredName := strings.ToLower(animeName)
	return strings.ReplaceAll(loweredName, " ", "-")
}
