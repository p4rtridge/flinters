package core

import (
	"testing"
)

func TestContextSetGet(t *testing.T) {
	ctx := make(Context)
	key := DataKey{id: "foo"}

	ctx[key] = 123

	if got := ctx[key]; got != 123 {
		t.Fatalf("expected 123, got %v", got)
	}
}
