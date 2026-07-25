package expr

import (
	"fmt"
	"strings"

	"github.com/tc4dy/godas/series"
)

type Expr interface {
	Eval(cols map[string]*series.Series, n int) ([]bool, error)
	String() string
}

type ColExpr struct {
	name string
}

func Col(name string) *ColExpr {
	return &ColExpr{name: name}
}

func (c *ColExpr) String() string { return c.name }

func (c *ColExpr) Eval(cols map[string]*series.Series, n int) ([]bool, error) {
	s, ok := cols[c.name]
	if !ok {
		return nil, fmt.Errorf("godas: column %q not found", c.name)
	}
	if s.DType() != series.Bool {
		return nil, fmt.Errorf("godas: column %q is not bool type", c.name)
	}
	mask := make([]bool, n)
	for i := 0; i < n; i++ {
		v, ok := s.GetBool(i)
		if ok {
			mask[i] = v
		}
	}
	return mask, nil
}

type cmpOp int

const (
	opGt cmpOp = iota
	opGte
	opLt
	opLte
	opEq
	opNeq
)

type CmpExpr struct {
	col string
	op  cmpOp
	val any
}

func (c *ColExpr) Gt(val any) *CmpExpr  { return &CmpExpr{col: c.name, op: opGt, val: val} }
func (c *ColExpr) Gte(val any) *CmpExpr { return &CmpExpr{col: c.name, op: opGte, val: val} }
func (c *ColExpr) Lt(val any) *CmpExpr  { return &CmpExpr{col: c.name, op: opLt, val: val} }
func (c *ColExpr) Lte(val any) *CmpExpr { return &CmpExpr{col: c.name, op: opLte, val: val} }
func (c *ColExpr) Eq(val any) *CmpExpr  { return &CmpExpr{col: c.name, op: opEq, val: val} }
func (c *ColExpr) Neq(val any) *CmpExpr { return &CmpExpr{col: c.name, op: opNeq, val: val} }

func (c *CmpExpr) String() string {
	ops := map[cmpOp]string{opGt: ">", opGte: ">=", opLt: "<", opLte: "<=", opEq: "==", opNeq: "!="}
	return fmt.Sprintf("%s %s %v", c.col, ops[c.op], c.val)
}

func (c *CmpExpr) Eval(cols map[string]*series.Series, n int) ([]bool, error) {
	s, ok := cols[c.col]
	if !ok {
		return nil, fmt.Errorf("godas: column %q not found", c.col)
	}
	mask := make([]bool, n)
	switch v := c.val.(type) {
	case float64:
		if s.DType() != series.Float64 && s.DType() != series.Int64 {
			return nil, fmt.Errorf("godas: column %q is not numeric", c.col)
		}
		for i := 0; i < n; i++ {
			fv, ok := s.GetFloat64(i)
			if !ok {
				continue
			}
			mask[i] = evalCmp(fv, v, c.op)
		}
	case int:
		if s.DType() != series.Float64 && s.DType() != series.Int64 {
			return nil, fmt.Errorf("godas: column %q is not numeric", c.col)
		}
		fv2 := float64(v)
		for i := 0; i < n; i++ {
			fv, ok := s.GetFloat64(i)
			if !ok {
				continue
			}
			mask[i] = evalCmp(fv, fv2, c.op)
		}
	case string:
		if s.DType() != series.String {
			return nil, fmt.Errorf("godas: column %q is not string", c.col)
		}
		for i := 0; i < n; i++ {
			sv, ok := s.GetString(i)
			if !ok {
				continue
			}
			switch c.op {
			case opEq:
				mask[i] = sv == v
			case opNeq:
				mask[i] = sv != v
			default:
				return nil, fmt.Errorf("godas: operator not supported for string column")
			}
		}
	default:
		return nil, fmt.Errorf("godas: unsupported comparison value type %T", c.val)
	}
	return mask, nil
}

