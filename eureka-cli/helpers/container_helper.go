package helpers

import (
	"fmt"
	"log/slog"
	"strconv"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/strslice"
	"github.com/docker/go-connections/nat"
	"github.com/folio-org/eureka-cli/constant"
	"github.com/folio-org/eureka-cli/field"
	v1 "github.com/opencontainers/image-spec/specs-go/v1"
)

func GetModuleNetworkConfig() *network.NetworkingConfig {
	return &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{constant.NetworkID: {
			NetworkID: constant.NetworkID,
			Aliases:   []string{constant.NetworkAlias},
		}},
	}
}

func GetPlatform() *v1.Platform {
	return &v1.Platform{}
}

func GetRestartPolicy() *container.RestartPolicy {
	return &container.RestartPolicy{
		Name: container.RestartPolicyAlways,
	}
}

func GetConfigSidecarCmd(cmd []string) strslice.StrSlice {
	if len(cmd) > 0 {
		return cmd
	}

	return nil
}

func GetSidecarName(moduleName string) string {
	return fmt.Sprintf("%s-sc", moduleName)
}

func CreateExposedPorts(privateServerPort int) *nat.PortSet {
	exposedPorts := make(map[nat.Port]struct{})
	exposedPorts[nat.Port(strconv.Itoa(privateServerPort))] = struct{}{}
	exposedPorts[nat.Port(constant.PrivateDebugPort)] = struct{}{}
	portSet := nat.PortSet(exposedPorts)

	return &portSet
}

func CreatePortBindings(hostServerPort int, hostServerDebugPort int, privateServerPort int) *nat.PortMap {
	var (
		serverPortBinding []nat.PortBinding
		debugPortBinding  []nat.PortBinding
	)
	serverPortBinding = append(serverPortBinding, nat.PortBinding{
		HostIP:   constant.HostIP,
		HostPort: strconv.Itoa(hostServerPort),
	})
	debugPortBinding = append(debugPortBinding, nat.PortBinding{
		HostIP:   constant.HostIP,
		HostPort: strconv.Itoa(hostServerDebugPort),
	})
	portBindings := make(map[nat.Port][]nat.PortBinding)
	portBindings[nat.Port(strconv.Itoa(privateServerPort))] = serverPortBinding
	portBindings[nat.Port(constant.PrivateDebugPort)] = debugPortBinding
	portMap := nat.PortMap(portBindings)

	return &portMap
}

func CreateResources(isModule bool, r map[string]any) *container.Resources {
	if len(r) == 0 {
		return createDefaultResources(isModule)
	}

	return &container.Resources{
		CPUCount:          GetIntOrDefault(r, field.ModuleResourceCpuCountEntry, constant.ModuleCPU),
		MemoryReservation: ConvertMemory(MibToBytes, GetIntOrDefault(r, field.ModuleResourceMemoryReservationEntry, constant.ModuleMemoryReservation)),
		Memory:            ConvertMemory(MibToBytes, GetIntOrDefault(r, field.ModuleResourceMemoryEntry, constant.ModuleMemory)),
		MemorySwap:        ConvertMemory(MibToBytes, GetIntOrDefault(r, field.ModuleResourceMemorySwapEntry, constant.ModuleSwap)),
		OomKillDisable:    BoolP(GetBoolOrDefault(r, field.ModuleResourceOomKillDisableEntry, false)),
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

func AppendRequiredContainers(actionName string, containers []string, backendModules map[string]any) []string {
	if IsModuleEnabled(constant.ModSearchModule, backendModules) {
		containers = append(containers, constant.ElasticsearchContainer)
	}
	if IsModuleEnabled(constant.ModDataExportWorkerModule, backendModules) {
		extraContainers := []string{constant.MinIOContainer, constant.CreateBucketsContainer, constant.FTPServerContainer}
		containers = append(containers, extraContainers...)
	}
	if len(containers) > 0 {
		slog.Info(actionName, "text", "Retrieved required containers", "containers", containers)
	}

	return containers
}
