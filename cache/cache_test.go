package cache

import (
	"runtime"
	"testing"
	"weak"
)

func TestCache(t *testing.T) {
	cache := New[string, string]()
	key := "key"
	p := weak.Make(&key)
	value := "value"
	cache.Set(&key, value)
	if v, ok := cache.Get(&key); !ok {
		t.Fatal("expected cached, got not")
	} else if v != value {
		t.Fatalf("expected %q, got %q", value, v)
	}
	if v, ok := cache.m.Load(p); !ok {
		t.Fatal("expected cached, got not")
	} else if v != value {
		t.Fatalf("expected %q, got %q", value, v)
	}
	runtime.GC()
	if v, ok := cache.Get(&key); !ok {
		t.Fatal("expected cached, got not")
	} else if v != value {
		t.Fatalf("expected %q, got %q", value, v)
	}
	if v, ok := cache.m.Load(p); !ok {
		t.Fatal("expected cached, got not")
	} else if v != value {
		t.Fatalf("expected %q, got %q", value, v)
	}
	runtime.GC()
	if _, ok := cache.m.Load(p); ok {
		t.Fatal("expected not cached, got cached")
	}
}
