package choice

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func choiceStr(choice any) string {
	if v, ok := choice.(Description); ok {
		return v.Description()
	} else if v, ok := choice.(fmt.Stringer); ok {
		return v.String()
	}
	return fmt.Sprint(choice)
}

func Menu[E Choice](choices []E) string {
	var digit int
	for n := len(choices); n != 0; digit++ {
		n /= 10
	}
	option := fmt.Sprintf("%%%dd", digit)
	var b strings.Builder
	for i, choice := range choices {
		fmt.Fprintf(&b, "%s. %s\n", fmt.Sprintf(option, i+1), choiceStr(choice))
	}
	return b.String()
}

func choose[E Choice](choice string, choices []E) (res any, err error) {
	n, err := strconv.Atoi(choice)
	if err != nil {
		return nil, choiceError(choice)
	}
	if length := len(choices); n < 1 || n > length {
		return nil, choiceError(fmt.Sprintf("out of range(1-%d): %d", length, n))
	}
	return choices[n-1].Run()
}

func Choose[E Choice](choices []E) (choice string, res any, err error) {
	if length := len(choices); length == 0 {
		return
	}
	fmt.Print(Menu(choices))
	fmt.Print("\nPlease choose: ")
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Scan()
	b := bytes.TrimSpace(scanner.Bytes())
	if bytes.EqualFold(b, []byte("q")) || bytes.Contains(b, []byte{27}) {
		return
	}
	res, err = choose(string(b), choices)
	return
}
