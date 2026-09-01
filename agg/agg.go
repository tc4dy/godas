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
	col   string
	kind  Kind
	alias string
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

	if a.kind == KindFirst || a.kind == KindLast {
		return a.applyFirstLast(s)
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
	}
	return math.NaN(), fmt.Errorf("godas: unknown aggregation kind %d", a.kind)
}

func (a Agg) applyFirstLast(s *series.Series) (float64, error) {
	switch s.DType() {
	case series.Float64:
		raw := s.RawFloats()
		nulls := s.Nulls()
		if a.kind == KindFirst {
			for i := 0; i < len(raw); i++ {
				if !nulls[i] {
					return raw[i], nil
				}
			}
		} else {
			for i := len(raw) - 1; i >= 0; i-- {
				if !nulls[i] {
					return raw[i], nil
				}
			}
		}
		return math.NaN(), nil

	case series.Int64:
		raw := s.RawInts()
		nulls := s.Nulls()
		if a.kind == KindFirst {
			for i := 0; i < len(raw); i++ {
				if !nulls[i] {
					return float64(raw[i]), nil
				}
			}
		} else {
			for i := len(raw) - 1; i >= 0; i-- {
				if !nulls[i] {
					return float64(raw[i]), nil
				}
			}
		}
		return math.NaN(), nil

	case series.Bool:
		raw := s.RawBools()
		nulls := s.Nulls()
		if a.kind == KindFirst {
			for i := 0; i < len(raw); i++ {
				if !nulls[i] {
					if raw[i] {
						return 1.0, nil
					}
					return 0.0, nil
				}
			}
		} else {
			for i := len(raw) - 1; i >= 0; i-- {
				if !nulls[i] {
					if raw[i] {
						return 1.0, nil
					}
					return 0.0, nil
				}
			}
		}
		return math.NaN(), nil

	case series.String:
		return math.NaN(), fmt.Errorf("godas: First/Last not supported for string column %q", s.Name())

	default:
		return math.NaN(), fmt.Errorf("godas: unsupported type %s for First/Last", s.DType())
	}
}
