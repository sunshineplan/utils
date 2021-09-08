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

// Execute gets the result from the same function with several args by specified method.
func Execute(method Method, args interface{}, fn func(interface{}) (interface{}, error)) (interface{}, error) {
	if reflect.TypeOf(args).Kind() != reflect.Slice {
		return nil, errors.New("args must be a slice")
	}

	src := reflect.ValueOf(args)
	count := src.Len()
	if count == 0 {
		return nil, errors.New("args can't be empty slice")
	}
	v := reflect.MakeSlice(src.Type(), count, count)
	reflect.Copy(v, src)

	if method == Random {
		rand.Shuffle(count, func(i, j int) {
			a := v.Index(i).Interface()
			b := v.Index(j).Interface()
			v.Index(j).Set(reflect.ValueOf(a))
			v.Index(i).Set(reflect.ValueOf(b))
		})
	}

	result := make(chan interface{}, 1)
	lasterr := make(chan error, 1)
	ctx := newContext(count, fn)
	defer ctx.cancel()

	for i := 0; i < v.Len(); i++ {
		switch method {
		case Concurrent:
			go ctx.run(v.Index(i).Interface(), result, lasterr)
		case Serial, Random:
			ctx.run(v.Index(i).Interface(), result, lasterr)
		default:
			return nil, errors.New("unknown method")
		}
	}

	if err := <-lasterr; err != nil {
		return nil, err
	}

	return <-result, nil
}

// ExecuteConcurrent gets the fastest result from the same function with several args.
func ExecuteConcurrent(args interface{}, fn func(interface{}) (interface{}, error)) (interface{}, error) {
	return Execute(Concurrent, args, fn)
}

// ExecuteSerial gets the result until success from the same function with several args used in order.
func ExecuteSerial(args interface{}, fn func(interface{}) (interface{}, error)) (interface{}, error) {
	return Execute(Serial, args, fn)
}

// ExecuteRandom gets the result until success from the same function with several args used randomly.
func ExecuteRandom(args interface{}, fn func(interface{}) (interface{}, error)) (interface{}, error) {
	return Execute(Random, args, fn)
}
