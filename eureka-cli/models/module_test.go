package models

import (
	"testing"

	"github.com/folio-org/eureka-setup/eureka-cli/action"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

// ==================== NewBackendModuleWithSidecar Tests ====================

func TestNewBackendModuleWithSidecar_Success_WithDeployModule(t *testing.T) {
	// Arrange
	viper.Set("application.port-start", 60000)
	viper.Set("application.port-end", 60999)
	t.Cleanup(func() {
		viper.Set("application.port-start", 0)
		viper.Set("application.port-end", 0)
	})

	params := &action.Param{}
	act := action.New("test-action", "http://localhost:%s", params)

	version := "19.2.3"
	port := 8081
	privatePort := 8080
	deploySidecar := true

	props := BackendModuleProperties{
		DeployModule:        true,
		DeploySidecar:       &deploySidecar,
		UseVault:            true,
		UseOkapiURL:         true,
		DisableSystemUser:   false,
		LocalDescriptorPath: "/descriptors/mod-users.json",
		Name:                "mod-users",
		Version:             &version,
		Port:                &port,
		PrivatePort:         &privatePort,
		Env: map[string]any{
			"JAVA_OPTIONS": "-Xmx512m",
		},
		Resources: map[string]any{
			"memory": "1Gi",
		},
		Volumes: []string{"/data:/data"},
	}

	// Act
	result, err := NewBackendModuleWithSidecar(act, props)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.DeployModule)
	assert.True(t, result.UseVault)
	assert.True(t, result.UseOkapiURL)
	assert.False(t, result.DisableSystemUser)
	assert.Equal(t, "mod-users", result.ModuleName)
	assert.Equal(t, "19.2.3", *result.ModuleVersion)
	assert.Equal(t, 8081, result.ModuleExposedServerPort)
	assert.Greater(t, result.ModuleExposedDebugPort, 0)
	assert.Equal(t, 8080, result.PrivatePort)
	assert.True(t, result.DeploySidecar)
	assert.Greater(t, result.SidecarExposedServerPort, 0)
	assert.Greater(t, result.SidecarExposedDebugPort, 0)
	assert.NotNil(t, result.ModuleExposedPorts)
	assert.NotNil(t, result.ModulePortBindings)
	assert.NotNil(t, result.SidecarExposedPorts)
	assert.NotNil(t, result.SidecarPortBindings)
	assert.Len(t, result.ModuleVolumes, 1)
	assert.Contains(t, result.ModuleVolumes, "/data:/data")
}

func TestNewBackendModuleWithSidecar_Success_WithoutDeployModule(t *testing.T) {
	// Arrange
	viper.Set("application.port-start", 60000)
	viper.Set("application.port-end", 60999)
	t.Cleanup(func() {
		viper.Set("application.port-start", 0)
		viper.Set("application.port-end", 0)
	})

	params := &action.Param{}
	act := action.New("test-action", "http://localhost:%s", params)

	version := "20.1.0"
	port := 8090
	privatePort := 8089
	deploySidecar := true

	props := BackendModuleProperties{
		DeployModule:  false,
		DeploySidecar: &deploySidecar,
		Name:          "mod-inventory",
		Version:       &version,
		Port:          &port,
		PrivatePort:   &privatePort,
		Env:           map[string]any{},
		Resources:     map[string]any{},
	}

	// Act
	result, err := NewBackendModuleWithSidecar(act, props)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.DeployModule)
	assert.Equal(t, 0, result.ModuleExposedDebugPort)
	assert.Equal(t, 0, result.SidecarExposedServerPort)
	assert.Equal(t, 0, result.SidecarExposedDebugPort)
	assert.True(t, result.DeploySidecar)
}

func TestNewBackendModuleWithSidecar_Error_NoFreePorts(t *testing.T) {
	// Arrange
	viper.Set("application.port-start", 0)
	viper.Set("application.port-end", 0)
	t.Cleanup(func() {
		viper.Set("application.port-start", 0)
		viper.Set("application.port-end", 0)
	})

	params := &action.Param{}
	act := action.New("test-action", "http://localhost:%s", params)

	version := "1.0.0"
	port := 8000
	privatePort := 8080
	deploySidecar := true

	props := BackendModuleProperties{
		DeployModule:  true,
		DeploySidecar: &deploySidecar,
		Name:          "mod-test",
		Version:       &version,
		Port:          &port,
		PrivatePort:   &privatePort,
		Env:           map[string]any{},
		Resources:     map[string]any{},
	}

	// Act
	result, err := NewBackendModuleWithSidecar(act, props)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find free TCP ports")
}

