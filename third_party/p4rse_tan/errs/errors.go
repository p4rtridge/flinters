package errs

import (
	"errors"
	"fmt"
)

// Engine Errors
var (
	ErrMissingRequirement = errors.New("missing requirement")
	ErrCircularDependency = errors.New("circular dependency detected")
	ErrPluginExecution    = errors.New("plugin execution failed")
	ErrAsyncExecution     = errors.New("pipeline asynchronous execution failed")
)

// Common Plugin Errors
var (
	ErrMissingKey  = errors.New("missing key in Context")
	ErrInvalidType = errors.New("invalid value type in Context")
)

// MissingKey returns a wrapped ErrMissingKey annotated with the provided identifier.
func MissingKey(key string) error {
	return fmt.Errorf("%w: %s", ErrMissingKey, key)
}

// InvalidType annotates ErrInvalidType with the offending key and the expected vs actual Go types.
func InvalidType(key string, expected, actual any) error {
	return fmt.Errorf("%w: key %s (expected %T, got %T)", ErrInvalidType, key, expected, actual)
}

// CircularDependency annotates ErrCircularDependency with the blocked plugin and requirement.
func CircularDependency(plugin, waiting string) error {
	return fmt.Errorf("%w: cycle involving plugin '%s' waiting for '%s'", ErrCircularDependency, plugin, waiting)
}

// MissingRequirement annotates ErrMissingRequirement with the plugin and missing key.
func MissingRequirement(plugin, requirement string) error {
	return fmt.Errorf("%w: plugin '%s' needs '%s' which is not provided by context or any plugin", ErrMissingRequirement, plugin, requirement)
}

// PluginExecution wraps plugin failures with their name for observability.
func PluginExecution(name string, cause error) error {
	return fmt.Errorf("%w: %s: %w", ErrPluginExecution, name, cause)
}

// AsyncExecution wraps errgroup failures raised after pipeline execution.
func AsyncExecution(cause error) error {
	return fmt.Errorf("%w: %w", ErrAsyncExecution, cause)
}
