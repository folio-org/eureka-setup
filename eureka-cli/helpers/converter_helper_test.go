package helpers_test

import (
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/stretchr/testify/assert"
)

func TestConvertMemory_BytesToMib(t *testing.T) {
	// Arrange
	bytes := int64(1048576) // 1 MiB

	// Act
	result := helpers.ConvertMemory(helpers.BytesToMib, bytes)

	// Assert
	assert.Equal(t, int64(1), result)
}

func TestConvertMemory_MibToBytes(t *testing.T) {
	// Arrange
	mib := int64(1)

	// Act
	result := helpers.ConvertMemory(helpers.MibToBytes, mib)

	// Assert
	assert.Equal(t, int64(1048576), result)
}

func TestConvertMemory_ZeroValue(t *testing.T) {
	// Arrange
	zero := int64(0)

	// Act
	resultBytesToMib := helpers.ConvertMemory(helpers.BytesToMib, zero)
	resultMibToBytes := helpers.ConvertMemory(helpers.MibToBytes, zero)

	// Assert
	assert.Equal(t, int64(0), resultBytesToMib)
	assert.Equal(t, int64(0), resultMibToBytes)
}

func TestConvertMemory_NegativeValue(t *testing.T) {
	// Arrange
	negative := int64(-100)

	// Act
	result := helpers.ConvertMemory(helpers.BytesToMib, negative)

	// Assert
	assert.Equal(t, int64(-100), result)
}

func TestConvertMemory_LargeValue(t *testing.T) {
	// Arrange
	largeBytes := int64(1073741824) // 1 GiB

	// Act
	result := helpers.ConvertMemory(helpers.BytesToMib, largeBytes)

	// Assert
	assert.Equal(t, int64(1024), result)
}

func TestConvertMemory_InvalidMode(t *testing.T) {
	// Arrange
	value := int64(1048576)
	invalidMode := helpers.ConversionMode(99)

	// Act
	result := helpers.ConvertMemory(invalidMode, value)

	// Assert
	assert.Equal(t, int64(1048576), result) // Returns original value unchanged
}

func TestConvertMapKeyToSlice_WithMultipleEntries(t *testing.T) {
	// Arrange
	testMap := map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}

	// Act
	result := helpers.ConvertMapKeyToSlice(testMap)

	// Assert
	assert.Len(t, result, 3)
	assert.Contains(t, result, "key1")
	assert.Contains(t, result, "key2")
	assert.Contains(t, result, "key3")
}

func TestConvertMapKeyToSlice_EmptyMap(t *testing.T) {
	// Arrange
	testMap := map[string]any{}

	// Act
	result := helpers.ConvertMapKeyToSlice(testMap)

	// Assert
	assert.Empty(t, result)
}

func TestConvertMapKeyToSlice_SingleEntry(t *testing.T) {
	// Arrange
	testMap := map[string]any{"onlyKey": "onlyValue"}

	// Act
	result := helpers.ConvertMapKeyToSlice(testMap)

	// Assert
	assert.Len(t, result, 1)
	assert.Equal(t, "onlyKey", result[0])
}
