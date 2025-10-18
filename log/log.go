package log

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"sync/atomic"
)

var defaultLogger atomic.Pointer[Logger]

func init() {
	defaultLogger.Store(newLogger(log.Default(), os.Stderr))
}

func Default() *Logger { return defaultLogger.Load() }

func SetDefault(l *Logger) {
	defaultLogger.Store(l)
}

func File() string {
	return Default().File()
}

func SetOutput(file string, extra io.Writer) {
	Default().SetOutput(file, extra)
}

func SetFile(file string) {
	Default().SetFile(file)
}

func SetExtra(extra io.Writer) {
	Default().SetExtra(extra)
}

func Rotate() {
	Default().Rotate()
}

func Flags() int {
	return Default().Flags()
}
func SetFlags(flag int) {
	Default().SetFlags(flag)
}
func Prefix() string {
	return Default().Prefix()
}
func SetPrefix(prefix string) {
	Default().SetPrefix(prefix)
}
func Writer() io.Writer {
	return Default().Writer()
}
func Print(v ...any) {
	Default().Print(v...)
}
func Printf(format string, v ...any) {
	Default().Printf(format, v...)
}
func Println(v ...any) {
	Default().Println(v...)
}
func Fatal(v ...any) {
	Default().Fatal(v...)
}
func Fatalf(format string, v ...any) {
	Default().Fatalf(format, v...)
}
func Fatalln(v ...any) {
	Default().Fatalln(v...)
}
func Panic(v ...any) {
	Default().Panic(v...)
}
func Panicf(format string, v ...any) {
	Default().Panicf(format, v...)
}
func Panicln(v ...any) {
	Default().Panicln(v...)
}
func Output(calldepth int, s string) error {
	return Default().Output(calldepth+1, s) // +1 for this frame.
}

func SetHandler(h slog.Handler) {
	Default().slog.Store(slog.New(h))
}
func Level() *slog.LevelVar {
	return Default().level
}
func SetLevel(level slog.Level) {
	Default().level.Set(level)
}
func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}
func DebugContext(ctx context.Context, msg string, args ...any) {
	Default().DebugContext(ctx, msg, args...)
}
func Enabled(ctx context.Context, level slog.Level) bool {
	return Default().Enabled(ctx, level)
}
func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}
func ErrorContext(ctx context.Context, msg string, args ...any) {
	Default().ErrorContext(ctx, msg, args...)
}
func Handler() slog.Handler {
	return Default().Handler()
}
func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}
func InfoContext(ctx context.Context, msg string, args ...any) {
	Default().InfoContext(ctx, msg, args...)
}
func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	Default().Log(ctx, level, msg, args...)
}
func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	Default().LogAttrs(ctx, level, msg, attrs...)
}
func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}
func WarnContext(ctx context.Context, msg string, args ...any) {
	Default().WarnContext(ctx, msg, args...)
}
func With(args ...any) *Logger {
	return Default().With(args...)
}
func WithGroup(name string) *Logger {
	return Default().WithGroup(name)
}
