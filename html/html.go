package html

import "html"

var (
	EscapeString   = html.EscapeString
	UnescapeString = html.UnescapeString
)

type HTML string

type HTMLer interface {
	HTML() HTML
}

func Background() *Element { return NewElement("") }

func NewHTML() *Element { return NewElement("html") }

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
