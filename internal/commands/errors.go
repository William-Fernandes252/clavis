package commands

import (
	"github.com/William-Fernandes252/clavis/internal/errors"
)

const errType = "command"

// NewInvalidArgumentsError creates a new error indicating that the command was called with invalid arguments.
func NewInvalidArgumentsError(message string) *errors.BaseError {
	return &errors.BaseError{
		ErrorData: errors.NewErrorData(errType, "invalid-arguments", message),
	}
}

// NewCommandError creates a new error for command execution failures.
func NewCommandError(code, message string, cause error) *errors.BaseError {
	return &errors.BaseError{
		ErrorData: errors.NewErrorData(errType, code, message),
		Cause:     &cause,
	}
}
