package validation

import (
	"testing"
)

func TestNewValidator(t *testing.T) {
	type args struct {
		validateFn func(value string, ctx Context) *ValidationError
	}
	tests := []struct {
		name string
		args args
		want Validator[string]
	}{
		{
			name: "creates validator with validation function",
			args: args{
				validateFn: func(value string, ctx Context) *ValidationError {
					if value == "error" {
						return NewValidationError(ctx.Target, value, "test error")
					}
					return nil
				},
			},
			want: Validator[string]{
				Name: nil,
				Validate: func(value string, ctx Context) *ValidationError {
					if value == "error" {
						return NewValidationError(ctx.Target, value, "test error")
					}
					return nil
				},
			},
		},
		{
			name: "creates validator that always passes",
			args: args{
				validateFn: func(value string, ctx Context) *ValidationError {
					return nil
				},
			},
			want: Validator[string]{
				Name: nil,
				Validate: func(value string, ctx Context) *ValidationError {
					return nil
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewValidator(tt.args.validateFn)

			// Test the structure
			if got.Name != nil {
				t.Errorf("NewValidator() Name = %v, want nil", got.Name)
			}

			if got.Validate == nil {
				t.Error("NewValidator() Validate function should not be nil")
			}

			// Test the function behavior
			ctx := NewContext("test")

			// Test with valid input
			if err := got.Validate("valid", ctx); err != nil && tt.name == "creates validator that always passes" {
				t.Errorf("NewValidator() validation failed for valid input: %v", err)
			}

			// Test with error input for the first test case
			if tt.name == "creates validator with validation function" {
				if err := got.Validate("error", ctx); err == nil {
					t.Error("NewValidator() should have returned error for 'error' input")
				}
			}
		})
	}
}

func TestValidator_WithName(t *testing.T) {
	type testCase struct {
		name         string
		validator    Validator[string]
		nameToAdd    string
		expectedName string
	}

	tests := []testCase{
		{
			name: "adds name to unnamed validator",
			validator: NewValidator(func(value string, ctx Context) *ValidationError {
				return nil
			}),
			nameToAdd:    "test_validator",
			expectedName: "test_validator",
		},
		{
			name: "overwrites existing name",
			validator: NewValidator(func(value string, ctx Context) *ValidationError {
				return nil
			}).WithName("old_name"),
			nameToAdd:    "new_name",
			expectedName: "new_name",
		},
		{
			name: "handles empty name",
			validator: NewValidator(func(value string, ctx Context) *ValidationError {
				return nil
			}),
			nameToAdd:    "",
			expectedName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.validator.WithName(tt.nameToAdd)

			if got.Name == nil {
				t.Error("WithName() should set Name field")
				return
			}

			if *got.Name != tt.expectedName {
				t.Errorf("WithName() = %v, want %v", *got.Name, tt.expectedName)
			}

			// Verify original validator is unchanged
			if tt.validator.Name != nil && tt.name == "adds name to unnamed validator" {
				t.Error("WithName() should not modify original validator")
			}
		})
	}
}

