package choice

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
)

// Description interface defines a method for obtaining the description of a choice.
type Description interface {
	Description() string
}

func choiceStr(choice any) string {
	if v, ok := choice.(Description); ok {
		return v.Description()
	} else if v, ok := choice.(fmt.Stringer); ok {
		return v.String()
	}
	return fmt.Sprint(choice)
}

// Menu function is used to generate a string representation of the menu, associating choices with numbers.
func Menu[E any](choices []E, showQuit bool) string {
	if len(choices) == 0 {
		return ""
	}
	var digit int
	for n := len(choices); n != 0; digit++ {
		n /= 10
	}
	option := fmt.Sprintf("%%%dd", digit)
	var b strings.Builder
	for i, choice := range choices {
		fmt.Fprintf(&b, "%s. %s\n", fmt.Sprintf(option, i+1), choiceStr(choice))
	}
	if showQuit {
		fmt.Fprintf(&b, "%s. Quit\n", fmt.Sprintf(fmt.Sprintf("%%%ds", digit), "q"))
	}
	return b.String()
}

// ErrBadChoice defines an error that represents an invalid choice made by the user.
var ErrBadChoice = errors.New("bad choice")

var _ error = choiceError("")

type choiceError string

func (err choiceError) Error() string {
	return "bad choice: " + string(err)
}

func (choiceError) Unwrap() error {
	return ErrBadChoice
}

func choose[E any](choice string, choices []E) (res E, err error) {
	n, err := strconv.Atoi(choice)
	if err != nil {
		err = choiceError(choice)
		return
	}
	if length := len(choices); n < 1 || n > length {
		err = choiceError(fmt.Sprintf("out of range(1-%d): %d", length, n))
		return
	}
	return choices[n-1], nil
}

// Choose function allows the user to make a choice from the given options with no default value.
func Choose[E any](choices []E) (choice bool, res E, err error) {
	return ChooseWithDefault(choices, 0)
}

// ChooseWithDefault function allows the user to make a choice from the given options with an optional default value.
func ChooseWithDefault[E any](choices []E, def int) (choice bool, res E, err error) {
	if n := len(choices); n == 0 {
		err = errors.New("no choices")
		return
	} else if def > n {
		err = errors.New("invalid default choice")
		return
	}
	var prompt string
	if def > 0 {
		prompt = fmt.Sprintf("Please choose (default: %d): ", def)
	} else {
		prompt = "Please choose: "
	}
	scanner := bufio.NewScanner(os.Stdin)
	var b []byte
	if def <= 0 {
		for len(b) == 0 {
			fmt.Print(prompt)
			scanner.Scan()
			b = bytes.TrimSpace(scanner.Bytes())
		}
	} else {
		fmt.Print(prompt)
		scanner.Scan()
		b = bytes.TrimSpace(scanner.Bytes())
		if len(b) == 0 {
			return true, choices[def-1], nil
		}
	}
	if bytes.EqualFold(b, []byte("q")) {
		return
	}
	choice = true
	res, err = choose(string(b), choices)
	return
}
