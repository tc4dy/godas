package stats_test

import (
	"testing"

	"github.com/tc4dy/godas/series"
	"github.com/tc4dy/godas/stats"
)

func BenchmarkStatsMean(b *testing.B) {
	s := series.NewFloat64("x", make([]float64, 1000000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats.Mean(s)
	}
}

func BenchmarkStatsStd(b *testing.B) {
	s := series.NewFloat64("x", make([]float64, 1000000))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stats.Std(s)
	}
}