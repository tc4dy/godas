package io_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/io"
	"github.com/tc4dy/godas/series"
)

func TestReadNDJSONBasic(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "data.ndjson")
	content := `{"name":"Alice","age":30}
{"name":"Bob","age":25}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	df, err := io.ReadNDJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if df.NRows() != 2 {
		t.Fatalf("expected 2 rows, got %d", df.NRows())
	}
	if df.NCols() != 2 {
		t.Fatalf("expected 2 cols, got %d", df.NCols())
	}
}

func TestReadNDJSONEmpty(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "empty.ndjson")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	df, err := io.ReadNDJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if df.NRows() != 0 {
		t.Fatalf("expected 0 rows, got %d", df.NRows())
	}
}

func TestReadNDJSONWithEmptyLines(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "data.ndjson")
	content := `{"name":"Alice","age":30}

{"name":"Bob","age":25}
`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	df, err := io.ReadNDJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if df.NRows() != 2 {
		t.Fatalf("expected 2 rows, got %d", df.NRows())
	}
}

func TestReadNDJSONMissingFile(t *testing.T) {
	_, err := io.ReadNDJSON("/nonexistent/file.ndjson")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestWriteNDJSON(t *testing.T) {
	s1 := series.NewString("name", []string{"Alice", "Bob"})
	s2 := series.NewFloat64("age", []float64{30, 25})
	df, _ := dataframe.New(s1, s2)
	tmp := t.TempDir()
	path := filepath.Join(tmp, "out.ndjson")
	err := io.WriteNDJSON(df, path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}
