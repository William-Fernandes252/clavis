package storage

import "github.com/William-Fernandes252/clavis/internal/errors"

// NewStorageError creates a new storage error with the given code, message, and cause
func NewStorageError(code, message string, cause error) *errors.BaseError {
	return &errors.BaseError{
		ErrorData: errors.NewErrorData("storage", code, message),
		Cause:     &cause,
	}
}
