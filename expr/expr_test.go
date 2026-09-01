package expr_test

import (
	"testing"

	"github.com/tc4dy/godas/expr"
	"github.com/tc4dy/godas/series"
)

func makeColMap() map[string]*series.Series {
	return map[string]*series.Series{
		"age":    series.NewFloat64("age", []float64{15, 25, 35, 45}),
		"city":   series.NewString("city", []string{"NY", "LA", "NY", "SF"}),
		"active": series.NewBool("active", []bool{true, false, true, true}),
	}
}

func TestGt(t *testing.T) {
	mask, err := expr.Col("age").Gt(20.0).Eval(makeColMap(), 4)  
	...
}

func TestLt(t *testing.T) {
	mask, err := expr.Col("age").Lt(30.0).Eval(makeColMap(), 4)  
	...
}

func TestEqString(t *testing.T) {
	mask, err := expr.Col("city").Eq("NY").Eval(makeColMap(), 4)
	if err != nil {
		t.Fatal(err)
	}
	if !mask[0] || mask[1] || !mask[2] || mask[3] {
		t.Fatalf("unexpected mask: %v", mask)
	}
}

func TestNeqString(t *testing.T) {
	mask, err := expr.Col("city").Neq("NY").Eval(makeColMap(), 4)
	if err != nil {
		t.Fatal(err)
	}
	if mask[0] || !mask[1] || mask[2] || !mask[3] {
		t.Fatalf("unexpected mask: %v", mask)
	}
}

func TestAnd(t *testing.T) {
	e := expr.And(expr.Col("age").Gt(20), expr.Col("city").Eq("NY"))
	mask, err := e.Eval(makeColMap(), 4)
	if err != nil {
		t.Fatal(err)
	}
	if mask[0] || mask[1] || !mask[2] || mask[3] {
		t.Fatalf("unexpected mask: %v", mask)
	}
}

func TestOr(t *testing.T) {
	e := expr.Or(expr.Col("age").Lt(20), expr.Col("city").Eq("SF"))
	mask, err := e.Eval(makeColMap(), 4)
	if err != nil {
		t.Fatal(err)
	}
	if !mask[0] || mask[1] || mask[2] || !mask[3] {
		t.Fatalf("unexpected mask: %v", mask)
	}
}

func TestNot(t *testing.T) {
	e := expr.Not(expr.Col("age").Gt(30))
	mask, err := e.Eval(makeColMap(), 4)
	if err != nil {
		t.Fatal(err)
	}
	if !mask[0] || !mask[1] || mask[2] || mask[3] {
		t.Fatalf("unexpected mask: %v", mask)
	}
}

func TestContains(t *testing.T) {
	mask, err := expr.Col("city").Contains("Y").Eval(makeColMap(), 4)
	if err != nil {
		t.Fatal(err)
	}
	if !mask[0] || mask[1] || !mask[2] || mask[3] {
		t.Fatalf("unexpected mask: %v", mask)
	}
}

func TestIn(t *testing.T) {
	mask, err := expr.Col("city").In("NY", "SF").Eval(makeColMap(), 4)
	if err != nil {
		t.Fatal(err)
	}
	if !mask[0] || mask[1] || !mask[2] || !mask[3] {
		t.Fatalf("unexpected mask: %v", mask)
	}
}

func TestMissingColumn(t *testing.T) {
	_, err := expr.Col("nonexistent").Gt(0).Eval(makeColMap(), 4)
	if err == nil {
		t.Fatal("expected error for missing column")
	}
}

func TestContainsOnNonString(t *testing.T) {
	_, err := expr.Col("age").Contains("5").Eval(makeColMap(), 4)
	if err == nil {
		t.Fatal("expected error for Contains on non-string column")
	}
}
