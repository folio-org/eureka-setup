package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// FlagReader interface allows us to accept flag structs without importing the flags package
type FlagReader interface {
	GetName() string
}

// ==================== Base Errors ====================

var (
	ErrNotFound         = errors.New("resource not found")
	ErrInvalidInput     = errors.New("invalid input")
	ErrTimeout          = errors.New("operation timed out")
	ErrNotReady         = errors.New("resource not ready")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrConfigMissing    = errors.New("configuration missing")
	ErrDeploymentFailed = errors.New("deployment failed")
)

// ==================== Generic Error Helpers ====================

func Wrap(err error, message string) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", message, err)
}

func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}
	message := fmt.Sprintf(format, args...)
	return fmt.Errorf("%s: %w", message, err)
}

func New(message string) error {
	return errors.New(message)
}

func Newf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

// ==================== Validation Errors ====================

func ActionNil() error {
	return errors.New("action cannot be nil")
}

func LoggerNil() error {
	return errors.New("logger cannot be nil")
}

func RequiredParameterMissing(param string) error {
	return fmt.Errorf("%w: %s parameter required", ErrInvalidInput, param)
}

// ==================== HTTP Errors ====================

func PingFailed(url string, err error) error {
	return fmt.Errorf("failed to ping %s: %w", url, err)
}

func PingFailedWithStatus(url string, statusCode int) error {
	return fmt.Errorf("failed to ping %s: received status code %d (%s)", url, statusCode, http.StatusText(statusCode))
}

func RequestFailed(statusCode int, method, url string) error {
	return fmt.Errorf("request failed with status %d for URL: %s %s", statusCode, method, url)
}

// ==================== Action Errors ====================

func UnsupportedPlatform(platform, address string) error {
	return fmt.Errorf("unsupported %s platform for %s", platform, address)
}

func GatewayURLConstructFailed(platform string, err error) error {
	return fmt.Errorf("failed to construct a gateway url for %s platform: %w", platform, err)
}

func NoFreeTCPPort(portStart, portEnd int) error {
	return fmt.Errorf("failed to find free TCP ports in range: %d-%d", portStart, portEnd)
}

// ==================== AWS Errors ====================

func AWSConfigLoadFailed(err error) error {
	return fmt.Errorf("failed to load AWS config: %w", err)
}

func ECRAuthFailed(err error) error {
	return fmt.Errorf("%w: failed to get ECR authorization token: %w", ErrUnauthorized, err)
}

func ECRNoAuthData() error {
	return fmt.Errorf("%w: no authorization data from ECR", ErrUnauthorized)
}

func ECRTokenNil() error {
	return fmt.Errorf("%w: ECR authorization token is nil", ErrUnauthorized)
}

func ECRTokenDecodeFailed(err error) error {
	return fmt.Errorf("%w: failed to decode ECR authorization token: %w", ErrUnauthorized, err)
}

// ==================== Consortium Errors ====================

func ConsortiumMissingCentralTenant(consortiumName string) error {
	return fmt.Errorf("%w: consortium %s does not contain a central tenant", ErrConfigMissing, consortiumName)
}

// ==================== File Errors ====================

func NotRegularFile(fileName string) error {
	return fmt.Errorf("%w: %s is not a regular file", ErrInvalidInput, fileName)
}

// ==================== Git Errors ====================

func CloneFailed(repoLabel string, err error) error {
	return fmt.Errorf("failed to clone repository %s: %w", repoLabel, err)
}

// ==================== Kafka Errors ====================

func KafkaNotReady(err error) error {
	return fmt.Errorf("%w: kafka not ready: %w", ErrNotReady, err)
}

func KafkaBrokerAPIFailed() error {
	return fmt.Errorf("%w: no output from Kafka broker API", ErrNotReady)
}

func ConsumerGroupRebalanceTimeout(consumerGroup string, err error) error {
	return fmt.Errorf("%w: consumer group %s rebalance exceeded: %w", ErrTimeout, consumerGroup, err)
}

