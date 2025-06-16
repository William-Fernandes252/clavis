package validation

import "fmt"

type KeyValidator func(string) error

type ValueValidator func(string, []byte) error

var ValidateNonEmptyKey KeyValidator = func(key string) error {
	if key == "" {
		return fmt.Errorf("key cannot be empty")
	}
	return nil
}

func ValidateKeyLength(maxLength int) KeyValidator {
	return func(key string) error {
		if len(key) > maxLength {
			return fmt.Errorf("key too long: maximum %d characters, got %d", maxLength, len(key))
		}
		return nil
	}
}

func ValidateValueSize(maxSize int64) ValueValidator {
	return func(key string, value []byte) error {
		if int64(len(value)) > maxSize {
			return fmt.Errorf("value too large: maximum %d bytes, got %d", maxSize, len(value))
		}
		return nil
	}
}

// ComposeKeyValidators combines multiple key validation functions
func ComposeKeyValidators(validators ...KeyValidator) KeyValidator {
	return func(key string) error {
		for _, validator := range validators {
			if err := validator(key); err != nil {
				return err
			}
		}
		return nil
	}
}

// ComposeValueValidators combines multiple value validation functions
func ComposeValueValidators(validators ...func(string, []byte) error) func(string, []byte) error {
	return func(key string, value []byte) error {
		for _, validator := range validators {
			if err := validator(key, value); err != nil {
				return err
			}
		}
		return nil
	}
}
