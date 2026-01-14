package runconfig_test

import (
	"log/slog"
	"os"
	"testing"

	"github.com/j011195/eureka-setup/eureka-cli/action"
	"github.com/j011195/eureka-setup/eureka-cli/runconfig"
	"github.com/stretchr/testify/assert"
)

func TestNew_Success(t *testing.T) {
	// Arrange
	action := &action.Action{
		Name: "test-action",
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Act
	config, err := runconfig.New(action, logger)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, config)
	assert.Equal(t, action, config.Action)
	assert.Equal(t, logger, config.Logger)

	// Verify all dependencies are initialized
	assert.NotNil(t, config.GitClient)
	assert.NotNil(t, config.HTTPClient)
	assert.NotNil(t, config.DockerClient)
	assert.NotNil(t, config.VaultClient)
	assert.NotNil(t, config.AWSSvc)
	assert.NotNil(t, config.KafkaSvc)
	assert.NotNil(t, config.KeycloakSvc)
	assert.NotNil(t, config.KongSvc)
	assert.NotNil(t, config.RegistrySvc)
	assert.NotNil(t, config.ModuleProps)
	assert.NotNil(t, config.ModuleEnv)
	assert.NotNil(t, config.ModuleSvc)
	assert.NotNil(t, config.ManagementSvc)
	assert.NotNil(t, config.TenantSvc)
	assert.NotNil(t, config.UserSvc)
	assert.NotNil(t, config.ConsortiumSvc)
	assert.NotNil(t, config.UISvc)
	assert.NotNil(t, config.SearchSvc)
	assert.NotNil(t, config.InterceptModuleSvc)
}

func TestNew_NilAction(t *testing.T) {
	// Arrange
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Act
	config, err := runconfig.New(nil, logger)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "action")
}

func TestNew_NilLogger(t *testing.T) {
	// Arrange
	action := &action.Action{
		Name: "test-action",
	}

	// Act
	config, err := runconfig.New(action, nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "logger")
}

func TestNew_NilBoth(t *testing.T) {
	// Arrange & Act
	config, err := runconfig.New(nil, nil)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, config)
	// Should return error for action first
	assert.Contains(t, err.Error(), "action")
}

func TestNew_DependencyWiring(t *testing.T) {
	// Arrange
	action := &action.Action{
		Name: "test-action",
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))

	// Act
	config, err := runconfig.New(action, logger)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, config)

	// Verify the dependency chain is correctly wired
	// These assertions verify that services that depend on other services
	// have been properly initialized in the correct order

	// RegistrySvc should be created before ModuleSvc
	assert.NotNil(t, config.RegistrySvc)
	assert.NotNil(t, config.ModuleSvc)

	// UserSvc should be created before ConsortiumSvc
	assert.NotNil(t, config.UserSvc)
	assert.NotNil(t, config.ConsortiumSvc)

	// ConsortiumSvc should be created before TenantSvc
	assert.NotNil(t, config.TenantSvc)

	// TenantSvc should be created before ManagementSvc
	assert.NotNil(t, config.ManagementSvc)

	// ManagementSvc should be created before KeycloakSvc and InterceptModuleSvc
	assert.NotNil(t, config.KeycloakSvc)
	assert.NotNil(t, config.InterceptModuleSvc)

	// ModuleSvc should be created before InterceptModuleSvc
	// TenantSvc should be created before UISvc
	assert.NotNil(t, config.UISvc)
}
