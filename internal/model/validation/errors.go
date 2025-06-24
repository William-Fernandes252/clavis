package validation

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/William-Fernandes252/clavis/internal/model/errors"
)

const (
	errType     = "validation"
	defaultCode = "validation-failed"
)

// ValidationError represents a validation-specific error
type ValidationError struct {
	*errors.ErrorData
	Target string `json:"target"`
	Value  any    `json:"value"`
}

// NewValidationError creates a new validation error
func NewValidationError(target string, value any, message string) *ValidationError {
	return &ValidationError{
		Target:    target,
		Value:     value,
		ErrorData: errors.NewErrorData(errType, defaultCode, &message),
	}
}

// WithMetadata adds metadata and returns the validation error
func (ve *ValidationError) WithMetadata(key string, value any) *ValidationError {
	ve.ErrorData.Metadata[key] = value
	return ve
}

// WithCode adds a code and returns the validation error
func (ve *ValidationError) WithCode(code string) *ValidationError {
	ve.ErrorData = ve.ErrorData.WithCode(code)
	return ve
}

// Error implements the error interface for ValidationError
func (ve ValidationError) Error() string {
	if ve.Message != nil {
		return *ve.Message
	}
	return fmt.Sprintf("%s: validation failed for \"%v\"", ve.Target, ve.Value)
}

// Type returns the broad category of error
func (ve ValidationError) Type() errors.ErrorType {
	return ve.ErrorData.Type
}

// Code returns a specific error code for programmatic handling
func (ve ValidationError) Code() string {
	return ve.ErrorData.Code
}

// Metadata returns additional context about the error
func (ve ValidationError) Metadata() map[string]any {
	return ve.ErrorData.Metadata
}

// ToJSON serializes the validation error to JSON
func (ve ValidationError) ToJSON() ([]byte, error) {
	// Create a combined structure for JSON serialization
	combined := struct {
		Type     errors.ErrorType `json:"type"`
		Code     string           `json:"code,omitempty"`
		Metadata map[string]any   `json:"metadata,omitempty"`
		Message  *string          `json:"message,omitempty"`
		Target   string           `json:"target"`
		Value    any              `json:"value"`
	}{
		Type:     ve.ErrorData.Type,
		Code:     ve.ErrorData.Code,
		Metadata: ve.ErrorData.Metadata,
		Message:  ve.Message,
		Target:   ve.Target,
		Value:    ve.Value,
	}

	return json.Marshal(combined)
}

// ValidationResult represents the result of a validation operation
// It contains a list of validation errors
// If there are no errors, it indicates a successful validation
type ValidationResult struct {
	Errors []ValidationError `json:"errors"`
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Errors: make([]ValidationError, 0),
	}
}

// Error implements the error interface for ValidationResult
func (ve ValidationResult) Error() string {
	if len(ve.Errors) == 0 {
		return "validation failed"
	}

	messages := make([]string, len(ve.Errors))
	for i, err := range ve.Errors {
		messages[i] = err.Error()
	}
	return strings.Join(messages, "; ")
}

// Type returns the broad category of error for ValidationResult
func (ve ValidationResult) ToJSON() ([]byte, error) {
	return json.Marshal(ve)
}

// HasErrors checks if the validation result contains any errors
func (ve ValidationResult) HasErrors() bool {
	return len(ve.Errors) > 0
}

// Add appends a new validation error to the result
func (ve *ValidationResult) Add(err ValidationError) {
	ve.Errors = append(ve.Errors, err)
}

var _ errors.Error = (*ValidationError)(nil)
