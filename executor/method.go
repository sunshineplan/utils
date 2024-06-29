package executor

import (
	"errors"
	"math/rand/v2"

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

// SetLimit sets default limit.
// If n < 0, represents no limit.
// If n = 0, limit will be set as CPU number.
func SetLimit(n int) {
	defaultLimit = n
}

// Execute gets the result from the functions with several args by specified method.
// If both argMethod and fnMethod is Concurrent, fnMethod will be first.
func Execute[Arg, Res any](argMethod, fnMethod Method, limit int, args []Arg, fn ...func(Arg) (Res, error)) (res Res, err error) {
	if len(fn) == 0 {
		err = errors.New("no function provided")
		return
	}

	var clone []Arg
	var count int
	var nilArg bool
	switch len(args) {
	case 0:
		nilArg = true
	default:
		count = len(args)
		clone = make([]Arg, count)
		copy(clone, args)
	}

	switch fnMethod {
	case Concurrent:
		result := make(chan Res, 1)
		lasterr := make(chan error, 1)

		if nilArg {
			ctx := argContext[Arg, Res](len(fn), *new(Arg))
			defer ctx.cancel()

			workers.RunSlice(limit, fn, func(_ int, fn func(Arg) (Res, error)) {
				ctx.runFn(fn, result, lasterr)
			})

			if err = <-lasterr; err != nil {
				return
			}

			return <-result, nil
		}

		if argMethod == Random {
			rand.Shuffle(count, func(i, j int) { clone[i], clone[j] = clone[j], clone[i] })
		}

		for i := range count {
			ctx := argContext[Arg, Res](len(fn), clone[i])
			workers.RunSlice(limit, fn, func(_ int, fn func(Arg) (Res, error)) {
				ctx.runFn(fn, result, lasterr)
			})

			if err = <-lasterr; err == nil {
				res = <-result
				return
			} else if i == count-1 {
				return
			}
			ctx.cancel()
		}
	case Serial, Random:
		if fnMethod == Random {
			rand.Shuffle(len(fn), func(i, j int) { fn[i], fn[j] = fn[j], fn[i] })
		}

		for i, f := range fn {
			result := make(chan Res, 1)
			lasterr := make(chan error, 1)

			ctx := fnContext(count, f)
			if nilArg {
				ctx.runArg(*new(Arg), result, lasterr)
			} else {
				var worker int
				switch argMethod {
				case Concurrent:
					worker = limit
				case Serial, Random:
					worker = 1
				default:
					err = errors.New("unknown arg method")
					return
				}

				if argMethod == Random {
					rand.Shuffle(count, func(i, j int) { clone[i], clone[j] = clone[j], clone[i] })
				}

				workers.RunSlice(worker, clone, func(_ int, arg Arg) {
					ctx.runArg(arg, result, lasterr)
				})
			}

			if err = <-lasterr; err == nil {
				res = <-result
				return
			} else if i == len(fn)-1 {
				return
			}
			ctx.cancel()
		}
	default:
		err = errors.New("unknown function method")
		return
	}
	err = errors.New("unknown error")
	return
}

// ExecuteConcurrentArg gets the fastest result from the functions with args, args will be run concurrently.
func ExecuteConcurrentArg[Arg, Res any](arg []Arg, fn ...func(Arg) (Res, error)) (Res, error) {
	return Execute(Concurrent, Serial, defaultLimit, arg, fn...)
}

// ExecuteConcurrentFn gets the fastest result from the functions with args, functions will be run concurrently.
func ExecuteConcurrentFn[Arg, Res any](arg []Arg, fn ...func(Arg) (Res, error)) (Res, error) {
	return Execute(Serial, Concurrent, defaultLimit, arg, fn...)
}

// ExecuteSerial gets the result until success from the functions with args in order.
func ExecuteSerial[Arg, Res any](arg []Arg, fn ...func(Arg) (Res, error)) (Res, error) {
	return Execute(Serial, Serial, defaultLimit, arg, fn...)
}

// ExecuteRandom gets the result until success from the functions with args randomly.
func ExecuteRandom[Arg, Res any](arg []Arg, fn ...func(Arg) (Res, error)) (Res, error) {
	return Execute(Random, Random, defaultLimit, arg, fn...)
}
