package text

import (
	"fmt"
	"regexp"
	"strings"
)

// Processor defines a generic interface for text processors.
// Each processor can optionally be executed only once (Once == true)
// and provides a human-readable description for debugging or logging.
type Processor interface {
	// Describe returns a short description of this processor.
	Describe() string
	// Once reports whether this processor should run only once.
	Once() bool
	// Process performs the actual text transformation.
	Process(string) (string, error)
}

var (
	_ Processor = new(processor)
	_ Processor = new(multiProcessor)
	_ Processor = RegexpRemover{}
	_ Processor = Cutter{}
	_ Processor = Trimmer{}
)

// processor is a generic implementation of Processor,
// allowing you to wrap any custom text function.
type processor struct {
	desc string
	once bool
	fn   func(string) (string, error)
}

// NewProcessor creates a new Processor from a function.
//
//	desc - short description for debugging
//	once - whether this processor should be executed only once
//	fn   - transformation function taking a string and returning a string/error
func NewProcessor(desc string, once bool, fn func(string) (string, error)) Processor {
	return &processor{desc, once, fn}
}

// Describe returns the processor's description string.
func (p *processor) Describe() string { return p.desc }

// Once reports whether this processor should run only once.
func (p *processor) Once() bool { return p.once }

// Process executes the wrapped function to transform the text.
func (p *processor) Process(s string) (string, error) {
	return p.fn(s)
}

// multiProcessor executes multiple sub-processors as a single Processor.
type multiProcessor struct {
	desc       string
	once       bool
	processors []Processor
}

// NewMultiProcessor creates a new MultiProcessor.
//
//	desc  - human-readable name
//	once  - whether this processor should execute only once
//	procs - list of sub-processors
func NewMultiProcessor(desc string, once bool, procs ...Processor) Processor {
	return &multiProcessor{desc, once, procs}
}

// Describe returns the description for debugging or logging.
func (m *multiProcessor) Describe() string { return m.desc }

// Once reports whether the MultiProcessor should run only once.
func (m *multiProcessor) Once() bool { return m.once }

// Process executes all sub-processors.
func (m *multiProcessor) Process(s string) (string, error) {
	for _, p := range m.processors {
		var err error
		s, err = p.Process(s)
		if err != nil {
			return "", err
		}
	}
	return s, nil
}

// RegexpRemover removes substrings that match the given regular expression.
type RegexpRemover struct {
	Re *regexp.Regexp
}

// Describe returns a string representation of the RegexpRemover.
func (p RegexpRemover) Describe() string { return fmt.Sprintf("RegexpRemover(%q)", p.Re.String()) }

// Once always returns false, meaning this processor can be applied repeatedly.
func (RegexpRemover) Once() bool { return false }

// Process removes all matches of the regular expression from the input string.
func (p RegexpRemover) Process(s string) (string, error) {
	return p.Re.ReplaceAllString(s, ""), nil
}

// Cutter splits the input by the given separator and keeps only the part before it.
type Cutter struct {
	Sep string
}

// Describe returns a string representation of the Cutter.
func (p Cutter) Describe() string { return fmt.Sprintf("Cutter(%q)", p.Sep) }

// Once always returns true, meaning this processor should be run only once.
func (Cutter) Once() bool { return true }

// Process cuts the string at the first occurrence of the separator and returns the left part.
func (p Cutter) Process(s string) (string, error) {
	before, _, _ := strings.Cut(s, p.Sep)
	return before, nil
}

// Trimmer removes all leading and trailing characters from the given cutset.
type Trimmer struct {
	Cutset string
}

// Describe returns a string representation of the Trimmer.
func (p Trimmer) Describe() string { return fmt.Sprintf("Trimmer(%q)", p.Cutset) }

// Once always returns false, meaning this processor can be applied repeatedly.
func (Trimmer) Once() bool { return false }

// Process trims all leading and trailing characters in Cutset from the input string.
func (p Trimmer) Process(s string) (string, error) {
	return strings.Trim(s, p.Cutset), nil
}

// TrimSpace returns a processor that removes leading and trailing spaces.
func TrimSpace() Processor {
	return NewProcessor("TrimSpace", false, WrapFunc(strings.TrimSpace))
}

// CutSpace returns a processor that extracts the first word in the input string.
func CutSpace() Processor {
	return NewProcessor("CutSpace", true, func(s string) (string, error) {
		if fs := strings.Fields(s); len(fs) > 0 {
			return fs[0], nil
		}
		return "", nil
	})
}

// RemoveParentheses returns a processor that remove both western and full-width parentheses.
func RemoveParentheses() Processor {
	return NewMultiProcessor(
		"RemoveParentheses",
		true,
		RegexpRemover{regexp.MustCompile(`\([^\)]*\)`)},
		RegexpRemover{regexp.MustCompile(`（[^）]*）`)},
	)
}

// WrapFunc wraps a simple string -> string function
// into a function matching the Processor signature.
func WrapFunc(fn func(string) string) func(string) (string, error) {
	return func(s string) (string, error) { return fn(s), nil }
}
