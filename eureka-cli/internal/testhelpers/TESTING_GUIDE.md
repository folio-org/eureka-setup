# Testing System Prompt for Eureka-CLI

## Overview

This document provides a comprehensive guide for writing consistent, maintainable, and efficient tests for the eureka-cli project using Go's standard library testing package and the testify assertion library.

## Core Testing Principles

### 1. Test Organization

- **File naming**: `<package>_test.go` (e.g., `search_svc_test.go`)
- **Package naming**: Use `<package>_test` for black-box testing (preferred for exported functions)
- **Package naming**: Use `<package>` for white-box testing (when testing internal/unexported functions)
- **Test function naming**: `Test<FunctionName>_<Scenario>` (e.g., `TestReindexInventoryRecords_Success`)
- **Subtest naming**: Subtests created with `t.Run()` MUST start with the parent test function name (e.g., `t.Run("TestFunctionName_Scenario", ...)` not `t.Run("Scenario", ...)`)

### 2. Test Structure (AAA Pattern)

Every test should follow the Arrange-Act-Assert pattern:

```go
func TestFunctionName_Scenario(t *testing.T) {
  // Arrange: Set up test data, mocks, and dependencies
  mockHTTP := &testhelpers.MockHTTPClient{}
  action := testhelpers.NewMockAction()
  svc := New(action, mockHTTP)
  
  // Act: Execute the function under test
  result, err := svc.SomeFunction(input)
  
  // Assert: Verify the results
  assert.NoError(t, err)
  assert.Equal(t, expected, result)
  mockHTTP.AssertExpectations(t)
}
```

**Important**: When using subtests with `t.Run()`, the subtest name MUST start with the parent test function name:

```go
func TestPingFailedWithStatus(t *testing.T) {
  t.Run("TestPingFailedWithStatus_Success", func(t *testing.T) {  // ✅ Correct
    // test code
  })
  
  t.Run("Success", func(t *testing.T) {  // ❌ Incorrect - missing parent name
    // test code
  })
}
```

### 3. Test Coverage Strategy

#### Unit Tests

- **Focus**: Individual functions/methods in isolation
- **Dependencies**: All external dependencies mocked
- **Scope**: Single package
- **Location**: Same directory as source code
- **Examples**:
  - Testing HTTP client methods with mock servers
  - Testing service methods with mock HTTP clients
  - Testing validation functions
  - Testing error handling paths

#### Integration Tests

- **Focus**: Multiple components working together
- **Dependencies**: Some real implementations, critical mocks only
- **Scope**: Cross-package interactions
- **Location**: Separate `integration_test` directory or marked with build tags
- **Examples**:
  - Testing service → HTTP client → mock server flow
  - Testing multiple services interacting
  - Testing full request-response cycles

### 4. HTTP Testing Patterns

#### Pattern A: Mock HTTP Server (for testing HTTP clients)

```go
func TestHTTPClient_GetRequest(t *testing.T) {
  // Arrange
  expectedResponse := models.SomeResponse{ID: "123", Name: "test"}
  mockServer := testhelpers.NewMockHTTPServer(t, 
    testhelpers.JSONResponse(http.StatusOK, expectedResponse))
  defer mockServer.Close()
  
  client := httpclient.New(action, logger)
  
  // Act
  var result models.SomeResponse
  err := client.GetRetryReturnStruct(mockServer.URL()+"/endpoint", nil, &result)
  
  // Assert
  assert.NoError(t, err)
  assert.Equal(t, expectedResponse, result)
  mockServer.AssertRequestCount(1)
  mockServer.AssertRequestMethod(0, "GET")
}
```

#### Pattern B: Mock HTTP Client (for testing services)

```go
func TestSearchSvc_ReindexInventoryRecords_Success(t *testing.T) {
  // Arrange
  mockHTTP := &testhelpers.MockHTTPClient{}
  action := testhelpers.NewMockAction()
  svc := New(action, mockHTTP)
  
  expectedJob := models.ReindexJobResponse{
    ID: "job-123",
    JobStatus: "COMPLETED",
  }
  
  mockHTTP.On("PostReturnStruct", 
    mock.Anything, 
    mock.Anything, 
    mock.Anything, 
    mock.Anything).
    Run(func(args mock.Arguments) {
      // Populate the target argument
      target := args.Get(3).(*models.ReindexJobResponse)
      *target = expectedJob
    }).
    Return(nil)
  
  // Act
  err := svc.ReindexInventoryRecords("test-tenant")
  
  // Assert
  assert.NoError(t, err)
  mockHTTP.AssertExpectations(t)
}
```

### 5. Table-Driven Tests

Use table-driven tests for multiple scenarios of the same function:

