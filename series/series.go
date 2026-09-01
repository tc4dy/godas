package series

import (
	"fmt"
	"math"
	"sort"
	"strconv"
)

type DType int

const (
	Float64 DType = iota
	Int64
	String
	Bool
)

func (d DType) String() string {
	switch d {
	case Float64:
		return "float64"
	case Int64:
		return "int64"
	case String:
		return "string"
	case Bool:
		return "bool"
	default:
		return "unknown"
	}
}

type Series struct {
	name   string
	dtype  DType
	floats []float64
	ints   []int64
	strs   []string
	bools  []bool
	nulls  []bool
	len    int
}

func NewFloat64(name string, data []float64) *Series {
	nulls := make([]bool, len(data))
	for i, v := range data {
		if math.IsNaN(v) {
			nulls[i] = true
		}
	}
	return &Series{name: name, dtype: Float64, floats: data, nulls: nulls, len: len(data)}
}

func NewInt64(name string, data []int64) *Series {
	nulls := make([]bool, len(data))
	return &Series{name: name, dtype: Int64, ints: data, nulls: nulls, len: len(data)}
}

func NewString(name string, data []string) *Series {
	nulls := make([]bool, len(data))
	return &Series{name: name, dtype: String, strs: data, nulls: nulls, len: len(data)}
}

func NewBool(name string, data []bool) *Series {
	nulls := make([]bool, len(data))
	return &Series{name: name, dtype: Bool, bools: data, nulls: nulls, len: len(data)}
}

func (s *Series) Name() string    { return s.name }
func (s *Series) DType() DType    { return s.dtype }
func (s *Series) Len() int        { return s.len }
func (s *Series) IsNull(i int) bool { return s.nulls[i] }

func (s *Series) SetName(name string) *Series {
	s.name = name
	return s
}

func (s *Series) GetFloat64(i int) (float64, bool) {
	if s.nulls[i] {
		return 0, false
	}
	switch s.dtype {
	case Float64:
		return s.floats[i], true
	case Int64:
		return float64(s.ints[i]), true
	default:
		return 0, false
	}
}

func (s *Series) GetString(i int) (string, bool) {
	if s.nulls[i] {
		return "", false
	}
	switch s.dtype {
	case String:
		return s.strs[i], true
	case Float64:
		return strconv.FormatFloat(s.floats[i], 'f', -1, 64), true
	case Int64:
		return strconv.FormatInt(s.ints[i], 10), true
	case Bool:
		return strconv.FormatBool(s.bools[i]), true
	default:
		return "", false
	}
}

func (s *Series) GetBool(i int) (bool, bool) {
	if s.nulls[i] {
		return false, false
	}
	if s.dtype != Bool {
		return false, false
	}
	return s.bools[i], true
}

func (s *Series) RawFloats() []float64 { return s.floats }
func (s *Series) RawInts() []int64     { return s.ints }
func (s *Series) RawStrings() []string { return s.strs }
func (s *Series) RawBools() []bool     { return s.bools }
func (s *Series) Nulls() []bool        { return s.nulls }

func (s *Series) Filter(mask []bool) *Series {
	if len(mask) != s.len {
		panic(fmt.Sprintf("godas: mask length %d does not match series length %d", len(mask), s.len))
	}
	count := 0
	for _, b := range mask {
		if b {
			count++
		}
	}
	out := &Series{name: s.name, dtype: s.dtype, len: count, nulls: make([]bool, count)}
	idx := 0
	switch s.dtype {
	case Float64:
		out.floats = make([]float64, count)
		for i, b := range mask {
			if b {
				out.floats[idx] = s.floats[i]
				out.nulls[idx] = s.nulls[i]
				idx++
			}
		}
	case Int64:
		out.ints = make([]int64, count)
		for i, b := range mask {
			if b {
				out.ints[idx] = s.ints[i]
				out.nulls[idx] = s.nulls[i]
				idx++
			}
		}
	case String:
		out.strs = make([]string, count)
		for i, b := range mask {
			if b {
				out.strs[idx] = s.strs[i]
				out.nulls[idx] = s.nulls[i]
				idx++
			}
		}
	case Bool:
		out.bools = make([]bool, count)
		for i, b := range mask {
			if b {
				out.bools[idx] = s.bools[i]
				out.nulls[idx] = s.nulls[i]
				idx++
			}
		}
	}
	return out
}

func (s *Series) Slice(start, end int) *Series {
	if start < 0 || end > s.len || start > end {
		panic(fmt.Sprintf("godas: slice [%d:%d] out of bounds for series length %d", start, end, s.len))
	}
	count := end - start
	out := &Series{name: s.name, dtype: s.dtype, len: count, nulls: make([]bool, count)}
	copy(out.nulls, s.nulls[start:end])
	switch s.dtype {
	case Float64:
		out.floats = make([]float64, count)
		copy(out.floats, s.floats[start:end])
	case Int64:
		out.ints = make([]int64, count)
		copy(out.ints, s.ints[start:end])
	case String:
		out.strs = make([]string, count)
		copy(out.strs, s.strs[start:end])
	case Bool:
		out.bools = make([]bool, count)
		copy(out.bools, s.bools[start:end])
	}
	return out
}

