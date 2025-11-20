package helpers_test

import (
	"testing"

	"github.com/folio-org/eureka-cli/helpers"
	"github.com/stretchr/testify/assert"
)

func TestStripModuleVersion_WithDashAndSuffix(t *testing.T) {
	// Arrange - Removes everything after LAST dash
	moduleName := "mod-users-19.3.0"

	// Act
	result := helpers.StripModuleVersion(moduleName)

	// Assert
	assert.Equal(t, "mod-users", result) // Removes "-19.3.0"
}

func TestStripModuleVersion_WithMultipleDashes(t *testing.T) {
	// Arrange - Removes everything after LAST dash
	moduleName := "mod-inventory-storage-25.0.0"

	// Act
	result := helpers.StripModuleVersion(moduleName)

	// Assert
	assert.Equal(t, "mod-inventory-storage", result) // Removes "-25.0.0"
}

func TestStripModuleVersion_NoDash(t *testing.T) {
	// Arrange
	moduleName := "module"

	// Act
	result := helpers.StripModuleVersion(moduleName)

	// Assert
	assert.Equal(t, "module", result)
}

func TestStripModuleVersion_EmptyString(t *testing.T) {
	// Arrange
	moduleName := ""

	// Act
	result := helpers.StripModuleVersion(moduleName)

	// Assert
	assert.Equal(t, "", result)
}

func TestStripModuleVersion_OnlyDash(t *testing.T) {
	// Arrange
	moduleName := "-"

	// Act
	result := helpers.StripModuleVersion(moduleName)

	// Assert
	assert.Equal(t, "", result)
}

func TestFilterEmptyLines_EmptyString(t *testing.T) {
	// Arrange
	input := ""

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "", result)
}

func TestFilterEmptyLines_SingleLine(t *testing.T) {
	// Arrange
	input := "single line"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "single line", result)
}

func TestFilterEmptyLines_MultipleLines(t *testing.T) {
	// Arrange
	input := "line1\nline2\nline3"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "line1\nline2\nline3", result)
}

func TestFilterEmptyLines_WithEmptyLines(t *testing.T) {
	// Arrange
	input := "line1\n\nline2\n\n\nline3"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "line1\nline2\nline3", result)
}

func TestFilterEmptyLines_WithWhitespaceLines(t *testing.T) {
	// Arrange
	input := "line1\n   \nline2\n\t\nline3"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "line1\nline2\nline3", result)
}

func TestFilterEmptyLines_OnlyEmptyLines(t *testing.T) {
	// Arrange
	input := "\n\n\n"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "", result)
}

func TestFilterEmptyLines_OnlyWhitespaceLines(t *testing.T) {
	// Arrange
	input := "   \n\t\t\n  \n"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "", result)
}

func TestFilterEmptyLines_LeadingAndTrailingEmpty(t *testing.T) {
	// Arrange
	input := "\n\nline1\nline2\n\n"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "line1\nline2", result)
}

func TestFilterEmptyLines_PreservesIndentation(t *testing.T) {
	// Arrange
	input := "  indented line\n\tnext line\nregular line"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	assert.Equal(t, "  indented line\n\tnext line\nregular line", result)
}

func TestFilterEmptyLines_MixedContent(t *testing.T) {
	// Arrange
	input := `
out: Listen on ipv4:

Address         Port        Address         Port


--------------- ----------  --------------- ----------

192.168.1.100   8081        127.0.0.1       9000

`

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert
	expected := `out: Listen on ipv4:
Address         Port        Address         Port
--------------- ----------  --------------- ----------
192.168.1.100   8081        127.0.0.1       9000`
	assert.Equal(t, expected, result)
}

func TestFilterEmptyLines_WindowsLineEndings(t *testing.T) {
	// Arrange
	input := "line1\r\n\r\nline2\r\n"

	// Act
	result := helpers.FilterEmptyLines(input)

	// Assert - Note: strings.Split on "\n" will leave "\r" in the lines
	assert.Contains(t, result, "line1")
	assert.Contains(t, result, "line2")
}
