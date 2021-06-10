package web

import (
	"fmt"
	"net/http"
)

// ResponseError defines an interface for HTTP error handling, using the custom format:
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
	e := Error(http.StatusBadRequest, errCode)

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
	e := Error(http.StatusNotAcceptable, errCode)

	for _, r := range resources {
		e.AddMessage(fmt.Sprintf("%s not found", r))
		e.With(r, "not found")
	}

	e.Write(w)
}

// ErrorResponse defines the response structure for all 4xx and 5xx errors
type ErrorResponse struct {
	Messages []string            `json:"messages"`
	Errors   map[string][]string `json:"errors"`
	Code     int                 `json:"code,omitempty"`
}

// LoginRequired renders an "Unauthorized error" to user
// Example:
// {
//   "messages": ["You must be logged in to access this page"],
//   "errors": {
//     "user": ["unauthorized"]
//   }
// }
func LoginRequired(w http.ResponseWriter) {
	response := &ErrorResponse{
		Messages: []string{"You must be logged in to access this page"},
		Errors: map[string][]string{
			"user": []string{"unauthorized"},
		},
	}
	WriteJSON(w, http.StatusUnauthorized, response)
}

// ServiceUnavailable renders a 503 error to user
// Example:
// {
//   "messages": ["There was an error, please try again later"],
//   "errors": {
//     "error": ["service unavailable"]
//   }
// }
func ServiceUnavailable(w http.ResponseWriter) {
	response := &ErrorResponse{
		Messages: []string{"There was an error, please try again later"},
		Errors: map[string][]string{
			"error": []string{"service unavailable"},
		},
	}
	WriteJSON(w, http.StatusServiceUnavailable, response)
}
