package agg_test

import (
	"testing"

	"github.com/tc4dy/godas/agg"
	"github.com/tc4dy/godas/series"
)

func BenchmarkAggSum(b *testing.B) {
	s := series.NewFloat64("x", make([]float64, 1000000))
	a := agg.Sum("x")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		a.Apply(s)
	}
}