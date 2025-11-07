package helpers_test

import (
	"testing"

	"github.com/folio-org/eureka-cli/helpers"
	"github.com/stretchr/testify/assert"
)

func TestGetBool_TrueValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"enabled": true}

	// Act
	result := helpers.GetBool(entry, "enabled")

	// Assert
	assert.True(t, result)
}

func TestGetBool_FalseValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"enabled": false}

	// Act
	result := helpers.GetBool(entry, "enabled")

	// Assert
	assert.False(t, result)
}

func TestGetBool_NilValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"enabled": nil}

	// Act
	result := helpers.GetBool(entry, "enabled")

	// Assert
	assert.False(t, result)
}

func TestGetBool_NonBoolValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"enabled": "true"}

	// Act
	result := helpers.GetBool(entry, "enabled")

	// Assert
	assert.False(t, result)
}

func TestGetBool_KeyNotExists(t *testing.T) {
	// Arrange
	entry := map[string]any{"other": true}

	// Act
	result := helpers.GetBool(entry, "enabled")

	// Assert
	assert.False(t, result)
}

func TestGetAnyOrDefault_ValueExists(t *testing.T) {
	// Arrange
	entry := map[string]any{"key": "value"}

	// Act
	result := helpers.GetAnyOrDefault(entry, "key", "default")

	// Assert
	assert.Equal(t, "value", result)
}

func TestGetAnyOrDefault_NilValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"key": nil}

	// Act
	result := helpers.GetAnyOrDefault(entry, "key", "default")

	// Assert
	assert.Equal(t, "default", result)
}

func TestGetAnyOrDefault_KeyNotExists(t *testing.T) {
	// Arrange
	entry := map[string]any{}

	// Act
	result := helpers.GetAnyOrDefault(entry, "key", "default")

	// Assert
	assert.Equal(t, "default", result)
}

func TestGetIntOrDefault_ValidInt(t *testing.T) {
	// Arrange
	entry := map[string]any{"count": 42}

	// Act
	result := helpers.GetIntOrDefault(entry, "count", 10)

	// Assert
	assert.Equal(t, int64(42), result)
}

func TestGetIntOrDefault_NilValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"count": nil}

	// Act
	result := helpers.GetIntOrDefault(entry, "count", 10)

	// Assert
	assert.Equal(t, int64(10), result)
}

func TestGetIntOrDefault_NonIntValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"count": "42"}

	// Act
	result := helpers.GetIntOrDefault(entry, "count", 10)

	// Assert
	assert.Equal(t, int64(10), result)
}

func TestGetIntOrDefault_KeyNotExists(t *testing.T) {
	// Arrange
	entry := map[string]any{}

	// Act
	result := helpers.GetIntOrDefault(entry, "count", 10)

	// Assert
	assert.Equal(t, int64(10), result)
}

func TestGetBoolOrDefault_TrueValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"enabled": true}

	// Act
	result := helpers.GetBoolOrDefault(entry, "enabled", false)

	// Assert
	assert.True(t, result)
}

func TestGetBoolOrDefault_FalseValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"enabled": false}

	// Act
	result := helpers.GetBoolOrDefault(entry, "enabled", true)

	// Assert
	assert.False(t, result)
}

func TestGetBoolOrDefault_NilValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"enabled": nil}

	// Act
	result := helpers.GetBoolOrDefault(entry, "enabled", true)

	// Assert
	assert.True(t, result)
}

func TestGetBoolOrDefault_NonBoolValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"enabled": "true"}

	// Act
	result := helpers.GetBoolOrDefault(entry, "enabled", false)

	// Assert
	assert.False(t, result)
}

func TestSetBool_ValidBoolValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"enabled": true}
	var result bool

	// Act
	helpers.SetBool(entry, "enabled", &result)

	// Assert
	assert.True(t, result)
}

func TestSetBool_NilValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"enabled": nil}
	result := true

	// Act
	helpers.SetBool(entry, "enabled", &result)

	// Assert
	assert.True(t, result) // Should remain unchanged
}

func TestSetBool_NonBoolValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"enabled": "true"}
	result := false

	// Act
	helpers.SetBool(entry, "enabled", &result)

	// Assert
	assert.False(t, result) // Should remain unchanged
}

func TestSetString_ValidStringValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"name": "test"}
	var result string

	// Act
	helpers.SetString(entry, "name", &result)

	// Assert
	assert.Equal(t, "test", result)
}

func TestSetString_NilValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"name": nil}
	result := "original"

	// Act
	helpers.SetString(entry, "name", &result)

	// Assert
	assert.Equal(t, "original", result) // Should remain unchanged
}

func TestSetString_NonStringValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"name": 123}
	result := "original"

	// Act
	helpers.SetString(entry, "name", &result)

	// Assert
	assert.Equal(t, "original", result) // Should remain unchanged
}
