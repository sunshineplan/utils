package choice

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

var (
	_ Choice      = a("")
	_ Description = a("")
)

type a string

func (s a) Run() (any, error) {
	return s, nil
}

func (s a) Description() string {
	return strings.Repeat(string(s), 2)
}

func (s a) String() string {
	return string(s)
}

func TestMenu(t *testing.T) {
	choices := []a{"a", "b", "c"}
	expect := `1. aa
2. bb
3. cc
`
	if s := Menu(choices); s != expect {
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
	} else if expect := "b"; fmt.Sprint(s) != expect {
		t.Errorf("expected %q; got %q", expect, s)
	}
}
