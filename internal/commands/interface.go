package commands

import (
	"encoding/json"

	"github.com/William-Fernandes252/clavis/internal/errors"
)

// Command defines the interface for a command that can be executed.
type Command[T any] interface {
	// Execute runs the command and returns a generic output.
	Execute() (*Output[T], errors.Error)

	// Name returns the name of the command.
	Name() string
}

// CommandFactory defines the interface for creating commands with arguments.
// This properly implements the Abstract Factory pattern by returning the Command interface.
type CommandFactory[T any] interface {
	// Name returns the name of the command this factory creates.
	Name() string

	// Create creates a new command instance with the provided arguments.
	Create(rawArgs json.RawMessage) (Command[T], errors.Error)
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
