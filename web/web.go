package web

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"runtime"

	"github.com/gorilla/mux"
	opentracing "github.com/opentracing/opentracing-go"
	"github.com/rs/cors"
	"github.com/starclusterteam/go-starbox/config"
	"github.com/starclusterteam/go-starbox/constants"
	"github.com/starclusterteam/go-starbox/constants/envvar"
	"github.com/starclusterteam/go-starbox/log"

	// This metrics import is used to initialize Prometheus HTTP endpoint server.
	_ "github.com/starclusterteam/go-starbox/metrics/auto"
	"github.com/starclusterteam/go-starbox/tracing"
)

const defaultPort = 8888

type Server interface {
	Run() error
	RunSSL(certFile, keyFile string) error
	Stop(ctx context.Context) error
}

// Web is a generic webserver.
type Web struct {
	server *http.Server
}

type serverOptions struct {
	addr           string
	tracer         opentracing.Tracer
	ping           bool
	pingPath       string
	defaultHandler http.Handler
	cors           *cors.Cors
}

// New returns new web instance that handle the given routes. If no port
// is given through the options, the port defined in the environment variable WEB_PORT
// will be used. If WEB_PORT is not defined, the default 8888 port is used.
func New(routes Routes, opts ...Option) Server {
	options := serverOptions{
		addr:     fmt.Sprintf(":%d", config.Int(envvar.WebPort, constants.WebPort)),
		tracer:   tracing.Tracer,
		ping:     true,
		pingPath: "/api/v1/ping",
	}

	for _, o := range opts {
		o(&options)
	}

	rs := make([]Route, len(routes))
	for i, r := range routes {
		rs[i] = r.WithMiddlewares(
			panicHandler,
			xRequestID,
			logger,
			TracingMiddleware(options.tracer, r.String()),
			defaultServerMetrics.Middleware(r.Pattern),
		)
	}

	if options.ping {
		rs = append(rs, NewRoute("GET", options.pingPath, Ping))
	}

	router := NewRouter(rs)

	if options.defaultHandler != nil {
		router.NotFoundHandler = options.defaultHandler
	}

	var finalHandler http.Handler = router

	if options.cors != nil {
		finalHandler = options.cors.Handler(finalHandler)
	}

	return &Web{
		server: &http.Server{
			Addr:    options.addr,
			Handler: finalHandler,
		},
	}
}

// Run starts the http server. It returns the error returned by http.Server.ListenAndServe, except if
// that error is http.ErrServerClosed.
func (w *Web) Run() error {
	log.Infof("Running web server on %s", w.server.Addr)

	if err := w.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

func (w *Web) RunSSL(certFile, keyFile string) error {
	log.Infof("Running web server on https://%s", w.server.Addr)

	if err := w.server.ListenAndServeTLS(certFile, keyFile); err != nil && err != http.ErrServerClosed {
		return err
	}

	return nil
}

// Stop gracefully shutdowns the http server.
func (w *Web) Stop(ctx context.Context) error {
	return w.server.Shutdown(ctx)
}

// Routes keeps all routes.
type Routes []Route

// Route object
type Route struct {
	Method  string
	Pattern string
	Handler http.Handler
	prefix  bool
}

// NewRoute returns a new route for this params.
func NewRoute(method, path string, handler http.Handler, opts ...RouteOption) Route {
	r := Route{
		Method:  method,
		Pattern: path,
		Handler: handler,
	}

	for _, o := range opts {
		o(&r)
	}

	return r
}

// WithMiddlewares returns a route with its handler wrapped with the given middlewares.
func (r Route) WithMiddlewares(middlewares ...Middleware) Route {
	return Route{
		Method:  r.Method,
		Pattern: r.Pattern,
		Handler: MiddlewareChain(middlewares...)(r.Handler),
		prefix:  r.prefix,
	}
}

// String returns the name of the root as a combination of the method and the route pattern.
// E.g.: "POST /api/v1/devices/{uuid}"
func (r Route) String() string {
	return fmt.Sprintf("%s %s", r.Method, r.Pattern)
}

// Option is used to specify functional options when creating a new web server.
type Option func(*serverOptions)

// WithTracer sets the tracer when creating a web server. The tracer will be added to each route.
func WithTracer(tracer opentracing.Tracer) Option {
	return func(o *serverOptions) {
		o.tracer = tracer
	}
}

// WithPing can be used to disable the ping route by giving WithPing(false) as option to New().
// The ping endpoint defaults to "/api/v1/ping". If you want to specify a custom path use `WithPingPath`.
func WithPing(ping bool) Option {
	return func(o *serverOptions) {
		o.ping = ping
	}
}

// WithPingPath sets the path of the ping endpoint. The ping endpoint defaults to "/api/v1/ping".
func WithPingPath(path string) Option {
	return func(o *serverOptions) {
		o.ping = true
		o.pingPath = path
	}
}

// WithPort overrides the default port used by New().
// The default port is the port given in the WEB_PORT environment variable, or 8888 if the environment variable is not set.
func WithPort(port int) Option {
	return func(o *serverOptions) {
		o.addr = fmt.Sprintf(":%d", port)
	}
}

// WithAddr sets the listening address of the server.
func WithAddr(addr string) Option {
	return func(o *serverOptions) {
		o.addr = addr
	}
}

// WithDefaultHandler sets the default handler for the router.
func WithDefaultHandler(handler http.Handler) Option {
	return func(o *serverOptions) {
		o.defaultHandler = handler
	}
}

func WithCORS(c *cors.Cors) Option {
	return func(o *serverOptions) {
		o.cors = c
	}
}

// RouteOption is a functional option for creating routes.
type RouteOption func(*Route)

// WithPrefix adds a matcher for the URL path prefix for a route.
// See http://www.gorillatoolkit.org/pkg/mux#Route.PathPrefix for more details.
func WithPrefix() RouteOption {
	return func(r *Route) {
		r.prefix = true
	}
}

// NewRouter returns a new router
func NewRouter(routes Routes) *mux.Router {
	router := mux.NewRouter().StrictSlash(true)

	for _, route := range routes {
		r := router.Methods(route.Method)
		if route.prefix {
			r = r.PathPrefix(route.Pattern)
		} else {
			r = r.Path(route.Pattern)
		}

		r = r.Handler(route.Handler)
	}

	return router
}

func panicHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if e := recover(); e != nil {
				stack := make([]byte, 1<<16)
				stackSize := runtime.Stack(stack, true)

				GetLogger(r).
					With("stack", string(stack[:stackSize])).
					Errorf("http handler panic: %v", e)
			}
		}()

		h.ServeHTTP(w, r)
	})
}

func xRequestID(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := generateID()

		SetLogger(r, GetLogger(r).With("request_id", id))

		r.Header.Set("X-Request-Id", id)
		w.Header().Set("X-Request-Id", id)

		h.ServeHTTP(w, r)
	})
}

func generateID() string {
	r := make([]byte, 16)
	_, err := rand.Read(r)
	if err != nil {
		panic(fmt.Sprintf("failed to generate random id: %v", err))
	}

	return hex.EncodeToString(r)
}