func TestValidator_GetName(t *testing.T) {
	tests := []struct {
		name      string
		validator Validator[string]
		want      string
	}{
		{
			name: "returns default name for unnamed validator",
			validator: NewValidator(func(value string, ctx Context) *ValidationError {
				return nil
			}),
			want: "unnamed-validator",
		},
		{
			name: "returns custom name",
			validator: NewValidator(func(value string, ctx Context) *ValidationError {
				return nil
			}).WithName("custom_validator"),
			want: "custom_validator",
		},
		{
			name: "returns empty name when set to empty",
			validator: NewValidator(func(value string, ctx Context) *ValidationError {
				return nil
			}).WithName(""),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.validator.GetName(); got != tt.want {
				t.Errorf("Validator.GetName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewValidatorChain(t *testing.T) {
	validator1 := NewValidator(func(value string, ctx Context) *ValidationError {
		return nil
	})
	validator2 := NewValidator(func(value string, ctx Context) *ValidationError {
		return nil
	})

	tests := []struct {
		name       string
		validators []Validator[string]
		wantCount  int
	}{
		{
			name:       "creates empty chain",
			validators: []Validator[string]{},
			wantCount:  0,
		},
		{
			name:       "creates chain with single validator",
			validators: []Validator[string]{validator1},
			wantCount:  1,
		},
		{
			name:       "creates chain with multiple validators",
			validators: []Validator[string]{validator1, validator2},
			wantCount:  2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewValidatorChain(tt.validators...)

			if len(got.validators) != tt.wantCount {
				t.Errorf("NewValidatorChain() validator count = %v, want %v", len(got.validators), tt.wantCount)
			}
		})
	}
}

func TestValidatorChain_Add(t *testing.T) {
	tests := []struct {
		name          string
		initialCount  int
		addValidator  bool
		expectedCount int
	}{
		{
			name:          "adds validator to empty chain",
			initialCount:  0,
			addValidator:  true,
			expectedCount: 1,
		},
		{
			name:          "adds validator to existing chain",
			initialCount:  2,
			addValidator:  true,
			expectedCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create initial chain
			var initialValidators []Validator[string]
			for i := 0; i < tt.initialCount; i++ {
				initialValidators = append(initialValidators, NewValidator(func(value string, ctx Context) *ValidationError {
					return nil
				}))
			}

			chain := NewValidatorChain(initialValidators...)

			// Add validator if specified
			if tt.addValidator {
				newValidator := NewValidator(func(value string, ctx Context) *ValidationError {
					return nil
				})
				result := chain.Add(newValidator)

				// Verify it returns the same chain instance
				if result != chain {
					t.Error("Add() should return the same chain instance for fluent interface")
				}
			}

			if len(chain.validators) != tt.expectedCount {
				t.Errorf("ValidatorChain.Add() validator count = %v, want %v", len(chain.validators), tt.expectedCount)
			}
		})
	}
}

func TestValidatorChain_Validate(t *testing.T) {
	tests := []struct {
		name           string
		validators     []func(value string, ctx Context) *ValidationError
		testValue      string
		expectedErrors int
	}{
		{
			name:           "empty chain returns no errors",
			validators:     []func(value string, ctx Context) *ValidationError{},
			testValue:      "any_value",
			expectedErrors: 0,
		},
		{
			name: "all validators pass",
			validators: []func(value string, ctx Context) *ValidationError{
				func(value string, ctx Context) *ValidationError { return nil },
				func(value string, ctx Context) *ValidationError { return nil },
			},
			testValue:      "valid_value",
			expectedErrors: 0,
		},
		{
			name: "some validators fail",
			validators: []func(value string, ctx Context) *ValidationError{
				func(value string, ctx Context) *ValidationError { return nil },
				func(value string, ctx Context) *ValidationError {
					return NewValidationError(ctx.Target, value, "error 1")
				},
				func(value string, ctx Context) *ValidationError {
					return NewValidationError(ctx.Target, value, "error 2")
				},
			},
			testValue:      "invalid_value",
			expectedErrors: 2,
		},
		{
			name: "all validators fail",
			validators: []func(value string, ctx Context) *ValidationError{
				func(value string, ctx Context) *ValidationError {
					return NewValidationError(ctx.Target, value, "error 1")
				},
				func(value string, ctx Context) *ValidationError {
					return NewValidationError(ctx.Target, value, "error 2")
				},
			},
			testValue:      "invalid_value",
			expectedErrors: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create validators
			var validators []Validator[string]
			for _, validateFn := range tt.validators {
				validators = append(validators, NewValidator(validateFn))
			}

			chain := NewValidatorChain(validators...)
			ctx := NewContext("test_field")

			result := chain.Validate(tt.testValue, ctx)

			if len(result.Errors) != tt.expectedErrors {
				t.Errorf("ValidatorChain.Validate() error count = %v, want %v", len(result.Errors), tt.expectedErrors)
			}

			if result.HasErrors() != (tt.expectedErrors > 0) {
				t.Errorf("ValidatorChain.Validate() HasErrors() = %v, want %v", result.HasErrors(), tt.expectedErrors > 0)
			}
		})
	}
}

func TestValidatorChain_ValidateFirst(t *testing.T) {
	tests := []struct {
		name         string
		validators   []func(value string, ctx Context) *ValidationError
		testValue    string
		expectError  bool
		errorMessage string
	}{
		{
			name:        "empty chain returns no error",
			validators:  []func(value string, ctx Context) *ValidationError{},
			testValue:   "any_value",
			expectError: false,
		},
		{
			name: "first validator fails",
			validators: []func(value string, ctx Context) *ValidationError{
				func(value string, ctx Context) *ValidationError {
					return NewValidationError(ctx.Target, value, "first error")
				},
				func(value string, ctx Context) *ValidationError {
					return NewValidationError(ctx.Target, value, "second error")
				},
			},
			testValue:    "invalid_value",
			expectError:  true,
			errorMessage: "first error",
		},
		{
			name: "second validator fails",
			validators: []func(value string, ctx Context) *ValidationError{
				func(value string, ctx Context) *ValidationError { return nil },
				func(value string, ctx Context) *ValidationError {
					return NewValidationError(ctx.Target, value, "second error")
				},
			},
			testValue:    "invalid_value",
			expectError:  true,
			errorMessage: "second error",
		},
		{
			name: "all validators pass",
			validators: []func(value string, ctx Context) *ValidationError{
				func(value string, ctx Context) *ValidationError { return nil },
				func(value string, ctx Context) *ValidationError { return nil },
			},
			testValue:   "valid_value",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create validators
			var validators []Validator[string]
			for _, validateFn := range tt.validators {
				validators = append(validators, NewValidator(validateFn))
			}

			chain := NewValidatorChain(validators...)
			ctx := NewContext("test_field")

			result := chain.ValidateFirst(tt.testValue, ctx)

			if tt.expectError {
				if result == nil {
					t.Error("ValidatorChain.ValidateFirst() expected error but got nil")
					return
				}
				if result.Error() != tt.errorMessage {
					t.Errorf("ValidatorChain.ValidateFirst() error message = %v, want %v", result.Error(), tt.errorMessage)
				}
			} else {
				if result != nil {
					t.Errorf("ValidatorChain.ValidateFirst() expected no error but got: %v", result)
				}
			}
		})
	}
}

