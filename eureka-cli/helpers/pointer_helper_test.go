package helpers_test

import (
	"testing"
	"time"

	"github.com/j011195/eureka-setup/eureka-cli/helpers"
	"github.com/stretchr/testify/assert"
)

func TestStringPtr(t *testing.T) {
	// Arrange
	value := "test-string"

	// Act
	result := helpers.StringPtr(value)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, value, *result)
	assert.Equal(t, &value, result)
}

func TestStringPtr_EmptyString(t *testing.T) {
	// Arrange
	value := ""

	// Act
	result := helpers.StringPtr(value)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, "", *result)
}

func TestBoolPtr_True(t *testing.T) {
	// Arrange
	value := true

	// Act
	result := helpers.BoolPtr(value)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, true, *result)
}

func TestBoolPtr_False(t *testing.T) {
	// Arrange
	value := false

	// Act
	result := helpers.BoolPtr(value)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, false, *result)
}

func TestIntPtr(t *testing.T) {
	// Arrange
	value := 42

	// Act
	result := helpers.IntPtr(value)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, 42, *result)
}

func TestIntPtr_Zero(t *testing.T) {
	// Arrange
	value := 0

	// Act
	result := helpers.IntPtr(value)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, 0, *result)
}

func TestIntPtr_Negative(t *testing.T) {
	// Arrange
	value := -10

	// Act
	result := helpers.IntPtr(value)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, -10, *result)
}

func TestDefaultInt_WithValue(t *testing.T) {
	// Arrange
	value := 100
	defaultValue := 50

	// Act
	result := helpers.DefaultInt(value, defaultValue)

	// Assert
	assert.Equal(t, 100, result)
}

func TestDefaultInt_WithZero(t *testing.T) {
	// Arrange
	value := 0
	defaultValue := 50

	// Act
	result := helpers.DefaultInt(value, defaultValue)

	// Assert
	assert.Equal(t, 50, result)
}

func TestDefaultInt_BothZero(t *testing.T) {
	// Arrange
	value := 0
	defaultValue := 0

	// Act
	result := helpers.DefaultInt(value, defaultValue)

	// Assert
	assert.Equal(t, 0, result)
}

func TestDefaultInt_NegativeValue(t *testing.T) {
	// Arrange
	value := -5
	defaultValue := 10

	// Act
	result := helpers.DefaultInt(value, defaultValue)

	// Assert
	assert.Equal(t, -5, result, "Negative values should not be replaced")
}

func TestDefaultDuration_WithValue(t *testing.T) {
	// Arrange
	value := 30 * time.Second
	defaultValue := 10 * time.Second

	// Act
	result := helpers.DefaultDuration(value, defaultValue)

	// Assert
	assert.Equal(t, 30*time.Second, result)
}

func TestDefaultDuration_WithZero(t *testing.T) {
	// Arrange
	value := time.Duration(0)
	defaultValue := 10 * time.Second

	// Act
	result := helpers.DefaultDuration(value, defaultValue)

	// Assert
	assert.Equal(t, 10*time.Second, result)
}

func TestDefaultDuration_BothZero(t *testing.T) {
	// Arrange
	value := time.Duration(0)
	defaultValue := time.Duration(0)

	// Act
	result := helpers.DefaultDuration(value, defaultValue)

	// Assert
	assert.Equal(t, time.Duration(0), result)
}

func TestDefaultDuration_Milliseconds(t *testing.T) {
	// Arrange
	value := 500 * time.Millisecond
	defaultValue := 1 * time.Second

	// Act
	result := helpers.DefaultDuration(value, defaultValue)

	// Assert
	assert.Equal(t, 500*time.Millisecond, result)
}

func TestDefaultDuration_Minutes(t *testing.T) {
	// Arrange
	value := 5 * time.Minute
	defaultValue := 1 * time.Minute

	// Act
	result := helpers.DefaultDuration(value, defaultValue)

	// Assert
	assert.Equal(t, 5*time.Minute, result)
}

func TestDefaultDuration_NegativeValue(t *testing.T) {
	// Arrange
	value := -5 * time.Second
	defaultValue := 10 * time.Second

	// Act
	result := helpers.DefaultDuration(value, defaultValue)

	// Assert
	assert.Equal(t, -5*time.Second, result, "Negative durations should not be replaced")
}
