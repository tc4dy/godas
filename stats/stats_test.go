package stats_test

import (
	"math"
	"testing"

	"github.com/tc4dy/godas/series"
	"github.com/tc4dy/godas/stats"
)

func TestSum(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2, 3})
	sum := stats.Sum(s)
	if sum != 6.0 {
		t.Fatalf("expected 6, got %v", sum)
	}
}

func TestMean(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2, 3})
	mean := stats.Mean(s)
	if mean != 2.0 {
		t.Fatalf("expected 2, got %v", mean)
	}
}

func TestMeanWithNaN(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, math.NaN(), 3})
	mean := stats.Mean(s)
	if mean != 2.0 {
		t.Fatalf("expected 2, got %v", mean)
	}
}

func TestMin(t *testing.T) {
	s := series.NewFloat64("x", []float64{3, 1, 2})
	min := stats.Min(s)
	if min != 1.0 {
		t.Fatalf("expected 1, got %v", min)
	}
}

func TestMax(t *testing.T) {
	s := series.NewFloat64("x", []float64{3, 1, 2})
	max := stats.Max(s)
	if max != 3.0 {
		t.Fatalf("expected 3, got %v", max)
	}
}

func TestStd(t *testing.T) {
	s := series.NewFloat64("x", []float64{2, 4, 4, 4, 5, 5, 7, 9})
	std := stats.Std(s)
	if math.Abs(std-2.138) > 0.01 {
		t.Fatalf("expected ~2.138, got %v", std)
	}
}

func TestMedian(t *testing.T) {
	s := series.NewFloat64("x", []float64{3, 1, 2})
	med := stats.Median(s)
	if med != 2.0 {
		t.Fatalf("expected 2, got %v", med)
	}
}

func TestCount(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, math.NaN(), 3})
	c := stats.Count(s)
	if c != 2 {
		t.Fatalf("expected 2, got %d", c)
	}
}

func TestNullCount(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, math.NaN(), 3})
	nc := stats.NullCount(s)
	if nc != 1 {
		t.Fatalf("expected 1, got %d", nc)
	}
}

func TestCorr(t *testing.T) {
	a := series.NewFloat64("a", []float64{1, 2, 3})
	b := series.NewFloat64("b", []float64{2, 4, 6})
	corr := stats.Corr(a, b)
	if math.Abs(corr-1.0) > 0.001 {
		t.Fatalf("expected ~1, got %v", corr)
	}
}

func TestCorrWithNaN(t *testing.T) {
	a := series.NewFloat64("a", []float64{1, 2, math.NaN()})
	b := series.NewFloat64("b", []float64{2, 4, 6})
	corr := stats.Corr(a, b)
	if math.Abs(corr-1.0) > 0.001 {
		t.Fatalf("expected ~1, got %v", corr)
	}
}

func TestDescribe(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2, 3, 4, 5})
	sum := stats.Describe(s)
	if sum == nil {
		t.Fatal("summary is nil")
	}
	if sum.Mean != 3.0 {
		t.Fatalf("expected mean 3, got %v", sum.Mean)
	}
	if sum.Min != 1.0 || sum.Max != 5.0 {
		t.Fatalf("expected min 1, max 5, got %v, %v", sum.Min, sum.Max)
	}
}
