package io

import (
	"fmt"
	"math"
	"os"

	"github.com/apache/arrow/go/v14/arrow"
	"github.com/apache/arrow/go/v14/arrow/array"
	"github.com/apache/arrow/go/v14/arrow/memory"
	"github.com/apache/arrow/go/v14/parquet"
	"github.com/apache/arrow/go/v14/parquet/arrow/reader"
	"github.com/apache/arrow/go/v14/parquet/arrow/writer"
	"github.com/apache/arrow/go/v14/parquet/file"

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

	arrowRdr, err := reader.NewParquetReader(rdr, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("godas: arrow reader error: %w", err)
	}
	defer arrowRdr.Close()

	allCols := make([]*series.Series, 0)
	colNames := make(map[string]bool)

	for {
		rec, err := arrowRdr.Read()
		if err != nil {
			break
		}

		for i := 0; i < int(rec.NumCols()); i++ {
			col := rec.Column(i)
			name := rec.Schema().Field(i).Name

			if colNames[name] {
				continue
			}
			colNames[name] = true

			s := arrowToSeries(name, col)
			allCols = append(allCols, s)
		}
		rec.Release()
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
		for i := 0; i < data.Len(); i++ {
			if data.IsNull(i) {
				vals[i] = math.NaN()
			} else {
				vals[i] = data.Value(i)
			}
		}
		return series.NewFloat64(name, vals)

	case arrow.INT64:
		data := col.(*array.Int64)
		vals := make([]int64, data.Len())
		for i := 0; i < data.Len(); i++ {
			vals[i] = data.Value(i)
		}
		return series.NewInt64(name, vals)

	case arrow.INT32:
		data := col.(*array.Int32)
		vals := make([]int64, data.Len())
		for i := 0; i < data.Len(); i++ {
			vals[i] = int64(data.Value(i))
		}
		return series.NewInt64(name, vals)

	case arrow.STRING:
		data := col.(*array.String)
		vals := make([]string, data.Len())
		for i := 0; i < data.Len(); i++ {
			if data.IsNull(i) {
				continue
			}
			vals[i] = data.Value(i)
		}
		return series.NewString(name, vals)

	case arrow.BOOLEAN:
		data := col.(*array.Boolean)
		vals := make([]bool, data.Len())
		for i := 0; i < data.Len(); i++ {
			vals[i] = data.Value(i)
		}
		return series.NewBool(name, vals)

	default:
		data := col.(*array.String)
		vals := make([]string, data.Len())
		for i := 0; i < data.Len(); i++ {
			if data.IsNull(i) {
				continue
			}
			vals[i] = data.Value(i)
		}
		return series.NewString(name, vals)
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
			for j, v := range raw {
				if s.IsNull(j) {
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
			for j, v := range raw {
				if s.IsNull(j) {
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
			for j, v := range raw {
				if s.IsNull(j) {
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
			for j, v := range raw {
				if s.IsNull(j) {
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
	wr, err := writer.NewParquetWriter(f, schema, props)
	if err != nil {
		return fmt.Errorf("godas: parquet writer error: %w", err)
	}
	defer wr.Close()

	if err := wr.Write(rec); err != nil {
		return fmt.Errorf("godas: parquet write error: %w", err)
	}

	return nil
}