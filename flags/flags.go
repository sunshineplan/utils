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

var Strict bool

var config string

func SetConfigFile(path string) { config = path }

func getArgs(strict bool) (args []string) {
	lines, err := txt.ReadFile(config)
	if err != nil {
		if !strict && errors.Is(err, fs.ErrNotExist) {
			fmt.Println("config file is missing")
			return
		}
		panic(err)
	}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			panic(fmt.Sprintf("cannot parse %q", line))
		}
		if key := strings.TrimSpace(parts[0]); flag.Lookup(key) != nil {
			args = append(args, fmt.Sprintf("-%s=%s", key, unquote(parts[1])))
		} else {
			if err := fmt.Sprintf("undefined flag %q", key); strict {
				panic(err)
			} else {
				fmt.Println("[Warning]", err)
			}
		}
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

func parse(strict bool) {
	if config != "" {
		flag.CommandLine.Parse(append(getArgs(strict), os.Args[1:]...))
		return
	}
	flag.Parse()
}

func UseStrict(strict bool) {
	Strict = strict
}

func Parse() {
	parse(Strict)
}

func ParseStrict() {
	parse(true)
}
