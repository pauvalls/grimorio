#!/bin/bash
set -euo pipefail

# bench.sh — Run Go benchmarks with optional baseline comparison

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"
BENCH_DIR="${ROOT_DIR}/internal/services"
BASELINE_FILE="${ROOT_DIR}/.benchmark-baseline.txt"

usage() {
    echo "Usage: $0 [run|compare|save-baseline]"
    echo ""
    echo "Commands:"
    echo "  run           Run benchmarks and print results (default)"
    echo "  compare       Run benchmarks and compare against saved baseline"
    echo "  save-baseline Run benchmarks and save results as baseline"
    exit 1
}

CMD="${1:-run}"

case "$CMD" in
    run)
        echo "=== Running benchmarks ==="
        cd "$ROOT_DIR"
        go test ./internal/services/... -bench=. -benchmem -run=^$ -benchtime=1s
        ;;
    compare)
        if [[ ! -f "$BASELINE_FILE" ]]; then
            echo "ERROR: No baseline found at $BASELINE_FILE"
            echo "Run: $0 save-baseline"
            exit 1
        fi
        echo "=== Running benchmarks (comparing against baseline) ==="
        cd "$ROOT_DIR"
        go test ./internal/services/... -bench=. -benchmem -run=^$ -benchtime=1s | tee /tmp/bench-current.txt

        echo ""
        echo "=== Comparison with baseline ==="
        # Simple line-by-line comparison
        if command -v benchstat &> /dev/null; then
            benchstat "$BASELINE_FILE" /tmp/bench-current.txt
        else
            echo "benchstat not installed. Showing raw diff:"
            diff -u "$BASELINE_FILE" /tmp/bench-current.txt || true
        fi
        ;;
    save-baseline)
        echo "=== Saving benchmark baseline ==="
        cd "$ROOT_DIR"
        go test ./internal/services/... -bench=. -benchmem -run=^$ -benchtime=1s | tee "$BASELINE_FILE"
        echo ""
        echo "Baseline saved to $BASELINE_FILE"
        ;;
    *)
        usage
        ;;
esac
