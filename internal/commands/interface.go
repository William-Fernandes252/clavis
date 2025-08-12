package commands

import "github.com/William-Fernandes252/clavis/internal/errors"

// Command defines the interface for a command that can be executed.
type Command[A any, R any] interface {
	// Execute runs the command with the provided arguments.
	Execute(args A) (*Output[R], errors.Error)

	// Name returns the name of the command.
	Name() string
}

type Output[T any] struct {
	// Value holds the result of the command execution.
	Value T `json:"value"`

	// Metadata contains additional information about the command execution.
	Metadata map[string]any `json:"metadata,omitempty"`
}

func (o *Output[T]) WithMetadata(key string, value any) *Output[T] {
	if o.Metadata == nil {
		o.Metadata = make(map[string]any)
	}
	o.Metadata[key] = value
	return o
}

// Exit creates an Output instance with the provided value.
func Exit[T any](value T) *Output[T] {
	return &Output[T]{Value: value}
}
