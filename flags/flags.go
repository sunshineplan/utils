package flags

import (
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"os"
	"strconv"
	"strings"

	"github.com/sunshineplan/utils/txt"
)

var (
	// SilentMissingConfig controls whether a warning message is printed
	// when the specified configuration file is missing.
	SilentMissingConfig bool

	// config stores the path of the configuration file.
	config string
)

// errMissingConfig is returned when the specified configuration file does not exist.
var errMissingConfig = errors.New("config file is missing")

// SetConfigFile sets the path of the configuration file to be used when parsing flags.
func SetConfigFile(path string) { config = path }

// getArgs reads the configuration file and converts its key-value pairs
// into command-line style arguments compatible with the flag package.
// Lines beginning with '#' or empty lines are ignored.
// Each valid line must be in the form "key=value".
// Returns a slice of arguments or an error if parsing fails.
func getArgs() (args []string, err error) {
	lines, err := txt.ReadFile(config)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, errMissingConfig
		}
		return
	}
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			err = fmt.Errorf("line %d: cannot parse %q", i+1, line)
			return
		}
		if key := strings.TrimSpace(parts[0]); flag.Lookup(key) != nil {
			args = append(args, fmt.Sprintf("-%s=%s", key, unquote(parts[1])))
		} else {
			err = fmt.Errorf("line %d: unknown flag %q", i+1, key)
			return
		}
	}
	return
}

// unquote removes surrounding quotes from a string if present.
// If the string cannot be unquoted, it is returned unchanged.
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if unq, err := strconv.Unquote(s); err == nil {
		return unq
	}
	return s
}

// Parse parses command-line flags and optionally merges them with values
// read from the configuration file specified via SetConfigFile.
// If a configuration file is provided, its flags are prepended to os.Args.
// Errors during parsing or reading are handled according to the flag package’s
// ErrorHandling setting.
func Parse() error {
	if config != "" {
		args, err := getArgs()
		if err != nil {
			handleError(err)
		}
		return flag.CommandLine.Parse(append(args, os.Args[1:]...))
	}
	return flag.CommandLine.Parse(os.Args[1:])
}

// handleError handles errors according to their type
// and the current flag.CommandLine.ErrorHandling mode.
func handleError(err error) {
	switch err {
	case errMissingConfig:
		if !SilentMissingConfig {
			fmt.Println("[flags]", err)
		}
	default:
		switch flag.CommandLine.ErrorHandling() {
		case flag.ContinueOnError:
			fmt.Println("[flags]", err)
		case flag.ExitOnError:
			os.Exit(2)
		case flag.PanicOnError:
			panic(err)
		}
	}
}

// init reinitializes the default CommandLine flag set to use ContinueOnError mode,
// preventing flag.Parse from exiting the program automatically.
func init() {
	flag.CommandLine.Init(flag.CommandLine.Name(), flag.ContinueOnError)
}
