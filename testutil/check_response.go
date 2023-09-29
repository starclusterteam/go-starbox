package testutil

import "net/http/httptest"

// CheckUnauthorized checks if response is web.CheckUnauthorized
func CheckUnauthorized(w *httptest.ResponseRecorder) {
	CheckResponseBody(w, 401, "{\"messages\":[\"You must be logged in to access this page\"],\"code\":105,\"errors\":{\"user\":[\"unauthorized\"]}}")
}

// CheckServiceUnavailable checks if response is web.ServiceUnavailable
func CheckServiceUnavailable(w *httptest.ResponseRecorder) {
	CheckResponseBody(w, 503, "{\"messages\":[\"There was an error, please try again later\"],\"code\":503,\"errors\":{\"error\":[\"service unavailable\"]}}")
}
