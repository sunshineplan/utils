package log

import (
	"context"
	"os"
	"os/signal"
)

// Rotatable defines an interface for objects that support log rotation.
type Rotatable interface {
	Rotate()
}

// ListenRotateSignal listens for the specified signals and triggers rotation on the Rotatable object.
// It stops listening when the context is canceled.
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