func TestNewConditionalValidator(t *testing.T) {
	condition := func(value string, ctx Context) bool {
		return value == "trigger"
	}
	validator := NewValidator(func(value string, ctx Context) *ValidationError {
		return NewValidationError(ctx.Target, value, "validation error")
	})

	tests := []struct {
		name      string
		condition func(value string, ctx Context) bool
		validator Validator[string]
	}{
		{
			name:      "creates conditional validator",
			condition: condition,
			validator: validator,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := NewConditionalValidator(tt.condition, tt.validator)

			if got.Condition == nil {
				t.Error("NewConditionalValidator() Condition should not be nil")
			}

			if got.Validator.Validate == nil {
				t.Error("NewConditionalValidator() Validator should not be nil")
			}
		})
	}
}

func TestConditionalValidator_Validate(t *testing.T) {
	tests := []struct {
		name        string
		condition   func(value string, ctx Context) bool
		validator   func(value string, ctx Context) *ValidationError
		testValue   string
		expectError bool
	}{
		{
			name: "condition met and validator fails",
			condition: func(value string, ctx Context) bool {
				return value == "trigger"
			},
			validator: func(value string, ctx Context) *ValidationError {
				return NewValidationError(ctx.Target, value, "validation error")
			},
			testValue:   "trigger",
			expectError: true,
		},
		{
			name: "condition not met",
			condition: func(value string, ctx Context) bool {
				return value == "trigger"
			},
			validator: func(value string, ctx Context) *ValidationError {
				return NewValidationError(ctx.Target, value, "validation error")
			},
			testValue:   "not_trigger",
			expectError: false,
		},
		{
			name: "condition met but validator passes",
			condition: func(value string, ctx Context) bool {
				return value == "trigger"
			},
			validator: func(value string, ctx Context) *ValidationError {
				return nil
			},
			testValue:   "trigger",
			expectError: false,
		},
		{
			name: "condition always false",
			condition: func(value string, ctx Context) bool {
				return false
			},
			validator: func(value string, ctx Context) *ValidationError {
				return NewValidationError(ctx.Target, value, "should not be called")
			},
			testValue:   "any_value",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NewValidator(tt.validator)
			conditional := NewConditionalValidator(tt.condition, validator)
			ctx := NewContext("test_field")

			result := conditional.Validate(tt.testValue, ctx)

			if tt.expectError {
				if result == nil {
					t.Error("ConditionalValidator.Validate() expected error but got nil")
				}
			} else {
				if result != nil {
					t.Errorf("ConditionalValidator.Validate() expected no error but got: %v", result)
				}
			}
		})
	}
}

