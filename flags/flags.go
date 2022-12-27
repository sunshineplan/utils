package flags

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/sunshineplan/utils/txt"
)

var config string

func SetConfigFile(path string) { config = path }

func getArgs() (args []string) {
	lines, err := txt.ReadFile(config)
	if err != nil {
		panic(err)
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			panic(fmt.Sprintln("cannot parse", line))
		}
		args = append(args, fmt.Sprintf("-%s=%s", strings.TrimSpace(parts[0]), unquote(parts[1])))
	}
	return
}

func unquote(s string) string {
	s = strings.TrimSpace(s)
	if s, err := strconv.Unquote(s); err == nil {
		return s
	}
	return s
}

func Parse() {
	if config != "" {
		flag.CommandLine.Parse(append(getArgs(), os.Args[1:]...))
		return
	}
	flag.Parse()
}
