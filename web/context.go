package web

import (
	"context"
	"net/http"

	"github.com/starclusterteam/go-starbox/log"
)

type key int

// Context keys
const (
	_ key = iota
	LOGGERKEY
)

// SetLogger sets logger to request context.
func SetLogger(r *http.Request, l log.Interface) *http.Request {
	ctx := context.WithValue(r.Context(), LOGGERKEY, l)
	*r = *r.WithContext(ctx)
	return r
}

// GetLogger retrieves logger from request context.
func GetLogger(r *http.Request) log.Interface {
	val, ok := r.Context().Value(LOGGERKEY).(log.Interface)
	if !ok {
		return log.Logger()
	}

	return val
}
