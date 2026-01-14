package errors_test

import (
	"errors"
	"fmt"
	"testing"

	apperrors "github.com/j011195/eureka-setup/eureka-cli/errors"
	"github.com/stretchr/testify/assert"
)

// ==================== Base Errors Tests ====================

func TestBaseErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{"ErrNotFound", apperrors.ErrNotFound, "resource not found"},
		{"ErrInvalidInput", apperrors.ErrInvalidInput, "invalid input"},
		{"ErrTimeout", apperrors.ErrTimeout, "operation timed out"},
		{"ErrNotReady", apperrors.ErrNotReady, "resource not ready"},
		{"ErrUnauthorized", apperrors.ErrUnauthorized, "unauthorized"},
		{"ErrConfigMissing", apperrors.ErrConfigMissing, "configuration missing"},
		{"ErrDeploymentFailed", apperrors.ErrDeploymentFailed, "deployment failed"},
		{"ErrAccessTokenBlank", apperrors.ErrAccessTokenBlank, "access token cannot be blank"},
		{"ErrTenantNameBlank", apperrors.ErrTenantNameBlank, "tenant name cannot be blank"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

// ==================== Generic Error Helpers Tests ====================

func TestWrap(t *testing.T) {
	t.Run("TestWrap_Success", func(t *testing.T) {
		// Arrange
		baseErr := errors.New("base error")
		message := "wrapped error"

		// Act
		result := apperrors.Wrap(baseErr, message)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), message)
		assert.Contains(t, result.Error(), "base error")
		assert.True(t, errors.Is(result, baseErr))
	})

	t.Run("TestWrap_NilError", func(t *testing.T) {
		// Arrange & Act
		result := apperrors.Wrap(nil, "message")

		// Assert
		assert.Nil(t, result)
	})
}

func TestWrapf(t *testing.T) {
	t.Run("TestWrapf_Success", func(t *testing.T) {
		// Arrange
		baseErr := errors.New("base error")
		format := "operation %s failed with code %d"

		// Act
		result := apperrors.Wrapf(baseErr, format, "test", 404)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "operation test failed with code 404")
		assert.Contains(t, result.Error(), "base error")
		assert.True(t, errors.Is(result, baseErr))
	})

	t.Run("TestWrapf_NilError", func(t *testing.T) {
		// Arrange & Act
		result := apperrors.Wrapf(nil, "format %s", "arg")

		// Assert
		assert.Nil(t, result)
	})
}

func TestNew(t *testing.T) {
	t.Run("TestNew_Success", func(t *testing.T) {
		// Arrange
		message := "custom error message"

		// Act
		result := apperrors.New(message)

		// Assert
		assert.Error(t, result)
		assert.Equal(t, message, result.Error())
	})
}

func TestNewf(t *testing.T) {
	t.Run("TestNewf_Success", func(t *testing.T) {
		// Arrange
		format := "error code %d: %s"

		// Act
		result := apperrors.Newf(format, 500, "internal server error")

		// Assert
		assert.Error(t, result)
		assert.Equal(t, "error code 500: internal server error", result.Error())
	})
}

// ==================== Validation Errors Tests ====================

func TestActionNil(t *testing.T) {
	t.Run("TestActionNil_Success", func(t *testing.T) {
		// Act
		result := apperrors.ActionNil()

		// Assert
		assert.Error(t, result)
		assert.Equal(t, "action cannot be nil", result.Error())
	})
}

func TestLoggerNil(t *testing.T) {
	t.Run("TestLoggerNil_Success", func(t *testing.T) {
		// Act
		result := apperrors.LoggerNil()

		// Assert
		assert.Error(t, result)
		assert.Equal(t, "logger cannot be nil", result.Error())
	})
}

func TestRequiredParameterMissing(t *testing.T) {
	t.Run("TestRequiredParameterMissing_Success", func(t *testing.T) {
		// Act
		result := apperrors.RequiredParameterMissing("tenant")

		// Assert
		assert.Error(t, result)
		assert.Equal(t, "invalid input: tenant parameter required", result.Error())
		assert.True(t, errors.Is(result, apperrors.ErrInvalidInput))
	})
}

