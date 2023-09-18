package log

import (
	"context"
	"io"
	"log"
	"log/slog"
)

var std = newLogger(log.Default(), nil)

func Default() *Logger { return std }

func SetOutput(file string, extra io.Writer) {
	std.SetOutput(file, extra)
}

func SetFile(file string) {
	std.SetFile(file)
}

func SetExtra(extra io.Writer) {
	std.SetExtra(extra)
}

func Rotate() {
	std.Rotate()
}

func Flags() int {
	return std.Flags()
}
func SetFlags(flag int) {
	std.SetFlags(flag)
}
func Prefix() string {
	return std.Prefix()
}
func SetPrefix(prefix string) {
	std.SetPrefix(prefix)
}
func Writer() io.Writer {
	return std.Writer()
}
func Print(v ...any) {
	std.Print(v...)
}
func Printf(format string, v ...any) {
	std.Printf(format, v...)
}
func Println(v ...any) {
	std.Println(v...)
}
func Fatal(v ...any) {
	std.Fatal(v...)
}
func Fatalf(format string, v ...any) {
	std.Fatalf(format, v...)
}
func Fatalln(v ...any) {
	std.Fatalln(v...)
}
func Panic(v ...any) {
	std.Panic(v...)
}
func Panicf(format string, v ...any) {
	std.Panicf(format, v...)
}
func Panicln(v ...any) {
	std.Panicln(v...)
}
func Output(calldepth int, s string) error {
	return std.Output(calldepth+1, s) // +1 for this frame.
}

func SetHandler(h slog.Handler) {
	std.mu.Lock()
	defer std.mu.Unlock()
	std.slog = slog.New(h)
}
func Level() *slog.LevelVar {
	return std.level
}
func SetLevel(level slog.Level) {
	std.level.Set(level)
}
func Debug(msg string, args ...any) {
	std.Debug(msg, args...)
}
func DebugContext(ctx context.Context, msg string, args ...any) {
	std.DebugContext(ctx, msg, args...)
}
func Enabled(ctx context.Context, level slog.Level) bool {
	return std.Enabled(ctx, level)
}
func Error(msg string, args ...any) {
	std.Error(msg, args...)
}
func ErrorContext(ctx context.Context, msg string, args ...any) {
	std.ErrorContext(ctx, msg, args...)
}
func Handler() slog.Handler {
	return std.Handler()
}
func Info(msg string, args ...any) {
	std.Info(msg, args...)
}
func InfoContext(ctx context.Context, msg string, args ...any) {
	std.InfoContext(ctx, msg, args...)
}
func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	std.Log(ctx, level, msg, args...)
}
func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	std.LogAttrs(ctx, level, msg, attrs...)
}
func Warn(msg string, args ...any) {
	std.Warn(msg, args...)
}
func WarnContext(ctx context.Context, msg string, args ...any) {
	std.WarnContext(ctx, msg, args...)
}
func With(args ...any) *Logger {
	std = std.With(args...)
	return std
}
func WithGroup(name string) *Logger {
	std = std.WithGroup(name)
	return std
}
