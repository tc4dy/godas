package series_test

import (
	"testing"

	"github.com/tc4dy/godas/series"
)

func BenchmarkSeriesFilter(b *testing.B) {
	s := series.NewFloat64("x", make([]float64, 1000000))
	mask := make([]bool, 1000000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Filter(mask)
	}
}

func BenchmarkSeriesSortedIndices(b *testing.B) {
	s := series.NewFloat64("x", make([]float64, 1000000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.SortedIndices(true)
	}
}