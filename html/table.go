package html

import "strconv"

var _ HTMLer = new(TableCell)

// TableCell wraps an Element and provides methods specific to <td> and <th> cells.
type TableCell struct{ *Element }

// Tr creates a new <tr> element containing the given table cells.
func Tr(element ...*TableCell) *Element {
	tr := NewElement("tr")
	for _, i := range element {
		tr.AppendContent(i)
	}
	return tr
}

// Th creates a new <th> (table header cell) element with optional content.
func Th(content any) *TableCell {
	return &TableCell{NewElement("th").Content(content)}
}

// Td creates a new <td> (table data cell) element with optional content.
func Td(content any) *TableCell {
	return &TableCell{NewElement("td").Content(content)}
}

// Abbr sets the "abbr" attribute on the table cell.
func (cell *TableCell) Abbr(abbr string) *TableCell {
	cell.Element.Attribute("abbr", abbr)
	return cell
}

// Colspan sets the "colspan" attribute, specifying how many columns the cell spans.
func (cell *TableCell) Colspan(n uint) *TableCell {
	cell.Element.Attribute("colspan", strconv.FormatUint(uint64(n), 10))
	return cell
}

// Headers sets the "headers" attribute, linking the cell to header IDs.
func (cell *TableCell) Headers(headers string) *TableCell {
	cell.Element.Attribute("headers", headers)
	return cell
}

// Rowspan sets the "rowspan" attribute, specifying how many rows the cell spans.
func (cell *TableCell) Rowspan(n uint) *TableCell {
	cell.Element.Attribute("rowspan", strconv.FormatUint(uint64(n), 10))
	return cell
}

// Scope sets the "scope" attribute, typically used in <th> to define scope of headers.
func (cell *TableCell) Scope(scope string) *TableCell {
	cell.Element.Attribute("scope", scope)
	return cell
}

// Class sets the "class" attribute for the table cell.
func (cell *TableCell) Class(class ...string) *TableCell {
	cell.Element.Class(class...)
	return cell
}

// Style sets the "style" attribute for the table cell.
func (cell *TableCell) Style(style string) *TableCell {
	cell.Element.Style(style)
	return cell
}
