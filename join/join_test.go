package join_test

import (
	"testing"

	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/join"
	"github.com/tc4dy/godas/series"
)

func TestInnerJoin(t *testing.T) {
	left := makeDF([]string{"id", "name"}, [][]any{{1, "Alice"}, {2, "Bob"}, {3, "Charlie"}})
	right := makeDF([]string{"id", "score"}, [][]any{{1, 95}, {2, 88}, {4, 70}})
	result, err := join.Inner(left, right, "id")
	if err != nil {
		t.Fatal(err)
	}
	if result.NRows() != 2 {
		t.Fatalf("expected 2 rows, got %d", result.NRows())
	}
}

func TestLeftJoin(t *testing.T) {
	left := makeDF([]string{"id", "name"}, [][]any{{1, "Alice"}, {2, "Bob"}, {3, "Charlie"}})
	right := makeDF([]string{"id", "score"}, [][]any{{1, 95}, {2, 88}, {4, 70}})
	result, err := join.Left(left, right, "id")
	if err != nil {
		t.Fatal(err)
	}
	if result.NRows() != 3 {
		t.Fatalf("expected 3 rows, got %d", result.NRows())
	}
}

func TestRightJoin(t *testing.T) {
	left := makeDF([]string{"id", "name"}, [][]any{{1, "Alice"}, {2, "Bob"}, {3, "Charlie"}})
	right := makeDF([]string{"id", "score"}, [][]any{{1, 95}, {2, 88}, {4, 70}})
	result, err := join.Right(left, right, "id")
	if err != nil {
		t.Fatal(err)
	}
	if result.NRows() != 3 {
		t.Fatalf("expected 3 rows, got %d", result.NRows())
	}
}

func TestJoinMissingKey(t *testing.T) {
	left := makeDF([]string{"id"}, [][]any{{1}})
	right := makeDF([]string{"x"}, [][]any{{1}})
	_, err := join.Inner(left, right, "id")
	if err == nil {
		t.Fatal("expected error for missing key")
	}
}

func TestJoinEmptyLeft(t *testing.T) {
	left, _ := dataframe.New()
	right := makeDF([]string{"id"}, [][]any{{1}})
	_, err := join.Inner(left, right, "id")
	if err == nil {
		t.Fatal("expected error for empty left")
	}
}

func makeDF(cols []string, rows [][]any) *dataframe.DataFrame {
	seriesList := make([]*series.Series, len(cols))
	for i, col := range cols {
		vals := make([]any, len(rows))
		for j, row := range rows {
			vals[j] = row[i]
		}
		switch vals[0].(type) {
		case int:
			ints := make([]int64, len(vals))
			for k, v := range vals {
				ints[k] = int64(v.(int))
			}
			seriesList[i] = series.NewInt64(col, ints)
		case float64:
			floats := make([]float64, len(vals))
			for k, v := range vals {
				floats[k] = v.(float64)
			}
			seriesList[i] = series.NewFloat64(col, floats)
		case string:
			strs := make([]string, len(vals))
			for k, v := range vals {
				strs[k] = v.(string)
			}
			seriesList[i] = series.NewString(col, strs)
		}
	}
	df, _ := dataframe.New(seriesList...)
	return df
}