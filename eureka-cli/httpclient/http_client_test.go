package httpclient_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/folio-org/eureka-setup/eureka-cli/httpclient"
	"github.com/stretchr/testify/assert"
)

type TestResponse struct {
	ID      int    `json:"id"`
	Message string `json:"message"`
}

func createTestAction() *action.Action {
	return &action.Action{Name: "TestAction"}
}

func createTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// GET Tests

func TestGetReturnStruct_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(TestResponse{ID: 42, Message: "test"})
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	var result TestResponse

	// Act
	err := client.GetReturnStruct(server.URL, nil, &result)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 42, result.ID)
	assert.Equal(t, "test", result.Message)
}

func TestGetReturnStruct_EmptyResponse(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	var result TestResponse

	// Act
	err := client.GetReturnStruct(server.URL, nil, &result)

	// Assert
	assert.NoError(t, err)
}

func TestGetReturnStruct_InvalidJSON(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	var result TestResponse

	// Act
	err := client.GetReturnStruct(server.URL, nil, &result)

	// Assert
	assert.Error(t, err)
}

func TestGetRetryReturnStruct_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(TestResponse{ID: 100, Message: "retry success"})
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	var result TestResponse

	// Act
	err := client.GetRetryReturnStruct(server.URL, nil, &result)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 100, result.ID)
	assert.Equal(t, "retry success", result.Message)
}

// POST Tests

func TestPostReturnNoContent_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "test")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	payload := []byte(`{"test": "data"}`)

	// Act
	err := client.PostReturnNoContent(server.URL, payload, nil)

	// Assert
	assert.NoError(t, err)
}

func TestPostReturnNoContent_WithNilPayload(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())

	// Act
	err := client.PostReturnNoContent(server.URL, nil, nil)

	// Assert
	assert.NoError(t, err)
}

func TestPostRetryReturnNoContent_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	payload := []byte(`{"retry": "test"}`)

	// Act
	err := client.PostRetryReturnNoContent(server.URL, payload, nil)

	// Assert
	assert.NoError(t, err)
}

func TestPostReturnStruct_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(TestResponse{ID: 200, Message: "created"})
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	payload := []byte(`{"name": "test"}`)
	var result TestResponse

	// Act
	err := client.PostReturnStruct(server.URL, payload, nil, &result)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 200, result.ID)
	assert.Equal(t, "created", result.Message)
}

func TestPostReturnStruct_EmptyResponse(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusCreated)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	var result TestResponse

	// Act
	err := client.PostReturnStruct(server.URL, []byte(`{}`), nil, &result)

	// Assert
	assert.NoError(t, err)
}

func TestPostReturnStruct_InvalidJSON(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	var result TestResponse

	// Act
	err := client.PostReturnStruct(server.URL, []byte(`{}`), nil, &result)

	// Assert
	assert.Error(t, err)
}

func TestPostReturnStruct_EOFResponse(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Empty body causes EOF
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	var result TestResponse

	// Act
	err := client.PostReturnStruct(server.URL, []byte(`{}`), nil, &result)

	// Assert
	assert.NoError(t, err) // EOF is ignored
}

func TestPostFormDataReturnStruct_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		// Read the body to get form data
		body, _ := io.ReadAll(r.Body)
		formStr := string(body)
		assert.Contains(t, formStr, "username=testuser")
		assert.Contains(t, formStr, "password=testpass")

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(TestResponse{ID: 1, Message: "logged in"})
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	formData := url.Values{}
	formData.Set("username", "testuser")
	formData.Set("password", "testpass")
	var result TestResponse

	// Act
	err := client.PostFormDataReturnStruct(server.URL, formData, nil, &result)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 1, result.ID)
	assert.Equal(t, "logged in", result.Message)
}

func TestPostFormDataReturnStruct_WithCustomHeaders(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/x-www-form-urlencoded", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(TestResponse{ID: 2, Message: "ok"})
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	formData := url.Values{}
	formData.Set("key", "value")
	headers := map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	}
	var result TestResponse

	// Act
	err := client.PostFormDataReturnStruct(server.URL, formData, headers, &result)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, 2, result.ID)
}

func TestPostFormDataReturnStruct_EmptyResponse(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "0")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	formData := url.Values{}
	var result TestResponse

	// Act
	err := client.PostFormDataReturnStruct(server.URL, formData, nil, &result)

	// Assert
	assert.NoError(t, err)
}

