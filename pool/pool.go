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

func GetFloat64(size int) []float64 {
	p := float64Pool.Get().(*[]float64)
	s := *p
	if cap(s) < size {
		s = make([]float64, 0, size)
	}
	s = s[:0]
	return s
}

func PutFloat64(s []float64) {
	if s == nil {
		return
	}
	s = s[:0]
	float64Pool.Put(&s)
}

func GetString(size int) []string {
	p := stringPool.Get().(*[]string)
	s := *p
	if cap(s) < size {
		s = make([]string, 0, size)
	}
	s = s[:0]
	return s
}

func PutString(s []string) {
	if s == nil {
		return
	}
	s = s[:0]
	stringPool.Put(&s)
}

func GetBool(size int) []bool {
	p := boolPool.Get().(*[]bool)
	s := *p
	if cap(s) < size {
		s = make([]bool, 0, size)
	}
	s = s[:0]
	return s
}

func PutBool(s []bool) {
	if s == nil {
		return
	}
	s = s[:0]
	boolPool.Put(&s)
}
