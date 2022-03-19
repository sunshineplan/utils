package executor

import (
	"errors"
	"math/rand"
	"time"

	"github.com/sunshineplan/utils/workers"
)

// Method represents the execute method.
type Method int

const (
	// Concurrent method. Executing at the same time.
	Concurrent Method = iota
	// Serial method. Executing in order.
	Serial
	// Random method. Executing randomly.
	Random
)

var defaultLimit = workers.NumCPU()

func init() {
	rand.Seed(time.Now().UnixNano())
}

// SetLimit sets default limit.
// If n < 0, represents no limit.
// If n = 0, limit will be set as CPU number.
func SetLimit(n int) {
	defaultLimit = n
}

// Execute gets the result from the functions with several args by specified method.
// If both argMethod and fnMethod is Concurrent, fnMethod will be first.
func Execute(argMethod, fnMethod Method, limit int, arg []interface{}, fn ...func(interface{}) (interface{}, error)) (interface{}, error) {
	if len(fn) == 0 {
		return nil, errors.New("no function provided")
	}

	var v []interface{}
	var count int
	var nilArg bool
	switch len(arg) {
	case 0:
		count = 1
		nilArg = true
	default:
		count = len(arg)
		v = make([]interface{}, count)
		copy(v, arg)
		if argMethod == Random {
			rand.Shuffle(len(v), func(i, j int) { v[i], v[j] = v[j], v[i] })
		}
	}

	if fnMethod == Random {
		rand.Shuffle(len(fn), func(i, j int) { fn[i], fn[j] = fn[j], fn[i] })
	}

	switch fnMethod {
	case Concurrent:
		result := make(chan interface{}, 1)
		lasterr := make(chan error, 1)

		if nilArg {
			ctx := newContext(len(fn), 0, nil)
			defer ctx.cancel()

			workers.RunSlice(limit, fn, func(_ int, fn func(interface{}) (interface{}, error)) {
				ctx.runFn(fn, result, lasterr)
			})

			if err := <-lasterr; err != nil {
				return nil, err
			}

			return <-result, nil
		}

		for i := 0; i < count; i++ {
			ctx := newContext(len(fn), argKey, v[i])
			workers.RunSlice(limit, fn, func(_ int, fn func(interface{}) (interface{}, error)) {
				ctx.runFn(fn, result, lasterr)
			})

			if err := <-lasterr; err == nil {
				return <-result, nil
			} else if i == count-1 {
				return nil, err
			}
			ctx.cancel()
		}
	case Serial, Random:
		for i, f := range fn {
			result := make(chan interface{}, 1)
			lasterr := make(chan error, 1)

			ctx := newContext(count, fnKey, f)
			if nilArg {
				ctx.runArg(nil, result, lasterr)
			} else {
				if argMethod == Random {
					rand.Shuffle(count, func(i, j int) { v[i], v[j] = v[j], v[i] })
				}

				var worker int
				switch argMethod {
				case Concurrent:
					worker = limit
				case Serial, Random:
					worker = 1
				default:
					return nil, errors.New("unknown arg method")
				}
				workers.RunSlice(worker, v, func(_ int, i interface{}) {
					ctx.runArg(i, result, lasterr)
				})
			}

			if err := <-lasterr; err == nil {
				return <-result, nil
			} else if i == len(fn)-1 {
				return nil, err
			}
			ctx.cancel()
		}
	default:
		return nil, errors.New("unknown function method")
	}

	return nil, errors.New("unknown error")
}

// ExecuteConcurrentArg gets the fastest result from the functions with args, args will be run concurrently.
func ExecuteConcurrentArg(arg []interface{}, fn ...func(interface{}) (interface{}, error)) (interface{}, error) {
	return Execute(Concurrent, Serial, defaultLimit, arg, fn...)
}

// ExecuteConcurrentFn gets the fastest result from the functions with args, functions will be run concurrently.
func ExecuteConcurrentFn(arg []interface{}, fn ...func(interface{}) (interface{}, error)) (interface{}, error) {
	return Execute(Serial, Concurrent, defaultLimit, arg, fn...)
}

// ExecuteSerial gets the result until success from the functions with args in order.
func ExecuteSerial(arg []interface{}, fn ...func(interface{}) (interface{}, error)) (interface{}, error) {
	return Execute(Serial, Serial, defaultLimit, arg, fn...)
}

// ExecuteRandom gets the result until success from the functions with args randomly.
func ExecuteRandom(arg []interface{}, fn ...func(interface{}) (interface{}, error)) (interface{}, error) {
	return Execute(Random, Random, defaultLimit, arg, fn...)
}
