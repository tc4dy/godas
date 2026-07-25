package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tc4dy/godas/agg"
	"github.com/tc4dy/godas/dataframe"
	"github.com/tc4dy/godas/expr"
	"github.com/tc4dy/godas/io"
	"github.com/tc4dy/godas/join"
	"github.com/tc4dy/godas/stats"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "godas",
		Short:   "Go Data Analysis Suite",
		Long:    "GODAS - High-performance data analysis for Go",
		Version: "1.0.0",
	}
	cmd.AddCommand(
		NewLoadCmd(),
		NewFilterCmd(),
		NewGroupCmd(),
		NewStatsCmd(),
		NewJoinCmd(),
		NewExportCmd(),
	)
	return cmd
}

func NewLoadCmd() *cobra.Command {
	var format string
	cmd := &cobra.Command{
		Use:   "load [file]",
		Short: "Load data from file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			file := args[0]
			var df *dataframe.DataFrame
			var err error
			switch format {
			case "csv":
				df, err = io.ReadCSV(file)
			case "json":
				df, err = io.ReadJSON(file)
			default:
				return fmt.Errorf("unsupported format: %s", format)
			}
			if err != nil {
				return err
			}
			df.Print()
			return nil
		},
	}
	cmd.Flags().StringVarP(&format, "format", "f", "csv", "Input format (csv, json)")
	return cmd
}

