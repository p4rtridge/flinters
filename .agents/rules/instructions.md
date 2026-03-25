---
trigger: always_on
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

## 7. Performance Tuning & Profiling (Optimization Phase)
- **Benchmarking First:** Any refactoring proposal related to performance (e.g., changing string parsing logic, altering buffer structures) MUST be accompanied by `Benchmark...` code using the `testing` package.
- **Memory Profiling:** The system must be designed to easily integrate `net/http/pprof` or `runtime/pprof`. When debugging memory leaks, the agent must focus on finding hanging Goroutines (goroutine leaks) or slice/map references to old data that aren't being garbage collected.
- **Escape Analysis:** When refactoring hot paths (loops processing large amounts of data), prioritize writing code so that variables are allocated on the Stack rather than the Heap (to reduce GC pressure). Avoid returning pointers (`*T`) in internal stream processing functions if the struct is small.

## 8. Advanced Memory Management (Zero-Allocation Goal)
- **Aggressive `sync.Pool`:** Mandate the use of `sync.Pool` for ALL frequent allocations occurring during streaming (e.g., `[]byte` buffers for individual file lines, or intermediate `DataBundle` structs).
- **String vs []byte:** During intermediate string processing steps, absolutely do not cast `[]byte` to `string` if it is only for comparison or regex matching. Use the functions in the `bytes` package to operate directly on `[]byte` to avoid zero-allocation overhead.

## 9. Refactoring Standards
- **Cyclomatic Complexity:** Issue a warning and require function extraction if a Plugin's `Execute()` method exceeds 50 lines or has too many nested loops/conditions.
- **Defensive Programming:** During refactoring, always add safety checks (nil pointer checks, out-of-bounds slice checks) so the tool doesn't panic when encountering "dirty" data.
- **Test Preservation:** No refactoring step is allowed to alter the output of existing Unit Tests.

## 10. Observability & Logging
- **Standard Library:** Exclusively use Go's built-in `log/slog` for structured logging. Avoid third-party logging frameworks to keep the core binary lightweight and dependency-free.
- **Context Injection:** The Logger must be injected into the pipeline via the Type-Safe Context.
    - Define `var KeyLogger = NewTypedKey[*slog.Logger]("engine_logger")` in the `core` package.
    - Plugins must retrieve this logger via `KeyLogger.Get(ctx)` instead of using the global `slog.Info()`.
- **Structured & Contextual:** Always bind the plugin's identity to its logs to make debugging trace graphs easier: 
    - e.g., `log := baseLog.With(slog.String("plugin", p.Name()))`
- **Zero-Cost Hot Paths (CRITICAL):**
    - NEVER execute logging statements (`Info`, `Debug`, `Error`) inside a high-frequency streaming loop (e.g., line-by-line processing). I/O operations will destroy the $O(1)$ streaming performance.
    - If row-level error tracking is required, use an error aggregation strategy (e.g., log a summary warning every 10,000 dropped rows) or emit metrics to a separate channel.