# p4rse_tan

p4rse_tan is a high-performance, streaming-first data processing framework for Go 1.26+, inspired by **Iframely** (architecture) and **OctoSQL** (data handling).

It connects modular **Plugins** using a **Data-Driven Dependency Graph**. The execution order is automatically resolved: a plugin runs only when its required data dependencies are available in the shared `Context`.

## Performance Benchmark

We include a benchmark script to compare `p4rse_tan` against standard tools like `octosql`.

### Setup
1.  **Install OctoSQL** (optional, for comparison):
    ```bash
    go install github.com/cube2222/octosql/cmd/octosql@latest
    ```
2.  **Run Benchmark:**
    ```bash
    cd third_party/p4rse_tan/benchmark
    ./run_benchmark.sh [row_count]
    ```
    Default row count is 100,000,000.

### Results (100M rows, CSV Group By Sum)

**Test Environment:**
- **OS:** Ubuntu 24.04.4 LTS (WSL2 on Windows 11)
- **CPU:** 11th Gen Intel i5-11400H (12 Cores) @ 2.687GHz
- **Memory:** 8GB (7829MiB)
- **Go Version:** 1.26.1 `linux/amd64`

| Tool | Time (Mean) | Peak Memory | Notes |
| :--- | :--- | :--- | :--- |
| **p4rse_tan** | **~10.68s** | **~13.5 MB** | Pure streaming, constant memory (O(1)). |
| **OctoSQL** | ~37.28s | ~119.7 MB | ~3.5x slower, ~9x higher memory usage. |

*(Benchmark run on 2026-03-25)*

## Why Streaming?

This project evolved from an initial attempt to process data by loading entire files into **SQLite** for querying. While valid for smaller datasets, that approach encountered significant bottlenecks with large-scale data:

- **High Memory Overhead:** Loading huge files or bulk-inserting them often leads to OOM (Out of Memory) crashes on constrained hardware.
- **Latency:** You must wait for the entire import process to finish before any analysis can begin.

**p4rse_tan** solves this by adoptiong a pure **Streaming Architecture**:
- **Instant Start:** Processing begins immediately as the first byte is read.
- **Constant Memory:** RAM usage stays low ($O(1)$) regardless of input size (whether 100MB or 100TB).
- **Zero Overhead:** No intermediate database storage is required; data flows directly from source to aggregation.

## Why Data-Driven Dependency?

The choice of a **Data-Driven Dependency** model (over a fixed procedural pipeline) is deliberate to maximize **Extensibility** and **Flexibility**:

- **Decoupled Plugins:** Users can write custom plugins that only know about their inputs and output. They don't need to know *who* provides the data or *who* consumes it.
- **Easy Extension:** Adding a new feature (e.g., a new metric or a different output format) is as simple as registering a new struct. The engine automatically rewires the execution graph.
- **Hot-Swappable Logic:** You can swap out a `CSVReader` for a `JSONReader` or a `PostgresReader` without changing a single line of code in analysis plugins, as long as they provide the same data keys (`iter.Seq`).

## Other Architectures Considered & Why They Failed

Before settling on the Data-Driven DAG, we evaluated other common Go patterns but found them lacking for our specific "extensible plugin" requirements:

1.  **Pure Function Chaining (Procedural):**
    *   *Approach:* Calling `Process(Read(file))`.
    *   *Issue:* Extremely rigid. Adding a new step (like "filter by date") requires rewriting the main function. It breaks the "Plugin" concept where users can just drop in new logic.

2.  **Manual Channel Orchestration (Standard Go Concurrency):**
    *   *Approach:* Manually wiring `chan string` between goroutines.
    *   *Issue:* While performant, managing shutdown signals, error propagation, and buffer sizes across 10+ stages becomes unmaintainable "spaghetti code." The **Data-Driven DAG** automates this safely.

3.  **Full-Scale ETL Engines (e.g., Apache Beam / Spark):**
    *   *Approach:* Using distributed systems.
    *   *Issue:* Overkill for a single-binary Go application. We needed embedded high-performance, not cluster management overhead.