func (s *Series) SortedIndices(ascending bool) []int {
	idx := make([]int, s.len)
	for i := range idx {
		idx[i] = i
	}
	sort.SliceStable(idx, func(a, b int) bool {
		ia, ib := idx[a], idx[b]
		if s.nulls[ia] {
			return false
		}
		if s.nulls[ib] {
			return true
		}
		switch s.dtype {
		case Float64:
			if ascending {
				return s.floats[ia] < s.floats[ib]
			}
			return s.floats[ia] > s.floats[ib]
		case Int64:
			if ascending {
				return s.ints[ia] < s.ints[ib]
			}
			return s.ints[ia] > s.ints[ib]
		case String:
			if ascending {
				return s.strs[ia] < s.strs[ib]
			}
			return s.strs[ia] > s.strs[ib]
		case Bool:
			if ascending {
				return !s.bools[ia] && s.bools[ib]
			}
			return s.bools[ia] && !s.bools[ib]
		}
		return false
	})
	return idx
}

func (s *Series) Reorder(indices []int) *Series {
	out := &Series{name: s.name, dtype: s.dtype, len: len(indices), nulls: make([]bool, len(indices))}
	switch s.dtype {
	case Float64:
		out.floats = make([]float64, len(indices))
		for i, idx := range indices {
			if idx < 0 || idx >= s.len {
				out.nulls[i] = true
				continue
			}
			out.floats[i] = s.floats[idx]
			out.nulls[i] = s.nulls[idx]
		}
	case Int64:
		out.ints = make([]int64, len(indices))
		for i, idx := range indices {
			if idx < 0 || idx >= s.len {
				out.nulls[i] = true
				continue
			}
			out.ints[i] = s.ints[idx]
			out.nulls[i] = s.nulls[idx]
		}
	case String:
		out.strs = make([]string, len(indices))
		for i, idx := range indices {
			if idx < 0 || idx >= s.len {
				out.nulls[i] = true
				continue
			}
			out.strs[i] = s.strs[idx]
			out.nulls[i] = s.nulls[idx]
		}
	case Bool:
		out.bools = make([]bool, len(indices))
		for i, idx := range indices {
			if idx < 0 || idx >= s.len {
				out.nulls[i] = true
				continue
			}
			out.bools[i] = s.bools[idx]
			out.nulls[i] = s.nulls[idx]
		}
	}
	return out
}

func (s *Series) Append(other *Series) (*Series, error) {
	if s.dtype != other.dtype {
		return nil, fmt.Errorf("godas: cannot append series of type %s to %s", other.dtype, s.dtype)
	}
	out := &Series{name: s.name, dtype: s.dtype, len: s.len + other.len}
	out.nulls = make([]bool, s.len+other.len)
	copy(out.nulls, s.nulls)
	copy(out.nulls[s.len:], other.nulls)
	switch s.dtype {
	case Float64:
		out.floats = make([]float64, s.len+other.len)
		copy(out.floats, s.floats)
		copy(out.floats[s.len:], other.floats)
	case Int64:
		out.ints = make([]int64, s.len+other.len)
		copy(out.ints, s.ints)
		copy(out.ints[s.len:], other.ints)
	case String:
		out.strs = make([]string, s.len+other.len)
		copy(out.strs, s.strs)
		copy(out.strs[s.len:], other.strs)
	case Bool:
		out.bools = make([]bool, s.len+other.len)
		copy(out.bools, s.bools)
		copy(out.bools[s.len:], other.bools)
	}
	return out, nil
}

func (s *Series) Clone() *Series {
	out := &Series{name: s.name, dtype: s.dtype, len: s.len}
	out.nulls = make([]bool, s.len)
	copy(out.nulls, s.nulls)
	switch s.dtype {
	case Float64:
		out.floats = make([]float64, s.len)
		copy(out.floats, s.floats)
	case Int64:
		out.ints = make([]int64, s.len)
		copy(out.ints, s.ints)
	case String:
		out.strs = make([]string, s.len)
		copy(out.strs, s.strs)
	case Bool:
		out.bools = make([]bool, s.len)
		copy(out.bools, s.bools)
	}
	return out
}

func (s *Series) ValueAt(i int) any {
	if s.nulls[i] {
		return nil
	}
	switch s.dtype {
	case Float64:
		return s.floats[i]
	case Int64:
		return s.ints[i]
	case String:
		return s.strs[i]
	case Bool:
		return s.bools[i]
	}
	return nil
}

func (s *Series) StringAt(i int) string {
	if s.nulls[i] {
		return "null"
	}
	switch s.dtype {
	case Float64:
		return strconv.FormatFloat(s.floats[i], 'f', 4, 64)
	case Int64:
		return strconv.FormatInt(s.ints[i], 10)
	case String:
		return s.strs[i]
	case Bool:
		return strconv.FormatBool(s.bools[i])
	default:
		return "null"
	}
}
