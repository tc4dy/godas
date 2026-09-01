![Godas Banner](godasbanner.png)
# GODAS | Go Data Analysis Suite 📁 (Godas Project) 

<div align="center">
  
[![Go Version](https://img.shields.io/badge/Go-1.22%2B-00ADD8?style=flat-square&logo=go)](https://golang.org/)
[![License](https://img.shields.io/badge/License-MIT-6A4E9C?style=flat-square)](LICENSE)
[![CI](https://img.shields.io/github/actions/workflow/status/tc4dy/godas/go.yml?branch=main&style=flat-square&label=CI&color=28A745)](https://github.com/tc4dy/godas/actions)
[![Awesome](https://img.shields.io/badge/Awesome-FFA500?style=flat-square&logo=awesome&logoColor=white)](https://github.com/avelino/awesome-go)
[![Coverage](https://img.shields.io/badge/coverage-80%25-6A4E9C?style=flat-square)](#testing)
[![Go Reference](https://img.shields.io/badge/Go-Reference-00ADD8?style=flat-square&logo=go)](https://pkg.go.dev/github.com/tc4dy/godas)

**A high-performance, idiomatic data analysis library for Go.**

*Type-safe columnar data structures · Expressive filtering · Concurrent aggregation · Multi-format I/O*

</div>

---

## Table of Contents

- [Overview](#overview)
- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Core Concepts](#core-concepts)
  - [Series](#series)
  - [DataFrame](#dataframe)
  - [Expressions](#expressions)
  - [Aggregations](#aggregations)
  - [Joins](#joins)
  - [Streaming](#streaming)
- [I/O Formats](#io-formats)
- [CLI Tool](#cli-tool)
- [Testing](#testing)
- [Benchmarks](#benchmarks)
- [Project Structure](#project-structure)
- [Contributing](#contributing)

---

## Overview

`godas` is a structured data manipulation library for Go, designed for analytical workloads that demand type safety, predictable memory layout, and concurrent execution. It provides a cohesive API for loading, transforming, aggregating, and exporting tabular data — with full support for null semantics across all operations.

It is built around a columnar in-memory model: each column is a strongly typed `Series`, and a `DataFrame` is an ordered collection of those columns. This design minimises allocations during bulk operations and makes parallelism straightforward.

---

## Features

### 🏗️ Data Structures
| Feature | Details |
|---|---|
| **Typed Series** | `float64`, `int64`, `string`, `bool` with per-element null tracking |
| **DataFrame** | Ordered, named columns with O(1) column lookup |
| **Null semantics** | NaN-aware for numerics; explicit null bitmap for all types |
| **Immutable operations** | Transformations return new structures; originals are never mutated |

### 🔍 Filtering & Expressions
| Feature | Details |
|---|---|
| **Comparison operators** | `Gt`, `Gte`, `Lt`, `Lte`, `Eq`, `Neq` for numeric and string columns |
| **String predicates** | `Contains`, `In` |
| **Logical composition** | `And`, `Or`, `Not` — composable to arbitrary depth |
| **Bool column filter** | Pass a pre-computed bool `Series` directly as a filter mask |

### 📊 Aggregations
| Function | Description |
|---|---|
| `Sum` | Sum of non-null values |
| `Mean` | Arithmetic mean, ignoring nulls |
| `Min` / `Max` | Extrema, ignoring nulls |
| `Count` | Count of non-null values |
| `Std` | Sample standard deviation |
| `Median` | Median via linear interpolation |
| `First` / `Last` | First or last non-null value |

All aggregations support `.As(alias)` for column renaming and are executed concurrently across groups during `GroupBy`.

### 🔗 Joins
- **Inner**, **Left**, and **Right** hash joins on a single key column
- Automatic suffix `_right` for name collisions from the right frame
- Null-row handling for unmatched keys in outer joins

### 📁 I/O Formats
| Format | Read | Write | Notes |
|---|---|---|---|
| CSV | ✅ | ✅ | Configurable delimiter, header, null token, worker count |
| JSON | ✅ | ✅ | Array-of-objects format |
| NDJSON | ✅ | ✅ | Newline-delimited JSON; concurrent column inference |
| Parquet | ✅ | ✅ | Via Apache Arrow; supports all four dtypes |

### ⚡ Performance
- Parallel CSV column parsing with a configurable worker pool
- Concurrent group aggregation via `sync.WaitGroup`
- Reusable slice allocations through `sync.Pool` (float64, string, bool)
- Streaming pipeline for large files via `stream.FromCSV(...).Chunk(...).Pipe(...).Collect()`

### 🖥️ CLI
A fully functional command-line interface (`godas`) for scripted and exploratory data work, built on [cobra](https://github.com/spf13/cobra).

---

## Installation

### Library

```bash
go get github.com/tc4dy/godas
```

Requires **Go 1.22** or later.

### CLI

```bash
go install github.com/tc4dy/godas/cmd/godas@latest
```

Or build from source:

```bash
git clone https://github.com/tc4dy/godas.git
cd godas
make build          # outputs to ./bin/godas
make install        # installs to $GOPATH/bin
```

---

## Quick Start

```go
package main

import (
    "fmt"

    "github.com/tc4dy/godas/agg"
    "github.com/tc4dy/godas/dataframe"
    "github.com/tc4dy/godas/expr"
    "github.com/tc4dy/godas/io"
    "github.com/tc4dy/godas/series"
)

func main() {
    // Build a DataFrame inline
    city := series.NewString("city", []string{"London", "Paris", "London", "Berlin", "Paris"})
    revenue := series.NewFloat64("revenue", []float64{120, 95, 140, 80, 110})
    active := series.NewBool("active", []bool{true, true, false, true, true})

    df, err := dataframe.New(city, revenue, active)
    if err != nil {
        panic(err)
    }

    // Filter: active rows with revenue > 100
    filtered, err := df.Filter(
        expr.And(
            expr.Col("active").Eq(true),
            expr.Col("revenue").Gt(100),
        ),
    )
    if err != nil {
        panic(err)
    }

    // Group by city and aggregate
    result, err := filtered.
        GroupBy("city").
        Aggregate(
            agg.Sum("revenue").As("total_revenue"),
            agg.Count("revenue").As("n_deals"),
        )
    if err != nil {
        panic(err)
    }

    result.Print()

    // Export
    if err := io.WriteCSV(result, "summary.csv"); err != nil {
        panic(err)
    }

    fmt.Printf("Wrote %d rows\n", result.NRows())
}
```

---

## Core Concepts

### Series

A `Series` is a typed, nullable, one-dimensional array. It is the fundamental building block of a `DataFrame`.

```go
// Construction
floats  := series.NewFloat64("price", []float64{9.99, 14.99, math.NaN()})
ints    := series.NewInt64("qty", []int64{1, 5, 3})
strings := series.NewString("sku", []string{"A001", "A002", "A003"})
bools   := series.NewBool("in_stock", []bool{true, false, true})

// Null awareness — NaN in float64 columns is automatically treated as null
floats.IsNull(2)          // true
v, ok := floats.GetFloat64(2)  // ok == false

// Slicing and reordering (zero-copy slice backed by original)
sub := floats.Slice(0, 2)
reordered := floats.Reorder([]int{2, 0, 1})

// Appending (returns a new Series)
merged, err := floats.Append(series.NewFloat64("price", []float64{19.99}))

// Cloning (deep copy)
copy := floats.Clone()
```

### DataFrame

A `DataFrame` is an ordered, indexed collection of `Series` columns with uniform row count.

```go
df, err := dataframe.New(city, revenue, active)

// Metadata
df.NRows()          // row count
df.NCols()          // column count
df.Shape()          // (rows, cols)
df.ColumnNames()    // []string
df.DTypes()         // map[string]series.DType
df.HasCol("city")   // bool

// Column access
s, err := df.Col("revenue")    // safe
s = df.MustCol("revenue")      // panics if missing

// Structural operations — all return new DataFrames
df.Select("city", "revenue")   // keep columns
df.Drop("active")              // remove column
df.WithColumn(newSeries)       // add or replace column
df.Rename("revenue", "sales")  // rename a column

// Row operations
df.Head(10)
df.Tail(10)
df.Slice(100, 200)
df.Sort("revenue", false)  // descending

// Combine
appended, err := df.Append(other)
clone := df.Clone()

// Summary statistics (numeric columns only)
desc := df.Describe()   // map[string]*stats.Summary
```

### Expressions

Expressions are composable predicates evaluated against a `DataFrame` to produce a boolean row mask.

```go
// Numeric comparisons
expr.Col("age").Gt(30)
expr.Col("age").Gte(30)
expr.Col("age").Lt(30)
expr.Col("age").Lte(30)
expr.Col("age").Eq(30)
expr.Col("age").Neq(30)

// String predicates
expr.Col("city").Eq("London")
expr.Col("city").Neq("Paris")
expr.Col("city").Contains("on")
expr.Col("city").In("London", "Berlin")

// Logical composition
expr.And(expr.Col("age").Gt(25), expr.Col("city").Eq("London"))
expr.Or(expr.Col("age").Lt(20), expr.Col("city").Eq("Paris"))
expr.Not(expr.Col("active").Eq(true))

// Nested
expr.And(
    expr.Col("age").Gt(25),
    expr.Or(
        expr.Col("city").Eq("London"),
        expr.Col("city").Eq("Berlin"),
    ),
)

// Apply
filtered, err := df.Filter(expr.Col("revenue").Gt(100.0))
```

### Aggregations

```go
// Available aggregations
agg.Sum("col")
agg.Mean("col")
agg.Min("col")
agg.Max("col")
agg.Count("col")   // counts non-null values
agg.Std("col")     // sample standard deviation
agg.Median("col")
agg.First("col")   // first non-null value
agg.Last("col")    // last non-null value

// Alias
agg.Sum("revenue").As("total_revenue")

// GroupBy + Aggregate
result, err := df.
    GroupBy("region", "category").
    Aggregate(
        agg.Sum("sales").As("total_sales"),
        agg.Mean("price").As("avg_price"),
        agg.Count("id").As("n_orders"),
    )

// Direct stats on a Series
s, _ := df.Col("revenue")
mean   := stats.Mean(s)
stddev := stats.Std(s)
corr   := stats.Corr(s1, s2)
desc   := stats.Describe(s)    // *stats.Summary
```

### Joins

```go
// Inner join — only matched rows
result, err := join.Inner(left, right, "id")

// Left join — all left rows, null-filled where unmatched
result, err := join.Left(left, right, "order_id")

// Right join — all right rows, null-filled where unmatched
result, err := join.Right(left, right, "customer_id")
```

Column names present in both frames (excluding the join key) are suffixed with `_right` on the right side automatically.

### Streaming

For files too large to load entirely into memory, use the streaming pipeline:

```go
df, err := stream.FromCSV("large_file.csv").
    Chunk(10000).          // process N rows at a time
    Pipe(stream.Func(func(chunk *dataframe.DataFrame) (*dataframe.DataFrame, error) {
        return chunk.Filter(expr.Col("status").Eq("active"))
    })).
    Collect()
```

Multiple `Pipe` stages are applied in order to each chunk before the results are concatenated.

---

## I/O Formats

### CSV

```go
// Reading
df, err := io.ReadCSV("data.csv")
df, err := io.ReadCSV("data.tsv",
    io.WithDelimiter('\t'),
    io.WithNullValue("NA"),
    io.WithWorkers(8),
)
df, err := io.ReadCSV("no_header.csv", io.WithNoHeader())

// Writing
err := io.WriteCSV(df, "output.csv")
err := io.WriteCSV(df, "output.tsv", io.WithWriteDelimiter('\t'))
```

Type inference runs per column in parallel: integer → float → bool → string, in that priority order.

### JSON

```go
// Array-of-objects format: [{"col": val, ...}, ...]
df, err := io.ReadJSON("data.json")
err     := io.WriteJSON(df, "output.json")
```

### NDJSON

```go
// One JSON object per line
df, err := io.ReadNDJSON("data.ndjson")
df, err := io.ReadNDJSON("data.ndjson", io.WithNDJSONWorkers(4))
err     := io.WriteNDJSON(df, "output.ndjson")
```

### Parquet

```go
// Apache Arrow-backed Parquet I/O
df, err := io.ReadParquet("data.parquet")
df, err := io.ReadParquet("data.parquet",
    io.WithParquetBatchSize(50000),
    io.WithParquetWorkers(4),
)
err := io.WriteParquet(df, "output.parquet")
```

---

## CLI Tool

The `godas` CLI exposes the most common operations for scripted use.

```
godas [command] [flags]
```

### Commands

#### `load` — Inspect a file
```bash
godas load data.csv
godas load data.json --format json
```

#### `filter` — Filter rows by condition
```bash
godas filter --file data.csv --condition "age > 30"
godas filter --file data.csv --condition "city == London"
godas filter --file data.csv --condition "price < 50.0" --format csv
```

#### `group` — Group and aggregate
```bash
godas group --file data.csv --by region --agg sum:revenue --agg mean:price
godas group --file data.csv --by category --agg count:id --agg max:score
```

Supported aggregation functions: `sum`, `mean`, `min`, `max`, `count`, `std`, `median`, `first`, `last`

#### `stats` — Column statistics
```bash
godas stats revenue --file data.csv

# Output:
# Count:  1000
# Mean:   142.350000
# Std:    38.120000
# Min:    10.000000
# 25%:    115.000000
# Median: 140.000000
# 75%:    168.000000
# Max:    299.000000
# Nulls:  3
```

#### `join` — Join two datasets
```bash
godas join --left customers.csv --right orders.csv --on customer_id --type inner
godas join --left customers.csv --right orders.csv --on id --type left
```

Join types: `inner`, `left`, `right`

#### `export` — Convert between formats
```bash
godas export --file data.csv --output data.json --format json
godas export --file data.json --output data.csv --format csv
```

---

## Testing

The test suite covers all packages with race detection enabled.

```bash
# Run all tests
make test

# Run with coverage report (opens browser)
make cover

# Run a specific package
go test -race -v ./dataframe/...
go test -race -v ./io/...
go test -race -v ./join/...

# Run a single test
go test -run TestGroupByAggregate ./dataframe/...
```

### Test Coverage by Package

| Package | Tests |
|---|---|
| `series` | Construction, null handling, filter, slice, sort, reorder, append, clone |
| `dataframe` | CRUD, filter, select, drop, sort, head/tail/slice, groupby, describe, append |
| `expr` | All comparison ops, string predicates, logical operators, error cases |
| `agg` | All aggregation kinds, alias, non-numeric error path |
| `stats` | Sum, mean, min, max, std, median, count, null count, correlation, describe |
| `io/csv` | Read/write, delimiter, null value, no-header, empty file, missing file |
| `io/json` | Read/write, empty, missing file, mixed types |
| `io/ndjson` | Read/write, empty lines, missing file |
| `io/parquet` | Round-trip for all dtypes, empty write guard, missing file |
| `join` | Inner/left/right, missing key, empty frame |
| `stream` | Basic collect, filter pipe, missing file |
| `pool` | Acquire/release for all three pool types |

---

## Benchmarks

```bash
# Run all benchmarks
make bench

# Run a specific package
go test -bench=. -benchmem ./stats/...
go test -bench=. -benchmem ./series/...
```

```bash
make lint    # golangci-lint (gofmt, govet, staticcheck, errcheck, …)
make test    # go test -race -cover ./...
make bench   # go test -bench=. -benchmem ./...
make build   # go build -o bin/godas ./cmd/godas
```

---

![Godas Logo](godaslogo.png)

<div align="center">

MIT License · © 2026 tc4dy

</div>
