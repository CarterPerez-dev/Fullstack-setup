// AngelaMos | 2026
// validation.go

package core

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

func FormatValidationError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		messages := make([]string, 0, len(ve))
		for _, fe := range ve {
			messages = append(messages, FormatFieldError(fe))
		}
		return strings.Join(messages, "; ")
	}
	return "validation failed"
}

func FormatFieldError(fe validator.FieldError) string {
	field := strings.ToLower(fe.Field())

	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "min":
		return field + " must be at least " + fe.Param() + " characters"
	case "max":
		return field + " must be at most " + fe.Param() + " characters"
	case "oneof":
		return field + " must be one of: " + fe.Param()
	default:
		return field + " is invalid"
	}
}
