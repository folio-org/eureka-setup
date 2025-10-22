package httpclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/folio-org/eureka-cli/action"
)

// Test retry-enabled GET method
func TestGetRetryDecodeReturnAny(t *testing.T) {
	tests := []struct {
		name           string
		responseBody   string
		statusCode     int
		expectedResult any
		expectError    bool
	}{
		{
			name:         "valid JSON response with retry",
			responseBody: `{"message": "success", "id": 123}`,
			statusCode:   http.StatusOK,
			expectedResult: map[string]any{
				"message": "success",
				"id":      float64(123),
			},
			expectError: false,
		},
		{
			name:           "empty response with retry",
			responseBody:   ``,
			statusCode:     http.StatusOK,
			expectedResult: nil,
			expectError:    false,
		},
		{
			name:         "invalid JSON with retry",
			responseBody: `{"invalid": json}`,
			statusCode:   http.StatusOK,
			expectError:  true,
		},
		{
			name:         "array response with retry",
			responseBody: `[1, 2, 3]`,
			statusCode:   http.StatusOK,
			expectedResult: []any{
				float64(1), float64(2), float64(3),
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			action := &action.Action{Name: "test-get-retry"}
			client := New(action)

			result, err := client.GetRetryDecodeReturnAny(server.URL, map[string]string{})

			if tt.expectError {
				if err == nil {
					t.Error("GetRetryDecodeReturnAny() expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("GetRetryDecodeReturnAny() unexpected error = %v", err)
			}

			if !anyEqual(result, tt.expectedResult) {
				t.Errorf("GetRetryDecodeReturnAny() = %v, want %v", result, tt.expectedResult)
			}
		})
	}
}

// Test retry-enabled POST method
func TestPostRetryReturnNoContent(t *testing.T) {
	requestBody := []byte(`{"retry": "test"}`)

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

		w.WriteHeader(http.StatusAccepted)
	}))
	defer server.Close()

	action := &action.Action{Name: "test-post-retry"}
	client := New(action)

	err := client.PostRetryReturnNoContent(server.URL, requestBody, map[string]string{})
	if err != nil {
		t.Fatalf("PostRetryReturnNoContent() error = %v", err)
	}
}

// Test retry functionality with server errors
func TestRetryFunctionality(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			// Return 503 to trigger retry
			w.WriteHeader(http.StatusServiceUnavailable)
			_, _ = w.Write([]byte(`{"error": "service unavailable"}`))
		} else {
			// Success on third attempt
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"success": true}`))
		}
	}))
	defer server.Close()

	action := &action.Action{Name: "test-retry"}
	client := New(action)

	// Test retry on GET
	result, err := client.GetRetryDecodeReturnAny(server.URL, map[string]string{})
	if err != nil {
		t.Fatalf("GetRetryDecodeReturnAny() with retry failed: %v", err)
	}

	expected := map[string]any{"success": true}
	if !anyEqual(result, expected) {
		t.Errorf("GetRetryDecodeReturnAny() after retry = %v, want %v", result, expected)
	}

	if attempts < 3 {
		t.Errorf("Expected at least 3 attempts, got %d", attempts)
	}
}

// Test retry client creation and configuration
func TestCreateRetryClient(t *testing.T) {
	retryClient := createRetryClient()

	if retryClient == nil {
		t.Fatal("createRetryClient() returned nil")
	}

	if retryClient.HTTPClient == nil {
		t.Error("createRetryClient() did not set HTTPClient")
	}

	if retryClient.RetryMax == 0 {
		t.Error("createRetryClient() did not set RetryMax")
	}

	if retryClient.RetryWaitMin == 0 {
		t.Error("createRetryClient() did not set RetryWaitMin")
	}

	if retryClient.RetryWaitMax == 0 {
		t.Error("createRetryClient() did not set RetryWaitMax")
	}

	if retryClient.CheckRetry == nil {
		t.Error("createRetryClient() did not set CheckRetry")
	}
}

// Test retry policy for different status codes
func TestRetryPolicy(t *testing.T) {
	retryClient := createRetryClient()

	tests := []struct {
		name        string
		statusCode  int
		shouldRetry bool
	}{
		{"429 Too Many Requests", http.StatusTooManyRequests, true},
		{"503 Service Unavailable", http.StatusServiceUnavailable, true},
		{"404 Not Found", http.StatusNotFound, true},
		{"200 OK", http.StatusOK, false},
		{"400 Bad Request", http.StatusBadRequest, false},
		{"500 Internal Server Error", http.StatusInternalServerError, true}, // 500 should retry based on default policy
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock response
			resp := &http.Response{
				StatusCode: tt.statusCode,
				Header:     make(http.Header),
			}

			shouldRetry, _ := retryClient.CheckRetry(context.Background(), resp, nil)
			if shouldRetry != tt.shouldRetry {
				t.Errorf("CheckRetry() for status %d = %v, want %v", tt.statusCode, shouldRetry, tt.shouldRetry)
			}
		})
	}
}