func evalCmp(a, b float64, op cmpOp) bool {
	switch op {
	case opGt:
		return a > b
	case opGte:
		return a >= b
	case opLt:
		return a < b
	case opLte:
		return a <= b
	case opEq:
		return a == b
	case opNeq:
		return a != b
	}
	return false
}

type AndExpr struct{ left, right Expr }
type OrExpr struct{ left, right Expr }
type NotExpr struct{ inner Expr }

func And(left, right Expr) *AndExpr { return &AndExpr{left: left, right: right} }
func Or(left, right Expr) *OrExpr   { return &OrExpr{left: left, right: right} }
func Not(inner Expr) *NotExpr       { return &NotExpr{inner: inner} }

func (a *AndExpr) String() string { return fmt.Sprintf("(%s AND %s)", a.left, a.right) }
func (a *OrExpr) String() string  { return fmt.Sprintf("(%s OR %s)", a.left, a.right) }
func (a *NotExpr) String() string { return fmt.Sprintf("NOT(%s)", a.inner) }

func (a *AndExpr) Eval(cols map[string]*series.Series, n int) ([]bool, error) {
	l, err := a.left.Eval(cols, n)
	if err != nil {
		return nil, err
	}
	r, err := a.right.Eval(cols, n)
	if err != nil {
		return nil, err
	}
	mask := make([]bool, n)
	for i := range mask {
		mask[i] = l[i] && r[i]
	}
	return mask, nil
}

func (a *OrExpr) Eval(cols map[string]*series.Series, n int) ([]bool, error) {
	l, err := a.left.Eval(cols, n)
	if err != nil {
		return nil, err
	}
	r, err := a.right.Eval(cols, n)
	if err != nil {
		return nil, err
	}
	mask := make([]bool, n)
	for i := range mask {
		mask[i] = l[i] || r[i]
	}
	return mask, nil
}

func (a *NotExpr) Eval(cols map[string]*series.Series, n int) ([]bool, error) {
	inner, err := a.inner.Eval(cols, n)
	if err != nil {
		return nil, err
	}
	mask := make([]bool, n)
	for i, v := range inner {
		mask[i] = !v
	}
	return mask, nil
}

type ContainsExpr struct {
	col    string
	substr string
}

func (c *ColExpr) Contains(substr string) *ContainsExpr {
	return &ContainsExpr{col: c.name, substr: substr}
}

func (c *ContainsExpr) String() string {
	return fmt.Sprintf("%s contains %q", c.col, c.substr)
}

func (c *ContainsExpr) Eval(cols map[string]*series.Series, n int) ([]bool, error) {
	s, ok := cols[c.col]
	if !ok {
		return nil, fmt.Errorf("godas: column %q not found", c.col)
	}
	if s.DType() != series.String {
		return nil, fmt.Errorf("godas: Contains requires string column, got %s", s.DType())
	}
	mask := make([]bool, n)
	for i := 0; i < n; i++ {
		sv, ok := s.GetString(i)
		if ok {
			mask[i] = strings.Contains(sv, c.substr)
		}
	}
	return mask, nil
}

type InExpr struct {
	col  string
	vals map[any]struct{}
}

func (c *ColExpr) In(vals ...any) *InExpr {
	set := make(map[any]struct{}, len(vals))
	for _, v := range vals {
		set[v] = struct{}{}
	}
	return &InExpr{col: c.name, vals: set}
}

func (e *InExpr) String() string { return fmt.Sprintf("%s in [...]", e.col) }

func (e *InExpr) Eval(cols map[string]*series.Series, n int) ([]bool, error) {
	s, ok := cols[e.col]
	if !ok {
		return nil, fmt.Errorf("godas: column %q not found", e.col)
	}
	mask := make([]bool, n)
	for i := 0; i < n; i++ {
		v := s.ValueAt(i)
		if v == nil {
			continue
		}
		_, found := e.vals[v]
		mask[i] = found
	}
	return mask, nil
}