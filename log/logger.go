package log

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"sync/atomic"

	"github.com/sunshineplan/utils/container"
)

// Logger implements a custom logger that combines the standard log.Logger with slog.Logger,
// providing flexible output destinations and log level control.
type Logger struct {
	*log.Logger                             // Underlying standard logger.
	file        atomic.Pointer[os.File]     // File handle for log output, managed atomically.
	extra       container.Value[io.Writer]  // Additional output destination (e.g., stderr).
	slog        atomic.Pointer[slog.Logger] // Structured logger for leveled logging.
	level       *slog.LevelVar              // Log level controller.
}

var (
	_ io.Writer = new(Logger)
	_ Rotatable = new(Logger)
)

// Constants for log flags, mirrored from the standard log package.
const (
	Ldate         = log.Ldate         // Include date in log output (e.g., 2009/01/23).
	Ltime         = log.Ltime         // Include time in log output (e.g., 01:23:23).
	Lmicroseconds = log.Lmicroseconds // Include microsecond resolution in time (requires Ltime).
	Llongfile     = log.Llongfile     // Include full file name and line number (e.g., /a/b/c/d.go:23).
	Lshortfile    = log.Lshortfile    // Include final file name element and line number (e.g., d.go:23), overrides Llongfile.
	LUTC          = log.LUTC          // Use UTC for date/time if Ldate or Ltime is set.
	Lmsgprefix    = log.Lmsgprefix    // Move prefix to before the message.
	LstdFlags     = log.LstdFlags     // Default flags: Ldate | Ltime.
)

// newLogger creates a new Logger instance with the specified standard logger and file handle.
// The logger is initialized with a default slog handler and log level.
func newLogger(l *log.Logger, file *os.File) *Logger {
	logger := &Logger{Logger: l, level: new(slog.LevelVar)}
	logger.file.Store(file)
	logger.slog.Store(slog.New(newDefaultHandler(l, logger.level)))
	return logger
}

// New creates a new Logger instance with the specified file path, prefix, and flags.
// If file is empty, logs are discarded. Panics if the file cannot be opened.
func New(file, prefix string, flag int) *Logger {
	if file == "" {
		return newLogger(log.New(io.Discard, prefix, flag), nil)
	}
	f, err := openFile(file)
	if err != nil {
		panic(err)
	}
	return newLogger(log.New(f, prefix, flag), f)
}

// File returns the current log file path, or an empty string if no file is set.
func (l *Logger) File() string {
	if file := l.file.Load(); file != nil {
		return file.Name()
	}
	return ""
}

// setOutput configures the logger's output destinations.
// It sets the file and extra writer, closing the old file if necessary.
// Logs are written to io.Discard if no outputs are provided.
func (l *Logger) setOutput(file *os.File, extra io.Writer) {
	var writers []io.Writer
	if file != nil {
		writers = append(writers, file)
	}
	if extra != nil {
		writers = append(writers, extra)
	}
	if len(writers) == 0 {
		writers = append(writers, io.Discard)
	}
	l.Logger.SetOutput(io.MultiWriter(writers...))
	if oldFile := l.file.Load(); oldFile != nil && oldFile != file {
		if err := oldFile.Close(); err != nil {
			l.Error("failed to close log file", "error", err)
		}
	}
	l.file.Store(file)
	if extra != nil {
		l.extra.Store(extra)
	}
}

// SetOutput sets the log output to the specified file path and extra writer.
// Returns an error if the file cannot be opened.
func (l *Logger) SetOutput(file string, extra io.Writer) error {
	f, err := openFile(file)
	if err != nil {
		return err
	}
	l.setOutput(f, extra)
	return nil
}

// SetFile sets the log file to the specified path, keeping the existing extra writer.
func (l *Logger) SetFile(file string) {
	l.SetOutput(file, l.extra.Load())
}

// SetExtra sets an additional output destination (e.g., stderr), keeping the existing file.
func (l *Logger) SetExtra(extra io.Writer) {
	l.setOutput(l.file.Load(), extra)
}

