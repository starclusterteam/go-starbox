package testutil

import (
	"net/http"
	"net/http/httptest"
	"strings"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("mock http", func() {
	hello := func(w http.ResponseWriter, req *http.Request) {
		if req.Method != "GET" {
			w.Header().Add("Allow", "GET")
			http.Error(w, "Only GET supported.", http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, world.\n"))
	}

	It("Calls GET correctly", func() {
		handler := http.HandlerFunc(hello)
		req := NewRequest("GET", "http://foo.example.com/bar", nil)
		req.Header.Set("Content-Type", "text/plain; charset=utf-8")
		respw := httptest.NewRecorder()
		handler.ServeHTTP(respw, req)
		wantHdr := make(http.Header)
		wantHdr.Add("Content-Type", "text/plain; charset=utf-8")
		CheckResponse(respw, http.StatusOK, wantHdr, "Hello, world.\n")
	})

	It("Calls PUT correctly", func() {
		handler := http.HandlerFunc(hello)
		body := strings.NewReader(`foo`)
		req := NewRequest("PUT", "http://foo.example.com/bar", body)
		req.Header.Set("Content-Type", "text/plain; charset=utf-8")
		respw := httptest.NewRecorder()
		handler.ServeHTTP(respw, req)
		wantHdr := make(http.Header)
		wantHdr.Add("Content-Type", "text/plain; charset=utf-8")
		wantHdr.Add("Allow", "GET")
		wantHdr.Add("X-Content-Type-Options", "nosniff")
		CheckResponse(respw, http.StatusMethodNotAllowed, wantHdr, "Only GET supported.\n")
	})
})
