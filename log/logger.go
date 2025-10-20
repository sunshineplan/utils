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
	*log.Logger
	file  atomic.Pointer[os.File]
	extra container.Value[io.Writer]
	slog  atomic.Pointer[slog.Logger]
	level *slog.LevelVar
}

func newLogger(l *log.Logger, file *os.File) *Logger {
	logger := &Logger{Logger: l, level: new(slog.LevelVar)}
	logger.file.Store(file)
	logger.slog.Store(slog.New(newDefaultHandler(l, logger.level)))
	return logger
}

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

func (l *Logger) File() string {
	if file := l.file.Load(); file != nil {
		return file.Name()
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

func (l *Logger) SetOutput(file string, extra io.Writer) error {
	f, err := openFile(file)
	if err != nil {
		return err
	}
	l.setOutput(f, extra)
	return nil
}

func (l *Logger) SetFile(file string) {
	l.SetOutput(file, l.extra.Load())
}

func (l *Logger) SetExtra(extra io.Writer) {
	l.setOutput(l.file.Load(), extra)
}

func (l *Logger) SetHandler(h slog.Handler) {
	l.slog.Store(slog.New(h))
}
func (l *Logger) Level() *slog.LevelVar {
	return l.level
}
func (l *Logger) SetLevel(level slog.Level) {
	l.level.Set(level)
}
func (l *Logger) Debug(msg string, args ...any) {
	l.slog.Load().Debug(msg, args...)
}
func (l *Logger) DebugContext(ctx context.Context, msg string, args ...any) {
	l.slog.Load().DebugContext(ctx, msg, args...)
}
func (l *Logger) Enabled(ctx context.Context, level slog.Level) bool {
	return l.slog.Load().Enabled(ctx, level)
}
func (l *Logger) Error(msg string, args ...any) {
	l.slog.Load().Error(msg, args...)
}
func (l *Logger) ErrorContext(ctx context.Context, msg string, args ...any) {
	l.slog.Load().ErrorContext(ctx, msg, args...)
}
func (l *Logger) Handler() slog.Handler {
	return l.slog.Load().Handler()
}
func (l *Logger) Info(msg string, args ...any) {
	l.slog.Load().Info(msg, args...)
}
func (l *Logger) InfoContext(ctx context.Context, msg string, args ...any) {
	l.slog.Load().InfoContext(ctx, msg, args...)
}
func (l *Logger) Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	l.slog.Load().Log(ctx, level, msg, args...)
}
func (l *Logger) LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	l.slog.Load().LogAttrs(ctx, level, msg, attrs...)
}
func (l *Logger) Warn(msg string, args ...any) {
	l.slog.Load().Warn(msg, args...)
}
func (l *Logger) WarnContext(ctx context.Context, msg string, args ...any) {
	l.slog.Load().WarnContext(ctx, msg, args...)
}
func (l *Logger) With(args ...any) *Logger {
	logger := &Logger{Logger: l.Logger, extra: l.extra, level: l.level}
	logger.file.Store(l.file.Load())
	if extra := l.extra.Load(); extra != nil {
		logger.extra.Store(extra)
	}
	logger.slog.Store(l.slog.Load().With(args...))
	return logger
}
func (l *Logger) WithGroup(name string) *Logger {
	logger := &Logger{Logger: l.Logger, extra: l.extra, level: l.level}
	logger.file.Store(l.file.Load())
	if extra := l.extra.Load(); extra != nil {
		logger.extra.Store(extra)
	}
	logger.slog.Store(l.slog.Load().WithGroup(name))
	return logger
}

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

func (l *Logger) Write(b []byte) (int, error) {
	return l.Writer().Write(b)
}

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
