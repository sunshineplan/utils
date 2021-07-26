package txt

import (
	"bufio"
	"io"
	"os"
)

// ReadAll reads all content from r.
func ReadAll(r io.Reader) ([]string, error) {
	var content []string
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		content = append(content, scanner.Text())
	}

	return content, scanner.Err()
}

// ReadFile reads all content from file.
func ReadFile(file string) ([]string, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}

	return ReadAll(f)
}
