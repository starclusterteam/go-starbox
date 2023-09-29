package errors

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShouldWrapError(t *testing.T) {
	err := Wrap(fmt.Errorf("test"), "test")
	assert.EqualError(t, err, "test: test")
}

func TestWrappedErrorIsEqual(t *testing.T) {
	err := New("test")

	wrappedErr := Wrap(err, "with wrap")
	assert.True(t, Is(wrappedErr, err))
}

func TestMessage(t *testing.T) {
	a := New("test")
	assert.Equal(t, "test", a.Error())

	b := Wrap(a, "wrapped once")
	assert.Equal(t, "wrapped once: test", b.Error())

	c := Wrap(b, "wrapped twice")
	assert.Equal(t, "wrapped twice: wrapped once: test", c.Error())

	d := Wrap(c, "wrapped thrice")
	assert.Equal(t, "wrapped thrice: wrapped twice: wrapped once: test", d.Error())

	assert.True(t, Is(d, a))
}

