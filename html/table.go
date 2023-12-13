package html

import "strconv"

func Tr[T *TableHeader | *TableData](element ...T) *Element {
	tr := NewElement("tr")
	for _, i := range element {
		tr.AppendContent(i)
	}
	return tr
}

var (
	_ HTMLer = new(TableHeader)
	_ HTMLer = new(TableData)
)

type (
	TableHeader struct{ *Element }
	TableData   struct{ *Element }
)

func Th(content any) *TableHeader {
	return &TableHeader{NewElement("th").Content(content)}
}

func (th *TableHeader) Abbr(abbr string) *TableHeader {
	th.Element.Attribute("abbr", abbr)
	return th
}

func (th *TableHeader) Colspan(n uint) *TableHeader {
	th.Element.Attribute("colspan", strconv.FormatUint(uint64(n), 10))
	return th
}

func (th *TableHeader) Headers(headers string) *TableHeader {
	th.Element.Attribute("headers", headers)
	return th
}

func (th *TableHeader) Rowspan(n uint) *TableHeader {
	th.Element.Attribute("rowspan", strconv.FormatUint(uint64(n), 10))
	return th
}

func (th *TableHeader) Scope(scope string) *TableHeader {
	th.Element.Attribute("scope", scope)
	return th
}

func Td(content any) *TableData {
	return &TableData{NewElement("td").Content(content)}
}

func (td *TableData) Colspan(n uint) *TableData {
	td.Element.Attribute("colspan", strconv.FormatUint(uint64(n), 10))
	return td
}

func (td *TableData) Headers(headers string) *TableData {
	td.Element.Attribute("headers", headers)
	return td
}

func (td *TableData) Rowspan(n uint) *TableData {
	td.Element.Attribute("rowspan", strconv.FormatUint(uint64(n), 10))
	return td
}
