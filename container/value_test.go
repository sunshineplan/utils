package container

import (
	"math/rand/v2"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/sunshineplan/utils"
)

func TestValue(t *testing.T) {
	v := NewValue[int]()
	if _, ok := v.Load(); ok {
		t.Fatal("initial Value is not nil")
	}
	v.Store(42)
	if i, ok := v.Load(); !ok || i != 42 {
		t.Fatalf("wrong value: got %d, want 42", i)
	}
	v.Store(84)
	if i, ok := v.Load(); !ok || i != 84 {
		t.Fatalf("wrong value: got %d, want 84", i)
	}
}

func TestValueLarge(t *testing.T) {
	v := NewValue[string]()
	v.Store("foo")
	if s, ok := v.Load(); !ok || s != "foo" {
		t.Fatalf("wrong value: got %s, want foo", s)
	}
	v.Store("barbaz")
	if s, ok := v.Load(); !ok || s != "barbaz" {
		t.Fatalf("wrong value: got %s, want barbaz", s)
	}
}

func TestValuePanic(t *testing.T) {
	const nilErr = "cache/value: store of nil value into Value"
	v := NewValue[any]()
	func() {
		defer func() {
			err := recover()
			if err != nilErr {
				t.Fatalf("inconsistent store panic: got '%v', want '%v'", err, nilErr)
			}
		}()
		v.Store(nil)
	}()
	v.Store(1)
	func() {
		defer func() {
			err := recover()
			if err != nilErr {
				t.Fatalf("inconsistent store panic: got '%v', want '%v'", err, nilErr)
			}
		}()
		v.Store(nil)
	}()
}

func TestPointType(t *testing.T) {
	v := NewValue[*int]()
	if v, stored := v.Load(); stored {
		t.Fatal("wrong stored status")
	} else if v != nil {
		t.Fatal("initial Value is not nil")
	}
	v.Store((*int)(nil))
	if v, stored := v.Load(); !stored {
		t.Fatal("wrong stored status")
	} else if v != nil {
		t.Fatalf("wrong value: got %v, want nil", v)
	}
	if old, stored := v.Swap(utils.Ptr(1)); !stored {
		t.Fatal("wrong stored status")
	} else if old != nil {
		t.Fatalf("wrong value: got %v, want nil", v)
	}
}

func TestValueConcurrent(t *testing.T) {
	tests := [][]any{
		{uint16(0), ^uint16(0), uint16(1 + 2<<8), uint16(3 + 4<<8)},
		{uint32(0), ^uint32(0), uint32(1 + 2<<16), uint32(3 + 4<<16)},
		{uint64(0), ^uint64(0), uint64(1 + 2<<32), uint64(3 + 4<<32)},
		{complex(0, 0), complex(1, 2), complex(3, 4), complex(5, 6)},
	}
	p := 4 * runtime.GOMAXPROCS(0)
	N := int(1e5)
	if testing.Short() {
		p /= 2
		N = 1e3
	}
	for _, test := range tests {
		v := NewValue[any]()
		done := make(chan bool, p)
		for i := 0; i < p; i++ {
			go func() {
				expected := true
			loop:
				for j := 0; j < N; j++ {
					x := test[rand.IntN(len(test))]
					v.Store(x)
					x = v.MustLoad()
					for _, x1 := range test {
						if x == x1 {
							continue loop
						}
					}
					t.Logf("loaded unexpected value %+v, want %+v", x, test)
					expected = false
					break
				}
				done <- expected
			}()
		}
		for i := 0; i < p; i++ {
			if !<-done {
				t.FailNow()
			}
		}
	}
}

func BenchmarkValueRead(b *testing.B) {
	v := NewValue[*int]()
	v.Store(new(int))
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			x := v.MustLoad()
			if *x != 0 {
				b.Fatalf("wrong value: got %v, want 0", *x)
			}
		}
	})
}

