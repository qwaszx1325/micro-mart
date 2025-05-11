package helper

import (
	"github.com/go-playground/validator/v10"
	"regexp"
)

// VerifiedCodeValidation is a custom validator for verified code
func VerifiedCodeValidation(fl validator.FieldLevel) bool {
	match, _ := regexp.MatchString(`^[A-Z]{3}-[0-9]{6}$`, fl.Field().String())
	return match
}

func ContainsAtLeastOneAlpha(fl validator.FieldLevel) bool {
	match, _ := regexp.MatchString(`[a-zA-Z]{1,}`, fl.Field().String())
	return match
}

func ContainsAtLeastOneNum(fl validator.FieldLevel) bool {
	match, _ := regexp.MatchString(`[0-9]{1,}`, fl.Field().String())
	return match
}
