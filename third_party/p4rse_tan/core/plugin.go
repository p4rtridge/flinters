package core

// Plugin defines the interface for a data-driven logic module.
type Plugin interface {
	// Name returns the unique identifier for the plugin.
	Name() string

	// Requires returns a list of data keys that must be present in the Context
	// before this plugin can execute.
	Requires() []DataKey

	// Provides returns a list of data keys that this plugin guarantees to write
	// into the Context upon successful execution.
	Provides() []DataKey

	// Cleanup runs after all plugins have executed and is used to close any open resources.
	Cleanup(ctx Context) error

	// Execute runs the plugin logic using the provided context.
	// The plugin should write its outputs directly into the context or return an updated context.
	Execute(ctx Context) error
}
