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

	done := make(chan bool)
	cache.Set("regenerate", "old", 2*time.Second, func() (interface{}, error) {
		defer func() { done <- true }()
		return "new", nil
	})
	cache.Set("expire", "value", 1*time.Second, nil)

	value, ok := cache.Get("expire")
	if !ok {
		t.Fatal("expected ok; got not")
	}
	if expect := "value"; value != expect {
		t.Errorf("expected %q; got %q", expect, value)
	}

	value, ok = cache.Get("regenerate")
	if !ok {
		t.Fatal("expected ok; got not")
	}
	if expect := "old"; value != expect {
		t.Errorf("expected %q; got %q", expect, value)
	}

	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	select {
	case <-done:
		if _, ok := cache.Get("expire"); ok {
			t.Error("expected not ok; got ok")
		}

		value, ok := cache.Get("regenerate")
		if !ok {
			t.Fatal("expected ok; got not")
		}
		if expect := "new"; value != expect {
			t.Errorf("expected %q; got %q", expect, value)
		}
	case <-ticker.C:
		t.Fatal("time out")
	}
}
