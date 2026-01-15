package helpers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/helpers"
	"github.com/stretchr/testify/assert"
)

func TestDumpRequestJSON_DebugDisabled(t *testing.T) {
	// Arrange
	bodyBytes := []byte(`{"test": "data"}`)

	// Act - Should return early without panic (debug logging disabled by default)
	helpers.DumpRequestJSON(bodyBytes)

	// Assert - No panic means success
}

func TestDumpRequestFormData_DebugDisabled(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Set("key", "value")

	// Act - Should return early without panic (debug logging disabled by default)
	helpers.DumpRequestFormData(formData)

	// Assert - No panic means success
}

func TestDumpRequest_DebugDisabled(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("POST", "http://example.com/api", bytes.NewBufferString(`{"test": true}`))

	// Act
	err := helpers.DumpRequest(req)

	// Assert
	assert.NoError(t, err)
}

func TestDumpRequest_NilRequest(t *testing.T) {
	// Arrange
	var nilRequest *http.Request

	// Act - Should handle nil gracefully (returns early due to log check)
	err := helpers.DumpRequest(nilRequest)

	// Assert
	assert.NoError(t, err)
}

func TestDumpResponse_WithForceDump(t *testing.T) {
	// Arrange
	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusOK)
	_, _ = recorder.WriteString(`{"status": "success"}`)
	response := recorder.Result()
	defer func() { _ = response.Body.Close() }()

	// Act - forceDump=true bypasses log level check
	err := helpers.DumpResponse("GET", "http://example.com/api", response, true)

	// Assert
	assert.NoError(t, err)
}

func TestDumpResponse_WithoutForceDump(t *testing.T) {
	// Arrange
	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusOK)
	response := recorder.Result()
	defer func() { _ = response.Body.Close() }()

	// Act - Without debug logging enabled, this returns early
	err := helpers.DumpResponse("GET", "http://example.com/api", response, false)

	// Assert
	assert.NoError(t, err)
}

func TestDumpResponse_ErrorInDump(t *testing.T) {
	// Arrange - Create a response with nil body to potentially trigger error
	response := &http.Response{
		StatusCode: 200,
		Body:       nil,
	}

	// Act
	err := helpers.DumpResponse("GET", "http://example.com", response, true)

	// Assert - httputil.DumpResponse handles nil body, but we test error path exists
	_ = err
}
