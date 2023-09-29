package web

import (
	"net/http"
	"strconv"
  "github.com/pkg/errors"

	"github.com/gorilla/mux"
)

func FetchIntVar(r *http.Request, name string) (int, *ErrorResponse) {
	vars := mux.Vars(r)
	value, ok := vars[name]
	if !ok {
		return 0, NewInternalError(errors.Errorf("variable %s not found", name))
	}

	id, err := strconv.Atoi(value)
	if err != nil {
		return 0, ErrInvalidRequestFormat
	}

	return id, nil
}

func FetchStringVar(r *http.Request, name string) (string, *ErrorResponse) {
	vars := mux.Vars(r)
	value, ok := vars[name]
	if !ok {
		return "", NewInternalError(errors.Errorf("variable %s not found", name))
	}

	return value, nil
}

