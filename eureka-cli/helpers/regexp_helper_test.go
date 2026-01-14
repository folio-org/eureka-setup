package helpers_test

import (
	"bytes"
	"testing"

	"github.com/j011195/eureka-setup/eureka-cli/helpers"
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
	// Arrange
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

func TestGetPortFromURL_WithPath(t *testing.T) {
	// Arrange
	url := "http://localhost:8080/api/v1"

	// Act
	result, err := helpers.GetPortFromURL(url)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 8080, result)
}

func TestGetPortFromURL_NoPort(t *testing.T) {
	// Arrange
	url := "http://localhost"

	// Act
	_, err := helpers.GetPortFromURL(url)

	// Assert
	assert.Error(t, err)
}

func TestGetPortFromURL_WithoutProtocol(t *testing.T) {
	// Arrange
	url := "localhost:9000"

	// Act
	result, err := helpers.GetPortFromURL(url)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 9000, result)
}

func TestGetPortFromURL_WithQueryString(t *testing.T) {
	// Arrange
	url := "http://localhost:8080?param=value"

	// Act
	result, err := helpers.GetPortFromURL(url)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 8080, result)
}

func TestGetPortFromURL_PurePortNumber(t *testing.T) {
	// Arrange
	url := "30300"

	// Act
	result, err := helpers.GetPortFromURL(url)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 30300, result)
}

func TestGetPortFromURL_ColonPrefixedPort(t *testing.T) {
	// Arrange
	url := ":30300"

	// Act
	result, err := helpers.GetPortFromURL(url)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 30300, result)
}

func TestGetPortFromURL_PurePortNumberWithWhitespace(t *testing.T) {
	// Arrange
	url := "  8081  "

	// Act
	result, err := helpers.GetPortFromURL(url)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 8081, result)
}

func TestGetPortFromURL_ColonPrefixedPortWithPath(t *testing.T) {
	// Arrange
	url := ":9000/api/health"

	// Act
	result, err := helpers.GetPortFromURL(url)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 9000, result)
}

func TestGetPortFromURL_InvalidPurePort(t *testing.T) {
	// Arrange
	url := "invalid"

	// Act
	_, err := helpers.GetPortFromURL(url)

	// Assert
	assert.Error(t, err)
}

func TestGetHostnameFromURL_WithHTTPProtocol(t *testing.T) {
	// Arrange
	url := "http://host.docker.internal:8081"

	// Act
	result := helpers.GetHostnameFromURL(url)

	// Assert
	assert.Equal(t, "host.docker.internal", result)
}

func TestGetHostnameFromURL_WithHTTPSProtocol(t *testing.T) {
	// Arrange
	url := "https://example.com:443"

	// Act
	result := helpers.GetHostnameFromURL(url)

	// Assert
	assert.Equal(t, "example.com", result)
}

func TestGetHostnameFromURL_WithoutProtocol(t *testing.T) {
	// Arrange
	url := "localhost:9000"

	// Act
	result := helpers.GetHostnameFromURL(url)

	// Assert
	assert.Equal(t, "localhost", result)
}

func TestGetHostnameFromURL_IPAddress(t *testing.T) {
	// Arrange
	url := "192.168.1.1:8080"

	// Act
	result := helpers.GetHostnameFromURL(url)

	// Assert
	assert.Equal(t, "192.168.1.1", result)
}

func TestGetHostnameFromURL_IPAddressWithoutPort(t *testing.T) {
	// Arrange
	url := "192.168.1.1"

	// Act
	result := helpers.GetHostnameFromURL(url)

	// Assert
	assert.Equal(t, "192.168.1.1", result)
}

func TestGetHostnameFromURL_HostnameOnly(t *testing.T) {
	// Arrange
	url := "host.docker.internal"

	// Act
	result := helpers.GetHostnameFromURL(url)

	// Assert
	assert.Equal(t, "host.docker.internal", result)
}

func TestGetHostnameFromURL_WithPath(t *testing.T) {
	// Arrange
	url := "http://example.com:8080/api/v1"

	// Act
	result := helpers.GetHostnameFromURL(url)

	// Assert
	assert.Equal(t, "example.com", result)
}

func TestGetHostnameFromURL_WithQueryString(t *testing.T) {
	// Arrange
	url := "http://example.com/api?param=value"

	// Act
	result := helpers.GetHostnameFromURL(url)

	// Assert
	assert.Equal(t, "example.com", result)
}

func TestGetHostnameFromURL_WithFragment(t *testing.T) {
	// Arrange
	url := "http://example.com#section"

	// Act
	result := helpers.GetHostnameFromURL(url)

	// Assert
	assert.Equal(t, "example.com", result)
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
	// Arrange
	var stdout bytes.Buffer
	stdout.WriteString("100\n50\n")

	// Act
	result := helpers.GetKafkaConsumerLagFromLogLine(stdout)

	// Assert
	assert.Equal(t, "150", result)
}

func TestGetKafkaConsumerLagFromLogLine_EmptyBuffer(t *testing.T) {
	// Arrange
	var stdout bytes.Buffer

	// Act
	result := helpers.GetKafkaConsumerLagFromLogLine(stdout)

	// Assert
	assert.Equal(t, "0", result)
}

func TestGetKafkaConsumerLagFromLogLine_WithDashes(t *testing.T) {
	// Arrange - Dashes should be treated as 0
	var stdout bytes.Buffer
	stdout.WriteString("-\n0")

	// Act
	result := helpers.GetKafkaConsumerLagFromLogLine(stdout)

	// Assert
	assert.Equal(t, "0", result)
}

func TestGetKafkaConsumerLagFromLogLine_MixedValues(t *testing.T) {
	// Arrange
	var stdout bytes.Buffer
	stdout.WriteString("-\n100\n-\n50\n25")

	// Act
	result := helpers.GetKafkaConsumerLagFromLogLine(stdout)

	// Assert
	assert.Equal(t, "175", result)
}

func TestGetKafkaConsumerLagFromLogLine_SingleValue(t *testing.T) {
	// Arrange
	var stdout bytes.Buffer
	stdout.WriteString("42")

	// Act
	result := helpers.GetKafkaConsumerLagFromLogLine(stdout)

	// Assert
	assert.Equal(t, "42", result)
}

func TestGetKafkaConsumerLagFromLogLine_AllDashes(t *testing.T) {
	// Arrange
	var stdout bytes.Buffer
	stdout.WriteString("-\n-\n-")

	// Act
	result := helpers.GetKafkaConsumerLagFromLogLine(stdout)

	// Assert
	assert.Equal(t, "0", result)
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
