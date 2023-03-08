package log

import (
	"fmt"
	"io"
	"log"
	"os"
)

var std = &Logger{Logger: log.Default()}

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
	std.Output(2, fmt.Sprintf(format, v...))
	os.Exit(1)
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
