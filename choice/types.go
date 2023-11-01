package choice

import "errors"

type Choice interface {
	Run() (any, error)
}

type Description interface {
	Description() string
}

var ErrBadChoice = errors.New("bad choice")

var _ error = choiceError("")

type choiceError string

func (err choiceError) Error() string {
	return "bad choice: " + string(err)
}

func (choiceError) Unwrap() error {
	return ErrBadChoice
}