// SetHandler sets the slog handler for structured logging.
// Note: The new handler may not respect the existing log level (l.level), potentially disabling level control.
// Ensure the provided handler is configured with the desired log level if needed.
func (l *Logger) SetHandler(h slog.Handler) {
	l.slog.Store(slog.New(h))
}

// Level returns the current log level.
func (l *Logger) Level() slog.Level {
	return l.level.Level()
}

// SetLevel sets the log level for structured logging.
func (l *Logger) SetLevel(level slog.Level) {
	l.level.Set(level)
}

// Debug logs a message at Debug level with the given arguments.
func (l *Logger) Debug(msg string, args ...any) {
	l.slog.Load().Debug(msg, args...)
}

// DebugContext logs a message at Debug level with the given context and arguments.
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.slog.Load().DebugContext(ctx, msg, args...)
}

// Enabled checks if the specified log level is enabled for the logger.
func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.slog.Load().Enabled(ctx, level)
}

// Error logs a message at Error level with the given arguments.
func (l *Logger) Error(msg string, args ...any) {
	l.slog.Load().Error(msg, args...)
}

// ErrorContext logs a message at Error level with the given context and arguments.
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.slog.Load().ErrorContext(ctx, msg, args...)
}

// SlogHandler returns the current slog handler.
func (l *Logger) SlogHandler() slog.Handler {
	return l.slog.Load().Handler()
}

// Info logs a message at Info level with the given arguments.
func (l *Logger) Info(msg string, args ...any) {
	l.slog.Load().Info(msg, args...)
}

// InfoContext logs a message at Info level with the given context and arguments.
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.slog.Load().InfoContext(ctx, msg, args...)
}

// Log logs a message at the specified level with the given context and arguments.
func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.slog.Load().Log(ctx, level, msg, args...)
}

// LogAttrs logs a message at the specified level with the given context and attributes.
func (l *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.slog.Load().LogAttrs(ctx, level, msg, attrs...)
}

// Warn logs a message at Warn level with the given arguments.
func (l *Logger) Warn(msg string, args ...any) {
	l.slog.Load().Warn(msg, args...)
}

// WarnContext logs a message at Warn level with the given context and arguments.
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.slog.Load().WarnContext(ctx, msg, args...)
}

// With returns a new Logger with the specified attributes, leaving the original unchanged.
func (l *Logger) With(args ...any) *Logger {
	logger := &Logger{Logger: l.Logger, extra: l.extra, level: l.level}
	logger.file.Store(l.file.Load())
	if extra := l.extra.Load(); extra != nil {
		logger.extra.Store(extra)
	}
	logger.slog.Store(l.slog.Load().With(args...))
	return logger
}

// WithGroup returns a new Logger with the specified group name, leaving the original unchanged.
func (l *Logger) WithGroup(name string) *Logger {
	logger := &Logger{Logger: l.Logger, extra: l.extra, level: l.level}
	logger.file.Store(l.file.Load())
	if extra := l.extra.Load(); extra != nil {
		logger.extra.Store(extra)
	}
	logger.slog.Store(l.slog.Load().WithGroup(name))
	return logger
}

// Rotate reopens the log file and rotates the extra writer if it implements Rotatable.
func (l *Logger) Rotate() {
	if extra := l.extra.Load(); extra != nil {
		if i, ok := extra.(Rotatable); ok {
			i.Rotate()
		}
	}
	if file := l.file.Load(); file != nil {
		l.SetFile(file.Name())
	}
}

// Write writes bytes to the logger's output destination, implementing io.Writer.
func (l *Logger) Write(b []byte) (int, error) {
	return l.Writer().Write(b)
}

// openFile opens a log file with the specified path in append mode.
// Returns nil if the path is empty, or an error if the file cannot be opened.
func openFile(file string) (*os.File, error) {
	if file != "" {
		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			return nil, err
		}
		return f, nil
	}
	return nil, nil
}
