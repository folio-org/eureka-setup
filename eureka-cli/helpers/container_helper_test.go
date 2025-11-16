package helpers_test

import (
	"testing"

	"github.com/docker/go-connections/nat"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/folio-org/eureka-cli/helpers"
	"github.com/stretchr/testify/assert"
)

func TestGetModuleNetworkConfig_ReturnsValidConfig(t *testing.T) {
	// Act
	result := helpers.GetModuleNetworkConfig()

	// Assert
	assert.NotNil(t, result)
	assert.NotNil(t, result.EndpointsConfig)
	assert.Contains(t, result.EndpointsConfig, constant.NetworkID)
	assert.Equal(t, constant.NetworkID, result.EndpointsConfig[constant.NetworkID].NetworkID)
	assert.Contains(t, result.EndpointsConfig[constant.NetworkID].Aliases, constant.NetworkAlias)
}

func TestGetPlatform_ReturnsEmptyPlatform(t *testing.T) {
	// Act
	result := helpers.GetPlatform()

	// Assert
	assert.NotNil(t, result)
}

func TestGetRestartPolicy_ReturnsAlwaysPolicy(t *testing.T) {
	// Act
	result := helpers.GetRestartPolicy()

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, "always", string(result.Name))
}

func TestGetConfigSidecarCmd_WithCmd(t *testing.T) {
	// Arrange
	cmd := []string{"./application", "-Dquarkus.http.host=0.0.0.0"}

	// Act
	result := helpers.GetConfigSidecarCmd(cmd)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, cmd, []string(result))
}

func TestGetConfigSidecarCmd_WithEmptyCmd(t *testing.T) {
	// Arrange
	cmd := []string{}

	// Act
	result := helpers.GetConfigSidecarCmd(cmd)

	// Assert
	assert.Nil(t, result)
}

func TestGetConfigSidecarCmd_WithNilCmd(t *testing.T) {
	// Act
	result := helpers.GetConfigSidecarCmd(nil)

	// Assert
	assert.Nil(t, result)
}

func TestGetSidecarName_StandardModule(t *testing.T) {
	// Arrange
	moduleName := "mod-users"

	// Act
	result := helpers.GetSidecarName(moduleName)

	// Assert
	assert.Equal(t, "mod-users-sc", result)
}

func TestGetSidecarName_EmptyModule(t *testing.T) {
	// Arrange
	moduleName := ""

	// Act
	result := helpers.GetSidecarName(moduleName)

	// Assert
	assert.Equal(t, "-sc", result)
}

func TestCreateExposedPorts_ValidPort(t *testing.T) {
	// Arrange
	privateServerPort := 8081

	// Act
	result := helpers.CreateExposedPorts(privateServerPort)

	// Assert
	assert.NotNil(t, result)
	assert.Len(t, *result, 2)
	assert.Contains(t, *result, nat.Port("8081"))
	assert.Contains(t, *result, nat.Port(constant.PrivateDebugPort))
}

func TestCreatePortBindings_ValidPorts(t *testing.T) {
	// Arrange
	hostServerPort := 9000
	hostServerDebugPort := 5005
	privateServerPort := 8081

	// Act
	result := helpers.CreatePortBindings(hostServerPort, hostServerDebugPort, privateServerPort)

	// Assert
	assert.NotNil(t, result)
	assert.Len(t, *result, 2)

	serverBindings := (*result)[nat.Port("8081")]
	assert.Len(t, serverBindings, 1)
	assert.Equal(t, constant.HostIP, serverBindings[0].HostIP)
	assert.Equal(t, "9000", serverBindings[0].HostPort)

	debugBindings := (*result)[nat.Port(constant.PrivateDebugPort)]
	assert.Len(t, debugBindings, 1)
	assert.Equal(t, constant.HostIP, debugBindings[0].HostIP)
	assert.Equal(t, "5005", debugBindings[0].HostPort)
}

func TestCreateResources_WithCustomResources(t *testing.T) {
	// Arrange
	resources := map[string]any{
		field.ModuleResourceCpuCountEntry:          2,
		field.ModuleResourceMemoryReservationEntry: 256,
		field.ModuleResourceMemoryEntry:            1024,
		field.ModuleResourceMemorySwapEntry:        2048,
		field.ModuleResourceOomKillDisableEntry:    true,
	}

	// Act
	result := helpers.CreateResources(true, resources)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, int64(2), result.CPUCount)
	assert.Equal(t, int64(268435456), result.MemoryReservation) // 256 MiB in bytes
	assert.Equal(t, int64(1073741824), result.Memory)           // 1024 MiB in bytes
	assert.Equal(t, int64(2147483648), result.MemorySwap)       // 2048 MiB in bytes
	assert.NotNil(t, result.OomKillDisable)
	assert.True(t, *result.OomKillDisable)
}

