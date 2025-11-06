package testhelpers

import (
	"net/http"
	"net/url"

	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/actionparams"
	"github.com/stretchr/testify/mock"
)

// MockHTTPClient is a mock implementation of httpclient.HTTPClientRunner
type MockHTTPClient struct {
	mock.Mock
}

func (m *MockHTTPClient) Ping(url string) error {
	args := m.Called(url)
	return args.Error(0)
}

func (m *MockHTTPClient) GetReturnResponse(url string, headers map[string]string) (*http.Response, error) {
	args := m.Called(url, headers)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*http.Response), args.Error(1)
}

func (m *MockHTTPClient) GetReturnStruct(url string, headers map[string]string, target any) error {
	args := m.Called(url, headers, target)
	return args.Error(0)
}

func (m *MockHTTPClient) GetRetryReturnStruct(url string, headers map[string]string, target any) error {
	args := m.Called(url, headers, target)
	return args.Error(0)
}

func (m *MockHTTPClient) PostReturnNoContent(url string, payload []byte, headers map[string]string) error {
	args := m.Called(url, payload, headers)
	return args.Error(0)
}

func (m *MockHTTPClient) PostRetryReturnNoContent(url string, payload []byte, headers map[string]string) error {
	args := m.Called(url, payload, headers)
	return args.Error(0)
}

func (m *MockHTTPClient) PostReturnStruct(url string, payload []byte, headers map[string]string, target any) error {
	args := m.Called(url, payload, headers, target)
	return args.Error(0)
}

func (m *MockHTTPClient) PostFormDataReturnStruct(urlStr string, formValues url.Values, headers map[string]string, target any) error {
	args := m.Called(urlStr, formValues, headers, target)
	return args.Error(0)
}

func (m *MockHTTPClient) PutReturnNoContent(url string, payload []byte, headers map[string]string) error {
	args := m.Called(url, payload, headers)
	return args.Error(0)
}

func (m *MockHTTPClient) DeleteReturnNoContent(url string, headers map[string]string) error {
	args := m.Called(url, headers)
	return args.Error(0)
}

func (m *MockHTTPClient) DeleteRetryReturnNoContent(url string, headers map[string]string) error {
	args := m.Called(url, headers)
	return args.Error(0)
}

func (m *MockHTTPClient) Delete(url string, headers map[string]string) error {
	args := m.Called(url, headers)
	return args.Error(0)
}

func (m *MockHTTPClient) DeleteWithBody(url string, payload []byte, headers map[string]string) error {
	args := m.Called(url, payload, headers)
	return args.Error(0)
}

// NewMockAction creates a minimal Action instance for testing
func NewMockAction() *action.Action {
	params := &actionparams.ActionParams{}
	return action.NewWithCredentials(
		"test-action",
		"http://localhost:%s", // Gateway URL template
		params,
		"test-vault-token",
		"test-keycloak-token",
		"test-master-token",
	)
}