func TestAccessTokenBlank(t *testing.T) {
	t.Run("TestAccessTokenBlank_Success", func(t *testing.T) {
		// Act
		result := apperrors.AccessTokenBlank()

		// Assert
		assert.Error(t, result)
		assert.Equal(t, "access token cannot be blank", result.Error())
		assert.True(t, errors.Is(result, apperrors.ErrAccessTokenBlank))
		assert.Equal(t, apperrors.ErrAccessTokenBlank, result)
	})
}

func TestTenantNameBlank(t *testing.T) {
	t.Run("TestTenantNameBlank_Success", func(t *testing.T) {
		// Act
		result := apperrors.TenantNameBlank()

		// Assert
		assert.Error(t, result)
		assert.Equal(t, "tenant name cannot be blank", result.Error())
		assert.True(t, errors.Is(result, apperrors.ErrTenantNameBlank))
		assert.Equal(t, apperrors.ErrTenantNameBlank, result)
	})
}

// ==================== HTTP Errors Tests ====================

func TestPingFailed(t *testing.T) {
	t.Run("TestPingFailed_Success", func(t *testing.T) {
		// Arrange
		url := "http://localhost:9130"
		baseErr := errors.New("connection refused")

		// Act
		result := apperrors.PingFailed(url, baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to ping http://localhost:9130")
		assert.True(t, errors.Is(result, baseErr))
	})
}

func TestPingFailedWithStatus(t *testing.T) {
	t.Run("TestPingFailedWithStatus_Success", func(t *testing.T) {
		// Arrange
		url := "http://example.com/health"
		statusCode := 404

		// Act
		result := apperrors.PingFailedWithStatus(url, statusCode)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to ping")
		assert.Contains(t, result.Error(), url)
		assert.Contains(t, result.Error(), "404")
		assert.Contains(t, result.Error(), "Not Found")
	})

	t.Run("TestPingFailedWithStatus_ServerError", func(t *testing.T) {
		// Arrange
		url := "http://localhost:8080"
		statusCode := 503

		// Act
		result := apperrors.PingFailedWithStatus(url, statusCode)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to ping http://localhost:8080")
		assert.Contains(t, result.Error(), "503")
		assert.Contains(t, result.Error(), "Service Unavailable")
	})
}

func TestRequestFailed(t *testing.T) {
	t.Run("TestRequestFailed_Success", func(t *testing.T) {
		// Arrange
		statusCode := 404
		method := "GET"
		url := "http://localhost:9130/api"

		// Act
		result := apperrors.RequestFailed(statusCode, method, url)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "request failed with status 404")
		assert.Contains(t, result.Error(), "GET http://localhost:9130/api")
	})
}

// ==================== Action Errors Tests ====================

func TestUnsupportedPlatform(t *testing.T) {
	t.Run("TestUnsupportedPlatform_Success", func(t *testing.T) {
		// Arrange
		platform := "docker"
		address := "mgr-tenants"

		// Act
		result := apperrors.UnsupportedPlatform(platform, address)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "unsupported docker platform for mgr-tenants")
	})
}

