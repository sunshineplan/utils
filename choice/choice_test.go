package choice

import (
	"errors"
	"strings"
	"testing"
)

var _ Description = a("")

type a string

func (s a) Description() string {
	return strings.Repeat(string(s), 2)
}

func (s a) String() string {
	return string(s)
}

func TestMenu(t *testing.T) {
	choices := []a{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j"}
	expect := ` 1. aa
 2. bb
 3. cc
 4. dd
 5. ee
 6. ff
 7. gg
 8. hh
 9. ii
10. jj
 0. Quit
`
	if s := Menu(choices, true); s != expect {
		t.Errorf("expected %q; got %q", expect, s)
	}
}

func TestChoose(t *testing.T) {
	choices := []a{"a", "b", "c"}
	if _, err := choose("4", choices); !errors.Is(err, ErrBadChoice) {
		t.Errorf("expected ErrBadChoice; got %s", err)
	}
	if s, err := choose("2", choices); err != nil {
		t.Fatal(err)
	} else if expect := "b"; s.String() != expect {
		t.Errorf("expected %q; got %q", expect, s)
	}
}

func TestError(t *testing.T) {
	if !errors.Is(choiceError(""), ErrBadChoice) {
		t.Error("expected err is ErrBadChoice; got not")
	}
}
