package httpclient

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/folio-org/eureka-cli/action"
)

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
