package playback

import (
	"fmt"
	"log"

	"github.com/alvarorichard/Goanime/internal/util"
)

func GetUserInput() string {
	// Display basic prompt
	fmt.Println("Playback Control")
	fmt.Print("'n' next | 'p' previous | 'e' select episode | 'q' quit > ")

	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		if err.Error() == "unexpected newline" {
			fmt.Println("No input detected, continuing to next episode")
			return "n"
		}
		fmt.Println("Error reading input - defaulting to continue")
		log.Printf("Error reading input: %v - defaulting to continue", util.ErrorHandler(err))
		return "n"
	}
	return input
}
