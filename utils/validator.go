package utils

import (
	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

// ValidateStruct validates the struct and returns error if there is any invalid values
func ValidateStruct(v interface{}) error {
	return validate.Struct(v)
}
