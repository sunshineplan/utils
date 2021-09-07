package utils

import (
	"fmt"
	"testing"
	"time"
)

func TestLoadBalancer(t *testing.T) {
	result, err := LoadBalancer(
		[]interface{}{1, 3, 5},
		func(n interface{}) (interface{}, error) {
			time.Sleep(time.Second * time.Duration(n.(int)))
			return n, nil
		},
	)
	if err != nil {
		t.Error(err)
	}
	if result != 1 {
		t.Errorf("expected %d; got %v", 1, result)
	}

	if _, err := LoadBalancer(
		[]interface{}{1, 3, 5},
		func(n interface{}) (interface{}, error) {
			time.Sleep(time.Second * time.Duration(n.(int)))
			return nil, fmt.Errorf("%v", n)
		},
	); err == nil || err.Error() != "5" {
		t.Errorf("expected error %d; got %v", 5, err)
	}
}
