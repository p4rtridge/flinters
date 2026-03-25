package core

import (
	"errors"
	"testing"

	"github.com/p4rtridge/p4rse_tan/errs"
)

func TestTypedKeySetGetSuccess(t *testing.T) {
	ctx := make(Context)
	key := NewTypedKey[int]("counter")

	key.Set(ctx, 42)

	val, err := key.Get(ctx)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if val != 42 {
		t.Fatalf("expected value 42, got %d", val)
	}
}

func TestTypedKeyGetMissingKey(t *testing.T) {
	ctx := make(Context)
	key := NewTypedKey[string]("missing")

	val, err := key.Get(ctx)
	if err == nil {
		t.Fatal("expected error for missing key")
	}
	if !errors.Is(err, errs.ErrMissingKey) {
		t.Fatalf("expected ErrMissingKey, got %v", err)
	}
	if val != "" {
		t.Fatalf("expected zero string, got %q", val)
	}
}

func TestTypedKeyGetInvalidType(t *testing.T) {
	ctx := make(Context)
	key := NewTypedKey[int]("number")
	ctx[key.Key] = "not-an-int"

	_, err := key.Get(ctx)
	if err == nil {
		t.Fatal("expected invalid type error")
	}
	if !errors.Is(err, errs.ErrInvalidType) {
		t.Fatalf("expected ErrInvalidType, got %v", err)
	}
}
