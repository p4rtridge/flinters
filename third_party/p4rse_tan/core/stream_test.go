package core

import (
	"io"
	"strings"
	"testing"

	"iter"
)

func TestKeyRecordStreamRoundTrip(t *testing.T) {
	ctx := make(Context)
	seq := iter.Seq[Record](func(yield func(Record) bool) {})
	KeyRecordStream.Set(ctx, seq)

	got, err := KeyRecordStream.Get(ctx)
	if err != nil {
		t.Fatalf("expected stream, got error %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil record stream")
	}
}

func TestKeyImportRoundTrip(t *testing.T) {
	ctx := make(Context)
	expected := Import{
		Name:   "input.csv",
		Reader: io.NopCloser(strings.NewReader("data")),
	}
	KeyImport.Set(ctx, expected)

	got, err := KeyImport.Get(ctx)
	if err != nil {
		t.Fatalf("expected import, got error %v", err)
	}
	if got.Name != expected.Name {
		t.Fatalf("expected name %s, got %s", expected.Name, got.Name)
	}
	if got.Reader == nil {
		t.Fatal("expected reader to be non-nil")
	}
	got.Reader.Close()
}

func TestKeyExportRoundTrip(t *testing.T) {
	ctx := make(Context)
	expected := Export{
		Name:   "output.csv",
		Reader: io.NopCloser(strings.NewReader("rows")),
	}
	KeyExport.Set(ctx, expected)

	got, err := KeyExport.Get(ctx)
	if err != nil {
		t.Fatalf("expected export, got error %v", err)
	}
	if got.Name != expected.Name {
		t.Fatalf("expected name %s, got %s", expected.Name, got.Name)
	}
	if got.Reader == nil {
		t.Fatal("expected reader to be non-nil")
	}
	got.Reader.Close()
}
