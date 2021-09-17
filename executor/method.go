package executor

import (
	"errors"
	"math/rand"
	"reflect"
	"time"
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

func init() {
	rand.Seed(time.Now().UnixNano())
}

// Execute gets the result from the functions with several args by specified method.
// If both argMethod and fnMethod is Concurrent, fnMethod will be first.
func Execute(argMethod, fnMethod Method, arg interface{}, fn ...func(interface{}) (interface{}, error)) (interface{}, error) {
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

	switch fnMethod {
	case Concurrent:
		result := make(chan interface{}, 1)
		lasterr := make(chan error, 1)

		if nilArg {
			ctx := newArgContext(len(fn), nil)
			defer ctx.cancel()

			for _, f := range fn {
				go ctx.run(f, result, lasterr)
			}

			if err := <-lasterr; err != nil {
				return nil, err
			}

			return <-result, nil
		}

		for i := 0; i < count; i++ {
			ctx := newArgContext(len(fn), v.Index(i).Interface())
			for _, f := range fn {
				go ctx.run(f, result, lasterr)
			}

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

			ctx := newFnContext(count, f)
			if nilArg {
				switch argMethod {
				case Concurrent:
					go ctx.run(nil, result, lasterr)
				case Serial, Random:
					ctx.run(nil, result, lasterr)
				default:
					return nil, errors.New("unknown arg method")
				}
			} else {
				if argMethod == Random {
					rand.Shuffle(count, func(i, j int) {
						a := v.Index(i).Interface()
						b := v.Index(j).Interface()
						v.Index(j).Set(reflect.ValueOf(a))
						v.Index(i).Set(reflect.ValueOf(b))
					})
				}

				for i := 0; i < count; i++ {
					switch argMethod {
					case Concurrent:
						go ctx.run(v.Index(i).Interface(), result, lasterr)
					case Serial, Random:
						ctx.run(v.Index(i).Interface(), result, lasterr)
					default:
						return nil, errors.New("unknown arg method")
					}
				}
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
	return Execute(Concurrent, Serial, arg, fn...)
}

// ExecuteConcurrentFn gets the fastest result from the functions with args, functions will be run concurrently.
func ExecuteConcurrentFn(arg interface{}, fn ...func(interface{}) (interface{}, error)) (interface{}, error) {
	return Execute(Serial, Concurrent, arg, fn...)
}

// ExecuteSerial gets the result until success from the functions with args in order.
func ExecuteSerial(arg interface{}, fn ...func(interface{}) (interface{}, error)) (interface{}, error) {
	return Execute(Serial, Serial, arg, fn...)
}

// ExecuteRandom gets the result until success from the functions with args randomly.
func ExecuteRandom(arg interface{}, fn ...func(interface{}) (interface{}, error)) (interface{}, error) {
	return Execute(Random, Random, arg, fn...)
}
