package core

import "iter"

// Record defines an abstracted accessor for tabular datasets natively without enforcing allocations.
type Record interface {
	Get(key string) ([]byte, bool)
}

// KeyRecordStream is the unified global pipeline dataset channel.
var KeyRecordStream = NewTypedKey[iter.Seq[Record]]("record_stream")
