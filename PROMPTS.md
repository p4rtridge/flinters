# PROMPTS.md

This file documents the raw prompts used to guide the development of the **p4rse_tan** project, as required by the `Requirement.md` for the Software Engineer Challenge.

## Prompt 1: Research & Performance Analysis (Initial Phase)

> Can you research the best approach for Go to handle this 1GB CSV file without loading it all into RAM? 
> **Please analyze different ways to read large datasets in Go (e.g., `os.ReadFile` vs `bufio.Scanner` vs `io.Reader`), and provide the pros and cons for each in terms of memory usage and speed.**
> 
> **Additionally, I'd like to see some references to open-source libraries or tools in Go that excel at processing large CSV data. What are the common architectural patterns they use?**
> 
> Also, look into "Iframely" for a data-driven plugin architecture and "OctoSQL" for streaming aggregation patterns. I want to build a minimal version first that can handle the aggregation logic.

---

## Prompt 2: Project Foundation & Architecture (Architecture Design Prompt)

The following instruction set was provided as the primary guiding document for the project's architecture, technical stack, and core philosophy. It follows a data-driven dependency model inspired by Iframely and a high-performance streaming-first approach inspired by OctoSQL.

---

# Project Instructions: p4rse_tan

## Location: third_party/p4rse_tan

## 1. Architectural & Technical References
To ensure consistency, all code logic must reference two primary design inspirations:
- **Architecture (Iframely Style):** Utilize a **Data-Driven Dependency** model. The system operates on a dependency graph. A module (Plugin) only executes when its required input data types (Dependencies) are available.
- **Data Handling (OctoSQL Style):** Focus strictly on **High-Performance Streaming** and **Incremental Aggregation**. Process data line-by-line, never load entire datasets into RAM, and allow computations (e.g., sum, count) on the fly during the reading process.

## 2. Core Philosophy & Identity
- **Name:** p4rse_tan
- **Concept:** A high-performance, streaming-first data processing tool (similar to OctoSQL).
- **Architecture:** Data-driven dependency (inspired by Iframely). Execution logic is driven by "Plugins" based on data availability.
- **Performance Mantra:** "Never load the whole file into memory." RAM usage must remain $O(1)$ (constant) relative to the file size.

## 3. Technical Stack (Go 1.26)
- **Streaming:** - Absolutely NO use of `os.ReadFile` or `ioutil.ReadAll` for data processing.
    - Always use `os.Open` combined with Buffered I/O (`bufio.Scanner` or `bufio.Reader`) for file reading.
- **Modern Go (1.23 - 1.26):** - Prioritize `iter.Seq` and `iter.Seq2` (Iterators) to define and pass data streams.
    - Use `errgroup` for parallel processing and `context` for cancellation/timeouts.
- **Memory Management:** - When using a Scanner, configure `Scanner.Buffer()` with a minimum of 1MB and a maximum of 10MB to handle exceptionally long data lines.
    - Use `sync.Pool` to reuse large buffers and reduce Garbage Collection (GC) pressure.

## 4. Architecture Rules (Type-Safe Plugin System)
- **Plugin Contract:** Every module must adhere to the `core.Plugin` interface. Explicitly define `Requires()` (list of required `DataKey`s) and `Provides()` (list of output `DataKey`s).
- **Type-Safe Context:** The system uses `core.Context` (`map[DataKey]any`).
    - **CRITICAL:** Never read/write to the map using raw strings or manual type casting (`val.(Type)`).
    - **Always** use `core.TypedKey[T]` via `Get(ctx)` and `Set(ctx, val)` to guarantee compile-time Type-Safety.
- **Concurrency & Lifecycle Hooks:** - Plugins should be designed to handle synchronous logic on data streams. The Core Engine (Dispatcher) handles Goroutine management and pipeline concurrency to keep Plugins pure.
- **Dependency Graph:** The Core Engine (Dispatcher) must automatically orchestrate the execution order of Plugins based on the dependency graph formed by their `DataKey`s.

## 5. Coding Standards
- **Error Handling:** Use `errors.Join` or `fmt.Errorf("...: %w", err)` to wrap errors. Never silently ignore errors (`_ = ...`).
- **Interfaces:** Keep interfaces small and focused. Prefer accepting `io.Reader` over passing raw file paths into logic functions.
- **Concurrency:** Ensure thread-safe code. Prevent race conditions when plugins read/write concurrently. Strictly follow the `core.Context` state management.

## 6. Agent Instructions (For Antigravity)
- **No Snippets:** Always provide the complete, runnable code for the requested function or file.
- **Context Awareness:** When manipulating data, always initialize and use `TypedKey[T]` according to the `core` package standards.
- **Streaming Check:** Before providing code, ask yourself: "If the input stream yields 100GB of data, will this crash the memory?". If yes, rewrite it.
- **Explicit Design:** Briefly explain your architectural reasoning (e.g., why this sorting algorithm, why this buffer size) before writing the code.

---

## Prompt 3: Error-handling Strategy (Error Handling Prompt)

> Handling 1GB of "dirty" CSV data requires more than just performance. 
> 
> Can you implement an **Error Aggregation** strategy? I don't want the logs to be flooded with millions of individual row errors. Instead, log a summary warning every 10,000 malformed rows and emit metrics to a separate channel.

---

## Prompt 4: Performance Optimization & Zero-Allocation (Historical Log)

> I want to reach absolute zero-allocation on the CSV scanner because current implementation use `strings.Split` and `strings.TrimSpace` which allocate memory for each strings.
> 
> Can you help me write custom CSV scanner in `./plugins/csv_scanner.go` using byte stream and byte comparison only?
> 
> Also, rewrite `loader` to use `[]byte` instead of `string` for all data keys and values to avoid heap allocation.

---

## Prompt 5: Observability & SOLID Refactoring (Historical Log)

> Refactor current observability system. Engine must not depend on logger package.
> 
> Move engine-internal logging to `slog.Default()` and plugin-level logging into `Context`.
> 
> Refactor `core` into smaller files: `key.go`, `context.go`, `keys.go`, `plugin.go`, `dag.go`, `pipeline.go` to follow SRP.

---