var Value_SwapTests = []struct {
	init any
	new  any
	want any
	err  any
}{
	{init: nil, new: nil, err: "cache/value: swap of nil value into Value"},
	{init: nil, new: true, want: nil, err: nil},
	{init: true, new: "", err: "sync/atomic: swap of inconsistently typed value into Value"},
	{init: true, new: false, want: true, err: nil},
}

func TestValue_Swap(t *testing.T) {
	for i, tt := range Value_SwapTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			v := NewValue[any]()
			if tt.init != nil {
				v.Store(tt.init)
			}
			defer func() {
				err := recover()
				switch {
				case tt.err == nil && err != nil:
					t.Errorf("should not panic, got %v", err)
				case tt.err != nil && err == nil:
					t.Errorf("should panic %v, got <nil>", tt.err)
				}
			}()
			if got, _ := v.Swap(tt.new); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
			if got, stored := v.Load(); !stored || got != tt.new {
				t.Errorf("got %v, want %v", got, tt.new)
			}
		})
	}
}

func TestValueSwapConcurrent(t *testing.T) {
	v := NewValue[uint64]()
	var count uint64
	var g sync.WaitGroup
	var m, n uint64 = 10000, 10000
	if testing.Short() {
		m = 1000
		n = 1000
	}
	for i := uint64(0); i < m*n; i += n {
		i := i
		g.Add(1)
		go func() {
			var c uint64
			for new := i; new < i+n; new++ {
				if old, stored := v.Swap(new); stored {
					c += old
				}
			}
			atomic.AddUint64(&count, c)
			g.Done()
		}()
	}
	g.Wait()
	if want, got := (m*n-1)*(m*n)/2, count+v.MustLoad(); got != want {
		t.Errorf("sum from 0 to %d was %d, want %v", m*n-1, got, want)
	}
}

var heapA, heapB = struct{ uint }{0}, struct{ uint }{0}

var Value_CompareAndSwapTests = []struct {
	init any
	new  any
	old  any
	want bool
	err  any
}{
	{init: nil, new: nil, old: nil, err: "cache/value: compare and swap of nil value into Value"},
	{init: nil, new: true, old: "", err: "sync/atomic: compare and swap of inconsistently typed values into Value"},
	{init: nil, new: true, old: true, want: false, err: nil},
	{init: nil, new: true, old: nil, want: true, err: nil},
	{init: true, new: "", err: "sync/atomic: compare and swap of inconsistently typed value into Value"},
	{init: true, new: true, old: false, want: false, err: nil},
	{init: true, new: true, old: true, want: true, err: nil},
	{init: heapA, new: struct{ uint }{1}, old: heapB, want: true, err: nil},
}

func TestValue_CompareAndSwap(t *testing.T) {
	for i, tt := range Value_CompareAndSwapTests {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			v := NewValue[any]()
			if tt.init != nil {
				v.Store(tt.init)
			}
			defer func() {
				err := recover()
				switch {
				case tt.err == nil && err != nil:
					t.Errorf("got %v, wanted no panic", err)
				case tt.err != nil && err == nil:
					t.Errorf("did not panic, want %v", tt.err)
				}
			}()
			if got := v.CompareAndSwap(tt.old, tt.new); got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValueCompareAndSwapConcurrent(t *testing.T) {
	v := NewValue[int]()
	var w sync.WaitGroup
	v.Store(0)
	m, n := 1000, 100
	if testing.Short() {
		m = 100
		n = 100
	}
	for i := 0; i < m; i++ {
		i := i
		w.Add(1)
		go func() {
			for j := i; j < m*n; runtime.Gosched() {
				if v.CompareAndSwap(j, j+1) {
					j += m
				}
			}
			w.Done()
		}()
	}
	w.Wait()
	if stop := v.MustLoad(); stop != m*n {
		t.Errorf("did not get to %v, stopped at %v", m*n, stop)
	}
}
