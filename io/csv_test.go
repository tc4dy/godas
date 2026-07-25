package io_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/io"
	"github.com/tc4dy/godas/series"
)

func TestReadCSVBasic(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "data.csv")
	content := "name,age\nAlice,30\nBob,25\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	df, err := io.ReadCSV(path)
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

func TestReadCSVNoHeader(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "data.csv")
	content := "Alice,30\nBob,25\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	df, err := io.ReadCSV(path, io.WithNoHeader())
	if err != nil {
		t.Fatal(err)
	}
	cols := df.ColumnNames()
	if cols[0] != "col0" || cols[1] != "col1" {
		t.Fatalf("expected col0,col1, got %v", cols)
	}
}

func TestReadCSVDelimiter(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "data.csv")
	content := "name;age\nAlice;30\nBob;25\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	df, err := io.ReadCSV(path, io.WithDelimiter(';'))
	if err != nil {
		t.Fatal(err)
	}
	if df.NRows() != 2 {
		t.Fatalf("expected 2 rows, got %d", df.NRows())
	}
}

func TestReadCSVWithNull(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "data.csv")
	content := "name,age\nAlice,30\nBob,NA\n"
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	df, err := io.ReadCSV(path, io.WithNullValue("NA"))
	if err != nil {
		t.Fatal(err)
	}
	ageCol, _ := df.Col("age")
	if !ageCol.IsNull(1) {
		t.Fatal("expected null for second row")
	}
}

func TestReadCSVEmpty(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "empty.csv")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}
	df, err := io.ReadCSV(path)
	if err != nil {
		t.Fatal(err)
	}
	if df.NRows() != 0 {
		t.Fatalf("expected 0 rows, got %d", df.NRows())
	}
}

func TestReadCSVMissingFile(t *testing.T) {
	_, err := io.ReadCSV("/nonexistent/file.csv")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestWriteCSV(t *testing.T) {
	s1 := series.NewString("name", []string{"Alice", "Bob"})
	s2 := series.NewFloat64("age", []float64{30, 25})
	df, _ := dataframe.New(s1, s2)
	tmp := t.TempDir()
	path := filepath.Join(tmp, "out.csv")
	err := io.WriteCSV(df, path)
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := "name,age\nAlice,30.0000\nBob,25.0000\n"
	if string(data) != expected {
		t.Fatalf("expected %q, got %q", expected, string(data))
	}
}

func TestWriteCSVWithDelimiter(t *testing.T) {
	s := series.NewString("x", []string{"a", "b"})
	df, _ := dataframe.New(s)
	tmp := t.TempDir()
	path := filepath.Join(tmp, "out.csv")
	err := io.WriteCSV(df, path, io.WithWriteDelimiter(';'))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	expected := "x\na\nb\n"
	if string(data) != expected {
		t.Fatalf("expected %q, got %q", expected, string(data))
	}
}