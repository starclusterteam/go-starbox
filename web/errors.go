package web

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/starclusterteam/go-starbox/errors"
)

// ResponseError defines an interface for HTTP error handling, using the custom format:
//
//	{
//		"messages": ["a message", "another message"],
//		"errors": {
//			"thing 1": ["value 1", "value 2"],
//			"thing 2": ["value 3"]
//		}
//	}
type ResponseError interface {
	AddMessage(message string) ResponseError
	With(key string, values ...string) ResponseError
	Write(w http.ResponseWriter)
}

type responseError struct {
	statusCode int
	errorCode  int
	messages   []string
	errors     map[string][]string
}

func (e *responseError) With(k string, vs ...string) ResponseError {
	if e.errors == nil {
		e.errors = make(map[string][]string)
	}

	e.errors[k] = vs
	return e
}

func (e *responseError) AddMessage(m string) ResponseError {
	e.messages = append(e.messages, m)
	return e
}

func (e *responseError) Write(w http.ResponseWriter) {
	resp := ErrorResponse{
		Errors:   e.errors,
		Messages: e.messages,
		Code:     e.errorCode,
	}

	WriteJSON(w, e.statusCode, &resp)
}

// Error creates a ResponseError that also writes the error code to the response writer.
func Error(statusCode, errCode int) ResponseError {
	return &responseError{statusCode: statusCode, errorCode: errCode}
}

// BadRequest writes to the ResponseWriter a bad request error.
func BadRequest(w http.ResponseWriter, errCode int, field string, messages ...string) {
	genericErrorResponse(w, http.StatusBadRequest, errCode, field, messages...)
}

func PreconditionFailed(w http.ResponseWriter, errCode int, field string, messages ...string) {
	genericErrorResponse(w, http.StatusPreconditionFailed, errCode, field, messages...)
}

func PayloadTooLarge(w http.ResponseWriter, errCode int, field string, messages ...string) {
	genericErrorResponse(w, http.StatusRequestEntityTooLarge, errCode, field, messages...)
}

// UnsupportedMediaType writes to the ResponseWriter an unsupported media type error.
func UnsupportedMediaType(w http.ResponseWriter, errCode int, field string, messages ...string) {
	genericErrorResponse(w, http.StatusUnsupportedMediaType, errCode, field, messages...)
}

func genericErrorResponse(w http.ResponseWriter, httpStatusCode int, errCode int, field string, messages ...string) {
	e := Error(httpStatusCode, errCode)

	for _, m := range messages {
		if field != "" {
			e.AddMessage(fmt.Sprintf("%s %s", field, m))
		} else {
			e.AddMessage(m)
		}
	}

	if field != "" {
		e.With(field, messages...)
	}
	e.Write(w)
}

// NotFound writes to the ResponseWriter a not found error.
func NotFound(w http.ResponseWriter, errCode int, resources ...string) {
	e := Error(http.StatusNotFound, errCode)

	for _, r := range resources {
		e.AddMessage(fmt.Sprintf("%s not found", r))
		e.With(r, "not found")
	}

	e.Write(w)
}

// ErrorResponse defines the response structure for all 4xx and 5xx errors
type ErrorResponse struct {
	Messages []string            `json:"messages"`
	Code     int                 `json:"code"`
	Errors   map[string][]string `json:"errors"`

	HTTPStatus int `json:"-"`

	isInternalError bool
	internalError   error
}

func (e *ErrorResponse) Write(w http.ResponseWriter) {
	WriteJSON(w, e.HTTPStatus, e)
}

func (e *ErrorResponse) IsInternalError() bool {
	return e.isInternalError
}

func (e *ErrorResponse) InternalError() error {
	return e.internalError
}

var (
	ErrMissingAuthorizationHeader = NewError("Missing authorization header", 101, http.StatusUnauthorized)
	ErrInvalidAuthorizationHeader = NewError("Invalid authorization header", 102, http.StatusUnauthorized)
	ErrInvalidAccessToken         = NewError("Invalid access token", 103, http.StatusUnauthorized)
	ErrInvalidCredentials         = NewError("Invalid credentials", 104, http.StatusUnauthorized)
	ErrUnauthorized               = NewError("Unauthorized", 105, http.StatusUnauthorized)

	ErrInvalidRequestFormat = NewError("Invalid request format", 201, http.StatusBadRequest)
)

