package httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/folio-org/eureka-cli/action"
)

func TestHTTPClientErrorHandling(t *testing.T) {
	// Test server that returns error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal server error"}`))
	}))
	defer server.Close()

	action := &action.Action{Name: "test-error"}
	client := New(action)

	// Test that error status codes are handled properly
	_, err := client.GetReturnResponse(server.URL, map[string]string{})
	if err == nil {
		t.Error("Expected error for 500 status, but got none")
	}

	err = client.PostReturnNoContent(server.URL, []byte(`{}`), map[string]string{})
	if err == nil {
		t.Error("Expected error for 500 status, but got none")
	}

	err = client.PutReturnNoContent(server.URL, []byte(`{}`), map[string]string{})
	if err == nil {
		t.Error("Expected error for 500 status, but got none")
	}

	err = client.Delete(server.URL, map[string]string{})
	if err == nil {
		t.Error("Expected error for 500 status, but got none")
	}
}

func TestHTTPClientWithCustomHeaders(t *testing.T) {
	customHeaders := map[string]string{
		"Authorization":   "Bearer token123",
		"X-Custom-Header": "custom-value",
		"Content-Type":    "application/xml",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify custom headers
		for key, expectedValue := range customHeaders {
			actualValue := r.Header.Get(key)
			if actualValue != expectedValue {
				t.Errorf("Header %s = %q, want %q", key, actualValue, expectedValue)
			}
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	action := &action.Action{Name: "test-headers"}
	client := New(action)

	_, err := client.GetReturnResponse(server.URL, customHeaders)
	if err != nil {
		t.Fatalf("GetReturnResponse() with custom headers error = %v", err)
	}
}
