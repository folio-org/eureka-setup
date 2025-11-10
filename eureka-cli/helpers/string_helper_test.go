package helpers_test

import (
	"testing"

	"github.com/folio-org/eureka-cli/helpers"
	"github.com/stretchr/testify/assert"
)

func TestTrimModuleName_WithDashAndSuffix(t *testing.T) {
	// Arrange - Removes everything after LAST dash
	moduleName := "mod-users-19.3.0"

	// Act
	result := helpers.TrimModuleName(moduleName)

	// Assert
	assert.Equal(t, "mod-users", result) // Removes "-19.3.0"
}

func TestTrimModuleName_WithMultipleDashes(t *testing.T) {
	// Arrange - Removes everything after LAST dash
	moduleName := "mod-inventory-storage-25.0.0"

	// Act
	result := helpers.TrimModuleName(moduleName)

	// Assert
	assert.Equal(t, "mod-inventory-storage", result) // Removes "-25.0.0"
}

func TestTrimModuleName_NoDash(t *testing.T) {
	// Arrange
	moduleName := "module"

	// Act
	result := helpers.TrimModuleName(moduleName)

	// Assert
	assert.Equal(t, "module", result)
}

func TestTrimModuleName_EmptyString(t *testing.T) {
	// Arrange
	moduleName := ""

	// Act
	result := helpers.TrimModuleName(moduleName)

	// Assert
	assert.Equal(t, "", result)
}

func TestTrimModuleName_OnlyDash(t *testing.T) {
	// Arrange
	moduleName := "-"

	// Act
	result := helpers.TrimModuleName(moduleName)

	// Assert
	assert.Equal(t, "", result)
}
