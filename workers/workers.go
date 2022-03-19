package workers

import (
	"runtime"
)

// NumCPU returns the number of logical CPUs usable by the current process.
var NumCPU = runtime.NumCPU

// Integer is a constraint that permits any integer type used by Range.
type Integer interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
}

var defaultWorkers = NumCPU()

// SetLimit sets default workers limit.
func SetLimit(n int) {
	defaultWorkers = n
}

// Slice runs the slice with default workers limit.
func Slice[E any](s []E, f func(int, E)) {
	RunSlice(defaultWorkers, s, f)
}

// Map runs the map with default workers limit.
func Map[K comparable, V any](m map[K]V, f func(K, V)) {
	RunMap(defaultWorkers, m, f)
}

// Range runs the range with default workers limit.
func Range[Int Integer](start, end Int, f func(Int)) {
	RunRange(defaultWorkers, start, end, f)
}

// RunSlice runs the slice with limit.
func RunSlice[E any](limit int, s []E, f func(int, E)) {
	if limit == 0 {
		limit = NumCPU()
	} else if limit < 0 {
		limit = len(s)
	}

	c := make(chan struct{}, limit)
	for i := 0; i < len(s); i++ {
		c <- struct{}{}
		go func(index int, value E) {
			defer func() { <-c }()
			f(index, value)
		}(i, s[i])
	}

	for i := 0; i < limit; i++ {
		c <- struct{}{}
	}
}

// RunMap runs the map with limit.
func RunMap[K comparable, V any](limit int, m map[K]V, f func(K, V)) {
	if limit == 0 {
		limit = NumCPU()
	} else if limit < 0 {
		limit = len(m)
	}

	c := make(chan struct{}, limit)
	for k, v := range m {
		c <- struct{}{}
		go func(k K, v V) {
			defer func() { <-c }()
			f(k, v)
		}(k, v)
	}

	for i := 0; i < limit; i++ {
		c <- struct{}{}
	}
}

// RunRange runs the range with limit.
func RunRange[Int Integer](limit int, start, end Int, f func(Int)) {
	if start > end {
		return
	}

	if limit == 0 {
		limit = NumCPU()
	} else if limit < 0 {
		limit = int(end-start) + 1
	}

	c := make(chan struct{}, limit)
	for i := start; i <= end; i++ {
		c <- struct{}{}
		go func(num Int) {
			defer func() { <-c }()
			f(num)
		}(i)
	}

	for i := 0; i < limit; i++ {
		c <- struct{}{}
	}
}