func TestGatewayURLConstructFailed(t *testing.T) {
	t.Run("TestGatewayURLConstructFailed_Success", func(t *testing.T) {
		// Arrange
		platform := "ecs"
		baseErr := errors.New("invalid config")

		// Act
		result := apperrors.GatewayURLConstructFailed(platform, baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to construct a gateway url for ecs platform")
		assert.True(t, errors.Is(result, baseErr))
	})
}

func TestNoFreeTCPPort(t *testing.T) {
	t.Run("TestNoFreeTCPPort_Success", func(t *testing.T) {
		// Arrange
		portStart := 8000
		portEnd := 9000

		// Act
		result := apperrors.NoFreeTCPPort(portStart, portEnd)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to find free TCP ports in range: 8000-9000")
	})
}

// ==================== AWS Errors Tests ====================

func TestAWSConfigLoadFailed(t *testing.T) {
	t.Run("TestAWSConfigLoadFailed_Success", func(t *testing.T) {
		// Arrange
		baseErr := errors.New("credentials not found")

		// Act
		result := apperrors.AWSConfigLoadFailed(baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to load AWS config")
		assert.True(t, errors.Is(result, baseErr))
	})
}

func TestECRAuthFailed(t *testing.T) {
	t.Run("TestECRAuthFailed_Success", func(t *testing.T) {
		// Arrange
		baseErr := errors.New("invalid credentials")

		// Act
		result := apperrors.ECRAuthFailed(baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to get ECR authorization token")
		assert.True(t, errors.Is(result, apperrors.ErrUnauthorized))
		assert.True(t, errors.Is(result, baseErr))
	})
}

func TestECRNoAuthData(t *testing.T) {
	t.Run("TestECRNoAuthData_Success", func(t *testing.T) {
		// Act
		result := apperrors.ECRNoAuthData()

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "no authorization data from ECR")
		assert.True(t, errors.Is(result, apperrors.ErrUnauthorized))
	})
}

func TestECRTokenNil(t *testing.T) {
	t.Run("TestECRTokenNil_Success", func(t *testing.T) {
		// Act
		result := apperrors.ECRTokenNil()

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "ECR authorization token is nil")
		assert.True(t, errors.Is(result, apperrors.ErrUnauthorized))
	})
}

func TestECRTokenDecodeFailed(t *testing.T) {
	t.Run("TestECRTokenDecodeFailed_Success", func(t *testing.T) {
		// Arrange
		baseErr := errors.New("invalid base64")

		// Act
		result := apperrors.ECRTokenDecodeFailed(baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to decode ECR authorization token")
		assert.True(t, errors.Is(result, apperrors.ErrUnauthorized))
		assert.True(t, errors.Is(result, baseErr))
	})
}

// ==================== Consortium Errors Tests ====================

func TestConsortiumMissingCentralTenant(t *testing.T) {
	t.Run("TestConsortiumMissingCentralTenant_Success", func(t *testing.T) {
		// Arrange
		consortiumName := "test-consortium"

		// Act
		result := apperrors.ConsortiumMissingCentralTenant(consortiumName)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "consortium test-consortium does not contain a central tenant")
		assert.True(t, errors.Is(result, apperrors.ErrConfigMissing))
	})
}

// ==================== File Errors Tests ====================

func TestNotRegularFile(t *testing.T) {
	t.Run("TestNotRegularFile_Success", func(t *testing.T) {
		// Arrange
		fileName := "/dev/null"

		// Act
		result := apperrors.NotRegularFile(fileName)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "/dev/null is not a regular file")
		assert.True(t, errors.Is(result, apperrors.ErrInvalidInput))
	})
}

// ==================== Git Errors Tests ====================

func TestCloneFailed(t *testing.T) {
	t.Run("TestCloneFailed_Success", func(t *testing.T) {
		// Arrange
		repoLabel := "platform-complete"
		baseErr := errors.New("repository not found")

		// Act
		result := apperrors.CloneFailed(repoLabel, baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to clone repository platform-complete")
		assert.True(t, errors.Is(result, baseErr))
	})
}

// ==================== Kafka Errors Tests ====================

func TestKafkaNotReady(t *testing.T) {
	t.Run("TestKafkaNotReady_Success", func(t *testing.T) {
		// Arrange
		baseErr := errors.New("connection timeout")

		// Act
		result := apperrors.KafkaNotReady(baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "kafka not ready")
		assert.True(t, errors.Is(result, apperrors.ErrNotReady))
		assert.True(t, errors.Is(result, baseErr))
	})
}

func TestKafkaBrokerAPIFailed(t *testing.T) {
	t.Run("TestKafkaBrokerAPIFailed_Success", func(t *testing.T) {
		// Act
		result := apperrors.KafkaBrokerAPIFailed()

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "no output from Kafka broker API")
		assert.True(t, errors.Is(result, apperrors.ErrNotReady))
	})
}

