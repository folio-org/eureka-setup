package helpers_test

import (
	"testing"

	"github.com/j011195/eureka-setup/eureka-cli/helpers"
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

// ==================== GetInt Tests ====================

func TestGetInt_ValueExists(t *testing.T) {
	// Arrange
	entry := map[string]any{"port": 8080}

	// Act
	result := helpers.GetInt(entry, "port")

	// Assert
	assert.Equal(t, 8080, result)
}

func TestGetInt_NilValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"port": nil}

	// Act
	result := helpers.GetInt(entry, "port")

	// Assert
	assert.Equal(t, 0, result)
}

func TestGetInt_KeyNotExists(t *testing.T) {
	// Arrange
	entry := map[string]any{}

	// Act
	result := helpers.GetInt(entry, "port")

	// Assert
	assert.Equal(t, 0, result)
}

func TestGetInt_WrongType(t *testing.T) {
	// Arrange
	entry := map[string]any{"port": "8080"}

	// Act
	result := helpers.GetInt(entry, "port")

	// Assert
	assert.Equal(t, 0, result)
}

// ==================== GetIntPtr Tests ====================

func TestGetIntPtr_ValueExists(t *testing.T) {
	// Arrange
	entry := map[string]any{"port": 8080}

	// Act
	result := helpers.GetIntPtr(entry, "port")

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, 8080, *result)
}

func TestGetIntPtr_NilValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"port": nil}

	// Act
	result := helpers.GetIntPtr(entry, "port")

	// Assert
	assert.Nil(t, result)
}

func TestGetIntPtr_KeyNotExists(t *testing.T) {
	// Arrange
	entry := map[string]any{}

	// Act
	result := helpers.GetIntPtr(entry, "port")

	// Assert
	assert.Nil(t, result)
}

func TestGetIntPtr_WrongType(t *testing.T) {
	// Arrange
	entry := map[string]any{"port": "8080"}

	// Act
	result := helpers.GetIntPtr(entry, "port")

	// Assert
	assert.Nil(t, result)
}

// ==================== GetBoolPtr Tests ====================

func TestGetBoolPtr_ValueExists(t *testing.T) {
	// Arrange
	entry := map[string]any{"enabled": true}

	// Act
	result := helpers.GetBoolPtr(entry, "enabled")

	// Assert
	assert.NotNil(t, result)
	assert.True(t, *result)
}

func TestGetBoolPtr_NilValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"enabled": nil}

	// Act
	result := helpers.GetBoolPtr(entry, "enabled")

	// Assert
	assert.Nil(t, result)
}

func TestGetBoolPtr_KeyNotExists(t *testing.T) {
	// Arrange
	entry := map[string]any{}

	// Act
	result := helpers.GetBoolPtr(entry, "enabled")

	// Assert
	assert.Nil(t, result)
}

func TestGetBoolPtr_WrongType(t *testing.T) {
	// Arrange
	entry := map[string]any{"enabled": "true"}

	// Act
	result := helpers.GetBoolPtr(entry, "enabled")

	// Assert
	assert.Nil(t, result)
}

// ==================== GetStringSlice Tests ====================

func TestGetStringSlice_ValueExists(t *testing.T) {
	// Arrange
	entry := map[string]any{"items": []any{"a", "b", "c"}}

	// Act
	result := helpers.GetStringSlice(entry, "items")

	// Assert
	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestGetStringSlice_MixedTypes(t *testing.T) {
	// Arrange
	entry := map[string]any{"items": []any{"a", 123, "b", true, "c"}}

	// Act
	result := helpers.GetStringSlice(entry, "items")

	// Assert
	assert.Equal(t, []string{"a", "b", "c"}, result)
}

func TestGetStringSlice_NilValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"items": nil}

	// Act
	result := helpers.GetStringSlice(entry, "items")

	// Assert
	assert.Empty(t, result)
}

func TestGetStringSlice_KeyNotExists(t *testing.T) {
	// Arrange
	entry := map[string]any{}

	// Act
	result := helpers.GetStringSlice(entry, "items")

	// Assert
	assert.Empty(t, result)
}

