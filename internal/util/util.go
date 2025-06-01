package util

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"runtime"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/manifoldco/promptui"
)

var (
	IsDebug       bool
	minNameLength = 4

	// Style definitions for help and errors
	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true).
			Underline(true)

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#4ECDC4")).
			Bold(true)

	optionStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#45B7D1")).
			Italic(true)

	exampleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#96CEB4")).
			Italic(true)

	// Error styling
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF4757")).
			Bold(true)

	debugErrorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B")).
			Bold(true).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#FF4757")).
			Padding(1, 2)

	warningStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFA726")).
			Bold(true)

	// Additional styling for prompts and success messages
	promptStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF69B4")).
			Bold(true)

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#00FF00")).
			Bold(true)
)

// SetDebugMode sets the debug mode
func SetDebugMode(debug bool) {
	IsDebug = debug
}

// GetAnimeName gets the anime name from command line arguments or user input
func GetAnimeName() (string, error) {
	// If the user has provided an anime name as an argument, we use it.
	var animeName string
	if len(flag.Args()) > 0 {
		animeName = strings.Join(flag.Args(), " ")
		// Check if it has some flags and remove them
		if strings.Contains(animeName, "-") {
			animeName = strings.Split(animeName, "-")[0]
		}
		// Display anime name with beautiful styling
		animeDisplay := titleStyle.Render("🎯 Target Anime: " + animeName)
		fmt.Println(animeDisplay)

		if len(animeName) < minNameLength {
			return "", fmt.Errorf("anime name must have at least %d characters, you entered: %v", minNameLength, animeName)
		}
		return TreatingAnimeName(animeName), nil
	}

	// Enhanced prompt with styling
	promptHeader := helpStyle.Render("🔍 Search for Anime")
	fmt.Println(promptHeader)

	animeName, err := getUserInput("Enter anime name")
	return TreatingAnimeName(animeName), err
}

// ErrorHandler returns a stylized error message with beautiful formatting
func ErrorHandler(err error) string {
	if IsDebug {
		// Create a beautiful debug error display with full details
		errorIcon := "🚨"
		debugIcon := "🔍"

		errorMessage := fmt.Sprintf("%s %s %s", errorIcon, "DEBUG ERROR", debugIcon)
		fullError := fmt.Sprintf("%+v", err)

		styledHeader := errorStyle.Render(errorMessage)
		styledError := debugErrorStyle.Render(fullError)

		return fmt.Sprintf("%s\n%s", styledHeader, styledError)
	} else {
		// Create a clean, styled error message for normal users
		errorIcon := "❌"
		hintIcon := "💡"

		baseError := fmt.Sprintf("%v", err)
		hint := "run the program with -debug to see details"

		styledError := errorStyle.Render(fmt.Sprintf("%s %s", errorIcon, baseError))
		styledHint := warningStyle.Render(fmt.Sprintf("%s %s", hintIcon, hint))

		return fmt.Sprintf("%s\n%s", styledError, styledHint)
	}
}

// Helper prints the help message
func Helper() {
	title := titleStyle.Render("🎌 GoAnime - Your Anime Streaming Companion")

	usage := helpStyle.Render("📖 Usage:")
	usageExamples := []string{
		"  goanime",
		"  goanime " + optionStyle.Render("[options]"),
		"  goanime " + optionStyle.Render("[options]") + " " + exampleStyle.Render("[anime name]"),
	}

	note := helpStyle.Render("📝 Note:") + " Don't use - in anime names, use spaces instead"
	example := "  Example: " + exampleStyle.Render("\"one piece\"") + " instead of " + exampleStyle.Render("\"one-piece\"")

	options := helpStyle.Render("⚙️  Options:")
	optionsList := []string{
		"  " + optionStyle.Render("-debug") + "    🐛 Enable debug mode with detailed information",
		"  " + optionStyle.Render("-help, -h") + " 📚 Show this help message",
		"  " + optionStyle.Render("-version") + "  ℹ️  Show version information",
	}

	fmt.Println(title)
	fmt.Println()
	fmt.Println(usage)
	for _, line := range usageExamples {
		fmt.Println(line)
	}
	fmt.Println()
	fmt.Println(note)
	fmt.Println(example)
	fmt.Println()
	fmt.Println(options)
	for _, line := range optionsList {
		fmt.Println(line)
	}
	fmt.Println()
}

