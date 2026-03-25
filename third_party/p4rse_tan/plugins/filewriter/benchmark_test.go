package filewriter

import (
	"bufio"
	"io"
	"os"
	"testing"
)

func benchmarkOld(b *testing.B, size int) {
	// Create a dummy file
	f, _ := os.Create("test_old.bin")
	defer os.Remove("test_old.bin")
	defer f.Close()

	data := make([]byte, size)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, w := io.Pipe()
		go func() {
			w.Write(data)
			w.Close()
		}()
		io.Copy(f, r)
		f.Seek(0, 0)
	}
}

func benchmarkNew(b *testing.B, size int) {
	f, _ := os.Create("test_new.bin")
	defer os.Remove("test_new.bin")
	defer f.Close()

	data := make([]byte, size)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r, w := io.Pipe()
		go func() {
			w.Write(data)
			w.Close()
		}()
		// Wrap output in bufio.Writer 8MB
		bw := bufio.NewWriterSize(f, 8*1024*1024)
		io.Copy(bw, r)
		bw.Flush()
		f.Seek(0, 0)
	}
}

func BenchmarkFileWriter_Old(b *testing.B) {
	benchmarkOld(b, 10*1024*1024) // 10MB write
}

func BenchmarkFileWriter_New(b *testing.B) {
	benchmarkNew(b, 10*1024*1024) // 10MB write
}
