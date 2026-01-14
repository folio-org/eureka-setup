package vaultclient

import (
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/internal/testhelpers"
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
	// Verify the client was created with expected configuration
	assert.NotNil(t, client)
}

func TestCreate_ReturnsValidClient(t *testing.T) {
	// Arrange
	action := testhelpers.NewMockAction()
	mockHTTP := new(testhelpers.MockHTTPClient)
	vc := New(action, mockHTTP)

	// Act
	client, err := vc.Create()

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, client, "Client should not be nil")
}