```go
func TestValidateInventoryRecordsResponse(t *testing.T) {
  tests := []struct {
    name        string
    job         models.ReindexJobResponse
    expectError bool
    errorType   error
  }{
    {
      name: "success with valid job",
      job: models.ReindexJobResponse{
        ID: "job-123",
        JobStatus: "COMPLETED",
      },
      expectError: false,
    },
    {
      name: "error when job has errors",
      job: models.ReindexJobResponse{
        ID: "job-123",
        Errors: []models.ReindexJobError{{Type: "ERROR", Message: "failed"}},
      },
      expectError: true,
    },
    {
      name: "error when job ID is blank",
      job: models.ReindexJobResponse{
        ID: "",
        JobStatus: "COMPLETED",
      },
      expectError: true,
    },
  }
  
  for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
      // Arrange
      svc := &SearchSvc{}
      
      // Act
      err := svc.validateInventoryRecordsResponse(tt.job)
      
      // Assert
      if tt.expectError {
        assert.Error(t, err)
      } else {
        assert.NoError(t, err)
      }
    })
  }
}
```

### 6. Testing Error Scenarios

Always test both happy path and error paths:

```go
func TestSearchSvc_ReindexInventoryRecords_HTTPError(t *testing.T) {
  // Arrange
  mockHTTP := &testhelpers.MockHTTPClient{}
  action := testhelpers.NewMockAction()
  svc := New(action, mockHTTP)
  
  expectedError := errors.New("network error")
  mockHTTP.On("PostReturnStruct", 
    mock.Anything, 
    mock.Anything, 
    mock.Anything, 
    mock.Anything).
    Return(expectedError)
  
  // Act
  err := svc.ReindexInventoryRecords("test-tenant")
  
  // Assert
  assert.NoError(t, err) // Function continues on error
  mockHTTP.AssertExpectations(t)
}
```

### 7. Assertion Library Usage

Use testify assertions for clear, readable tests:

```go
// Basic assertions
assert.Equal(t, expected, actual)
assert.NotEqual(t, unexpected, actual)
assert.True(t, condition)
assert.False(t, condition)
assert.Nil(t, obj)
assert.NotNil(t, obj)

// Error assertions
assert.NoError(t, err)
assert.Error(t, err)
assert.EqualError(t, err, "expected error message")
assert.ErrorIs(t, err, expectedErrorType)

// Collection assertions
assert.Len(t, collection, expectedLength)
assert.Contains(t, collection, element)
assert.ElementsMatch(t, expected, actual) // Order-independent
assert.Empty(t, collection)
assert.NotEmpty(t, collection)

// Use require for critical assertions (stops test on failure)
require.NoError(t, err)
require.NotNil(t, obj)
```

### 8. Test Helpers Location

- **Package**: `internal/testhelpers`
- **Purpose**: Reusable testing utilities
- **Contents**:
  - HTTP mock helpers (`http_helpers.go`)
  - Interface mocks (`mocks.go`)
  - Common test data builders
  - Assertion helpers

### 9. Testing Best Practices

#### DO

✅ Test one thing per test function
✅ Use descriptive test names that explain the scenario
✅ Mock external dependencies (HTTP, database, file system)
✅ Test both success and failure paths
✅ Use table-driven tests for multiple similar scenarios
✅ Clean up resources (defer server.Close(), etc.)
✅ Verify mock expectations (mockHTTP.AssertExpectations(t))
✅ Keep tests independent (no shared state)
✅ Test edge cases and boundary conditions
✅ Test with empty strings and nil values where applicable
✅ Test URL query parameter formatting and escaping
✅ Remove TODO comments from source code once tests are written
✅ Run tests after writing them to verify they pass

#### DON'T

❌ Don't test external services (use mocks)
❌ Don't share state between tests
❌ Don't use time.Sleep() (use channels or sync primitives)
❌ Don't ignore errors in test setup
❌ Don't write flaky tests
❌ Don't test implementation details (test behavior)
❌ Don't copy-paste test code (use helpers)
❌ Don't forget to test error paths (header errors, HTTP errors)
❌ Don't leave duplicate test names in the file

### 10. Service-Specific Testing Guidelines

#### SearchSvc Testing

- **Focus**: Reindexing operations, validation logic
- **Mocks**: HTTPClient
- **Key scenarios**:
  - Successful reindex with valid response
  - HTTP request failures (continue processing)
  - Invalid job responses (errors, blank IDs)
  - Multiple inventory records iteration

#### KeycloakSvc Testing

- **Focus**: User/role/capability management
- **Mocks**: HTTPClient, VaultClient
- **Key scenarios**:
  - CRUD operations for users/roles/capabilities
  - Token-based authentication
  - Tenant-specific operations
  - Error handling for 404, 409, etc.

