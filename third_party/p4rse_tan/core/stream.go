package core

import "io"

// Import represents a generic byte stream from any source (File, Network, Memory).
type Import struct {
	Name   string
	Reader io.ReadCloser
}

// KeyImport is the universal key for passing byte streams into parsers.
var KeyImport = NewTypedKey[Import]("import_stream")

// Export represents a generic byte payload destined for any output sink (File, Network, DB).
type Export struct {
	Name   string
	Reader io.ReadCloser
}

// KeyExport is the universal key for transferring encoded bytes to output plugins.
var KeyExport = NewTypedKey[Export]("export_stream")
