package log

import (
	"bytes"
	"context"
	"log"
	"log/slog"
	"sync"
)

var _ slog.Handler = new(defaultHandler)

type defaultHandler struct {
	*sync.Mutex
	*bytes.Buffer
	*log.Logger
	slog.Handler
}

func newDefaultHandler(logger *log.Logger, level *slog.LevelVar) *defaultHandler {
	buf := new(bytes.Buffer)
	return &defaultHandler{new(sync.Mutex), buf, logger, slog.NewTextHandler(buf, &slog.HandlerOptions{Level: level})}
}

func (h *defaultHandler) Handle(ctx context.Context, r slog.Record) error {
	h.Lock()
	defer h.Unlock()
	if err := h.Handler.Handle(ctx, r); err != nil {
		return err
	}
	if _, err := h.Writer().Write(h.Bytes()); err != nil {
		return err
	}
	h.Reset()
	return nil
}

func (h *defaultHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &defaultHandler{h.Mutex, h.Buffer, h.Logger, h.Handler.WithAttrs(attrs)}
}

func (h *defaultHandler) WithGroup(name string) slog.Handler {
	return &defaultHandler{h.Mutex, h.Buffer, h.Logger, h.Handler.WithGroup(name)}
}
