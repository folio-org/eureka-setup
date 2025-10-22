package httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/folio-org/eureka-cli/action"
)

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
