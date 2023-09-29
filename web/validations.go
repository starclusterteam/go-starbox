package web

import (
	"encoding/json"
	"net/http"

	"github.com/starclusterteam/go-starbox/errors"
)

type Validatable interface {
	Validate() []ValidationError
}

type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func NewValidationError(field, message string) ValidationError {
	return ValidationError{
		Field:   field,
		Message: message,
	}
}

func ReadJSON(r *http.Request, v Validatable) *ErrorResponse {
	err := DecodeJSON(r, v)
	if err != nil {
		return ErrInvalidRequestFormat
	}

	errs := v.Validate()
	if len(errs) > 0 {
		return NewValidationErrorsResponse(errs)
	}

	return nil
}

func DecodeJSON(r *http.Request, v interface{}) error {
	if r.Body == nil {
		return errors.New("empty request body")
	}

	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(v)
	if err != nil {
		return errors.Wrap(err, "failed to decode json")
	}

	return nil
}
