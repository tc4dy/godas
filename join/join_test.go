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
	resultCols := make([]*series.Series, 0, left.NCols()+right.NCols()-1)

	for _, name := range left.ColumnNames() {
		src := left.MustCol(name)
		resultCols = append(resultCols, buildJoinedSeries(src, leftRows, n, name))
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
		resultCols = append(resultCols, buildJoinedSeries(src, rightRows, n, outName))
	}

	return dataframe.New(resultCols...)
}

func buildJoinedSeries(src *series.Series, rows []int, n int, name string) *series.Series {
	switch src.DType() {
	case series.Float64:
		vals := make([]float64, n)
		nulls := make([]bool, n)
		raw := src.RawFloats()
		srcNulls := src.Nulls()
		for i, r := range rows {
			if r < 0 {
				nulls[i] = true
				vals[i] = 0
			} else {
				vals[i] = raw[r]
				nulls[i] = srcNulls[r]
			}
		}
		s := series.NewFloat64(name, vals)
		copy(s.Nulls(), nulls)
		return s

	case series.Int64:
		vals := make([]int64, n)
		nulls := make([]bool, n)
		raw := src.RawInts()
		srcNulls := src.Nulls()
		for i, r := range rows {
			if r < 0 {
				nulls[i] = true
				vals[i] = 0
			} else {
				vals[i] = raw[r]
				nulls[i] = srcNulls[r]
			}
		}
		s := series.NewInt64(name, vals)
		copy(s.Nulls(), nulls)
		return s

	case series.String:
		vals := make([]string, n)
		nulls := make([]bool, n)
		raw := src.RawStrings()
		srcNulls := src.Nulls()
		for i, r := range rows {
			if r < 0 {
				nulls[i] = true
				vals[i] = ""
			} else {
				vals[i] = raw[r]
				nulls[i] = srcNulls[r]
			}
		}
		s := series.NewString(name, vals)
		copy(s.Nulls(), nulls)
		return s

	case series.Bool:
		vals := make([]bool, n)
		nulls := make([]bool, n)
		raw := src.RawBools()
		srcNulls := src.Nulls()
		for i, r := range rows {
			if r < 0 {
				nulls[i] = true
				vals[i] = false
			} else {
				vals[i] = raw[r]
				nulls[i] = srcNulls[r]
			}
		}
		s := series.NewBool(name, vals)
		copy(s.Nulls(), nulls)
		return s

	default:
		vals := make([]string, n)
		nulls := make([]bool, n)
		srcNulls := src.Nulls()
		for i, r := range rows {
			if r < 0 {
				nulls[i] = true
				vals[i] = ""
			} else {
				vals[i] = src.StringAt(r)
				nulls[i] = srcNulls[r]
			}
		}
		s := series.NewString(name, vals)
		copy(s.Nulls(), nulls)
		return s
	}
}
