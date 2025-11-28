package text

import "fmt"

var MaxIter = 100

// Tasks represents an ordered list of text processors.
// Each processor in the list will be executed sequentially,
// and repeated until the text no longer changes.
type Tasks struct {
	tasks []Processor
}

// NewTasks creates a new Tasks instance with the provided processors.
func NewTasks(tasks ...Processor) *Tasks {
	return &Tasks{tasks}
}

// Process executes all configured processors on a single string.
func (t *Tasks) Process(s string) (output string, err error) {
	output = s
	first := true
	for range MaxIter {
		for _, task := range t.tasks {
			if first || !task.Once() {
				s, err = task.Process(s)
				if err != nil {
					return "", fmt.Errorf("%s error: %w", task.Describe(), err)
				}
			}
		}
		if s == output {
			return
		}
		output = s
		if first {
			first = false
		}
	}
	err = fmt.Errorf("max iteration limit reached")
	return
}

// ProcessAll applies all processors to a slice of strings, returning the processed results.
// If any processor returns an error, processing stops and the error is returned.
func (t *Tasks) ProcessAll(inputs []string) ([]string, error) {
	output := make([]string, len(inputs))
	for i, in := range inputs {
		out, err := t.Process(in)
		if err != nil {
			return nil, err
		}
		output[i] = out
	}
	return output, nil
}

// Append adds one or more processors to the task list and returns the updated instance.
func (t *Tasks) Append(tasks ...Processor) *Tasks {
	t.tasks = append(t.tasks, tasks...)
	return t
}
