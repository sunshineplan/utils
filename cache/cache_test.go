package cache

import (
	"testing"
	"time"
)

func TestSetGetDelete(t *testing.T) {
	cache := New[string, string](false)
	cache.Set("key", "value", 0, nil)
	if value, ok := cache.Get("key"); !ok {
		t.Fatal("expected ok; got not")
	} else if value != "value" {
		t.Errorf("expected value; got %q", value)
	}
	cache.Delete("key")
	if _, ok := cache.Get("key"); ok {
		t.Error("expected not ok; got ok")
	}
}

func TestEmpty(t *testing.T) {
	cache := New[string, int](false)
	cache.Set("a", 1, 0, nil)
	cache.Set("b", 2, 0, nil)
	cache.Set("c", 3, 0, nil)
	for _, i := range []string{"a", "b", "c"} {
		if _, ok := cache.Get(i); !ok {
			t.Error("expected ok; got not")
		}
	}
	cache.Empty()
	for _, i := range []string{"a", "b", "c"} {
		if _, ok := cache.Get(i); ok {
			t.Error("expected not ok; got ok")
		}
	}
}

func TestRenew(t *testing.T) {
	cache := New[string, string](true)
	expire := make(chan struct{})
	cache.Set("renew", "old", 2*time.Second, func() (string, error) {
		defer func() { close(expire) }()
		return "new", nil
	})
	cache.Set("expire", "value", 1*time.Second, nil)
	if value, ok := cache.Get("expire"); !ok {
		t.Fatal("expected ok; got not")
	} else if expect := "value"; value != expect {
		t.Errorf("expected %q; got %q", expect, value)
	}
	if value, ok := cache.Get("renew"); !ok {
		t.Fatal("expected ok; got not")
	} else if expect := "old"; value != expect {
		t.Errorf("expected %q; got %q", expect, value)
	}
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()
	select {
	case <-expire:
		time.Sleep(100 * time.Millisecond)
		if _, ok := cache.Get("expire"); ok {
			t.Error("expected not ok; got ok")
		}
		value, ok := cache.Get("renew")
		if !ok {
			t.Fatal("expected ok; got not")
		} else if expect := "new"; value != expect {
			t.Errorf("expected %q; got %q", expect, value)
		}
	case <-ticker.C:
		t.Fatal("time out")
	}
}
