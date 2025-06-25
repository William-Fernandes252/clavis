package validators

import (
	"testing"

	"github.com/William-Fernandes252/clavis/internal/model/validation"
)

func TestNumericRange(t *testing.T) {
	tests := []struct {
		name        string
		min         int
		max         int
		value       int
		expectError bool
		errorCode   string
	}{
		{
			name:        "value within range",
			min:         1,
			max:         10,
			value:       5,
			expectError: false,
		},
		{
			name:        "value at minimum boundary",
			min:         1,
			max:         10,
			value:       1,
			expectError: false,
		},
		{
			name:        "value at maximum boundary",
			min:         1,
			max:         10,
			value:       10,
			expectError: false,
		},
		{
			name:        "value below minimum",
			min:         1,
			max:         10,
			value:       0,
			expectError: true,
			errorCode:   "numeric-range",
		},
		{
			name:        "value above maximum",
			min:         1,
			max:         10,
			value:       11,
			expectError: true,
			errorCode:   "numeric-range",
		},
		{
			name:        "negative range",
			min:         -10,
			max:         -1,
			value:       -5,
			expectError: false,
		},
		{
			name:        "negative range violation",
			min:         -10,
			max:         -1,
			value:       0,
			expectError: true,
			errorCode:   "numeric-range",
		},
		{
			name:        "zero range (min equals max)",
			min:         5,
			max:         5,
			value:       5,
			expectError: false,
		},
		{
			name:        "zero range violation",
			min:         5,
			max:         5,
			value:       4,
			expectError: true,
			errorCode:   "numeric-range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NumericRange(tt.min, tt.max)
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("NumericRange(%d, %d) expected error but got nil for value %d", tt.min, tt.max, tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("NumericRange(%d, %d) error code = %v, want %v", tt.min, tt.max, result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("NumericRange(%d, %d) expected no error but got: %v for value %d", tt.min, tt.max, result, tt.value)
				}
			}
		})
	}
}

func TestNumericRangeFloat(t *testing.T) {
	tests := []struct {
		name        string
		min         float64
		max         float64
		value       float64
		expectError bool
		errorCode   string
	}{
		{
			name:        "float value within range",
			min:         1.5,
			max:         10.5,
			value:       5.7,
			expectError: false,
		},
		{
			name:        "float value at minimum boundary",
			min:         1.5,
			max:         10.5,
			value:       1.5,
			expectError: false,
		},
		{
			name:        "float value at maximum boundary",
			min:         1.5,
			max:         10.5,
			value:       10.5,
			expectError: false,
		},
		{
			name:        "float value below minimum",
			min:         1.5,
			max:         10.5,
			value:       1.4,
			expectError: true,
			errorCode:   "numeric-range",
		},
		{
			name:        "float value above maximum",
			min:         1.5,
			max:         10.5,
			value:       10.6,
			expectError: true,
			errorCode:   "numeric-range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NumericRange(tt.min, tt.max)
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("NumericRange(%v, %v) expected error but got nil for value %v", tt.min, tt.max, tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("NumericRange(%v, %v) error code = %v, want %v", tt.min, tt.max, result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("NumericRange(%v, %v) expected no error but got: %v for value %v", tt.min, tt.max, result, tt.value)
				}
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name        string
		min         int
		value       int
		expectError bool
		errorCode   string
	}{
		{
			name:        "value above minimum",
			min:         5,
			value:       10,
			expectError: false,
		},
		{
			name:        "value at minimum",
			min:         5,
			value:       5,
			expectError: false,
		},
		{
			name:        "value below minimum",
			min:         5,
			value:       3,
			expectError: true,
			errorCode:   "min-value",
		},
		{
			name:        "negative minimum with valid value",
			min:         -10,
			value:       -5,
			expectError: false,
		},
		{
			name:        "negative minimum with invalid value",
			min:         -10,
			value:       -15,
			expectError: true,
			errorCode:   "min-value",
		},
		{
			name:        "zero minimum",
			min:         0,
			value:       0,
			expectError: false,
		},
		{
			name:        "zero minimum with negative value",
			min:         0,
			value:       -1,
			expectError: true,
			errorCode:   "min-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := Min(tt.min)
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("Min(%d) expected error but got nil for value %d", tt.min, tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("Min(%d) error code = %v, want %v", tt.min, result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("Min(%d) expected no error but got: %v for value %d", tt.min, result, tt.value)
				}
			}
		})
	}
}

func TestMinFloat(t *testing.T) {
	tests := []struct {
		name        string
		min         float64
		value       float64
		expectError bool
		errorCode   string
	}{
		{
			name:        "float value above minimum",
			min:         5.5,
			value:       10.7,
			expectError: false,
		},
		{
			name:        "float value at minimum",
			min:         5.5,
			value:       5.5,
			expectError: false,
		},
		{
			name:        "float value below minimum",
			min:         5.5,
			value:       3.2,
			expectError: true,
			errorCode:   "min-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := Min(tt.min)
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("Min(%v) expected error but got nil for value %v", tt.min, tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("Min(%v) error code = %v, want %v", tt.min, result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("Min(%v) expected no error but got: %v for value %v", tt.min, result, tt.value)
				}
			}
		})
	}
}

