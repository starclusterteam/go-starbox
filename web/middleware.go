package web

import (
	"net/http"
)

// ContextKey is a type alias for string to prevent key collisions across packages
type ContextKey string

var (
	// SessionContextKey is the request context key where session is kept
	SessionContextKey = ContextKey("session")
)

// Middleware defines an HTTP middleware: a decorator that takes an HTTP handler and returns an HTTP handler.
type Middleware func(next http.Handler) http.Handler

// MiddlewareChain returns a middleware composed of sequentially applied middlewares.
// It is implemented as a recursive function, where the first middleware wraps the
// resulted chain of the following middlewares.
func MiddlewareChain(middlewares ...Middleware) Middleware {
	// Return a no-op middleware if no middlewares are given.
	if len(middlewares) == 0 {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return func(next http.Handler) http.Handler {
		return middlewares[0](MiddlewareChain(middlewares[1:]...)(next))
	}
}
