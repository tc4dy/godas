package pool

import "sync"

var float64Pool = sync.Pool{
	New: func() any {
		s := make([]float64, 0, 1024)
		return &s
	},
}

var stringPool = sync.Pool{
	New: func() any {
		s := make([]string, 0, 1024)
		return &s
	},
}

var boolPool = sync.Pool{
	New: func() any {
		s := make([]bool, 0, 1024)
		return &s
	},
}

func GetFloat64(cap int) []float64 {
	p := float64Pool.Get().(*[]float64)
	s := *p
	if cap > 0 && cap(s) < cap {
		s = make([]float64, 0, cap)
	}
	return s[:0]
}

func PutFloat64(s []float64) {
	p := &s
	float64Pool.Put(p)
}

func GetString(cap int) []string {
	p := stringPool.Get().(*[]string)
	s := *p
	if cap > 0 && cap(s) < cap {
		s = make([]string, 0, cap)
	}
	return s[:0]
}

func PutString(s []string) {
	p := &s
	stringPool.Put(p)
}

func GetBool(cap int) []bool {
	p := boolPool.Get().(*[]bool)
	s := *p
	if cap > 0 && cap(s) < cap {
		s = make([]bool, 0, cap)
	}
	return s[:0]
}

func PutBool(s []bool) {
	p := &s
	boolPool.Put(p)
}