package validators

import (
	"fmt"

	"github.com/William-Fernandes252/clavis/internal/model/validation"
)

// NumericRange validates that a number is within the specified range
func NumericRange[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](min, max T) Validator[T] {
	return NewValidator(func(value T, ctx validation.Context) *validation.ValidationError {
		if value < min || value > max {
			msg := fmt.Sprintf("%s: must be between %v and %v, got %v",
				ctx.Target, min, max, value)
			err := validation.NewValidationError(
				ctx.Target,
				value,
				msg,
			).WithMetadata("min", min).
				WithMetadata("max", max).
				WithMetadata("actual", value).
				WithCode("numeric-range")
			return err
		}
		return nil
	}).WithName("numeric-range")
}

// Min validates minimum value
func Min[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](min T) Validator[T] {
	return NewValidator(func(value T, ctx validation.Context) *validation.ValidationError {
		if value < min {
			msg := fmt.Sprintf("%s: must be at least %v, got %v",
				ctx.Target, min, value)
			err := validation.NewValidationError(
				ctx.Target,
				value,
				msg,
			).WithMetadata("min", min).
				WithMetadata("actual", value).
				WithCode("min-value")
			return err
		}
		return nil
	}).WithName("min-value")
}

// Max validates maximum value
func Max[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](max T) Validator[T] {
	return NewValidator(func(value T, ctx validation.Context) *validation.ValidationError {
		if value > max {
			msg := fmt.Sprintf("%s: must be at most %v, got %v",
				ctx.Target, max, value)
			err := validation.NewValidationError(
				ctx.Target,
				value,
				msg,
			).WithMetadata("max", max).
				WithMetadata("actual", value).
				WithCode("max-value")
			return err
		}
		return nil
	}).WithName("max-value")
}

// Positive validates that a number is positive
func Positive[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64]() Validator[T] {
	return NewValidator(func(value T, ctx validation.Context) *validation.ValidationError {
		if value <= 0 {
			msg := fmt.Sprintf("%s: must be positive, got %v", ctx.Target, value)
			err := validation.NewValidationError(
				ctx.Target,
				value,
				msg,
			).WithCode("non-positive")
			return err
		}
		return nil
	}).WithName("positive")
}

// NonNegative validates that a number is non-negative
func NonNegative[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64]() Validator[T] {
	return NewValidator(func(value T, ctx validation.Context) *validation.ValidationError {
		if value < 0 {
			msg := fmt.Sprintf("%s: must be non-negative, got %v", ctx.Target, value)
			err := validation.NewValidationError(
				ctx.Target,
				value,
				msg,
			).WithCode("non-negative")
			return err
		}
		return nil
	}).WithName("non-negative")
}
