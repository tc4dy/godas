package io

import (
	"context"
	"fmt"
	"math"
	"os"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"github.com/apache/arrow/go/v14/parquet"
	"github.com/apache/arrow/go/v14/parquet/file"
	"github.com/apache/arrow/go/v14/parquet/pqarrow"

	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/series"
)

type ParquetReadOptions struct {
	BatchSize int64
	Workers   int
}

func defaultParquetReadOptions() ParquetReadOptions {
	return ParquetReadOptions{
		BatchSize: 100000,
		Workers:   4,
	}
}

type ParquetOption func(*ParquetReadOptions)

func WithParquetBatchSize(size int64) ParquetOption {
	return func(o *ParquetReadOptions) { o.BatchSize = size }
}

func WithParquetWorkers(n int) ParquetOption {
	return func(o *ParquetReadOptions) { o.Workers = n }
}

func ReadParquet(path string, opts ...ParquetOption) (*dataframe.DataFrame, error) {
	options := defaultParquetReadOptions()
	for _, o := range opts {
		o(&options)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("godas: cannot open %q: %w", path, err)
	}
	defer f.Close()

	rdr, err := file.NewParquetReader(f)
	if err != nil {
		return nil, fmt.Errorf("godas: parquet reader error: %w", err)
	}
	defer rdr.Close()

	arrowRdr, err := pqarrow.NewFileReader(rdr, pqarrow.ArrowReadProperties{}, memory.DefaultAllocator)
	if err != nil {
		return nil, fmt.Errorf("godas: arrow reader error: %w", err)
	}

	tbl, err := arrowRdr.ReadTable(context.Background())
	if err != nil {
		return nil, fmt.Errorf("godas: read table error: %w", err)
	}
	defer tbl.Release()

	allCols := make([]*series.Series, 0, int(tbl.NumCols()))

	for i := 0; i < int(tbl.NumCols()); i++ {
		col := tbl.Column(i)
		name := tbl.Schema().Field(i).Name
		chunks := col.Data().Chunks()
		if len(chunks) == 0 {
			continue
		}
		s := arrowToSeries(name, chunks[0])
		allCols = append(allCols, s)
	}

	if len(allCols) == 0 {
		return dataframe.New()
	}

	return dataframe.New(allCols...)
}

func arrowToSeries(name string, col arrow.Array) *series.Series {
	switch col.DataType().ID() {
	case arrow.FLOAT64:
		data := col.(*array.Float64)
		vals := make([]float64, data.Len())
		nulls := make([]bool, data.Len())
		for i := 0; i < data.Len(); i++ {
			if data.IsNull(i) {
				nulls[i] = true
				vals[i] = math.NaN()
			} else {
				vals[i] = data.Value(i)
			}
		}
		s := series.NewFloat64(name, vals)
		copy(s.Nulls(), nulls)
		return s

	case arrow.INT64:
		data := col.(*array.Int64)
		vals := make([]int64, data.Len())
		nulls := make([]bool, data.Len())
		for i := 0; i < data.Len(); i++ {
			if data.IsNull(i) {
				nulls[i] = true
				continue
			}
			vals[i] = data.Value(i)
		}
		s := series.NewInt64(name, vals)
		copy(s.Nulls(), nulls)
		return s

	case arrow.INT32:
		data := col.(*array.Int32)
		vals := make([]int64, data.Len())
		nulls := make([]bool, data.Len())
		for i := 0; i < data.Len(); i++ {
			if data.IsNull(i) {
				nulls[i] = true
				continue
			}
			vals[i] = int64(data.Value(i))
		}
		s := series.NewInt64(name, vals)
		copy(s.Nulls(), nulls)
		return s

	case arrow.STRING:
		data := col.(*array.String)
		vals := make([]string, data.Len())
		nulls := make([]bool, data.Len())
		for i := 0; i < data.Len(); i++ {
			if data.IsNull(i) {
				nulls[i] = true
				continue
			}
			vals[i] = data.Value(i)
		}
		s := series.NewString(name, vals)
		copy(s.Nulls(), nulls)
		return s

	case arrow.BOOL:
		data := col.(*array.Boolean)
		vals := make([]bool, data.Len())
		nulls := make([]bool, data.Len())
		for i := 0; i < data.Len(); i++ {
			if data.IsNull(i) {
				nulls[i] = true
				continue
			}
			vals[i] = data.Value(i)
		}
		s := series.NewBool(name, vals)
		copy(s.Nulls(), nulls)
		return s

	default:
		return series.NewString(name, []string{})
	}
}

func WriteParquet(df *dataframe.DataFrame, path string) error {
	if df.NRows() == 0 {
		return fmt.Errorf("godas: cannot write empty dataframe to parquet")
	}

	names := df.ColumnNames()
	fields := make([]arrow.Field, len(names))
	arrays := make([]arrow.Array, len(names))

	for i, name := range names {
		s := df.MustCol(name)

		switch s.DType() {
		case series.Float64:
			fields[i] = arrow.Field{Name: name, Type: arrow.PrimitiveTypes.Float64}
			builder := array.NewFloat64Builder(memory.DefaultAllocator)
			raw := s.RawFloats()
			nulls := s.Nulls()
			for j, v := range raw {
				if nulls[j] {
					builder.AppendNull()
				} else {
					builder.Append(v)
				}
			}
			arrays[i] = builder.NewArray()
			builder.Release()

		case series.Int64:
			fields[i] = arrow.Field{Name: name, Type: arrow.PrimitiveTypes.Int64}
			builder := array.NewInt64Builder(memory.DefaultAllocator)
			raw := s.RawInts()
			nulls := s.Nulls()
			for j, v := range raw {
				if nulls[j] {
					builder.AppendNull()
				} else {
					builder.Append(v)
				}
			}
			arrays[i] = builder.NewArray()
			builder.Release()

		case series.String:
			fields[i] = arrow.Field{Name: name, Type: arrow.BinaryTypes.String}
			builder := array.NewStringBuilder(memory.DefaultAllocator)
			raw := s.RawStrings()
			nulls := s.Nulls()
			for j, v := range raw {
				if nulls[j] {
					builder.AppendNull()
				} else {
					builder.Append(v)
				}
			}
			arrays[i] = builder.NewArray()
			builder.Release()

		case series.Bool:
			fields[i] = arrow.Field{Name: name, Type: arrow.FixedWidthTypes.Boolean}
			builder := array.NewBooleanBuilder(memory.DefaultAllocator)
			raw := s.RawBools()
			nulls := s.Nulls()
			for j, v := range raw {
				if nulls[j] {
					builder.AppendNull()
				} else {
					builder.Append(v)
				}
			}
			arrays[i] = builder.NewArray()
			builder.Release()
		}
	}

	schema := arrow.NewSchema(fields, nil)
	rec := array.NewRecord(schema, arrays, int64(df.NRows()))
	defer rec.Release()

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("godas: cannot create %q: %w", path, err)
	}
	defer f.Close()

	props := parquet.NewWriterProperties()
	arrowProps := pqarrow.NewArrowWriterProperties()
	wr, err := pqarrow.NewFileWriter(schema, f, props, arrowProps)
	if err != nil {
		return fmt.Errorf("godas: parquet writer error: %w", err)
	}
	defer wr.Close()

	if err := wr.Write(rec); err != nil {
		return fmt.Errorf("godas: parquet write error: %w", err)
	}

	return nil
}
