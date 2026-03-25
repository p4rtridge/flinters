package csvwriter

import (
	"bufio"
	"encoding/csv"
	"io"
	"sync"
	"testing"
)

func benchmarkOld(b *testing.B, rows int) {
	header := []string{"campaign_id", "clicks", "spend"}
	row := []string{"camp_123", "100.5", "50.25"}

	for i := 0; i < b.N; i++ {
		pr, pw := io.Pipe()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			io.Copy(io.Discard, pr)
		}()

		writer := csv.NewWriter(pw)
		writer.Write(header)

		for r := 0; r < rows; r++ {
			writer.Write(row)
		}
		writer.Flush() 
		pw.Close()
		wg.Wait()
	}
}

func benchmarkNew(b *testing.B, rows int) {
	header := []string{"campaign_id", "clicks", "spend"}
	row := []string{"camp_123", "100.5", "50.25"}

	for i := 0; i < b.N; i++ {
		pr, pw := io.Pipe()
		var wg sync.WaitGroup
		wg.Add(1)
		go func() {
			defer wg.Done()
			io.Copy(io.Discard, pr)
		}()

		bw := bufio.NewWriterSize(pw, 1024*1024) // 1MB buffer
		writer := csv.NewWriter(bw)
		writer.Write(header)

		for r := 0; r < rows; r++ {
			writer.Write(row)
		}
		writer.Flush()
		bw.Flush()
		pw.Close()
		wg.Wait()
	}
}

func BenchmarkCSVWriter_FlushLoop(b *testing.B) {
	benchmarkOld(b, 1000)
}

func BenchmarkCSVWriter_FlushEnd(b *testing.B) {
	benchmarkNew(b, 1000)
}
