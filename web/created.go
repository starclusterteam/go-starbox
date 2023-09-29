package web

import (
	"net/http"
)

func Created(w http.ResponseWriter, _ *http.Request) {
	WriteJSON(w, http.StatusCreated, EmptyResponse{})
}

