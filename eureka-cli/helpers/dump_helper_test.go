package helpers

import (
	"bytes"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/folio-org/eureka-cli/action"
)

func TestDumpRequestJSON(t *testing.T) {
	tests := []struct {
		name      string
		bodyBytes []byte
	}{
		{
			name:      "valid JSON body",
			bodyBytes: []byte(`{"name": "test", "value": 123}`),
		},
		{
			name:      "empty body",
			bodyBytes: []byte{},
		},
		{
			name:      "invalid JSON body",
			bodyBytes: []byte(`{"name": "test", "value":}`),
		},
		{
			name:      "large body",
			bodyBytes: []byte(strings.Repeat("a", 1000)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// This should not panic regardless of debug level
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("DumpRequestJSON panicked: %v", r)
				}

				// Restore stdout
				_ = w.Close()
				os.Stdout = old

				// Read captured output
				var output bytes.Buffer
				_, _ = io.Copy(&output, r)
			}()

			DumpRequestJSON(tt.bodyBytes)
		})
	}
}

func TestDumpRequestFormData(t *testing.T) {
	tests := []struct {
		name     string
		formData url.Values
	}{
		{
			name: "basic form data",
			formData: url.Values{
				"username": []string{"testuser"},
				"password": []string{"testpass"},
			},
		},
		{
			name:     "empty form data",
			formData: url.Values{},
		},
		{
			name: "multiple values for same key",
			formData: url.Values{
				"tags": []string{"tag1", "tag2", "tag3"},
			},
		},
		{
			name: "special characters in form data",
			formData: url.Values{
				"special": []string{"value with spaces & symbols!"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			// This should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("DumpRequestFormData panicked: %v", r)
				}

				// Restore stdout
				_ = w.Close()
				os.Stdout = old

				// Read captured output
				var output bytes.Buffer
				_, _ = io.Copy(&output, r)
			}()

			DumpRequestFormData(tt.formData)
		})
	}
}

func TestDumpRequest(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		url         string
		body        string
		headers     map[string]string
		expectError bool
	}{
		{
			name:        "GET request",
			method:      "GET",
			url:         "http://example.com/api",
			body:        "",
			headers:     map[string]string{"Accept": "application/json"},
			expectError: false,
		},
		{
			name:        "POST request with body",
			method:      "POST",
			url:         "http://example.com/api",
			body:        `{"name": "test"}`,
			headers:     map[string]string{"Content-Type": "application/json"},
			expectError: false,
		},
		{
			name:   "request with multiple headers",
			method: "PUT",
			url:    "http://example.com/api/123",
			body:   "update data",
			headers: map[string]string{
				"Authorization":   "Bearer token123",
				"Content-Type":    "text/plain",
				"X-Custom-Header": "custom-value",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			var bodyReader io.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			}

			req := httptest.NewRequest(tt.method, tt.url, bodyReader)

			// Add headers
			for key, value := range tt.headers {
				req.Header.Set(key, value)
			}

			// Create action
			testAction := &action.Action{Name: "test-action"}

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			var err error
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("DumpRequest panicked: %v", r)
				}

				// Restore stdout
				_ = w.Close()
				os.Stdout = old

				// Read captured output
				var output bytes.Buffer
				_, _ = io.Copy(&output, r)
			}()

			err = DumpRequest(testAction, req)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDumpResponse(t *testing.T) {
	tests := []struct {
		name        string
		statusCode  int
		body        string
		headers     map[string]string
		forceDump   bool
		expectError bool
	}{
		{
			name:        "successful response",
			statusCode:  200,
			body:        `{"status": "success"}`,
			headers:     map[string]string{"Content-Type": "application/json"},
			forceDump:   false,
			expectError: false,
		},
		{
			name:        "error response",
			statusCode:  500,
			body:        `{"error": "internal server error"}`,
			headers:     map[string]string{"Content-Type": "application/json"},
			forceDump:   false,
			expectError: false,
		},
		{
			name:        "force dump response",
			statusCode:  404,
			body:        "Not Found",
			headers:     map[string]string{"Content-Type": "text/plain"},
			forceDump:   true,
			expectError: false,
		},
		{
			name:        "empty response body",
			statusCode:  204,
			body:        "",
			headers:     map[string]string{},
			forceDump:   false,
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Set headers
				for key, value := range tt.headers {
					w.Header().Set(key, value)
				}

				w.WriteHeader(tt.statusCode)
				_, _ = w.Write([]byte(tt.body))
			}))
			defer server.Close()

			// Make request to get response
			resp, err := http.Get(server.URL)
			if err != nil {
				t.Fatalf("Failed to make request: %v", err)
			}
			defer func() {
				_ = resp.Body.Close()
			}()

			// Create action
			testAction := &action.Action{Name: "test-action"}

			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("DumpResponse panicked: %v", r)
				}

				// Restore stdout
				_ = w.Close()
				os.Stdout = old

				// Read captured output
				var output bytes.Buffer
				_, _ = io.Copy(&output, r)
			}()

			err = DumpResponse(testAction, resp, tt.forceDump)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func TestDumpFunctionsWithDebugLevel(t *testing.T) {
	// Test that dump functions respect debug log level
	// Save original log level
	originalLevel := slog.SetLogLoggerLevel(slog.LevelDebug)
	defer slog.SetLogLoggerLevel(originalLevel)

	// Test with debug level enabled
	t.Run("debug level enabled", func(t *testing.T) {
		// These should not panic when debug is enabled
		DumpRequestJSON([]byte(`{"test": "data"}`))

		formData := url.Values{"key": []string{"value"}}
		DumpRequestFormData(formData)

		req := httptest.NewRequest("GET", "http://example.com", nil)
		action := &action.Action{Name: "test"}
		err := DumpRequest(action, req)
		if err != nil {
			t.Errorf("DumpRequest failed: %v", err)
		}
	})

	// Test with debug level disabled
	t.Run("debug level disabled", func(t *testing.T) {
		slog.SetLogLoggerLevel(slog.LevelInfo)

		// These should still not panic when debug is disabled
		DumpRequestJSON([]byte(`{"test": "data"}`))

		formData := url.Values{"key": []string{"value"}}
		DumpRequestFormData(formData)

		req := httptest.NewRequest("GET", "http://example.com", nil)
		action := &action.Action{Name: "test"}
		err := DumpRequest(action, req)
		if err != nil {
			t.Errorf("DumpRequest failed: %v", err)
		}
	})
}