func TestMax(t *testing.T) {
	tests := []struct {
		name        string
		max         int
		value       int
		expectError bool
		errorCode   string
	}{
		{
			name:        "value below maximum",
			max:         10,
			value:       5,
			expectError: false,
		},
		{
			name:        "value at maximum",
			max:         10,
			value:       10,
			expectError: false,
		},
		{
			name:        "value above maximum",
			max:         10,
			value:       15,
			expectError: true,
			errorCode:   "max-value",
		},
		{
			name:        "negative maximum with valid value",
			max:         -5,
			value:       -10,
			expectError: false,
		},
		{
			name:        "negative maximum with invalid value",
			max:         -5,
			value:       -3,
			expectError: true,
			errorCode:   "max-value",
		},
		{
			name:        "zero maximum",
			max:         0,
			value:       0,
			expectError: false,
		},
		{
			name:        "zero maximum with positive value",
			max:         0,
			value:       1,
			expectError: true,
			errorCode:   "max-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := Max(tt.max)
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("Max(%d) expected error but got nil for value %d", tt.max, tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("Max(%d) error code = %v, want %v", tt.max, result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("Max(%d) expected no error but got: %v for value %d", tt.max, result, tt.value)
				}
			}
		})
	}
}

func TestMaxFloat(t *testing.T) {
	tests := []struct {
		name        string
		max         float64
		value       float64
		expectError bool
		errorCode   string
	}{
		{
			name:        "float value below maximum",
			max:         10.5,
			value:       5.2,
			expectError: false,
		},
		{
			name:        "float value at maximum",
			max:         10.5,
			value:       10.5,
			expectError: false,
		},
		{
			name:        "float value above maximum",
			max:         10.5,
			value:       15.7,
			expectError: true,
			errorCode:   "max-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := Max(tt.max)
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("Max(%v) expected error but got nil for value %v", tt.max, tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("Max(%v) error code = %v, want %v", tt.max, result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("Max(%v) expected no error but got: %v for value %v", tt.max, result, tt.value)
				}
			}
		})
	}
}

func TestPositive(t *testing.T) {
	tests := []struct {
		name        string
		value       int
		expectError bool
		errorCode   string
	}{
		{
			name:        "positive integer",
			value:       5,
			expectError: false,
		},
		{
			name:        "large positive integer",
			value:       1000,
			expectError: false,
		},
		{
			name:        "zero is not positive",
			value:       0,
			expectError: true,
			errorCode:   "non-positive",
		},
		{
			name:        "negative integer",
			value:       -5,
			expectError: true,
			errorCode:   "non-positive",
		},
		{
			name:        "large negative integer",
			value:       -1000,
			expectError: true,
			errorCode:   "non-positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := Positive[int]()
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("Positive() expected error but got nil for value %d", tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("Positive() error code = %v, want %v", result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("Positive() expected no error but got: %v for value %d", result, tt.value)
				}
			}
		})
	}
}

