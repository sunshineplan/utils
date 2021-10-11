package watcher

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type watcher struct {
	sync.Mutex
	file     string
	size     int64
	modTime  time.Time
	ticker   *time.Ticker
	interval time.Duration
	c        chan<- os.FileInfo
}

func (w *watcher) start() {
	w.ticker = time.NewTicker(w.interval)
	for {
		<-w.ticker.C

		info, err := os.Stat(w.file)
		if err != nil {
			log.Printf("Failed to watch %s: %v", w.file, err)
			w.ticker.Stop()
			break
		}
		if w.size != info.Size() || w.modTime != info.ModTime() {
			w.Lock()
			w.size = info.Size()
			w.modTime = info.ModTime()
			w.Unlock()
			w.c <- info
		}
	}
}

// Watcher watches whether file is changed.
type Watcher struct {
	C <-chan os.FileInfo
	w watcher
}

// New creates a Watcher for file. The refresh period is specified by the interval argument.
func New(file string, interval time.Duration) *Watcher {
	if interval <= 0 {
		panic(fmt.Errorf("non-positive interval"))
	}

	file, err := filepath.Abs(file)
	if err != nil {
		panic(err)
	}
	info, err := os.Stat(file)
	if err != nil {
		panic(err)
	}
	if info.IsDir() {
		panic(fmt.Errorf("%s is a directory", file))
	}

	c := make(chan os.FileInfo, 1)
	w := &Watcher{
		C: c,
		w: watcher{
			file:     file,
			size:     info.Size(),
			modTime:  info.ModTime(),
			interval: interval,
			c:        c,
		},
	}
	go w.w.start()
	return w
}

// File returns which file is being watched.
func (w *Watcher) File() string {
	return w.w.file
}

// Stop stops watching file.
func (w *Watcher) Stop() {
	w.w.ticker.Stop()
}
