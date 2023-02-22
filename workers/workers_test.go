package workers

import (
	"context"
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestFunction(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	var n uint32
	Function(ctx, func(ctx context.Context) {
		if ctx.Err() == nil {
			atomic.AddUint32(&n, 1)
			time.Sleep(2 * time.Second)
		}
	})
	if expect, n := uint32(defaultWorkers), atomic.LoadUint32(&n); n != expect {
		t.Errorf("expected %v; got %v", expect, n)
	}
}

func TestSlice(t *testing.T) {
	type test struct {
		char  string
		times int
	}
	slice := []test{{"a", 1}, {"b", 2}, {"c", 3}}

	result := make([]string, len(slice))
	Slice(slice, func(i int, item test) {
		result[i] = strings.Repeat(item.char, item.times)
	})

	if expect := []string{"a", "bb", "ccc"}; !reflect.DeepEqual(expect, result) {
		t.Errorf("expected %v; got %v", expect, result)
	}
}

func TestMap(t *testing.T) {
	var m sync.Mutex
	var result []string
	Map(map[string]int{"a": 1, "b": 2, "c": 3}, func(k string, v int) {
		m.Lock()
		result = append(result, strings.Repeat(k, v))
		m.Unlock()
	})

	sort.Strings(result)
	if expect := []string{"a", "bb", "ccc"}; !reflect.DeepEqual(expect, result) {
		t.Errorf("expected %v; got %v", expect, result)
	}
}

func TestRange(t *testing.T) {
	end := 3
	items := []string{"a", "b", "c"}
	result := make([]string, end)
	Range(1, end, func(num int) {
		result[num-1] = strings.Repeat(items[num-1], num)
	})

	if expect := []string{"a", "bb", "ccc"}; !reflect.DeepEqual(expect, result) {
		t.Errorf("expected %v; got %v", expect, result)
	}
}

func TestLimit(t *testing.T) {
	limit := rand.Intn(1000) + 51
	workers := rand.Intn(50) + 1

	var mu1, mu2, mu3, mu4 sync.Mutex
	var count1, count2, count3, count4 int
	go Range(1, limit, func(_ int) {
		mu1.Lock()
		count1++
		mu1.Unlock()
		for {
			select {}
		}
	})
	go RunRange(0, 1, limit, func(_ int) {
		mu2.Lock()
		count2++
		mu2.Unlock()
		for {
			select {}
		}
	})
	go RunRange(-1, 1, limit, func(_ int) {
		mu3.Lock()
		count3++
		mu3.Unlock()
		for {
			select {}
		}
	})
	go RunRange(workers, 1, limit, func(_ int) {
		mu4.Lock()
		count4++
		mu4.Unlock()
		for {
			select {}
		}
	})

	time.Sleep(time.Second)

	mu1.Lock()
	mu2.Lock()
	mu3.Lock()
	mu4.Lock()

	if cpu := NumCPU(); count1 != cpu {
		t.Errorf("expected %d; got %d", cpu, count1)
	}
	if count2 != count1 {
		t.Errorf("expected %d; got %d", count1, count2)
	}
	if count3 != limit {
		t.Errorf("expected %d; got %d", limit, count3)
	}
	if count4 != workers {
		t.Errorf("expected %d; got %d", workers, count4)
	}
}

func TestSetLimit(t *testing.T) {
	SetLimit(5)
	if defaultWorkers != 5 {
		t.Errorf("expected %d; got %d", 5, defaultWorkers)
	}

	SetLimit(100)
	if defaultWorkers != 100 {
		t.Errorf("expected %d; got %d", 100, defaultWorkers)
	}
}
