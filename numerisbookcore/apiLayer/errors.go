package apiLayer

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"strings"
)

type ValidationError struct {
	Errors map[string]interface{}
}

func NewValidationError(validationErrors validator.ValidationErrors) *ValidationError {
	errors := make(map[string]interface{})
	for _, err := range validationErrors {
		fmt.Println(err.Field())

		field := err.Field()
		tag := err.Tag()
		message := fmt.Sprintf("%s is %s", field, tag)
		errors[field] = message
	}
	fmt.Println("errors >>>. ", errors)
	return &ValidationError{Errors: errors}
}

func (ve *ValidationError) Error() string {
	var errorMsgs []string
	for _, message := range ve.Errors {
		errorMsgs = append(errorMsgs, fmt.Sprintf("%s", message))
	}
	return strings.Join(errorMsgs, "; ")
}
