package agg

import (
	"fmt"
	"math"

	"github.com/tc4dy/godas/series"
	"github.com/tc4dy/godas/stats"
)

type Kind int

const (
	KindSum Kind = iota
	KindMean
	KindMin
	KindMax
	KindCount
	KindStd
	KindMedian
	KindFirst
	KindLast
)

type Agg struct {
	col    string
	kind   Kind
	alias  string
}

func Sum(col string) Agg    { return Agg{col: col, kind: KindSum, alias: "sum_" + col} }
func Mean(col string) Agg   { return Agg{col: col, kind: KindMean, alias: "mean_" + col} }
func Min(col string) Agg    { return Agg{col: col, kind: KindMin, alias: "min_" + col} }
func Max(col string) Agg    { return Agg{col: col, kind: KindMax, alias: "max_" + col} }
func Count(col string) Agg  { return Agg{col: col, kind: KindCount, alias: "count_" + col} }
func Std(col string) Agg    { return Agg{col: col, kind: KindStd, alias: "std_" + col} }
func Median(col string) Agg { return Agg{col: col, kind: KindMedian, alias: "median_" + col} }
func First(col string) Agg  { return Agg{col: col, kind: KindFirst, alias: "first_" + col} }
func Last(col string) Agg   { return Agg{col: col, kind: KindLast, alias: "last_" + col} }

func (a Agg) As(alias string) Agg {
	a.alias = alias
	return a
}

func (a Agg) Col() string   { return a.col }
func (a Agg) Alias() string { return a.alias }

func (a Agg) Apply(s *series.Series) (float64, error) {
	if a.kind == KindCount {
		return float64(stats.Count(s)), nil
	}
	if s.DType() != series.Float64 && s.DType() != series.Int64 {
		return math.NaN(), fmt.Errorf("godas: aggregation %q requires numeric column, got %s", a.alias, s.DType())
	}
	switch a.kind {
	case KindSum:
		return stats.Sum(s), nil
	case KindMean:
		return stats.Mean(s), nil
	case KindMin:
		return stats.Min(s), nil
	case KindMax:
		return stats.Max(s), nil
	case KindStd:
		return stats.Std(s), nil
	case KindMedian:
		return stats.Median(s), nil
	case KindFirst:
		for i := 0; i < s.Len(); i++ {
			v, ok := s.GetFloat64(i)
			if ok {
				return v, nil
			}
		}
		return math.NaN(), nil
	case KindLast:
		for i := s.Len() - 1; i >= 0; i-- {
			v, ok := s.GetFloat64(i)
			if ok {
				return v, nil
			}
		}
		return math.NaN(), nil
	}
	return math.NaN(), fmt.Errorf("godas: unknown aggregation kind %d", a.kind)
}