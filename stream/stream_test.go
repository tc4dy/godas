package stream_test

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/expr"
	"github.com/tc4dy/godas/stream"
)

func makeCSVFile(t *testing.T, rows int) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "*.csv")
	if err != nil {
		t.Fatal(err)
	}
	var sb strings.Builder
	sb.WriteString("id,value\n")
	for i := 0; i < rows; i++ {
		sb.WriteString(fmt.Sprintf("%d,%d\n", i, i*10))
	}
	f.WriteString(sb.String())
	f.Close()
	return f.Name()
}

func TestStreamBasic(t *testing.T) {
	path := makeCSVFile(t, 500)
	df, err := stream.FromCSV(path).Chunk(100).Collect()
	if err != nil {
		t.Fatal(err)
	}
	if df.NRows() != 500 {
		t.Fatalf("expected 500 rows, got %d", df.NRows())
	}
}

func TestStreamWithFilter(t *testing.T) {
	path := makeCSVFile(t, 1000)
	df, err := stream.FromCSV(path).
		Chunk(200).
		Pipe(stream.Func(func(chunk *dataframe.DataFrame) (*dataframe.DataFrame, error) {
			return chunk.Filter(expr.Col("value").Gt(float64(4990)))
		})).
		Collect()
	if err != nil {
		t.Fatal(err)
	}
	if df.NRows() == 0 {
		t.Fatal("expected some rows after filter")
	}
}

func TestStreamMissingFile(t *testing.T) {
	_, err := stream.FromCSV("/nonexistent/file.csv").Collect()
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}