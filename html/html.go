package html

import "html"

var (
	// EscapeString is alias of [html.EscapeString],
	// used for encoding HTML entities.
	EscapeString = html.EscapeString
	// UnescapeString is alias of [html.UnescapeString],
	// used for decoding HTML entities.
	UnescapeString = html.UnescapeString
)

// HTML represents a string that contains valid HTML markup.
type HTML string

// HTMLer defines types that can render themselves as HTML.
type HTMLer interface {
	HTML() HTML
}

// Background creates an element with no tag, typically used for raw content.
func Background() *Element { return NewElement("") }

// NewHTML creates a new <html> element.
func NewHTML() *Element { return NewElement("html") }

// Common HTML element constructors for convenience.

func A() *Element     { return NewElement("a") }
func B() *Element     { return NewElement("b") }
func Body() *Element  { return NewElement("body") }
func Br() *Element    { return NewElement("br") }
func Div() *Element   { return NewElement("div") }
func Em() *Element    { return NewElement("em") }
func Form() *Element  { return NewElement("form") }
func H1() *Element    { return NewElement("h1") }
func H2() *Element    { return NewElement("h2") }
func Head() *Element  { return NewElement("head") }
func I() *Element     { return NewElement("i") }
func Img() *Element   { return NewElement("img") }
func Input() *Element { return NewElement("input") }
func Label() *Element { return NewElement("label") }
func Li() *Element    { return NewElement("li") }
func Meta() *Element  { return NewElement("meta") }
func P() *Element     { return NewElement("p") }
func Span() *Element  { return NewElement("span") }
func Style() *Element { return NewElement("style") }
func Svg() *Element   { return NewElement("svg") }
func Table() *Element { return NewElement("table") }
func Tbody() *Element { return NewElement("tbody") }
func Title() *Element { return NewElement("title") }
func Thead() *Element { return NewElement("thead") }
func Ul() *Element    { return NewElement("ul") }
