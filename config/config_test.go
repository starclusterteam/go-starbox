package config

import (
	"os"
	"testing"
	"time"
)

const (
	DurationEnv             = "TEST_DURATION"
	DurationValue           = time.Minute
	DurationDefaultValue    = time.Second
	DurationUnparsableValue = "asd"
)

func TestDurationDefaultValue(t *testing.T) {
	os.Unsetenv(DurationEnv)

	d, err := Duration(DurationEnv, DurationDefaultValue)
	if err != nil {
		t.Fatalf("Expected %+v to not have occured", err)
	}

	if d != DurationDefaultValue {
		t.Fatalf("Expected %s to be %s", d, DurationDefaultValue)
	}
}

func TestDuration(t *testing.T) {
	os.Unsetenv(DurationEnv)
	defer os.Unsetenv(DurationEnv)

	os.Setenv(DurationEnv, DurationValue.String())

	d, err := Duration(DurationEnv, 0)
	if err != nil {
		t.Fatalf("Expected %+v to not have occured", err)
	}

	if d != DurationValue {
		t.Fatalf("Expected %s to be %s", d, DurationValue)
	}
}

func TestDurationParseError(t *testing.T) {
	os.Unsetenv(DurationEnv)
	defer os.Unsetenv(DurationEnv)

	os.Setenv(DurationEnv, DurationUnparsableValue)

	_, err := Duration(DurationEnv, 0)

	if err == nil {
		t.Fatalf("Expected error to have occured")
	}
}
