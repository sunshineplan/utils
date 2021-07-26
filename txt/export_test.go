package txt

import (
	"bytes"
	"testing"
)

func TestExport(t *testing.T) {
	testcase := []string{"A", "B", "C"}
	result := `A
B
C
`

	var b bytes.Buffer
	if err := Export(testcase, &b); err != nil {
		t.Error(err)
	}
	if r := b.String(); r != result {
		t.Errorf("expected %q; got %q", result, r)
	}

}