func TestCreateResources_EmptyResourcesForModule(t *testing.T) {
	// Arrange
	resources := map[string]any{}

	// Act
	result := helpers.CreateResources(true, resources)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, int64(constant.ModuleCPU), result.CPUCount)
	assert.Equal(t, int64(constant.ModuleMemoryReservation*1024*1024), result.MemoryReservation)
	assert.Equal(t, int64(constant.ModuleMemory*1024*1024), result.Memory)
	assert.Equal(t, int64(constant.ModuleSwap), result.MemorySwap) // -1 stays -1
	assert.NotNil(t, result.OomKillDisable)
	assert.False(t, *result.OomKillDisable)
}

func TestCreateResources_EmptyResourcesForSidecar(t *testing.T) {
	// Arrange
	resources := map[string]any{}

	// Act
	result := helpers.CreateResources(false, resources)

	// Assert
	assert.NotNil(t, result)
	assert.Equal(t, int64(constant.SidecarCPU), result.CPUCount)
	assert.Equal(t, int64(constant.SidecarMemoryReservation*1024*1024), result.MemoryReservation)
	assert.Equal(t, int64(constant.SidecarMemory*1024*1024), result.Memory)
	assert.Equal(t, int64(constant.SidecarSwap), result.MemorySwap) // -1 stays -1
	assert.NotNil(t, result.OomKillDisable)
	assert.False(t, *result.OomKillDisable)
}

func TestAppendRequiredContainers_WithSearchModule(t *testing.T) {
	// Arrange
	actionName := "TestAction"
	requiredContainers := []string{}
	configBackendModules := map[string]any{
		constant.ModSearchModule: map[string]any{
			field.ModuleDeployModuleEntry: true,
		},
	}

	// Act
	result := helpers.AppendRequiredContainers(actionName, requiredContainers, configBackendModules)

	// Assert
	assert.Contains(t, result, constant.ElasticsearchContainer)
}

func TestAppendRequiredContainers_WithDataExportWorkerModule(t *testing.T) {
	// Arrange
	actionName := "TestAction"
	requiredContainers := []string{}
	configBackendModules := map[string]any{
		constant.ModDataExportWorkerModule: map[string]any{
			field.ModuleDeployModuleEntry: true,
		},
	}

	// Act
	result := helpers.AppendRequiredContainers(actionName, requiredContainers, configBackendModules)

	// Assert
	assert.Contains(t, result, constant.MinIOContainer)
	assert.Contains(t, result, constant.CreateBucketsContainer)
	assert.Contains(t, result, constant.FTPServerContainer)
}

func TestAppendRequiredContainers_WithBothModules(t *testing.T) {
	// Arrange
	actionName := "TestAction"
	requiredContainers := []string{"existing-container"}
	configBackendModules := map[string]any{
		constant.ModSearchModule: map[string]any{
			field.ModuleDeployModuleEntry: true,
		},
		constant.ModDataExportWorkerModule: map[string]any{
			field.ModuleDeployModuleEntry: true,
		},
	}

	// Act
	result := helpers.AppendRequiredContainers(actionName, requiredContainers, configBackendModules)

	// Assert
	assert.Contains(t, result, "existing-container")
	assert.Contains(t, result, constant.ElasticsearchContainer)
	assert.Contains(t, result, constant.MinIOContainer)
	assert.Contains(t, result, constant.CreateBucketsContainer)
	assert.Contains(t, result, constant.FTPServerContainer)
	assert.Len(t, result, 5)
}

func TestAppendRequiredContainers_NoModulesEnabled(t *testing.T) {
	// Arrange
	actionName := "TestAction"
	requiredContainers := []string{}
	configBackendModules := map[string]any{}

	// Act
	result := helpers.AppendRequiredContainers(actionName, requiredContainers, configBackendModules)

	// Assert
	assert.Empty(t, result)
}

func TestAppendRequiredContainers_ModulesDisabled(t *testing.T) {
	// Arrange
	actionName := "TestAction"
	requiredContainers := []string{}
	configBackendModules := map[string]any{
		constant.ModSearchModule: map[string]any{
			field.ModuleDeployModuleEntry: false,
		},
	}

	// Act
	result := helpers.AppendRequiredContainers(actionName, requiredContainers, configBackendModules)

	// Assert
	assert.Empty(t, result)
}
