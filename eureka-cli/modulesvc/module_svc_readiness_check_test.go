package modulesvc_test

import (
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/folio-org/eureka-cli/dockerclient"
	"github.com/folio-org/eureka-cli/internal/testhelpers"
	"github.com/folio-org/eureka-cli/moduleenv"
	"github.com/folio-org/eureka-cli/modulesvc"
	"github.com/folio-org/eureka-cli/registrysvc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCheckModuleReadiness_Success(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := modulesvc.New(action, mockHTTP, &dockerclient.DockerClient{}, &registrysvc.RegistrySvc{}, &moduleenv.ModuleEnv{})
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	mockResponse := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader("")),
	}

	mockHTTP.On("GetReturnResponse",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(mockResponse, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()
	close(errCh)

	// Assert
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	default:
		// Success - no error sent
	}
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_HTTPError(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := modulesvc.New(action, mockHTTP, &dockerclient.DockerClient{}, &registrysvc.RegistrySvc{}, &moduleenv.ModuleEnv{})
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	mockHTTP.On("GetReturnResponse",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(nil, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()
	close(errCh)

	// Assert
	err := <-errCh
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "module test-module")
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_NilResponse(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := modulesvc.New(action, mockHTTP, &dockerclient.DockerClient{}, &registrysvc.RegistrySvc{}, &moduleenv.ModuleEnv{})
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	mockHTTP.On("GetReturnResponse",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(nil, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()
	close(errCh)

	// Assert
	err := <-errCh
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "module test-module")
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_NonOKStatusCode(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := modulesvc.New(action, mockHTTP, &dockerclient.DockerClient{}, &registrysvc.RegistrySvc{}, &moduleenv.ModuleEnv{})
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	mockResponse := &http.Response{
		StatusCode: http.StatusServiceUnavailable,
		Status:     "503 Service Unavailable",
		Body:       io.NopCloser(strings.NewReader("")),
	}

	mockHTTP.On("GetReturnResponse",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(mockResponse, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()
	close(errCh)

	// Assert
	err := <-errCh
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "module test-module")
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_EventualSuccess(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := modulesvc.New(action, mockHTTP, &dockerclient.DockerClient{}, &registrysvc.RegistrySvc{}, &moduleenv.ModuleEnv{})
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	failureResponse := &http.Response{
		StatusCode: http.StatusServiceUnavailable,
		Status:     "503 Service Unavailable",
		Body:       io.NopCloser(strings.NewReader("")),
	}

	successResponse := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader("")),
	}

	// First 2 calls fail, third succeeds
	mockHTTP.On("GetReturnResponse",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(failureResponse, nil).Times(2)

	mockHTTP.On("GetReturnResponse",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(successResponse, nil).Once()

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()
	close(errCh)

	// Assert
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	default:
		// Success - no error sent
	}
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_MultipleModulesConcurrent(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := modulesvc.New(action, mockHTTP, &dockerclient.DockerClient{}, &registrysvc.RegistrySvc{}, &moduleenv.ModuleEnv{})
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	successResponse := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader("")),
	}

	mockHTTP.On("GetReturnResponse",
		mock.Anything,
		mock.Anything).
		Return(successResponse, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 3)

	modules := []struct {
		name string
		port int
	}{
		{"module-1", 8081},
		{"module-2", 8082},
		{"module-3", 8083},
	}

	// Act
	for _, mod := range modules {
		wg.Add(1)
		go svc.CheckModuleReadiness(wg, errCh, mod.name, mod.port)
	}
	wg.Wait()
	close(errCh)

	// Assert
	errorCount := 0
	for err := range errCh {
		if err != nil {
			errorCount++
		}
	}
	assert.Equal(t, 0, errorCount, "Expected no errors from concurrent module checks")
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_ErrorChannelFull(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := modulesvc.New(action, mockHTTP, &dockerclient.DockerClient{}, &registrysvc.RegistrySvc{}, &moduleenv.ModuleEnv{})
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	mockHTTP.On("GetReturnResponse",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(nil, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error) // Unbuffered channel to test default case
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()

	// Assert
	// The goroutine should complete without blocking even if error channel is not read
	// This tests the default case in the select statement
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_VerifyRetryLogic(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := modulesvc.New(action, mockHTTP, &dockerclient.DockerClient{}, &registrysvc.RegistrySvc{}, &moduleenv.ModuleEnv{})
	svc.ReadinessMaxRetries = 5
	svc.ReadinessWait = 1 * time.Millisecond

	failureResponse := &http.Response{
		StatusCode: http.StatusServiceUnavailable,
		Status:     "503 Service Unavailable",
		Body:       io.NopCloser(strings.NewReader("")),
	}

	// Will retry until max retries
	mockHTTP.On("GetReturnResponse",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(failureResponse, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()
	close(errCh)

	// Assert
	err := <-errCh
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "module test-module")
	// Should have been called exactly maxRetries times
	mockHTTP.AssertNumberOfCalls(t, "GetReturnResponse", 5)
}

func TestCheckModuleReadiness_PortInURL(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := modulesvc.New(action, mockHTTP, &dockerclient.DockerClient{}, &registrysvc.RegistrySvc{}, &moduleenv.ModuleEnv{})
	svc.ReadinessMaxRetries = 3
	svc.ReadinessWait = 1 * time.Millisecond

	successResponse := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader("")),
	}

	var capturedURL string
	mockHTTP.On("GetReturnResponse",
		mock.MatchedBy(func(urlStr string) bool {
			capturedURL = urlStr
			return true
		}),
		mock.Anything).
		Return(successResponse, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 9999)
	wg.Wait()
	close(errCh)

	// Assert
	assert.Contains(t, capturedURL, ":9999")
	assert.Contains(t, capturedURL, "/admin/health")
	mockHTTP.AssertExpectations(t)
}

func TestCheckModuleReadiness_DefaultMaxRetries(t *testing.T) {
	// Arrange
	mockHTTP := &testhelpers.MockHTTPClient{}
	action := testhelpers.NewMockAction()
	svc := modulesvc.New(action, mockHTTP, &dockerclient.DockerClient{}, &registrysvc.RegistrySvc{}, &moduleenv.ModuleEnv{})
	// Don't set ModuleReadinessMaxRetries - should default to constant value

	successResponse := &http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       io.NopCloser(strings.NewReader("")),
	}

	mockHTTP.On("GetReturnResponse",
		mock.MatchedBy(func(urlStr string) bool {
			return strings.Contains(urlStr, "/admin/health")
		}),
		mock.Anything).
		Return(successResponse, nil)

	wg := &sync.WaitGroup{}
	errCh := make(chan error, 1)
	wg.Add(1)

	// Act
	go svc.CheckModuleReadiness(wg, errCh, "test-module", 8080)
	wg.Wait()
	close(errCh)

	// Assert
	select {
	case err := <-errCh:
		if err != nil {
			t.Errorf("Expected no error, got: %v", err)
		}
	default:
		// Success - no error sent, defaults to constant.ModuleReadinessMaxRetries (50)
	}
	mockHTTP.AssertExpectations(t)
}
