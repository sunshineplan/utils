package log

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"log/slog"
	"strings"
	"sync"
	"time"
)

var _ slog.Handler = new(defaultHandler)

type defaultHandler struct {
	*sync.Mutex
	slog.Handler
	*log.Logger
	*bytes.Buffer
}

func newDefaultHandler(mu *sync.Mutex, logger *log.Logger, opts *slog.HandlerOptions) *defaultHandler {
	buf := new(bytes.Buffer)
	return &defaultHandler{mu, slog.NewTextHandler(buf, opts), logger, buf}
}

func (h *defaultHandler) Handle(ctx context.Context, r slog.Record) error {
	h.Lock()
	defer h.Unlock()
	msg := fmt.Sprintf("%s %s", r.Level, r.Message)
	r.Time, r.Message, r.Level = time.Time{}, "", 0
	h.Handler.Handle(ctx, r)
	if log := strings.TrimSpace(strings.Replace(strings.Replace(h.String(), "level=INFO", "", 1), ` msg=""`, "", 1)); log == "" {
		h.Print(msg)
	} else {
		h.Println(msg, log)
	}
	h.Reset()
	return nil
}

func (h *defaultHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &defaultHandler{h.Mutex, h.Handler.WithAttrs(attrs), h.Logger, h.Buffer}
}

func (h *defaultHandler) WithGroup(name string) slog.Handler {
	return &defaultHandler{h.Mutex, h.Handler.WithGroup(name), h.Logger, h.Buffer}
}
