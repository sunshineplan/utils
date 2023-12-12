package html

import (
	"fmt"
	"html"
	"strings"
)

var (
	EscapeString   = html.EscapeString
	UnescapeString = html.UnescapeString
)

type Attribute struct {
	Name  string
	Value string
}

func Attributes(pairs ...string) (attributes []Attribute) {
	if len(pairs)%2 != 0 {
		panic("pairs must have even number of elements")
	}
	for i := 0; i < len(pairs); i = i + 2 {
		attributes = append(attributes, Attribute{pairs[i], pairs[i+1]})
	}
	return
}

func (attr Attribute) String() string {
	if attr.Value == "" || attr.Value == "true" {
		return attr.Name
	}
	return fmt.Sprintf("%s=%q", attr.Name, attr.Value)
}

type HTML string

func Element[T HTML | string](tag string, attributes []Attribute, content T) HTML {
	var b strings.Builder
	fmt.Fprint(&b, "<", tag)
	for _, i := range attributes {
		fmt.Fprint(&b, " ", i)
	}
	fmt.Fprint(&b, ">")
	switch any(content).(type) {
	case HTML:
		fmt.Fprint(&b, content)
	default:
		fmt.Fprint(&b, EscapeString(string(content)))
	}
	fmt.Fprintf(&b, "</%s>", tag)
	return HTML(b.String())
}