## Core Philosophy

1.  **Streaming First:** Never load entire datasets into RAM. RAM usage must remain $O(1)$ relative to the file size.
2.  **Data-Driven Architecture:** You don't manually order function calls. You define *what data you need* and *what data you produce*. The engine builds the DAG (Directed Acyclic Graph) for you.
3.  **Type Safety:** The `Context` uses `core.TypedKey[T]` to ensure strict compile-time type checking, avoiding raw string keys and unsafe type assertions.

## Features

- **Automatic Dependency Resolution:** The `BuildPipeline` function performs a topological sort of your plugins based on `Requires()` and `Provides()`.
- **Zero-Allocation Focus:** Encourages the use of `sync.Pool` and Go 1.23+ Iterators (`iter.Seq`) for high-throughput data streams.
- **Structured Error Handling:** Built-in error wrapping and context management.

## Installation

```bash
go get github.com/p4rtridge/p4rse_tan
```

## Usage

### 1. Define a Plugin

Implement the `core.Plugin` interface.

```go
package main

import (
    "fmt"
    "github.com/p4rtridge/p4rse_tan/core"
)

// Define Type-Safe Keys
var (
    InputFileKey   = core.NewTypedKey[string]("config.input_file")
    RecordCountKey = core.NewTypedKey[int]("metrics.record_count")
)

type CounterPlugin struct{}

func (p *CounterPlugin) Name() string {
    return "CounterPlugin"
}

func (p *CounterPlugin) Requires() []core.DataKey {
    // This plugin needs an input filename to start
    return []core.DataKey{InputFileKey.Key}
}

func (p *CounterPlugin) Provides() []core.DataKey {
    // This plugin will produce a record count
    return []core.DataKey{RecordCountKey.Key} 
}

func (p *CounterPlugin) Execute(ctx core.Context) error {
    // 1. Get Input (safely typed)
    inputFile, err := InputFileKey.Get(ctx)
    if err != nil {
        return err // Handle missing dependency errors
    }

    fmt.Printf("Processing %s...\n", inputFile)

    // ... (streaming logic would go here) ...
    count := 42 

    // 2. Set Output
    RecordCountKey.Set(ctx, count)
    
    return nil
}

func (p *CounterPlugin) Cleanup(ctx core.Context) error {
    return nil // Close files or resources here
}
```

### 2. Build and Run a Pipeline

```go
package main

import (
    "github.com/p4rtridge/p4rse_tan/core"
)

func main() {
    // 1. Initialize Context
    ctx := make(core.Context)
    execCtx := context.Background()
    
    // 2. Seed initial data (e.g., config / CLI args)
    InputFileKey.Set(ctx, "data.csv")

    // 3. Register Plugins
    plugins := []core.Plugin{
        &CounterPlugin{},
        // Add other plugins (e.g., CSVReader, Aggregator, SQLWriter) here.
        // The order in this list DOES NOT matter.
    }

    // 4. Build Pipeline (Resolves DAG)
    pipeline, err := core.BuildPipeline(ctx, plugins)
    if err != nil {
        panic(err)
    }

    // 5. Execute
    if err := pipeline.Execute(execCtx, ctx); err != nil {
        panic(err)
    }
}
```

## Architecture

### Context & Keys
State is passed via specific generic helpers to ensure safety: 
- `core.NewTypedKey[T](id string)`: Creates a typed key definition.
- `key.Get(ctx)`: Retrieves a value of type `T`.
- `key.Set(ctx, val)`: Sets a value of type `T`.

### Plugins
Each plugin is an isolated unit of logic. It should not know about other plugins, only about the **Data** it needs.

## Directory Structure

- `core/`: The heart of the framework (DAG, Pipeline, Context).
- `plugins/`: Standalone implementations (CSV Reader, Writer, Aggregator).
- `errs/`: Standardized error handling.
- `logger/`: structed logging interface.

## Requirements

- Go 1.26 or higher.
