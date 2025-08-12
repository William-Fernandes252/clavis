package commands

import (
	"reflect"
	"strings"

	"github.com/William-Fernandes252/clavis/internal/errors"
	"github.com/go-playground/validator/v10"
)

// Args defines the interface for command arguments.
// It can be any type that implements the Validate method.
// This allows for flexible argument structures while ensuring validation.
// Custom validation messages can be provided using the "error" tag.
type Args any

var validate = validator.New()

// ValidateArgs checks if the provided arguments are valid.
// It calls the Validate method if it exists.
func ValidateArgs(args Args) errors.Error {
	if v, ok := args.(interface{ Validate() errors.Error }); ok {
		return v.Validate()
	}

	if err := validate.Struct(args); err != nil {
		messages := make([]string, 0, len(err.(validator.ValidationErrors)))

		for _, fe := range err.(validator.ValidationErrors) {
			field, _ := reflect.TypeOf(args).Elem().FieldByName(fe.StructField())
			customMsg := field.Tag.Get("error")
			if customMsg != "" {
				messages = append(messages, customMsg)
			} else {
				messages = append(messages, fe.Error())
			}
		}

		return NewInvalidArgumentsError(strings.Join(messages, ", "))
	}

	return nil
}