func TestConsumerGroupRebalanceTimeout(t *testing.T) {
	t.Run("TestConsumerGroupRebalanceTimeout_Success", func(t *testing.T) {
		// Arrange
		consumerGroup := "test-group"
		baseErr := errors.New("max retries exceeded")

		// Act
		result := apperrors.ConsumerGroupRebalanceTimeout(consumerGroup, baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "consumer group test-group rebalance exceeded")
		assert.True(t, errors.Is(result, apperrors.ErrTimeout))
		assert.True(t, errors.Is(result, baseErr))
	})
}

func TestConsumerGroupPollTimeout(t *testing.T) {
	t.Run("TestConsumerGroupPollTimeout_Success", func(t *testing.T) {
		// Arrange
		consumerGroup := "test-group"
		maxRetries := 10

		// Act
		result := apperrors.ConsumerGroupPollTimeout(consumerGroup, maxRetries)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "consumer group test-group polling exceeded maximum retries (10)")
		assert.True(t, errors.Is(result, apperrors.ErrTimeout))
	})
}

func TestContainerCommandFailed(t *testing.T) {
	t.Run("TestContainerCommandFailed_Success", func(t *testing.T) {
		// Arrange
		stderr := "permission denied"

		// Act
		result := apperrors.ContainerCommandFailed(stderr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to execute container command, stderr: permission denied")
	})
}

// ==================== Keycloak Errors Tests ====================

func TestAccessTokenNotFound(t *testing.T) {
	t.Run("TestAccessTokenNotFound_Success", func(t *testing.T) {
		// Arrange
		requestURL := "http://localhost:8080/token"

		// Act
		result := apperrors.AccessTokenNotFound(requestURL)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "access token from response: http://localhost:8080/token")
		assert.True(t, errors.Is(result, apperrors.ErrNotFound))
	})
}

func TestClientNotFound(t *testing.T) {
	t.Run("TestClientNotFound_Success", func(t *testing.T) {
		// Arrange
		clientID := "test-client"

		// Act
		result := apperrors.ClientNotFound(clientID)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "expected exactly 1 client with id test-client")
		assert.True(t, errors.Is(result, apperrors.ErrNotFound))
	})
}

func TestRoleNotFound(t *testing.T) {
	t.Run("TestRoleNotFound_Success", func(t *testing.T) {
		// Arrange
		roleName := "admin"

		// Act
		result := apperrors.RoleNotFound(roleName)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "expected exactly 1 role with name admin")
		assert.True(t, errors.Is(result, apperrors.ErrNotFound))
	})
}

func TestUserNotFound(t *testing.T) {
	t.Run("TestUserNotFound_Success", func(t *testing.T) {
		// Arrange
		username := "testuser"
		tenantName := "diku"

		// Act
		result := apperrors.UserNotFound(username, tenantName)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "user testuser in tenant diku")
		assert.True(t, errors.Is(result, apperrors.ErrNotFound))
	})
}

// ==================== Kong Errors Tests ====================

func TestKongRoutesNotReady(t *testing.T) {
	t.Run("TestKongRoutesNotReady_Success", func(t *testing.T) {
		// Arrange
		expected := 5

		// Act
		result := apperrors.KongRoutesNotReady(expected)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "kong routes 5")
		assert.True(t, errors.Is(result, apperrors.ErrNotReady))
	})
}

func TestKongAdminAPIFailed(t *testing.T) {
	t.Run("TestKongAdminAPIFailed_Success", func(t *testing.T) {
		// Arrange
		statusCode := 500
		status := "Internal Server Error"

		// Act
		result := apperrors.KongAdminAPIFailed(statusCode, status)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "kong admin API failed: 500 Internal Server Error")
	})
}

// ==================== Application Errors Tests ====================

func TestApplicationNotFound(t *testing.T) {
	t.Run("TestApplicationNotFound_Success", func(t *testing.T) {
		// Arrange
		applicationName := "platform-complete"

		// Act
		result := apperrors.ApplicationNotFound(applicationName)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to find the latest application for platform-complete profile")
		assert.True(t, errors.Is(result, apperrors.ErrNotFound))
	})
}

