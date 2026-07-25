package io

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"sync"

	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/series"
)

type NDJSONReadOptions struct {
	Workers int
}

func defaultNDJSONReadOptions() NDJSONReadOptions {
	return NDJSONReadOptions{
		Workers: runtime.NumCPU(),
	}
}

type NDJSONOption func(*NDJSONReadOptions)

func WithNDJSONWorkers(n int) NDJSONOption {
	return func(o *NDJSONReadOptions) { o.Workers = n }
}

func ReadNDJSON(path string, opts ...NDJSONOption) (*dataframe.DataFrame, error) {
	options := defaultNDJSONReadOptions()
	for _, o := range opts {
		o(&options)
	}

	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("godas: cannot open %q: %w", path, err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024*10)

	records := make([]map[string]any, 0)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var rec map[string]any
		if err := json.Unmarshal(line, &rec); err != nil {
			return nil, fmt.Errorf("godas: NDJSON parse error: %w", err)
		}
		records = append(records, rec)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("godas: NDJSON scan error: %w", err)
	}

	if len(records) == 0 {
		return dataframe.New()
	}

	keys := make([]string, 0)
	seen := make(map[string]struct{})
	for _, rec := range records {
		for k := range rec {
			if _, ok := seen[k]; !ok {
				keys = append(keys, k)
				seen[k] = struct{}{}
			}
		}
	}

	n := len(records)
	cols := make([]*series.Series, len(keys))
	var wg sync.WaitGroup
	sem := make(chan struct{}, options.Workers)

	for ki, k := range keys {
		wg.Add(1)
		sem <- struct{}{}
		go func(ki int, k string) {
			defer wg.Done()
			defer func() { <-sem }()
			raw := make([]any, n)
			for i, rec := range records {
				raw[i] = rec[k]
			}
			cols[ki] = inferJSONSeries(k, raw)
		}(ki, k)
	}
	wg.Wait()

	return dataframe.New(cols...)
}

func WriteNDJSON(df *dataframe.DataFrame, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("godas: cannot create %q: %w", path, err)
	}
	defer f.Close()

	names := df.ColumnNames()
	n := df.NRows()
	writer := bufio.NewWriterSize(f, 1<<20)
	enc := json.NewEncoder(writer)

	for i := 0; i < n; i++ {
		rec := make(map[string]any, len(names))
		for _, name := range names {
			s := df.MustCol(name)
			rec[name] = s.ValueAt(i)
		}
		if err := enc.Encode(rec); err != nil {
			return fmt.Errorf("godas: NDJSON encode: %w", err)
		}
	}

	if err := writer.Flush(); err != nil {
		return fmt.Errorf("godas: NDJSON flush: %w", err)
	}
	return nil
}