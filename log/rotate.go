package log

import (
	"context"
	"os"
	"os/signal"
)

type Rotatable interface {
	Rotate()
}

func ListenRotateSignal(ctx context.Context, r Rotatable, sig ...os.Signal) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, sig...)
	go func() {
		for {
			select {
			case <-ctx.Done():
				signal.Stop(c)
				return
			case <-c:
				r.Rotate()
			}
		}
	}()
}
