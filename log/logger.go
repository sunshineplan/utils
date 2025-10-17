package log

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"sync"
)

var (
	_ io.Writer = new(Logger)
	_ Rotatable = new(Logger)
)

const (
	Ldate         = log.Ldate         // the date in the local time zone: 2009/01/23
	Ltime         = log.Ltime         // the time in the local time zone: 01:23:23
	Lmicroseconds = log.Lmicroseconds // microsecond resolution: 01:23:23.123123.  assumes Ltime.
	Llongfile     = log.Llongfile     // full file name and line number: /a/b/c/d.go:23
	Lshortfile    = log.Lshortfile    // final file name element and line number: d.go:23. overrides Llongfile
	LUTC          = log.LUTC          // if Ldate or Ltime is set, use UTC rather than the local time zone
	Lmsgprefix    = log.Lmsgprefix    // move the "prefix" from the beginning of the line to before the message
	LstdFlags     = log.LstdFlags     // initial values for the standard logger
)

type Logger struct {
	mu sync.Mutex
	*log.Logger

	file  *os.File
	extra io.Writer

	slog  *slog.Logger
	level *slog.LevelVar
}

func newLogger(l *log.Logger, file *os.File) *Logger {
	logger := &Logger{Logger: l, file: file, level: new(slog.LevelVar)}
	logger.slog = slog.New(newDefaultHandler(&logger.mu, l, logger.level))
	return logger
}

func New(file, prefix string, flag int) *Logger {
	if file == "" {
		return newLogger(log.New(io.Discard, prefix, flag), nil)
	}
	f := openFile(file)
	return newLogger(log.New(f, prefix, flag), f)
}

func (l *Logger) File() string {
	if l.file != nil {
		return l.file.Name()
	}
	return ""
}

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
	if l.file != nil && l.file != file {
		l.file.Close()
	}
	l.file = file
	l.extra = extra
}

func (l *Logger) SetOutput(file string, extra io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.setOutput(openFile(file), extra)
}

func (l *Logger) SetFile(file string) {
	l.SetOutput(file, l.extra)
}

func (l *Logger) SetExtra(extra io.Writer) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.setOutput(l.file, extra)
}

func (l *Logger) SetHandler(h slog.Handler) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.slog = slog.New(h)
}
func (l *Logger) Level() *slog.LevelVar {
	return l.level
}
func (l *Logger) SetLevel(level slog.Level) {
	l.level.Set(level)
}
func (l *Logger) Debug(msg string, args ...any) {
	l.slog.Debug(msg, args...)
}
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.slog.DebugContext(ctx, msg, args...)
}
func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.slog.Enabled(ctx, level)
}
func (l *Logger) Error(msg string, args ...any) {
	l.slog.Error(msg, args...)
}
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.slog.ErrorContext(ctx, msg, args...)
}
func (l *Logger) Handler() slog.Handler {
	return l.slog.Handler()
}
func (l *Logger) Info(msg string, args ...any) {
	l.slog.Info(msg, args...)
}
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.slog.InfoContext(ctx, msg, args...)
}
func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.slog.Log(ctx, level, msg, args...)
}
func (l *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.slog.LogAttrs(ctx, level, msg, attrs...)
}
func (l *Logger) Warn(msg string, args ...any) {
	l.slog.Warn(msg, args...)
}
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.slog.WarnContext(ctx, msg, args...)
}
func (l *Logger) With(args ...any) *Logger {
	return &Logger{Logger: l.Logger, file: l.file, extra: l.extra, slog: l.slog.With(args...), level: l.level}
}
func (l *Logger) WithGroup(name string) *Logger {
	return &Logger{Logger: l.Logger, file: l.file, extra: l.extra, slog: l.slog.WithGroup(name), level: l.level}
}

func (l *Logger) Rotate() {
	if i, ok := l.extra.(Rotatable); ok {
		i.Rotate()
	}
	if l.file != nil {
		l.SetFile(l.file.Name())
	}
}

func (l *Logger) Write(b []byte) (int, error) {
	return l.Writer().Write(b)
}

func openFile(file string) *os.File {
	if file != "" {
		f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0640)
		if err != nil {
			panic(err)
		}
		return f
	}
	return nil
}
