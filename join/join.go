package join

import (
	"fmt"

	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/series"
)

type joinType int

const (
	joinInner joinType = iota
	joinLeft
	joinRight
)

func Inner(left, right *dataframe.DataFrame, on string) (*dataframe.DataFrame, error) {
	return perform(left, right, on, joinInner)
}

func Left(left, right *dataframe.DataFrame, on string) (*dataframe.DataFrame, error) {
	return perform(left, right, on, joinLeft)
}

func Right(left, right *dataframe.DataFrame, on string) (*dataframe.DataFrame, error) {
	return perform(left, right, on, joinRight)
}

func perform(left, right *dataframe.DataFrame, on string, kind joinType) (*dataframe.DataFrame, error) {
	lk, err := left.Col(on)
	if err != nil {
		return nil, fmt.Errorf("godas join: left key %w", err)
	}
	rk, err := right.Col(on)
	if err != nil {
		return nil, fmt.Errorf("godas join: right key %w", err)
	}

	rIndex := make(map[string][]int, rk.Len())
	for i := 0; i < rk.Len(); i++ {
		k := rk.StringAt(i)
		rIndex[k] = append(rIndex[k], i)
	}

	leftRows := []int{}
	rightRows := []int{}
	rightSeen := make([]bool, rk.Len())

	for li := 0; li < lk.Len(); li++ {
		k := lk.StringAt(li)
		matches, found := rIndex[k]
		if found {
			for _, ri := range matches {
				leftRows = append(leftRows, li)
				rightRows = append(rightRows, ri)
				rightSeen[ri] = true
			}
		} else if kind == joinLeft {
			leftRows = append(leftRows, li)
			rightRows = append(rightRows, -1)
		}
	}

	if kind == joinRight {
		for ri := 0; ri < rk.Len(); ri++ {
			if !rightSeen[ri] {
				leftRows = append(leftRows, -1)
				rightRows = append(rightRows, ri)
			}
		}
	}

	n := len(leftRows)
	result := make([]*series.Series, 0)

	for _, name := range left.ColumnNames() {
		src := left.MustCol(name)
		result = append(result, buildJoinedSeries(src, leftRows, n, name))
	}

	for _, name := range right.ColumnNames() {
		if name == on {
			continue
		}
		src := right.MustCol(name)
		outName := name
		if left.HasCol(name) {
			outName = name + "_right"
		}
		result = append(result, buildJoinedSeries(src, rightRows, n, outName))
	}

	return dataframe.New(result...)
}

func buildJoinedSeries(src *series.Series, rows []int, n int, name string) *series.Series {
	switch src.DType() {
	case series.Float64:
		vals := make([]float64, n)
		raw := src.RawFloats()
		srcNulls := src.Nulls()
		nulls := make([]bool, n)
		for i, r := range rows {
			if r < 0 {
				nulls[i] = true
			} else {
				vals[i] = raw[r]
				nulls[i] = srcNulls[r]
			}
		}
		_ = nulls
		return series.NewFloat64(name, vals)
	case series.Int64:
		vals := make([]int64, n)
		raw := src.RawInts()
		for i, r := range rows {
			if r >= 0 {
				vals[i] = raw[r]
			}
		}
		return series.NewInt64(name, vals)
	case series.String:
		vals := make([]string, n)
		raw := src.RawStrings()
		for i, r := range rows {
			if r >= 0 {
				vals[i] = raw[r]
			}
		}
		return series.NewString(name, vals)
	case series.Bool:
		vals := make([]bool, n)
		raw := src.RawBools()
		for i, r := range rows {
			if r >= 0 {
				vals[i] = raw[r]
			}
		}
		return series.NewBool(name, vals)
	}
	return series.NewString(name, make([]string, n))
}