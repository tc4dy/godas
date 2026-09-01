package dataframe

import (
	"fmt"
	"math"
	"os"
	"strings"
	"sync"
	"text/tabwriter"

	"github.com/tc4dy/godas/agg"
	"github.com/tc4dy/godas/expr"
	"github.com/tc4dy/godas/series"
	"github.com/tc4dy/godas/stats"
)

type DataFrame struct {
	cols    []*series.Series
	colIdx  map[string]int
	nrows   int
}

func New(cols ...*series.Series) (*DataFrame, error) {
	if len(cols) == 0 {
		return &DataFrame{colIdx: make(map[string]int)}, nil
	}
	n := cols[0].Len()
	for _, c := range cols[1:] {
		if c.Len() != n {
			return nil, fmt.Errorf("godas: column length mismatch: %d vs %d", n, c.Len())
		}
	}
	idx := make(map[string]int, len(cols))
	for i, c := range cols {
		if _, dup := idx[c.Name()]; dup {
			return nil, fmt.Errorf("godas: duplicate column name %q", c.Name())
		}
		idx[c.Name()] = i
	}
	return &DataFrame{cols: cols, colIdx: idx, nrows: n}, nil
}

func (df *DataFrame) NRows() int { return df.nrows }
func (df *DataFrame) NCols() int { return len(df.cols) }

func (df *DataFrame) ColumnNames() []string {
	names := make([]string, len(df.cols))
	for i, c := range df.cols {
		names[i] = c.Name()
	}
	return names
}

func (df *DataFrame) Col(name string) (*series.Series, error) {
	i, ok := df.colIdx[name]
	if !ok {
		return nil, fmt.Errorf("godas: column %q not found", name)
	}
	return df.cols[i], nil
}

func (df *DataFrame) MustCol(name string) *series.Series {
	s, err := df.Col(name)
	if err != nil {
		panic(err)
	}
	return s
}

func (df *DataFrame) HasCol(name string) bool {
	_, ok := df.colIdx[name]
	return ok
}

func (df *DataFrame) colMap() map[string]*series.Series {
	m := make(map[string]*series.Series, len(df.cols))
	for _, c := range df.cols {
		m[c.Name()] = c
	}
	return m
}

func (df *DataFrame) Filter(e expr.Expr) (*DataFrame, error) {
	mask, err := e.Eval(df.colMap(), df.nrows)
	if err != nil {
		return nil, err
	}
	newCols := make([]*series.Series, len(df.cols))
	for i, c := range df.cols {
		newCols[i] = c.Filter(mask)
	}
	return New(newCols...)
}

func (df *DataFrame) Select(names ...string) (*DataFrame, error) {
	cols := make([]*series.Series, 0, len(names))
	for _, n := range names {
		s, err := df.Col(n)
		if err != nil {
			return nil, err
		}
		cols = append(cols, s)
	}
	return New(cols...)
}

func (df *DataFrame) Drop(names ...string) (*DataFrame, error) {
	drop := make(map[string]struct{}, len(names))
	for _, n := range names {
		drop[n] = struct{}{}
	}
	cols := make([]*series.Series, 0, len(df.cols))
	for _, c := range df.cols {
		if _, ok := drop[c.Name()]; !ok {
			cols = append(cols, c)
		}
	}
	if len(cols) == len(df.cols) {
		return nil, fmt.Errorf("godas: none of the specified columns found")
	}
	return New(cols...)
}

func (df *DataFrame) WithColumn(s *series.Series) (*DataFrame, error) {
	if s.Len() != df.nrows && df.nrows > 0 {
		return nil, fmt.Errorf("godas: column length %d does not match dataframe rows %d", s.Len(), df.nrows)
	}
	newCols := make([]*series.Series, len(df.cols))
	copy(newCols, df.cols)
	if i, ok := df.colIdx[s.Name()]; ok {
		newCols[i] = s
	} else {
		newCols = append(newCols, s)
	}
	return New(newCols...)
}

