package stats

import (
	"math"
	"sort"

	"github.com/tc4dy/godas/series"
)

type Summary struct {
	Count  int
	Mean   float64
	Std    float64
	Min    float64
	Q25    float64
	Median float64
	Q75    float64
	Max    float64
	NNull  int
}

func Describe(s *series.Series) *Summary {
	if s.DType() != series.Float64 && s.DType() != series.Int64 {
		return nil
	}
	vals := make([]float64, 0, s.Len())
	nulls := 0
	for i := 0; i < s.Len(); i++ {
		v, ok := s.GetFloat64(i)
		if !ok {
			nulls++
			continue
		}
		vals = append(vals, v)
	}
	if len(vals) == 0 {
		return &Summary{NNull: nulls}
	}
	sort.Float64s(vals)
	sum := Sum(s)
	n := float64(len(vals))
	mean := sum / n
	variance := 0.0
	for _, v := range vals {
		d := v - mean
		variance += d * d
	}
	std := 0.0
	if n > 1 {
		std = math.Sqrt(variance / (n - 1))
	}
	return &Summary{
		Count:  len(vals),
		Mean:   mean,
		Std:    std,
		Min:    vals[0],
		Q25:    percentile(vals, 0.25),
		Median: percentile(vals, 0.50),
		Q75:    percentile(vals, 0.75),
		Max:    vals[len(vals)-1],
		NNull:  nulls,
	}
}

func percentile(sorted []float64, p float64) float64 {
	if len(sorted) == 0 {
		return math.NaN()
	}
	idx := p * float64(len(sorted)-1)
	lo := int(math.Floor(idx))
	hi := int(math.Ceil(idx))
	if lo == hi {
		return sorted[lo]
	}
	return sorted[lo]*(float64(hi)-idx) + sorted[hi]*(idx-float64(lo))
}

func Sum(s *series.Series) float64 {
	total := 0.0
	for i := 0; i < s.Len(); i++ {
		v, ok := s.GetFloat64(i)
		if ok {
			total += v
		}
	}
	return total
}

func Mean(s *series.Series) float64 {
	total := 0.0
	count := 0
	for i := 0; i < s.Len(); i++ {
		v, ok := s.GetFloat64(i)
		if ok {
			total += v
			count++
		}
	}
	if count == 0 {
		return math.NaN()
	}
	return total / float64(count)
}

func Min(s *series.Series) float64 {
	min := math.Inf(1)
	found := false
	for i := 0; i < s.Len(); i++ {
		v, ok := s.GetFloat64(i)
		if ok {
			if v < min {
				min = v
			}
			found = true
		}
	}
	if !found {
		return math.NaN()
	}
	return min
}

func Max(s *series.Series) float64 {
	max := math.Inf(-1)
	found := false
	for i := 0; i < s.Len(); i++ {
		v, ok := s.GetFloat64(i)
		if ok {
			if v > max {
				max = v
			}
			found = true
		}
	}
	if !found {
		return math.NaN()
	}
	return max
}

func Std(s *series.Series) float64 {
	vals := make([]float64, 0, s.Len())
	for i := 0; i < s.Len(); i++ {
		v, ok := s.GetFloat64(i)
		if ok {
			vals = append(vals, v)
		}
	}
	if len(vals) < 2 {
		return math.NaN()
	}
	m := 0.0
	for _, v := range vals {
		m += v
	}
	m /= float64(len(vals))
	variance := 0.0
	for _, v := range vals {
		d := v - m
		variance += d * d
	}
	return math.Sqrt(variance / float64(len(vals)-1))
}

func Median(s *series.Series) float64 {
	vals := make([]float64, 0, s.Len())
	for i := 0; i < s.Len(); i++ {
		v, ok := s.GetFloat64(i)
		if ok {
			vals = append(vals, v)
		}
	}
	if len(vals) == 0 {
		return math.NaN()
	}
	sort.Float64s(vals)
	return percentile(vals, 0.5)
}

func Count(s *series.Series) int {
	n := 0
	for i := 0; i < s.Len(); i++ {
		if !s.IsNull(i) {
			n++
		}
	}
	return n
}

func NullCount(s *series.Series) int {
	n := 0
	for i := 0; i < s.Len(); i++ {
		if s.IsNull(i) {
			n++
		}
	}
	return n
}

func Corr(a, b *series.Series) float64 {
	if a.Len() != b.Len() {
		return math.NaN()
	}
	n := a.Len()
	sumA, sumB, sumAB, sumA2, sumB2 := 0.0, 0.0, 0.0, 0.0, 0.0
	count := 0
	for i := 0; i < n; i++ {
		av, aok := a.GetFloat64(i)
		bv, bok := b.GetFloat64(i)
		if !aok || !bok {
			continue
		}
		sumA += av
		sumB += bv
		sumAB += av * bv
		sumA2 += av * av
		sumB2 += bv * bv
		count++
	}
	if count == 0 {
		return math.NaN()
	}
	fn := float64(count)
	num := fn*sumAB - sumA*sumB
	den := math.Sqrt((fn*sumA2 - sumA*sumA) * (fn*sumB2 - sumB*sumB))
	if den == 0 {
		return math.NaN()
	}
	return num / den
}