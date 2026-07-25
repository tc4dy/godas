package dataframe_test

import (
	"testing"

	"github.com/tc4dy/godas/agg"
	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/expr"
	"github.com/tc4dy/godas/series"
)

func TestNewDataFrame(t *testing.T) {
	s1 := series.NewFloat64("a", []float64{1, 2, 3})
	s2 := series.NewString("b", []string{"x", "y", "z"})
	df, err := dataframe.New(s1, s2)
	if err != nil {
		t.Fatal(err)
	}
	if df.NRows() != 3 {
		t.Fatalf("expected 3 rows, got %d", df.NRows())
	}
	if df.NCols() != 2 {
		t.Fatalf("expected 2 cols, got %d", df.NCols())
	}
}

func TestNewDataFrameLengthMismatch(t *testing.T) {
	s1 := series.NewFloat64("a", []float64{1, 2})
	s2 := series.NewString("b", []string{"x"})
	_, err := dataframe.New(s1, s2)
	if err == nil {
		t.Fatal("expected error for length mismatch")
	}
}

func TestDataFrameCol(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2})
	df, _ := dataframe.New(s)
	col, err := df.Col("x")
	if err != nil {
		t.Fatal(err)
	}
	if col.Len() != 2 {
		t.Fatalf("expected len 2, got %d", col.Len())
	}
}

func TestDataFrameColNotFound(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2})
	df, _ := dataframe.New(s)
	_, err := df.Col("y")
	if err == nil {
		t.Fatal("expected error for missing column")
	}
}

func TestDataFrameFilter(t *testing.T) {
	age := series.NewFloat64("age", []float64{10, 20, 30})
	df, _ := dataframe.New(age)
	filtered, err := df.Filter(expr.Col("age").Gt(15))
	if err != nil {
		t.Fatal(err)
	}
	if filtered.NRows() != 2 {
		t.Fatalf("expected 2 rows, got %d", filtered.NRows())
	}
}

func TestDataFrameSelect(t *testing.T) {
	s1 := series.NewFloat64("a", []float64{1, 2})
	s2 := series.NewString("b", []string{"x", "y"})
	df, _ := dataframe.New(s1, s2)
	sel, err := df.Select("b")
	if err != nil {
		t.Fatal(err)
	}
	if sel.NCols() != 1 {
		t.Fatalf("expected 1 col, got %d", sel.NCols())
	}
}

func TestDataFrameDrop(t *testing.T) {
	s1 := series.NewFloat64("a", []float64{1, 2})
	s2 := series.NewString("b", []string{"x", "y"})
	df, _ := dataframe.New(s1, s2)
	dropped, err := df.Drop("a")
	if err != nil {
		t.Fatal(err)
	}
	if dropped.NCols() != 1 {
		t.Fatalf("expected 1 col, got %d", dropped.NCols())
	}
}

func TestDataFrameWithColumn(t *testing.T) {
	s1 := series.NewFloat64("a", []float64{1, 2})
	df, _ := dataframe.New(s1)
	s2 := series.NewString("b", []string{"x", "y"})
	df2, err := df.WithColumn(s2)
	if err != nil {
		t.Fatal(err)
	}
	if df2.NCols() != 2 {
		t.Fatalf("expected 2 cols, got %d", df2.NCols())
	}
}

func TestDataFrameRename(t *testing.T) {
	s := series.NewFloat64("old", []float64{1, 2})
	df, _ := dataframe.New(s)
	renamed, err := df.Rename("old", "new")
	if err != nil {
		t.Fatal(err)
	}
	if !renamed.HasCol("new") {
		t.Fatal("column new not found")
	}
}

func TestDataFrameSort(t *testing.T) {
	s := series.NewFloat64("x", []float64{3, 1, 2})
	df, _ := dataframe.New(s)
	sorted, err := df.Sort("x", true)
	if err != nil {
		t.Fatal(err)
	}
	col, _ := sorted.Col("x")
	if col.RawFloats()[0] != 1 {
		t.Fatalf("expected first 1, got %v", col.RawFloats()[0])
	}
}

func TestDataFrameHead(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2, 3, 4, 5})
	df, _ := dataframe.New(s)
	head := df.Head(2)
	if head.NRows() != 2 {
		t.Fatalf("expected 2 rows, got %d", head.NRows())
	}
}

func TestDataFrameTail(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2, 3, 4, 5})
	df, _ := dataframe.New(s)
	tail := df.Tail(2)
	if tail.NRows() != 2 {
		t.Fatalf("expected 2 rows, got %d", tail.NRows())
	}
	col, _ := tail.Col("x")
	if col.RawFloats()[0] != 4 {
		t.Fatalf("expected first 4, got %v", col.RawFloats()[0])
	}
}

func TestDataFrameSlice(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2, 3, 4, 5})
	df, _ := dataframe.New(s)
	sl, err := df.Slice(1, 4)
	if err != nil {
		t.Fatal(err)
	}
	if sl.NRows() != 3 {
		t.Fatalf("expected 3 rows, got %d", sl.NRows())
	}
}

func TestDataFrameGroupByAggregate(t *testing.T) {
	city := series.NewString("city", []string{"A", "B", "A", "B"})
	val := series.NewFloat64("val", []float64{10, 20, 30, 40})
	df, _ := dataframe.New(city, val)
	grouped := df.GroupBy("city")
	result, err := grouped.Aggregate(agg.Mean("val"), agg.Sum("val"))
	if err != nil {
		t.Fatal(err)
	}
	if result.NRows() != 2 {
		t.Fatalf("expected 2 groups, got %d", result.NRows())
	}
}

func TestDataFrameDescribe(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2, 3})
	df, _ := dataframe.New(s)
	desc := df.Describe()
	if len(desc) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(desc))
	}
	if desc["x"].Mean != 2.0 {
		t.Fatalf("expected mean 2, got %v", desc["x"].Mean)
	}
}

func TestDataFrameAppend(t *testing.T) {
	s1 := series.NewFloat64("x", []float64{1, 2})
	s2 := series.NewFloat64("x", []float64{3, 4})
	df1, _ := dataframe.New(s1)
	df2, _ := dataframe.New(s2)
	app, err := df1.Append(df2)
	if err != nil {
		t.Fatal(err)
	}
	if app.NRows() != 4 {
		t.Fatalf("expected 4 rows, got %d", app.NRows())
	}
}

func TestDataFrameClone(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2, 3})
	df, _ := dataframe.New(s)
	clone := df.Clone()
	cloneCol, _ := clone.Col("x")
	cloneCol.RawFloats()[0] = 99
	origCol, _ := df.Col("x")
	if origCol.RawFloats()[0] == 99 {
		t.Fatal("clone modified original")
	}
}

func TestDataFrameShape(t *testing.T) {
	s := series.NewFloat64("x", []float64{1, 2})
	df, _ := dataframe.New(s)
	r, c := df.Shape()
	if r != 2 || c != 1 {
		t.Fatalf("expected (2,1), got (%d,%d)", r, c)
	}
}

func TestDataFrameDTypes(t *testing.T) {
	s1 := series.NewFloat64("a", []float64{1})
	s2 := series.NewString("b", []string{"x"})
	df, _ := dataframe.New(s1, s2)
	dtypes := df.DTypes()
	if dtypes["a"] != series.Float64 || dtypes["b"] != series.String {
		t.Fatal("dtype mismatch")
	}
}