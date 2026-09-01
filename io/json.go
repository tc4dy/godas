package io

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/series"
)

func ReadJSON(path string) (*dataframe.DataFrame, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("godas: cannot read %q: %w", path, err)
	}

	var records []map[string]any
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("godas: JSON parse error in %q: %w", path, err)
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
	cols := make([]*series.Series, 0, len(keys))

	for _, k := range keys {
		raw := make([]any, n)
		for i, rec := range records {
			raw[i] = rec[k]
		}
		s := inferJSONSeries(k, raw)
		cols = append(cols, s)
	}

	return dataframe.New(cols...)
}

func inferJSONSeries(name string, vals []any) *series.Series {
	allFloat := true
	allBool := true
	allString := true

	for _, v := range vals {
		if v == nil {
			continue
		}
		switch v.(type) {
		case float64:
			allBool = false
		case bool:
			allFloat = false
		case string:
			allFloat = false
			allBool = false
		default:
			allFloat = false
			allBool = false
			allString = false
		}
	}

	if allBool {
		data := make([]bool, len(vals))
		nulls := make([]bool, len(vals))
		for i, v := range vals {
			if v == nil {
				nulls[i] = true
				continue
			}
			if b, ok := v.(bool); ok {
				data[i] = b
			} else {
				nulls[i] = true
			}
		}
		s := series.NewBool(name, data)
		copy(s.Nulls(), nulls)
		return s
	}

	if allFloat {
		data := make([]float64, len(vals))
		for i, v := range vals {
			if v == nil {
				data[i] = math.NaN()
				continue
			}
			if f, ok := v.(float64); ok {
				data[i] = f
			} else {
				data[i] = math.NaN()
			}
		}
		return series.NewFloat64(name, data)
	}

	if allString {
		data := make([]string, len(vals))
		nulls := make([]bool, len(vals))
		for i, v := range vals {
			if v == nil {
				nulls[i] = true
				continue
			}
			if s, ok := v.(string); ok {
				data[i] = s
			} else {
				nulls[i] = true
			}
		}
		s := series.NewString(name, data)
		copy(s.Nulls(), nulls)
		return s
	}

	data := make([]string, len(vals))
	nulls := make([]bool, len(vals))
	for i, v := range vals {
		if v == nil {
			nulls[i] = true
			continue
		}
		switch val := v.(type) {
		case float64:
			data[i] = strconv.FormatFloat(val, 'f', -1, 64)
		case bool:
			data[i] = strconv.FormatBool(val)
		case string:
			data[i] = val
		default:
			b, _ := json.Marshal(val)
			data[i] = string(b)
		}
	}
	s := series.NewString(name, data)
	copy(s.Nulls(), nulls)
	return s
}

func WriteJSON(df *dataframe.DataFrame, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("godas: cannot create %q: %w", path, err)
	}
	defer f.Close()

	names := df.ColumnNames()
	n := df.NRows()
	records := make([]map[string]any, n)

	for ri := 0; ri < n; ri++ {
		rec := make(map[string]any, len(names))
		for _, name := range names {
			s := df.MustCol(name)
			rec[name] = s.ValueAt(ri)
		}
		records[ri] = rec
	}

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(records); err != nil {
		return fmt.Errorf("godas: JSON encode: %w", err)
	}
	return nil
}
