package strings

import (
	"strings"
)

// ContainsAnyOf returns whether the given string contains any of the following
// strings.
func ContainsAnyOf(str string, args ...string) bool {
	for _, arg := range args {
		if strings.Contains(str, arg) {
			return true
		}
	}
	return false
}
