package web_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/starclusterteam/go-starbox/web"
	"github.com/stretchr/testify/assert"
)

func testHandler(s string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", s)
	})
}

func appendMiddleware(s string) web.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Fprintf(w, "%s", s)
			next.ServeHTTP(w, r)
		})
	}
}

func TestMiddlewareChain(t *testing.T) {
	tests := []struct {
		name        string
		handler     http.Handler
		middlewares []web.Middleware
		expected    string
	}{
		{"no middleware", testHandler("x"), []web.Middleware{}, "x"},
		{"one middleware", testHandler("y"), []web.Middleware{appendMiddleware("m1")}, "m1y"},
		{"multiple middlewares", testHandler("z"), []web.Middleware{appendMiddleware("m1"), appendMiddleware("m2"), appendMiddleware("m3")}, "m1m2m3z"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			m := web.MiddlewareChain(test.middlewares...)

			r := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			m(test.handler).ServeHTTP(w, r)

			assert.Equal(t, test.expected, string(w.Body.Bytes()))
		})
	}
}
