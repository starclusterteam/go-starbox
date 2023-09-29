package web_test

import (
	"encoding/json"
	"net/http/httptest"
	"testing"

	. "github.com/onsi/ginkgo"
	"github.com/starclusterteam/go-starbox/testutil"
	"github.com/starclusterteam/go-starbox/web"
	"github.com/stretchr/testify/assert"
)

var _ = Describe("errors", func() {
	Describe("LoginRequired", func() {
		It("Renders correctly", func() {
			respw := httptest.NewRecorder()
			web.NewLoginRequired().Write(respw)
			testutil.CheckUnauthorized(respw)
		})
	})

	Describe("ServiceUnavailable", func() {
		It("Renders correctly", func() {
			respw := httptest.NewRecorder()
			web.NewServiceUnavailable().Write(respw)
			testutil.CheckServiceUnavailable(respw)
		})
	})
})

func TestNotFound(t *testing.T) {
	w := httptest.NewRecorder()
	web.NotFound(w, 1337, "post", "comment")

	var resp web.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	expected := web.ErrorResponse{
		Messages: []string{"post not found", "comment not found"},
		Errors: map[string][]string{
			"post":    []string{"not found"},
			"comment": []string{"not found"},
		},
		Code: 1337,
	}

	assert.Equal(t, expected, resp)
}

func TestBadRequest(t *testing.T) {
	w := httptest.NewRecorder()
	web.BadRequest(w, 42, "email", "not valid", "too short")

	var resp web.ErrorResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(t, err)

	expected := web.ErrorResponse{
		Messages: []string{"email not valid", "email too short"},
		Errors: map[string][]string{
			"email": []string{"not valid", "too short"},
		},
		Code: 42,
	}

	assert.Equal(t, expected, resp)
}
