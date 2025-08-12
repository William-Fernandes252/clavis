package keys

import (
	"regexp"

	"github.com/William-Fernandes252/clavis/internal/errors"
)

type Key string

const (
	keyMaxLength = 255
	keyPattern   = `^[a-zA-Z][a-zA-Z0-9._:-]*$`
)

// Validate checks if the key meets the required constraints:
// - Cannot be empty
// - Cannot exceed 255 characters in length
// - Must match the pattern ^[a-zA-Z][a-zA-Z0-9._:-]*$
func (k Key) Validate() *errors.ValidationError {
	if len(k) == 0 {
		return errors.NewValidationError("key", k, "cannot be empty").WithCode("key-empty")
	}
	if len(k) > keyMaxLength {
		return errors.NewValidationError("key", k, "cannot exceed 255 characters").
			WithCode("key-too-long").
			WithMetadata("max-length", keyMaxLength)
	}
	if !regexp.MustCompile(keyPattern).MatchString(string(k)) {
		return errors.NewValidationError("key", k, "must match the pattern ^[a-zA-Z][a-zA-Z0-9._:-]*$").
			WithCode("key-invalid").
			WithMetadata("pattern", keyPattern)
	}
	return nil
}

// String returns the string representation of the Key.
func (k Key) String() string {
	return string(k)
}

// Bytes returns the byte representation of the Key.
func (k Key) Bytes() []byte {
	return []byte(k)
}