// ==================== Module Errors Tests ====================

func TestModulesNotDeployed(t *testing.T) {
	t.Run("TestModulesNotDeployed_Success", func(t *testing.T) {
		// Arrange
		expectedModules := 3

		// Act
		result := apperrors.ModulesNotDeployed(expectedModules)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "3 modules not deployed")
	})
}

func TestModuleNotReady(t *testing.T) {
	t.Run("TestModuleNotReady_Success", func(t *testing.T) {
		// Arrange
		moduleName := "mod-inventory"

		// Act
		result := apperrors.ModuleNotReady(moduleName)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "module mod-inventory")
		assert.True(t, errors.Is(result, apperrors.ErrNotReady))
	})
}

func TestModulePullFailed(t *testing.T) {
	t.Run("TestModulePullFailed_Success", func(t *testing.T) {
		// Arrange
		imageName := "mod-inventory:1.0.0"
		baseErr := errors.New("image not found")

		// Act
		result := apperrors.ModulePullFailed(imageName, baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to pull module image mod-inventory:1.0.0")
		assert.True(t, errors.Is(result, apperrors.ErrDeploymentFailed))
		assert.True(t, errors.Is(result, baseErr))
	})
}

func TestSidecarDeployFailed(t *testing.T) {
	t.Run("TestSidecarDeployFailed_Success", func(t *testing.T) {
		// Arrange
		sidecarName := "mod-inventory-sidecar"
		baseErr := errors.New("port already allocated")

		// Act
		result := apperrors.SidecarDeployFailed(sidecarName, baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to deploy sidecar mod-inventory-sidecar")
		assert.True(t, errors.Is(result, apperrors.ErrDeploymentFailed))
		assert.True(t, errors.Is(result, baseErr))
	})
}

func TestSidecarVersionNotFound(t *testing.T) {
	t.Run("TestSidecarVersionNotFound_Success", func(t *testing.T) {
		// Act
		result := apperrors.SidecarVersionNotFound()

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "sidecar version in registry")
		assert.True(t, errors.Is(result, apperrors.ErrNotFound))
	})
}

func TestLocalDescriptorNotFound(t *testing.T) {
	t.Run("TestLocalDescriptorNotFound_Success", func(t *testing.T) {
		// Arrange
		path := "/path/to/descriptor.json"
		moduleName := "mod-inventory"

		// Act
		result := apperrors.LocalDescriptorNotFound(path, moduleName)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "local descriptor /path/to/descriptor.json for module mod-inventory")
		assert.True(t, errors.Is(result, apperrors.ErrNotFound))
	})
}

func TestEmptyLineNotFound(t *testing.T) {
	t.Run("TestEmptyLineNotFound_Success", func(t *testing.T) {
		// Arrange
		id := "test-id"

		// Act
		result := apperrors.EmptyLineNotFound(id)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "response does not contain an empty line using id test-id")
	})
}

func TestImageKeyNotSet(t *testing.T) {
	t.Run("TestImageKeyNotSet_Success", func(t *testing.T) {
		// Arrange
		imageName := "mod-inventory"
		fieldName := "version"

		// Act
		result := apperrors.ImageKeyNotSet(imageName, fieldName)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "cannot run image mod-inventory, key version not set in config")
		assert.True(t, errors.Is(result, apperrors.ErrConfigMissing))
	})
}

func TestModuleDiscoveryNotFound(t *testing.T) {
	t.Run("TestModuleDiscoveryNotFound_Success", func(t *testing.T) {
		// Arrange
		moduleName := "mod-inventory"

		// Act
		result := apperrors.ModuleDiscoveryNotFound(moduleName)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "module discovery mod-inventory in application")
		assert.True(t, errors.Is(result, apperrors.ErrNotFound))
	})
}

