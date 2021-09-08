package txt

import (
	"reflect"
	"strings"
	"testing"
)

func TestReader(t *testing.T) {
	txt := `A
B
C
`
	content, err := ReadAll(strings.NewReader(txt))
	if err != nil {
		t.Fatal(err)
	}

	if expect := []string{"A", "B", "C"}; !reflect.DeepEqual(expect, content) {
		t.Errorf("expected %v; got %v", expect, content)
	}
}
