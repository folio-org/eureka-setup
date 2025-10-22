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
		_, _ = w.Write([]byte(`{"message": "success"}`))
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
				_, _ = w.Write([]byte(tt.responseBody))
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
