package commands

import (
	"encoding/json"
	"testing"

	"github.com/William-Fernandes252/clavis/internal/errors"
)

// Mock command for testing
type MockCommand struct {
	name  string
	value any
}

func (m *MockCommand) Execute() (*Output[any], errors.Error) {
	return Exit[any]("mock-result"), nil
}

func (m *MockCommand) Name() string {
	return m.name
}

// Mock factory for testing
type MockCommandFactory struct{}

func (f *MockCommandFactory) Name() string {
	return "mock"
}

func (f *MockCommandFactory) Create(rawArgs json.RawMessage) (Command[any], errors.Error) {
	return &MockCommand{name: "mock", value: "test"}, nil
}

func TestRegistry_ExecuteCommand_ReturnType(t *testing.T) {
	registry := NewRegistry()
	factory := &MockCommandFactory{}
	registry.Register(factory)

	// Test ExecuteCommand returns Output[any]
	args, _ := json.Marshal(map[string]string{"key": "test-key"})
	result, err := registry.ExecuteCommand("mock", args)
	if err != nil {
		t.Fatalf("ExecuteCommand failed: %v", err)
	}

	if result == nil {
		t.Fatal("Expected result to be non-nil")
	}

	// Check that the result is an Output[any] with the correct value
	if result.Value != "mock-result" {
		t.Errorf("Expected value 'mock-result', got %v", result.Value)
	}

	// Verify the returned type is *Output[any]
	var _ *Output[any] = result

	// Test with non-existent command
	_, err = registry.ExecuteCommand("non-existent", args)
	if err == nil {
		t.Error("Expected error for non-existent command")
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	factory := &MockCommandFactory{}

	// Register should work normally
	registry.Register(factory)

	// Registering the same command twice should panic
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when registering duplicate command")
		}
	}()
	registry.Register(factory)
}
