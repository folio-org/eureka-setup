package httpclient

import (
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/folio-org/eureka-cli/action"
)

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
