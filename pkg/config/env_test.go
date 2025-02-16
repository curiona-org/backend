package config_test

import (
	"os"
	"testing"
	"time"

	"github.com/roadmap-thesis/backend/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestConfig_LookupEnv(t *testing.T) {
	t.Parallel()

	t.Run("String", func(t *testing.T) {
		t.Parallel()

		t.Setenv("TEST_STRING", "test_value")
		defer os.Unsetenv("TEST_STRING")

		result := config.LookupEnv("TEST_STRING", "default_value")
		assert.Equal(t, "test_value", result)
	})

	t.Run("StringDefault", func(t *testing.T) {
		t.Parallel()

		result := config.LookupEnv("NON_EXISTENT_STRING", "default_value")
		assert.Equal(t, "default_value", result)
	})

	t.Run("Int", func(t *testing.T) {
		t.Parallel()

		t.Setenv("TEST_INT", "42")
		defer os.Unsetenv("TEST_INT")

		result := config.LookupEnv("TEST_INT", 0)
		assert.Equal(t, 42, result)
	})

	t.Run("IntDefault", func(t *testing.T) {
		t.Parallel()

		result := config.LookupEnv("NON_EXISTENT_INT", 0)
		assert.Equal(t, 0, result)
	})

	t.Run("Int32", func(t *testing.T) {
		t.Parallel()

		t.Setenv("TEST_INT32", "32")
		defer os.Unsetenv("TEST_INT32")

		result := config.LookupEnv("TEST_INT32", int32(0))
		assert.Equal(t, int32(32), result)
	})

	t.Run("Int32Default", func(t *testing.T) {
		t.Parallel()

		result := config.LookupEnv("NON_EXISTENT_INT32", int32(0))
		assert.Equal(t, int32(0), result)
	})

	t.Run("Int64", func(t *testing.T) {
		t.Parallel()

		t.Setenv("TEST_INT64", "64")
		defer os.Unsetenv("TEST_INT64")

		result := config.LookupEnv("TEST_INT64", int64(0))
		assert.Equal(t, int64(64), result)
	})

	t.Run("Int64Default", func(t *testing.T) {
		t.Parallel()

		result := config.LookupEnv("NON_EXISTENT_INT64", int64(0))
		assert.Equal(t, int64(0), result)
	})

	t.Run("Duration", func(t *testing.T) {
		t.Parallel()

		t.Setenv("TEST_DURATION", "1h")
		defer os.Unsetenv("TEST_DURATION")

		result := config.LookupEnv("TEST_DURATION", time.Minute)
		assert.Equal(t, time.Hour, result)
	})

	t.Run("DurationDefault", func(t *testing.T) {
		t.Parallel()

		result := config.LookupEnv("NON_EXISTENT_DURATION", time.Minute)
		assert.Equal(t, time.Minute, result)
	})

	t.Run("DurationInvalid", func(t *testing.T) {
		t.Parallel()

		t.Setenv("TEST_DURATION_INVALID", "invalid")
		defer os.Unsetenv("TEST_DURATION_INVALID")

		result := config.LookupEnv("TEST_DURATION_INVALID", time.Minute)
		assert.Equal(t, time.Minute, result)
	})

	t.Run("StringSlice", func(t *testing.T) {
		t.Parallel()

		t.Setenv("TEST_STRING_SLICE", "a,b,c")
		defer os.Unsetenv("TEST_STRING_SLICE")

		result := config.LookupEnv("TEST_STRING_SLICE", []string{"default"})
		assert.Equal(t, []string{"a", "b", "c"}, result)
	})

	t.Run("StringSliceDefault", func(t *testing.T) {
		t.Parallel()

		result := config.LookupEnv("NON_EXISTENT_STRING_SLICE", []string{"default"})
		assert.Equal(t, []string{"default"}, result)
	})
}
