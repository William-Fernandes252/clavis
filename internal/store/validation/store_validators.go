package validation

import (
	"fmt"

	"github.com/William-Fernandes252/clavis/internal/model/validation"
	"github.com/William-Fernandes252/clavis/internal/model/validation/validators"
)

// StoreKeyValidator wraps a chain of validators for store keys
type StoreKeyValidator struct {
	chain *validators.ValidatorChain[string]
}

// NewStoreKeyValidator creates a new store key validator with a chain of validators
func NewStoreKeyValidator(validatorList ...validators.Validator[string]) *StoreKeyValidator {
	return &StoreKeyValidator{
		chain: validators.NewValidatorChain(validatorList...),
	}
}

// Validate validates a store key and returns a traditional error for compatibility
func (skv *StoreKeyValidator) Validate(key string) error {
	ctx := validation.NewContext("key")
	result := skv.chain.Validate(key, ctx)

	if result.HasErrors() {
		// Return the first error for compatibility with existing interface
		return fmt.Errorf("%s", result.Errors[0].Error())
	}
	return nil
}

// StoreValueValidator wraps validators for store values
type StoreValueValidator struct {
	validators []func(key string, value []byte, ctx validation.Context) *validation.ValidationError
}

// NewStoreValueValidator creates a new store value validator
func NewStoreValueValidator(validatorFuncs ...func(key string, value []byte, ctx validation.Context) *validation.ValidationError) *StoreValueValidator {
	return &StoreValueValidator{
		validators: validatorFuncs,
	}
}

// Validate validates a store value and returns a traditional error for compatibility
func (svv *StoreValueValidator) Validate(key string, value []byte) error {
	ctx := validation.NewContext("value").WithMetadata("key", key)

	for _, validator := range svv.validators {
		if err := validator(key, value, ctx); err != nil {
			return fmt.Errorf("%s", err.Error())
		}
	}
	return nil
}

// Predefined key validators using the new validation framework

// NonEmptyKeyValidator validates that a key is not empty
func NonEmptyKeyValidator() validators.Validator[string] {
	return validators.NotEmpty().WithName("non-empty-key")
}

// KeyLengthValidator validates key length within a maximum limit
func KeyLengthValidator(maxLength int) validators.Validator[string] {
	return validators.MaxLength(maxLength).WithName("key-length")
}

// KeyPatternValidator validates key matches a pattern
func KeyPatternValidator(pattern string) validators.Validator[string] {
	return validators.Pattern(pattern).WithName("key-pattern")
}

// Predefined value validators

// ValueSizeValidator validates that value size is within limit
func ValueSizeValidator(maxSize int64) func(string, []byte, validation.Context) *validation.ValidationError {
	return func(key string, value []byte, ctx validation.Context) *validation.ValidationError {
		if int64(len(value)) > maxSize {
			msg := fmt.Sprintf("value too large: maximum %d bytes, got %d", maxSize, len(value))
			return validation.NewValidationError(
				ctx.Target,
				len(value),
				msg,
			).WithMetadata("max-size", maxSize).
				WithMetadata("actual-size", len(value)).
				WithMetadata("key", key).
				WithCode("value-too-large")
		}
		return nil
	}
}

// ValueContentValidator validates value content
func ValueContentValidator(validateFn func([]byte) bool, errorMsg string) func(string, []byte, validation.Context) *validation.ValidationError {
	return func(key string, value []byte, ctx validation.Context) *validation.ValidationError {
		if !validateFn(value) {
			msg := fmt.Sprintf("value validation failed: %s", errorMsg)
			return validation.NewValidationError(
				ctx.Target,
				string(value),
				msg,
			).WithMetadata("key", key).
				WithCode("value-content-invalid")
		}
		return nil
	}
}

// Helper functions for creating default validators

// DefaultKeyValidators returns the standard set of key validators
func DefaultKeyValidators() []validators.Validator[string] {
	return []validators.Validator[string]{
		NonEmptyKeyValidator(),
		KeyLengthValidator(1024),
	}
}

// DefaultValueValidators returns the standard set of value validators
func DefaultValueValidators() []func(string, []byte, validation.Context) *validation.ValidationError {
	return []func(string, []byte, validation.Context) *validation.ValidationError{
		ValueSizeValidator(100 * 1024 * 1024), // 100MB
	}
}
