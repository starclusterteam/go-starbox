package web

import "net/http"

// Get returns a new route for this params for a GET request
func Get(path string, handler http.Handler) Route {
	return NewRoute("GET", path, handler)
}

// Post returns a new route for this params for a POST request
func Post(path string, handler http.Handler) Route {
	return NewRoute("POST", path, handler)
}

// Put returns a new route for this params for a PUT request
func Put(path string, handler http.Handler) Route {
	return NewRoute("PUT", path, handler)
}

// Patch returns a new route for this params for a PATCH request
func Patch(path string, handler http.Handler) Route {
	return NewRoute("PATCH", path, handler)
}

// Delete returns a new route for this params for a DELETE request
func Delete(path string, handler http.Handler) Route {
	return NewRoute("DELETE", path, handler)
}
