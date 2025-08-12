package data

import "github.com/William-Fernandes252/clavis/internal/errors"

const stringDataType = Type("string")

// StringCodec represents sequences of bytes, including text, serialized objects, and binary data.
// It implements the Data interface for string values.
type StringCodec struct{}

// Type returns the type name ("string")
func (s *StringCodec) Type() Type {
	return stringDataType
}

// Value returns the in-memory representation of the data
func (s *StringCodec) Serialize(value string) ([]byte, errors.Error) {
	return []byte(value), nil
}

// Deserialize reads the bytes back into the data type
// It returns an empty string if the input is nil
func (s *StringCodec) Deserialize(raw []byte) (string, errors.Error) {
	if raw == nil {
		return "", nil
	}

	return string(raw), nil
}

var _ Codec[string] = (*StringCodec)(nil)
