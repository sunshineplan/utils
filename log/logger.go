package log

import (
	"io"
	"log"
	"os"
	"sync"
)

var (
	_ io.Writer = (*Logger)(nil)
	_ Rotatable = (*Logger)(nil)
)

type Logger struct {
	*log.Logger
	m     sync.Mutex
	file  *os.File
	extra io.Writer
}

func New(file, prefix string, flag int) *Logger {
	if file == "" {
		return &Logger{Logger: log.New(io.Discard, prefix, flag)}
	}
	f := openFile(file)
	return &Logger{Logger: log.New(f, prefix, flag), file: f}
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
}

func (l *Logger) SetOutput(file string, extra io.Writer) {
	l.m.Lock()
	defer l.m.Unlock()
	f := openFile(file)
	l.setOutput(f, extra)
	l.file = f
	l.extra = extra
}

func (l *Logger) SetFile(file string) {
	l.m.Lock()
	defer l.m.Unlock()
	f := openFile(file)
	l.setOutput(f, l.extra)
	l.file = f
}

func (l *Logger) SetExtra(extra io.Writer) {
	l.m.Lock()
	defer l.m.Unlock()
	l.setOutput(l.file, extra)
	l.extra = extra
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
