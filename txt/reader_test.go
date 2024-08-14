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
	res := ReadAll(strings.NewReader(txt))
	if expect := []string{"A", "B", "C"}; !slices.Equal(expect, res) {
		t.Errorf("expected %v; got %v", expect, res)
	}
}
