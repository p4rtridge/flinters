# FV-SEC001 - Ad Performance Aggregator

This repository contains the solution for the **Software Engineer Challenge (FV-SEC001)**. It is a high-performance console application designed to process large CSV datasets (~1GB+) with minimal memory footprint using a custom streaming engine.

## 🚀 Performance Highlights

Processing **100,000,000 rows** (~3GB+ of generated data) on a standard laptop:

| Metric | Result | Notes |
| :--- | :--- | :--- |
| **Execution Time** | **~10.6 seconds** | Single-pass streaming architecture. |
| **Peak Memory Usage** | **~13.5 MB** | Constant $O(1)$ memory usage, regardless of file size. |

*(Benchmark performed on Intel i5-11400H, 8GB RAM, WSL2 Ubuntu)*

---

## 🛠 Prerequisites

- **Go 1.26** or higher (Required for generic iterator support).
- Make sure to have `make` installed (optional, for easier commands).

## 📥 Installation

1.  **Clone the repository:**
    ```bash
    git clone https://github.com/p4rtridge/flinters.git
    cd flinters
    ```

2.  **Download Input Data:**
    Place your `ad_data.csv` file in the `data/` directory.

3.  **Build the application:**
    ```bash
    go build -o aggregator aggregator.go
    ```

##  ▶️ Usage

Run the aggregator via CLI:

```bash
./aggregator --input data/ad_data.csv --results results/
```

### Flags

| Flag | Default | Description |
| :--- | :--- | :--- |
| `--input` | `data/ad_data.csv` | Path to the source CSV file. |
| `--results` | `results/` | Directory where `top_10_ctr.csv` and `top_10_cpa.csv` will be saved. |

---

## 🏗 Architecture & Design Decisions

### Why Custom Streaming Engine (`p4rse_tan`)?
The challenge requires handling large files efficiently. Standard approaches often fail at scale:
 - **Load-all-to-memory:** Causes OOM crashes on large files.
 - **Naive Line-by-Line:** Often results in "spaghetti code" when validation, aggregation, and filtering logic mix.

We implemented **p4rse_tan**, a Data-Driven DAG Framework (inspired by *Iframely* and *OctoSQL*). 
- **System:** Plugins declare *Dependencies* (Inputs) and *Provisions* (Outputs).
- **Core:** The engine automatically resolves the execution order (Topological Sort).
- **Memory:** Zero-allocation hot paths using `sync.Pool` and Go 1.23+ Iterators.

### Project Structure

```
.
├── aggregator.go          # Main CLI entry point (Wires plugins together)
├── campaign/              # Challenge-specific logic
│   ├── models/            # Domain models (Campaign, Stats)
│   └── plugins/           # Business Logic Plugins
│       ├── ctr/           # Calculates CTR & finds Top 10
│       ├── cpa/           # Calculates CPA & finds Top 10
│       └── writer/        # Writes results to CSV
├── third_party/
│   └── p4rse_tan/         # The custom streaming engine Core
└── results/               # Output directory
```

---

## 🧪 Benchmark & Verification

To verify performance claims, a benchmark script is included in the engine directory.

```bash
# Generate 100M rows and benchmark
cd third_party/p4rse_tan/benchmark
./run_benchmark.sh 100000000
```

*See `third_party/p4rse_tan/README.md` for detailed comparisons against OctoSQL.*

---

## 📚 Libraries Used

- **Standard Library:** `encoding/csv`, `bufio`, `sort`, `flag`, `context`.
- **p4rse_tan (Internal):** Custom streaming framework for dependency injection and pipeline management.
- **github.com/p4rtridge/p4rse_tan:** (Local replace in `go.mod`).

##  License
MIT
