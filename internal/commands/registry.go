package commands

import (
	"encoding/json"
	"fmt"

	"github.com/William-Fernandes252/clavis/internal/errors"
)

// Registry holds all the registered command factories and handles their execution.
type Registry struct {
	factories map[string]CommandFactory[any]
}

// NewRegistry creates a new command registry.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]CommandFactory[any]),
	}
}

// Register adds a command factory to the registry.
// It will panic if a command with the same name is already registered.
func (r *Registry) Register(factory CommandFactory[any]) {
	name := factory.Name()
	if _, exists := r.factories[name]; exists {
		panic(fmt.Sprintf("command %s already registered", name))
	}
	r.factories[name] = factory
}

// ExecuteCommand finds a command factory by name, creates the command with arguments, and executes it.
// args are passed as raw JSON to be decoded into the specific command's argument struct.
func (r *Registry) ExecuteCommand(name string, rawArgs json.RawMessage) (*Output[any], errors.Error) {
	factory, ok := r.factories[name]
	if !ok {
		return nil, NewCommandError("command-not-found", fmt.Sprintf("command '%s' not found", name), nil)
	}

	// Create the command with the provided arguments
	cmd, err := factory.Create(rawArgs)
	if err != nil {
		return nil, err
	}

	// Execute the command using the Command interface
	return cmd.Execute()
}
