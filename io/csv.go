package io

import (
	"bufio"
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"

	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/series"
)

type CSVReadOptions struct {
	Delimiter rune
	HasHeader bool
	Workers   int
	NullValue string
}

func defaultCSVReadOptions() CSVReadOptions {
	return CSVReadOptions{
		Delimiter: ',',
		HasHeader: true,
		Workers:   runtime.NumCPU(),
		NullValue: "",
	}
}

type CSVOption func(*CSVReadOptions)

func WithDelimiter(d rune) CSVOption    { return func(o *CSVReadOptions) { o.Delimiter = d } }
func WithNoHeader() CSVOption           { return func(o *CSVReadOptions) { o.HasHeader = false } }
func WithWorkers(n int) CSVOption       { return func(o *CSVReadOptions) { o.Workers = n } }
func WithNullValue(v string) CSVOption  { return func(o *CSVReadOptions) { o.NullValue = v } }

func ReadCSV(path string, opts ...CSVOption) (*dataframe.DataFrame, error) {
	options := defaultCSVReadOptions()
	for _, o := range opts {
		o(&options)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("godas: cannot open %q: %w", path, err)
	}
	defer f.Close()

	r := csv.NewReader(bufio.NewReaderSize(f, 1<<20))
	r.Comma = options.Delimiter
	r.ReuseRecord = true

	records, err := r.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("godas: CSV parse error in %q: %w", path, err)
	}

	if len(records) == 0 {
		return dataframe.New()
	}

	var headers []string
	startRow := 0
	if options.HasHeader {
		if len(records) == 0 {
			return dataframe.New()
		}
		headers = make([]string, len(records[0]))
		copy(headers, records[0])
		startRow = 1
	} else {
		headers = make([]string, len(records[0]))
		for i := range headers {
			headers[i] = fmt.Sprintf("col%d", i)
		}
	}

	if startRow >= len(records) {
		emptyCols := make([]*series.Series, len(headers))
		for i, h := range headers {
			emptyCols[i] = series.NewString(h, []string{})
		}
		return dataframe.New(emptyCols...)
	}

	dataRows := records[startRow:]
	nRows := len(dataRows)
	nCols := len(headers)

	raw := make([][]string, nCols)
	for i := range raw {
		raw[i] = make([]string, nRows)
	}
	for ri, row := range dataRows {
		for ci := 0; ci < nCols && ci < len(row); ci++ {
			raw[ci][ri] = row[ci]
		}
	}

	cols := make([]*series.Series, nCols)
	var wg sync.WaitGroup
	errs := make([]error, nCols)

	sem := make(chan struct{}, options.Workers)
	for ci := 0; ci < nCols; ci++ {
		wg.Add(1)
		sem <- struct{}{}
		go func(ci int) {
			defer wg.Done()
			defer func() { <-sem }()
			cols[ci] = inferAndBuild(headers[ci], raw[ci], options.NullValue)
		}(ci)
	}
	wg.Wait()

	for _, e := range errs {
		if e != nil {
			return nil, e
		}
	}
	return dataframe.New(cols...)
}

func inferAndBuild(name string, vals []string, nullVal string) *series.Series {
	allInt := true
	allFloat := true
	allBool := true

	for _, v := range vals {
		if v == nullVal {
			continue
		}
		if _, err := strconv.ParseInt(v, 10, 64); err != nil {
			allInt = false
		}
		if _, err := strconv.ParseFloat(v, 64); err != nil {
			allFloat = false
		}
		lower := strings.ToLower(v)
		if lower != "true" && lower != "false" {
			allBool = false
		}
	}

	if allInt {
		data := make([]int64, len(vals))
		for i, v := range vals {
			if v == nullVal {
				continue
			}
			data[i], _ = strconv.ParseInt(v, 10, 64)
		}
		return series.NewInt64(name, data)
	}

	if allFloat {
		data := make([]float64, len(vals))
		for i, v := range vals {
			if v == nullVal {
				data[i] = math.NaN()
				continue
			}
			data[i], _ = strconv.ParseFloat(v, 64)
		}
		return series.NewFloat64(name, data)
	}

	if allBool {
		data := make([]bool, len(vals))
		for i, v := range vals {
			if v == nullVal {
				continue
			}
			data[i] = strings.ToLower(v) == "true"
		}
		return series.NewBool(name, data)
	}

	data := make([]string, len(vals))
	copy(data, vals)
	return series.NewString(name, data)
}

type CSVWriteOptions struct {
	Delimiter rune
}

type CSVWriteOption func(*CSVWriteOptions)

func WithWriteDelimiter(d rune) CSVWriteOption {
	return func(o *CSVWriteOptions) { o.Delimiter = d }
}

func WriteCSV(df *dataframe.DataFrame, path string, opts ...CSVWriteOption) error {
	options := CSVWriteOptions{Delimiter: ','}
	for _, o := range opts {
		o(&options)
	}

	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("godas: cannot create %q: %w", path, err)
	}
	defer f.Close()

	bw := bufio.NewWriterSize(f, 1<<20)
	w := csv.NewWriter(bw)
	w.Comma = options.Delimiter

	if err := w.Write(df.ColumnNames()); err != nil {
		return fmt.Errorf("godas: CSV write header: %w", err)
	}

	names := df.ColumnNames()
	n := df.NRows()
	row := make([]string, len(names))

	for ri := 0; ri < n; ri++ {
		for ci, name := range names {
			s := df.MustCol(name)
			row[ci] = s.StringAt(ri)
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("godas: CSV write row %d: %w", ri, err)
		}
	}

	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("godas: CSV flush: %w", err)
	}
	return bw.Flush()
}