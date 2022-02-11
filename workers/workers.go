package workers

import (
	"fmt"
	"reflect"
	"runtime"
)

// NumCPU returns the number of logical CPUs usable by the current process.
var NumCPU = runtime.NumCPU

var defaultWorkers = &Workers{limit: NumCPU()}

// Workers can run jobs concurrently with limit.
type Workers struct {
	limit int
}

// New creates a new Workers with the limit.
func New(n int) *Workers {
	return &Workers{limit: n}
}

// SetLimit sets default workers limit.
func SetLimit(n int) {
	defaultWorkers.limit = n
}

// Slice runs the slice on default workers.
func Slice(i interface{}, f func(int, interface{})) error {
	return defaultWorkers.Slice(i, f)
}

// Map runs the map on default workers.
func Map(i interface{}, f func(interface{}, interface{})) error {
	return defaultWorkers.Map(i, f)
}

// Range runs the range on default workers.
func Range(start, end int, f func(int)) error {
	return defaultWorkers.Range(start, end, f)
}

// Slice runs the slice on workers.
func (w *Workers) Slice(i interface{}, f func(int, interface{})) error {
	return runSlice(w.limit, i, f)
}

// Map runs the map on workers.
func (w *Workers) Map(i interface{}, f func(interface{}, interface{})) error {
	return runMap(w.limit, i, f)
}

// Range runs the range on workers.
func (w *Workers) Range(start, end int, f func(int)) error {
	return runRange(w.limit, start, end, f)
}

func runSlice(limit int, i interface{}, f func(int, interface{})) error {
	if reflect.TypeOf(i).Kind() != reflect.Slice {
		return fmt.Errorf("items must be a slice")
	}

	values := reflect.ValueOf(i)

	if limit == 0 {
		limit = NumCPU()
	} else if limit < 0 {
		limit = values.Len()
	}

	c := make(chan struct{}, limit)
	for i := 0; i < values.Len(); i++ {
		c <- struct{}{}
		go func(index int, value interface{}) {
			defer func() { <-c }()
			f(index, value)
		}(i, values.Index(i).Interface())
	}

	for i := 0; i < limit; i++ {
		c <- struct{}{}
	}

	return nil
}

func runMap(limit int, i interface{}, f func(interface{}, interface{})) error {
	if reflect.TypeOf(i).Kind() != reflect.Map {
		return fmt.Errorf("item must be a map")
	}

	value := reflect.ValueOf(i)

	if limit == 0 {
		limit = NumCPU()
	} else if limit < 0 {
		limit = len(value.MapKeys())
	}

	iter := value.MapRange()
	c := make(chan struct{}, limit)
	for iter.Next() {
		k := iter.Key()
		v := iter.Value()
		c <- struct{}{}
		go func(k, v interface{}) {
			defer func() { <-c }()
			f(k, v)
		}(k.Interface(), v.Interface())
	}

	for i := 0; i < limit; i++ {
		c <- struct{}{}
	}

	return nil
}

func runRange(limit, start, end int, f func(int)) error {
	if start > end {
		return nil
	}

	if limit == 0 {
		limit = NumCPU()
	} else if limit < 0 {
		limit = end - start + 1
	}

	c := make(chan struct{}, limit)
	for i := start; i <= end; i++ {
		c <- struct{}{}
		go func(num int) {
			defer func() { <-c }()
			f(num)
		}(i)
	}

	for i := 0; i < limit; i++ {
		c <- struct{}{}
	}

	return nil
}

//func runSlice[Slice ~[]E, E any](limit int, s Slice, f func(int, E)) {
//	if limit == 0 {
//		limit = NumCPU()
//	} else if limit < 0 {
//		limit = len(s)
//	}
//
//	c := make(chan struct{}, limit)
//	for i := 0; i < len(s); i++ {
//		c <- struct{}{}
//		go func(index int, value E) {
//			defer func() { <-c }()
//			f(index, value)
//		}(i, s[i])
//	}
//
//	for i := 0; i < limit; i++ {
//		c <- struct{}{}
//	}
//}
//
//func runMap[Map ~map[K]V, K comparable, V any](limit int, m Map, f func(K, V)) {
//	if limit == 0 {
//		limit = NumCPU()
//	} else if limit < 0 {
//		limit = len(m)
//	}
//
//	c := make(chan struct{}, limit)
//	for k, v := range m {
//		c <- struct{}{}
//		go func(k K, v V) {
//			defer func() { <-c }()
//			f(k, v)
//		}(k, v)
//	}
//
//	for i := 0; i < limit; i++ {
//		c <- struct{}{}
//	}
//}