func TestGetStringSlice_WrongType(t *testing.T) {
	// Arrange
	entry := map[string]any{"items": "not-a-slice"}

	// Act
	result := helpers.GetStringSlice(entry, "items")

	// Assert
	assert.Empty(t, result)
}

func TestGetStringSlice_EmptySlice(t *testing.T) {
	// Arrange
	entry := map[string]any{"items": []any{}}

	// Act
	result := helpers.GetStringSlice(entry, "items")

	// Assert
	assert.Empty(t, result)
}

// ==================== GetMap Tests ====================

func TestGetMap_ValueExists(t *testing.T) {
	// Arrange
	innerMap := map[string]any{"key": "value"}
	entry := map[string]any{"config": innerMap}

	// Act
	result := helpers.GetMap(entry, "config")

	// Assert
	assert.Equal(t, innerMap, result)
}

func TestGetMap_NilValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"config": nil}

	// Act
	result := helpers.GetMap(entry, "config")

	// Assert
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestGetMap_KeyNotExists(t *testing.T) {
	// Arrange
	entry := map[string]any{}

	// Act
	result := helpers.GetMap(entry, "config")

	// Assert
	assert.NotNil(t, result)
	assert.Empty(t, result)
}

func TestGetMap_WrongType(t *testing.T) {
	// Arrange
	entry := map[string]any{"config": "not-a-map"}

	// Act
	result := helpers.GetMap(entry, "config")

	// Assert
	assert.NotNil(t, result)
	assert.Empty(t, result)
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

func TestGetString_ValidString(t *testing.T) {
	// Arrange
	entry := map[string]any{"name": "test-value"}

	// Act
	result := helpers.GetString(entry, "name")

	// Assert
	assert.Equal(t, "test-value", result)
}

func TestGetString_EmptyString(t *testing.T) {
	// Arrange
	entry := map[string]any{"name": ""}

	// Act
	result := helpers.GetString(entry, "name")

	// Assert
	assert.Equal(t, "", result)
}

func TestGetString_NilValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"name": nil}

	// Act
	result := helpers.GetString(entry, "name")

	// Assert
	assert.Equal(t, "", result)
}

func TestGetString_NonStringValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"name": 123}

	// Act
	result := helpers.GetString(entry, "name")

	// Assert
	assert.Equal(t, "", result)
}

func TestGetString_KeyNotExists(t *testing.T) {
	// Arrange
	entry := map[string]any{"other": "value"}

	// Act
	result := helpers.GetString(entry, "name")

	// Assert
	assert.Equal(t, "", result)
}

func TestGetStringOrDefault_ValidString(t *testing.T) {
	// Arrange
	entry := map[string]any{"name": "actual-value"}

	// Act
	result := helpers.GetStringOrDefault(entry, "name", "default-value")

	// Assert
	assert.Equal(t, "actual-value", result)
}

func TestGetStringOrDefault_EmptyString(t *testing.T) {
	// Arrange
	entry := map[string]any{"name": ""}

	// Act
	result := helpers.GetStringOrDefault(entry, "name", "default-value")

	// Assert
	assert.Equal(t, "", result) // Empty string is a valid value, not default
}

func TestGetStringOrDefault_NilValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"name": nil}

	// Act
	result := helpers.GetStringOrDefault(entry, "name", "default-value")

	// Assert
	assert.Equal(t, "default-value", result)
}

func TestGetStringOrDefault_NonStringValue(t *testing.T) {
	// Arrange
	entry := map[string]any{"name": 123}

	// Act
	result := helpers.GetStringOrDefault(entry, "name", "default-value")

	// Assert
	assert.Equal(t, "default-value", result)
}

func TestGetStringOrDefault_KeyNotExists(t *testing.T) {
	// Arrange
	entry := map[string]any{"other": "value"}

	// Act
	result := helpers.GetStringOrDefault(entry, "name", "default-value")

	// Assert
	assert.Equal(t, "default-value", result)
}

func TestGetBoolOrDefault_KeyNotExists(t *testing.T) {
	// Arrange
	entry := map[string]any{"other": true}

	// Act
	result := helpers.GetBoolOrDefault(entry, "enabled", true)

	// Assert
	assert.True(t, result)
}