func TestPostFormDataReturnStruct_InvalidJSON(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{bad json`))
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	formData := url.Values{}
	var result TestResponse

	// Act
	err := client.PostFormDataReturnStruct(server.URL, formData, nil, &result)

	// Assert
	assert.Error(t, err)
}

func TestPostFormDataReturnStruct_ValidationError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	formData := url.Values{}
	var result TestResponse

	// Act
	err := client.PostFormDataReturnStruct(server.URL, formData, nil, &result)

	// Assert
	assert.Error(t, err)
}

func TestPostFormDataReturnStruct_EOFResponse(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Empty body causes EOF
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	formData := url.Values{}
	var result TestResponse

	// Act
	err := client.PostFormDataReturnStruct(server.URL, formData, nil, &result)

	// Assert
	assert.NoError(t, err) // EOF is ignored
}

// PUT Tests

func TestPutReturnNoContent_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPut, r.Method)
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "updated")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	payload := []byte(`{"updated": "data"}`)

	// Act
	err := client.PutReturnNoContent(server.URL, payload, nil)

	// Assert
	assert.NoError(t, err)
}

func TestPutReturnNoContent_WithHeaders(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	headers := map[string]string{"Content-Type": "application/json"}
	payload := []byte(`{"id": 1}`)

	// Act
	err := client.PutReturnNoContent(server.URL, payload, headers)

	// Assert
	assert.NoError(t, err)
}

// DELETE Tests

func TestDelete_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())

	// Act
	err := client.Delete(server.URL, nil)

	// Assert
	assert.NoError(t, err)
}

func TestDelete_WithHeaders(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer token", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	headers := map[string]string{"Authorization": "Bearer token"}

	// Act
	err := client.Delete(server.URL, headers)

	// Assert
	assert.NoError(t, err)
}

func TestDeleteWithPayloadReturnStruct_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodDelete, r.Method)
		body, _ := io.ReadAll(r.Body)
		assert.Contains(t, string(body), "reason")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status": "deleted"}`))
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	payload := []byte(`{"reason": "test deletion"}`)
	var response map[string]any

	// Act
	err := client.DeleteWithPayloadReturnStruct(server.URL, payload, nil, &response)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, "deleted", response["status"])
}

// Ping Tests

func TestPingRetry_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())

	// Act
	err := client.PingRetry(server.URL)

	// Assert
	assert.NoError(t, err)
}

func TestPingRetry_Failure(t *testing.T) {
	// Arrange
	client := httpclient.New(createTestAction(), createTestLogger())

	// Act
	err := client.PingRetry("http://localhost:99999/nonexistent")

	// Assert
	assert.Error(t, err)
}

func TestPing_Success(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())

	// Act
	statusCode, err := client.Ping(server.URL)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, statusCode)
}

func TestPing_NonOKStatus(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())

	// Act
	statusCode, err := client.Ping(server.URL)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, statusCode)
}

func TestPing_NetworkError(t *testing.T) {
	// Arrange
	client := httpclient.New(createTestAction(), createTestLogger())

	// Act
	statusCode, err := client.Ping("http://localhost:99999/nonexistent")

	// Assert
	assert.Error(t, err)
	assert.Equal(t, 0, statusCode)
}

// CloseResponse Tests

func TestCloseResponse_WithValidResponse(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("test data"))
	}))
	defer server.Close()

	resp, _ := http.Get(server.URL)

	// Act & Assert - Should not panic
	httpclient.CloseResponse(resp)
}

func TestCloseResponse_WithNilResponse(t *testing.T) {
	// Act & Assert - Should not panic
	httpclient.CloseResponse(nil)
}

func TestCloseResponse_WithNilBody(t *testing.T) {
	// Arrange
	response := &http.Response{
		StatusCode: 200,
		Body:       nil,
	}

	// Act & Assert - Should not panic
	httpclient.CloseResponse(response)
}

// Constructor Test

func TestNew_CreatesClient(t *testing.T) {
	// Arrange
	action := createTestAction()
	logger := createTestLogger()

	// Act
	client := httpclient.New(action, logger)

	// Assert
	assert.NotNil(t, client)
}

// Edge Cases and Error Handling

func TestGetReturnStruct_ServerError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error": "internal error"}`))
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	var result TestResponse

	// Act
	err := client.GetReturnStruct(server.URL, nil, &result)

	// Assert
	assert.Error(t, err)
}

func TestPostReturnNoContent_ServerError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())

	// Act
	err := client.PostReturnNoContent(server.URL, []byte(`{}`), nil)

	// Assert
	assert.Error(t, err)
}

func TestDelete_ServerError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())

	// Act
	err := client.Delete(server.URL, nil)

	// Assert
	assert.Error(t, err)
}

func TestPutReturnNoContent_ServerError(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())

	// Act
	err := client.PutReturnNoContent(server.URL, []byte(`{}`), nil)

	// Assert
	assert.Error(t, err)
}

func TestGetReturnStruct_EOFHandling(t *testing.T) {
	// Arrange
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		// Write nothing, causing EOF on decode
	}))
	defer server.Close()

	client := httpclient.New(createTestAction(), createTestLogger())
	var result TestResponse

	// Act
	err := client.GetReturnStruct(server.URL, nil, &result)

	// Assert
	assert.NoError(t, err) // EOF is handled gracefully
}
