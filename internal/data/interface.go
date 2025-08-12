package data

import (
	"github.com/William-Fernandes252/clavis/internal/errors"
)

const dataErrorType = "data"

type Type string

// Type represents the type of data stored in the system.
func (t Type) String() string {
	return string(t)
}

// Serializer is an interface for data types that can be serialized to bytes
type Serializer[T any] interface {
	// Serialize converts the data into a byte slice for storage
	//
	// e.g. []byte("hello"), JSON marshal for maps, gob for structs, etc.
	Serialize(value T) ([]byte, errors.Error)
}

// Deserializer is an interface for data types that can be deserialized from bytes
type Deserializer[T any] interface {
	// Deserialize reads the bytes back into the data type
	//
	// e.g. []byte("hello") ? "hello", JSON?unmarshal for maps, gob for structs, etc.
	Deserialize(raw []byte) (T, errors.Error)
}

// Codec is an interface that combines serialization and deserialization.
// It defines how to convert data to and from a byte slice.
// It also provides the type of data it handles (e.g. "string", "hash", "list").
// T is the type of data being serialized/deserialized (e.g. string, map[string]string, []string, etc.)
type Codec[T any] interface {
	Serializer[T]
	Deserializer[T]

	// Type returns the type name ("string", "hash", "list")
	Type() Type
}

func NewDataError(code, message string, cause error) *errors.BaseError {
	return &errors.BaseError{
		ErrorData: errors.NewErrorData(dataErrorType, code, message),
		Cause:     &cause,
	}
}
