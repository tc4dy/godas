package series_test

import (
	"math"
	"testing"

	"github.com/tc4dy/godas/series"
)

func TestNewFloat64(t *testing.T) {
	s := series.NewFloat64("x", []float64{1.1, 2.2})
	if s.Len() != 2 {
		t.Fatalf("expected len 2, got %d", s.Len())
	}
	if s.DType() != series.Float64 {
		t.Fatalf("expected Float64, got %v", s.DType())
	}
	if s.Name() != "x" {
		t.Fatalf("expected name x, got %s", s.Name())
	}
}

func TestNewInt64(t *testing.T) {
	s := series.NewInt64("x", []int64{1, 2})
	if s.Len() != 2 {
		t.Fatalf("expected len 2, got %d", s.Len())
	}
}

func TestNewString(t *testing.T) {
	s := series.NewString("x", []string{"a", "b"})
	if s.Len() != 2 {
		t.Fatalf("expected len 2, got %d", s.Len())
	}
}

func TestNewBool(t *testing.T) {
	s := series.NewBool("x", []bool{true, false})
	if s.Len() != 2 {
		t.Fatalf("expected len 2, got %d", s.Len())
	}
}

func TestSeriesGetFloat64(t *testing.T) {
	s := series.NewFloat64("x", []float64{1.5, math.NaN(), 3.0})
	v, ok := s.GetFloat64(0)
	if !ok || v != 1.5 {
		t.Fatalf("expected 1.5, got %v", v)
	}
	_, ok = s.GetFloat64(1)
	if ok {
		t.Fatal("expected ok false for NaN")
	}
}

func TestSeriesGetString(t *testing.T) {
	s := series.NewString("x", []string{"a", "b"})
	v, ok := s.GetString(0)
	if !ok || v != "a" {
		t.Fatalf("expected a, got %v", v)
	}
}

func TestSeriesGetBool(t *testing.T) {
	s := series.NewBool("x", []bool{true, false})
	v, ok := s.GetBool(0)
	if !ok || v != true {
		t.Fatalf("expected true, got %v", v)
	}
}

func TestSeriesFilter(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2, 3})
	mask := []bool{true, false, true}
	f := s.Filter(mask)
	if f.Len() != 2 {
		t.Fatalf("expected len 2, got %d", f.Len())
	}
	vals := f.RawFloats()
	if vals[0] != 1 || vals[1] != 3 {
		t.Fatalf("expected [1,3], got %v", vals)
	}
}

func TestSeriesSlice(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2, 3, 4})
	sl := s.Slice(1, 3)
	if sl.Len() != 2 {
		t.Fatalf("expected len 2, got %d", sl.Len())
	}
	vals := sl.RawFloats()
	if vals[0] != 2 || vals[1] != 3 {
		t.Fatalf("expected [2,3], got %v", vals)
	}
}

func TestSeriesSortedIndices(t *testing.T) {
	s := series.NewFloat64("x", []float64{3, 1, 2})
	idx := s.SortedIndices(true)
	expected := []int{1, 2, 0}
	for i, v := range expected {
		if idx[i] != v {
			t.Fatalf("expected %v, got %v", expected, idx)
		}
	}
}

func TestSeriesReorder(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2, 3})
	indices := []int{2, 0, 1}
	r := s.Reorder(indices)
	vals := r.RawFloats()
	expected := []float64{3, 1, 2}
	for i, v := range expected {
		if vals[i] != v {
			t.Fatalf("expected %v, got %v", expected, vals)
		}
	}
}

func TestSeriesAppend(t *testing.T) {
	s1 := series.NewFloat64("x", []float64{1, 2})
	s2 := series.NewFloat64("x", []float64{3, 4})
	app, err := s1.Append(s2)
	if err != nil {
		t.Fatal(err)
	}
	if app.Len() != 4 {
		t.Fatalf("expected len 4, got %d", app.Len())
	}
	vals := app.RawFloats()
	expected := []float64{1, 2, 3, 4}
	for i, v := range expected {
		if vals[i] != v {
			t.Fatalf("expected %v, got %v", expected, vals)
		}
	}
}

func TestSeriesAppendTypeMismatch(t *testing.T) {
	s1 := series.NewFloat64("x", []float64{1})
	s2 := series.NewString("x", []string{"a"})
	_, err := s1.Append(s2)
	if err == nil {
		t.Fatal("expected error for type mismatch")
	}
}

func TestSeriesClone(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2, 3})
	c := s.Clone()
	c.RawFloats()[0] = 99
	if s.RawFloats()[0] == 99 {
		t.Fatal("clone modified original")
	}
}

func TestSeriesValueAt(t *testing.T) {
	s := series.NewFloat64("x", []float64{1.5, math.NaN()})
	v := s.ValueAt(0)
	if v != 1.5 {
		t.Fatalf("expected 1.5, got %v", v)
	}
	v = s.ValueAt(1)
	if v != nil {
		t.Fatalf("expected nil for NaN, got %v", v)
	}
}

func TestSeriesStringAt(t *testing.T) {
	s := series.NewFloat64("x", []float64{1.5, math.NaN()})
	str := s.StringAt(0)
	if str != "1.5000" {
		t.Fatalf("expected 1.5000, got %s", str)
	}
	str = s.StringAt(1)
	if str != "null" {
		t.Fatalf("expected null, got %s", str)
	}
}

func TestSeriesSetName(t *testing.T) {
	s := series.NewFloat64("old", []float64{1})
	s.SetName("new")
	if s.Name() != "new" {
		t.Fatalf("expected new, got %s", s.Name())
	}
}