func ConsumerGroupPollTimeout(consumerGroup string, maxRetries int) error {
	return fmt.Errorf("%w: consumer group %s polling exceeded maximum retries (%d)", ErrTimeout, consumerGroup, maxRetries)
}

func ContainerCommandFailed(stderr string) error {
	return fmt.Errorf("failed to execute container command, stderr: %s", stderr)
}

// ==================== Keycloak Errors ====================

func AccessTokenNotFound(requestURL string) error {
	return fmt.Errorf("%w: access token from response: %s", ErrNotFound, requestURL)
}

func ClientNotFound(clientID string) error {
	return fmt.Errorf("%w: expected exactly 1 client with id %s", ErrNotFound, clientID)
}

func RoleNotFound(roleName string) error {
	return fmt.Errorf("%w: expected exactly 1 role with name %s", ErrNotFound, roleName)
}

func UserNotFound(username, tenantName string) error {
	return fmt.Errorf("%w: user %s in tenant %s", ErrNotFound, username, tenantName)
}

// ==================== Kong Errors ====================

func KongRoutesNotReady(expected int) error {
	return fmt.Errorf("%w: kong routes %d", ErrNotReady, expected)
}

func KongAdminAPIFailed(statusCode int, status string) error {
	return fmt.Errorf("kong admin API failed: %d %s", statusCode, status)
}

// ==================== Module Errors ====================

func ModulesNotDeployed(expectedModules int) error {
	return fmt.Errorf("%d modules not deployed", expectedModules)
}

func ModuleNotReady(moduleName string) error {
	return fmt.Errorf("%w: module %s", ErrNotReady, moduleName)
}

func ModulePullFailed(imageName string, err error) error {
	return fmt.Errorf("%w: failed to pull module image %s: %w", ErrDeploymentFailed, imageName, err)
}

func SidecarDeployFailed(sidecarName string, err error) error {
	return fmt.Errorf("%w: failed to deploy sidecar %s: %w", ErrDeploymentFailed, sidecarName, err)
}

func SidecarVersionNotFound() error {
	return fmt.Errorf("%w: sidecar version in registry", ErrNotFound)
}

func LocalDescriptorNotFound(path, moduleName string) error {
	return fmt.Errorf("%w: local descriptor %s for module %s", ErrNotFound, path, moduleName)
}

func EmptyLineNotFound(id string) error {
	return fmt.Errorf("response does not contain an empty line using id %s", id)
}

func ImageKeyNotSet(imageName, fieldName string) error {
	return fmt.Errorf("%w: cannot run image %s, key %s not set in config", ErrConfigMissing, imageName, fieldName)
}

func ModuleDiscoveryNotFound(moduleName string) error {
	return fmt.Errorf("%w: module discovery %s in application", ErrNotFound, moduleName)
}

// ==================== Tenant Errors ====================

func TenantNotFound(tenantName string) error {
	return fmt.Errorf("%w: tenant %s in config", ErrNotFound, tenantName)
}

func CentralTenantNotFound(consortiumName string) error {
	return fmt.Errorf("%w: central tenant in consortium %s", ErrNotFound, consortiumName)
}

func TenantNotCreated(tenantName string) error {
	return fmt.Errorf("%w: consortium tenant %s not created", ErrDeploymentFailed, tenantName)
}

// ==================== Search/Reindex Errors ====================

func ReindexJobHasErrors(jobErrors []any) error {
	return fmt.Errorf("reindex job has %d error(s): %+v", len(jobErrors), jobErrors)
}

func ReindexJobIDBlank() error {
	return errors.New("reindex job id is blank")
}

// ==================== Registry Errors ====================

func LocalInstallFileNotFound(err error) error {
	return fmt.Errorf("%w: failed to find local install file: %w", ErrNotFound, err)
}

// ==================== Flag Errors ====================

func RegisterFlagCompletionFailed(err error) error {
	return fmt.Errorf("failed to register flag completion function: %w", err)
}

func MarkFlagRequiredFailed(flag FlagReader, err error) error {
	return fmt.Errorf("failed to mark %s flag as required: %w", flag.GetName(), err)
}
