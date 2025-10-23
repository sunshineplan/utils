package log

import (
	"bytes"
	"context"
	"log"
	"log/slog"
	"strings"
	"sync"
	"time"
)

// defaultHandler combines a standard log.Logger with an slog.Handler for flexible logging.
type defaultHandler struct {
	*sync.Mutex   // Mutex for thread-safe buffer access.
	*bytes.Buffer // Buffer for formatting log messages.
	*log.Logger   // Underlying standard logger for output.
	slog.Handler  // Structured logging handler.
}

var _ slog.Handler = new(defaultHandler)

// newDefaultHandler creates a new defaultHandler with the specified logger and log level.
func newDefaultHandler(logger *log.Logger, level *slog.LevelVar) *defaultHandler {
	buf := new(bytes.Buffer)
	return &defaultHandler{new(sync.Mutex), buf, logger, slog.NewTextHandler(buf, &slog.HandlerOptions{Level: level})}
}

// Handle formats and outputs a log record using the slog.Handler and log.Logger.
func (h *defaultHandler) Handle(ctx context.Context, r slog.Record) error {
	h.Lock()
	defer h.Unlock()
	r.Time = time.Time{}
	if err := h.Handler.Handle(ctx, r); err != nil {
		return err
	}
	h.Print(strings.TrimPrefix(h.String(), "level="))
	h.Reset()
	return nil
}

// WithAttrs returns a new handler with the specified attributes.
func (h *defaultHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &defaultHandler{h.Mutex, h.Buffer, h.Logger, h.Handler.WithAttrs(attrs)}
}

// WithGroup returns a new handler with the specified group name.
func (h *defaultHandler) WithGroup(name string) slog.Handler {
	return &defaultHandler{h.Mutex, h.Buffer, h.Logger, h.Handler.WithGroup(name)}
}