#### ManagementSvc Testing

- **Focus**: Tenant and application management
- **Mocks**: HTTPClient, TenantSvc
- **Key scenarios**:
  - Tenant creation/deletion
  - Application deployment
  - Module descriptor handling
  - Batch operations
  - Tenant entitlement operations (create, remove, retrieve)
  - Application version retrieval (latest version queries)
  - Module discovery operations

#### ModuleSvc Testing

- **Focus**: Module provisioning and image management
- **Mocks**: HTTPClient, RegistrySvc, ModuleEnv, DockerClient
- **Key scenarios**:
  - Module image formatting (standard, custom, local)
  - Sidecar image resolution
  - Environment variable configuration
  - Module readiness checks
  - Version resolution logic
  - Edge cases with empty/special characters in parameters

#### HTTPClient Testing

- **Focus**: HTTP operations, retry logic, error handling
- **Mocks**: httptest.Server
- **Key scenarios**:
  - Successful requests (GET, POST, PUT, DELETE)
  - Error responses (4xx, 5xx)
  - Retry behavior
  - Response body parsing
  - Connection reuse

### 11. Test Data Management

Create builders for complex test data:

```go
// test_builders.go
package testhelpers

func NewTestReindexJobResponse() *models.ReindexJobResponse {
    return &models.ReindexJobResponse{
        ID: "test-job-id",
        JobStatus: "COMPLETED",
    }
}

func NewTestReindexJobResponseWithErrors(errors ...models.ReindexJobError) *models.ReindexJobResponse {
    return &models.ReindexJobResponse{
        ID: "test-job-id",
        Errors: errors,
    }
}
```

### 12. Running Tests

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./searchsvc/...

# Run specific test
go test ./searchsvc/ -run TestReindexInventoryRecords_Success

# Run with race detection
go test -race ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 13. Continuous Testing Strategy

#### Phase 1: Core Utilities (Priority: HIGH)

1. HTTP client methods (GET, POST, PUT, DELETE)
2. Error package functions
3. Helper functions (converters, validators)

#### Phase 2: Service Layer (Priority: MEDIUM)

1. SearchSvc - reindexing operations
2. KeycloakSvc - authentication and user management
3. ManagementSvc - tenant and application operations
4. RegistrySvc - module registry operations
5. ModuleSvc - module provisioning and image management
6. TenantSvc - tenant parameters and configuration

#### Phase 3: Integration Tests (Priority: LOW)

1. End-to-end service flows
2. Multi-service interactions
3. Error propagation across layers

### 14. Test Maintenance

- **Review**: Tests should be reviewed as part of PR process
- **Update**: Tests must be updated when functionality changes
- **Refactor**: Refactor tests when they become hard to read
- **Coverage**: Aim for >80% coverage on critical paths
- **Performance**: Keep tests fast (<1s per test)

## Example Test File Template

```go
package searchsvc_test

import (
    "testing"
    
    "github.com/j011195/eureka-setup/eureka-cli/internal/testhelpers"
    "github.com/j011195/eureka-setup/eureka-cli/models"
    "github.com/j011195/eureka-setup/eureka-cli/searchsvc"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestServiceFunction_SuccessScenario(t *testing.T) {
    // Arrange
    mockHTTP := &testhelpers.MockHTTPClient{}
    action := testhelpers.NewMockAction()
    svc := searchsvc.New(action, mockHTTP)
    
    // Setup expectations
    mockHTTP.On("Method", mock.Anything).Return(nil)
    
    // Act
    result, err := svc.ServiceFunction()
    
    // Assert
    assert.NoError(t, err)
    assert.NotNil(t, result)
    mockHTTP.AssertExpectations(t)
}

func TestServiceFunction_ErrorScenario(t *testing.T) {
    // Similar structure, test error path
}
```

## Token Efficiency Guidelines

To maximize test coverage while minimizing token usage:

1. **Use table-driven tests** for multiple scenarios
2. **Reuse test helpers** instead of duplicating setup code
3. **Focus on critical paths** first, edge cases second
4. **Batch related tests** in the same file
5. **Use descriptive but concise test names**
6. **Leverage testhelpers package** for common patterns
7. **Start with unit tests** before integration tests
8. **Test exported functions** before internal helpers

## Summary

This system prompt ensures:

- ✅ Consistent test structure across the codebase
- ✅ Efficient use of mocking and test helpers
- ✅ Clear separation between unit and integration tests
- ✅ Comprehensive coverage of success and error paths
- ✅ Maintainable and readable test code
- ✅ Token-efficient test development strategy
