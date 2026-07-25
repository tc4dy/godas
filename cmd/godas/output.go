package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/tc4dy/godas/dataframe"
)

type OutputFormat int

const (
	Table OutputFormat = iota
	JSON
	CSV
)

func printDataFrame(df *dataframe.DataFrame, format OutputFormat) error {
	switch format {
	case Table:
		return printTable(df)
	case JSON:
		return printJSON(df)
	case CSV:
		return printCSV(df)
	default:
		return fmt.Errorf("unsupported output format")
	}
}

func printTable(df *dataframe.DataFrame) error {
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(df.ColumnNames())
	rows := df.NRows()
	for i := 0; i < rows && i < 50; i++ {
		row := make([]string, df.NCols())
		for j, col := range df.ColumnNames() {
			s := df.MustCol(col)
			row[j] = s.StringAt(i)
		}
		table.Append(row)
	}
	if rows > 50 {
		table.Append([]string{"..."})
	}
	table.Render()
	return nil
}

func printJSON(df *dataframe.DataFrame) error {
	names := df.ColumnNames()
	n := df.NRows()
	records := make([]map[string]any, n)
	for i := 0; i < n; i++ {
		rec := make(map[string]any)
		for _, name := range names {
			s := df.MustCol(name)
			rec[name] = s.ValueAt(i)
		}
		records[i] = rec
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}

func printCSV(df *dataframe.DataFrame) error {
	w := csv.NewWriter(os.Stdout)
	if err := w.Write(df.ColumnNames()); err != nil {
		return err
	}
	names := df.ColumnNames()
	n := df.NRows()
	for i := 0; i < n; i++ {
		row := make([]string, len(names))
		for j, name := range names {
			s := df.MustCol(name)
			row[j] = s.StringAt(i)
		}
		if err := w.Write(row); err != nil {
			return err
		}
	}
	w.Flush()
	return w.Error()
}