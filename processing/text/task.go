package text

import (
	"regexp"
	"strings"
)

type Tasks struct {
	tasks []Processor
}

func NewTasks(tasks ...Processor) *Tasks {
	return &Tasks{tasks}
}

func (t *Tasks) Process(s string) (string, error) {
	var err error
	for first, output := true, s; ; {
		for _, task := range t.tasks {
			if first || !task.Once() {
				s, err = task.Process(s)
				if err != nil {
					return "", err
				}
			}
		}
		if s == output {
			return output, nil
		} else {
			output = s
		}
		if first {
			first = false
		}
	}
}

func (t *Tasks) ProcessAll(s []string) ([]string, error) {
	var output []string
	for _, i := range s {
		s, err := t.Process(i)
		if err != nil {
			return nil, err
		}
		output = append(output, s)
	}
	return output, nil
}

func (t *Tasks) Append(tasks ...Processor) *Tasks {
	t.tasks = append(t.tasks, tasks...)
	return t
}

func (t *Tasks) TrimSpace() *Tasks {
	return t.Append(NewProcessor(false, WrapFunc(strings.TrimSpace)))
}

func (t *Tasks) CutSpace() *Tasks {
	return t.Append(NewProcessor(true, func(s string) (string, error) {
		if fs := strings.Fields(s); len(fs) > 0 {
			return fs[0], nil
		}
		return "", nil
	}))
}

func (t *Tasks) RemoveParentheses() *Tasks {
	return t.Append(
		RemoveByRegexp{regexp.MustCompile(`\([^)]*\)`)},
		RemoveByRegexp{regexp.MustCompile(`（[^）]*）`)},
	)
}
