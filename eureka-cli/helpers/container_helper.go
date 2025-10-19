package helpers

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/folio-org/eureka-cli/action"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	"github.com/spf13/viper"
)

func NewModuleNetworkConfig() *network.NetworkingConfig {
	endpointConfig := map[string]*network.EndpointSettings{constant.DefaultNetworkID: {
		NetworkID: constant.DefaultNetworkID,
		Aliases:   []string{constant.DefaultNetworkAlias},
	}}

	return &network.NetworkingConfig{EndpointsConfig: endpointConfig}
}

func CreateExposedPorts(serverPort int) *nat.PortSet {
	moduleExposedPorts := make(map[nat.Port]struct{})
	moduleExposedPorts[nat.Port(strconv.Itoa(serverPort))] = struct{}{}
	moduleExposedPorts[nat.Port(constant.DefaultDebugPort)] = struct{}{}

	portSet := nat.PortSet(moduleExposedPorts)

	return &portSet
}

func CreatePortBindings(hostServerPort int, hostServerDebugPort int, serverPort int) *nat.PortMap {
	var (
		serverPortBinding      []nat.PortBinding
		serverDebugPortBinding []nat.PortBinding
	)

	serverPortBinding = append(serverPortBinding, nat.PortBinding{
		HostIP:   constant.DefaultHostIP,
		HostPort: strconv.Itoa(hostServerPort),
	})

	serverDebugPortBinding = append(serverDebugPortBinding, nat.PortBinding{
		HostIP:   constant.DefaultHostIP,
		HostPort: strconv.Itoa(hostServerDebugPort),
	})

	portBindings := make(map[nat.Port][]nat.PortBinding)
	portBindings[nat.Port(strconv.Itoa(serverPort))] = serverPortBinding
	portBindings[nat.Port(constant.DefaultDebugPort)] = serverDebugPortBinding

	portMap := nat.PortMap(portBindings)

	return &portMap
}

func CreateResources(isModule bool, resources map[string]any) *container.Resources {
	if len(resources) == 0 {
		return createDefaultResources(isModule)
	}

	return &container.Resources{
		CPUCount:          GetIntOrDefault(resources, field.ModuleResourceCpuCountEntry, constant.DefaultModuleCPU),
		MemoryReservation: ConvertMiBToBytes(GetIntOrDefault(resources, field.ModuleResourceMemoryReservationEntry, constant.DefaultModuleMemoryReservation)),
		Memory:            ConvertMiBToBytes(GetIntOrDefault(resources, field.ModuleResourceMemoryEntry, constant.DefaultModuleMemory)),
		MemorySwap:        ConvertMiBToBytes(GetIntOrDefault(resources, field.ModuleResourceMemorySwapEntry, constant.DefaultModuleSwap)),
		OomKillDisable:    BoolP(GetBoolOrDefault(resources, field.ModuleResourceOomKillDisableEntry, false)),
	}
}

func createDefaultResources(isModule bool) *container.Resources {
	if isModule {
		return &container.Resources{
			CPUCount:          constant.DefaultModuleCPU,
			MemoryReservation: ConvertMiBToBytes(constant.DefaultModuleMemoryReservation),
			Memory:            ConvertMiBToBytes(constant.DefaultModuleMemory),
			MemorySwap:        ConvertMiBToBytes(constant.DefaultModuleSwap),
			OomKillDisable:    BoolP(false),
		}
	}

	return &container.Resources{
		CPUCount:          constant.DefaultSidecarCPU,
		MemoryReservation: ConvertMiBToBytes(constant.DefaultSidecarMemoryReservation),
		Memory:            ConvertMiBToBytes(constant.DefaultSidecarMemory),
		MemorySwap:        ConvertMiBToBytes(constant.DefaultSidecarSwap),
		OomKillDisable:    BoolP(false),
	}
}

func AppendAdditionalRequiredContainers(action *action.Action, initialRequiredContainers []string) []string {
	if IsModuleEnabled(constant.ModSearchModuleName) {
		initialRequiredContainers = append(initialRequiredContainers, constant.ElasticsearchContainerName)
	}
	if IsModuleEnabled(constant.ModDataExportWorkerModuleName) {
		extraContainers := []string{constant.MinioContainerName, constant.CreateBucketsContainerName, constant.FtpServerContainerName}
		initialRequiredContainers = append(initialRequiredContainers, extraContainers...)
	}
	if len(initialRequiredContainers) > 0 {
		slog.Info(action.Name, "text", fmt.Sprintf("Retrieved required containers: %s", initialRequiredContainers))
	}

	return initialRequiredContainers
}

// IsModuleEnabled returns true if the module is present and enabled, or if its value is nil or missing the deploy entry
func IsModuleEnabled(module string) bool {
	value, exists := viper.GetStringMap(field.BackendModules)[module]
	if !exists || value == nil {
		return false
	}

	entry, ok := value.(map[string]any)
	if !ok {
		return false
	}

	deploy, ok := entry[field.ModuleDeployModuleEntry]
	if !ok {
		return true // If the deploy entry is missing, treat as enabled (legacy behavior)
	}
	enabled, ok := deploy.(bool)

	return ok && enabled
}

// IsUIEnabled returns true if the tenant has UI deployment enabled
func IsUIEnabled(tenant string) bool {
	value, exists := viper.GetStringMap(field.Tenants)[tenant]
	if !exists || value == nil {
		return false
	}

	entry, ok := value.(map[string]any)
	if !ok {
		return false
	}
	deploy, ok := entry[field.TenantsDeployUIEntry]
	enabled, isBool := deploy.(bool)

	return ok && isBool && enabled
}
