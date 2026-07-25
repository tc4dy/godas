package stream_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tc4dy/godas/stream"
)

func makeBenchCSV(b *testing.B, rows int) string {
	b.Helper()
	tmp := b.TempDir()
	path := filepath.Join(tmp, "bench.csv")
	f, err := os.Create(path)
	if err != nil {
		b.Fatal(err)
	}
	defer f.Close()
	var sb strings.Builder
	sb.WriteString("id,value\n")
	for i := 0; i < rows; i++ {
		sb.WriteString(fmt.Sprintf("%d,%d\n", i, i*10))
	}
	if _, err := f.WriteString(sb.String()); err != nil {
		b.Fatal(err)
	}
	return path
}

func BenchmarkStreamCollect(b *testing.B) {
	path := makeBenchCSV(b, 100000)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		stream.FromCSV(path).Collect()
	}
}