package web

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/starclusterteam/go-starbox/apm"
)

// WriteJSON encodes given value to JSON and responds with <code>
func WriteJSON(w http.ResponseWriter, code int, value interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	valueJSON, _ := json.Marshal(value)
	w.Write(valueJSON)
}

// HandleError logs the error then responds with generic 500 message
func HandleError(w http.ResponseWriter, req *http.Request, err error) {
	apm.GlobalReporter.Report(req.Context(), err)

	WriteJSON(w, http.StatusInternalServerError, &ErrorResponse{
		Messages: []string{"Something went wrong, error has been reported and we'll look into it as soon as possible"},
		Errors: map[string][]string{
			"error": {"internal server error"},
		},
	})
}

// blacklistHeaders takes a map of headers and returns a map without
// the header matching "X-Vulcand"
// TODO: make it configurable
func blacklistHeaders(headers map[string]string) map[string]string {
	res := make(map[string]string)
	for k, v := range headers {
		if !strings.Contains(strings.ToLower(k), strings.ToLower("X-Vulcand")) {
			res[k] = v
		}
	}
	return res
}
