package core

import "github.com/p4rtridge/p4rse_tan/errs"

// DataKey is an encapsulated unique identifier for data within the Context.
// It is intentionally a struct to prevent arbitrary string casting and encourage
// the usage of Type-Safe keys declared by plugins.
type DataKey struct {
	id string
}

// UntypedKey returns the exact underlying string identifier. Used minimally for debugging.
func (k DataKey) UntypedKey() string { return k.id }

// TypedKey provides a way to interact with Context in a type-safe manner.
type TypedKey[T any] struct {
	Key DataKey
}

func NewTypedKey[T any](id string) TypedKey[T] {
	return TypedKey[T]{Key: DataKey{id: id}}
}

func (k TypedKey[T]) Get(ctx Context) (T, error) {
	var zero T

	if ctx == nil {
		return zero, errs.MissingKey(k.Key.id)
	}

	val, ok := ctx[k.Key]
	if !ok {
		return zero, errs.MissingKey(k.Key.id)
	}

	typedVal, ok := val.(T)
	if !ok {
		return zero, errs.InvalidType(k.Key.id, zero, val)
	}

	return typedVal, nil
}

func (k TypedKey[T]) Set(ctx Context, val T) {
	ctx[k.Key] = val
}
