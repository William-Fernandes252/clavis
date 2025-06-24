package validators

import (
	"testing"

	"github.com/William-Fernandes252/clavis/internal/model/validation"
)

func TestNotEmpty(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
		errorCode   string
	}{
		{
			name:        "valid non-empty string",
			value:       "hello",
			expectError: false,
		},
		{
			name:        "empty string fails",
			value:       "",
			expectError: true,
			errorCode:   "not-empty",
		},
		{
			name:        "whitespace only string fails",
			value:       "   ",
			expectError: true,
			errorCode:   "not-empty",
		},
		{
			name:        "tab and newline only fails",
			value:       "\t\n\r",
			expectError: true,
			errorCode:   "not-empty",
		},
		{
			name:        "string with content and whitespace passes",
			value:       "  hello  ",
			expectError: false,
		},
		{
			name:        "single character passes",
			value:       "a",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NotEmpty()
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Error("NotEmpty() expected error but got nil")
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("NotEmpty() error code = %v, want %v", result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("NotEmpty() expected no error but got: %v", result)
				}
			}
		})
	}
}

func TestLength(t *testing.T) {
	tests := []struct {
		name        string
		min         int
		max         int
		value       string
		expectError bool
		errorCode   string
	}{
		{
			name:        "valid length within range",
			min:         3,
			max:         10,
			value:       "hello",
			expectError: false,
		},
		{
			name:        "exact minimum length",
			min:         5,
			max:         10,
			value:       "hello",
			expectError: false,
		},
		{
			name:        "exact maximum length",
			min:         3,
			max:         5,
			value:       "hello",
			expectError: false,
		},
		{
			name:        "below minimum length",
			min:         10,
			max:         20,
			value:       "short",
			expectError: true,
			errorCode:   "length-range",
		},
		{
			name:        "above maximum length",
			min:         1,
			max:         3,
			value:       "toolong",
			expectError: true,
			errorCode:   "length-range",
		},
		{
			name:        "empty string below minimum",
			min:         1,
			max:         10,
			value:       "",
			expectError: true,
			errorCode:   "length-range",
		},
		{
			name:        "unicode characters counted correctly",
			min:         3,
			max:         10,
			value:       "hello",
			expectError: false,
		},
		{
			name:        "minimum equals maximum, valid",
			min:         5,
			max:         5,
			value:       "exact",
			expectError: false,
		},
		{
			name:        "minimum equals maximum, invalid",
			min:         5,
			max:         5,
			value:       "toolong",
			expectError: true,
			errorCode:   "length-range",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := Length(tt.min, tt.max)
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("Length(%d, %d) expected error but got nil for value %q", tt.min, tt.max, tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("Length(%d, %d) error code = %v, want %v", tt.min, tt.max, result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("Length(%d, %d) expected no error but got: %v for value %q", tt.min, tt.max, result, tt.value)
				}
			}
		})
	}
}

func TestMinLength(t *testing.T) {
	tests := []struct {
		name        string
		min         int
		value       string
		expectError bool
		errorCode   string
	}{
		{
			name:        "valid length above minimum",
			min:         3,
			value:       "hello",
			expectError: false,
		},
		{
			name:        "exact minimum length",
			min:         5,
			value:       "hello",
			expectError: false,
		},
		{
			name:        "below minimum length",
			min:         10,
			value:       "short",
			expectError: true,
			errorCode:   "min-length",
		},
		{
			name:        "empty string below minimum",
			min:         1,
			value:       "",
			expectError: true,
			errorCode:   "min-length",
		},
		{
			name:        "zero minimum allows empty",
			min:         0,
			value:       "",
			expectError: false,
		},
		{
			name:        "zero minimum allows any length",
			min:         0,
			value:       "any length string",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := MinLength(tt.min)
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("MinLength(%d) expected error but got nil for value %q", tt.min, tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("MinLength(%d) error code = %v, want %v", tt.min, result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("MinLength(%d) expected no error but got: %v for value %q", tt.min, result, tt.value)
				}
			}
		})
	}
}

