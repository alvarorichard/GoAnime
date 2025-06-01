package playback

import (
	"fmt"
	"log"

	"github.com/alvarorichard/Goanime/internal/util"
)

func GetUserInput() string {
	fmt.Print("Press 'n' for next episode, 'p' for previous episode, 'e' to select episode, 'q' to quit: ")
	var input string
	_, err := fmt.Scanln(&input)
	if err != nil {
		if err.Error() == "unexpected newline" {
			log.Println("No input detected, continuing playback")
			return "n"
		}
		log.Printf("Error reading input: %v - defaulting to continue", util.ErrorHandler(err))
		return "n"
	}
	return input
}