func TestPositiveFloat(t *testing.T) {
	tests := []struct {
		name        string
		value       float64
		expectError bool
		errorCode   string
	}{
		{
			name:        "positive float",
			value:       5.7,
			expectError: false,
		},
		{
			name:        "small positive float",
			value:       0.001,
			expectError: false,
		},
		{
			name:        "zero float is not positive",
			value:       0.0,
			expectError: true,
			errorCode:   "non-positive",
		},
		{
			name:        "negative float",
			value:       -5.7,
			expectError: true,
			errorCode:   "non-positive",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := Positive[float64]()
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("Positive() expected error but got nil for value %v", tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("Positive() error code = %v, want %v", result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("Positive() expected no error but got: %v for value %v", result, tt.value)
				}
			}
		})
	}
}

func TestNonNegative(t *testing.T) {
	tests := []struct {
		name        string
		value       int
		expectError bool
		errorCode   string
	}{
		{
			name:        "positive integer",
			value:       5,
			expectError: false,
		},
		{
			name:        "zero is non-negative",
			value:       0,
			expectError: false,
		},
		{
			name:        "large positive integer",
			value:       1000,
			expectError: false,
		},
		{
			name:        "negative integer",
			value:       -5,
			expectError: true,
			errorCode:   "non-negative",
		},
		{
			name:        "large negative integer",
			value:       -1000,
			expectError: true,
			errorCode:   "non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NonNegative[int]()
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("NonNegative() expected error but got nil for value %d", tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("NonNegative() error code = %v, want %v", result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("NonNegative() expected no error but got: %v for value %d", result, tt.value)
				}
			}
		})
	}
}

func TestNonNegativeFloat(t *testing.T) {
	tests := []struct {
		name        string
		value       float64
		expectError bool
		errorCode   string
	}{
		{
			name:        "positive float",
			value:       5.7,
			expectError: false,
		},
		{
			name:        "zero float is non-negative",
			value:       0.0,
			expectError: false,
		},
		{
			name:        "small positive float",
			value:       0.001,
			expectError: false,
		},
		{
			name:        "negative float",
			value:       -5.7,
			expectError: true,
			errorCode:   "non-negative",
		},
		{
			name:        "small negative float",
			value:       -0.001,
			expectError: true,
			errorCode:   "non-negative",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NonNegative[float64]()
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("NonNegative() expected error but got nil for value %v", tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("NonNegative() error code = %v, want %v", result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("NonNegative() expected no error but got: %v for value %v", result, tt.value)
				}
			}
		})
	}
}

// Test different numeric types with the same validator
func TestNumericValidatorWithDifferentTypes(t *testing.T) {
	t.Run("int8 positive", func(t *testing.T) {
		validator := Positive[int8]()
		ctx := validation.NewContext("test_field")

		result := validator.Validate(int8(5), ctx)
		if result != nil {
			t.Errorf("Positive[int8]() expected no error but got: %v", result)
		}

		result = validator.Validate(int8(-1), ctx)
		if result == nil {
			t.Error("Positive[int8]() expected error for negative value but got nil")
		}
	})

	t.Run("uint range", func(t *testing.T) {
		validator := NumericRange[uint](uint(1), uint(10))
		ctx := validation.NewContext("test_field")

		result := validator.Validate(uint(5), ctx)
		if result != nil {
			t.Errorf("NumericRange[uint]() expected no error but got: %v", result)
		}

		result = validator.Validate(uint(15), ctx)
		if result == nil {
			t.Error("NumericRange[uint]() expected error for out-of-range value but got nil")
		}
	})

	t.Run("float32 min", func(t *testing.T) {
		validator := Min[float32](float32(1.5))
		ctx := validation.NewContext("test_field")

		result := validator.Validate(float32(2.5), ctx)
		if result != nil {
			t.Errorf("Min[float32]() expected no error but got: %v", result)
		}

		result = validator.Validate(float32(0.5), ctx)
		if result == nil {
			t.Error("Min[float32]() expected error for below-minimum value but got nil")
		}
	})
}

