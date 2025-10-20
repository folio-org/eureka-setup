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
	endpointConfig := map[string]*network.EndpointSettings{constant.NetworkID: {
		NetworkID: constant.NetworkID,
		Aliases:   []string{constant.NetworkAlias},
	}}

	return &network.NetworkingConfig{EndpointsConfig: endpointConfig}
}

func CreateExposedPorts(serverPort int) *nat.PortSet {
	moduleExposedPorts := make(map[nat.Port]struct{})
	moduleExposedPorts[nat.Port(strconv.Itoa(serverPort))] = struct{}{}
	moduleExposedPorts[nat.Port(constant.DebugPort)] = struct{}{}

	portSet := nat.PortSet(moduleExposedPorts)

	return &portSet
}

func CreatePortBindings(hostServerPort int, hostServerDebugPort int, serverPort int) *nat.PortMap {
	var (
		serverPortBinding      []nat.PortBinding
		serverDebugPortBinding []nat.PortBinding
	)

	serverPortBinding = append(serverPortBinding, nat.PortBinding{
		HostIP:   constant.HostIP,
		HostPort: strconv.Itoa(hostServerPort),
	})

	serverDebugPortBinding = append(serverDebugPortBinding, nat.PortBinding{
		HostIP:   constant.HostIP,
		HostPort: strconv.Itoa(hostServerDebugPort),
	})

	portBindings := make(map[nat.Port][]nat.PortBinding)
	portBindings[nat.Port(strconv.Itoa(serverPort))] = serverPortBinding
	portBindings[nat.Port(constant.DebugPort)] = serverDebugPortBinding

	portMap := nat.PortMap(portBindings)

	return &portMap
}

func CreateResources(isModule bool, resources map[string]any) *container.Resources {
	if len(resources) == 0 {
		return createDefaultResources(isModule)
	}

	return &container.Resources{
		CPUCount:          GetIntOrDefault(resources, field.ModuleResourceCpuCountEntry, constant.ModuleCPU),
		MemoryReservation: ConvertMiBToBytes(GetIntOrDefault(resources, field.ModuleResourceMemoryReservationEntry, constant.ModuleMemoryReservation)),
		Memory:            ConvertMiBToBytes(GetIntOrDefault(resources, field.ModuleResourceMemoryEntry, constant.ModuleMemory)),
		MemorySwap:        ConvertMiBToBytes(GetIntOrDefault(resources, field.ModuleResourceMemorySwapEntry, constant.ModuleSwap)),
		OomKillDisable:    BoolP(GetBoolOrDefault(resources, field.ModuleResourceOomKillDisableEntry, false)),
	}
}

func createDefaultResources(isModule bool) *container.Resources {
	if isModule {
		return &container.Resources{
			CPUCount:          constant.ModuleCPU,
			MemoryReservation: ConvertMiBToBytes(constant.ModuleMemoryReservation),
			Memory:            ConvertMiBToBytes(constant.ModuleMemory),
			MemorySwap:        ConvertMiBToBytes(constant.ModuleSwap),
			OomKillDisable:    BoolP(false),
		}
	}

	return &container.Resources{
		CPUCount:          constant.SidecarCPU,
		MemoryReservation: ConvertMiBToBytes(constant.SidecarMemoryReservation),
		Memory:            ConvertMiBToBytes(constant.SidecarMemory),
		MemorySwap:        ConvertMiBToBytes(constant.SidecarSwap),
		OomKillDisable:    BoolP(false),
	}
}

func AppendAdditionalRequiredContainers(action *action.Action, initialRequiredContainers []string) []string {
	if IsModuleEnabled(constant.ModSearchModule) {
		initialRequiredContainers = append(initialRequiredContainers, constant.ElasticsearchContainer)
	}
	if IsModuleEnabled(constant.ModDataExportWorkerModule) {
		extraContainers := []string{constant.MinIOContainer, constant.CreateBucketsContainer, constant.FTPServerContainer}
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
