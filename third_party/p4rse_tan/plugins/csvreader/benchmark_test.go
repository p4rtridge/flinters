package csvreader

import (
	"bufio"
	"encoding/csv"
	"io"
	"strings"
	"testing"
)

// benchmarkOld: encoding/csv with per-row []string allocation
func benchmarkOld(b *testing.B, csvData string) {
	reader := csv.NewReader(strings.NewReader(csvData))
	headerRecord, _ := reader.Read()
	headerMap := make(map[string]int)
	for i, col := range headerRecord {
		headerMap[strings.TrimSpace(col)] = i
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		reader = csv.NewReader(strings.NewReader(csvData))
		reader.Read() // skip header
		for {
			rec, err := reader.Read()
			if err == io.EOF {
				break
			}
			// convert []string → [][]byte to match current csvRecord type
			bRow := make([][]byte, len(rec))
			for j, v := range rec {
				bRow[j] = []byte(v)
			}
			wrapper := &csvRecord{header: headerMap, row: bRow}
			_ = wrapper
		}
	}
}

// benchmarkNew: custom bufio.Scanner + pre-allocated [][]byte — zero string heap allocations
func benchmarkNew(b *testing.B, csvData string) {
	// parse header once
	firstScanner := bufio.NewScanner(strings.NewReader(csvData))
	firstScanner.Scan()
	headerRecord := strings.Split(string(firstScanner.Bytes()), ",")
	headerMap := make(map[string]int)
	for i, col := range headerRecord {
		headerMap[strings.TrimSpace(col)] = i
	}

	cols := make([][]byte, len(headerRecord))
	wrapper := &csvRecord{header: headerMap, row: cols}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		scanner := bufio.NewScanner(strings.NewReader(csvData))
		scanner.Scan() // skip header
		for scanner.Scan() {
			line := scanner.Bytes()
			idx := 0
			start := 0
			n := len(line)
			for j := 0; j < n; j++ {
				if line[j] == ',' {
					if idx < len(cols) {
						cols[idx] = line[start:j]
					}
					start = j + 1
					idx++
				}
			}
			if idx < len(cols) {
				cols[idx] = line[start:n]
			}
			for k := idx + 1; k < len(cols); k++ {
				cols[k] = nil
			}
			_ = wrapper
		}
	}
}

func generateCSV(rows int) string {
	var sb strings.Builder
	sb.WriteString("col1,col2,col3\n")
	for i := 0; i < rows; i++ {
		sb.WriteString("val1,val2,val3\n")
	}
	return sb.String()
}

func BenchmarkCSVReader_Old(b *testing.B) {
	data := generateCSV(1000)
	benchmarkOld(b, data)
}

func BenchmarkCSVReader_New(b *testing.B) {
	data := generateCSV(1000)
	benchmarkNew(b, data)
}
