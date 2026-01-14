# Internal Test Helpers Package

## Overview

This package provides reusable testing utilities for the eureka-cli project, enabling consistent, maintainable, and efficient unit and integration tests.

## Package Contents

### 1. `http_helpers.go`

HTTP testing utilities for mocking HTTP servers and responses:

- **MockHTTPServer**: Test HTTP server with request capture and assertion helpers
- **JSONResponse**: Handler for JSON responses
- **ErrorResponse**: Handler for error responses  
- **EmptyResponse**: Handler for empty responses
- **SequentialResponses**: Handler for sequential different responses
- **Assertion helpers**: Request count, method, path, headers, body validation

### 2. `mocks.go`

Mock implementations of key interfaces:

- **MockHTTPClient**: Mock for `httpclient.HTTPClientRunner` interface
- **MockAction**: Mock for `action.Action` struct
- **MockCommandExecutor**: Mock for `execsvc.CommandRunner` interface
- **MockRegistrySvc**: Mock for `registrysvc.RegistryProcessor` interface
- **MockModuleEnv**: Mock for `moduleenv.ModuleEnvProcessor` interface
- **MockDockerClient**: Mock for `dockerclient.DockerClientRunner` interface
- **MockTenantSvc**: Mock for `tenantsvc.TenantProcessor` interface
- Additional mocks can be added as needed

### 3. `doc.go`

Package documentation

### 4. `TESTING_GUIDE.md`

Comprehensive testing system prompt containing:

- Testing principles and patterns
- Test organization guidelines
- HTTP testing patterns (mock server vs mock client)
- Table-driven test examples
- Error scenario testing
- Service-specific testing strategies
- Test maintenance and coverage guidelines
- Token-efficient testing approach

## Quick Start

### Install Dependencies

```bash
go get github.com/stretchr/testify
```

### Example: Testing with Mock HTTP Server

```go
package httpclient_test

import (
  "testing"
  "net/http"
  
  "github.com/j011195/eureka-setup/eureka-cli/internal/testhelpers"
  "github.com/stretchr/testify/assert"
)

func TestHTTPClient_GetRequest(t *testing.T) {
  // Create mock server
  mockServer := testhelpers.NewMockHTTPServer(t, 
    testhelpers.JSONResponse(http.StatusOK, map[string]string{"status": "ok"}))
  defer mockServer.Close()
  
  // Test code here...
  
  // Assertions
  mockServer.AssertRequestCount(1)
  mockServer.AssertRequestMethod(0, "GET")
}
```

### Example: Testing with Mock HTTP Client

```go
package searchsvc_test

import (
  "testing"
  
  "github.com/j011195/eureka-setup/eureka-cli/internal/testhelpers"
  "github.com/j011195/eureka-setup/eureka-cli/searchsvc"
  "github.com/stretchr/testify/assert"
  "github.com/stretchr/testify/mock"
)

func TestSearchSvc_ReindexInventoryRecords(t *testing.T) {
  // Setup mocks
  mockHTTP := &testhelpers.MockHTTPClient{}
  action := testhelpers.NewMockAction()
  svc := searchsvc.New(action, mockHTTP)
  
  // Configure expectations
  mockHTTP.On("PostReturnStruct", 
    mock.Anything, 
    mock.Anything, 
    mock.Anything, 
    mock.Anything).Return(nil)
  
  // Test code here...
  
  // Verify expectations
  mockHTTP.AssertExpectations(t)
}
```

## Testing Strategy

### Phase 1: Core Utilities (High Priority)

- HTTP client methods
- Error package functions
- Helper utilities

### Phase 2: Service Layer (Medium Priority)

- SearchSvc - reindexing operations
- KeycloakSvc - authentication and user management
- ManagementSvc - tenant and application operations
- RegistrySvc - module registry operations
- ModuleSvc - module provisioning and image management
- TenantSvc - tenant parameters and configuration

### Phase 3: Integration Tests (Low Priority)

- End-to-end flows
- Multi-service interactions

## Best Practices

✅ **DO**:

- Use table-driven tests for multiple scenarios
- Mock all external dependencies
- Test both success and error paths
- Keep tests independent
- Use descriptive test names
- Test edge cases with empty strings and special characters
- Verify URL parameter formatting and escaping
- Remove TODO comments once tests are written
- Run tests immediately after writing them

❌ **DON'T**:

- Test external services directly
- Share state between tests
- Use time.Sleep()
- Ignore test setup errors
- Write flaky tests
- Leave duplicate test function names
- Forget to test error paths (headers, HTTP errors)

## Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./searchsvc/...

# Run with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Contributing

When adding new test helpers:

1. Add to appropriate file (`http_helpers.go` or `mocks.go`)
2. Document with clear comments
3. Add usage examples to TESTING_GUIDE.md
4. Ensure helpers are reusable across packages
5. Keep helpers simple and focused

## References

- [TESTING_GUIDE.md](./TESTING_GUIDE.md) - Comprehensive testing system prompt
- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify](https://github.com/stretchr/testify)
- [Go Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)
