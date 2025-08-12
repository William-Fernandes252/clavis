package keys

import (
	"fmt"

	"github.com/William-Fernandes252/clavis/internal/errors"
)

const (
	KeyErrorType    = "keys"
	KeyNotFoundCode = "key-not-found"
	InvalidKeyCode  = "invalid-key"
)

func NewKeyNotFoundError(key Key, cause error) *errors.BaseError {
	return NewKeyError(KeyNotFoundCode, fmt.Sprintf("%s not found", key), cause).
		WithMetadata("key", key)
}

func NewInvalidKeyError(key Key, cause error) *errors.BaseError {
	return NewKeyError(InvalidKeyCode, fmt.Sprintf("%s is not a valid key", key), cause).
		WithMetadata("key", key)
}

func NewKeyError(code, message string, cause error) *errors.BaseError {
	return &errors.BaseError{
		ErrorData: errors.NewErrorData(KeyErrorType, code, message),
		Cause:     &cause,
	}
}
