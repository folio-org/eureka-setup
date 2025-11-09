package testhelpers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockHTTPServer creates a test HTTP server with predefined responses
type MockHTTPServer struct {
	Server   *httptest.Server
	Requests []*http.Request
	t        *testing.T
}

// NewMockHTTPServer creates a new mock HTTP server
func NewMockHTTPServer(t *testing.T, handler http.HandlerFunc) *MockHTTPServer {
	t.Helper()

	mock := &MockHTTPServer{
		Requests: []*http.Request{},
		t:        t,
	}

	// Wrap handler to capture requests
	wrappedHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Capture request for assertion
		mock.Requests = append(mock.Requests, r)
		handler(w, r)
	})

	mock.Server = httptest.NewServer(wrappedHandler)
	return mock
}

// Close closes the mock HTTP server
func (m *MockHTTPServer) Close() {
	m.Server.Close()
}

// URL returns the base URL of the mock server
func (m *MockHTTPServer) URL() string {
	return m.Server.URL
}

// AssertRequestCount asserts the number of requests received
func (m *MockHTTPServer) AssertRequestCount(expected int) {
	m.t.Helper()
	assert.Equal(m.t, expected, len(m.Requests), "unexpected number of requests")
}

// AssertRequestMethod asserts the HTTP method of a specific request
func (m *MockHTTPServer) AssertRequestMethod(index int, expectedMethod string) {
	m.t.Helper()
	require.Less(m.t, index, len(m.Requests), "request index out of bounds")
	assert.Equal(m.t, expectedMethod, m.Requests[index].Method, "unexpected request method")
}

// AssertRequestPath asserts the path of a specific request
func (m *MockHTTPServer) AssertRequestPath(index int, expectedPath string) {
	m.t.Helper()
	require.Less(m.t, index, len(m.Requests), "request index out of bounds")
	assert.Equal(m.t, expectedPath, m.Requests[index].URL.Path, "unexpected request path")
}

// AssertRequestHeader asserts a header value of a specific request
func (m *MockHTTPServer) AssertRequestHeader(index int, headerName, expectedValue string) {
	m.t.Helper()
	require.Less(m.t, index, len(m.Requests), "request index out of bounds")
	assert.Equal(m.t, expectedValue, m.Requests[index].Header.Get(headerName), "unexpected header value")
}

// GetRequestBody returns the body of a specific request as bytes
func (m *MockHTTPServer) GetRequestBody(index int) []byte {
	m.t.Helper()
	require.Less(m.t, index, len(m.Requests), "request index out of bounds")
	body, err := io.ReadAll(m.Requests[index].Body)
	require.NoError(m.t, err, "failed to read request body")
	return body
}

// AssertRequestJSONBody asserts the JSON body of a specific request
func (m *MockHTTPServer) AssertRequestJSONBody(index int, expected any) {
	m.t.Helper()
	body := m.GetRequestBody(index)

	var actual any
	err := json.Unmarshal(body, &actual)
	require.NoError(m.t, err, "failed to unmarshal request body")

	expectedJSON, err := json.Marshal(expected)
	require.NoError(m.t, err, "failed to marshal expected value")

	var expectedMap any
	err = json.Unmarshal(expectedJSON, &expectedMap)
	require.NoError(m.t, err, "failed to unmarshal expected JSON")

	assert.Equal(m.t, expectedMap, actual, "unexpected request body")
}

// JSONResponse creates an HTTP handler that returns a JSON response
func JSONResponse(statusCode int, body any) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)

		if body != nil {
			if err := json.NewEncoder(w).Encode(body); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

// ErrorResponse creates an HTTP handler that returns an error response
func ErrorResponse(statusCode int, message string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(statusCode)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": message})
	}
}

// EmptyResponse creates an HTTP handler that returns an empty response
func EmptyResponse(statusCode int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(statusCode)
	}
}

// SequentialResponses creates an HTTP handler that returns different responses in sequence
func SequentialResponses(responses ...http.HandlerFunc) http.HandlerFunc {
	index := 0
	return func(w http.ResponseWriter, r *http.Request) {
		if index >= len(responses) {
			http.Error(w, "no more responses configured", http.StatusInternalServerError)
			return
		}
		handler := responses[index]
		index++
		handler(w, r)
	}
}

// ReadJSONBody reads and unmarshals JSON from an io.Reader
func ReadJSONBody(t *testing.T, body io.Reader, target any) {
	t.Helper()
	data, err := io.ReadAll(body)
	require.NoError(t, err, "failed to read body")

	err = json.Unmarshal(data, target)
	require.NoError(t, err, "failed to unmarshal JSON")
}

// CreateJSONReader creates an io.Reader from a JSON-serializable object
func CreateJSONReader(t *testing.T, obj any) io.Reader {
	t.Helper()
	data, err := json.Marshal(obj)
	require.NoError(t, err, "failed to marshal JSON")
	return bytes.NewReader(data)
}

// AssertNoError is a helper to assert no error occurred
func AssertNoError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	assert.NoError(t, err, msgAndArgs...)
}

// AssertError is a helper to assert an error occurred
func AssertError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	assert.Error(t, err, msgAndArgs...)
}

// AssertEqual is a helper for equality assertions
func AssertEqual(t *testing.T, expected, actual any, msgAndArgs ...any) {
	t.Helper()
	assert.Equal(t, expected, actual, msgAndArgs...)
}

// AssertNotNil is a helper to assert not nil
func AssertNotNil(t *testing.T, obj any, msgAndArgs ...any) {
	t.Helper()
	assert.NotNil(t, obj, msgAndArgs...)
}

// AssertNil is a helper to assert nil
func AssertNil(t *testing.T, obj any, msgAndArgs ...any) {
	t.Helper()
	assert.Nil(t, obj, msgAndArgs...)
}

// RequireNoError is a helper that fails the test immediately if there's an error
func RequireNoError(t *testing.T, err error, msgAndArgs ...any) {
	t.Helper()
	require.NoError(t, err, msgAndArgs...)
}