func NewFilterCmd() *cobra.Command {
	var file, condition, format string
	cmd := &cobra.Command{
		Use:   "filter",
		Short: "Filter rows by condition",
		RunE: func(cmd *cobra.Command, args []string) error {
			df, err := loadData(file, format)
			if err != nil {
				return err
			}
			e := parseExpr(condition)
			filtered, err := df.Filter(e)
			if err != nil {
				return err
			}
			filtered.Print()
			return nil
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Input file")
	cmd.Flags().StringVarP(&condition, "condition", "c", "", "Filter condition")
	cmd.Flags().StringVarP(&format, "format", "t", "csv", "File format")
	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("condition")
	return cmd
}

func NewGroupCmd() *cobra.Command {
	var file, by, format string
	var aggs []string
	cmd := &cobra.Command{
		Use:   "group",
		Short: "Group by columns and aggregate",
		RunE: func(cmd *cobra.Command, args []string) error {
			df, err := loadData(file, format)
			if err != nil {
				return err
			}
			grouped := df.GroupBy(by)
			result, err := grouped.Aggregate(parseAggs(aggs)...)
			if err != nil {
				return err
			}
			result.Print()
			return nil
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Input file")
	cmd.Flags().StringVarP(&by, "by", "b", "", "Group by column")
	cmd.Flags().StringSliceVarP(&aggs, "agg", "a", []string{}, "Aggregations (e.g., sum:col, mean:col)")
	cmd.Flags().StringVarP(&format, "format", "t", "csv", "File format")
	cmd.MarkFlagRequired("file")
	cmd.MarkFlagRequired("by")
	return cmd
}

func NewStatsCmd() *cobra.Command {
	var file, format string
	cmd := &cobra.Command{
		Use:   "stats [column]",
		Short: "Show statistics for a column",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			col := args[0]
			df, err := loadData(file, format)
			if err != nil {
				return err
			}
			s, err := df.Col(col)
			if err != nil {
				return err
			}
			sum := stats.Describe(s)
			fmt.Printf("Count: %d\n", sum.Count)
			fmt.Printf("Mean: %f\n", sum.Mean)
			fmt.Printf("Std: %f\n", sum.Std)
			fmt.Printf("Min: %f\n", sum.Min)
			fmt.Printf("25%%: %f\n", sum.Q25)
			fmt.Printf("Median: %f\n", sum.Median)
			fmt.Printf("75%%: %f\n", sum.Q75)
			fmt.Printf("Max: %f\n", sum.Max)
			fmt.Printf("Nulls: %d\n", sum.NNull)
			return nil
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Input file")
	cmd.Flags().StringVarP(&format, "format", "t", "csv", "File format")
	cmd.MarkFlagRequired("file")
	return cmd
}

func NewJoinCmd() *cobra.Command {
	var file1, file2, on, format, joinType string
	cmd := &cobra.Command{
		Use:   "join",
		Short: "Join two datasets",
		RunE: func(cmd *cobra.Command, args []string) error {
			left, err := loadData(file1, format)
			if err != nil {
				return err
			}
			right, err := loadData(file2, format)
			if err != nil {
				return err
			}
			var result *dataframe.DataFrame
			switch joinType {
			case "inner":
				result, err = join.Inner(left, right, on)
			case "left":
				result, err = join.Left(left, right, on)
			case "right":
				result, err = join.Right(left, right, on)
			default:
				return fmt.Errorf("unsupported join type: %s", joinType)
			}
			if err != nil {
				return err
			}
			result.Print()
			return nil
		},
	}
	cmd.Flags().StringVarP(&file1, "left", "l", "", "Left file")
	cmd.Flags().StringVarP(&file2, "right", "r", "", "Right file")
	cmd.Flags().StringVarP(&on, "on", "o", "", "Join column")
	cmd.Flags().StringVarP(&format, "format", "t", "csv", "File format")
	cmd.Flags().StringVarP(&joinType, "type", "y", "inner", "Join type (inner, left, right)")
	cmd.MarkFlagRequired("left")
	cmd.MarkFlagRequired("right")
	cmd.MarkFlagRequired("on")
	return cmd
}

func NewExportCmd() *cobra.Command {
	var file, output, format string
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export data to file",
		RunE: func(cmd *cobra.Command, args []string) error {
			df, err := loadData(file, format)
			if err != nil {
				return err
			}
			switch format {
			case "csv":
				err = io.WriteCSV(df, output)
			case "json":
				err = io.WriteJSON(df, output)
			default:
				return fmt.Errorf("unsupported export format: %s", format)
			}
			if err != nil {
				return err
			}
			fmt.Printf("Exported to %s\n", output)
			return nil
		},
	}
	cmd.Flags().StringVarP(&file, "file", "f", "", "Input file")
	cmd.Flags().StringVarP(&output, "output", "o", "output.csv", "Output file")
	cmd.Flags().StringVarP(&format, "format", "t", "csv", "Output format (csv, json)")
	cmd.MarkFlagRequired("file")
	return cmd
}

func loadData(file, format string) (*dataframe.DataFrame, error) {
	switch format {
	case "csv":
		return io.ReadCSV(file)
	case "json":
		return io.ReadJSON(file)
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

func parseExpr(cond string) expr.Expr {
	parts := strings.SplitN(cond, ">", 2)
	if len(parts) == 2 {
		col := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		return expr.Col(col).Gt(parseFloat(val))
	}
	parts = strings.SplitN(cond, "<", 2)
	if len(parts) == 2 {
		col := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		return expr.Col(col).Lt(parseFloat(val))
	}
	parts = strings.SplitN(cond, "==", 2)
	if len(parts) == 2 {
		col := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return expr.Col(col).Eq(f)
		}
		return expr.Col(col).Eq(val)
	}
	return nil
}

func parseFloat(s string) float64 {
	f, _ := strconv.ParseFloat(s, 64)
	return f
}

func parseAggs(aggs []string) []agg.Agg {
	out := make([]agg.Agg, 0, len(aggs))
	for _, a := range aggs {
		parts := strings.SplitN(a, ":", 2)
		if len(parts) != 2 {
			continue
		}
		fn, col := parts[0], parts[1]
		switch fn {
		case "sum":
			out = append(out, agg.Sum(col))
		case "mean":
			out = append(out, agg.Mean(col))
		case "count":
			out = append(out, agg.Count(col))
		case "min":
			out = append(out, agg.Min(col))
		case "max":
			out = append(out, agg.Max(col))
		case "std":
			out = append(out, agg.Std(col))
		case "median":
			out = append(out, agg.Median(col))
		case "first":
			out = append(out, agg.First(col))
		case "last":
			out = append(out, agg.Last(col))
		}
	}
	return out
}