package io_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tc4dy/godas/io"
)

func BenchmarkReadCSV(b *testing.B) {
	tmp := b.TempDir()
	path := filepath.Join(tmp, "bench.csv")
	f, _ := os.Create(path)
	defer f.Close()
	for i := 0; i < 1000000; i++ {
		f.WriteString("col1,col2\n")
	}
	f.Sync()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		io.ReadCSV(path)
	}
}