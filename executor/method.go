package executor

import (
	"errors"
	"math/rand"
	"reflect"
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
func Execute(argMethod, fnMethod Method, limit int, arg interface{}, fn ...func(interface{}) (interface{}, error)) (interface{}, error) {
	if len(fn) == 0 {
		return nil, errors.New("no function provided")
	}

	src := reflect.ValueOf(arg)

	var v reflect.Value
	var count int
	var nilArg bool
	switch {
	case arg == nil:
		count = 1
		nilArg = true
	case reflect.TypeOf(arg).Kind() == reflect.Slice:
		count = src.Len()
		if count == 0 {
			return nil, errors.New("arg can not be empty slice")
		}
		v = reflect.MakeSlice(src.Type(), count, count)
		reflect.Copy(v, src)

		if argMethod == Random {
			rand.Shuffle(count, func(i, j int) {
				a := v.Index(i).Interface()
				b := v.Index(j).Interface()
				v.Index(j).Set(reflect.ValueOf(a))
				v.Index(i).Set(reflect.ValueOf(b))
			})
		}
	default:
		count = 1
		v = reflect.MakeSlice(reflect.SliceOf(src.Type()), 0, 0)
		v = reflect.Append(v, src)
	}

	if fnMethod == Random {
		rand.Shuffle(len(fn), func(i, j int) { fn[i], fn[j] = fn[j], fn[i] })
	}

	ws := workers.New(limit)
	single := workers.New(1)
	switch fnMethod {
	case Concurrent:
		result := make(chan interface{}, 1)
		lasterr := make(chan error, 1)

		if nilArg {
			ctx := newContext(len(fn), 0, nil)
			defer ctx.cancel()

			ws.Slice(fn, func(_ int, fn interface{}) {
				ctx.runFn(fn.(func(interface{}) (interface{}, error)), result, lasterr)
			})

			if err := <-lasterr; err != nil {
				return nil, err
			}

			return <-result, nil
		}

		for i := 0; i < count; i++ {
			ctx := newContext(len(fn), argKey, v.Index(i).Interface())
			ws.Slice(fn, func(_ int, fn interface{}) {
				ctx.runFn(fn.(func(interface{}) (interface{}, error)), result, lasterr)
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
					rand.Shuffle(count, func(i, j int) {
						a := v.Index(i).Interface()
						b := v.Index(j).Interface()
						v.Index(j).Set(reflect.ValueOf(a))
						v.Index(i).Set(reflect.ValueOf(b))
					})
				}

				var worker *workers.Workers
				switch argMethod {
				case Concurrent:
					worker = ws
				case Serial, Random:
					worker = single
				default:
					return nil, errors.New("unknown arg method")
				}
				worker.Slice(v.Interface(), func(_ int, i interface{}) {
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
func ExecuteConcurrentArg(arg interface{}, fn ...func(interface{}) (interface{}, error)) (interface{}, error) {
	return Execute(Concurrent, Serial, defaultLimit, arg, fn...)
}

// ExecuteConcurrentFn gets the fastest result from the functions with args, functions will be run concurrently.
func ExecuteConcurrentFn(arg interface{}, fn ...func(interface{}) (interface{}, error)) (interface{}, error) {
	return Execute(Serial, Concurrent, defaultLimit, arg, fn...)
}

// ExecuteSerial gets the result until success from the functions with args in order.
func ExecuteSerial(arg interface{}, fn ...func(interface{}) (interface{}, error)) (interface{}, error) {
	return Execute(Serial, Serial, defaultLimit, arg, fn...)
}

// ExecuteRandom gets the result until success from the functions with args randomly.
func ExecuteRandom(arg interface{}, fn ...func(interface{}) (interface{}, error)) (interface{}, error) {
	return Execute(Random, Random, defaultLimit, arg, fn...)
}
