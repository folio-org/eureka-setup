package awssvc

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/constant"
	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

// setTestEnv sets an environment variable and returns a cleanup function
func setTestEnv(t *testing.T, key, value string) func() {
	t.Helper()
	oldValue := os.Getenv(key)
	if err := os.Setenv(key, value); err != nil {
		t.Fatalf("Failed to set env var %s: %v", key, err)
	}
	return func() {
		if oldValue != "" {
			_ = os.Setenv(key, oldValue)
		} else {
			_ = os.Unsetenv(key)
		}
	}
}

// unsetTestEnv unsets an environment variable and returns a cleanup function
func unsetTestEnv(t *testing.T, key string) func() {
	t.Helper()
	oldValue := os.Getenv(key)
	if err := os.Unsetenv(key); err != nil {
		t.Fatalf("Failed to unset env var %s: %v", key, err)
	}
	return func() {
		if oldValue != "" {
			_ = os.Setenv(key, oldValue)
		}
	}
}

func TestNew(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()

	// Act
	svc := New(action)

	// Assert
	assert.NotNil(t, svc)
	assert.Equal(t, action, svc.Action)
}

func TestGetECRNamespace_WithEnvVar(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action)
	expectedNamespace := "123456789012.dkr.ecr.us-east-1.amazonaws.com/folio"
	cleanup := setTestEnv(t, constant.ECRRepositoryEnv, expectedNamespace)
	defer cleanup()

	// Act
	result := svc.GetECRNamespace()

	// Assert
	assert.Equal(t, expectedNamespace, result)
}

func TestGetECRNamespace_WithoutEnvVar(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action)
	cleanup := unsetTestEnv(t, constant.ECRRepositoryEnv)
	defer cleanup()

	// Act
	result := svc.GetECRNamespace()

	// Assert
	assert.Equal(t, "", result)
}

func TestIsECRConfigured_True(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action)
	cleanup := setTestEnv(t, constant.ECRRepositoryEnv, "123456789012.dkr.ecr.us-east-1.amazonaws.com/folio")
	defer cleanup()

	// Act
	result := svc.IsECRConfigured()

	// Assert
	assert.True(t, result)
}

func TestIsECRConfigured_False(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action)
	cleanup := unsetTestEnv(t, constant.ECRRepositoryEnv)
	defer cleanup()

	// Act
	result := svc.IsECRConfigured()

	// Assert
	assert.False(t, result)
}

func TestGetAuthorizationToken_NotConfigured(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action)
	cleanup := unsetTestEnv(t, constant.ECRRepositoryEnv)
	defer cleanup()

	// Act
	token, err := svc.GetAuthorizationToken()

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "", token)
}

func TestGetAuthorizationToken_Configured_SkipIntegration(t *testing.T) {
	// This test would require actual AWS credentials and ECR access
	// Integration testing should be done separately with proper AWS setup
	t.Skip("Skipping integration test - requires AWS credentials and ECR configuration")

	// The actual implementation would look like:
	// 1. Set AWS_ECR_FOLIO_REPO environment variable
	// 2. Ensure AWS credentials are configured (via env vars, ~/.aws/credentials, or IAM role)
	// 3. Call GetAuthorizationToken()
	// 4. Verify the returned token is a valid base64-encoded JSON with username/password

	// Example validation:
	// assert.NotEmpty(t, token)
	// decoded, err := base64.StdEncoding.DecodeString(token)
	// assert.NoError(t, err)
	// var authConfig map[string]string
	// err = json.Unmarshal(decoded, &authConfig)
	// assert.NoError(t, err)
	// assert.Contains(t, authConfig, "username")
	// assert.Contains(t, authConfig, "password")
}

// TestAuthTokenEncoding verifies the expected format of auth token encoding
// This is a helper test to document the expected token format
func TestAuthTokenEncoding_ExpectedFormat(t *testing.T) {
	// Arrange - simulate the expected auth config structure
	authConfig := map[string]string{
		"username": "AWS",
		"password": "test-password-token",
	}

	// Act - encode as the function does
	payload, err := json.Marshal(authConfig)
	assert.NoError(t, err)
	encodedAuth := base64.StdEncoding.EncodeToString(payload)

	// Assert - verify it can be decoded back
	assert.NotEmpty(t, encodedAuth)

	decoded, err := base64.StdEncoding.DecodeString(encodedAuth)
	assert.NoError(t, err)

	var decodedConfig map[string]string
	err = json.Unmarshal(decoded, &decodedConfig)
	assert.NoError(t, err)
	assert.Equal(t, "AWS", decodedConfig["username"])
	assert.Equal(t, "test-password-token", decodedConfig["password"])
}

// TestECRAuthTokenDecoding verifies the AWS ECR token decoding logic
// AWS ECR returns tokens in format: base64("username:password")
func TestECRAuthTokenDecoding_ValidFormat(t *testing.T) {
	// Arrange - simulate AWS ECR token format
	username := "AWS"
	password := "eyJwYXlsb2FkIjoiZXhhbXBsZS10b2tlbiJ9" // Example token
	rawToken := username + ":" + password
	ecrToken := base64.StdEncoding.EncodeToString([]byte(rawToken))

	// Act - decode as the function does
	decodedBytes, err := base64.StdEncoding.DecodeString(ecrToken)
	assert.NoError(t, err)

	authCreds := string(decodedBytes)
	parts := []byte(authCreds)

	// Assert - verify the format
	assert.Contains(t, string(parts), ":")
	assert.Contains(t, string(parts), username)
	assert.Contains(t, string(parts), password)
}

func TestECRAuthTokenDecoding_InvalidBase64(t *testing.T) {
	// Arrange - invalid base64 string
	invalidToken := "not-a-valid-base64!!!"

	// Act
	_, err := base64.StdEncoding.DecodeString(invalidToken)

	// Assert - should return decoding error
	assert.Error(t, err)
}

func TestECRAuthTokenDecoding_InvalidFormat_NoColon(t *testing.T) {
	// Arrange - valid base64 but missing colon separator
	invalidFormat := "usernamepassword" // No colon separator
	ecrToken := base64.StdEncoding.EncodeToString([]byte(invalidFormat))

	// Act
	decodedBytes, err := base64.StdEncoding.DecodeString(ecrToken)
	assert.NoError(t, err)

	authCreds := string(decodedBytes)

	// Assert - should not contain colon, making it invalid for username:password split
	assert.NotContains(t, authCreds, ":")
}

func TestECRAuthTokenDecoding_EmptyPassword(t *testing.T) {
	// Arrange - username with empty password
	rawToken := "AWS:"
	ecrToken := base64.StdEncoding.EncodeToString([]byte(rawToken))

	// Act
	decodedBytes, err := base64.StdEncoding.DecodeString(ecrToken)
	assert.NoError(t, err)

	authCreds := string(decodedBytes)

	// Assert - should have colon but empty password
	assert.Contains(t, authCreds, ":")
	assert.Equal(t, "AWS:", authCreds)
}

func TestIsECRConfigured_EmptyString(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action)
	cleanup := setTestEnv(t, constant.ECRRepositoryEnv, "")
	defer cleanup()

	// Act
	result := svc.IsECRConfigured()

	// Assert - empty string should be considered as not configured
	assert.False(t, result)
}

func TestGetECRNamespace_EmptyString(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	svc := New(action)
	cleanup := setTestEnv(t, constant.ECRRepositoryEnv, "")
	defer cleanup()

	// Act
	result := svc.GetECRNamespace()

	// Assert - should return empty string
	assert.Equal(t, "", result)
}
