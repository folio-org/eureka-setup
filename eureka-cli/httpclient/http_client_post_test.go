package httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/folio-org/eureka-cli/action"
)

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
		_, _ = w.Write([]byte(responseBody))
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