// getUserInput prompts the user for input the anime name and returns it
func getUserInput(label string) (string, error) {
	// Use simpler input method on Windows to avoid readline ANSI issues
	if runtime.GOOS == "windows" {
		return getSimpleInput(label)
	}

	// Create styled prompt for non-Windows systems
	styledLabel := promptStyle.Render("🎮 " + label)

	prompt := promptui.Prompt{
		Label: styledLabel,
	}

	animeName, err := prompt.Run()
	if err != nil {
		return "", err
	}
	if len(animeName) < minNameLength {
		return "", fmt.Errorf("anime name must have at least %d characters, you entered: %v", minNameLength, animeName)
	}

	// Display success message
	successMsg := successStyle.Render("✓ Anime name received: " + animeName)
	fmt.Println(successMsg)

	return animeName, nil
}

// getSimpleInput provides a fallback input method for Windows
func getSimpleInput(label string) (string, error) {
	// Display styled prompt
	styledLabel := promptStyle.Render("🎮 " + label + ": ")
	fmt.Print(styledLabel)

	// Use simple buffered reader
	reader := bufio.NewReader(os.Stdin)
	animeName, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	// Clean up the input
	animeName = strings.TrimSpace(animeName)
	if len(animeName) < minNameLength {
		return "", fmt.Errorf("anime name must have at least %d characters, you entered: %v", minNameLength, animeName)
	}

	// Display success message
	successMsg := successStyle.Render("✓ Anime name received: " + animeName)
	fmt.Println(successMsg)

	return animeName, nil
}

// SelectMenuItem provides a cross-platform way to select from a menu
// Windows-compatible alternative to promptui.Select
func SelectMenuItem(label string, items []string) (int, string, error) {
	// Use simpler input method on Windows to avoid readline ANSI issues
	if runtime.GOOS == "windows" {
		return simpleSelectMenu(label, items)
	}

	// Create styled prompt for non-Windows systems
	styledLabel := promptStyle.Render(label)
	prompt := promptui.Select{
		Label: styledLabel,
		Items: items,
	}

	index, result, err := prompt.Run()
	if err != nil {
		return -1, "", err
	}

	// Display success message
	successMsg := successStyle.Render("✓ Selected: " + result)
	fmt.Println(successMsg)

	return index, result, nil
}

// simpleSelectMenu provides a simple menu selection for Windows systems
func simpleSelectMenu(label string, items []string) (int, string, error) {
	// Display styled prompt and menu options
	styledLabel := promptStyle.Render(label)
	fmt.Println(styledLabel)

	// Display all options with numbers
	for i, item := range items {
		fmt.Printf("%d. %s\n", i+1, item)
	}

	// Get user input
	fmt.Print(promptStyle.Render("Enter selection (1-" + fmt.Sprintf("%d", len(items)) + "): "))
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return -1, "", err
	}

	// Parse selection
	input = strings.TrimSpace(input)
	var selection int
	_, err = fmt.Sscanf(input, "%d", &selection)
	if err != nil || selection < 1 || selection > len(items) {
		return -1, "", fmt.Errorf("invalid selection: %s", input)
	}

	// Adjust for 0-based index
	selection--

	// Display success message
	successMsg := successStyle.Render("✓ Selected: " + items[selection])
	fmt.Println(successMsg)

	return selection, items[selection], nil
}

// TreatingAnimeName removes special characters and spaces from the anime name.
func TreatingAnimeName(animeName string) string {
	loweredName := strings.ToLower(animeName)
	return strings.ReplaceAll(loweredName, " ", "-")
}
