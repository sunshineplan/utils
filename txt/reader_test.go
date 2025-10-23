package txt

import (
	"slices"
	"strings"
	"testing"
)

func TestReader(t *testing.T) {
	txt := `A
B
C
`
	res, err := ReadAll(strings.NewReader(txt))
	if err != nil {
		t.Fatal(err)
	}
	if expect := []string{"A", "B", "C"}; !slices.Equal(expect, res) {
		t.Errorf("expected %v; got %v", expect, res)
	}
}
