package workers

import (
	"math/rand"
	"reflect"
	"sort"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSlice(t *testing.T) {
	type test struct {
		char  string
		times int
	}
	slice := []test{{"a", 1}, {"b", 2}, {"c", 3}}

	result := make([]string, len(slice))
	if err := Slice(slice, func(i int, item interface{}) {
		result[i] = strings.Repeat(item.(test).char, item.(test).times)
	}); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual([]string{"a", "bb", "ccc"}, result) {
		t.Errorf("expected %q; got %q", []string{"a", "bb", "ccc"}, result)
	}
}

func TestMap(t *testing.T) {
	var m sync.Mutex
	var result []string
	if err := Map(map[string]int{"a": 1, "b": 2, "c": 3}, func(k, v interface{}) {
		m.Lock()
		result = append(result, strings.Repeat(k.(string), v.(int)))
		m.Unlock()
	}); err != nil {
		t.Fatal(err)
	}

	sort.Strings(result)
	if !reflect.DeepEqual([]string{"a", "bb", "ccc"}, result) {
		t.Errorf("expected %q; got %q", []string{"a", "bb", "ccc"}, result)
	}
}

func TestRange(t *testing.T) {
	end := 3
	items := []string{"a", "b", "c"}
	result := make([]string, end)
	if err := Range(1, end, func(num int) {
		result[num-1] = strings.Repeat(items[num-1], num)
	}); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual([]string{"a", "bb", "ccc"}, result) {
		t.Errorf("expected %q; got %q", []string{"a", "bb", "ccc"}, result)
	}
}

func TestLimit(t *testing.T) {
	rand.Seed(time.Now().UnixNano())

	limit := rand.Intn(1000) + 51
	workers := rand.Intn(50) + 1

	var count1, count2, count3, count4 int
	go Range(1, limit, func(_ int) {
		count1++
		for {
			select {}
		}
	})
	go New(0).Range(1, limit, func(_ int) {
		count2++
		for {
			select {}
		}
	})
	go New(-1).Range(1, limit, func(_ int) {
		count3++
		for {
			select {}
		}
	})
	go New(workers).Range(1, limit, func(_ int) {
		count4++
		for {
			select {}
		}
	})

	time.Sleep(time.Second)

	if count1 != NumCPU() {
		t.Errorf("expected %d; got %d", NumCPU(), count1)
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
	if limit := defaultWorkers.limit; limit != 5 {
		t.Errorf("expected %d; got %d", 5, limit)
	}

	SetLimit(100)
	if limit := defaultWorkers.limit; limit != 100 {
		t.Errorf("expected %d; got %d", 100, limit)
	}
}