// Test validator composition with numeric validators
func TestNumericValidatorComposition(t *testing.T) {
	tests := []struct {
		name        string
		validators  []Validator[int]
		value       int
		expectError bool
		errorCode   string
	}{
		{
			name:        "Positive and Min both pass",
			validators:  []Validator[int]{Positive[int](), Min(1)},
			value:       5,
			expectError: false,
		},
		{
			name:        "Positive passes, Min fails",
			validators:  []Validator[int]{Positive[int](), Min(10)},
			value:       5,
			expectError: true,
			errorCode:   "min-value",
		},
		{
			name:        "Min and Max both pass",
			validators:  []Validator[int]{Min(1), Max(10)},
			value:       5,
			expectError: false,
		},
		{
			name:        "Range and Positive both pass",
			validators:  []Validator[int]{NumericRange(1, 10), Positive[int]()},
			value:       5,
			expectError: false,
		},
		{
			name:        "NonNegative passes, Max fails",
			validators:  []Validator[int]{NonNegative[int](), Max(3)},
			value:       5,
			expectError: true,
			errorCode:   "max-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := validation.NewContext("test_field")

			// Test each validator in the chain
			for _, validator := range tt.validators {
				result := validator.Validate(tt.value, ctx)

				if tt.expectError {
					if result != nil && result.Code() == tt.errorCode {
						// Found the expected error, test passes
						return
					}
				} else {
					if result != nil {
						t.Errorf("Validator chain expected no error but got: %v for value %d", result, tt.value)
						return
					}
				}
			}

			// If we expected an error but didn't find it
			if tt.expectError {
				t.Errorf("Validator chain expected error with code %v but got none for value %d", tt.errorCode, tt.value)
			}
		})
	}
}

// Test WithName functionality for numeric validators
func TestNumericValidatorWithName(t *testing.T) {
	tests := []struct {
		name         string
		validator    Validator[int]
		customName   string
		value        int
		expectError  bool
		expectedName string
	}{
		{
			name:         "Positive with custom name",
			validator:    Positive[int]().WithName("custom-positive"),
			customName:   "custom-positive",
			value:        -1,
			expectError:  true,
			expectedName: "custom-positive",
		},
		{
			name:         "Range with custom name",
			validator:    NumericRange(1, 10).WithName("age-range"),
			customName:   "age-range",
			value:        15,
			expectError:  true,
			expectedName: "age-range",
		},
		{
			name:         "Min with custom name",
			validator:    Min(5).WithName("minimum-value"),
			customName:   "minimum-value",
			value:        3,
			expectError:  true,
			expectedName: "minimum-value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := validation.NewContext("test_field")

			result := tt.validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("Expected error but got nil for value %d", tt.value)
					return
				}

				// Check that the validator name is preserved
				if tt.validator.GetName() != tt.expectedName {
					t.Errorf("Expected validator name %q but got %q", tt.expectedName, tt.validator.GetName())
				}
			} else {
				if result != nil {
					t.Errorf("Expected no error but got: %v for value %d", result, tt.value)
				}
			}
		})
	}
}

// Test context metadata propagation for numeric validators
func TestNumericValidatorContextMetadata(t *testing.T) {
	validator := Min(10)

	ctx := validation.NewContext("age_field").
		WithMetadata("input_type", "user_age").
		WithMetadata("validation_rule", "minimum_age")

	result := validator.Validate(5, ctx)

	if result == nil {
		t.Fatal("Expected validation error but got nil")
	}

	if result.Target != "age_field" {
		t.Errorf("Expected target 'age_field' but got %q", result.Target)
	}

	// Check that metadata is included in the error
	metadata := result.Metadata()
	if len(metadata) == 0 {
		t.Error("Expected metadata in validation error but got none")
	}

	// Should include validator-specific metadata
	if _, exists := metadata["min"]; !exists {
		t.Error("Expected 'min' metadata in validation error")
	}

	if _, exists := metadata["actual"]; !exists {
		t.Error("Expected 'actual' metadata in validation error")
	}
}
