package pool_test

import (
	"testing"

	"github.com/tc4dy/godas/pool"
)

func TestFloat64Pool(t *testing.T) {
	s := pool.GetFloat64(10)
	if cap(s) < 10 {
		t.Fatalf("expected capacity >=10, got %d", cap(s))
	}
	s = append(s, 1.0, 2.0)
	if len(s) != 2 {
		t.Fatalf("expected len 2, got %d", len(s))
	}
	pool.PutFloat64(s)
	s2 := pool.GetFloat64(5)
	if cap(s2) < 5 {
		t.Fatalf("expected capacity >=5, got %d", cap(s2))
	}
}

func TestStringPool(t *testing.T) {
	s := pool.GetString(5)
	if cap(s) < 5 {
		t.Fatalf("expected capacity >=5, got %d", cap(s))
	}
	s = append(s, "a", "b")
	pool.PutString(s)
	s2 := pool.GetString(3)
	if cap(s2) < 3 {
		t.Fatalf("expected capacity >=3, got %d", cap(s2))
	}
}

func TestBoolPool(t *testing.T) {
	s := pool.GetBool(4)
	if cap(s) < 4 {
		t.Fatalf("expected capacity >=4, got %d", cap(s))
	}
	s = append(s, true, false)
	pool.PutBool(s)
	s2 := pool.GetBool(2)
	if cap(s2) < 2 {
		t.Fatalf("expected capacity >=2, got %d", cap(s2))
	}
}