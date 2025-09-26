package confirm

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

// Do asks the user for confirmation.
func Do(prompt string, attempts int) bool {
	return do(prompt, attempts, os.Stdout, os.Stdin)
}

func do(prompt string, attempts int, w io.Writer, r io.Reader) bool {
	if prompt == "" {
		prompt = "Are you sure?"
	}
	if attempts <= 0 {
		attempts = 3
	}

	if _, err := fmt.Fprintf(w, "%s (yes/no): ", prompt); err != nil {
		fmt.Println("Error writing to output:", err)
		return false
	}
	br := bufio.NewReader(r)
	for ; attempts > 0; attempts-- {
		input, err := br.ReadString('\n')
		if err != nil {
			fmt.Fprintln(w, "Error reading input:", err)
			continue
		}
		switch strings.ToLower(strings.TrimSpace(input)) {
		case "y", "yes":
			return true
		case "n", "no":
			return false
		default:
			if attempts > 1 {
				fmt.Fprint(w, "Please type 'yes' or 'no': ")
			}
		}
	}
	fmt.Fprintln(w, "Max retries exceeded.")
	return false
}
