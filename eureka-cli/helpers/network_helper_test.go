package helpers_test

import (
	"errors"
	"testing"

	apperrors "github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/j011195/eureka-setup/eureka-cli/helpers"
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

func TestSecureOkapiApplicationJSONHeaders_ValidToken(t *testing.T) {
	// Arrange
	accessToken := "token123"

	// Act
	result, err := helpers.SecureOkapiApplicationJSONHeaders(accessToken)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "application/json", result["Content-Type"])
	assert.Equal(t, "token123", result["X-Okapi-Token"])
}

func TestSecureOkapiApplicationJSONHeaders_BlankToken(t *testing.T) {
	// Arrange
	accessToken := ""

	// Act
	result, err := helpers.SecureOkapiApplicationJSONHeaders(accessToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "access token cannot be blank")
	assert.True(t, errors.Is(err, apperrors.AccessTokenBlank()))
}

func TestTenantSecureApplicationJSONHeaders_ValidInputs(t *testing.T) {
	// Arrange
	tenantName := "diku"
	accessToken := "token123"

	// Act
	result, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, accessToken)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "application/json", result["Content-Type"])
	assert.Equal(t, "diku", result["X-Okapi-Tenant"])
	assert.Equal(t, "token123", result["X-Okapi-Token"])
}

func TestSecureOkapiTenantApplicationJSONHeaders_BlankTenant(t *testing.T) {
	// Arrange
	tenantName := ""
	accessToken := "token123"

	// Act
	result, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, accessToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "tenant name cannot be blank")
	assert.True(t, errors.Is(err, apperrors.TenantNameBlank()))
}

func TestSecureOkapiTenantApplicationJSONHeaders_BlankToken(t *testing.T) {
	// Arrange
	tenantName := "diku"
	accessToken := ""

	// Act
	result, err := helpers.SecureOkapiTenantApplicationJSONHeaders(tenantName, accessToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "access token cannot be blank")
	assert.True(t, errors.Is(err, apperrors.AccessTokenBlank()))
}

func TestTenantSecureNonOkapiApplicationJSONHeaders_ValidInputs(t *testing.T) {
	// Arrange
	tenantName := "diku"
	accessToken := "token123"

	// Act
	result, err := helpers.SecureTenantApplicationJSONHeaders(tenantName, accessToken)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 3)
	assert.Equal(t, "application/json", result["Content-Type"])
	assert.Equal(t, "diku", result["X-Okapi-Tenant"])
	assert.Equal(t, "Bearer token123", result["Authorization"])
}

func TestSecureTenantApplicationJSONHeaders_BlankTenant(t *testing.T) {
	// Arrange
	tenantName := ""
	accessToken := "token123"

	// Act
	result, err := helpers.SecureTenantApplicationJSONHeaders(tenantName, accessToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "tenant name cannot be blank")
	assert.True(t, errors.Is(err, apperrors.TenantNameBlank()))
}

func TestSecureTenantApplicationJSONHeaders_BlankToken(t *testing.T) {
	// Arrange
	tenantName := "diku"
	accessToken := ""

	// Act
	result, err := helpers.SecureTenantApplicationJSONHeaders(tenantName, accessToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "access token cannot be blank")
	assert.True(t, errors.Is(err, apperrors.AccessTokenBlank()))
}

func TestSecureApplicationJSONHeaders_ValidToken(t *testing.T) {
	// Arrange
	accessToken := "token123"

	// Act
	result, err := helpers.SecureApplicationJSONHeaders(accessToken)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "application/json", result["Content-Type"])
	assert.Equal(t, "Bearer token123", result["Authorization"])
}

func TestSecureApplicationJSONHeaders_BlankToken(t *testing.T) {
	// Arrange
	accessToken := ""

	// Act
	result, err := helpers.SecureApplicationJSONHeaders(accessToken)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.EqualError(t, err, "access token cannot be blank")
	assert.True(t, errors.Is(err, apperrors.AccessTokenBlank()))
}

func TestApplicationFormURLEncodedHeaders_NoInputs(t *testing.T) {
	// Act
	result := helpers.ApplicationFormURLEncodedHeaders()

	// Assert
	assert.Len(t, result, 1)
	assert.Equal(t, "application/x-www-form-urlencoded", result["Content-Type"])
}

func TestGetSidecarURL_EdgeModule(t *testing.T) {
	// Arrange
	moduleName := "edge-oai-pmh"
	privatePort := 8081

	// Act
	result := helpers.GetSidecarURL(moduleName, privatePort)

	// Assert
	assert.Equal(t, "http://edge-oai-pmh.eureka:8081", result)
}

func TestGetSidecarURL_EdgeModuleWithDifferentPort(t *testing.T) {
	// Arrange
	moduleName := "edge-orders"
	privatePort := 9000

	// Act
	result := helpers.GetSidecarURL(moduleName, privatePort)

	// Assert
	assert.Equal(t, "http://edge-orders.eureka:9000", result)
}

func TestGetSidecarURL_NonEdgeModule(t *testing.T) {
	// Arrange
	moduleName := "mod-inventory"
	privatePort := 8081

	// Act
	result := helpers.GetSidecarURL(moduleName, privatePort)

	// Assert
	assert.Equal(t, "http://mod-inventory-sc.eureka:8081", result)
}

func TestGetSidecarURL_NonEdgeModuleWithDifferentPort(t *testing.T) {
	// Arrange
	moduleName := "mod-users"
	privatePort := 9090

	// Act
	result := helpers.GetSidecarURL(moduleName, privatePort)

	// Assert
	assert.Equal(t, "http://mod-users-sc.eureka:9090", result)
}

func TestGetSidecarURL_EmptyModuleName(t *testing.T) {
	// Arrange
	moduleName := ""
	privatePort := 8081

	// Act
	result := helpers.GetSidecarURL(moduleName, privatePort)

	// Assert
	assert.Equal(t, "http://-sc.eureka:8081", result)
}

func TestGetSidecarURL_ModuleNameStartingWithEdgeButNotFullWord(t *testing.T) {
	// Arrange
	moduleName := "edges-test"
	privatePort := 8081

	// Act
	result := helpers.GetSidecarURL(moduleName, privatePort)

	// Assert
	assert.Equal(t, "http://edges-test.eureka:8081", result)
}
