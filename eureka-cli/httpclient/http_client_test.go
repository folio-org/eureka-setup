package httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
)

func TestNew(t *testing.T) {
	action := &action.Action{Name: "test-action"}
	client := New(action)

	if client == nil {
		t.Fatal("New() returned nil client")
	}

	if client.Action != action {
		t.Error("New() did not set action correctly")
	}

	if client.customClient == nil {
		t.Error("New() did not initialize custom client")
	}

	if client.retryClient == nil {
		t.Error("New() did not initialize retry client")
	}
}

func TestSetRequestHeaders(t *testing.T) {
	tests := []struct {
		name            string
		headers         map[string]string
		expectedHeaders map[string]string
	}{
		{
			name:    "empty headers - should set default JSON content type",
			headers: map[string]string{},
			expectedHeaders: map[string]string{
				constant.ContentTypeHeader: constant.ApplicationJSON,
			},
		},
		{
			name:    "nil headers - should set default JSON content type",
			headers: nil,
			expectedHeaders: map[string]string{
				constant.ContentTypeHeader: constant.ApplicationJSON,
			},
		},
		{
			name: "custom headers - should set provided headers",
			headers: map[string]string{
				"Authorization":            "Bearer token123",
				"X-Custom-Header":          "custom-value",
				constant.ContentTypeHeader: "application/xml",
			},
			expectedHeaders: map[string]string{
				"Authorization":            "Bearer token123",
				"X-Custom-Header":          "custom-value",
				constant.ContentTypeHeader: "application/xml",
			},
		},
		{
			name: "single custom header",
			headers: map[string]string{
				"User-Agent": "eureka-cli/1.0",
			},
			expectedHeaders: map[string]string{
				"User-Agent": "eureka-cli/1.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("GET", "http://example.com", nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			SetRequestHeaders(req, tt.headers)

			for expectedKey, expectedValue := range tt.expectedHeaders {
				actualValue := req.Header.Get(expectedKey)
				if actualValue != expectedValue {
					t.Errorf("Header %s = %q, want %q", expectedKey, actualValue, expectedValue)
				}
			}
		})
	}
}

func TestValidateResponse(t *testing.T) {
	action := &action.Action{Name: "test-action"}
	client := New(action)

	tests := []struct {
		name          string
		statusCode    int
		expectError   bool
		errorContains string
	}{
		{
			name:        "200 OK - should pass",
			statusCode:  http.StatusOK,
			expectError: false,
		},
		{
			name:        "201 Created - should pass",
			statusCode:  http.StatusCreated,
			expectError: false,
		},
		{
			name:        "204 No Content - should pass",
			statusCode:  http.StatusNoContent,
			expectError: false,
		},
		{
			name:        "299 edge case - should pass",
			statusCode:  299,
			expectError: false,
		},
		{
			name:          "400 Bad Request - should fail",
			statusCode:    http.StatusBadRequest,
			expectError:   true,
			errorContains: "unacceptable request status 400",
		},
		{
			name:          "404 Not Found - should fail",
			statusCode:    http.StatusNotFound,
			expectError:   true,
			errorContains: "unacceptable request status 404",
		},
		{
			name:          "500 Internal Server Error - should fail",
			statusCode:    http.StatusInternalServerError,
			expectError:   true,
			errorContains: "unacceptable request status 500",
		},
		{
			name:          "300 Multiple Choices - should fail",
			statusCode:    http.StatusMultipleChoices,
			expectError:   true,
			errorContains: "unacceptable request status 300",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server with the desired status code
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte("test response"))
			}))
			defer server.Close()

			// Create a request to the test server
			req, err := http.NewRequest("GET", server.URL, nil)
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}

			// Execute the request to get a response
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Fatalf("Failed to execute request: %v", err)
			}
			defer resp.Body.Close()

			// Test ValidateResponse
			err = client.ValidateResponse(resp)

			if tt.expectError {
				if err == nil {
					t.Errorf("ValidateResponse() expected error for status %d, but got none", tt.statusCode)
				} else if tt.errorContains != "" && !containsString(err.Error(), tt.errorContains) {
					t.Errorf("ValidateResponse() error = %q, want to contain %q", err.Error(), tt.errorContains)
				}
			} else {
				if err != nil {
					t.Errorf("ValidateResponse() unexpected error for status %d: %v", tt.statusCode, err)
				}
			}
		})
	}
}

func TestCloseResponse(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	}))
	defer server.Close()

	// Create and execute a request
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("Failed to execute request: %v", err)
	}

	// Test CloseResponse - this should not panic
	CloseResponse(resp)

	// Verify the body is closed by trying to read from it
	buf := make([]byte, 1)
	_, err = resp.Body.Read(buf)
	if err == nil {
		t.Error("CloseResponse() did not close the response body")
	}
}

// Helper function to check if a string contains a substring
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr ||
		(len(s) > len(substr) &&
			(s[:len(substr)] == substr ||
				s[len(s)-len(substr):] == substr ||
				indexOf(s, substr) >= 0)))
}

// Simple indexOf implementation
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