func TestNewBackendModuleWithSidecar_Success_WithEmptyEnvAndResources(t *testing.T) {
	// Arrange
	viper.Set("application.port-start", 61000)
	viper.Set("application.port-end", 61999)
	t.Cleanup(func() {
		viper.Set("application.port-start", 0)
		viper.Set("application.port-end", 0)
	})

	params := &action.Param{}
	act := action.New("test-action", "http://localhost:%s", params)

	port := 8100
	privatePort := 8099
	deploySidecar := false

	props := BackendModuleProperties{
		DeployModule:  false,
		DeploySidecar: &deploySidecar,
		Name:          "mod-simple",
		Port:          &port,
		PrivatePort:   &privatePort,
	}

	// Act
	result, err := NewBackendModuleWithSidecar(act, props)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.DeploySidecar)
	assert.Nil(t, result.ModuleEnv)
	assert.Nil(t, result.ModuleVolumes)
}

// ==================== NewBackendModule Tests ====================

func TestNewBackendModule_Success(t *testing.T) {
	// Arrange
	viper.Set("application.port-start", 62000)
	viper.Set("application.port-end", 62999)
	t.Cleanup(func() {
		viper.Set("application.port-start", 0)
		viper.Set("application.port-end", 0)
	})

	params := &action.Param{}
	act := action.New("test-action", "http://localhost:%s", params)

	version := "23.5.0"
	port := 9801
	privatePort := 9800

	props := BackendModuleProperties{
		DeployModule:        true,
		UseVault:            true,
		UseOkapiURL:         false,
		DisableSystemUser:   true,
		LocalDescriptorPath: "/path/to/descriptor.json",
		Name:                "mod-circulation",
		Version:             &version,
		Port:                &port,
		PrivatePort:         &privatePort,
		Env: map[string]any{
			"DB_HOST": "postgres",
		},
		Resources: map[string]any{
			"memory": "750m",
		},
		Volumes: []string{"/logs:/logs"},
	}

	// Act
	result, err := NewBackendModule(act, props)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.DeployModule)
	assert.True(t, result.UseVault)
	assert.False(t, result.UseOkapiURL)
	assert.True(t, result.DisableSystemUser)
	assert.Equal(t, "mod-circulation", result.ModuleName)
	assert.Equal(t, "23.5.0", *result.ModuleVersion)
	assert.Equal(t, 9801, result.ModuleExposedServerPort)
	assert.Greater(t, result.ModuleExposedDebugPort, 0)
	assert.Equal(t, 9800, result.PrivatePort)
	assert.False(t, result.DeploySidecar)
	assert.Nil(t, result.SidecarExposedPorts)
	assert.Nil(t, result.SidecarPortBindings)
	assert.NotNil(t, result.ModuleExposedPorts)
	assert.NotNil(t, result.ModulePortBindings)
	assert.Len(t, result.ModuleVolumes, 1)
}

func TestNewBackendModule_Error_NoFreePorts(t *testing.T) {
	// Arrange
	viper.Set("application.port-start", 0)
	viper.Set("application.port-end", 0)
	t.Cleanup(func() {
		viper.Set("application.port-start", 0)
		viper.Set("application.port-end", 0)
	})

	params := &action.Param{}
	act := action.New("test-action", "http://localhost:%s", params)

	version := "1.0.0"
	port := 8000
	privatePort := 8080

	props := BackendModuleProperties{
		DeployModule: true,
		Name:         "mod-test",
		Version:      &version,
		Port:         &port,
		PrivatePort:  &privatePort,
	}

	// Act
	result, err := NewBackendModule(act, props)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to find free TCP ports")
}

