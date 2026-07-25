package io_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/io"
	"github.com/tc4dy/godas/series"
)

func TestReadJSONBasic(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "data.json")
	content := `[{"name":"Alice","age":30},{"name":"Bob","age":25}]`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	df, err := io.ReadJSON(path)
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

func TestReadJSONEmpty(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "empty.json")
	if err := os.WriteFile(path, []byte("[]"), 0644); err != nil {
		t.Fatal(err)
	}
	df, err := io.ReadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	if df.NRows() != 0 {
		t.Fatalf("expected 0 rows, got %d", df.NRows())
	}
}

func TestReadJSONMissingFile(t *testing.T) {
	_, err := io.ReadJSON("/nonexistent/file.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestReadJSONMixedTypes(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "mixed.json")
	content := `[{"id":1,"active":true},{"id":2,"active":false}]`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	df, err := io.ReadJSON(path)
	if err != nil {
		t.Fatal(err)
	}
	active, _ := df.Col("active")
	if active.DType() != series.Bool {
		t.Fatalf("expected bool type, got %v", active.DType())
	}
}

func TestWriteJSON(t *testing.T) {
	s1 := series.NewString("name", []string{"Alice", "Bob"})
	s2 := series.NewFloat64("age", []float64{30, 25})
	df, _ := dataframe.New(s1, s2)
	tmp := t.TempDir()
	path := filepath.Join(tmp, "out.json")
	err := io.WriteJSON(df, path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := `[
  {
    "age": 30,
    "name": "Alice"
  },
  {
    "age": 25,
    "name": "Bob"
  }
]
`
	if string(data) != expected {
		t.Fatalf("expected %q, got %q", expected, string(data))
	}
}