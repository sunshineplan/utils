package executor

import (
	"fmt"
	"reflect"
	"sort"
	"testing"
	"time"
)

func TestExecuteConcurrent(t *testing.T) {
	c := make(chan error)
	var result interface{}
	var err1, err2 error

	t1 := time.NewTicker(time.Second + 200*time.Millisecond)
	defer t1.Stop()
	go func() {
		result, err1 = ExecuteConcurrent(
			[]int{1, 2, 3},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * time.Duration(n.(int)))
				return n, nil
			},
		)
		c <- err1
	}()
	select {
	case err := <-c:
		if err != nil {
			t.Errorf("expected nil error; got non-nil error %v", err)
		}
		if result != 1 {
			t.Errorf("expected %d; got %v", 1, result)
		}
	case <-t1.C:
		t.Error("time out")
	}

	t2 := time.NewTicker(3*time.Second + 200*time.Millisecond)
	defer t2.Stop()
	go func() {
		_, err2 = ExecuteConcurrent(
			[]int{1, 2, 3},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * time.Duration(n.(int)))
				return nil, fmt.Errorf("%v", n)
			},
		)
		c <- err2
	}()
	select {
	case err := <-c:
		if err == nil || err.Error() != "3" {
			t.Errorf("expected error %d; got %v", 3, err)
		}
	case <-t2.C:
		t.Error("time out")
	}
}

func TestExecuteSerial(t *testing.T) {
	var result interface{}
	var err error

	c1 := make(chan error)
	t1 := time.NewTicker(time.Second + 200*time.Millisecond)
	defer t1.Stop()
	go func() {
		result, err = ExecuteSerial(
			[]int{1, 2, 3},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * time.Duration(n.(int)))
				return n, nil
			},
		)
		c1 <- err
	}()
	select {
	case err := <-c1:
		if err != nil {
			t.Error(err)
		}
		if result != 1 {
			t.Errorf("expected %d; got %v", 1, result)
		}
	case <-t1.C:
		t.Error("time out")
	}

	c2 := make(chan error)
	t2 := time.NewTicker(3*time.Second + 600*time.Millisecond)
	defer t2.Stop()
	go func() {
		_, err = ExecuteSerial(
			[]int{0, 1, 2},
			func(n interface{}) (interface{}, error) {
				time.Sleep(time.Second * time.Duration(n.(int)))
				return nil, fmt.Errorf("%v", n)
			},
		)
		c2 <- err
	}()
	select {
	case err := <-c2:
		if err == nil || err.Error() != "2" {
			t.Errorf("expected error %d; got %v", 2, err)
		}
	case <-t2.C:
		t.Error("time out")
	}
}

func TestExecuteRandom(t *testing.T) {
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
		t.Error("expected error; got nil")
	}
	if reflect.DeepEqual(testcase, result) {
		t.Error("expected not equal; got equal")
	}
	sort.Strings(result)
	if !reflect.DeepEqual(testcase, result) {
		t.Error("expected equal; got not equal")
	}
}