func TestModuleDescriptorNotFound(t *testing.T) {
	t.Run("TestModuleDescriptorNotFound_Success", func(t *testing.T) {
		// Arrange
		moduleName := "mod-inventory"
		moduleVersion := "1.2.3"
		descriptorPath := "/path/to/target/ModuleDescriptor.json"

		// Act
		result := apperrors.ModuleDescriptorNotFound(moduleName, moduleVersion, descriptorPath)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "module descriptor for mod-inventory-1.2.3")
		assert.Contains(t, result.Error(), "/path/to/target/ModuleDescriptor.json")
		assert.True(t, errors.Is(result, apperrors.ErrNotFound))
	})
}

func TestModulePathNotFound(t *testing.T) {
	t.Run("TestModulePathNotFound_Success", func(t *testing.T) {
		// Arrange
		modulePath := "/path/to/module"

		// Act
		result := apperrors.ModulePathNotFound(modulePath)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "module path does not exist: /path/to/module")
	})
}

func TestModulePathAccessFailed(t *testing.T) {
	t.Run("TestModulePathAccessFailed_Success", func(t *testing.T) {
		// Arrange
		modulePath := "/restricted/path"
		baseErr := errors.New("permission denied")

		// Act
		result := apperrors.ModulePathAccessFailed(modulePath, baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to access module path /restricted/path")
		assert.True(t, errors.Is(result, baseErr))
	})
}

func TestModulePathNotDirectory(t *testing.T) {
	t.Run("TestModulePathNotDirectory_Success", func(t *testing.T) {
		// Arrange
		modulePath := "/path/to/file.txt"

		// Act
		result := apperrors.ModulePathNotDirectory(modulePath)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "module path is not a directory: /path/to/file.txt")
	})
}

// ==================== Tenant Errors Tests ====================

func TestTenantNotFound(t *testing.T) {
	t.Run("TestTenantNotFound_Success", func(t *testing.T) {
		// Arrange
		tenantName := "diku"

		// Act
		result := apperrors.TenantNotFound(tenantName)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "tenant diku in config")
		assert.True(t, errors.Is(result, apperrors.ErrNotFound))
	})
}

func TestCentralTenantNotFound(t *testing.T) {
	t.Run("TestCentralTenantNotFound_Success", func(t *testing.T) {
		// Arrange
		consortiumName := "test-consortium"

		// Act
		result := apperrors.CentralTenantNotFound(consortiumName)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "central tenant in consortium test-consortium")
		assert.True(t, errors.Is(result, apperrors.ErrNotFound))
	})
}

func TestTenantNotCreated(t *testing.T) {
	t.Run("TestTenantNotCreated_Success", func(t *testing.T) {
		// Arrange
		tenantName := "diku"

		// Act
		result := apperrors.TenantNotCreated(tenantName)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "consortium tenant diku not created")
		assert.True(t, errors.Is(result, apperrors.ErrDeploymentFailed))
	})
}

// ==================== Search/Reindex Errors Tests ====================

func TestReindexJobHasErrors(t *testing.T) {
	t.Run("TestReindexJobHasErrors_Success", func(t *testing.T) {
		// Arrange
		jobErrors := []any{"error1", "error2", "error3"}

		// Act
		result := apperrors.ReindexJobHasErrors(jobErrors)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "reindex job has 3 error(s)")
		assert.Contains(t, result.Error(), fmt.Sprintf("%+v", jobErrors))
	})
}

func TestReindexJobIDBlank(t *testing.T) {
	t.Run("TestReindexJobIDBlank_Success", func(t *testing.T) {
		// Act
		result := apperrors.ReindexJobIDBlank()

		// Assert
		assert.Error(t, result)
		assert.Equal(t, "reindex job id is blank", result.Error())
	})
}

// ==================== Registry Errors Tests ====================

