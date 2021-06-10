package web_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/starclusterteam/go-starbox/web"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRouteWithPrefix(t *testing.T) {
	var routes web.Routes
	routes = append(routes, web.NewRoute("GET", "/path/", handler("/path/"), web.WithPrefix()))
	routes = append(routes, web.NewRoute("GET", "/", handler("/"), web.WithPrefix()))

	tests := []struct {
		url      string
		expected string
	}{
		{"/", "/"},
		{"/pa", "/"},
		{"/pa.txt", "/"},
		{"/notpath/test", "/"},
		{"/notpath/test/", "/"},
		{"/path2/", "/"},
		{"/path/", "/path/"},
		{"/path/test", "/path/"},
		{"/path/test/", "/path/"},
	}
	s := httptest.NewServer(web.NewRouter(routes))
	defer s.Close()

	for _, test := range tests {
		t.Run(test.url, func(t *testing.T) {
			resp, err := http.Get(s.URL + test.url)
			require.NoError(t, err)

			actual, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.Equal(t, test.expected, string(actual))
		})
	}
}

func handler(name string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "%s", name)
	})
}
