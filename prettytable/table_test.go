package prettytable

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrintsTable(t *testing.T) {
	header := []string{"Key", "Name"}
	rows := [][]string{
		{"A", "B"},
		{"C", "D"},
	}

	expected := "Key   Name   \nA     B      \nC     D      \n"

	var buf bytes.Buffer
	PrintTable(&buf, header, rows...)

	assert.Equal(t, buf.String(), expected)
}

