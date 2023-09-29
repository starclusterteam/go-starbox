package web

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type fakeContext struct {
	muxVars map[string]string
}

func (c *fakeContext) Value(e interface{}) interface{} {
	return c.muxVars
}

func (c *fakeContext) Deadline() (deadline time.Time, ok bool) {
	return
}

func (c *fakeContext) Done() <-chan struct{} {
	return nil
}

func (c *fakeContext) Err() error {
	return nil
}

func TestWithInvalidChars(t *testing.T) {
	fakeCtx := &fakeContext{
		muxVars: map[string]string{
			"project_id": "value",
		},
	}

	req := &http.Request{}
	req = req.WithContext(fakeCtx)

	val, err := FetchIntVar(req, "project_id")
	assert.NotNil(t, err)
	assert.Equal(t, 0, val)
	assert.Equal(t, ErrInvalidRequestFormat, err)

	val, err = FetchIntVar(req, "xx_id")
	assert.NotNil(t, err)
	assert.Equal(t, 0, val)
	assert.True(t, err.isInternalError)
	assert.Equal(t, "variable xx_id not found", err.internalError.Error())
}

func TestReturnsCorrectValue(t *testing.T) {
	fakeCtx := &fakeContext{
		muxVars: map[string]string{
			"project_id": "123",
		},
	}

	req := &http.Request{}
	req = req.WithContext(fakeCtx)

	val, err := FetchIntVar(req, "project_id")
	assert.Nil(t, err)
	assert.Equal(t, 123, val)
}

