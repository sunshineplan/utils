package choice

import (
	"errors"
	"testing"
)

func TestError(t *testing.T) {
	if !errors.Is(choiceError(""), ErrBadChoice) {
		t.Error("expected err is ErrBadChoice; got not")
	}
}
