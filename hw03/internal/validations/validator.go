package validations

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type CustomValidator struct {
	Validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	v := validator.New()
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 1)[0]
		if name == "-" {
			return ""
		}
		return name
	})
	return &CustomValidator{Validator: v}
}

func (cv *CustomValidator) Validate(i any) error {
	return cv.Validator.Struct(i) //nolint:wrapcheck
}

func FormatValidationErrors(err error) map[string]string {
	var validationErrors validator.ValidationErrors

	if !errors.As(err, &validationErrors) {
		return map[string]string{"body": "Invalid request format"}
	}

	errorMap := make(map[string]string)

	for _, fieldError := range validationErrors {
		fieldName := fieldError.Field()

		switch fieldError.Tag() {
		case "required":
			errorMap[fieldName] = "This field is required."
		case "uuid":
			errorMap[fieldName] = "ID is not a valid UUID."
		case "datetime":
			errorMap[fieldName] = "Invalid date formats."
		case "gte":
			var capitalized string
			if len(fieldError.Field()) > 0 {
				capitalized = strings.ToUpper(fieldError.Field()[:1]) + fieldError.Field()[1:]
			}
			errorMap[fieldName] = capitalized + " must be non-negative integer."
		case "lte":
			errorMap[fieldName] = "Value must be less than or equal to " + fieldError.Param()
		default:
			errorMap[fieldName] = "Invalid value."
		}
	}

	return errorMap
}
