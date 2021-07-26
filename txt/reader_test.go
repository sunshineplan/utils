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

	if !reflect.DeepEqual([]string{"A", "B", "C"}, content) {
		t.Errorf("expected %v; got %v", []string{"A", "B", "C"}, content)
	}
}