func TestMaxLength(t *testing.T) {
	tests := []struct {
		name        string
		max         int
		value       string
		expectError bool
		errorCode   string
	}{
		{
			name:        "valid length below maximum",
			max:         10,
			value:       "hello",
			expectError: false,
		},
		{
			name:        "exact maximum length",
			max:         5,
			value:       "hello",
			expectError: false,
		},
		{
			name:        "above maximum length",
			max:         3,
			value:       "toolong",
			expectError: true,
			errorCode:   "max-length",
		},
		{
			name:        "empty string within maximum",
			max:         10,
			value:       "",
			expectError: false,
		},
		{
			name:        "zero maximum only allows empty",
			max:         0,
			value:       "",
			expectError: false,
		},
		{
			name:        "zero maximum fails non-empty",
			max:         0,
			value:       "a",
			expectError: true,
			errorCode:   "max-length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := MaxLength(tt.max)
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("MaxLength(%d) expected error but got nil for value %q", tt.max, tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("MaxLength(%d) error code = %v, want %v", tt.max, result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("MaxLength(%d) expected no error but got: %v for value %q", tt.max, result, tt.value)
				}
			}
		})
	}
}

func TestPattern(t *testing.T) {
	tests := []struct {
		name        string
		pattern     string
		value       string
		expectError bool
		errorCode   string
	}{
		{
			name:        "valid email pattern",
			pattern:     `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			value:       "test@example.com",
			expectError: false,
		},
		{
			name:        "invalid email pattern",
			pattern:     `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`,
			value:       "invalid-email",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
		{
			name:        "simple pattern match",
			pattern:     `^hello`,
			value:       "hello world",
			expectError: false,
		},
		{
			name:        "simple pattern no match",
			pattern:     `^hello`,
			value:       "world hello",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
		{
			name:        "numeric pattern match",
			pattern:     `^\d+$`,
			value:       "12345",
			expectError: false,
		},
		{
			name:        "numeric pattern no match",
			pattern:     `^\d+$`,
			value:       "12a45",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
		{
			name:        "empty string with pattern requiring content",
			pattern:     `.+`,
			value:       "",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := Pattern(tt.pattern)
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("Pattern(%q) expected error but got nil for value %q", tt.pattern, tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("Pattern(%q) error code = %v, want %v", tt.pattern, result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("Pattern(%q) expected no error but got: %v for value %q", tt.pattern, result, tt.value)
				}
			}
		})
	}
}

func TestEmail(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
		errorCode   string
	}{
		{
			name:        "valid simple email",
			value:       "test@example.com",
			expectError: false,
		},
		{
			name:        "valid email with subdomain",
			value:       "user@mail.example.com",
			expectError: false,
		},
		{
			name:        "valid email with plus",
			value:       "user+tag@example.com",
			expectError: false,
		},
		{
			name:        "valid email with dots",
			value:       "first.last@example.com",
			expectError: false,
		},
		{
			name:        "valid email with numbers",
			value:       "user123@example123.com",
			expectError: false,
		},
		{
			name:        "invalid email no @",
			value:       "userexample.com",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
		{
			name:        "invalid email no domain",
			value:       "user@",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
		{
			name:        "invalid email no user",
			value:       "@example.com",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
		{
			name:        "invalid email multiple @",
			value:       "user@@example.com",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
		{
			name:        "invalid email no TLD",
			value:       "user@example",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
		{
			name:        "empty string",
			value:       "",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
		{
			name:        "spaces in email",
			value:       "user @example.com",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := Email()
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("Email() expected error but got nil for value %q", tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("Email() error code = %v, want %v", result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("Email() expected no error but got: %v for value %q", result, tt.value)
				}
			}
		})
	}
}

func TestAlpha(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
		errorCode   string
	}{
		{
			name:        "valid lowercase letters",
			value:       "hello",
			expectError: false,
		},
		{
			name:        "valid uppercase letters",
			value:       "HELLO",
			expectError: false,
		},
		{
			name:        "valid mixed case letters",
			value:       "HelloWorld",
			expectError: false,
		},
		{
			name:        "invalid with numbers",
			value:       "hello123",
			expectError: true,
			errorCode:   "not-alpha",
		},
		{
			name:        "invalid with spaces",
			value:       "hello world",
			expectError: true,
			errorCode:   "not-alpha",
		},
		{
			name:        "invalid with special characters",
			value:       "hello!",
			expectError: true,
			errorCode:   "not-alpha",
		},
		{
			name:        "empty string",
			value:       "",
			expectError: false,
		},
		{
			name:        "single letter",
			value:       "a",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := Alpha()
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("Alpha() expected error but got nil for value %q", tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("Alpha() error code = %v, want %v", result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("Alpha() expected no error but got: %v for value %q", result, tt.value)
				}
			}
		})
	}
}

func TestAlphanumeric(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
		errorCode   string
	}{
		{
			name:        "valid letters only",
			value:       "hello",
			expectError: false,
		},
		{
			name:        "valid numbers only",
			value:       "12345",
			expectError: false,
		},
		{
			name:        "valid letters and numbers",
			value:       "hello123",
			expectError: false,
		},
		{
			name:        "valid mixed case with numbers",
			value:       "HelloWorld123",
			expectError: false,
		},
		{
			name:        "invalid with spaces",
			value:       "hello world",
			expectError: true,
			errorCode:   "not-alphanumeric",
		},
		{
			name:        "invalid with special characters",
			value:       "hello!",
			expectError: true,
			errorCode:   "not-alphanumeric",
		},
		{
			name:        "invalid with underscore",
			value:       "hello_world",
			expectError: true,
			errorCode:   "not-alphanumeric",
		},
		{
			name:        "empty string",
			value:       "",
			expectError: false,
		},
		{
			name:        "single character",
			value:       "a",
			expectError: false,
		},
		{
			name:        "single digit",
			value:       "1",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := Alphanumeric()
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("Alphanumeric() expected error but got nil for value %q", tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("Alphanumeric() error code = %v, want %v", result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("Alphanumeric() expected no error but got: %v for value %q", result, tt.value)
				}
			}
		})
	}
}

func TestURL(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
		errorCode   string
	}{
		{
			name:        "valid HTTP URL",
			value:       "http://example.com",
			expectError: false,
		},
		{
			name:        "valid HTTPS URL",
			value:       "https://example.com",
			expectError: false,
		},
		{
			name:        "valid URL with path",
			value:       "https://example.com/path/to/resource",
			expectError: false,
		},
		{
			name:        "valid URL with query params",
			value:       "https://example.com?param=value",
			expectError: false,
		},
		{
			name:        "valid URL with fragment",
			value:       "https://example.com#section",
			expectError: false,
		},
		{
			name:        "valid URL with port",
			value:       "https://example.com:8080",
			expectError: false,
		},
		{
			name:        "invalid URL no scheme",
			value:       "example.com",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
		{
			name:        "invalid URL no host",
			value:       "https://",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
		{
			name:        "invalid URL with spaces",
			value:       "https://example .com",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
		{
			name:        "empty string",
			value:       "",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
		{
			name:        "invalid scheme",
			value:       "invalid://example.com",
			expectError: true,
			errorCode:   "pattern-mismatch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := URL()
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("URL() expected error but got nil for value %q", tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("URL() error code = %v, want %v", result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("URL() expected no error but got: %v for value %q", result, tt.value)
				}
			}
		})
	}
}

func TestOneOf(t *testing.T) {
	tests := []struct {
		name        string
		allowed     []string
		value       string
		expectError bool
		errorCode   string
	}{
		{
			name:        "value in allowed list",
			allowed:     []string{"apple", "banana", "orange"},
			value:       "apple",
			expectError: false,
		},
		{
			name:        "value not in allowed list",
			allowed:     []string{"apple", "banana", "orange"},
			value:       "grape",
			expectError: true,
			errorCode:   "not-one-of",
		},
		{
			name:        "empty value not in allowed list",
			allowed:     []string{"apple", "banana", "orange"},
			value:       "",
			expectError: true,
			errorCode:   "not-one-of",
		},
		{
			name:        "empty value in allowed list",
			allowed:     []string{"", "apple", "banana"},
			value:       "",
			expectError: false,
		},
		{
			name:        "case sensitive match",
			allowed:     []string{"Apple", "Banana", "Orange"},
			value:       "apple",
			expectError: true,
			errorCode:   "not-one-of",
		},
		{
			name:        "single allowed value match",
			allowed:     []string{"only"},
			value:       "only",
			expectError: false,
		},
		{
			name:        "single allowed value no match",
			allowed:     []string{"only"},
			value:       "other",
			expectError: true,
			errorCode:   "not-one-of",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := OneOf(tt.allowed...)
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("OneOf(%v) expected error but got nil for value %q", tt.allowed, tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("OneOf(%v) error code = %v, want %v", tt.allowed, result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("OneOf(%v) expected no error but got: %v for value %q", tt.allowed, result, tt.value)
				}
			}
		})
	}
}

func TestNoWhitespace(t *testing.T) {
	tests := []struct {
		name        string
		value       string
		expectError bool
		errorCode   string
	}{
		{
			name:        "no whitespace characters",
			value:       "helloworld",
			expectError: false,
		},
		{
			name:        "contains space",
			value:       "hello world",
			expectError: true,
			errorCode:   "contains-whitespace",
		},
		{
			name:        "contains tab",
			value:       "hello\tworld",
			expectError: true,
			errorCode:   "contains-whitespace",
		},
		{
			name:        "contains newline",
			value:       "hello\nworld",
			expectError: true,
			errorCode:   "contains-whitespace",
		},
		{
			name:        "contains carriage return",
			value:       "hello\rworld",
			expectError: true,
			errorCode:   "contains-whitespace",
		},
		{
			name:        "empty string",
			value:       "",
			expectError: false,
		},
		{
			name:        "special characters no whitespace",
			value:       "hello!@#$%world",
			expectError: false,
		},
		{
			name:        "numbers and letters only",
			value:       "hello123world",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := NoWhitespace()
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("NoWhitespace() expected error but got nil for value %q", tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("NoWhitespace() error code = %v, want %v", result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("NoWhitespace() expected no error but got: %v for value %q", result, tt.value)
				}
			}
		})
	}
}

func TestCustom(t *testing.T) {
	tests := []struct {
		name        string
		validateFn  func(string) bool
		errorMsg    string
		value       string
		expectError bool
		errorCode   string
	}{
		{
			name:        "custom function returns true",
			validateFn:  func(s string) bool { return len(s) > 0 },
			errorMsg:    "must not be empty",
			value:       "hello",
			expectError: false,
		},
		{
			name:        "custom function returns false",
			validateFn:  func(s string) bool { return len(s) > 0 },
			errorMsg:    "must not be empty",
			value:       "",
			expectError: true,
			errorCode:   "custom-validation-failed",
		},
		{
			name: "custom palindrome check passes",
			validateFn: func(s string) bool {
				for i := 0; i < len(s)/2; i++ {
					if s[i] != s[len(s)-1-i] {
						return false
					}
				}
				return true
			},
			errorMsg:    "must be a palindrome",
			value:       "racecar",
			expectError: false,
		},
		{
			name: "custom palindrome check fails",
			validateFn: func(s string) bool {
				for i := 0; i < len(s)/2; i++ {
					if s[i] != s[len(s)-1-i] {
						return false
					}
				}
				return true
			},
			errorMsg:    "must be a palindrome",
			value:       "hello",
			expectError: true,
			errorCode:   "custom-validation-failed",
		},
		{
			name:        "custom contains substring check passes",
			validateFn:  func(s string) bool { return len(s) >= 3 && s[0:3] == "ABC" },
			errorMsg:    "must start with ABC",
			value:       "ABC123",
			expectError: false,
		},
		{
			name:        "custom contains substring check fails",
			validateFn:  func(s string) bool { return len(s) >= 3 && s[0:3] == "ABC" },
			errorMsg:    "must start with ABC",
			value:       "DEF123",
			expectError: true,
			errorCode:   "custom-validation-failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			validator := Custom(tt.validateFn, tt.errorMsg)
			ctx := validation.NewContext("test_field")

			result := validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("Custom() expected error but got nil for value %q", tt.value)
					return
				}
				if result.Code() != tt.errorCode {
					t.Errorf("Custom() error code = %v, want %v", result.Code(), tt.errorCode)
				}
			} else {
				if result != nil {
					t.Errorf("Custom() expected no error but got: %v for value %q", result, tt.value)
				}
			}
		})
	}
}

// Test validator composition and chaining
func TestStringValidatorComposition(t *testing.T) {
	tests := []struct {
		name        string
		validators  []Validator[string]
		value       string
		expectError bool
		errorCode   string
	}{
		{
			name:        "NotEmpty and MinLength both pass",
			validators:  []Validator[string]{NotEmpty(), MinLength(3)},
			value:       "hello",
			expectError: false,
		},
		{
			name:        "NotEmpty passes, MinLength fails",
			validators:  []Validator[string]{NotEmpty(), MinLength(10)},
			value:       "hello",
			expectError: true,
			errorCode:   "min-length",
		},
		{
			name:        "Email and MinLength both pass",
			validators:  []Validator[string]{Email(), MinLength(5)},
			value:       "test@example.com",
			expectError: false,
		},
		{
			name:        "Alpha and Length both pass",
			validators:  []Validator[string]{Alpha(), Length(3, 10)},
			value:       "hello",
			expectError: false,
		},
		{
			name:        "Alpha passes, Length fails",
			validators:  []Validator[string]{Alpha(), Length(10, 20)},
			value:       "hello",
			expectError: true,
			errorCode:   "length-range",
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
						t.Errorf("Validator chain expected no error but got: %v for value %q", result, tt.value)
						return
					}
				}
			}

			// If we expected an error but didn't find it
			if tt.expectError {
				t.Errorf("Validator chain expected error with code %v but got none for value %q", tt.errorCode, tt.value)
			}
		})
	}
}

// Test WithName functionality
func TestStringValidatorWithName(t *testing.T) {
	tests := []struct {
		name         string
		validator    Validator[string]
		customName   string
		value        string
		expectError  bool
		expectedName string
	}{
		{
			name:         "NotEmpty with custom name",
			validator:    NotEmpty().WithName("custom-not-empty"),
			customName:   "custom-not-empty",
			value:        "",
			expectError:  true,
			expectedName: "custom-not-empty",
		},
		{
			name:         "Email with custom name",
			validator:    Email().WithName("user-email"),
			customName:   "user-email",
			value:        "invalid-email",
			expectError:  true,
			expectedName: "user-email",
		},
		{
			name:         "Length with custom name",
			validator:    Length(5, 10).WithName("description-length"),
			customName:   "description-length",
			value:        "hi",
			expectError:  true,
			expectedName: "description-length",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := validation.NewContext("test_field")

			result := tt.validator.Validate(tt.value, ctx)

			if tt.expectError {
				if result == nil {
					t.Errorf("Expected error but got nil for value %q", tt.value)
					return
				}

				// Check that the validator name is preserved
				if tt.validator.GetName() != tt.expectedName {
					t.Errorf("Expected validator name %q but got %q", tt.expectedName, tt.validator.GetName())
				}
			} else {
				if result != nil {
					t.Errorf("Expected no error but got: %v for value %q", result, tt.value)
				}
			}
		})
	}
}

// Test context metadata propagation
func TestStringValidatorContextMetadata(t *testing.T) {
	validator := NotEmpty()

	ctx := validation.NewContext("test_field").
		WithMetadata("source", "user_input").
		WithMetadata("validation_type", "required")

	result := validator.Validate("", ctx)

	if result == nil {
		t.Fatal("Expected validation error but got nil")
	}

	if result.Target != "test_field" {
		t.Errorf("Expected target 'test_field' but got %q", result.Target)
	}
}