func (df *DataFrame) Rename(oldName, newName string) (*DataFrame, error) {
	i, ok := df.colIdx[oldName]
	if !ok {
		return nil, fmt.Errorf("godas: column %q not found", oldName)
	}
	if _, exists := df.colIdx[newName]; exists {
		return nil, fmt.Errorf("godas: column %q already exists", newName)
	}
	newCols := make([]*series.Series, len(df.cols))
	copy(newCols, df.cols)
	newCols[i] = df.cols[i].Clone().SetName(newName)
	return New(newCols...)
}

func (df *DataFrame) Sort(col string, ascending bool) (*DataFrame, error) {
	s, err := df.Col(col)
	if err != nil {
		return nil, err
	}
	indices := s.SortedIndices(ascending)
	newCols := make([]*series.Series, len(df.cols))
	var wg sync.WaitGroup
	errs := make([]error, len(df.cols))
	for i, c := range df.cols {
		wg.Add(1)
		go func(i int, c *series.Series) {
			defer wg.Done()
			newCols[i] = c.Reorder(indices)
		}(i, c)
	}
	wg.Wait()
	for _, e := range errs {
		if e != nil {
			return nil, e
		}
	}
	return New(newCols...)
}

func (df *DataFrame) Head(n int) *DataFrame {
	if n >= df.nrows {
		return df
	}
	newCols := make([]*series.Series, len(df.cols))
	for i, c := range df.cols {
		newCols[i] = c.Slice(0, n)
	}
	out, _ := New(newCols...)
	return out
}

func (df *DataFrame) Tail(n int) *DataFrame {
	if n >= df.nrows {
		return df
	}
	start := df.nrows - n
	newCols := make([]*series.Series, len(df.cols))
	for i, c := range df.cols {
		newCols[i] = c.Slice(start, df.nrows)
	}
	out, _ := New(newCols...)
	return out
}

func (df *DataFrame) Slice(start, end int) (*DataFrame, error) {
	if start < 0 || end > df.nrows || start > end {
		return nil, fmt.Errorf("godas: slice [%d:%d] out of bounds for %d rows", start, end, df.nrows)
	}
	newCols := make([]*series.Series, len(df.cols))
	for i, c := range df.cols {
		newCols[i] = c.Slice(start, end)
	}
	return New(newCols...)
}

type GroupedDF struct {
	df   *DataFrame
	keys []string
}

func (df *DataFrame) GroupBy(keys ...string) *GroupedDF {
	return &GroupedDF{df: df, keys: keys}
}

func (g *GroupedDF) Aggregate(aggs ...agg.Agg) (*DataFrame, error) {
    for _, k := range g.keys {
        if !g.df.HasCol(k) {
            return nil, fmt.Errorf("godas: group-by key %q not found", k)
        }
    }
    for _, a := range aggs {
        if !g.df.HasCol(a.Col()) {
            return nil, fmt.Errorf("godas: aggregation column %q not found", a.Col())
        }
    }

    type groupKey = string
    groupOrder := []groupKey{}
    groupRows := map[groupKey][]int{}

    for row := 0; row < g.df.nrows; row++ {
        var kb strings.Builder
        for ki, k := range g.keys {
            s := g.df.MustCol(k)
            kb.WriteString(s.StringAt(row))
            if ki < len(g.keys)-1 {
                kb.WriteByte('\x00')
            }
        }
        key := kb.String()
        if _, seen := groupRows[key]; !seen {
            groupOrder = append(groupOrder, key)
        }
        groupRows[key] = append(groupRows[key], row)
    }

    nGroups := len(groupOrder)
    keySeries := make([]*series.Series, len(g.keys))
    for ki, k := range g.keys {
        src := g.df.MustCol(k)
        switch src.DType() {
        case series.String:
            vals := make([]string, nGroups)
            for gi, key := range groupOrder {
                rows := groupRows[key]
                sv, _ := src.GetString(rows[0])
                vals[gi] = sv
            }
            keySeries[ki] = series.NewString(k, vals)
        case series.Float64:
            vals := make([]float64, nGroups)
            for gi, key := range groupOrder {
                rows := groupRows[key]
                fv, _ := src.GetFloat64(rows[0])
                vals[gi] = fv
            }
            keySeries[ki] = series.NewFloat64(k, vals)
        case series.Int64:
            vals := make([]int64, nGroups)
            for gi, key := range groupOrder {
                rows := groupRows[key]
                s2 := g.df.MustCol(k)
                iv := s2.RawInts()[rows[0]]
                vals[gi] = iv
            }
            keySeries[ki] = series.NewInt64(k, vals)
        default:
            vals := make([]string, nGroups)
            for gi, key := range groupOrder {
                rows := groupRows[key]
                vals[gi] = src.StringAt(rows[0])
            }
            keySeries[ki] = series.NewString(k, vals)
        }
    }

    aggSeries := make([]*series.Series, len(aggs))
    for ai, a := range aggs {
        src := g.df.MustCol(a.Col())
        vals := make([]float64, nGroups)
        var mu sync.Mutex
        var wg sync.WaitGroup
        var firstErr error 
        errOccurred := false
        
        for gi, key := range groupOrder {
            wg.Add(1)
            go func(gi int, key string) {
                defer wg.Done()
                rows := groupRows[key]
                sub := subSeries(src, rows)
                v, err := a.Apply(sub)
                mu.Lock()
                defer mu.Unlock()
                if err != nil && !errOccurred {
                    firstErr = err
                    errOccurred = true
                    vals[gi] = math.NaN()
                } else if err != nil {
                    vals[gi] = math.NaN()
                } else {
                    vals[gi] = v
                }
            }(gi, key)
        }
        wg.Wait()
        
        if errOccurred {
            return nil, firstErr
        }
        
        aggSeries[ai] = series.NewFloat64(a.Alias(), vals)
    }

    allCols := make([]*series.Series, 0, len(keySeries)+len(aggSeries))
    allCols = append(allCols, keySeries...)
    allCols = append(allCols, aggSeries...)
    return New(allCols...)
}

