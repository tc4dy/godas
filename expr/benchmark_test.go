package expr_test

import (
	"testing"

	"github.com/tc4dy/godas/expr"
	"github.com/tc4dy/godas/series"
)

func BenchmarkExprEval(b *testing.B) {
	cols := map[string]*series.Series{
		"x": series.NewFloat64("x", make([]float64, 1000000)),
	}
	e := expr.Col("x").Gt(0.5)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		e.Eval(cols, 1000000)
	}
}