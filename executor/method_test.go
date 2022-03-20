package executor

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestExecuteConcurrent1(t *testing.T) {
	result, err := ExecuteConcurrentArg(
		[]int{0, 1, 2},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * time.Duration(n))
			return n * 2, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if expect := 0; result != expect {
		t.Errorf("expected %d; got %v", expect, result)
	}

	_, err = ExecuteConcurrentArg(
		[]int{0, 1, 2},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * time.Duration(n))
			return nil, fmt.Errorf("%v", n*2)
		},
	)
	if expect := "4"; err.Error() != expect {
		t.Errorf("expected error %s; got %v", expect, err)
	}
}

func TestExecuteConcurrent2(t *testing.T) {
	result, err := ExecuteConcurrentFn(
		[]int{1},
		func(n int) (interface{}, error) {
			return n * 0 * 2, nil
		},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * 1 * time.Duration(n))
			return n * 1 * 2, nil
		},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * 2 * time.Duration(n))
			return n * 2 * 2, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if expect := 0; result != expect {
		t.Errorf("expected %d; got %v", expect, result)
	}

	_, err = ExecuteConcurrentFn(
		[]int{1},
		func(n int) (interface{}, error) {
			return nil, fmt.Errorf("%v", n*0*2)
		},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * 1 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*1*2)
		},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * 2 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*2*2)
		},
	)
	if expect := "4"; err.Error() != expect {
		t.Errorf("expected error %s; got %v", expect, err)
	}
}

func TestExecuteSerial1(t *testing.T) {
	result, err := ExecuteSerial(
		[]int{0, 1, 2},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * time.Duration(n))
			return n * 2, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if expect := 0; result != expect {
		t.Errorf("expected %d; got %v", expect, result)
	}

	_, err = ExecuteSerial(
		[]int{0, 1, 2},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * time.Duration(n))
			return nil, fmt.Errorf("%v", n*2)
		},
	)
	if expect := "4"; err.Error() != expect {
		t.Errorf("expected error %s; got %v", expect, err)
	}
}

func TestExecuteSerial2(t *testing.T) {
	result, err := ExecuteSerial(
		[]int{1},
		func(n int) (interface{}, error) {
			return n * 0 * 2, nil
		},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * 1 * time.Duration(n))
			return n * 1 * 2, nil
		},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * 2 * time.Duration(n))
			return n * 2 * 2, nil
		},
	)
	if err != nil {
		t.Fatal(err)
	}
	if expect := 0; result != expect {
		t.Errorf("expected %d; got %v", expect, result)
	}

	_, err = ExecuteSerial(
		[]int{1},
		func(n int) (interface{}, error) {
			return nil, fmt.Errorf("%v", n*0*2)
		},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * 1 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*1*2)
		},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * 2 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*2*2)
		},
	)
	if expect := "4"; err.Error() != expect {
		t.Errorf("expected error %s; got %v", expect, err)
	}
}

func TestExecuteRandom1(t *testing.T) {
	testcase := []interface{}{"a", "b", "c", "d", "e", "f", "g"}
	var result []interface{}
	_, err := ExecuteRandom(
		testcase,
		func(i interface{}) (interface{}, error) {
			result = append(result, i)
			return nil, fmt.Errorf("%v", i)
		},
	)
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if reflect.DeepEqual(testcase, result) {
		t.Error("expected not equal; got equal")
	}
	sort.Slice(result, func(i, j int) bool { return result[i].(string) < result[j].(string) })
	if !reflect.DeepEqual(testcase, result) {
		t.Errorf("expected %v; got %v", testcase, result)
	}
}

func TestExecuteRandom2(t *testing.T) {
	var result []string
	_, err := ExecuteRandom(
		nil,
		func(i interface{}) (interface{}, error) {
			result = append(result, "a")
			return nil, fmt.Errorf("a")
		},
		func(i interface{}) (interface{}, error) {
			result = append(result, "b")
			return nil, fmt.Errorf("b")
		},
		func(i interface{}) (interface{}, error) {
			result = append(result, "c")
			return nil, fmt.Errorf("c")
		},
		func(i interface{}) (interface{}, error) {
			result = append(result, "d")
			return nil, fmt.Errorf("d")
		},
		func(i interface{}) (interface{}, error) {
			result = append(result, "e")
			return nil, fmt.Errorf("e")
		},
		func(i interface{}) (interface{}, error) {
			result = append(result, "f")
			return nil, fmt.Errorf("f")
		},
		func(i interface{}) (interface{}, error) {
			result = append(result, "g")
			return nil, fmt.Errorf("g")
		},
	)
	if err == nil {
		t.Fatal("expected error; got nil")
	}

	expect := []string{"a", "b", "c", "d", "e", "f", "g"}
	if reflect.DeepEqual(expect, result) {
		t.Errorf("expected not equal; got equal: %v", result)
	}
	sort.Strings(result)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("expected equal; got not equal: %v", result)
	}
}

func TestLimit(t *testing.T) {
	_, err := Execute(
		Serial,
		Concurrent,
		defaultLimit,
		[]int{1},
		func(n int) (interface{}, error) {
			return nil, fmt.Errorf("%v", n*0*2)
		},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * 1 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*1*2)
		},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * 2 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*2*2)
		},
	)
	if expect := "4"; err.Error() != expect {
		t.Errorf("expected error %s; got %v", expect, err)
	}

	_, err = Execute(
		Serial,
		Concurrent,
		1,
		[]int{1},
		func(n int) (interface{}, error) {
			return nil, fmt.Errorf("%v", n*0*2)
		},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * 1 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*1*2)
		},
		func(n int) (interface{}, error) {
			time.Sleep(time.Second * 2 * time.Duration(n))
			return nil, fmt.Errorf("%v", n*2*2)
		},
	)
	if expect := "4"; err.Error() != expect {
		t.Errorf("expected error %s; got %v", expect, err)
	}
}
