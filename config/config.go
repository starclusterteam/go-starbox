package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/starcluster/go-starbox/const/envvar"
)

// GoEnv keeps go environment
type GoEnv struct {
	Name        string
	Development bool
	Test        bool
	Staging     bool
	Production  bool
}

var (
	envPrefix string
	goEnv     GoEnv
)

// SetPrefix sets environment variables prefix during development
func SetPrefix(name string) {
	envPrefix = fmt.Sprintf("%s_", name)
}

// FetchGoEnv returns go struct
func FetchGoEnv() GoEnv {
	v := String("GO_ENV", "development")

	goEnv = GoEnv{
		Name:        v,
		Development: v == "development",
		Test:        v == "test",
		Staging:     v == "staging",
		Production:  v == "production",
	}
	return goEnv
}

// String returns config for this key or default value
func String(name, defaultValue string) string {
	return getOrDefault(name, defaultValue)
}

// StringEnv fetches from TEST_<var> if running in test mode
func StringEnv(name, defaultValue string, env GoEnv) string {
	if env.Test {
		return String("TEST_"+name, defaultValue)
	}
	return String(name, defaultValue)
}

// Duration parses the duration from the `name` environment variable.
// If the environment variable is not defined, it returns the given default value.
// It returns a duration or an error on parse error.
func Duration(name string, defaultValue time.Duration) (time.Duration, error) {
	s := String(name, "")
	if s == "" {
		return defaultValue, nil
	}

	d, err := time.ParseDuration(s)
	return d, errors.Wrapf(err, `failed to parse duration from "%s"`, name)
}

// Int returns config for this key as int or default value
func Int(name string, defaultValue int) int {
	val := getOrDefault(name, strconv.Itoa(defaultValue))

	i, err := strconv.Atoi(val)
	if err != nil {
		panic(fmt.Sprintf("could not parse Int value: %v", err))
	}

	return i
}

// Float returns config for this key as float64 or default value
func Float(name string, defaultValue float64) float64 {
	val := getEnv(name)
	if val == "" {
		return defaultValue
	}

	res, err := strconv.ParseFloat(val, 64)
	if err != nil {
		panic(fmt.Sprintf("could not parse float from envvar(%s): %v", name, err))
	}

	return res
}

// Bool returns false if variable is not set or set to false, true otherwise
func Bool(name string, defaultValue bool) bool {
	value := getEnv(name)

	if value == "" {
		return defaultValue
	}

	if value == "false" || value == "0" {
		return false
	}

	return true
}

// RedisURL returns value of REDIS_URL
// if running in test mode returns TEST_REDIS_URL
func RedisURL() string {
	return StringEnv(envvar.RedisURL, "", goEnv)
}

// ZipkinURL returns the URL of the Zipkin collector endpoint
func ZipkinURL() string {
	return StringEnv(envvar.ZipkinCollectorURL, "", goEnv)
}

// TracingSampleRate parses the TRACING_SAMPLE_RATE environment variable value to a float.
func TracingSampleRate(defaultValue float64) float64 {
	return Float("TRACING_SAMPLE_RATE", defaultValue)
}

// DatabaseURL returns value of DATABASE_URL
// if running in test mode returns TEST_DATABASE_URL
func DatabaseURL() string {
	return StringEnv(envvar.DatabaseURL, "", goEnv)
}

// SentryDSN returns the Sentry DSN (https://docs.sentry.io/quickstart/#configure-the-dsn)
// from an environment variable
func SentryDSN() string {
	return StringEnv(envvar.SentryDSN, "", goEnv)
}

func CommitHash() string {
	return String("COMMIT_HASH", "")
}

// GetListen returns :8080 and 8080 for ListenAddr and Port from PORT env var
func GetListen(defaultPort string) (string, string) {
	return GetListenByName("PORT", defaultPort)
}

// GetListenByName returns :8080 and 8080 for ListenAddr and Port, from key env var
func GetListenByName(key string, defaultPort string) (string, string) {
	p := String(key, defaultPort)
	addr := ":" + p
	return addr, p
}

func getOrDefault(name, defaultValue string) string {
	val := getEnv(name)

	if val == "" {
		return defaultValue
	}

	return val
}

func getEnv(name string) string {
	// First try if PREFIX_CONFIG_VAR
	val := os.Getenv(envPrefix + name)
	if val != "" {
		return val
	}

	// Then try CONFIG_VAR
	return os.Getenv(name)
}
