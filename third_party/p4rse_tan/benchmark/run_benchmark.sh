#!/bin/bash
set -e

# Data size (override with $1)
ROWS=${1:-100000000}
DATA_FILE="benchmark_data.csv"

echo "==================================================="
echo "     FULL APPLICATION BENCHMARK (TIME & MEMORY)    "
echo "==================================================="

# 1. Build the p4rse_tan
echo "[1/4] Building p4rse_tan..."
go build -o p4rse_tan ./main.go

# 2. Generate Data
echo "[2/4] Preparing test data ($ROWS rows)..."
if [ ! -f "$DATA_FILE" ]; then
    echo " -> Generating new data file: $DATA_FILE"
    ./p4rse_tan -mode=gen -rows=$ROWS
else
    echo " -> Data file $DATA_FILE exists."
fi

# ---------------------------------------------------
# 3. MEMORY BENCHMARK (Peak RAM / RSS)
# ---------------------------------------------------
echo "---------------------------------------------------"
echo "[3/4] Running MEMORY Benchmark (Peak RAM usage)..."

# Hàm phụ để lấy Peak RAM
get_peak_ram() {
    local cmd="$1"
    # Dọn cache trước khi đo RAM để kết quả chính xác
    sync; echo 3 | sudo tee /proc/sys/vm/drop_caches > /dev/null
    
    # Chạy lệnh và lọc ra dòng "Maximum resident set size"
    # Dùng awk để chuyển đổi KB sang MB
    /usr/bin/time -v bash -c "$cmd" 2>&1 | \
    grep "Maximum resident set size" | \
    awk -F': ' '{printf "%.2f MB\n", $2/1024}'
}

if [ -f /usr/bin/time ]; then
    echo -n " -> p4rse_tan Peak RAM: "
    get_peak_ram "./p4rse_tan -mode=run_p4rse > /dev/null"

    if command -v octosql &> /dev/null; then
        echo -n " -> octosql Peak RAM:   "
        get_peak_ram "octosql \"SELECT region, SUM(amount) FROM \\\"$DATA_FILE\\\" GROUP BY region\" > /dev/null"
    fi
else
    echo " -> /usr/bin/time not found. Install it via: sudo apt install time"
fi

# ---------------------------------------------------
# 4. EXECUTION TIME BENCHMARK (Hyperfine)
# ---------------------------------------------------
echo "---------------------------------------------------"
echo "[4/4] Running EXECUTION TIME Benchmark (Hyperfine)..."

if command -v hyperfine &> /dev/null; then
    if command -v octosql &> /dev/null; then
        hyperfine --warmup 1 \
          --runs 5 \
          --prepare 'sync; echo 3 | sudo tee /proc/sys/vm/drop_caches > /dev/null' \
          './p4rse_tan -mode=run_p4rse' \
          "octosql \"SELECT region, SUM(amount) FROM \\\"$DATA_FILE\\\" GROUP BY region\""
    else
        echo " -> octosql not found. Benchmarking p4rse_tan only..."
        hyperfine --warmup 1 \
          --runs 5 \
          --prepare 'sync; echo 3 | sudo tee /proc/sys/vm/drop_caches > /dev/null' \
          './p4rse_tan -mode=run_p4rse'
    fi
else
    echo " -> hyperfine not found! Install via: sudo apt install hyperfine"
    echo " -> Falling back to basic execution time..."
    time ./p4rse_tan -mode=run_p4rse
fi

echo "==================================================="
echo "                  BENCHMARK DONE                   "
echo "==================================================="