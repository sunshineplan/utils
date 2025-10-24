package log

import (
	"context"
	"io"
	"log"
	"log/slog"
	"os"
	"sync/atomic"
)

// defaultLogger holds the default Logger instance, managed atomically for thread safety.
var defaultLogger atomic.Pointer[Logger]

// init initializes the default logger with stderr output.
func init() {
	defaultLogger.Store(newLogger(log.Default(), os.Stderr))
}

// Default returns the current default Logger instance.
func Default() *Logger { return defaultLogger.Load() }

// SetDefault sets the default Logger instance.
func SetDefault(l *Logger) {
	defaultLogger.Store(l)
}

// File returns the current log file path of the default Logger.
func File() string {
	return Default().File()
}

// SetOutput sets the output destination for the default Logger.
// The file parameter specifies the log file path; if empty, no file output is used.
// The extra parameter allows an additional output destination (e.g., stderr).
func SetOutput(file string, extra io.Writer) {
	Default().SetOutput(file, extra)
}

// SetFile sets the log file path for the default Logger, keeping the existing extra writer.
func SetFile(file string) {
	Default().SetFile(file)
}

// SetExtra sets an additional output destination for the default Logger, keeping the existing file.
func SetExtra(extra io.Writer) {
	Default().SetExtra(extra)
}

// Rotate reopens the log file and rotates the extra writer for the default Logger if applicable.
func Rotate() {
	Default().Rotate()
}

// Flags returns the current log flags of the default Logger.
func Flags() int {
	return Default().Flags()
}

// SetFlags sets the log flags for the default Logger.
func SetFlags(flag int) {
	Default().SetFlags(flag)
}

// Prefix returns the current log prefix of the default Logger.
func Prefix() string {
	return Default().Prefix()
}

// SetPrefix sets the log prefix for the default Logger.
func SetPrefix(prefix string) {
	Default().SetPrefix(prefix)
}

// Writer returns the current output writer of the default Logger.
func Writer() io.Writer {
	return Default().Writer()
}

// Print logs a message using the default Logger's Print method.
func Print(v ...any) {
	Default().Print(v...)
}

// Printf logs a formatted message using the default Logger's Printf method.
func Printf(format string, v ...any) {
	Default().Printf(format, v...)
}

// Println logs a message with a newline using the default Logger's Println method.
func Println(v ...any) {
	Default().Println(v...)
}

// Fatal logs a message and exits using the default Logger's Fatal method.
func Fatal(v ...any) {
	Default().Fatal(v...)
}

// Fatalf logs a formatted message and exits using the default Logger's Fatalf method.
func Fatalf(format string, v ...any) {
	Default().Fatalf(format, v...)
}

// Fatalln logs a message with a newline and exits using the default Logger's Fatalln method.
func Fatalln(v ...any) {
	Default().Fatalln(v...)
}

// Panic logs a message and panics using the default Logger's Panic method.
func Panic(v ...any) {
	Default().Panic(v...)
}

// Panicf logs a formatted message and panics using the default Logger's Panicf method.
func Panicf(format string, v ...any) {
	Default().Panicf(format, v...)
}

// Panicln logs a message with a newline and panics using the default Logger's Panicln method.
func Panicln(v ...any) {
	Default().Panicln(v...)
}

// Output logs a message with the specified call depth using the default Logger's Output method.
func Output(calldepth int, s string) error {
	return Default().Output(calldepth+1, s) // +1 for this frame.
}

// SetHandler sets the slog handler for the default Logger.
// Note: The new handler may not respect the existing log level (obtained via [Level]), potentially disabling level control.
// Ensure the provided handler is configured with the desired log level if needed.
func SetHandler(h slog.Handler) {
	Default().slog.Store(slog.New(h))
}

// Level returns the log level of the default Logger.
func Level() slog.Level {
	return Default().Level()
}

// SetLevel sets the log level for the default Logger.
func SetLevel(level slog.Level) {
	Default().level.Set(level)
}

// Debug logs a message at Debug level using the default Logger.
func Debug(msg string, args ...any) {
	Default().Debug(msg, args...)
}

// DebugContext logs a message at Debug level with context using the default Logger.
func DebugContext(ctx context.Context, msg string, args ...any) {
	Default().DebugContext(ctx, msg, args...)
}

// Enabled checks if the specified log level is enabled for the default Logger.
func Enabled(ctx context.Context, level slog.Level) bool {
	return Default().Enabled(ctx, level)
}

// Error logs a message at Error level using the default Logger.
func Error(msg string, args ...any) {
	Default().Error(msg, args...)
}

// ErrorContext logs a message at Error level with context using the default Logger.
func ErrorContext(ctx context.Context, msg string, args ...any) {
	Default().ErrorContext(ctx, msg, args...)
}

// SlogHandler returns the slog handler of the default Logger.
func SlogHandler() slog.Handler {
	return Default().SlogHandler()
}

// Info logs a message at Info level using the default Logger.
func Info(msg string, args ...any) {
	Default().Info(msg, args...)
}

// InfoContext logs a message at Info level with context using the default Logger.
func InfoContext(ctx context.Context, msg string, args ...any) {
	Default().InfoContext(ctx, msg, args...)
}

// Log logs a message at the specified level with context using the default Logger.
func Log(ctx context.Context, level slog.Level, msg string, args ...any) {
	Default().Log(ctx, level, msg, args...)
}

// LogAttrs logs a message at the specified level with attributes using the default Logger.
func LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr) {
	Default().LogAttrs(ctx, level, msg, attrs...)
}

// Warn logs a message at Warn level using the default Logger.
func Warn(msg string, args ...any) {
	Default().Warn(msg, args...)
}

// WarnContext logs a message at Warn level with context using the default Logger.
func WarnContext(ctx context.Context, msg string, args ...any) {
	Default().WarnContext(ctx, msg, args...)
}

// With returns a new Logger with the specified attributes, leaving the default Logger unchanged.
func With(args ...any) *Logger {
	return Default().With(args...)
}

// WithGroup returns a new Logger with the specified group name, leaving the default Logger unchanged.
func WithGroup(name string) *Logger {
	return Default().WithGroup(name)
}
