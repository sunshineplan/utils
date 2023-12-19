package html

import "strconv"

var _ HTMLer = new(TableCell)

type TableCell struct{ *Element }

func Tr(element ...*TableCell) *Element {
	tr := NewElement("tr")
	for _, i := range element {
		tr.AppendContent(i)
	}
	return tr
}

func Th(content any) *TableCell {
	return &TableCell{NewElement("th").Content(content)}
}

func Td(content any) *TableCell {
	return &TableCell{NewElement("td").Content(content)}
}

func (cell *TableCell) Abbr(abbr string) *TableCell {
	cell.Element.Attribute("abbr", abbr)
	return cell
}

func (cell *TableCell) Colspan(n uint) *TableCell {
	cell.Element.Attribute("colspan", strconv.FormatUint(uint64(n), 10))
	return cell
}

func (cell *TableCell) Headers(headers string) *TableCell {
	cell.Element.Attribute("headers", headers)
	return cell
}

func (cell *TableCell) Rowspan(n uint) *TableCell {
	cell.Element.Attribute("rowspan", strconv.FormatUint(uint64(n), 10))
	return cell
}

func (cell *TableCell) Scope(scope string) *TableCell {
	cell.Element.Attribute("scope", scope)
	return cell
}

func (cell *TableCell) Class(class ...string) *TableCell {
	cell.Element.Class(class...)
	return cell
}

func (cell *TableCell) Style(style string) *TableCell {
	cell.Element.Style(style)
	return cell
}
