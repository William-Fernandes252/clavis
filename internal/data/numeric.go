package data

import (
	"fmt"

	"github.com/William-Fernandes252/clavis/internal/errors"
)

const numericDataType = Type("numeric")

// NumberCodec represents a numeric data type. Can be used to interpret data as floating-point numbers, allowing operations like increment, decrement, and arithmetic operations.
// It implements the Data interface for numeric values.
type NumberCodec struct{}

// Type returns the type name ("numeric")
func (n *NumberCodec) Type() Type {
	return numericDataType
}

// Serialize converts the numeric value into a byte slice for storage
func (n *NumberCodec) Serialize(value float64) ([]byte, errors.Error) {
	return fmt.Appendf(nil, "%v", value), nil
}

// Deserialize reads the bytes back into the numeric data type
// It returns 0 if the input is nil or cannot be parsed
func (n *NumberCodec) Deserialize(raw []byte) (value float64, err errors.Error) {
	if raw == nil {
		return 0, nil
	}
	_, stdErr := fmt.Sscanf(string(raw), "%f", &value)
	if stdErr != nil {
		return 0, NewDataError("deserialization-failed", "Failed to deserialize numeric data", stdErr)
	}
	return
}

var _ Codec[float64] = (*NumberCodec)(nil)
