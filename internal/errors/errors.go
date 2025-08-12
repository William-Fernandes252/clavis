package errors

import (
	"encoding/json"
	"fmt"
)

type BaseError struct {
	*ErrorData
	Cause *error `json:"cause,omitempty"`
}

func (e *BaseError) Code() string {
	return e.ErrorData.Code
}

func (e *BaseError) Message() string {
	return e.ErrorData.Message
}

func (e *BaseError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %v", e.ErrorData.Message, e.Cause)
	}
	return e.ErrorData.Message
}

func (e *BaseError) Type() ErrorType {
	return e.ErrorData.Type
}

func (e *BaseError) Metadata() map[string]any {
	return e.ErrorData.Metadata
}

func (e *BaseError) WithMetadata(key string, value any) *BaseError {
	if e.ErrorData.Metadata == nil {
		e.ErrorData.Metadata = make(map[string]any)
	}
	e.ErrorData.Metadata[key] = value
	return e
}

func (e *BaseError) ToJSON() ([]byte, error) {
	data := struct {
		Type     string         `json:"type"`
		Code     string         `json:"code"`
		Message  string         `json:"message"`
		Metadata map[string]any `json:"metadata,omitempty"`
		Cause    string         `json:"cause,omitempty"`
	}{
		Type:     string(e.Type()),
		Code:     e.Code(),
		Message:  e.Message(),
		Metadata: e.Metadata(),
	}

	if e.Cause != nil {
		data.Cause = (*e.Cause).Error()
	}

	return json.Marshal(data)
}

var _ Error = (*BaseError)(nil)

const (
	ValidationErrType          = "validation"
	DefaultValidationErrorCode = "validation-failed"
)

// ValidationError represents a validation-specific error
type ValidationError struct {
	*ErrorData
	Target string `json:"target"`
	Value  any    `json:"value"`
}

// NewValidationError creates a new validation error
func NewValidationError(target string, value any, message string) *ValidationError {
	return &ValidationError{
		Target:    target,
		Value:     value,
		ErrorData: NewErrorData(ValidationErrType, DefaultValidationErrorCode, message),
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
	return fmt.Sprintf("%s: %s", ve.Target, ve.Message)
}

// Type returns the broad category of error
func (ve ValidationError) Type() ErrorType {
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
		Type     ErrorType      `json:"type"`
		Code     string         `json:"code,omitempty"`
		Metadata map[string]any `json:"metadata,omitempty"`
		Message  string         `json:"message,omitempty"`
		Target   string         `json:"target"`
		Value    any            `json:"value"`
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

var _ Error = (*ValidationError)(nil)
