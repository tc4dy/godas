package agg_test

import (
	"math"
	"testing"

	"github.com/tc4dy/godas/agg"
	"github.com/tc4dy/godas/series"
)

func TestSumAgg(t *testing.T) {
	s := series.NewFloat64("x", []float64{10, 20, 30})
	v, err := agg.Sum("x").Apply(s)
	if err != nil {
		t.Fatal(err)
	}
	if v != 60.0 {
		t.Fatalf("expected 60, got %v", v)
	}
}

func TestMeanAgg(t *testing.T) {
	s := series.NewFloat64("x", []float64{10, 20, 30})
	v, err := agg.Mean("x").Apply(s)
	if err != nil {
		t.Fatal(err)
	}
	if v != 20.0 {
		t.Fatalf("expected 20, got %v", v)
	}
}

func TestMinAgg(t *testing.T) {
	s := series.NewFloat64("x", []float64{5, 3, 8})
	v, err := agg.Min("x").Apply(s)
	if err != nil {
		t.Fatal(err)
	}
	if v != 3.0 {
		t.Fatalf("expected 3, got %v", v)
	}
}

func TestMaxAgg(t *testing.T) {
	s := series.NewFloat64("x", []float64{5, 3, 8})
	v, err := agg.Max("x").Apply(s)
	if err != nil {
		t.Fatal(err)
	}
	if v != 8.0 {
		t.Fatalf("expected 8, got %v", v)
	}
}

func TestCountAgg(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, math.NaN(), 3})
	v, err := agg.Count("x").Apply(s)
	if err != nil {
		t.Fatal(err)
	}
	if v != 2.0 {
		t.Fatalf("expected 2, got %v", v)
	}
}

func TestStdAgg(t *testing.T) {
	s := series.NewFloat64("x", []float64{2, 4, 4, 4, 5, 5, 7, 9})
	v, err := agg.Std("x").Apply(s)
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(v-2.138) > 0.01 {
		t.Fatalf("expected ~2.138, got %v", v)
	}
}

func TestAggAlias(t *testing.T) {
	a := agg.Sum("revenue").As("total_revenue")
	if a.Alias() != "total_revenue" {
		t.Fatalf("expected total_revenue, got %s", a.Alias())
	}
}

func TestFirstLastAgg(t *testing.T) {
	s := series.NewFloat64("x", []float64{10, 20, 30})
	first, _ := agg.First("x").Apply(s)
	last, _ := agg.Last("x").Apply(s)
	if first != 10.0 {
		t.Fatalf("expected first 10, got %v", first)
	}
	if last != 30.0 {
		t.Fatalf("expected last 30, got %v", last)
	}
}

func TestAggOnNonNumeric(t *testing.T) {
	s := series.NewString("x", []string{"a", "b"})
	_, err := agg.Sum("x").Apply(s)
	if err == nil {
		t.Fatal("expected error for non-numeric column")
	}
}
