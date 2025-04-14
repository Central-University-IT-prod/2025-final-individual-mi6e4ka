package validators

import (
	"github.com/go-playground/validator/v10"
)

// кастомный валидатор аналогичный gtefield для полей которые non-required
func GteNRFieldValidator(fl validator.FieldLevel) bool {
	param := fl.Param()
	targetField := fl.Parent().FieldByName(param)
	if !targetField.IsValid() || targetField.IsNil() {
		return true
	}
	fieldValue := fl.Field()

	return fieldValue.Uint() >= targetField.Elem().Uint()
}
