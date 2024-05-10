package text

import (
	"regexp"
	"strings"
)

type Processor interface {
	Once() bool
	Process(string) (string, error)
}

var (
	_ Processor = processor{}
	_ Processor = RemoveByRegexp{}
	_ Processor = Cut{}
	_ Processor = Trim{}
)

type processor struct {
	once bool
	fn   func(string) (string, error)
}

func NewProcessor(once bool, fn func(string) (string, error)) Processor {
	return processor{once, fn}
}

func WrapFunc(fn func(string) string) func(string) (string, error) {
	return func(s string) (string, error) { return fn(s), nil }
}

func (p processor) Once() bool { return p.once }
func (p processor) Process(s string) (string, error) {
	return p.fn(s)
}

type RemoveByRegexp struct {
	*regexp.Regexp
}

func (RemoveByRegexp) Once() bool { return false }
func (p RemoveByRegexp) Process(s string) (string, error) {
	return p.ReplaceAllString(s, ""), nil
}

type Cut struct {
	Sep string
}

func (Cut) Once() bool { return true }
func (p Cut) Process(s string) (string, error) {
	before, _, _ := strings.Cut(s, p.Sep)
	return before, nil
}

type Trim struct {
	Cutset string
}

func (Trim) Once() bool { return false }
func (p Trim) Process(s string) (string, error) {
	return strings.Trim(s, p.Cutset), nil
}
