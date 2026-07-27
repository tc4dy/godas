package io_test

import (
	"path/filepath"
	"testing"

	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/io"
	"github.com/tc4dy/godas/series"
)

func TestReadParquetBasic(t *testing.T) {
	s1 := series.NewString("name", []string{"Alice", "Bob"})
	s2 := series.NewFloat64("age", []float64{30, 25})
	df, _ := dataframe.New(s1, s2)

	tmp := t.TempDir()
	path := filepath.Join(tmp, "data.parquet")

	err := io.WriteParquet(df, path)
	if err != nil {
		t.Fatal(err)
	}

	result, err := io.ReadParquet(path)
	if err != nil {
		t.Fatal(err)
	}

	if result.NRows() != 2 {
		t.Fatalf("expected 2 rows, got %d", result.NRows())
	}
	if result.NCols() != 2 {
		t.Fatalf("expected 2 cols, got %d", result.NCols())
	}
}

func TestReadParquetMissingFile(t *testing.T) {
	_, err := io.ReadParquet("/nonexistent/file.parquet")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestWriteParquetEmpty(t *testing.T) {
	df, _ := dataframe.New()
	tmp := t.TempDir()
	path := filepath.Join(tmp, "empty.parquet")

	err := io.WriteParquet(df, path)
	if err == nil {
		t.Fatal("expected error for empty dataframe")
	}
}

func TestParquetRoundTrip(t *testing.T) {
	s1 := series.NewString("name", []string{"Alice", "Bob", "Charlie"})
	s2 := series.NewInt64("id", []int64{1, 2, 3})
	s3 := series.NewFloat64("score", []float64{95.5, 87.3, 92.1})
	s4 := series.NewBool("active", []bool{true, false, true})

	df, _ := dataframe.New(s1, s2, s3, s4)

	tmp := t.TempDir()
	path := filepath.Join(tmp, "roundtrip.parquet")

	err := io.WriteParquet(df, path)
	if err != nil {
		t.Fatal(err)
	}

	result, err := io.ReadParquet(path)
	if err != nil {
		t.Fatal(err)
	}

	if result.NRows() != 3 {
		t.Fatalf("expected 3 rows, got %d", result.NRows())
	}
	if result.NCols() != 4 {
		t.Fatalf("expected 4 cols, got %d", result.NCols())
	}

	nameCol, _ := result.Col("name")
	if nameCol.Len() != 3 {
		t.Fatalf("expected name col len 3, got %d", nameCol.Len())
	}
}
