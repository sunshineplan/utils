package utils

import (
	"fmt"
	"strings"
)

// Confirm asks the user for confirmation.
func Confirm(prompt string, attempts int) bool {
	if prompt == "" {
		prompt = "Are you sure?"
	}
	if attempts <= 0 {
		attempts = 3
	}

	fmt.Print(prompt, " (yes/no): ")
	var input string
	for ; attempts > 0; attempts-- {
		if _, err := fmt.Scanln(&input); err != nil {
			fmt.Println(err)
		}
		switch strings.ToLower(strings.TrimSpace(input)) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			if attempts > 1 {
				fmt.Print("Please type 'yes' or 'no': ")
			}
		}
	}
	fmt.Println("Max retries exceeded.")
	return false
}