func TestConditionalValidator_WithName(t *testing.T) {
	tests := []struct {
		name         string
		originalName *string
		newName      string
		expectedName string
	}{
		{
			name:         "adds name to unnamed conditional validator",
			originalName: nil,
			newName:      "conditional_test",
			expectedName: "conditional_test",
		},
		{
			name:         "overwrites existing name",
			originalName: stringPtr("old_name"),
			newName:      "new_name",
			expectedName: "new_name",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			condition := func(value string, ctx Context) bool { return true }
			validator := NewValidator(func(value string, ctx Context) *ValidationError {
				return nil
			})

			if tt.originalName != nil {
				validator = validator.WithName(*tt.originalName)
			}

			conditional := NewConditionalValidator(condition, validator)
			named := conditional.WithName(tt.newName)

			if named.Validator.Name == nil {
				t.Error("ConditionalValidator.WithName() should set inner validator name")
				return
			}

			if *named.Validator.Name != tt.expectedName {
				t.Errorf("ConditionalValidator.WithName() = %v, want %v", *named.Validator.Name, tt.expectedName)
			}

			// Verify original is unchanged
			if conditional.Validator.Name != nil && tt.originalName == nil {
				t.Error("ConditionalValidator.WithName() should not modify original")
			}
		})
	}
}

// Integration tests
func TestValidatorIntegration(t *testing.T) {
	t.Run("complex validator chain", func(t *testing.T) {
		// Create a complex validation scenario
		minLength := NewValidator(func(value string, ctx Context) *ValidationError {
			if len(value) < 3 {
				return NewValidationError(ctx.Target, value, "too short")
			}
			return nil
		}).WithName("min-length")

		notEmpty := NewValidator(func(value string, ctx Context) *ValidationError {
			if value == "" {
				return NewValidationError(ctx.Target, value, "cannot be empty")
			}
			return nil
		}).WithName("not-empty")

		chain := NewValidatorChain(notEmpty, minLength)
		ctx := NewContext("username")

		// Test valid value
		result := chain.Validate("valid_username", ctx)
		if result.HasErrors() {
			t.Errorf("Valid value should pass: %v", result.Error())
		}

		// Test invalid value
		result = chain.Validate("ab", ctx)
		if !result.HasErrors() {
			t.Error("Invalid value should fail")
		}
		if len(result.Errors) != 1 {
			t.Errorf("Expected 1 error, got %d", len(result.Errors))
		}
	})

	t.Run("conditional validator with context metadata", func(t *testing.T) {
		condition := func(value string, ctx Context) bool {
			required, exists := ctx.Metadata["required"]
			return exists && required.(bool)
		}

		validator := NewValidator(func(value string, ctx Context) *ValidationError {
			if value == "" {
				return NewValidationError(ctx.Target, value, "required field cannot be empty")
			}
			return nil
		})

		conditional := NewConditionalValidator(condition, validator)

		// Test with required=true
		ctx := NewContext("field").WithMetadata("required", true)
		result := conditional.Validate("", ctx)
		if result == nil {
			t.Error("Required field validation should fail for empty value")
		}

		// Test with required=false
		ctx = NewContext("field").WithMetadata("required", false)
		result = conditional.Validate("", ctx)
		if result != nil {
			t.Errorf("Non-required field should pass: %v", result)
		}

		// Test without required metadata
		ctx = NewContext("field")
		result = conditional.Validate("", ctx)
		if result != nil {
			t.Errorf("Field without required metadata should pass: %v", result)
		}
	})
}
