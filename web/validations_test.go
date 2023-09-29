package web

import (
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type simpleDecode struct {
	Field string `json:"field"`
}

func (s *simpleDecode) Validate() []ValidationError {
	return nil
}

func TestReadJSONSimple(t *testing.T) {
	r := &http.Request{
		Body: io.NopCloser(strings.NewReader(`{"field": "value"}`)),
	}

	var v simpleDecode
	err := ReadJSON(r, &v)

	assert.Nil(t, err)
	assert.Equal(t, "value", v.Field)
}

type simpleDecodeWithValidation struct {
	Field string `json:"field"`
}

func (s *simpleDecodeWithValidation) Validate() []ValidationError {
	var errs []ValidationError

	if s.Field == "" {
		errs = append(errs, NewValidationError("field", "field is required"))
	}

	return errs
}

func TestValidatesJSON(t *testing.T) {
	requests := []string{
		`{}`,
		`{"field": ""}`,
	}

	for _, input := range requests {
		fmt.Printf("Testing input: %s\n", input)
		r := &http.Request{
			Body: io.NopCloser(strings.NewReader(input)),
		}

		var v simpleDecodeWithValidation
		err := ReadJSON(r, &v)

		assert.NotNil(t, err)
		assert.Equal(t, 400, err.HTTPStatus)
		assert.Equal(t, 400, err.Code)
		assert.Equal(t, "field is required", err.Errors["field"][0])
		assert.Equal(t, "field is required", err.Messages[0])
	}
}

type simpleDecodeWithValidationAndOtherFields struct {
	Field string `json:"field"`
	Other *int   `json:"other"`
}

func (s *simpleDecodeWithValidationAndOtherFields) Validate() []ValidationError {
	var errs []ValidationError

	if s.Field == "" {
		errs = append(errs, NewValidationError("field", "field is required"))
	}

	if s.Other == nil {
		errs = append(errs, NewValidationError("other", "other is required"))
	}

	return errs
}

func TestItReturnsInvalidFormatIfFieldIsDifferentType(t *testing.T) {
	requests := []string{
		`{"field": 1, "other": "string"}`,
		`{"field": 1, "other": 1.1}`,
		`{"field": nil, "other": 1}`,
		`{"field": "string, "other": nil}`,
	}

	for _, input := range requests {
		fmt.Printf("Testing input: %s\n", input)
		r := &http.Request{
			Body: io.NopCloser(strings.NewReader(input)),
		}

		var v simpleDecodeWithValidationAndOtherFields
		err := ReadJSON(r, &v)

		assert.NotNil(t, err)
		// fmt.Printf("code=%d http_status=%d errors=%v messages=%v\n", err.Code, err.HTTPStatus, err.Errors, err.Messages)
		assert.Equal(t, 400, err.HTTPStatus)
		assert.Equal(t, 201, err.Code)
		assert.Equal(t, "Invalid request format", err.Errors["error"][0])
		assert.Equal(t, "Invalid request format", err.Messages[0])
	}
}

func TestItParsesCorrectlyWithTwoTypes(t *testing.T) {
	r := &http.Request{
		Body: io.NopCloser(strings.NewReader(`{"field": "string", "other": 1}`)),
	}

	var v simpleDecodeWithValidationAndOtherFields
	err := ReadJSON(r, &v)

	assert.Nil(t, err)
	assert.Equal(t, "string", v.Field)
	assert.Equal(t, 1, *v.Other)
}

type complexInput struct {
	Field    string        `json:"field"`
	Other    *int          `json:"other"`
	SomeBool bool          `json:"some_bool"`
	Double   *float64      `json:"double"`
	Inner    *simpleDecode `json:"inner"`
}

func (s *complexInput) Validate() []ValidationError {
	var errs []ValidationError

	if s.Field == "" {
		errs = append(errs, NewValidationError("field", "field is required"))
	}

	if s.Other == nil {
		errs = append(errs, NewValidationError("other", "other is required"))
	}

	if s.Double == nil {
		errs = append(errs, NewValidationError("double", "double is required"))
	}

	if s.Inner == nil {
		errs = append(errs, NewValidationError("inner", "inner is required"))
	}

	if innerErrs := s.Inner.Validate(); len(innerErrs) > 0 {
		errs = append(errs, innerErrs...)
	}

	return errs
}

func TestComplexInputValidation(t *testing.T) {
	r := &http.Request{
		Body: io.NopCloser(strings.NewReader(`{}`)),
	}

	var v complexInput
	err := ReadJSON(r, &v)

	assert.NotNil(t, err)
	assert.Equal(t, 400, err.HTTPStatus)
	assert.Equal(t, 400, err.Code)

	assert.Equal(t, "field is required", err.Errors["field"][0])
	assert.Equal(t, "other is required", err.Errors["other"][0])
	assert.Equal(t, "double is required", err.Errors["double"][0])
	assert.Equal(t, "inner is required", err.Errors["inner"][0])

	assert.Equal(t, "field is required", err.Messages[0])
	assert.Equal(t, "other is required", err.Messages[1])
	assert.Equal(t, "double is required", err.Messages[2])
	assert.Equal(t, "inner is required", err.Messages[3])
}

func TestComplexInputParsing(t *testing.T) {
	r := &http.Request{
		Body: io.NopCloser(strings.NewReader(`{"field": "string", "other": 1, "some_bool": true, "double": 1.1, "inner": {"field": "otherstring"}}`)),
	}

	var v complexInput
	err := ReadJSON(r, &v)

	assert.Nil(t, err)
	assert.Equal(t, "string", v.Field)
	assert.Equal(t, 1, *v.Other)
	assert.Equal(t, true, v.SomeBool)
	assert.Equal(t, 1.1, *v.Double)
	assert.Equal(t, "otherstring", v.Inner.Field)
}
