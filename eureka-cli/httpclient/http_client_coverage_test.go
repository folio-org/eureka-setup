package httpclient

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
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

// Test form data POST method
func TestPostFormDataReturnMapStringAny(t *testing.T) {
	tests := []struct {
		name         string
		formData     url.Values
		responseBody string
		statusCode   int
		expected     map[string]any
		expectError  bool
	}{
		{
			name: "valid form data",
			formData: url.Values{
				"username": []string{"testuser"},
				"password": []string{"testpass"},
				"remember": []string{"true"},
			},
			responseBody: `{"token": "abc123", "expires": 3600}`,
			statusCode:   http.StatusOK,
			expected: map[string]any{
				"token":   "abc123",
				"expires": float64(3600),
			},
			expectError: false,
		},
		{
			name:         "empty form data",
			formData:     url.Values{},
			responseBody: `{"status": "empty"}`,
			statusCode:   http.StatusOK,
			expected: map[string]any{
				"status": "empty",
			},
			expectError: false,
		},
		{
			name: "error response",
			formData: url.Values{
				"invalid": []string{"data"},
			},
			responseBody: `{"error": "bad request"}`,
			statusCode:   http.StatusBadRequest,
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("Expected POST request, got %s", r.Method)
				}

				// Check content type for form data
				contentType := r.Header.Get("Content-Type")
				if !strings.Contains(contentType, "application/x-www-form-urlencoded") &&
					!strings.Contains(contentType, "application/json") {
					t.Logf("Unexpected content type: %s", contentType)
				}

				// Parse form data from body (since it's sent as body, not query params)
				body := make([]byte, 1024)
				n, _ := r.Body.Read(body)
				formData := string(body[:n])

				// Basic verification that form data is in the body
				for key := range tt.formData {
					if !strings.Contains(formData, key) {
						t.Errorf("Form data missing key: %s", key)
					}
				}

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			action := &action.Action{Name: "test-form-data"}
			client := New(action)

			result, err := client.PostFormDataReturnMapStringAny(server.URL, tt.formData, map[string]string{})

			if tt.expectError {
				if err == nil {
					t.Error("PostFormDataReturnMapStringAny() expected error, but got none")
				}
				return
			}

			if err != nil {
				t.Fatalf("PostFormDataReturnMapStringAny() unexpected error = %v", err)
			}

			if !mapsEqual(result, tt.expected) {
				t.Errorf("PostFormDataReturnMapStringAny() = %v, want %v", result, tt.expected)
			}
		})
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

// Test logging round tripper
func TestLoggingRoundTripper(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"logged": true}`))
	}))
	defer server.Close()

	// Create a logger
	logger := slog.Default()

	// Create logging round tripper
	roundTripper := &LoggingRoundTripper{
		logger: logger,
		next:   http.DefaultTransport,
	}

	// Create request
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Test successful round trip
	resp, err := roundTripper.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip() failed: %v", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("RoundTrip() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}
}

// Test logging round tripper with error
func TestLoggingRoundTripperError(t *testing.T) {
	logger := slog.Default()

	// Create logging round tripper with a failing transport
	roundTripper := &LoggingRoundTripper{
		logger: logger,
		next:   &failingTransport{},
	}

	// Create request to non-existent server
	req, err := http.NewRequest("GET", "http://invalid-host-that-does-not-exist.local", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Test error round trip
	_, err = roundTripper.RoundTrip(req)
	if err == nil {
		t.Error("RoundTrip() expected error for invalid host, but got none")
	}
}

// Test custom client creation
func TestCreateCustomClient(t *testing.T) {
	client := createCustomClient()

	if client == nil {
		t.Fatal("createCustomClient() returned nil")
	}

	if client.Timeout == 0 {
		t.Error("createCustomClient() did not set timeout")
	}

	if client.Transport == nil {
		t.Error("createCustomClient() did not set transport")
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

// Test edge cases for doRequest method
func TestDoRequestEdgeCases(t *testing.T) {
	action := &action.Action{Name: "test-edge-cases"}
	client := New(action)

	// Test with nil body
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"nil_body": true}`))
	}))
	defer server.Close()

	resp, err := client.GetReturnResponse(server.URL, nil)
	if err != nil {
		t.Fatalf("doRequest() with nil body failed: %v", err)
	}
	defer CloseResponse(resp)

	// Test with both retry and non-retry paths
	_, err = client.GetRetryDecodeReturnAny(server.URL, nil)
	if err != nil {
		t.Fatalf("doRequest() with retry failed: %v", err)
	}
}

// Helper types and functions

// failingTransport simulates a transport that always fails
type failingTransport struct{}

func (ft *failingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, &mockNetError{}
}

// mockNetError simulates a network error
type mockNetError struct{}

func (e *mockNetError) Error() string   { return "mock network error" }
func (e *mockNetError) Timeout() bool   { return false }
func (e *mockNetError) Temporary() bool { return false }

// Helper function to compare any values including slices
func anyEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	// Handle slice comparisons
	sliceA, isSliceA := a.([]any)
	sliceB, isSliceB := b.([]any)
	if isSliceA && isSliceB {
		if len(sliceA) != len(sliceB) {
			return false
		}
		for i := range sliceA {
			if !anyEqual(sliceA[i], sliceB[i]) {
				return false
			}
		}
		return true
	}

	// Handle map comparisons
	mapA, isMapA := a.(map[string]any)
	mapB, isMapB := b.(map[string]any)
	if isMapA && isMapB {
		return mapsEqual(mapA, mapB)
	}

	// Handle numeric type conversions
	return valuesEqual(a, b)
}
