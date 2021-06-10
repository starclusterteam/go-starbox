package log_test

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/starclusterteam/go-starbox/log"
)

func TestWith(t *testing.T) {
	var buf bytes.Buffer
	log.LogrusLogger().Out = &buf

	log.Logger().With("key", 3).With("field", "xxx").Info("test")

	data := make(map[string]interface{})
	err := json.NewDecoder(&buf).Decode(&data)
	require.NoError(t, err)

	assert.Contains(t, data, "key")
	assert.Contains(t, data, "field")
	assert.Contains(t, data, "msg")
	assert.Contains(t, data, "time")
	assert.Contains(t, data, "level")

	assert.EqualValues(t, data["key"], 3)
	assert.Equal(t, data["msg"], "test")
	assert.Equal(t, data["field"], "xxx")

	buf.Reset()
	log.Logger().Info("test")
	data = make(map[string]interface{})
	err = json.NewDecoder(&buf).Decode(&data)
	require.NoError(t, err)

	assert.NotContains(t, data, "key")
	assert.NotContains(t, data, "field")
}
