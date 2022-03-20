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
func Execute[T any](argMethod, fnMethod Method, limit int, args []T, fn ...func(T) (interface{}, error)) (interface{}, error) {
	if len(fn) == 0 {
		return nil, errors.New("no function provided")
	}

	var clone []T
	var count int
	var nilArg bool
	switch len(args) {
	case 0:
		nilArg = true
	default:
		count = len(args)
		clone = make([]T, count)
		copy(clone, args)
	}

	switch fnMethod {
	case Concurrent:
		result := make(chan interface{}, 1)
		lasterr := make(chan error, 1)

		if nilArg {
			ctx := argContext(len(fn), *new(T))
			defer ctx.cancel()

			workers.RunSlice(limit, fn, func(_ int, fn func(T) (interface{}, error)) {
				ctx.runFn(fn, result, lasterr)
			})

			if err := <-lasterr; err != nil {
				return nil, err
			}

			return <-result, nil
		}

		if argMethod == Random {
			rand.Shuffle(count, func(i, j int) { clone[i], clone[j] = clone[j], clone[i] })
		}

		for i := 0; i < count; i++ {
			ctx := argContext(len(fn), clone[i])
			workers.RunSlice(limit, fn, func(_ int, fn func(T) (interface{}, error)) {
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
		if fnMethod == Random {
			rand.Shuffle(len(fn), func(i, j int) { fn[i], fn[j] = fn[j], fn[i] })
		}

		for i, f := range fn {
			result := make(chan interface{}, 1)
			lasterr := make(chan error, 1)

			ctx := fnContext(count, f)
			if nilArg {
				ctx.runArg(*new(T), result, lasterr)
			} else {
				var worker int
				switch argMethod {
				case Concurrent:
					worker = limit
				case Serial, Random:
					worker = 1
				default:
					return nil, errors.New("unknown arg method")
				}

				if argMethod == Random {
					rand.Shuffle(count, func(i, j int) { clone[i], clone[j] = clone[j], clone[i] })
				}

				workers.RunSlice(worker, clone, func(_ int, arg T) {
					ctx.runArg(arg, result, lasterr)
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
func ExecuteConcurrentArg[T any](arg []T, fn ...func(T) (interface{}, error)) (interface{}, error) {
	return Execute(Concurrent, Serial, defaultLimit, arg, fn...)
}

// ExecuteConcurrentFn gets the fastest result from the functions with args, functions will be run concurrently.
func ExecuteConcurrentFn[T any](arg []T, fn ...func(T) (interface{}, error)) (interface{}, error) {
	return Execute(Serial, Concurrent, defaultLimit, arg, fn...)
}

// ExecuteSerial gets the result until success from the functions with args in order.
func ExecuteSerial[T any](arg []T, fn ...func(T) (interface{}, error)) (interface{}, error) {
	return Execute(Serial, Serial, defaultLimit, arg, fn...)
}

// ExecuteRandom gets the result until success from the functions with args randomly.
func ExecuteRandom[T any](arg []T, fn ...func(T) (interface{}, error)) (interface{}, error) {
	return Execute(Random, Random, defaultLimit, arg, fn...)
}
