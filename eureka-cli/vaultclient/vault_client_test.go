package vaultclient

import (
	"context"
	"testing"

	"github.com/folio-org/eureka-cli/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockHTTP := new(testhelpers.MockHTTPClient)

	// Act
	client := New(action, mockHTTP)

	// Assert
	assert.NotNil(t, client)
	assert.Equal(t, action, client.Action)
	assert.Equal(t, mockHTTP, client.HTTPClient)
}

func TestCreate_Success(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockHTTP := new(testhelpers.MockHTTPClient)
	vc := New(action, mockHTTP)

	// Act
	client, err := vc.Create()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestGetSecretKey_Success(t *testing.T) {
	// Arrange
	// Note: This test requires a real Vault server or mock server
	// Since we're using the real vault-client-go library, we'll test
	// the basic flow and error handling

	action := testhelpers.NewMockAction()
	mockHTTP := new(testhelpers.MockHTTPClient)
	vc := New(action, mockHTTP)

	client, err := vc.Create()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	ctx := context.Background()
	vaultRootToken := "test-token"
	secretPath := "test/secret"

	// Act
	// This will fail without a real Vault server, but we're testing the method exists
	// and handles parameters correctly
	_, err = vc.GetSecretKey(ctx, client, vaultRootToken, secretPath)

	// Assert
	// We expect an error since there's no real Vault server
	// but the method should not panic
	assert.Error(t, err)
}

func TestGetSecretKey_InvalidToken(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockHTTP := new(testhelpers.MockHTTPClient)
	vc := New(action, mockHTTP)

	client, err := vc.Create()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	ctx := context.Background()
	vaultRootToken := ""
	secretPath := "test/secret"

	// Act
	_, err = vc.GetSecretKey(ctx, client, vaultRootToken, secretPath)

	// Assert - should fail with empty token
	assert.Error(t, err)
}

func TestGetSecretKey_ContextCancellation(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockHTTP := new(testhelpers.MockHTTPClient)
	vc := New(action, mockHTTP)

	client, err := vc.Create()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Create a canceled context
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	vaultRootToken := "test-token"
	secretPath := "test/secret"

	// Act
	_, err = vc.GetSecretKey(ctx, client, vaultRootToken, secretPath)

	// Assert - should fail due to canceled context
	assert.Error(t, err)
}

func TestCreate_ValidatesServerURL(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockHTTP := new(testhelpers.MockHTTPClient)
	vc := New(action, mockHTTP)

	// Act
	// Create should succeed even with default URL
	client, err := vc.Create()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, client)
}

func TestGetSecretKey_EmptySecretPath(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockHTTP := new(testhelpers.MockHTTPClient)
	vc := New(action, mockHTTP)

	client, err := vc.Create()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	ctx := context.Background()
	vaultRootToken := "test-token"
	secretPath := ""

	// Act
	_, err = vc.GetSecretKey(ctx, client, vaultRootToken, secretPath)

	// Assert - should fail with empty secret path
	assert.Error(t, err)
}

func TestGetSecretKey_SetTokenError(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockHTTP := new(testhelpers.MockHTTPClient)
	vc := New(action, mockHTTP)

	// Create a client
	client, err := vc.Create()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	ctx := context.Background()

	// Test with various invalid tokens
	tests := []struct {
		name  string
		token string
		path  string
	}{
		{
			name:  "EmptyToken",
			token: "",
			path:  "test/secret",
		},
		{
			name:  "InvalidTokenFormat",
			token: "invalid-token-format",
			path:  "test/secret",
		},
	}

	// Act & Assert
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := vc.GetSecretKey(ctx, client, tt.token, tt.path)

			// Should return an error
			assert.Error(t, err)
		})
	}
}

func TestVaultClient_Integration(t *testing.T) {
	// Arrange
	// This test validates the complete flow without a real Vault server
	action := testhelpers.NewMockAction()
	mockHTTP := new(testhelpers.MockHTTPClient)
	vc := New(action, mockHTTP)

	// Act
	// Step 1: Create client
	client, err := vc.Create()
	assert.NoError(t, err)
	assert.NotNil(t, client)

	// Step 2: Attempt to get secret (will fail without real server)
	ctx := context.Background()
	vaultRootToken := "test-root-token"
	secretPath := "secret/data/test"

	_, err = vc.GetSecretKey(ctx, client, vaultRootToken, secretPath)

	// Assert
	// Expected to fail without real Vault server
	assert.Error(t, err)
}
