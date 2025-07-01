package errors

import (
	"encoding/json"
	"fmt"
)

// ErrorType represents broad categories of errors
type ErrorType string

// Error provides a rich error interface with metadata and JSON serialization
type Error interface {
	error

	// Type returns the broad category of error
	Type() ErrorType

	// Code returns a specific error code for programmatic handling
	Code() string

	// Metadata returns additional context about the error
	Metadata() map[string]any

	// ToJSON serializes the error to JSON
	ToJSON() ([]byte, error)
}

type ErrorData struct {
	Type     ErrorType      `json:"type"`
	Code     string         `json:"code,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
	Message  string         `json:"message,omitempty"`
}

// NewErrorData creates a new ErrorData instance
func NewErrorData(errType ErrorType, code string, message string) *ErrorData {
	return &ErrorData{
		Type:     errType,
		Code:     code,
		Metadata: make(map[string]any),
		Message:  message,
	}
}

// WithMetadata adds metadata to the error
func (e *ErrorData) WithMetadata(key string, value any) *ErrorData {
	if e.Metadata == nil {
		e.Metadata = make(map[string]any)
	}
	e.Metadata[key] = value
	return e
}

// WithCode sets a specific error code
func (e *ErrorData) WithCode(code string) *ErrorData {
	e.Code = code
	return e
}

// ToJSON serializes the error data to JSON
func (e *ErrorData) ToJSON() ([]byte, error) {
	data, err := json.Marshal(e)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize error data: %w", err)
	}
	return data, nil
}
