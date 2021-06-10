package web

import (
	"net/http"

	"github.com/felixge/httpsnoop"
)

func logger(inner http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		SetLogger(
			r,
			GetLogger(r).
				With("method", r.Method).
				With("url", "http://"+r.Host+r.URL.String()),
		)

		m := httpsnoop.CaptureMetrics(inner, w, r)

		GetLogger(r).
			With("latency", m.Duration.String()).
			With("status", m.Code).
			Info("request")
	})
}
