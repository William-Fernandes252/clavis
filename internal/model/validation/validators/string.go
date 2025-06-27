package validators

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"

	"github.com/William-Fernandes252/clavis/internal/model/validation"
)

type StringValidator = validation.Validator[string]

// NotEmpty validates that a string is not empty
func NotEmpty() StringValidator {
	return validation.NewValidator(func(value string, ctx validation.Context) *validation.ValidationError {
		if strings.TrimSpace(value) == "" {
			msg := fmt.Sprintf("%s: must not be empty", ctx.Target)
			err := validation.NewValidationError(
				ctx.Target,
				value,
				msg,
			).WithCode("not-empty")
			return err
		}
		return nil
	})
}

// Length validates string length is within the specified range
func Length(min, max int) StringValidator {
	return validation.NewValidator(func(value string, ctx validation.Context) *validation.ValidationError {
		length := len(value)
		if length < min || length > max {
			msg := fmt.Sprintf("%s: must be between %d and %d characters, got %d",
				ctx.Target, min, max, length)
			err := validation.NewValidationError(
				ctx.Target,
				value,
				msg,
			).WithMetadata("min-length", min).
				WithMetadata("max-length", max).
				WithMetadata("actual-length", length).
				WithCode("length-range")
			return err
		}
		return nil
	})
}

// MinLength validates minimum string length
func MinLength(min int) StringValidator {
	return validation.NewValidator(func(value string, ctx validation.Context) *validation.ValidationError {
		if len(value) < min {
			msg := fmt.Sprintf("%s: must be at least %d characters, got %d",
				ctx.Target, min, len(value))
			err := validation.NewValidationError(
				ctx.Target,
				value,
				msg,
			).WithMetadata("min", min).
				WithMetadata("actual", len(value)).
				WithCode("min-length")
			return err
		}
		return nil
	})
}

// MaxLength validates maximum string length
func MaxLength(max int) StringValidator {
	return validation.NewValidator(func(value string, ctx validation.Context) *validation.ValidationError {
		if len(value) > max {
			msg := fmt.Sprintf("%s: must be at most %d characters, got %d",
				ctx.Target, max, len(value))
			err := validation.NewValidationError(
				ctx.Target,
				value,
				msg,
			).WithMetadata("max", max).
				WithMetadata("actual", len(value)).
				WithCode("max-length")
			return err
		}
		return nil
	})
}

// Pattern validates string against a regular expression
func Pattern(pattern string) StringValidator {
	regex := regexp.MustCompile(pattern)
	return validation.NewValidator(func(value string, ctx validation.Context) *validation.ValidationError {
		if !regex.MatchString(value) {
			msg := fmt.Sprintf("%s: does not match required pattern", ctx.Target)
			err := validation.NewValidationError(
				ctx.Target,
				value,
				msg,
			).WithMetadata("pattern", pattern).WithCode("pattern-mismatch")
			return err
		}
		return nil
	})
}

// Email validates email format (reuses Pattern logic)
func Email() StringValidator {
	return Pattern(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
}

// URL validates URL format (reuses Pattern logic)
func URL() StringValidator {
	return Pattern(`^https?://[^\s/$.?#].[^\s]*$`)
}

// OneOf validates that the string is one of the allowed values
func OneOf(allowed ...string) StringValidator {
	allowedSet := make(map[string]bool)
	for _, val := range allowed {
		allowedSet[val] = true
	}

	return validation.NewValidator(func(value string, ctx validation.Context) *validation.ValidationError {
		if !allowedSet[value] {
			msg := fmt.Sprintf("%s: must be one of %s",
				ctx.Target, strings.Join(allowed, ", "))
			err := validation.NewValidationError(
				ctx.Target,
				value,
				msg,
			).WithMetadata("allowed-values", allowed).WithCode("not-one-of")
			return err
		}
		return nil
	})
}

// NoWhitespace validates that the string contains no whitespace
func NoWhitespace() StringValidator {
	return validation.NewValidator(func(value string, ctx validation.Context) *validation.ValidationError {
		for _, r := range value {
			if unicode.IsSpace(r) {
				msg := fmt.Sprintf("%s: cannot contain whitespace", ctx.Target)
				err := validation.NewValidationError(
					ctx.Target,
					value,
					msg,
				).WithCode("contains-whitespace")
				return err
			}
		}
		return nil
	})
}

// Alpha validates that the string contains only alphabetic characters
func Alpha() StringValidator {
	return validation.NewValidator(func(value string, ctx validation.Context) *validation.ValidationError {
		for _, r := range value {
			if !unicode.IsLetter(r) {
				msg := fmt.Sprintf("%s: must contain only alphabetic characters", ctx.Target)
				err := validation.NewValidationError(
					ctx.Target,
					value,
					msg,
				).WithCode("not-alpha")
				return err
			}
		}
		return nil
	})
}

// Alphanumeric validates that the string contains only alphanumeric characters
func Alphanumeric() StringValidator {
	return validation.NewValidator(func(value string, ctx validation.Context) *validation.ValidationError {
		for _, r := range value {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
				msg := fmt.Sprintf("%s: must contain only alphanumeric characters", ctx.Target)
				err := validation.NewValidationError(
					ctx.Target,
					value,
					msg,
				).WithCode("not-alphanumeric")
				return err
			}
		}
		return nil
	})
}

// Custom allows creating custom string validators
func Custom(validateFn func(value string) bool, errorMsg string) StringValidator {
	return validation.NewValidator(func(value string, ctx validation.Context) *validation.ValidationError {
		if !validateFn(value) {
			msg := fmt.Sprintf("%s: %s", ctx.Target, errorMsg)
			err := validation.NewValidationError(
				ctx.Target,
				value,
				msg,
			).WithCode("custom-validation-failed")
			return err
		}
		return nil
	})
}