func subSeries(src *series.Series, rows []int) *series.Series {
    return src.Reorder(rows)
}

func (df *DataFrame) Describe() map[string]*stats.Summary {
	result := make(map[string]*stats.Summary)
	for _, c := range df.cols {
		if c.DType() == series.Float64 || c.DType() == series.Int64 {
			result[c.Name()] = stats.Describe(c)
		}
	}
	return result
}

func (df *DataFrame) Print() {
	df.Fprint(os.Stdout)
}

func (df *DataFrame) Fprint(w *os.File) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	header := strings.Join(df.ColumnNames(), "\t")
	fmt.Fprintln(tw, header)
	sep := make([]string, len(df.cols))
	for i, c := range df.cols {
		sep[i] = strings.Repeat("-", len(c.Name()))
	}
	fmt.Fprintln(tw, strings.Join(sep, "\t"))
	maxRows := df.nrows
	if maxRows > 50 {
		maxRows = 50
	}
	for row := 0; row < maxRows; row++ {
		parts := make([]string, len(df.cols))
		for ci, c := range df.cols {
			parts[ci] = c.StringAt(row)
		}
		fmt.Fprintln(tw, strings.Join(parts, "\t"))
	}
	if df.nrows > 50 {
		fmt.Fprintf(tw, "... %d more rows\n", df.nrows-50)
	}
	tw.Flush()
	fmt.Fprintf(w, "\n[%d rows x %d cols]\n", df.nrows, len(df.cols))
}

func (df *DataFrame) Append(other *DataFrame) (*DataFrame, error) {
	if len(df.cols) != len(other.cols) {
		return nil, fmt.Errorf("godas: column count mismatch: %d vs %d", len(df.cols), len(other.cols))
	}
	newCols := make([]*series.Series, len(df.cols))
	for i, c := range df.cols {
		oc, err := other.Col(c.Name())
		if err != nil {
			return nil, err
		}
		merged, err := c.Append(oc)
		if err != nil {
			return nil, err
		}
		newCols[i] = merged
	}
	return New(newCols...)
}

func (df *DataFrame) Clone() *DataFrame {
	newCols := make([]*series.Series, len(df.cols))
	for i, c := range df.cols {
		newCols[i] = c.Clone()
	}
	out, _ := New(newCols...)
	return out
}

func (df *DataFrame) Shape() (int, int) {
	return df.nrows, len(df.cols)
}

func (df *DataFrame) DTypes() map[string]series.DType {
	m := make(map[string]series.DType, len(df.cols))
	for _, c := range df.cols {
		m[c.Name()] = c.DType()
	}
	return m
}
