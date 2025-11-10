package helpers_test

import (
	"bytes"
	"testing"

	"github.com/folio-org/eureka-cli/helpers"
	"github.com/stretchr/testify/assert"
)

func TestGetVaultRootTokenFromLogs_ValidLogLine(t *testing.T) {
	// Arrange
	logLine := "Root Token: hvs.1234567890abcdef"

	// Act
	result := helpers.GetVaultRootTokenFromLogs(logLine)

	// Assert
	assert.Equal(t, "hvs.1234567890abcdef", result)
}

func TestGetVaultRootTokenFromLogs_WithWhitespace(t *testing.T) {
	// Arrange
	logLine := "Root Token:   hvs.token123   "

	// Act
	result := helpers.GetVaultRootTokenFromLogs(logLine)

	// Assert
	assert.Equal(t, "hvs.token123", result)
}

func TestGetPortFromURL_ValidURL(t *testing.T) {
	// Arrange - Pattern ".*:" removes everything up to and including the colon
	url := "http://localhost:8080"

	// Act
	result, err := helpers.GetPortFromURL(url)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 8080, result)
}

func TestGetPortFromURL_InvalidPort(t *testing.T) {
	// Arrange
	url := "http://localhost:invalid/api"

	// Act
	_, err := helpers.GetPortFromURL(url)

	// Assert
	assert.Error(t, err)
}

func TestGetModuleNameFromID_StandardModule(t *testing.T) {
	// Arrange
	moduleID := "mod-users-19.3.0"

	// Act
	result := helpers.GetModuleNameFromID(moduleID)

	// Assert
	assert.Equal(t, "mod-users", result)
}

func TestGetModuleNameFromID_SnapshotVersion(t *testing.T) {
	// Arrange
	moduleID := "mod-inventory-1.0.0-SNAPSHOT.123"

	// Act
	result := helpers.GetModuleNameFromID(moduleID)

	// Assert
	assert.Equal(t, "mod-inventory", result)
}

func TestGetModuleNameFromID_ComplexName(t *testing.T) {
	// Arrange
	moduleID := "mod-inventory-storage-25.0.0"

	// Act
	result := helpers.GetModuleNameFromID(moduleID)

	// Assert
	assert.Equal(t, "mod-inventory-storage", result)
}

func TestGetOptionalModuleVersion_WithVersion(t *testing.T) {
	// Arrange
	moduleID := "mod-users-19.3.0"

	// Act
	result := helpers.GetOptionalModuleVersion(moduleID)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, "19.3.0", *result)
}

func TestGetOptionalModuleVersion_NoVersion(t *testing.T) {
	// Arrange - Pattern doesn't match, returns original string minus module name
	moduleID := "mod-users"

	// Act
	result := helpers.GetOptionalModuleVersion(moduleID)

	// Assert - When pattern doesn't match fully, returns non-empty result
	assert.NotNil(t, result)
	assert.Equal(t, "-users", *result)
}

func TestGetModuleVersionFromID_StandardVersion(t *testing.T) {
	// Arrange
	moduleID := "mod-users-19.3.0"

	// Act
	result := helpers.GetModuleVersionFromID(moduleID)

	// Assert
	assert.Equal(t, "19.3.0", result)
}

func TestGetModuleVersionFromID_SnapshotVersion(t *testing.T) {
	// Arrange
	moduleID := "mod-users-19.3.0-SNAPSHOT.456"

	// Act
	result := helpers.GetModuleVersionFromID(moduleID)

	// Assert
	assert.Equal(t, "19.3.0-SNAPSHOT.456", result)
}

func TestGetModuleVersionFromID_NoVersion(t *testing.T) {
	// Arrange - Pattern extracts groups 2 and 3, even if not a valid version
	moduleID := "mod-users"

	// Act
	result := helpers.GetModuleVersionFromID(moduleID)

	// Assert - Returns whatever matches groups 2+3 in the pattern
	assert.Equal(t, "-users", result)
}

func TestGetKafkaConsumerLagFromLogLine_WithNewlines(t *testing.T) {
	// Arrange - Pattern `[\r\n\s-]+` removes newlines, spaces, and dashes
	var stdout bytes.Buffer
	stdout.WriteString("Consumer lag: 100\nAnother line\n")

	// Act
	result := helpers.GetKafkaConsumerLagFromLogLine(stdout)

	// Assert
	assert.Equal(t, "Consumerlag:100Anotherline", result)
}

func TestGetKafkaConsumerLagFromLogLine_EmptyBuffer(t *testing.T) {
	// Arrange
	var stdout bytes.Buffer

	// Act
	result := helpers.GetKafkaConsumerLagFromLogLine(stdout)

	// Assert
	assert.Equal(t, "", result)
}

func TestMatchesModuleName_Matching(t *testing.T) {
	// Arrange
	moduleID := "mod-users-19.3.0"
	moduleName := "mod-users"

	// Act
	result := helpers.MatchesModuleName(moduleID, moduleName)

	// Assert
	assert.True(t, result)
}

func TestMatchesModuleName_NotMatching(t *testing.T) {
	// Arrange
	moduleID := "mod-inventory-19.3.0"
	moduleName := "mod-users"

	// Act
	result := helpers.MatchesModuleName(moduleID, moduleName)

	// Assert
	assert.False(t, result)
}

func TestMatchesModuleName_NoVersion(t *testing.T) {
	// Arrange
	moduleID := "mod-users"
	moduleName := "mod-users"

	// Act
	result := helpers.MatchesModuleName(moduleID, moduleName)

	// Assert
	assert.False(t, result)
}

func TestMatchesModuleName_PartialMatch(t *testing.T) {
	// Arrange
	moduleID := "mod-users-extra-19.3.0"
	moduleName := "mod-users"

	// Act
	result := helpers.MatchesModuleName(moduleID, moduleName)

	// Assert
	assert.False(t, result)
}
