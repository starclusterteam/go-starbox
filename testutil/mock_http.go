package testutil

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"

	. "github.com/onsi/gomega"
)

// NewRequest is like http.NewRequest, but returns no error.
func NewRequest(method, url string, body io.Reader) *http.Request {
	req, err := http.NewRequest(method, url, body)
	Expect(err).To(BeNil(), fmt.Sprintf("Bug in test: cannot construct http.Request from method=%q, url=%q, body=%#v: %s", method, url, body, err))
	return req
}

// CheckResponse checks that the http response matches expectations
func CheckResponse(w *httptest.ResponseRecorder, wantStatus int, wantHeaders http.Header, wantBody string) {
	Expect(w.Code).To(Equal(wantStatus), "Bad HTTP status")
	Expect(w.HeaderMap).To(Equal(wantHeaders), "Bad HTTP response headers")
	Expect(w.Body.String()).To(Equal(wantBody), "Bad HTTP response body")
}

// CheckResponseBody checks that the http response matches expectations
func CheckResponseBody(w *httptest.ResponseRecorder, wantStatus int, wantBody string) {
	Expect(w.Code).To(Equal(wantStatus), "Bad HTTP status")
	Expect(w.Body.String()).To(Equal(wantBody), "Bad HTTP response body")
}
