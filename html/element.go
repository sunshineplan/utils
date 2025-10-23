package html

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/sunshineplan/utils/pool"
)

var _ HTMLer = new(Element)

// Element represents a single HTML element, including its tag name,
// attributes, and inner HTML content.
type Element struct {
	tag     string
	attrs   map[string]string
	content HTML
}

// Attribute sets or updates an attribute on the element.
// If attrs is nil, it initializes the map.
func (e *Element) Attribute(name, value string) *Element {
	if e.attrs == nil {
		e.attrs = make(map[string]string)
	}
	e.attrs[name] = value
	return e
}

// Class sets the "class" attribute. Multiple classes can be provided.
func (e *Element) Class(class ...string) *Element {
	return e.Attribute("class", strings.Join(class, " "))
}

// Href sets the "href" attribute.
func (e *Element) Href(href string) *Element {
	return e.Attribute("href", href)
}

// Name sets the "name" attribute.
func (e *Element) Name(name string) *Element {
	return e.Attribute("name", name)
}

// Src sets the "src" attribute.
func (e *Element) Src(src string) *Element {
	return e.Attribute("src", src)
}

// Style sets the "style" attribute.
func (e *Element) Style(style string) *Element {
	return e.Attribute("style", style)
}

// Title sets the "title" attribute.
func (e *Element) Title(title string) *Element {
	return e.Attribute("title", title)
}

// content converts an arbitrary value into escaped HTML text.
// It handles HTMLer, HTML, string, and other types gracefully.
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

// Content replaces the current content of the element with new values.
func (e *Element) Content(v ...any) *Element {
	e.content = ""
	return e.AppendContent(v...)
}

// Contentf formats a string using fmt.Sprintf and sets it as the element content.
func (e *Element) Contentf(format string, a ...any) *Element {
	return e.Content(fmt.Sprintf(format, a...))
}

// HTMLContent inserts raw (unescaped) HTML into the element content.
func (e *Element) HTMLContent(html string) *Element {
	return e.Content(HTML(html))
}

// AppendContent appends additional content to the element.
func (e *Element) AppendContent(v ...any) *Element {
	for _, v := range v {
		e.content += content(v)
	}
	return e
}

// AppendChild appends child elements to the current element.
func (e *Element) AppendChild(child ...*Element) *Element {
	for _, i := range child {
		e.content += i.HTML()
	}
	return e
}

// AppendHTML appends one or more raw HTML strings directly to the element.
func (e *Element) AppendHTML(html ...string) *Element {
	for _, i := range html {
		e.AppendContent(HTML(i))
	}
	return e
}

// https://developer.mozilla.org/en-US/docs/Glossary/Void_element
// Map of HTML5 void elements that do not require a closing tag.
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

// isVoidElement reports whether the element is a void (self-closing) element.
func (e Element) isVoidElement() bool {
	_, ok := voidElements[strings.ToLower(e.tag)]
	return ok
}

// builderPool provides a pool of reusable strings.Builder instances
// to reduce memory allocations during HTML rendering.
var builderPool = pool.New[strings.Builder]()

// https://developer.mozilla.org/en-US/docs/Web/HTML/Attributes
// printAttrs returns a formatted string representation of all attributes
// sorted by key. Boolean attributes (true or empty) are printed without a value.
// Attributes with value "false" are omitted.
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

// String returns the serialized HTML representation of the element.
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

// HTML returns the element as HTML type, implementing [HTMLer].
func (e *Element) HTML() HTML {
	return HTML(e.String())
}

// Clone creates a deep copy of the element and its attributes.
func (e *Element) Clone() *Element {
	attrs := make(map[string]string, len(e.attrs))
	maps.Copy(attrs, e.attrs)
	return &Element{
		tag:     e.tag,
		attrs:   attrs,
		content: e.content,
	}
}

// NewElement creates and returns a new HTML element with the given tag name.
func NewElement(tag string) *Element {
	return &Element{tag: tag, attrs: make(map[string]string)}
}
