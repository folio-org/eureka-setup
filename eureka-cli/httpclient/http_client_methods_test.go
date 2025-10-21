package httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/folio-org/eureka-cli/action"
)

func TestGetReturnResponse(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Check headers
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type header to be application/json, got %s", r.Header.Get("Content-Type"))
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"message": "success"}`))
	}))
	defer server.Close()

	action := &action.Action{Name: "test-get"}
	client := New(action)

	resp, err := client.GetReturnResponse(server.URL, map[string]string{})
	if err != nil {
		t.Fatalf("GetReturnResponse() error = %v", err)
	}
	defer CloseResponse(resp)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("GetReturnResponse() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

func TestGetDecodeReturnMapStringAny(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		statusCode     int
		expectedResult map[string]any
		expectError    bool
	}{
		{
			name:         "valid JSON response",
			responseBody: `{"name": "test", "count": 42, "enabled": true}`,
			statusCode:   http.StatusOK,
			expectedResult: map[string]any{
				"name":    "test",
				"count":   float64(42), // JSON numbers decode as float64
				"enabled": true,
			},
			expectError: false,
		},
		{
			name:           "empty JSON response",
			responseBody:   `{}`,
			statusCode:     http.StatusOK,
			expectedResult: map[string]any{},
			expectError:    false,
		},
		{
			name:         "invalid JSON response",
			responseBody: `{"invalid": json}`,
			statusCode:   http.StatusOK,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			action := &action.Action{Name: "test-get-decode"}
			client := New(action)

			result, err := client.GetDecodeReturnMapStringAny(server.URL, map[string]string{})

			if tt.expectError {
				if err == nil {
					t.Error("GetDecodeReturnMapStringAny() expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetDecodeReturnMapStringAny() unexpected error = %v", err)
			}

			if !mapsEqual(result, tt.expectedResult) {
				t.Errorf("GetDecodeReturnMapStringAny() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

func TestPostReturnNoContent(t *testing.T) {
	requestBody := []byte(`{"test": "data"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify request body
		buf := make([]byte, len(requestBody))
		n, _ := r.Body.Read(buf)
		if string(buf[:n]) != string(requestBody) {
			t.Errorf("Expected request body %s, got %s", string(requestBody), string(buf[:n]))
		}

		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	action := &action.Action{Name: "test-post"}
	client := New(action)

	err := client.PostReturnNoContent(server.URL, requestBody, map[string]string{})
	if err != nil {
		t.Fatalf("PostReturnNoContent() error = %v", err)
	}
}

func TestPostReturnMapStringAny(t *testing.T) {
	requestBody := []byte(`{"input": "test"}`)
	responseBody := `{"result": "success", "id": 123}`

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseBody))
	}))
	defer server.Close()

	action := &action.Action{Name: "test-post-return"}
	client := New(action)

	result, err := client.PostReturnMapStringAny(server.URL, requestBody, map[string]string{})
	if err != nil {
		t.Fatalf("PostReturnMapStringAny() error = %v", err)
	}

	expected := map[string]any{
		"result": "success",
		"id":     float64(123),
	}

	if !mapsEqual(result, expected) {
		t.Errorf("PostReturnMapStringAny() = %v, want %v", result, expected)
	}
}

func TestPutReturnNoContent(t *testing.T) {
	requestBody := []byte(`{"update": "data"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Errorf("Expected PUT request, got %s", r.Method)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	action := &action.Action{Name: "test-put"}
	client := New(action)

	err := client.PutReturnNoContent(server.URL, requestBody, map[string]string{})
	if err != nil {
		t.Fatalf("PutReturnNoContent() error = %v", err)
	}
}

func TestDelete(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	action := &action.Action{Name: "test-delete"}
	client := New(action)

	err := client.Delete(server.URL, map[string]string{})
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
}

func TestDeleteWithBody(t *testing.T) {
	requestBody := []byte(`{"deleteFilter": "criteria"}`)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		// Verify request body
		buf := make([]byte, len(requestBody))
		n, _ := r.Body.Read(buf)
		if string(buf[:n]) != string(requestBody) {
			t.Errorf("Expected request body %s, got %s", string(requestBody), string(buf[:n]))
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	action := &action.Action{Name: "test-delete-body"}
	client := New(action)

	err := client.DeleteWithBody(server.URL, requestBody, map[string]string{})
	if err != nil {
		t.Fatalf("DeleteWithBody() error = %v", err)
	}
}

func TestHTTPClientErrorHandling(t *testing.T) {
	// Test server that returns error status
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`))
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
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	action := &action.Action{Name: "test-headers"}
	client := New(action)

	_, err := client.GetReturnResponse(server.URL, customHeaders)
	if err != nil {
		t.Fatalf("GetReturnResponse() with custom headers error = %v", err)
	}
}

// Helper function to compare maps
func mapsEqual(a, b map[string]any) bool {
	if len(a) != len(b) {
		return false
	}

	for key, valueA := range a {
		valueB, exists := b[key]
		if !exists {
			return false
		}

		// Handle different numeric types that JSON might produce
		if !valuesEqual(valueA, valueB) {
			return false
		}
	}

	return true
}

// Helper function to compare values with type flexibility
func valuesEqual(a, b any) bool {
	if a == b {
		return true
	}

	// Handle numeric conversions
	aFloat, aIsFloat := a.(float64)
	bFloat, bIsFloat := b.(float64)
	if aIsFloat && bIsFloat {
		return aFloat == bFloat
	}

	aInt, aIsInt := a.(int)
	bInt, bIsInt := b.(int)
	if aIsInt && bIsInt {
		return aInt == bInt
	}

	// Cross-type numeric comparison
	if aIsFloat && bIsInt {
		return aFloat == float64(bInt)
	}
	if aIsInt && bIsFloat {
		return float64(aInt) == bFloat
	}

	return false
}
