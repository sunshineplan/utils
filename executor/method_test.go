package executor

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestExecuteConcurrent1(t *testing.T) {
	c := make(chan error)
	var result interface{}
	var err1, err2 error

	t1 := time.NewTicker(200 * time.Millisecond)
	defer t1.Stop()
	go func() {
		result, err1 = ExecuteConcurrentArg(
			[]int{0, 1, 2},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * time.Duration(n.(int)))
				return n.(int) * 2, nil
			},
		)
		c <- err1
	}()
	select {
	case err := <-c:
		if err != nil {
			t.Fatal(err)
		}
		if expect := 0; result != expect {
			t.Errorf("expected %d; got %v", expect, result)
		}
	case <-t1.C:
		t.Error("t1 time out")
	}

	t2 := time.NewTicker(2*time.Second + 200*time.Millisecond)
	defer t2.Stop()
	go func() {
		_, err2 = ExecuteConcurrentArg(
			[]int{0, 1, 2},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * time.Duration(n.(int)))
				return nil, fmt.Errorf("%v", n.(int)*2)
			},
		)
		c <- err2
	}()
	select {
	case err := <-c:
		if expect := "4"; err.Error() != expect {
			t.Errorf("expected error %s; got %v", expect, err)
		}
	case <-t2.C:
		t.Error("t2 time out")
	}
}

func TestExecuteConcurrent2(t *testing.T) {
	c := make(chan error)
	var result interface{}
	var err1, err2 error

	t1 := time.NewTicker(200 * time.Millisecond)
	defer t1.Stop()
	go func() {
		result, err1 = ExecuteConcurrentFn(
			1,
			func(n interface{}) (interface{}, error) {
				return n.(int) * 0 * 2, nil
			},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * 1 * time.Duration(n.(int)))
				return n.(int) * 1 * 2, nil
			},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * 2 * time.Duration(n.(int)))
				return n.(int) * 2 * 2, nil
			},
		)
		c <- err1
	}()
	select {
	case err := <-c:
		if err != nil {
			t.Fatal(err)
		}
		if expect := 0; result != expect {
			t.Errorf("expected %d; got %v", expect, result)
		}
	case <-t1.C:
		t.Error("t1 time out")
	}

	t2 := time.NewTicker(2*time.Second + 200*time.Millisecond)
	defer t2.Stop()
	go func() {
		_, err2 = ExecuteConcurrentFn(
			1,
			func(n interface{}) (interface{}, error) {
				return nil, fmt.Errorf("%v", n.(int)*0*2)
			},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * 1 * time.Duration(n.(int)))
				return nil, fmt.Errorf("%v", n.(int)*1*2)
			},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * 2 * time.Duration(n.(int)))
				return nil, fmt.Errorf("%v", n.(int)*2*2)
			},
		)
		c <- err2
	}()
	select {
	case err := <-c:
		if expect := "4"; err.Error() != expect {
			t.Errorf("expected error %s; got %v", expect, err)
		}
	case <-t2.C:
		t.Error("t2 time out")
	}
}

func TestExecuteSerial1(t *testing.T) {
	var result interface{}
	var err error

	c1 := make(chan error)
	t1 := time.NewTicker(200 * time.Millisecond)
	defer t1.Stop()
	go func() {
		result, err = ExecuteSerial(
			[]int{0, 1, 2},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * time.Duration(n.(int)))
				return n.(int) * 2, nil
			},
		)
		c1 <- err
	}()
	select {
	case err := <-c1:
		if err != nil {
			t.Fatal(err)
		}
		if expect := 0; result != expect {
			t.Errorf("expected %d; got %v", expect, result)
		}
	case <-t1.C:
		t.Error("t1 time out")
	}

	c2 := make(chan error)
	t2 := time.NewTicker(3*time.Second + 600*time.Millisecond)
	defer t2.Stop()
	go func() {
		_, err = ExecuteSerial(
			[]int{0, 1, 2},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * time.Duration(n.(int)))
				return nil, fmt.Errorf("%v", n.(int)*2)
			},
		)
		c2 <- err
	}()
	select {
	case err := <-c2:
		if expect := "4"; err.Error() != expect {
			t.Errorf("expected error %s; got %v", expect, err)
		}
	case <-t2.C:
		t.Error("t2 time out")
	}
}

func TestExecuteSerial2(t *testing.T) {
	var result interface{}
	var err error

	c1 := make(chan error)
	t1 := time.NewTicker(200 * time.Millisecond)
	defer t1.Stop()
	go func() {
		result, err = ExecuteSerial(
			1,
			func(n interface{}) (interface{}, error) {
				return n.(int) * 0 * 2, nil
			},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * 1 * time.Duration(n.(int)))
				return n.(int) * 1 * 2, nil
			},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * 2 * time.Duration(n.(int)))
				return n.(int) * 2 * 2, nil
			},
		)
		c1 <- err
	}()
	select {
	case err := <-c1:
		if err != nil {
			t.Fatal(err)
		}
		if expect := 0; result != expect {
			t.Errorf("expected %d; got %v", expect, result)
		}
	case <-t1.C:
		t.Error("t1 time out")
	}

	c2 := make(chan error)
	t2 := time.NewTicker(3*time.Second + 600*time.Millisecond)
	defer t2.Stop()
	go func() {
		_, err = ExecuteSerial(
			1,
			func(n interface{}) (interface{}, error) {
				return nil, fmt.Errorf("%v", n.(int)*0*2)
			},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * 1 * time.Duration(n.(int)))
				return nil, fmt.Errorf("%v", n.(int)*1*2)
			},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * 2 * time.Duration(n.(int)))
				return nil, fmt.Errorf("%v", n.(int)*2*2)
			},
		)
		c2 <- err
	}()
	select {
	case err := <-c2:
		if expect := "4"; err.Error() != expect {
			t.Errorf("expected error %s; got %v", expect, err)
		}
	case <-t2.C:
		t.Error("t2 time out")
	}
}

func TestExecuteRandom1(t *testing.T) {
	testcase := []string{"a", "b", "c", "d", "e"}
	var result []string
	_, err := ExecuteRandom(
		testcase,
		func(i interface{}) (interface{}, error) {
			result = append(result, i.(string))
			return nil, fmt.Errorf("%v", i)
		},
	)
	if err == nil {
		t.Fatal("expected error; got nil")
	}
	if reflect.DeepEqual(testcase, result) {
		t.Error("expected not equal; got equal")
	}
	sort.Strings(result)
	if !reflect.DeepEqual(testcase, result) {
		t.Error("expected equal; got not equal")
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
	)
	if err == nil {
		t.Fatal("expected error; got nil")
	}

	expect := []string{"a", "b", "c", "d", "e"}
	if reflect.DeepEqual(expect, result) {
		t.Errorf("expected not equal; got equal: %v", result)
	}
	sort.Strings(result)
	if !reflect.DeepEqual(expect, result) {
		t.Errorf("expected equal; got not equal: %v", result)
	}
}