func TestLocalInstallFileNotFound(t *testing.T) {
	t.Run("TestLocalInstallFileNotFound_Success", func(t *testing.T) {
		// Arrange
		baseErr := errors.New("file not found")

		// Act
		result := apperrors.LocalInstallFileNotFound(baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to find local install file")
		assert.Contains(t, result.Error(), "file not found")
		assert.True(t, errors.Is(result, apperrors.ErrNotFound))
		assert.True(t, errors.Is(result, baseErr))
	})
}

// ==================== Flag Error Messages Tests ====================

func TestRegisterFlagCompletionFailed(t *testing.T) {
	t.Run("TestRegisterFlagCompletionFailed_Success", func(t *testing.T) {
		// Arrange
		baseErr := errors.New("completion function error")

		// Act
		result := apperrors.RegisterFlagCompletionFailed(baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to register flag completion function")
		assert.Contains(t, result.Error(), "completion function error")
		assert.True(t, errors.Is(result, baseErr))
	})
}

func TestMarkFlagRequiredFailed(t *testing.T) {
	// Create mock flag structs for testing
	type mockFlag struct {
		name string
	}

	mockFlagGetName := func(f mockFlag) string {
		return f.name
	}

	// Create a simple flag implementation
	moduleName := mockFlag{name: "moduleName"}
	tenant := mockFlag{name: "tenant"}

	t.Run("TestMarkFlagRequiredFailed_ModuleName", func(t *testing.T) {
		// Arrange
		baseErr := errors.New("mark required error")

		// Create adapter that implements Flag interface
		flagAdapter := flagAdapter{name: mockFlagGetName(moduleName)}

		// Act
		result := apperrors.MarkFlagRequiredFailed(flagAdapter, baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to mark moduleName flag as required")
		assert.Contains(t, result.Error(), "mark required error")
		assert.True(t, errors.Is(result, baseErr))
	})

	t.Run("TestMarkFlagRequiredFailed_Tenant", func(t *testing.T) {
		// Arrange
		baseErr := errors.New("mark required error")

		// Create adapter that implements Flag interface
		flagAdapter := flagAdapter{name: mockFlagGetName(tenant)}

		// Act
		result := apperrors.MarkFlagRequiredFailed(flagAdapter, baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "failed to mark tenant flag as required")
		assert.Contains(t, result.Error(), "mark required error")
		assert.True(t, errors.Is(result, baseErr))
	})
}

// ==================== Version Errors Tests ====================

func TestVersionEmpty(t *testing.T) {
	t.Run("TestVersionEmpty_Success", func(t *testing.T) {
		// Act
		result := apperrors.VersionEmpty()

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "version cannot be empty")
		assert.True(t, errors.Is(result, apperrors.ErrInvalidInput))
	})
}

func TestNotSnapshotVersion(t *testing.T) {
	t.Run("TestNotSnapshotVersion_Success", func(t *testing.T) {
		// Arrange
		version := "1.0.0"

		// Act
		result := apperrors.NotSnapshotVersion(version)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "1.0.0")
		assert.Contains(t, result.Error(), "not a SNAPSHOT version with build number")
		assert.True(t, errors.Is(result, apperrors.ErrInvalidInput))
	})
}

func TestInvalidSnapshotFormat(t *testing.T) {
	t.Run("TestInvalidSnapshotFormat_Success", func(t *testing.T) {
		// Arrange
		version := "1.0.0-SNAPSHOT"

		// Act
		result := apperrors.InvalidSnapshotFormat(version)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "1.0.0-SNAPSHOT")
		assert.Contains(t, result.Error(), "invalid SNAPSHOT version format")
		assert.True(t, errors.Is(result, apperrors.ErrInvalidInput))
	})
}

func TestInvalidBuildNumber(t *testing.T) {
	t.Run("TestInvalidBuildNumber_Success", func(t *testing.T) {
		// Arrange
		version := "1.0.0-SNAPSHOT.abc"
		baseErr := errors.New("parse error")

		// Act
		result := apperrors.InvalidBuildNumber(version, baseErr)

		// Assert
		assert.Error(t, result)
		assert.Contains(t, result.Error(), "1.0.0-SNAPSHOT.abc")
		assert.Contains(t, result.Error(), "invalid build number")
		assert.True(t, errors.Is(result, apperrors.ErrInvalidInput))
		assert.True(t, errors.Is(result, baseErr))
	})
}

// flagAdapter is a simple adapter to implement the Flag interface for testing
type flagAdapter struct {
	name string
}

func (f flagAdapter) GetName() string {
	return f.name
}
