package html

import "testing"

func TestTable(t *testing.T) {
	table := `<table><thead><tr><th colspan="2">H1</th><th>H2</th></tr></thead><tbody><tr><td>B1</td><td>B2</td></tr></tbody></table>`
	if e := Table().AppendChild(
		Thead().AppendChild(Tr(Th("H1").Colspan(2), Th("H2"))),
		Tbody().AppendChild(Tr(Td("B1"), Td("B2"))),
	); string(e.HTML()) != table {
		t.Errorf("expected %q; got %q", table, e.HTML())
	}
}
