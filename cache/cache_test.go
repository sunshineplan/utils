package cache

import (
	"testing"
	"time"
)

func TestSetGetDelete(t *testing.T) {
	cache := New(false)

	cache.Set("key", "value", 0, nil)

	value, ok := cache.Get("key")
	if !ok {
		t.Fatal("expected ok; got not")
	}
	if value != "value" {
		t.Errorf("expected value; got %q", value)
	}

	cache.Delete("key")
	_, ok = cache.Get("key")
	if ok {
		t.Error("expected not ok; got ok")
	}
}

func TestEmpty(t *testing.T) {
	cache := New(false)

	cache.Set("a", 1, 0, nil)
	cache.Set("b", 2, 0, nil)
	cache.Set("c", 3, 0, nil)

	for _, i := range []string{"a", "b", "c"} {
		_, ok := cache.Get(i)
		if !ok {
			t.Error("expected ok; got not")
		}
	}

	cache.Empty()

	for _, i := range []string{"a", "b", "c"} {
		_, ok := cache.Get(i)
		if ok {
			t.Error("expected not ok; got ok")
		}
	}
}

func TestAutoCleanRegenerate(t *testing.T) {
	cache := New(true)

	var newValue = []string{"1", "2", "3"}
	c := make(chan string, 3)
	for _, i := range newValue {
		c <- i
	}

	cache.Set("regenerate", "old", 2*time.Second, func() (interface{}, error) {
		return <-c, nil
	})
	cache.Set("expire", "value", 2*time.Second, nil)

	for _, i := range []string{"regenerate", "expire"} {
		_, ok := cache.Get(i)
		if !ok {
			t.Error("expected ok; got not")
		}
	}

	for _, i := range newValue {
		time.Sleep(3 * time.Second)

		_, ok := cache.Get("expire")
		if ok {
			t.Error("expected not ok; got ok")
		}
		value, ok := cache.Get("regenerate")
		if !ok {
			t.Fatal("expected ok; got not")
		}
		if value != i {
			t.Errorf("expected %q; got %q", i, value)
		}
	}
}