func TestNewBackendModule_Success_MinimalConfiguration(t *testing.T) {
	// Arrange
	viper.Set("application.port-start", 63000)
	viper.Set("application.port-end", 63999)
	t.Cleanup(func() {
		viper.Set("application.port-start", 0)
		viper.Set("application.port-end", 0)
	})

	params := &action.Param{}
	act := action.New("test-action", "http://localhost:%s", params)

	port := 8200
	privatePort := 8199

	props := BackendModuleProperties{
		DeployModule: false,
		Name:         "mod-minimal",
		Port:         &port,
		PrivatePort:  &privatePort,
	}

	// Act
	result, err := NewBackendModule(act, props)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.False(t, result.DeployModule)
	assert.Equal(t, "mod-minimal", result.ModuleName)
	assert.Nil(t, result.ModuleVersion)
	assert.False(t, result.UseVault)
	assert.False(t, result.UseOkapiURL)
	assert.False(t, result.DisableSystemUser)
	assert.Empty(t, result.LocalDescriptorPath)
}

func TestNewBackendModule_Success_WithMultipleVolumes(t *testing.T) {
	// Arrange
	viper.Set("application.port-start", 64000)
	viper.Set("application.port-end", 64999)
	t.Cleanup(func() {
		viper.Set("application.port-start", 0)
		viper.Set("application.port-end", 0)
	})

	params := &action.Param{}
	act := action.New("test-action", "http://localhost:%s", params)

	version := "5.0.0"
	port := 8300
	privatePort := 8299

	props := BackendModuleProperties{
		DeployModule: true,
		Name:         "mod-storage",
		Version:      &version,
		Port:         &port,
		PrivatePort:  &privatePort,
		Volumes: []string{
			"/data:/data",
			"/logs:/logs",
			"/config:/config",
		},
	}

	// Act
	result, err := NewBackendModule(act, props)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result.ModuleVolumes, 3)
	assert.Contains(t, result.ModuleVolumes, "/data:/data")
	assert.Contains(t, result.ModuleVolumes, "/logs:/logs")
	assert.Contains(t, result.ModuleVolumes, "/config:/config")
}

// ==================== Constructor Comparison Tests ====================

func TestBackendModule_WithSidecarVsWithout_StructureDifference(t *testing.T) {
	// Arrange
	viper.Set("application.port-start", 65000)
	viper.Set("application.port-end", 65999)
	t.Cleanup(func() {
		viper.Set("application.port-start", 0)
		viper.Set("application.port-end", 0)
	})

	params := &action.Param{}
	act := action.New("test-action", "http://localhost:%s", params)

	version := "1.0.0"
	port := 8400
	privatePort := 8399
	deploySidecar := true

	propsWithSidecar := BackendModuleProperties{
		DeployModule:  true,
		DeploySidecar: &deploySidecar,
		Name:          "mod-test",
		Version:       &version,
		Port:          &port,
		PrivatePort:   &privatePort,
	}

	propsWithoutSidecar := BackendModuleProperties{
		DeployModule: true,
		Name:         "mod-test",
		Version:      &version,
		Port:         &port,
		PrivatePort:  &privatePort,
	}

	// Act
	withSidecar, err1 := NewBackendModuleWithSidecar(act, propsWithSidecar)
	withoutSidecar, err2 := NewBackendModule(act, propsWithoutSidecar)

	// Assert
	assert.NoError(t, err1)
	assert.NoError(t, err2)

	// Module with sidecar should have sidecar ports allocated
	assert.True(t, withSidecar.DeploySidecar)
	assert.Greater(t, withSidecar.SidecarExposedServerPort, 0)
	assert.Greater(t, withSidecar.SidecarExposedDebugPort, 0)
	assert.NotNil(t, withSidecar.SidecarExposedPorts)
	assert.NotNil(t, withSidecar.SidecarPortBindings)

	// Module without sidecar should have no sidecar configuration
	assert.False(t, withoutSidecar.DeploySidecar)
	assert.Equal(t, 0, withoutSidecar.SidecarExposedServerPort)
	assert.Equal(t, 0, withoutSidecar.SidecarExposedDebugPort)
	assert.Nil(t, withoutSidecar.SidecarExposedPorts)
	assert.Nil(t, withoutSidecar.SidecarPortBindings)
}
