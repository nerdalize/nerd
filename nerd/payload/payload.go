package payload

import "github.com/go-playground/validator"

//Validator allows validation of payloads with custom errors and/or types
type Validator struct {
	v *validator.Validate
}

//NewValidator sets up the payload validator
func NewValidator() *Validator {
	return &Validator{v: validator.New()}
}

//Validate takes a payload and returns nil or a validation error
func (v *Validator) Validate(payload interface{}) error {
	return v.v.Struct(payload)
}
