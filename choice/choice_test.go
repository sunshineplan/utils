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

func TestChooseWithDefault(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		choices    []string
		def        int
		wantChoice bool
		wantRes    string
		wantErr    bool
	}{
		{
			name:       "valid choice",
			input:      "2\n",
			choices:    []string{"a", "b", "c"},
			def:        0,
			wantChoice: true,
			wantRes:    "b",
			wantErr:    false,
		},
		{
			name:       "default used when empty input",
			input:      "\n",
			choices:    []string{"a", "b", "c"},
			def:        2,
			wantChoice: true,
			wantRes:    "b",
			wantErr:    false,
		},
		{
			name:       "quit with 0",
			input:      "0\n",
			choices:    []string{"a", "b"},
			def:        0,
			wantChoice: false,
			wantRes:    "",
			wantErr:    false,
		},
		{
			name:       "quit with q",
			input:      "q\n",
			choices:    []string{"a", "b"},
			def:        0,
			wantChoice: false,
			wantRes:    "",
			wantErr:    false,
		},
		{
			name:       "invalid input",
			input:      "x\n",
			choices:    []string{"a", "b"},
			def:        0,
			wantChoice: true,
			wantRes:    "",
			wantErr:    true,
		},
		{
			name:       "out of range",
			input:      "5\n",
			choices:    []string{"a", "b"},
			def:        0,
			wantChoice: true,
			wantRes:    "",
			wantErr:    true,
		},
		{
			name:       "no choices",
			input:      "1\n",
			choices:    []string{},
			def:        0,
			wantChoice: false,
			wantRes:    "",
			wantErr:    true,
		},
		{
			name:       "invalid default",
			input:      "1\n",
			choices:    []string{"a"},
			def:        5,
			wantChoice: false,
			wantRes:    "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := strings.NewReader(tt.input)
			choice, res, err := chooseWithDefault(r, tt.choices, tt.def)
			if choice != tt.wantChoice {
				t.Errorf("got choice=%v, want %v", choice, tt.wantChoice)
			}
			if res != tt.wantRes {
				t.Errorf("got res=%q, want %q", res, tt.wantRes)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("got err=%v, wantErr=%v", err, tt.wantErr)
			}
		})
	}
}
