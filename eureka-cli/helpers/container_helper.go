package helpers

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
)

func NewModuleNetworkConfig() *network.NetworkingConfig {
	endpointConfig := map[string]*network.EndpointSettings{constant.NetworkID: {
		NetworkID: constant.NetworkID,
		Aliases:   []string{constant.NetworkAlias},
	}}

	return &network.NetworkingConfig{EndpointsConfig: endpointConfig}
}

func GetSidecarName(moduleName string) string {
	return fmt.Sprintf("%s-sc", moduleName)
}

func CreateExposedPorts(serverPort int) *nat.PortSet {
	moduleExposedPorts := make(map[nat.Port]struct{})
	moduleExposedPorts[nat.Port(strconv.Itoa(serverPort))] = struct{}{}
	moduleExposedPorts[nat.Port(constant.PrivateDebugPort)] = struct{}{}
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
	portBindings[nat.Port(constant.PrivateDebugPort)] = serverDebugPortBinding
	portMap := nat.PortMap(portBindings)

	return &portMap
}

func CreateResources(isModule bool, resources map[string]any) *container.Resources {
	if len(resources) == 0 {
		return createDefaultResources(isModule)
	}

	return &container.Resources{
		CPUCount:          GetIntOrDefault(resources, field.ModuleResourceCpuCountEntry, constant.ModuleCPU),
		MemoryReservation: ConvertMemory(MibToBytes, GetIntOrDefault(resources, field.ModuleResourceMemoryReservationEntry, constant.ModuleMemoryReservation)),
		Memory:            ConvertMemory(MibToBytes, GetIntOrDefault(resources, field.ModuleResourceMemoryEntry, constant.ModuleMemory)),
		MemorySwap:        ConvertMemory(MibToBytes, GetIntOrDefault(resources, field.ModuleResourceMemorySwapEntry, constant.ModuleSwap)),
		OomKillDisable:    BoolP(GetBoolOrDefault(resources, field.ModuleResourceOomKillDisableEntry, false)),
	}
}

func createDefaultResources(isModule bool) *container.Resources {
	if isModule {
		return &container.Resources{
			CPUCount:          constant.ModuleCPU,
			MemoryReservation: ConvertMemory(MibToBytes, constant.ModuleMemoryReservation),
			Memory:            ConvertMemory(MibToBytes, constant.ModuleMemory),
			MemorySwap:        ConvertMemory(MibToBytes, constant.ModuleSwap),
			OomKillDisable:    BoolP(false),
		}
	}

	return &container.Resources{
		CPUCount:          constant.SidecarCPU,
		MemoryReservation: ConvertMemory(MibToBytes, constant.SidecarMemoryReservation),
		Memory:            ConvertMemory(MibToBytes, constant.SidecarMemory),
		MemorySwap:        ConvertMemory(MibToBytes, constant.SidecarSwap),
		OomKillDisable:    BoolP(false),
	}
}

func AppendAdditionalRequiredContainers(actionName string, initialRequiredContainers []string, configBackendModules map[string]any) []string {
	if IsModuleEnabled(constant.ModSearchModule, configBackendModules) {
		initialRequiredContainers = append(initialRequiredContainers, constant.ElasticsearchContainer)
	}
	if IsModuleEnabled(constant.ModDataExportWorkerModule, configBackendModules) {
		extraContainers := []string{constant.MinIOContainer, constant.CreateBucketsContainer, constant.FTPServerContainer}
		initialRequiredContainers = append(initialRequiredContainers, extraContainers...)
	}
	if len(initialRequiredContainers) > 0 {
		slog.Info(actionName, "text", "Retrieved required containers", "containers", initialRequiredContainers)
	}

	return initialRequiredContainers
}

func IsModuleEnabled(module string, configBackendModules map[string]any) bool {
	value, exists := configBackendModules[module]
	if !exists || value == nil {
		return false
	}

	entry, ok := value.(map[string]any)
	if !ok {
		return false
	}

	deploy, ok := entry[field.ModuleDeployModuleEntry]
	if !ok {
		return true
	}
	enabled, ok := deploy.(bool)

	return ok && enabled
}

func IsUIEnabled(tenantName string, configTenants map[string]any) bool {
	value, exists := configTenants[tenantName]
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
