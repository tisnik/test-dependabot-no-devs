package request

import (
	"errors"

	"github.com/course-go/reelgoofy/internal/core/response"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// ValidatedStruct is a custom interface, which represents structs
// that can be validated in request_validator.
type ValidatedStruct interface {
	CanBeValidated()
}

// ValidateStruct performs a validation on struct and return a map of custom errors.
func ValidateStruct(s ValidatedStruct) map[string]string {
	validate := validator.New()

	err := validate.Struct(s)
	if err == nil {
		return nil
	}

	errs := make(map[string]string)
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, e := range ve {
			switch e.Tag() {
			case "required":
				errs[e.Field()] = string(response.RequiredFieldMessage)
			case "uuid4":
				errs[e.Field()] = string(response.InvalidUUIDMessage)
			case "datetime":
				errs[e.Field()] = string(response.InvalidDateTimeFormatMessage)
			case "min":
				errs[e.Field()] = string(response.ValueTooSmallMessage)
			case "max":
				errs[e.Field()] = string(response.ValueTooLargeMessage)
			default:
				errs[e.Field()] = string(response.InvalidValueMessage)
			}
		}
	}

	return errs
}

// ValidateUuid checks whether a string representation of UUID is valid
// and returns a map of custom errors along with valid UUID.
func ValidateUuid(id string, name string) (uuid.UUID, map[string]string) {
	parsedUuid, err := uuid.Parse(id)
	if err != nil {
		errMap := map[string]string{
			name: string(response.InvalidUUIDMessage),
		}

		return uuid.Nil, errMap
	}

	return parsedUuid, nil
}
