package app

import "github.com/go-playground/validator/v10"

// customValidator は Echo の c.Validate() を go-playground/validator に接続する。
type customValidator struct {
	validator *validator.Validate
}

func newValidator() *customValidator {
	return &customValidator{validator: validator.New()}
}

func (cv *customValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}
