package html

import (
	"fmt"
	"slices"
	"strings"
)

var _ HTMLer = new(Element)

type Element struct {
	tag     string
	attrs   map[string]string
	content HTML
}

func (e *Element) Attribute(name, value string) *Element {
	e.attrs[name] = value
	return e
}

func (e *Element) Class(class ...string) *Element {
	return e.Attribute("class", strings.Join(class, " "))
}

func (e *Element) Href(href string) *Element {
	return e.Attribute("href", href)
}

func (e *Element) Name(name string) *Element {
	return e.Attribute("name", name)
}

func (e *Element) Src(src string) *Element {
	return e.Attribute("src", src)
}

func (e *Element) Style(style string) *Element {
	return e.Attribute("style", style)
}

func (e *Element) Title(title string) *Element {
	return e.Attribute("title", title)
}

func content(v any) HTML {
	switch v := v.(type) {
	case nil:
		return ""
	case HTML:
		return v
	case HTMLer:
		return v.HTML()
	case string:
		return HTML(EscapeString(v))
	default:
		return HTML(EscapeString(fmt.Sprint(v)))
	}
}

func (e *Element) Content(v ...any) *Element {
	e.content = ""
	return e.AppendContent(v...)
}

func (e *Element) Contentf(format string, a ...any) *Element {
	return e.Content(fmt.Sprintf(format, a...))
}

func (e *Element) HTMLContent(html string) *Element {
	return e.Content(HTML(html))
}

func (e *Element) AppendContent(v ...any) *Element {
	for _, v := range v {
		e.content += content(v)
	}
	return e
}

func (e *Element) AppendChild(child ...*Element) *Element {
	for _, i := range child {
		e.AppendContent(i)
	}
	return e
}

func (e *Element) AppendHTML(html ...string) *Element {
	for _, i := range html {
		e.AppendContent(HTML(i))
	}
	return e
}

// https://developer.mozilla.org/en-US/docs/Glossary/Void_element
func (e Element) isVoidElement() bool {
	return slices.Contains([]string{
		"area",
		"base",
		"br",
		"col",
		"embed",
		"hr",
		"img",
		"input",
		"link",
		"meta",
		"param",
		"source",
		"track",
		"wbr",
	}, strings.ToLower(e.tag))
}

// https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes
func (e Element) printAttrs() string {
	var s []string
	for k, v := range e.attrs {
		if v == "" || v == "true" {
			s = append(s, k)
		} else if v == "false" {
			continue
		} else {
			s = append(s, fmt.Sprintf("%s=%q", k, v))
		}
	}
	slices.Sort(s)
	return strings.Join(s, " ")
}

func (e *Element) String() string {
	var b strings.Builder
	if e.tag != "" {
		fmt.Fprint(&b, "<", e.tag)
		if attrs := e.printAttrs(); attrs != "" {
			fmt.Fprint(&b, " ", attrs)
		}
	}
	if e.tag == "" {
		fmt.Fprint(&b, e.content)
	} else if e.isVoidElement() {
		fmt.Fprint(&b, ">")
	} else {
		fmt.Fprint(&b, ">", e.content)
		fmt.Fprintf(&b, "</%s>", e.tag)
	}
	return b.String()
}

func (e *Element) HTML() HTML {
	return HTML(e.String())
}

func NewElement(tag string) *Element {
	return &Element{tag, make(map[string]string), ""}
}
