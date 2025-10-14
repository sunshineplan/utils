package html

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/sunshineplan/utils/pool"
)

var _ HTMLer = new(Element)

type Element struct {
	tag     string
	attrs   map[string]string
	content HTML
}

func (e *Element) Attribute(name, value string) *Element {
	if e.attrs == nil {
		e.attrs = make(map[string]string)
	}
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
		e.content += i.HTML()
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
var voidElements = map[string]struct{}{
	"area":   {},
	"base":   {},
	"br":     {},
	"col":    {},
	"embed":  {},
	"hr":     {},
	"img":    {},
	"input":  {},
	"link":   {},
	"meta":   {},
	"param":  {},
	"source": {},
	"track":  {},
	"wbr":    {},
}

func (e Element) isVoidElement() bool {
	_, ok := voidElements[strings.ToLower(e.tag)]
	return ok
}

var builderPool = pool.New[strings.Builder]()

// https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes
func (e Element) printAttrs() string {
	if len(e.attrs) == 0 {
		return ""
	}
	keys := make([]string, 0, len(e.attrs))
	for k := range e.attrs {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	b := builderPool.Get()
	defer func() {
		b.Reset()
		builderPool.Put(b)
	}()
	first := true
	for _, k := range keys {
		v := e.attrs[k]
		switch v {
		case "", "true":
			if !first {
				b.WriteByte(' ')
			}
			b.WriteString(k)
			first = false
		case "false":
			continue
		default:
			if !first {
				b.WriteByte(' ')
			}
			b.WriteString(k)
			b.WriteByte('=')
			b.WriteByte('"')
			b.WriteString(EscapeString(v))
			b.WriteByte('"')
			first = false
		}
	}
	return b.String()
}

func (e *Element) String() string {
	if e.tag == "" {
		return string(e.content)
	}
	b := builderPool.Get()
	defer func() {
		b.Reset()
		builderPool.Put(b)
	}()
	b.WriteByte('<')
	b.WriteString(e.tag)
	if attrs := e.printAttrs(); attrs != "" {
		b.WriteByte(' ')
		b.WriteString(attrs)
	}
	b.WriteByte('>')
	if !e.isVoidElement() {
		b.WriteString(string(e.content))
		b.WriteString("</")
		b.WriteString(e.tag)
		b.WriteByte('>')
	}
	return b.String()
}

func (e *Element) HTML() HTML {
	return HTML(e.String())
}

func (e *Element) Clone() *Element {
	attrs := make(map[string]string, len(e.attrs))
	maps.Copy(attrs, e.attrs)
	return &Element{
		tag:     e.tag,
		attrs:   attrs,
		content: e.content,
	}
}

func NewElement(tag string) *Element {
	return &Element{tag: tag, attrs: make(map[string]string)}
}
