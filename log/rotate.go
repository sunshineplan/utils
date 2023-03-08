package log

import (
	"os"
	"os/signal"
)

type Rotatable interface {
	Rotate()
}

func ListenRotateSignal(r Rotatable, sig ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, sig...)
	go func() {
		for range c {
			r.Rotate()
		}
	}()
}
