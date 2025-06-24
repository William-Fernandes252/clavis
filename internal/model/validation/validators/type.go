package validators

import (
	"github.com/William-Fernandes252/clavis/internal/model/validation"
)

// Validator is a generic type that defines a validation function for a specific type T.
type Validator[T any] struct {
	Name     *string // Optional name of the validator
	Validate func(value T, ctx validation.Context) *validation.ValidationError
}

// NewValidator creates a new validator with just the validation function
func NewValidator[T any](validateFn func(value T, ctx validation.Context) *validation.ValidationError) Validator[T] {
	return Validator[T]{
		Name:     nil,
		Validate: validateFn,
	}
}

// WithName returns a new validator with the specified name
func (v Validator[T]) WithName(name string) Validator[T] {
	return Validator[T]{
		Name:     &name,
		Validate: v.Validate,
	}
}

// GetName returns the validator name or a default if not set
func (v Validator[T]) GetName() string {
	if v.Name != nil {
		return *v.Name
	}
	return "unnamed-validator"
}

// ValidatorChain allows chaining multiple validators together
type ValidatorChain[T any] struct {
	validators []Validator[T]
}

// NewValidatorChain creates a new validator chain
func NewValidatorChain[T any](validators ...Validator[T]) *ValidatorChain[T] {
	return &ValidatorChain[T]{
		validators: validators,
	}
}

// Add adds a validator to the chain
func (vc *ValidatorChain[T]) Add(validator Validator[T]) *ValidatorChain[T] {
	vc.validators = append(vc.validators, validator)
	return vc
}

// Validate runs all validators in the chain and returns all errors
func (vc *ValidatorChain[T]) Validate(value T, ctx validation.Context) *validation.ValidationResult {
	errs := validation.NewValidationResult()

	for _, validator := range vc.validators {
		if err := validator.Validate(value, ctx); err != nil {
			errs.Add(*err)
		}
	}

	return errs
}

// ValidateFirst runs validators until the first error and returns it
func (vc *ValidatorChain[T]) ValidateFirst(value T, ctx validation.Context) *validation.ValidationError {
	for _, validator := range vc.validators {
		if err := validator.Validate(value, ctx); err != nil {
			return err
		}
	}
	return nil
}

// ConditionalValidator wraps a validator with a condition
type ConditionalValidator[T any] struct {
	Condition func(value T, ctx validation.Context) bool
	Validator Validator[T]
}

// NewConditionalValidator creates a validator that only runs if the condition is met
func NewConditionalValidator[T any](condition func(value T, ctx validation.Context) bool, validator Validator[T]) ConditionalValidator[T] {
	return ConditionalValidator[T]{
		Condition: condition,
		Validator: validator,
	}
}

// Validate runs the validator only if the condition is met
func (cv ConditionalValidator[T]) Validate(value T, ctx validation.Context) *validation.ValidationError {
	if cv.Condition(value, ctx) {
		return cv.Validator.Validate(value, ctx)
	}
	return nil
}

// WithName allows adding a name to a conditional validator
func (cv ConditionalValidator[T]) WithName(name string) ConditionalValidator[T] {
	return ConditionalValidator[T]{
		Condition: cv.Condition,
		Validator: cv.Validator.WithName(name),
	}
}