func NewError(message string, code, httpStatus int) *ErrorResponse {
	return &ErrorResponse{
		Messages:   []string{message},
		Code:       code,
		HTTPStatus: httpStatus,
		Errors: map[string][]string{
			"error": {message},
		},
	}
}

// NewValidationError returns a new error response with 400 status code
func NewBadRequest(message string) *ErrorResponse {
	return &ErrorResponse{
		Messages:   []string{message},
		Code:       400,
		HTTPStatus: http.StatusBadRequest,
		Errors: map[string][]string{
			"bad_request": {message},
		},
	}
}

func NewBadRequestNotFound(field, resource string) *ErrorResponse {
	return &ErrorResponse{
		Messages:   []string{resource + " not found"},
		Code:       400,
		HTTPStatus: http.StatusBadRequest,
		Errors: map[string][]string{
			field: {resource + " not found"},
		},
	}
}

func NewLoginRequired() *ErrorResponse {
	return &ErrorResponse{
		Messages:   []string{"You must be logged in to access this page"},
		Code:       105,
		HTTPStatus: http.StatusUnauthorized,
		Errors: map[string][]string{
			"user": []string{"unauthorized"},
		},
	}
}

// NewForbidden returns a new error response with 403 status code
func NewForbidden(message string) *ErrorResponse {
	return &ErrorResponse{
		Messages:   []string{message},
		Code:       403,
		HTTPStatus: http.StatusForbidden,
		Errors: map[string][]string{
			"forbidden": {message},
		},
	}
}

// NewNotFound returns a new error response with 404 status code
func NewNotFound(message string) *ErrorResponse {
	return &ErrorResponse{
		Messages:   []string{message},
		Code:       404,
		HTTPStatus: http.StatusNotFound,
		Errors: map[string][]string{
			"not_found": {message},
		},
	}
}

// NewConflict returns a new error response with 409 status code
func NewConflict(message string) *ErrorResponse {
	return &ErrorResponse{
		Messages:   []string{message},
		Code:       409,
		HTTPStatus: http.StatusConflict,
		Errors: map[string][]string{
			"conflict": {message},
		},
	}
}

func NewInternalError(err error) *ErrorResponse {
	return &ErrorResponse{
		Messages:        []string{},
		Code:            500,
		HTTPStatus:      http.StatusInternalServerError,
		Errors:          map[string][]string{},
		isInternalError: true,
		internalError:   err,
	}
}

func NewServiceUnavailable() *ErrorResponse {
	return &ErrorResponse{
		Messages: []string{"There was an error, please try again later"},
		Errors: map[string][]string{
			"error": []string{"service unavailable"},
		},
		HTTPStatus: http.StatusServiceUnavailable,
		Code:       503,
	}
}

func NewValidationErrorsResponse(errs []ValidationError) *ErrorResponse {
	messages := make([]string, len(errs))
	errors := make(map[string][]string)

	for i, err := range errs {
		messages[i] = err.Message
		errors[err.Field] = append(errors[err.Field], err.Message)
	}

	return &ErrorResponse{
		Messages:   messages,
		Code:       400,
		HTTPStatus: http.StatusBadRequest,
		Errors:     errors,
	}
}

func HandleErrorResponse(w http.ResponseWriter, req *http.Request, err *ErrorResponse) {
	if err.isInternalError {
		HandleError(w, req, err.internalError)
		return
	}

	WriteJSON(w, err.HTTPStatus, err)
}

func ParseErrorResponse(body []byte) (*ErrorResponse, error) {
	e := &ErrorResponse{}
	if err := json.Unmarshal(body, e); err != nil {
		return nil, errors.Wrap(err, "failed to parse response")
	}

	if len(e.Messages) == 0 {
		return nil, errors.New("missing required fields in error response: messages, errors")
	}

	if len(e.Errors) == 0 {
		return nil, errors.New("missing required fields in error response: messages, errors")
	}

	if e.HTTPStatus != 0 {
		return nil, errors.New("http status should not be provided in response: http_status")
	}

	return e, nil
}
