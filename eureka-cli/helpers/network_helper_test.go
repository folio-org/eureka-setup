package helpers_test

import (
	"testing"

	"github.com/folio-org/eureka-cli/helpers"
	"github.com/stretchr/testify/assert"
)

func TestIsHostnameReachable_ValidHostname(t *testing.T) {
	// Arrange
	actionName := "TestAction"
	hostname := "localhost"

	// Act
	err := helpers.IsHostnameReachable(actionName, hostname)

	// Assert
	assert.NoError(t, err)
}

func TestIsHostnameReachable_InvalidHostname(t *testing.T) {
	// Arrange
	actionName := "TestAction"
	hostname := "this-hostname-should-not-exist-12345.invalid"

	// Act
	err := helpers.IsHostnameReachable(actionName, hostname)

	// Assert
	assert.Error(t, err)
}

func TestConstructURL_WithHTTPPrefix(t *testing.T) {
	// Arrange
	url := "http://example.com/api"
	gatewayURL := "http://gateway.com"

	// Act
	result := helpers.ConstructURL(url, gatewayURL)

	// Assert
	assert.Equal(t, "http://example.com/api", result)
}

func TestConstructURL_WithHTTPSPrefix(t *testing.T) {
	// Arrange
	url := "https://example.com/api"
	gatewayURL := "http://gateway.com"

	// Act
	result := helpers.ConstructURL(url, gatewayURL)

	// Assert
	assert.Equal(t, "https://example.com/api", result)
}

func TestConstructURL_WithoutPrefix(t *testing.T) {
	// Arrange
	url := "8080"
	gatewayURL := "http://gateway.com"

	// Act
	result := helpers.ConstructURL(url, gatewayURL)

	// Assert
	assert.Equal(t, "http://gateway.com:8080", result)
}

func TestExtractPortFromURL_ValidURL(t *testing.T) {
	// Arrange - Pattern removes everything up to and including last colon
	url := "http://localhost:9000"

	// Act
	result, err := helpers.ExtractPortFromURL(url)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 9000, result)
}

func TestExtractPortFromURL_InvalidPort(t *testing.T) {
	// Arrange
	url := "http://localhost:invalid/api"

	// Act
	_, err := helpers.ExtractPortFromURL(url)

	// Assert
	assert.Error(t, err)
}

func TestTenantSecureApplicationJSONHeaders_ValidInputs(t *testing.T) {
	// Arrange
	tenantName := "diku"
	accessToken := "token123"

	// Act
	result := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, accessToken)

	// Assert
	assert.Len(t, result, 3)
	assert.Equal(t, "application/json", result["Content-Type"])
	assert.Equal(t, "diku", result["X-Okapi-Tenant"])
	assert.Equal(t, "token123", result["X-Okapi-Token"])
}

func TestTenantSecureNonOkapiApplicationJSONHeaders_ValidInputs(t *testing.T) {
	// Arrange
	tenantName := "diku"
	accessToken := "token123"

	// Act
	result := helpers.SecureTenantApplicationJSONHeaders(tenantName, accessToken)

	// Assert
	assert.Len(t, result, 3)
	assert.Equal(t, "application/json", result["Content-Type"])
	assert.Equal(t, "diku", result["X-Okapi-Tenant"])
	assert.Equal(t, "Bearer token123", result["Authorization"])
}

func TestSecureApplicationJSONHeaders_ValidToken(t *testing.T) {
	// Arrange
	accessToken := "token123"

	// Act
	result := helpers.SecureApplicationJSONHeaders(accessToken)

	// Assert
	assert.Len(t, result, 2)
	assert.Equal(t, "application/json", result["Content-Type"])
	assert.Equal(t, "Bearer token123", result["Authorization"])
}

func TestApplicationFormURLEncodedHeaders_NoInputs(t *testing.T) {
	// Act
	result := helpers.ApplicationFormURLEncodedHeaders()

	// Assert
	assert.Len(t, result, 1)
	assert.Equal(t, "application/x-www-form-urlencoded", result["Content-Type"])
}
