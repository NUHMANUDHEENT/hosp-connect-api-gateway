package utils

import (
	"strings"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateInput function checks for validation errors in the input struct
func ValidateInput(input interface{}) (bool, string) {
	err := validate.Struct(input)
	if err != nil {
		var errorMessages []string
		for _, err := range err.(validator.ValidationErrors) {
			errorMessages = append(errorMessages, err.Field()+" is required or invalid")
		}
		return false, strings.Join(errorMessages, ", ")
	}
	return true, ""
}
