package join_test

import (
	"testing"

	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/join"
	"github.com/tc4dy/godas/series"
)

func BenchmarkJoinInner(b *testing.B) {
	leftCol := series.NewInt64("id", make([]int64, 100000))
	rightCol := series.NewInt64("id", make([]int64, 100000))
	left, _ := dataframe.New(leftCol)
	right, _ := dataframe.New(rightCol)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		join.Inner(left, right, "id")
	}
}