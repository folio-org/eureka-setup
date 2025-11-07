package helpers

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

// These tests are in the same package (helpers) not helpers_test
// so we can test the unexported internal functions

func Test_dumpRequestJSONInternal(t *testing.T) {
	// Arrange
	bodyBytes := []byte(`{"test": "data", "value": 123}`)

	// Act - Should not panic
	dumpRequestJSONInternal(bodyBytes)

	// Assert - If we got here without panic, success
}

func Test_dumpRequestJSONInternal_EmptyBody(t *testing.T) {
	// Arrange
	bodyBytes := []byte{}

	// Act
	dumpRequestJSONInternal(bodyBytes)

	// Assert - No panic
}

func Test_dumpRequestFormDataInternal(t *testing.T) {
	// Arrange
	formData := url.Values{}
	formData.Set("username", "test")
	formData.Set("password", "secret")

	// Act
	dumpRequestFormDataInternal(formData)

	// Assert - No panic
}

func Test_dumpRequestFormDataInternal_EmptyForm(t *testing.T) {
	// Arrange
	formData := url.Values{}

	// Act
	dumpRequestFormDataInternal(formData)

	// Assert - No panic
}

func Test_dumpRequestInternal_ValidRequest(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("POST", "http://example.com/api", bytes.NewBufferString(`{"test": true}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer token123")

	// Act
	err := dumpRequestInternal(req)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func Test_dumpRequestInternal_GetRequest(t *testing.T) {
	// Arrange
	req := httptest.NewRequest("GET", "http://example.com/users?page=1", nil)

	// Act
	err := dumpRequestInternal(req)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func Test_dumpResponseInternal_Success(t *testing.T) {
	// Arrange
	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusOK)
	_, _ = recorder.WriteString(`{"id": 1, "name": "test"}`)
	response := recorder.Result()
	defer func() { _ = response.Body.Close() }()

	// Act
	err := dumpResponseInternal("GET", "http://example.com/api/users/1", response)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func Test_dumpResponseInternal_ErrorResponse(t *testing.T) {
	// Arrange
	recorder := httptest.NewRecorder()
	recorder.WriteHeader(http.StatusInternalServerError)
	_, _ = recorder.WriteString(`{"error": "Internal Server Error"}`)
	response := recorder.Result()
	defer func() { _ = response.Body.Close() }()

	// Act
	err := dumpResponseInternal("POST", "http://example.com/api/create", response)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func Test_dumpResponseInternal_NilBody(t *testing.T) {
	// Arrange
	response := &http.Response{
		StatusCode: 204,
		Header:     http.Header{},
		Body:       nil,
	}

	// Act
	err := dumpResponseInternal("DELETE", "http://example.com/api/delete/1", response)

	// Assert - May or may not error, but shouldn't panic
	_ = err
}
